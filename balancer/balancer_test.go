package balancer

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthCheck(t *testing.T) {
    healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer healthyServer.Close()

    unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer unhealthyServer.Close()

    servers := []*Server{
        {URL: healthyServer.URL, Healthy: false},
        {URL: unhealthyServer.URL, Healthy: false},
    }
    lb := NewLoadBalancer(servers)

    lb.HealthCheck()

    if !lb.servers[0].Healthy {
        t.Errorf("Expected server %s to be healthy", lb.servers[0].URL)
    }

    if lb.servers[1].Healthy {
        t.Errorf("Expected server %s to be unhealthy", lb.servers[1].URL)
    }
}

func TestNextServer(t *testing.T) {
    servers := []*Server{
        {URL: "http://server1", Healthy: true},
        {URL: "http://server2", Healthy: false},
        {URL: "http://server3", Healthy: true},
    }
    lb := NewLoadBalancer(servers)

    // First server should be healthy server1
    server, err := lb.NextServer()
    if err != nil {
        t.Fatalf("Unexpected error: %s", err)
    }
    if server.URL != "http://server1" {
        t.Errorf("Expected server1, got %s", server.URL)
    }

    // Second should skip server2 (unhealthy) and return server3
    server, err = lb.NextServer()
    if err != nil {
        t.Fatalf("Unexpected error: %s", err)
    }
    if server.URL != "http://server3" {
        t.Errorf("Expected server3, got %s", server.URL)
    }

    // Third should go back to server1
    server, err = lb.NextServer()
    if err != nil {
        t.Fatalf("Unexpected error: %s", err)
    }
    if server.URL != "http://server1" {
        t.Errorf("Expected server1, got %s", server.URL)
    }
}

func TestServeHTTP(t *testing.T) {
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello from backend"))
    }))
    defer backend.Close()

    servers := []*Server{
        {URL: backend.URL, Healthy: true},
    }
    lb := NewLoadBalancer(servers)

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rr := httptest.NewRecorder()

    lb.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Expected status code 200, got %d", rr.Code)
    }

    if rr.Body.String() != "Hello from backend" {
        t.Errorf("Expected response body 'Hello from backend', got %s", rr.Body.String())
    }
}
