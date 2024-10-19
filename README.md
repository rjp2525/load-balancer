## Go Load Balancer

A simple load balancer implemented in Go with round-robin request forwarding, health checks and connection pooling. The load balancer forwards requests to backend servers, performs periodic health checks and reuses HTTP connections to optimize performance.

## Features

- **Round-Robin Load Balancing**: Distributes incoming requests across multiple backend servers in a round-robin fashion
- **Health Checks**: Periodically checks backend servers for availability and skips unhealthy servers
- **Connection Pooling**: Reuses existing HTTP connections to reduce the overhead of creating new connections
- **Concurrent Request Handling**: Uses Go's `http.ListenAndServe` to handle requests concurrently

## Requirements

- **Go** 1.23+ is required to build and run this project.

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/rjp2525/load-balancer.git && cd load-balancer
```

### Install Dependencies

```bash
go mod tidy
```

## Running the Load Balancer

1. You'll need at least two backend servers listening on different ports to act as targets for the load balancer. The balancer is setup to check `/up` for Laravel projects, but you can create a fake backend server to listen on for testing purposes:
  ```go
  package main

  import (
      "fmt"
      "net/http"
  )

  func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
      fmt.Fprintln(w, "OK")
  }

  func main() {
      http.HandleFunc("/up", healthCheckHandler)

      fmt.Println("Starting server on :8000")
      if err := http.ListenAndServe(":8000", nil); err != nil {
          fmt.Printf("Error starting server: %s\n", err)
      }
  }
  ```

  Run this server on different ports (e.g. `8000`, `8001`) to simulate the backend servers.

2. Once your backend servers are running, run the load balancer:
  ```bash
  go run main.go
  ```

## Example

Once the load balancer is running, you can forward requests to the backend servers via the load balancer:
```bash
curl http://127.0.0.1:8080/
```

The load balancer will forward the request to one of the healthy backend servers in a round-robin fashion, and you will see the response from the backend server.

## Configuration (Optional)

You can configure the backend servers and port for the load balancer using a config.json file, for example:

```json
{
    "port": 8080,
    "servers": [
        "http://127.0.0.1:8000",
        "http://127.0.0.1:8001"
    ]
}
```

To use this configuration, ensure the file is located in the root of the project and the load balancer will read from it automatically.

## Testing

Unit tests are available for the load balancer including health checks, connection pooling and round-robin request distribution.

Run the tests using:
```bash
go test ./balancer
```

You should see output indicating whether the tests passed or failed.

## License

This project is licensed under the [MIT License](LICENSE.md).
