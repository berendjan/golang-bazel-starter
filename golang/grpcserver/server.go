package main

import (
	"context"
	"log"
	"sync"

	configApi "github.com/berendjan/golang-bazel-starter/golang/config/api"
	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
)

type GrpcServerDependencies struct {
	doOnce            sync.Once
	doTwice           sync.Once
	accountRepository *repository.AccountDbRepository
	configApi         *configApi.ConfigurationApi[*repository.AccountDbRepository]
}

func (p *GrpcServerDependencies) GetAccountRepository() *repository.AccountDbRepository {
	p.doOnce.Do(func() {
		log.Println("Initializing AccountRepository singleton")
		// Get singleton database pool (automatically runs migrations)
		pool := db.MustNewPool(context.Background(), db.DefaultConfig(repository.DbName))

		// Create account repository
		p.accountRepository = repository.NewAccountRepository(pool)
	})
	return p.accountRepository
}

func (p *GrpcServerDependencies) GetAccountApi() *configApi.ConfigurationApi[*repository.AccountDbRepository] {
	p.doTwice.Do(func() {
		log.Println("Initializing AccountRepository singleton")

		// Create configuration API
		p.configApi = configApi.NewConfigurationApi(p)
	})
	return p.configApi
}
