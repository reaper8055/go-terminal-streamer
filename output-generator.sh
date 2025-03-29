#!/usr/bin/env bash

# test-output-generator.sh
# Generates random stdout and stderr output until terminated with Ctrl+C

# Ensure cleanup on script exit
cleanup() {
  echo "Cleaning up and exiting..." >&2
  exit 0
}

# Set up trap for clean exit on SIGINT (Ctrl+C)
trap cleanup SIGINT SIGTERM

# Define absolute path for logs
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
LOG_FILE="${SCRIPT_DIR}/output_generator.log"

# Log levels with corresponding functions
log_info() {
  echo "[INFO] $(date +"%Y-%m-%d %H:%M:%S") - $1"
}

log_warn() {
  echo "[WARN] $(date +"%Y-%m-%d %H:%M:%S") - $1" >&2
}

log_error() {
  echo "[ERROR] $(date +"%Y-%m-%d %H:%M:%S") - $1" >&2
}

# Function to generate random build-like output
generate_build_output() {
  local modules=("core" "api" "services" "utils" "models" "database" "frontend" "auth" "metrics" "logging")
  local status=("Compiling" "Building" "Testing" "Analyzing" "Optimizing" "Linking" "Packaging")
  local result=("successful" "completed" "done" "finished")
  local errors=("syntax error" "undefined reference" "dependency not found" "type mismatch" "failed assertion")
  
  local module="${modules[$RANDOM % ${#modules[@]}]}"
  local action="${status[$RANDOM % ${#status[@]}]}"
  
  # Determine if this should be success (stdout) or error (stderr)
  if (( RANDOM % 10 < 8 )); then
    # 80% chance of success message to stdout
    local outcome="${result[$RANDOM % ${#result[@]}]}"
    log_info "${action} module '${module}' ${outcome} in $((RANDOM % 5 + 1))s"
  else
    # 20% chance of error message to stderr
    local error="${errors[$RANDOM % ${#errors[@]}]}"
    log_error "${action} module '${module}' failed: ${error} at line $((RANDOM % 1000 + 1))"
  fi
}

# Function to generate random system statistics
generate_system_stats() {
  local cpu_usage=$((RANDOM % 100))
  local memory_usage=$((RANDOM % 8192 + 1024))
  local disk_io=$((RANDOM % 1000))
  
  if (( cpu_usage > 80 )); then
    log_warn "High CPU usage detected: ${cpu_usage}%"
  else
    log_info "System stats - CPU: ${cpu_usage}%, Memory: ${memory_usage}MB, Disk I/O: ${disk_io}KB/s"
  fi
}

# Function to simulate progress with progress bar
show_progress() {
  local percent=$((RANDOM % 101))
  local width=50
  local completed=$((width * percent / 100))
  local remaining=$((width - completed))
  
  local progress="["
  for ((i=0; i<completed; i++)); do
    progress+="#"
  done
  
  for ((i=0; i<remaining; i++)); do
    progress+="."
  done
  
  progress+="] ${percent}%"
  log_info "${progress}"
}

# Main execution loop
echo "Starting output generator. Press Ctrl+C to stop."
echo "Output will be mixed between stdout and stderr randomly."
echo "========================================================="

COUNTER=0
while true; do
  # Increment counter
  COUNTER=$((COUNTER + 1))
  
  # Every 10th iteration, display a milestone
  if (( COUNTER % 10 == 0 )); then
    log_info "===== Milestone reached: ${COUNTER} iterations ====="
  fi
  
  # Generate different types of output randomly
  case $((RANDOM % 5)) in
    0|1|2) # 60% chance for build output
      generate_build_output
      ;;
    3) # 20% chance for system stats
      generate_system_stats
      ;;
    4) # 20% chance for progress bar
      show_progress
      ;;
  esac
  
  # Random sleep between 0.1 and 1.5 seconds
  sleep "0.$(( RANDOM % 15 + 1 ))"
done
