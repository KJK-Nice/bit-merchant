# Unified Server Scaffold Template

Generate the core unified server infrastructure in `internal/common/server/`. This replaces the standalone `RunHTTPServer` / `RunGRPCServer` functions with a composable `server.New(...).Run(ctx)` pattern that supports multiple transports with explicit shutdown ordering.

Created once per project. Individual transports (`WithWatermillRouter`) can be added later.

## Placeholders

- `{{module_common}}` — Go module path to `internal/common` (e.g., `github.com/example/myproject/internal/common`)

## File 1: `internal/common/server/server.go`

```go
package server

import (
	"context"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	components    map[string]component
	startOrder    []string
	shutdownSteps []ShutdownStep
}

type component struct {
	name  string
	start func(ctx context.Context) error
	stop  func(ctx context.Context) error
}

type Option func(*Server)

func New(opts ...Option) *Server {
	s := &Server{
		components: make(map[string]component),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) addComponent(name string, c component) {
	if _, exists := s.components[name]; exists {
		panic("duplicate component name: " + name)
	}
	s.components[name] = c
	s.startOrder = append(s.startOrder, name)
}

func (s *Server) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, len(s.components))
	for _, name := range s.startOrder {
		c := s.components[name]
		go func(c component) {
			logrus.WithField("component", c.name).Info("Starting")
			if err := c.start(ctx); err != nil {
				errCh <- err
			}
		}(c)
	}

	select {
	case <-ctx.Done():
		logrus.Info("Shutdown signal received")
	case err := <-errCh:
		logrus.WithError(err).Error("Component failed, initiating shutdown")
	}

	s.executeShutdown()
	return nil
}

func (s *Server) executeShutdown() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stopped := map[string]bool{}

	for _, step := range s.shutdownSteps {
		if step.fn != nil {
			logrus.Info("Running shutdown func")
			if err := step.fn(shutdownCtx); err != nil {
				logrus.WithError(err).Error("Shutdown func failed")
			}
			continue
		}

		var wg sync.WaitGroup
		for _, name := range step.componentNames {
			c, ok := s.components[name]
			if !ok {
				logrus.WithField("component", name).Warn("Unknown component in OnShutdown")
				continue
			}
			stopped[name] = true
			wg.Add(1)
			go func(c component) {
				defer wg.Done()
				logrus.WithField("component", c.name).Info("Stopping")
				if err := c.stop(shutdownCtx); err != nil {
					logrus.WithError(err).WithField("component", c.name).Error("Stop failed")
				}
			}(c)
		}
		wg.Wait()
	}

	// Safety net: stop any components not mentioned in OnShutdown
	var wg sync.WaitGroup
	for name, c := range s.components {
		if stopped[name] {
			continue
		}
		wg.Add(1)
		go func(c component) {
			defer wg.Done()
			logrus.WithField("component", c.name).Warn("Stopping (not in OnShutdown — add it)")
			if err := c.stop(shutdownCtx); err != nil {
				logrus.WithError(err).WithField("component", c.name).Error("Stop failed")
			}
		}(c)
	}
	wg.Wait()
}
```

## File 2: `internal/common/server/shutdown.go`

```go
package server

import "context"

// ShutdownStep is one step in the shutdown sequence.
type ShutdownStep struct {
	componentNames []string
	fn             func(ctx context.Context) error
}

// Stop creates a shutdown step that stops named components.
// Multiple names in one call = parallel shutdown within the step.
func Stop(names ...string) ShutdownStep {
	return ShutdownStep{componentNames: names}
}

// StopFunc creates a shutdown step that runs an arbitrary cleanup function.
func StopFunc(fn func()) ShutdownStep {
	return ShutdownStep{
		fn: func(ctx context.Context) error {
			fn()
			return nil
		},
	}
}

// StopFuncWithErr creates a shutdown step with error return.
func StopFuncWithErr(fn func(ctx context.Context) error) ShutdownStep {
	return ShutdownStep{fn: fn}
}

// OnShutdown declares the shutdown sequence.
// Steps execute top-to-bottom. Each step completes before the next starts.
// Components not mentioned are stopped last with a warning.
func OnShutdown(steps ...ShutdownStep) Option {
	return func(s *Server) {
		s.shutdownSteps = steps
	}
}
```

## File 3: `internal/common/server/http.go` (replace existing)

```go
package server

import (
	"context"
	"net/http"
	"os"

	"{{module_common}}/auth"
	"{{module_common}}/logs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

func WithHTTPHandler(name string, createHandler func(chi.Router) http.Handler) Option {
	return func(s *Server) {
		addr := ":" + os.Getenv("PORT")
		srv := &http.Server{Addr: addr}

		s.addComponent(name, component{
			name: name,
			start: func(ctx context.Context) error {
				apiRouter := chi.NewRouter()
				setMiddlewares(apiRouter)
				rootRouter := chi.NewRouter()
				rootRouter.Mount("/api", createHandler(apiRouter))
				srv.Handler = rootRouter

				logrus.WithField("addr", addr).Info("Starting HTTP server")
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					return err
				}
				return nil
			},
			stop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
	}
}

// setMiddlewares, addAuthMiddleware, addCorsMiddleware — same as existing
```

## File 4: `internal/common/server/grpc.go` (replace existing)

```go
package server

import (
	"context"
	"net"
	"os"

	"{{module_common}}/logs"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func WithGRPCServer(name string, registerServer func(*grpc.Server)) Option {
	return func(s *Server) {
		logrusEntry := logrus.NewEntry(logrus.StandardLogger())

		grpcSrv := grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_logrus.UnaryServerInterceptor(logrusEntry),
			),
			grpc_middleware.WithStreamServerChain(
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_logrus.StreamServerInterceptor(logrusEntry),
			),
		)
		registerServer(grpcSrv)

		port := os.Getenv("GRPC_PORT")
		if port == "" {
			port = "8080"
		}
		addr := ":" + port

		s.addComponent(name, component{
			name: name,
			start: func(ctx context.Context) error {
				lis, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}
				logrus.WithField("addr", addr).Info("Starting gRPC server")
				return grpcSrv.Serve(lis)
			},
			stop: func(ctx context.Context) error {
				grpcSrv.GracefulStop()
				return nil
			},
		})
	}
}
```

## Post-Creation Instructions

After creating the unified server:

1. Remove or replace the old `RunHTTPServer` / `RunGRPCServer` standalone functions
2. Update all `main.go` files to use `server.New(...).Run(ctx)` with `OnShutdown`
3. Add `/threedots scaffold watermill_router` to add Watermill support
4. Every component MUST appear in `OnShutdown` — the safety net logs warnings for forgotten ones
