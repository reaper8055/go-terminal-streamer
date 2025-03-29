package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// OutputBuffer maintains a circular buffer of recent output and handles client broadcasting
type OutputBuffer struct {
	mu            sync.RWMutex
	recentLines   []string
	maxBufferSize int
	clients       map[*websocket.Conn]bool
	clientsMu     sync.RWMutex
}

// NewOutputBuffer creates a new output buffer with the specified capacity
func NewOutputBuffer(capacity int) *OutputBuffer {
	return &OutputBuffer{
		recentLines:   make([]string, 0, capacity),
		maxBufferSize: capacity,
		clients:       make(map[*websocket.Conn]bool),
	}
}

// AddLine adds a new line to the buffer and broadcasts to all clients
func (b *OutputBuffer) AddLine(line string) {
	b.mu.Lock()
	if len(b.recentLines) >= b.maxBufferSize {
		// Shift array to remove oldest line
		b.recentLines = append(b.recentLines[1:], line)
	} else {
		b.recentLines = append(b.recentLines, line)
	}
	b.mu.Unlock()

	// Broadcast to all clients
	b.Broadcast(line)
}

// GetLines returns a copy of the current lines
func (b *OutputBuffer) GetLines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	lines := make([]string, len(b.recentLines))
	copy(lines, b.recentLines)
	return lines
}

// AddClient registers a new WebSocket client
func (b *OutputBuffer) AddClient(ws *websocket.Conn) {
	b.clientsMu.Lock()
	b.clients[ws] = true
	b.clientsMu.Unlock()

	// Send buffer history to new client
	b.SendHistoryToClient(ws)
}

// RemoveClient unregisters a WebSocket client
func (b *OutputBuffer) RemoveClient(ws *websocket.Conn) {
	b.clientsMu.Lock()
	delete(b.clients, ws)
	b.clientsMu.Unlock()
}

// SendHistoryToClient sends all buffered lines to a specific client
func (b *OutputBuffer) SendHistoryToClient(ws *websocket.Conn) {
	lines := b.GetLines()
	for _, line := range lines {
		websocket.JSON.Send(ws, map[string]string{"line": line})
	}
}

// Broadcast sends a line to all connected clients
func (b *OutputBuffer) Broadcast(line string) {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	for client := range b.clients {
		err := websocket.JSON.Send(client, map[string]string{"line": line})
		if err != nil {
			// Non-blocking error handling - log and continue
			log.Printf("Error sending to client: %v", err)
			// Client will be properly removed when its handler exits
		}
	}
}

// ProcessStream reads from a reader and adds each line to the buffer
func (b *OutputBuffer) ProcessStream(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if prefix != "" {
			line = fmt.Sprintf("[%s] %s", prefix, line)
		}
		b.AddLine(line)
	}

	if err := scanner.Err(); err != nil {
		b.AddLine(fmt.Sprintf("[ERROR] Scanner error: %v", err))
	}
}

// ExecuteCommand runs a command and streams its output to the buffer
func ExecuteCommand(ctx context.Context, buffer *OutputBuffer, command string) error {
	// Split the command into command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Process stdout and stderr in separate goroutines
	go buffer.ProcessStream(stdout, "stdout")
	go buffer.ProcessStream(stderr, "stderr")

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
		buffer.AddLine(fmt.Sprintf("[SYSTEM] Command exited with error: %v", err))
		return err
	}

	buffer.AddLine(fmt.Sprintf("[SYSTEM] Command completed successfully"))
	return nil
}

// WSHandler handles WebSocket connections
func WSHandler(buffer *OutputBuffer) websocket.Handler {
	return func(ws *websocket.Conn) {
		buffer.AddClient(ws)
		defer buffer.RemoveClient(ws)

		// Keep connection alive
		for {
			var message string
			err := websocket.JSON.Receive(ws, &message)
			if err != nil {
				if err != io.EOF {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			// Currently we don't handle incoming messages from clients
		}
	}
}

// Serve the HTML page
func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Terminal Output Streamer</title>
    <style>
        body {
            font-family: monospace;
            background-color: #1e1e1e;
            color: #f0f0f0;
            margin: 0;
            padding: 10px;
        }
        #output {
            padding: 10px;
            white-space: pre-wrap;
            overflow-y: auto;
            height: calc(100vh - 60px);
            border: 1px solid #444;
            background-color: #2d2d2d;
        }
        .stdout { color: #a8ff60; }
        .stderr { color: #f08080; }
        .system { color: #80a0ff; }
    </style>
</head>
<body>
    <h2>Terminal Output</h2>
    <div id="output"></div>

    <script>
        const output = document.getElementById('output');
        const ws = new WebSocket('ws://' + window.location.host + '/ws');
        
        ws.onopen = function() {
            console.log('WebSocket connection established');
        };
        
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            const line = data.line;
            const div = document.createElement('div');
            div.textContent = line;
            
            if (line.includes('[stdout]')) {
                div.className = 'stdout';
            } else if (line.includes('[stderr]')) {
                div.className = 'stderr';
            } else if (line.includes('[SYSTEM]')) {
                div.className = 'system';
            }
            
            output.appendChild(div);
            output.scrollTop = output.scrollHeight; // Auto-scroll to bottom
        };
        
        ws.onclose = function() {
            const div = document.createElement('div');
            div.className = 'system';
            div.textContent = '[SYSTEM] WebSocket connection closed';
            output.appendChild(div);
        };
        
        ws.onerror = function(error) {
            console.error('WebSocket error:', error);
            const div = document.createElement('div');
            div.className = 'stderr';
            div.textContent = '[SYSTEM] WebSocket error occurred';
            output.appendChild(div);
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP service address")
	command := flag.String("cmd", "", "Command to execute")
	bufferSize := flag.Int("buffer", 1000, "Number of lines to buffer")
	flag.Parse()

	if *command == "" {
		fmt.Println("Please provide a command to execute using the -cmd flag")
		os.Exit(1)
	}

	// Create output buffer
	buffer := NewOutputBuffer(*bufferSize)

	// Set up HTTP server
	http.Handle("/ws", WSHandler(buffer))
	http.HandleFunc("/", serveHTML)

	// Run the command in a goroutine
	go func() {
		ctx := context.Background()
		buffer.AddLine(fmt.Sprintf("[SYSTEM] Executing command: %s", *command))
		err := ExecuteCommand(ctx, buffer, *command)
		if err != nil {
			log.Printf("Command execution error: %v", err)
		}
	}()

	// Start the server
	log.Printf("Starting server on %s", *addr)
	log.Printf("Open http://localhost%s in your browser", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
