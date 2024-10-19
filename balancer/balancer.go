package balancer

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Server struct {
    URL     string
    Healthy bool
}

type LoadBalancer struct {
    servers []*Server
    idx     int
}

func NewLoadBalancer(servers []*Server) *LoadBalancer {
    return &LoadBalancer{
        servers: servers,
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

    res, err := lb.ForwardRequest(server, r.RequestURI)
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

    w.Write(body)
}

func (lb *LoadBalancer) ForwardRequest(server *Server, uri string) (*http.Response, error) {
    fullUrl := server.URL + uri
    fmt.Printf("Forwarding request to: %s\n", fullUrl)

    res, err := http.Get(fullUrl)
    if err != nil {
        return nil, err
    }

    return res, nil
}
