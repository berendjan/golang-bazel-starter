package test

import (
	"sync"

	configApi "github.com/berendjan/golang-bazel-starter/golang/config/api"
	configDbMigrations "github.com/berendjan/golang-bazel-starter/golang/config/db"
	configRepository "github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
	grpcserver "github.com/berendjan/golang-bazel-starter/golang/grpcserver"
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
	GrpcServer ServerConfig = ServerConfig{server: grpc, provider: func(tcp *TestContextProvider) *serverbase.ServerBase { return grpcserver.NewGrpcServer(tcp).ServerBase }}
)

type TestContextProvider struct {
	accountRepositoryOnce sync.Once
	configApiOnce         sync.Once
	accountRepository     *configRepository.AccountDbRepository
	configApi             *configApi.ConfigurationApi[*configRepository.AccountDbRepository]
	dbContexts            map[database]*TestDBContext
}

func NewTestContextProvider(dbContexts map[database]*TestDBContext) *TestContextProvider {
	return &TestContextProvider{
		dbContexts: dbContexts,
	}
}

func (p *TestContextProvider) GetAccountRepository() *configRepository.AccountDbRepository {
	p.accountRepositoryOnce.Do(func() {
		// Get singleton database pool (automatically runs migrations)
		// The database name already includes the testID suffix (e.g., "config_abc123")
		// because it was created that way in createDatabase()
		dbPool := p.dbContexts[config].client

		// Create account repository
		p.accountRepository = configRepository.NewAccountRepository(dbPool)
	})
	return p.accountRepository
}

func (p *TestContextProvider) GetAccountApi() *configApi.ConfigurationApi[*configRepository.AccountDbRepository] {
	p.configApiOnce.Do(func() {
		// Create configuration API
		p.configApi = configApi.NewConfigurationApi(p)
	})
	return p.configApi
}
