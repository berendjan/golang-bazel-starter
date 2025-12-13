package test

import (
	"path/filepath"
	"sync"

	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	configRepository "github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
	grpcserver "github.com/berendjan/golang-bazel-starter/golang/grpcserver"
	"github.com/berendjan/golang-bazel-starter/golang/grpcserver/messenger"
	"github.com/berendjan/golang-bazel-starter/golang/middleware/middletwo"
)

type database string

const (
	configDb database = database(configRepository.DbName)
)

type server string

const (
	grpcServer server = "grpc-server"
)

var (
	// Use dbmate migrations from db/config/migrations
	// Path is relative to runfiles/_main/golang/test, so go up to _main first
	ConfigDb DatabaseConfig = DatabaseConfig{
		database:      configDb,
		migrationsDir: filepath.Join("..", "..", "db", "config", "migrations"),
	}
)

var (
	GrpcServer ServerConfig = ServerConfig{server: grpcServer, provider: func(tcp *TestContextProvider) *serverbase.ServerBase {
		return grpcserver.NewGrpcServer(tcp.createMessenger()).ServerBase
	}}
)

type TestContextProvider struct {
	messengerOnce sync.Once
	messenger     *messenger.GrpcMessenger
	dbContexts    map[database]*TestDBContext
}

func NewTestContextProvider(dbContexts map[database]*TestDBContext) *TestContextProvider {
	return &TestContextProvider{
		dbContexts: dbContexts,
	}
}

func (tcp *TestContextProvider) createMessenger() *messenger.GrpcMessenger {
	tcp.messengerOnce.Do(func() {

		// Get database pool
		pool := tcp.dbContexts[configDb].client

		// Create repository
		accountRepo := repository.NewAccountRepository(pool)

		// Interchangable test middleware
		middlewareOne := &TestMiddleOne{}
		middlewareTwo := &middletwo.MiddleTwo{}

		// Create messenger with all dependencies
		tcp.messenger = messenger.NewGrpcMessenger(
			accountRepo,
			middlewareOne,
			middlewareTwo,
		)
	})

	return tcp.messenger
}
