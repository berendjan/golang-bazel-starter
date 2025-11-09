package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	gw "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1/gateway"
)

var (
	port = flag.Int("port", 50051, "The server port (serves both gRPC and HTTP)")
)

func main() {
	flag.Parse()

	// Create the Configuration service server instance
	configServer := NewConfigurationServer()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register the Configuration service
	gw.RegisterConfigurationServer(grpcServer, configServer)

	// Register reflection service for grpcurl/grpc_cli
	reflection.Register(grpcServer)

	// Create gRPC-Gateway mux
	ctx := context.Background()
	gwMux := runtime.NewServeMux()

	// Register the gateway handler directly with the server instance (no network call)
	if err := gw.RegisterConfigurationHandlerServer(ctx, gwMux, configServer); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// Create a multiplexer that can handle both gRPC and HTTP requests
	mixedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
			grpcServer.ServeHTTP(w, r)
		} else {
			gwMux.ServeHTTP(w, r)
		}
	})

	// Wrap with h2c (HTTP/2 cleartext) to support both HTTP/1.1 and HTTP/2
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: h2c.NewHandler(mixedHandler, &http2.Server{}),
	}

	// Setup graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on port %d (gRPC and HTTP)", *port)
	log.Printf("  - gRPC: use grpcurl or gRPC clients")
	log.Printf("  - HTTP: curl http://localhost:%d/v1/...", *port)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to serve: %v", err)
	}
}
