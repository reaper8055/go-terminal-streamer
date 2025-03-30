# Terminal Output Streamer

A lightweight Go application that streams terminal command output to a web browser in real-time.

## Overview

Terminal Output Streamer captures the stdout and stderr streams from a running command and broadcasts them in real-time to connected web clients using WebSockets. This is particularly useful for monitoring long-running processes like build commands, deployments, or any command that produces output over time.

## Features

- **Real-time streaming** of command output to web browsers
- **Separate stdout/stderr handling** with distinct visual styling
- **Output buffering** to allow late-joining clients to see previous output
- **Simple web interface** that requires no additional dependencies
- **Standard library implementation** with minimal external dependencies

## Requirements

- Go 1.16+
- Standard library and `golang.org/x/net/websocket`

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/terminal-streamer
cd terminal-streamer

# Build the application
go build -o terminal-streamer
```

## Usage

```bash
# Run a simple command
./terminal-streamer -cmd "ls -la" -addr ":8080"

# Stream a long-running build process
./terminal-streamer -cmd "bazel build //..." -addr ":8080"

# Use the test generator script
./terminal-streamer -cmd "./test-output-generator.sh" -addr ":8080"
```

### Command-line Options

- `-cmd`: The command to execute and stream (required)
- `-addr`: HTTP service address (default: ":8080")
- `-buffer`: Number of output lines to buffer (default: 1000)

## Architecture

The application consists of several core components:

1. **Command Execution**: Uses Go's `os/exec` package to run the specified command and capture its output streams.

2. **Output Buffer**: Maintains a circular buffer of recent output lines and handles client broadcasting.

3. **WebSocket Server**: Provides real-time communication between the terminal process and web clients.

4. **Web Interface**: A simple HTML/CSS/JavaScript interface that displays the streamed output.

## How Linux Handles stdout/stderr

This project showcases how Linux handles standard output streams:

- In Linux, stdout is file descriptor 1, and stderr is file descriptor 2
- Each process gets its own set of file descriptors
- When a process writes to stdout/stderr, it writes to these file descriptors
- The kernel routes this output according to where the descriptors point
- For terminal applications, they typically point to the controlling terminal
- When using pipes or redirections, these descriptors can be rerouted

## Testing with the Output Generator

The repository includes a bash script `test-output-generator.sh` that randomly generates stdout and stderr messages to help test the terminal streamer:

```bash
# Make the script executable
chmod +x test-output-generator.sh

# Run it standalone to see its output
./test-output-generator.sh

# Use it with the terminal streamer
./terminal-streamer -cmd "./test-output-generator.sh"
```

This script runs until terminated with Ctrl+C and produces a mix of informational, warning, and error messages with varying time intervals.

## Future Enhancements

Potential improvements for future versions:

- Authentication and secure connections (HTTPS/WSS)
- Command input from the web interface (bidirectional communication)
- Session management for multiple commands
- Log file persistence
- Custom styling and themes

## License

MIT License
