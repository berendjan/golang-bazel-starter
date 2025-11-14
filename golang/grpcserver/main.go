package main

import (
	"github.com/berendjan/golang-bazel-starter/golang/config/api"
	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
)

var serverDeps *GrpcServerDependencies

type DepsProvider interface {
	repository.AccountRepositoryProvider[*repository.AccountDbRepository]
	api.AccountApiProvider[*repository.AccountDbRepository]
}

type GrpcServer struct {
	*serverbase.ServerBase
	depsProvider DepsProvider
}

func (g *GrpcServer) Register(sb *serverbase.ServerBuilder, grpcPort, httpPort int) error {
	// Setup dependencies
	sb.RegisterService(grpcPort, httpPort, g.depsProvider.GetAccountApi())
	return nil
}

func NewGrpcServer(depsProvider DepsProvider) *GrpcServer {

	// Create gRPC server
	grpcServer := &GrpcServer{
		ServerBase:   serverbase.NewServerBase(),
		depsProvider: depsProvider,
	}
	grpcServer.ServerBase.ServerInterface = grpcServer

	return grpcServer
}

func main() {
	// Dependencies
	serverDeps = &GrpcServerDependencies{}

	// grpc server
	grpcServer := NewGrpcServer(serverDeps)

	// Launch server
	grpcServer.LaunchWithDefaultPorts()
}
