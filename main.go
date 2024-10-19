package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/rjp2525/load-balancer/balancer"
)

func main() {
    lb := balancer.NewLoadBalancer(
        []string{"http://127.0.0.1:8000", "http://127.0.0.1:80"},
        balancer.NewOpts().
            Timeout(10 * time.Second).
            MaxConnections(100),
    )

    lb.RunHealthCheck()

    fmt.Println("Starting load balancer on :8080")
    err := http.ListenAndServe(":8080", lb)
    if err != nil {
        log.Fatalf("Error starting server: %s", err)
    }
}
