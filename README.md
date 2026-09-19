# Go Load Balancer

A simple Layer 7 Load Balancer written in Go. It forwards incoming HTTP traffic to multiple backend servers using a Round-Robin algorithm and performs background health checks to skip servers that go offline.

## Features

- **Round-Robin Routing**: Distributes user traffic evenly across a list of backend servers.
- **Health Checks**: Pings servers in the background using TCP. If a server goes offline, it is temporarily removed from the rotation.
- **Thread-Safe**: Uses Mutexes and atomic counters to safely handle multiple requests at the exact same time without crashing.
- **Environment Variables**: Reads the list of backend servers from an environment variable, making it easy to change settings without changing the code.

## Folder Structure

- `cmd/lb/main.go`: The main load balancer program.
- `cmd/backend/main.go`: A basic dummy web server used for local testing.
- `internal/loadbalancer/`: Holds the logic for tracking the servers and proxying the HTTP traffic.
- `internal/health/`: Holds the background ticker that pings the servers.

## How to Run Locally

1. **Start the testing servers** in separate terminal tabs:
   ```bash
   PORT=8081 go run cmd/backend/main.go
   PORT=8082 go run cmd/backend/main.go
   PORT=8083 go run cmd/backend/main.go
   ```

2. **Start the Load Balancer** in a new terminal tab:
   ```bash
   PORT=8080 go run cmd/lb/main.go
   ```
   *(If you don't provide a list of URLs, it will automatically look for ports 8081, 8082, and 8083).*

3. **Test the routing**:
   ```bash
   curl http://localhost:8080
   ```
   Run this command multiple times. You will see the Load Balancer passing the request to a different backend server each time. If you kill one of the dummy servers, the Load Balancer will automatically skip it.
