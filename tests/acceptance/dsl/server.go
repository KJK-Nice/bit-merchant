package dsl

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// TestServer wraps an HTTP server for Rod-based testing
type TestServer struct {
	server   *http.Server
	echo     *echo.Echo
	port     int
	baseURL  string
	listener net.Listener
}

// StartTestServer starts an HTTP server for the test application
func StartTestServer(t *testing.T, app *TestApplication) *TestServer {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	server := &http.Server{
		Handler: app.GetEcho(),
	}

	testServer := &TestServer{
		server:   server,
		echo:     app.echo,
		port:     port,
		baseURL:  baseURL,
		listener: listener,
	}

	// Start server in goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/kitchen")
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	return testServer
}

// BaseURL returns the base URL of the test server
func (ts *TestServer) BaseURL() string {
	return ts.baseURL
}

// Port returns the port the server is listening on
func (ts *TestServer) Port() int {
	return ts.port
}

// Stop stops the test server
func (ts *TestServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ts.server.Shutdown(ctx)
}

// StopWithContext stops the test server with a custom context
func (ts *TestServer) StopWithContext(ctx context.Context) error {
	return ts.server.Shutdown(ctx)
}

