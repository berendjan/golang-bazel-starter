package server_builder

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// GRPCServiceRegistrar registers a gRPC service with a gRPC server
type GRPCServiceRegistrar interface {
	RegisterGRPC(s grpc.ServiceRegistrar)
}

// HTTPGatewayRegistrar registers an HTTP gateway handler with a ServeMux
type HTTPGatewayRegistrar interface {
	RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error
}

// ServiceRegistrar combines both gRPC and HTTP gateway registration
type ServiceRegistrar interface {
	GRPCServiceRegistrar
	HTTPGatewayRegistrar
}

// ServerBuilder builds and manages gRPC and HTTP servers
type ServerBuilder struct {
	grpcPort    int
	httpPort    int
	gwMux       *runtime.ServeMux
	grpcServer  *grpc.Server
	grpcOpts    []grpc.ServerOption
	shutdownCtx context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// New creates a new ServerBuilder with default ports (gRPC: 25000, HTTP: 26000)
func New() *ServerBuilder {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerBuilder{
		grpcPort:    25000,
		httpPort:    26000,
		gwMux:       runtime.NewServeMux(),
		shutdownCtx: ctx,
		cancel:      cancel,
	}
}

// WithPorts sets custom ports for gRPC and HTTP servers
func (sb *ServerBuilder) WithPorts(grpcPort, httpPort int) *ServerBuilder {
	sb.grpcPort = grpcPort
	sb.httpPort = httpPort
	return sb
}

// WithGRPCOptions sets gRPC server options
func (sb *ServerBuilder) WithGRPCOptions(opts ...grpc.ServerOption) *ServerBuilder {
	sb.grpcOpts = append(sb.grpcOpts, opts...)
	return sb
}

// RegisterService registers a service that implements both gRPC and HTTP gateway
func (sb *ServerBuilder) RegisterService(service ServiceRegistrar) *ServerBuilder {
	// Ensure gRPC server is created
	if sb.grpcServer == nil {
		sb.grpcServer = grpc.NewServer(sb.grpcOpts...)
	}

	// Register gRPC service
	service.RegisterGRPC(sb.grpcServer)

	// Register HTTP gateway
	ctx := context.Background()
	if err := service.RegisterGateway(ctx, sb.gwMux); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	return sb
}

// RegisterGRPCService registers only a gRPC service
func (sb *ServerBuilder) RegisterGRPCService(service GRPCServiceRegistrar) *ServerBuilder {
	// Ensure gRPC server is created
	if sb.grpcServer == nil {
		sb.grpcServer = grpc.NewServer(sb.grpcOpts...)
	}

	service.RegisterGRPC(sb.grpcServer)
	return sb
}

// RegisterGateway registers only an HTTP gateway service
func (sb *ServerBuilder) RegisterGateway(service HTTPGatewayRegistrar) *ServerBuilder {
	ctx := context.Background()
	if err := service.RegisterGateway(ctx, sb.gwMux); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}
	return sb
}

// GRPCServer returns the underlying gRPC server for direct access
// Useful for registering additional services like reflection
func (sb *ServerBuilder) GRPCServer() *grpc.Server {
	if sb.grpcServer == nil {
		sb.grpcServer = grpc.NewServer(sb.grpcOpts...)
	}
	return sb.grpcServer
}

// Run starts all configured servers and blocks until shutdown
func (sb *ServerBuilder) Run() error {
	if sb.grpcServer == nil {
		return fmt.Errorf("no services registered")
	}

	// Setup graceful shutdown
	sb.setupGracefulShutdown()

	// Start gRPC server
	sb.wg.Add(1)
	go func() {
		defer sb.wg.Done()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", sb.grpcPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %d: %v", sb.grpcPort, err)
		}

		log.Printf("Starting gRPC server on port %d", sb.grpcPort)

		// Setup shutdown listener
		go func() {
			<-sb.shutdownCtx.Done()
			log.Printf("Shutting down gRPC server on port %d", sb.grpcPort)
			sb.grpcServer.GracefulStop()
		}()

		if err := sb.grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	// Start HTTP gateway server
	sb.wg.Add(1)
	go func() {
		defer sb.wg.Done()

		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", sb.httpPort),
			Handler: sb.gwMux,
		}

		log.Printf("Starting HTTP gateway on port %d", sb.httpPort)

		// Setup shutdown listener
		go func() {
			<-sb.shutdownCtx.Done()
			log.Printf("Shutting down HTTP server on port %d", sb.httpPort)
			if err := httpServer.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP server shutdown error: %v", err)
			}
		}()

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server stopped: %v", err)
		}
	}()

	// Wait for all servers to complete
	sb.wg.Wait()
	return nil
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func (sb *ServerBuilder) setupGracefulShutdown() {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal, shutting down all servers...")
		sb.cancel()
	}()
}

// Shutdown gracefully shuts down all servers
func (sb *ServerBuilder) Shutdown() {
	sb.cancel()
}
