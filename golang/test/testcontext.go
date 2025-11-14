// Package test provides utilities for integration testing with isolated database contexts.
//
// TestContext creates isolated PostgreSQL databases for each test, automatically running
// migrations and providing a clean database state. A single shared PostgreSQL container
// is reused across all tests for efficiency.
//
// Key features:
// - Single shared container across all tests (created once via sync.Once)
// - Isolated databases per test (each test gets unique database(s))
// - Automatic migration execution via golang-migrate
// - Support for multiple databases and servers per test
// - Builder pattern for flexible configuration
//
// Example usage with builder:
//
//	func TestMyFeature(t *testing.T) {
//	    ctx := context.Background()
//	    tc, err := NewTestContextBuilder().
//	        WithDatabase("main", configDb.MigrationsFS).
//	        WithDatabase("analytics", analyticsDb.MigrationsFS).
//	        Build(ctx)
//	    if err != nil {
//	        t.Fatalf("Failed to create test context: %v", err)
//	    }
//	    defer tc.CleanUp(ctx)
//
//	    // Use named database clients
//	    mainDB := tc.Database("main")
//	    analyticsDB := tc.Database("analytics")
//	}
package test

import (
	"context"
	"embed"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
	frameworkDB "github.com/berendjan/golang-bazel-starter/golang/framework/db"
	"github.com/berendjan/golang-bazel-starter/golang/framework/serverbase"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	initSQL = `
CREATE TABLE IF NOT EXISTS test_databases (
    dbname TEXT PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`
)

var (
	// Singleton container and connection info for test contexts
	sharedContainer     testcontainers.Container
	sharedContainerHost string
	sharedContainerPort int
	sharedContainerOnce sync.Once
	sharedContainerErr  error
)

// TestContext provides isolated database and server instances for testing
type TestContext struct {
	testID              string
	container           testcontainers.Container
	databases           map[database]*TestDBContext
	servers             map[server]*TestServerContext
	postgresClient      *db.DBPool
	testContextProvider *TestContextProvider
}

// TestDBContext manages a test database connection
type TestDBContext struct {
	client     *db.DBPool
	dbName     string
	dbURL      string
	migrations embed.FS
	testConfig *DatabaseConfig
}

// TestServerContext manages a test server instance
type TestServerContext struct {
	grpcPort   int
	httpPort   int
	server     *serverbase.ServerBase
	serverDone chan struct{}
}

// Shutdown gracefully shuts down the test server and waits for it to complete
func (s *TestServerContext) Shutdown() {
	if s.server != nil {
		s.server.Shutdown()
		// Wait for server to finish shutting down
		if s.serverDone != nil {
			<-s.serverDone
		}
	}
}

// DatabaseConfig holds configuration for a database to be created
type DatabaseConfig struct {
	database
	migrations embed.FS
}

// ServerConfig holds configuration for a server to be created
type ServerConfig struct {
	server
	provider func(*TestContextProvider) *serverbase.ServerBase
}

// TestContextBuilder builds a TestContext with multiple databases and servers
type TestContextBuilder struct {
	databases []DatabaseConfig
	servers   []ServerConfig
}

// NewTestContextBuilder creates a new TestContextBuilder
func NewTestContextBuilder() *TestContextBuilder {
	return &TestContextBuilder{
		databases: []DatabaseConfig{},
		servers:   []ServerConfig{},
	}
}

// WithDatabase adds a database configuration to the builder
func (b *TestContextBuilder) WithDatabase(databaseConfig DatabaseConfig) *TestContextBuilder {
	b.databases = append(b.databases, databaseConfig)
	return b
}

// WithServer adds a server configuration to the builder
func (b *TestContextBuilder) WithServer(serverConfig ServerConfig) *TestContextBuilder {
	b.servers = append(b.servers, serverConfig)
	return b
}

