package serverbase

type ServerInterface interface {
	Register(sb *ServerBuilder, grpcPort, httpPort int) error
}
