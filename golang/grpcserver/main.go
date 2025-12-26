package main

import (
	"context"
	"log"

	"github.com/berendjan/golang-bazel-starter/golang/config/api"
	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
	"github.com/berendjan/golang-bazel-starter/golang/grpcserver/messenger"
	"github.com/berendjan/golang-bazel-starter/golang/middleware/auth"
	"github.com/berendjan/golang-bazel-starter/golang/middleware/middleone"
	"github.com/berendjan/golang-bazel-starter/golang/middleware/middletwo"
)

type GrpcServer struct {
	*serverbase.ServerBase
	accountApi *api.ConfigurationApi
	messenger  *messenger.GrpcMessenger
}

func (g *GrpcServer) Register(sb *serverbase.ServerBuilder, grpcPort, httpPort int) error {
	// Register health endpoint
	sb.RegisterHealthEndpoint(httpPort)

	// Setup dependencies - register the AccountApi
	sb.RegisterService(grpcPort, httpPort, g.accountApi)
	return nil
}

func NewGrpcServer(messenger *messenger.GrpcMessenger) *GrpcServer {
	// Create API with messenger as the sendable interface
	accountApi := api.NewConfigurationApi(messenger)

	// Create gRPC server
	grpcServer := &GrpcServer{
		ServerBase: serverbase.NewServerBase(),
		accountApi: accountApi,
		messenger:  messenger,
	}
	grpcServer.ServerBase.ServerInterface = grpcServer

	return grpcServer
}

func createMessenger() *messenger.GrpcMessenger {
	// Initialize database pool
	pool := db.MustNewPool(context.Background(), db.DefaultConfig(repository.DbName))

	// Create repository
	accountRepo := repository.NewAccountRepository(pool)

	// Create auth middleware (Kratos public API)
	authMiddleware := auth.NewAuthMiddleware("http://kratos.app-namespace.svc.cluster.local:4433")

	// Create middleware chain
	middlewareOne := middleone.NewMiddleOne(authMiddleware)
	middlewareTwo := &middletwo.MiddleTwo{}

	// Create messenger with all dependencies
	grpcMessenger := messenger.NewGrpcMessenger(
		accountRepo,
		middlewareOne,
		middlewareTwo,
	)
	return grpcMessenger
}

func main() {
	// TLS certificate and key files
	certFile := "/mnt/server-certs/tls.crt"
	keyFile := "/mnt/server-certs/tls.key"
	caFile := "/mnt/server-certs/ca.crt"

	// Create and launch gRPC server with mTLS
	// Health port 27000 is non-TLS for Kubernetes probes
	grpcServer := NewGrpcServer(createMessenger()).
		WithTLS(certFile, keyFile).
		WithClientCA(caFile).
		WithHealthPort(27000)
	log.Println("Starting gRPC server with messenger")

	// Launch server
	grpcServer.LaunchWithDefaultPorts()
}
