// Package proxy implements the local HTTP proxy server for AI API requests.
package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/royisme/bobamixer/internal/logging"
)

const (
	// DefaultAddr is the default proxy server address
	DefaultAddr = "127.0.0.1:7777"

	// ReadTimeout is the timeout for reading request
	ReadTimeout = 30 * time.Second

	// WriteTimeout is the timeout for writing response
	WriteTimeout = 30 * time.Second

	// IdleTimeout is the timeout for idle connections
	IdleTimeout = 120 * time.Second
)

// Server represents the local proxy server
type Server struct {
	addr       string
	httpServer *http.Server
	handler    *Handler
	mu         sync.Mutex
	running    bool
	listener   net.Listener
}

// NewServer creates a new proxy server
func NewServer(addr, dbPath string) (*Server, error) {
	if addr == "" {
		addr = DefaultAddr
	}

	handler, err := NewHandler(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	s := &Server{
		addr:    addr,
		handler: handler,
	}

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	return s, nil
}

// Start starts the proxy server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	// Use ListenConfig with context for proper context support
	ctx := context.Background()
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	s.listener = listener
	s.running = true

	logging.Info("Proxy server starting", logging.String("addr", s.addr))

	// Start serving in background
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logging.Error("Proxy server error", logging.String("error", err.Error()))
		}
	}()

	return nil
}

// Stop gracefully stops the proxy server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	logging.Info("Proxy server stopping...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.running = false
	logging.Info("Proxy server stopped")

	return nil
}

// IsRunning returns whether the server is currently running
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Addr returns the server address
func (s *Server) Addr() string {
	return s.addr
}

// Stats returns current proxy statistics
func (s *Server) Stats() *Stats {
	return s.handler.Stats()
}
