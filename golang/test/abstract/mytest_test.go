package abstract_test

import "testing"

type ServerInterface interface {
	register(grpcPort, httpPort int) error
}

type ServerBase struct {
	ServerInterface
}

func (s *ServerBase) Launch(grpcPort, httpPort int) error {
	return s.register(grpcPort, httpPort)
}

type MyServer struct {
	*ServerBase
}

func (ms *MyServer) register(grpcPort, httpPort int) error {
	println("Registering server on gRPC port:", grpcPort, "and HTTP port:", httpPort)
	return nil
}

type MyServer2 struct {
	*ServerBase
}

func (ms *MyServer2) register(grpcPort, httpPort int) error {
	println("Registering server2 on gRPC port:", grpcPort, "and HTTP port:", httpPort)
	return nil
}

func testUsage(serverBase *ServerBase) {
	println("sdad")
}

func TestAbstract(t *testing.T) {
	myServer := &MyServer{
		ServerBase: &ServerBase{},
	}
	myServer.ServerBase.ServerInterface = myServer
	myServer.Launch(12, 23)

	testUsage(myServer.ServerBase)

	myServer2 := &MyServer2{
		ServerBase: &ServerBase{},
	}
	myServer2.ServerBase.ServerInterface = myServer2
	myServer2.Launch(34, 45)
}
