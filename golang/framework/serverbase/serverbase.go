package serverbase

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
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServerBase struct {
	ServerInterface
	shutdownCtx context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	ready       chan struct{}
	readyOnce   sync.Once
}

func NewServerBase() *ServerBase {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerBase{
		shutdownCtx: ctx,
		cancel:      cancel,
		ready:       make(chan struct{}),
	}
}

// markReady signals that the server is ready to accept connections
func (s *ServerBase) markReady() {
	s.readyOnce.Do(func() {
		close(s.ready)
	})
}

// WaitUntilReady blocks until the server is ready or the timeout expires
func (s *ServerBase) WaitUntilReady(timeout time.Duration) error {
	select {
	case <-s.ready:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server did not start within %v", timeout)
	}
}

func (s *ServerBase) LaunchWithDefaultPorts() error {
	const grpcPort = 25000
	const httpPort = 26000
	return s.Launch(grpcPort, httpPort)
}

func (s *ServerBase) Launch(grpcPort, httpPort int) error {

	// Create server builder
	sb := NewServerBuilder()

	// Register services with both gRPC and HTTP gateway on specified ports
	s.Register(sb, grpcPort, httpPort)

	// Add reflection for debugging with grpcurl
	reflection.Register(sb.GRPCServer(grpcPort))

	// Run all servers
	if err := s.runServer(sb); err != nil {
		log.Fatalf("Failed to run servers: %v", err)
		return err
	}

	return nil
}

// Run starts all configured servers and blocks until shutdown
func (s *ServerBase) runServer(sb *ServerBuilder) error {
	if len(sb.grpcServers) == 0 && len(sb.httpServers) == 0 {
		return fmt.Errorf("no services registered")
	}

	// Setup graceful shutdown
	s.setupGracefulShutdown()

	// Channel to track when all servers have started listening
	serversReady := make(chan struct{})
	totalServers := len(sb.grpcServers) + len(sb.httpServers)
	startedCount := 0
	var startedMu sync.Mutex

	serverStarted := func() {
		startedMu.Lock()
		defer startedMu.Unlock()
		startedCount++
		if startedCount == totalServers {
			close(serversReady)
		}
	}

	// Start all gRPC servers
	for grpcPort, grpcServer := range sb.grpcServers {
		s.wg.Add(1)
		go s.startGRPCServer(grpcPort, grpcServer, serverStarted)
	}

	// Start all HTTP servers
	for httpPort, httpMux := range sb.httpServers {
		s.wg.Add(1)
		go s.startHTTPServer(httpPort, httpMux, serverStarted)
	}

	// Wait for all servers to start listening, then mark as ready
	go func() {
		<-serversReady
		s.markReady()
		log.Println("All servers are ready to accept connections")
	}()

	// Wait for all servers to complete
	s.wg.Wait()
	return nil
}

// startGRPCServer starts a single gRPC server instance
func (s *ServerBase) startGRPCServer(grpcPort int, grpcServer *grpc.Server, onReady func()) {
	defer s.wg.Done()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %d: %v", grpcPort, err)
	}

	// Notify that this server is ready
	onReady()

	// Setup shutdown listener
	go func() {
		<-s.shutdownCtx.Done()
		log.Printf("Shutting down gRPC server on port %d", grpcPort)
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Printf("gRPC server on port %d stopped: %v", grpcPort, err)
	}
}

// startHTTPServer starts a single HTTP gateway server instance
func (s *ServerBase) startHTTPServer(httpPort int, httpMux *runtime.ServeMux, onReady func()) {
	defer s.wg.Done()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: httpMux,
	}

	// Create listener first to ensure port is bound before marking ready
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		log.Fatalf("Failed to listen on HTTP port %d: %v", httpPort, err)
	}

	// Notify that this server is ready
	onReady()

	// Setup shutdown listener
	go func() {
		<-s.shutdownCtx.Done()
		log.Printf("Shutting down HTTP server on port %d", httpPort)
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server on port %d shutdown error: %v", httpPort, err)
		}
	}()

	if err := httpServer.Serve(lis); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server on port %d stopped: %v", httpPort, err)
	}
}

// setupGracefulShutdown sets up signal handling for graceful shutdown
func (s *ServerBase) setupGracefulShutdown() {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Received shutdown signal, shutting down all servers...")
		s.cancel()
	}()
}

// Shutdown gracefully shuts down all servers
func (s *ServerBase) Shutdown() {
	s.cancel()
}
