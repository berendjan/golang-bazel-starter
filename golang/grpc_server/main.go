package main

import (
	"log"

	"github.com/berendjan/golang-bazel-starter/golang/config/server"
	config "github.com/berendjan/golang-bazel-starter/golang/config/db"
	"github.com/berendjan/golang-bazel-starter/golang/framework/server_builder"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Get singleton database pool (automatically runs migrations)
	pool := config.GetPool()

	// Create account repository
	accountRepo := config.NewAccountRepository(pool)

	// Create server builder (defaults to gRPC: 25000, HTTP: 26000)
	builder := server_builder.New()

	// Register services with both gRPC and HTTP gateway
	builder.RegisterService(server.NewConfigurationServer(accountRepo))

	// Add reflection for debugging with grpcurl
	reflection.Register(builder.GRPCServer())

	// Add more services if needed:
	// builder.RegisterService(NewAnotherServer(pool))

	// Run all servers
	if err := builder.Run(); err != nil {
		log.Fatalf("Failed to run servers: %v", err)
	}
}
