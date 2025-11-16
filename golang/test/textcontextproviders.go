package test

import (
	"sync"

	configDbMigrations "github.com/berendjan/golang-bazel-starter/golang/config/db"
	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	configRepository "github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
	grpcserver "github.com/berendjan/golang-bazel-starter/golang/grpcserver"
	"github.com/berendjan/golang-bazel-starter/golang/grpcserver/messenger"
	"github.com/berendjan/golang-bazel-starter/golang/middleware/middletwo"
)

type database string

const (
	config database = database(configRepository.DbName)
)

type server string

const (
	grpc server = "grpc-server"
)

var (
	ConfigDb DatabaseConfig = DatabaseConfig{database: config, migrations: configDbMigrations.MigrationsFS}
)

var (
	GrpcServer ServerConfig = ServerConfig{server: grpc, provider: func(tcp *TestContextProvider) *serverbase.ServerBase {
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
		pool := tcp.dbContexts[config].client

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
