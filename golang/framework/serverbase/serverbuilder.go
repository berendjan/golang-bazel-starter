package serverbase

import (
	"context"
	"log"

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

// ServerBuilder builds and manages multiple gRPC and HTTP servers
type ServerBuilder struct {
	grpcServers map[int]*grpc.Server        // map of grpcPort -> grpc.Server
	httpServers map[int]*runtime.ServeMux   // map of httpPort -> ServeMux
	grpcOpts    map[int][]grpc.ServerOption // map of grpcPort -> server options
}

// New creates a new ServerBuilder
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		grpcServers: make(map[int]*grpc.Server),
		httpServers: make(map[int]*runtime.ServeMux),
		grpcOpts:    make(map[int][]grpc.ServerOption),
	}
}

// WithGRPCOptions sets gRPC server options for a specific port
func (sb *ServerBuilder) WithGRPCOptions(grpcPort int, opts ...grpc.ServerOption) *ServerBuilder {
	sb.grpcOpts[grpcPort] = append(sb.grpcOpts[grpcPort], opts...)
	return sb
}

// RegisterService registers a service on specified ports
// Creates gRPC and HTTP servers on the given ports if they don't exist
func (sb *ServerBuilder) RegisterService(grpcPort, httpPort int, service ServiceRegistrar) *ServerBuilder {
	log.Printf("RegisterService called with grpcPort=%d httpPort=%d service=%T", grpcPort, httpPort, service)

	// Get or create gRPC server for this port
	grpcServer, exists := sb.grpcServers[grpcPort]
	if !exists {
		opts := sb.grpcOpts[grpcPort] // Get port-specific options
		grpcServer = grpc.NewServer(opts...)
		sb.grpcServers[grpcPort] = grpcServer
	}

	// Get or create HTTP ServeMux for this port
	httpMux, exists := sb.httpServers[httpPort]
	if !exists {
		httpMux = runtime.NewServeMux()
		sb.httpServers[httpPort] = httpMux
	}

	// Register gRPC service
	service.RegisterGRPC(grpcServer)

	// Register HTTP gateway
	ctx := context.Background()
	if err := service.RegisterGateway(ctx, httpMux); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	return sb
}

// RegisterGRPCService registers only a gRPC service on specified port
func (sb *ServerBuilder) RegisterGRPCService(grpcPort int, service GRPCServiceRegistrar) *ServerBuilder {
	// Get or create gRPC server for this port
	grpcServer, exists := sb.grpcServers[grpcPort]
	if !exists {
		opts := sb.grpcOpts[grpcPort] // Get port-specific options
		grpcServer = grpc.NewServer(opts...)
		sb.grpcServers[grpcPort] = grpcServer
	}

	service.RegisterGRPC(grpcServer)
	return sb
}

// RegisterGateway registers only an HTTP gateway service on specified port
func (sb *ServerBuilder) RegisterGateway(httpPort int, service HTTPGatewayRegistrar) *ServerBuilder {
	// Get or create HTTP ServeMux for this port
	httpMux, exists := sb.httpServers[httpPort]
	if !exists {
		httpMux = runtime.NewServeMux()
		sb.httpServers[httpPort] = httpMux
	}

	ctx := context.Background()
	if err := service.RegisterGateway(ctx, httpMux); err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}
	return sb
}

// GRPCServer returns the underlying gRPC server for a specific port
// Useful for registering additional services like reflection
// Returns nil if no server exists on that port
func (sb *ServerBuilder) GRPCServer(grpcPort int) *grpc.Server {
	return sb.grpcServers[grpcPort]
}
