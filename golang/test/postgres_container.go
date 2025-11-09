package test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer     testcontainers.Container
	pgPool          *pgxpool.Pool
	containerOnce   sync.Once
	containerConfig *PostgresContainerConfig
)

// PostgresContainerConfig holds configuration for the test PostgreSQL container
type PostgresContainerConfig struct {
	Database string
	Username string
	Password string
	Image    string
}

// DefaultPostgresConfig returns default PostgreSQL container configuration
func DefaultPostgresConfig() *PostgresContainerConfig {
	return &PostgresContainerConfig{
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		Image:    "postgres:16-alpine",
	}
}

// GetPostgresContainer returns a singleton PostgreSQL testcontainer
// The container is created once and reused across all tests
func GetPostgresContainer(ctx context.Context) (testcontainers.Container, error) {
	var err error
	containerOnce.Do(func() {
		if containerConfig == nil {
			containerConfig = DefaultPostgresConfig()
		}

		req := testcontainers.ContainerRequest{
			Image:        containerConfig.Image,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       containerConfig.Database,
				"POSTGRES_USER":     containerConfig.Username,
				"POSTGRES_PASSWORD": containerConfig.Password,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		}

		pgContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})

		if err != nil {
			log.Printf("Failed to start PostgreSQL container: %v", err)
			return
		}

		log.Println("PostgreSQL testcontainer started successfully")
	})

	return pgContainer, err
}

// GetPostgresPool returns a singleton connection pool to the test PostgreSQL container
// It automatically creates and starts the container if needed
func GetPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	if pgPool != nil {
		return pgPool, nil
	}

	container, err := GetPostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	if containerConfig == nil {
		containerConfig = DefaultPostgresConfig()
	}

	// Create database configuration
	config := &db.Config{
		Host:              host,
		Port:              port.Int(),
		User:              containerConfig.Username,
		Password:          containerConfig.Password,
		Database:          containerConfig.Database,
		SSLMode:           "disable",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}

	// Create connection pool
	pgPool, err = db.NewPool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	log.Printf("Connected to PostgreSQL testcontainer at %s:%d", host, port.Int())
	return pgPool, nil
}

// CleanupPostgresContainer stops and removes the PostgreSQL container
// This should be called in TestMain or after all tests complete
func CleanupPostgresContainer(ctx context.Context) error {
	if pgPool != nil {
		pgPool.Close()
		pgPool = nil
	}

	if pgContainer != nil {
		log.Println("Stopping PostgreSQL testcontainer...")
		if err := pgContainer.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
		log.Println("PostgreSQL testcontainer stopped")
	}

	return nil
}

// ResetDatabase truncates all tables in the test database
// Useful for cleaning up between tests
func ResetDatabase(ctx context.Context) error {
	if pgPool == nil {
		return fmt.Errorf("database pool not initialized")
	}

	// Truncate all tables (add your tables here)
	tables := []string{"accounts"}

	for _, table := range tables {
		_, err := pgPool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}
