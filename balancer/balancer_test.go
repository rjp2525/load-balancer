package balancer

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
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

    lb := NewLoadBalancer([]string{healthyServer.URL, unhealthyServer.URL}, NewOpts())
    lb.HealthCheck()

    // Check that the healthy server is marked healthy
    if !lb.servers[0].Healthy {
        t.Errorf("Expected server %s to be healthy", lb.servers[0].URL)
    }

    // Check that the unhealthy server is marked unhealthy
    if lb.servers[1].Healthy {
        t.Errorf("Expected server %s to be unhealthy", lb.servers[1].URL)
    }
}

func TestNextServer(t *testing.T) {
    lb := NewLoadBalancer([]string{"http://server1", "http://server2", "http://server3"}, NewOpts())

    lb.servers[0].Healthy = true
    lb.servers[1].Healthy = false
    lb.servers[2].Healthy = true

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

func TestConnectionPool_Get(t *testing.T) {
    cp := NewConnectionPool(NewOpts().MaxConnections(2).Timeout(5 * time.Second))

    // First call should create a new client
    client := cp.Get("http://server1")
    if client == nil {
        t.Fatalf("Expected non-nil client")
    }

    // Push the client back to the pool
    err := cp.Push("http://server1", client)
    if err != nil {
        t.Fatalf("Unexpected error pushing client: %s", err)
    }

    // Second call should reuse the same client
    reusedClient := cp.Get("http://server1")
    if reusedClient != client {
        t.Errorf("Expected to reuse the same client, but got a different client")
    }
}

func TestConnectionPool_ExceedMaxConnections(t *testing.T) {
    cp := NewConnectionPool(NewOpts().MaxConnections(1).Timeout(5 * time.Second))

    // Create and push one client
    client := &http.Client{}
    err := cp.Push("http://server1", client)
    if err != nil {
        t.Fatalf("Unexpected error pushing client: %s", err)
    }

    // Try to push a second client, expect an error
    err = cp.Push("http://server1", &http.Client{})
    if err == nil {
        t.Errorf("Expected error when exceeding max connections, but got none")
    }
}

func TestServeHTTP(t *testing.T) {
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello from backend"))
    }))
    defer backend.Close()

    lb := NewLoadBalancer([]string{backend.URL}, NewOpts())

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
