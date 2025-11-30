package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string

	// SSL Certificate paths
	SSLCert     string // Path to client certificate
	SSLKey      string // Path to client private key
	SSLRootCert string // Path to CA certificate

	// Connection pool settings
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig(dbName string) *Config {
	return &Config{
		Host:              "app-postgres-rw.app-namespace.svc.cluster.local",
		Port:              5432,
		User:              "grpcserver",
		Password:          "", // Not used with certificate authentication
		Database:          dbName,
		SSLMode:           "verify-full",
		SSLCert:           "/mnt/client-certs/tls.crt",
		SSLKey:            "/mnt/client-certs/tls.key",
		SSLRootCert:       "/mnt/postgres-ca/ca-bundle.crt",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

type DBPool struct {
	*pgxpool.Pool
	database string
}

// ConnectionString builds a PostgreSQL connection string from the config
func (c *Config) ConnectionString() string {
	var connStr string

	// When using SSL certificate authentication, omit the password
	if c.SSLCert != "" {
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Database, c.SSLMode,
		)
		connStr += fmt.Sprintf(" sslcert=%s", c.SSLCert)
		connStr += fmt.Sprintf(" sslkey=%s", c.SSLKey)
		connStr += fmt.Sprintf(" sslrootcert=%s", c.SSLRootCert)
	} else {
		// Traditional password authentication
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
		)
	}

	return connStr
}

// NewPool creates a new PostgreSQL connection pool
func NewPool(ctx context.Context, cfg *Config) (*DBPool, error) {
	// Build pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to PostgreSQL at %s:%d (database: %s)", cfg.Host, cfg.Port, cfg.Database)
	return &DBPool{pool, cfg.Database}, nil
}

// MustNewPool creates a new connection pool or panics on error
func MustNewPool(ctx context.Context, cfg *Config) *DBPool {
	pool, err := NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return pool
}

// Close gracefully closes the database connection pool
func (pool *DBPool) Close() {
	if pool == nil || pool.Pool == nil {
		log.Printf("WARN: Attempted to close nil pool")
		return
	}
	log.Printf("Closing database connection pool (database: %s, stats before: %+v)", pool.database, pool.Pool.Stat())
	pool.Pool.Close()
	log.Printf("Database connection pool closed (database: %s)", pool.database)
}
