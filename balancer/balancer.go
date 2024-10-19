package balancer

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
	"sync"
)

type Server struct {
    URL     string
    Healthy bool
}

type LoadBalancer struct {
    servers []*Server
	mu      sync.Mutex
    idx     int
	cp      *ConnectionPool
}

func NewLoadBalancer(urls []string, opts *Opts) *LoadBalancer {
    servers := make([]*Server, len(urls))
    for i, url := range urls {
        servers[i] = &Server{
            URL:     url,
            Healthy: true,
        }
    }

    return &LoadBalancer{
        servers: servers,
        cp:      NewConnectionPool(opts),
    }
}

func (lb *LoadBalancer) HealthCheck() {
    for _, server := range lb.servers {
		// Laravel has a /up page generally with a 200 status code
        res, err := http.Get(server.URL + "/up")
        if err != nil || res.StatusCode != http.StatusOK {
            server.Healthy = false
            fmt.Printf("Server [%s] is down\n", server.URL)
        } else {
            server.Healthy = true
            fmt.Printf("Server [%s] is up\n", server.URL)
        }
    }
}

func (lb *LoadBalancer) RunHealthCheck() {
    ticker := time.NewTicker(10 * time.Second)

    go func() {
        for {
            <-ticker.C
            lb.HealthCheck()
        }
    }()
}

func (lb *LoadBalancer) NextServer() (*Server, error) {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    for i := 0; i < len(lb.servers); i++ {
        server := lb.servers[lb.idx]
        lb.idx = (lb.idx + 1) % len(lb.servers)

        if server.Healthy {
            return server, nil
        }
    }

    return nil, errors.New("no healthy servers available")
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    server, err := lb.NextServer()
    if err != nil {
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
    }

    client := lb.cp.Get(server.URL)
    res, err := lb.ForwardRequest(client, server, r.RequestURI)
    if err != nil {
        http.Error(w, "Error forwarding request", http.StatusInternalServerError)
        return
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        http.Error(w, "Error reading response", http.StatusInternalServerError)
        return
    }

    _, err = w.Write(body)
    if err != nil {
        http.Error(w, "Error writing response", http.StatusInternalServerError)
    }

    lb.cp.Push(server.URL, client) // Return the connection to the pool
}

func (lb *LoadBalancer) ForwardRequest(client *http.Client, server *Server, uri string) (*http.Response, error) {
    fullUrl := server.URL + uri
    fmt.Printf("Forwarding request to: %s\n", fullUrl)

    res, err := client.Get(fullUrl)
    if err != nil {
        return nil, err
    }

    return res, nil
}
