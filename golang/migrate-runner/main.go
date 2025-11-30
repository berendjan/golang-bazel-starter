package main

import (
	"context"
	"log"
	"os"
	"time"

	configdb "github.com/berendjan/golang-bazel-starter/golang/config/db"
	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
)

func main() {
	log.Println("Starting migration runner...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get database configuration
	cfg := getDatabaseConfig()

	// Connect to database
	log.Printf("Connecting to database at %s:%d...", cfg.Host, cfg.Port)
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	db.MustRunMigrations(cfg.ConnectionString(), configdb.MigrationsFS)

	log.Println("Migration runner completed successfully")
}

// getDatabaseConfig returns database configuration from environment or defaults
func getDatabaseConfig() *db.Config {
	cfg := &db.Config{
		Host:              getEnv("DB_HOST", "app-postgres-rw.app-namespace.svc.cluster.local"),
		Port:              5432,
		User:              getEnv("DB_USER", "migrate-runner"),
		Password:          "", // Not used with certificate authentication
		Database:          getEnv("DB_NAME", "app"),
		SSLMode:           "verify-full",
		SSLCert:           "/mnt/client-certs/tls.crt",
		SSLKey:            "/mnt/client-certs/tls.key",
		SSLRootCert:       "/mnt/postgres-ca/ca-bundle.crt",
		MaxConns:          5,  // Lower for migration job
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