// Build creates the TestContext with all configured databases and servers
func (b *TestContextBuilder) Build(ctx context.Context) (*TestContext, error) {
	testID := uuid.New().String()[:8]

	// Get or create the shared container
	pgContainer, host, port, err := getOrCreateContainer(ctx)
	if err != nil {
		return nil, err
	}

	// Connect to postgres database for this test context
	testConfig := &db.Config{
		Host:              host,
		Port:              port,
		User:              "postgres",
		Password:          "postgres",
		Database:          "postgres",
		SSLMode:           "disable",
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}

	postgresClient, err := db.NewPool(ctx, testConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	// Create test_databases table if it doesn't exist (idempotent)
	_, err = postgresClient.Exec(ctx, initSQL)
	if err != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("failed to create test_databases table: %w", err)
	}

	// Create all configured databases
	databases := make(map[database]*TestDBContext)
	for _, dbConfig := range b.databases {
		dbCtx, err := createDatabase(ctx, testID, dbConfig, host, port, postgresClient)
		if err != nil {
			// Clean up any created databases before returning error
			for _, db := range databases {
				if db.client != nil {
					db.client.Close()
				}
			}
			postgresClient.Close()
			return nil, fmt.Errorf("failed to create database '%s': %w", dbConfig.database, err)
		}
		databases[dbConfig.database] = dbCtx
	}

	// get Test Context Depedency Provider
	dependencyProvider := NewTestContextProvider(databases)

	// Create all configured servers
	servers := make(map[server]*TestServerContext)
	for _, srvConfig := range b.servers {
		srvCtx, err := createServer(ctx, srvConfig, dependencyProvider)
		if err != nil {
			// Clean up before returning error
			for _, db := range databases {
				db.client.Close()
			}
			postgresClient.Close()
			return nil, fmt.Errorf("failed to create server '%s': %w", srvConfig.server, err)
		}
		servers[srvConfig.server] = srvCtx
	}

	return &TestContext{
		testID:              testID,
		container:           pgContainer,
		databases:           databases,
		servers:             servers,
		postgresClient:      postgresClient,
		testContextProvider: dependencyProvider,
	}, nil
}

// getOrCreateContainer returns the singleton container, creating it if necessary
func getOrCreateContainer(ctx context.Context) (testcontainers.Container, string, int, error) {
	sharedContainerOnce.Do(func() {
		log.Println("=== Initializing shared PostgreSQL test container (this should only happen ONCE) ===")

		req := testcontainers.ContainerRequest{
			Image:        "postgres:17",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "postgres",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
			Name: "test_postgres",
			HostConfigModifier: (func(hc *container.HostConfig) {
				hc.AutoRemove = false
			}),
		}

		pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Reuse:            true,
		})
		if err != nil {
			sharedContainerErr = fmt.Errorf("failed to start container: %w", err)
			return
		}

		host, err := pgContainer.Host(ctx)
		if err != nil {
			sharedContainerErr = fmt.Errorf("failed to get container host: %w", err)
			return
		}

		port, err := pgContainer.MappedPort(ctx, "5432")
		if err != nil {
			sharedContainerErr = fmt.Errorf("failed to get container port: %w", err)
			return
		}

		sharedContainer = pgContainer
		sharedContainerHost = host
		sharedContainerPort = mustParsePort(port.Port())

		log.Printf("=== Shared PostgreSQL test container ready at %s:%s ===", host, port.Port())
	})

	if sharedContainerErr != nil {
		return nil, "", 0, sharedContainerErr
	}

	return sharedContainer, sharedContainerHost, sharedContainerPort, nil
}

