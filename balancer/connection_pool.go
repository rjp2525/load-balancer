package balancer

import (
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Opts struct {
    maxConnections int
    timeout        time.Duration
}

type ConnectionPool struct {
    *Opts
    clients map[string][]*http.Client
    mu      sync.Mutex
}

// Creates default connection pool options with 10 max connections and a 5-second timeout.
func NewOpts() *Opts {
    return &Opts{
        maxConnections: 10,
        timeout:        5 * time.Second,
    }
}

// Sets the maximum number of connections for a server
func (opts *Opts) MaxConnections(maxConnections int) *Opts {
    opts.maxConnections = maxConnections
    return opts
}

// Sets the timeout for HTTP connections
func (opts *Opts) Timeout(timeout time.Duration) *Opts {
    opts.timeout = timeout
    return opts
}

// Initializes a new connection pool with the provided options
func NewConnectionPool(opts *Opts) *ConnectionPool {
    return &ConnectionPool{
        Opts:    opts,
        clients: make(map[string][]*http.Client),
    }
}

// Retrieves a connection from the pool, or creates a new one if no connection is available
func (cp *ConnectionPool) Get(server string) *http.Client {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if clients, ok := cp.clients[server]; ok && len(clients) > 0 {
        client := clients[len(clients)-1]
        cp.clients[server] = clients[:len(clients)-1]
        return client
    }

    return &http.Client{
        Timeout: cp.timeout,
    }
}

// Returns a connection to the pool
func (cp *ConnectionPool) Push(server string, client *http.Client) error {
    cp.mu.Lock()
    defer cp.mu.Unlock()

    if len(cp.clients[server]) >= cp.maxConnections {
        return fmt.Errorf("connection pool limit exceeded for server '%s'", server)
    }

    cp.clients[server] = append(cp.clients[server], client)
    return nil
}
