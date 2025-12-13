package serverbase

import (
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc/reflection"
)

type ServerBase struct {
	ServerInterface
	shutdownCtx context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	tlsConfig   *tls.Config
}

func NewServerBase() *ServerBase {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerBase{
		shutdownCtx: ctx,
		cancel:      cancel,
	}
}

// WithTLS configures TLS for both gRPC and HTTP servers using certificate files
func (s *ServerBase) WithTLS(certFile, keyFile string) *ServerBase {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("TLS disabled: failed to load certificates from %s and %s: %v", certFile, keyFile, err)
		return s
	}

	s.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	log.Printf("TLS enabled using certificate: %s", certFile)
	return s
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

	// Start all gRPC servers
	log.Printf("Starting %d gRPC server(s) and %d HTTP server(s)...", len(sb.grpcServers), len(sb.httpServers))
	for grpcPort, grpcServer := range sb.grpcServers {
		s.wg.Add(1)
		go s.startGRPCServer(grpcPort, grpcServer)
	}

	// Start all HTTP servers
	for httpPort, httpMux := range sb.httpServers {
		s.wg.Add(1)
		go s.startHTTPServer(httpPort, httpMux)
	}

	// Wait for all servers to complete
	s.wg.Wait()
	return nil
}

// startGRPCServer starts a single gRPC server instance
func (s *ServerBase) startGRPCServer(grpcPort int, grpcServer *grpc.Server) {
	defer s.wg.Done()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %d: %v", grpcPort, err)
	}

	// Wrap listener with TLS if configured
	if s.tlsConfig != nil {
		lis = tls.NewListener(lis, s.tlsConfig)
		log.Printf("gRPC server listening on port %d (TLS)", grpcPort)
	} else {
		log.Printf("gRPC server listening on port %d", grpcPort)
	}

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
func (s *ServerBase) startHTTPServer(httpPort int, httpMux *runtime.ServeMux) {
	defer s.wg.Done()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: httpMux,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", httpPort))
	if err != nil {
		log.Fatalf("Failed to listen on HTTP port %d: %v", httpPort, err)
	}

	// Wrap listener with TLS if configured
	if s.tlsConfig != nil {
		lis = tls.NewListener(lis, s.tlsConfig)
		log.Printf("HTTPS server listening on port %d (TLS)", httpPort)
	} else {
		log.Printf("HTTP server listening on port %d", httpPort)
	}

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