// createDatabase creates a single test database with migrations
func createDatabase(ctx context.Context, testID string, config DatabaseConfig, host string, port int, postgresClient *db.DBPool) (*TestDBContext, error) {
	dbName := fmt.Sprintf("%s_%s", config.database, testID)

	// Insert database name into test_databases table
	_, err := postgresClient.Exec(ctx,
		"INSERT INTO test_databases (dbname) VALUES ($1)",
		dbName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert db_name: %w", err)
	}

	// Create the test database
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err = postgresClient.Exec(ctx, createDBQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	dbURL := fmt.Sprintf("postgres://postgres:postgres@%s:%d/%s?sslmode=disable",
		host, port, dbName)

	err = frameworkDB.RunMigrations(dbURL, config.migrations)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	log.Printf("Migrations completed successfully for database %s", dbName)

	// Connect to the test database
	dbConfig := &db.Config{
		Host:              host,
		Port:              port,
		User:              "postgres",
		Password:          "postgres",
		Database:          dbName,
		SSLMode:           "disable",
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}

	client, err := db.NewPool(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &TestDBContext{
		client:     client,
		dbName:     dbName,
		dbURL:      dbURL,
		migrations: config.migrations,
	}, nil
}

// createServer creates a test server instance
func createServer(_ context.Context, config ServerConfig, dependencyProvider *TestContextProvider) (*TestServerContext, error) {

	// Generate random ports in range 30000-40000
	grpcPort := 30000 + rand.Intn(10000)
	httpPort := grpcPort + 1

	server := config.provider(dependencyProvider)

	// Channel to signal when server has completely shut down
	serverDone := make(chan struct{})

	// Launch server in background
	go func() {
		defer close(serverDone)
		if err := server.Launch(grpcPort, httpPort); err != nil {
			log.Printf("Server launch error: %v", err)
		}
	}()

	// Wait for server to be ready with timeout
	if err := server.WaitUntilReady(10 * time.Second); err != nil {
		server.Shutdown()
		// Wait for the goroutine to clean up
		<-serverDone
		return nil, fmt.Errorf("server startup failed: %w", err)
	}

	return &TestServerContext{
		server:     server,
		grpcPort:   grpcPort,
		httpPort:   httpPort,
		serverDone: serverDone,
	}, nil
}

// mustParsePort converts string port to int, panics on error
func mustParsePort(port string) int {
	var p int
	if _, err := fmt.Sscanf(port, "%d", &p); err != nil {
		panic(fmt.Sprintf("invalid port: %s", port))
	}
	return p
}

func (tx *TestContext) GetGrpcClient(server ServerConfig) string {
	var serverContext *TestServerContext
	if serverContext = tx.servers[server.server]; serverContext == nil {
		panic(fmt.Sprintf("Server not registered: %s", server.server))
	}
	return fmt.Sprintf("localhost:%d", serverContext.grpcPort)
}

// CleanUp tears down the test context, dropping all test databases and shutting down servers
// Note: This does NOT terminate the shared container, which is reused across tests
func (tc *TestContext) CleanUp(ctx context.Context) error {
	// Shutdown all servers
	for name, srv := range tc.servers {
		if srv != nil {
			srv.Shutdown()
			log.Printf("Shut down test server: %s", name)
		}
	}

	// Close all database clients
	for _, db := range tc.databases {
		if db.client != nil {
			db.client.Close()
			log.Printf("Closed database client: %s", db.dbName)
		}
	}

	// Delete database entries and drop databases
	if tc.postgresClient != nil {
		for _, db := range tc.databases {

			// Delete from test_databases table
			_, err := tc.postgresClient.Exec(ctx,
				"DELETE FROM test_databases WHERE dbname = $1",
				db.dbName,
			)
			if err != nil {
				log.Printf("Warning: failed to delete db_name %s from test_databases table: %v", db.dbName, err)
			}

			// Drop the test database (best effort - may fail if connections still open)
			dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", db.dbName)
			_, err = tc.postgresClient.Exec(ctx, dropQuery)
			if err != nil {
				log.Printf("Warning: failed to drop test database %s: %v", db.dbName, err)
			} else {
				log.Printf("Dropped test database: %s", db.dbName)
			}
		}

		// Close test client
		tc.postgresClient.Close()
		tc.postgresClient = nil
	}

	// DO NOT terminate the container - it's shared across all tests and reused
	// The container will be cleaned up when the test process exits

	return nil
}

// TerminateSharedContainer terminates the shared PostgreSQL container
// We dont call this, keep container alive to ensure tests are quick
// This should typically only be called in TestMain after all tests complete
// Example:
//
//	func TestMain(m *testing.M) {
//	    code := m.Run()
//	    ctx := context.Background()
//	    if err := test.TerminateSharedContainer(ctx); err != nil {
//	        log.Printf("Warning: failed to terminate shared container: %v", err)
//	    }
//	    os.Exit(code)
//	}
func TerminateSharedContainer(ctx context.Context) error {
	if sharedContainer != nil {
		log.Println("Terminating shared PostgreSQL test container...")
		if err := sharedContainer.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate shared container: %w", err)
		}
		sharedContainer = nil
		log.Println("Shared PostgreSQL test container terminated")
	}
	return nil
}
