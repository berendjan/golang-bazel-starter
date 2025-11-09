package test

import (
	"context"
	"testing"
)

// TestPostgresContainer tests that we can create and connect to a PostgreSQL container
func TestPostgresContainer(t *testing.T) {
	ctx := context.Background()

	// Get the container (singleton - will be created once)
	container, err := GetPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to get postgres container: %v", err)
	}

	if container == nil {
		t.Fatal("Container should not be nil")
	}

	// Verify container is running
	state, err := container.State(ctx)
	if err != nil {
		t.Fatalf("Failed to get container state: %v", err)
	}

	if !state.Running {
		t.Fatal("Container should be running")
	}

	t.Log("PostgreSQL container is running successfully")
}

// TestPostgresPool tests that we can get a connection pool
func TestPostgresPool(t *testing.T) {
	ctx := context.Background()

	// Get the pool (singleton)
	pool, err := GetPostgresPool(ctx)
	if err != nil {
		t.Fatalf("Failed to get postgres pool: %v", err)
	}

	if pool == nil {
		t.Fatal("Pool should not be nil")
	}

	// Test connection
	err = pool.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Log("Successfully connected to PostgreSQL testcontainer")
}

// TestPostgresPoolReuse tests that the pool is reused
func TestPostgresPoolReuse(t *testing.T) {
	ctx := context.Background()

	// Get pool first time
	pool1, err := GetPostgresPool(ctx)
	if err != nil {
		t.Fatalf("Failed to get first pool: %v", err)
	}

	// Get pool second time - should be the same instance
	pool2, err := GetPostgresPool(ctx)
	if err != nil {
		t.Fatalf("Failed to get second pool: %v", err)
	}

	// Verify they are the same instance
	if pool1 != pool2 {
		t.Fatal("Pool instances should be the same (singleton)")
	}

	t.Log("Pool singleton pattern working correctly")
}

// TestResetDatabase tests that we can reset the database
func TestResetDatabase(t *testing.T) {
	ctx := context.Background()

	// Get pool first
	pool, err := GetPostgresPool(ctx)
	if err != nil {
		t.Fatalf("Failed to get pool: %v", err)
	}

	// Create accounts table if it doesn't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id BYTEA PRIMARY KEY,
			type INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = pool.Exec(ctx, `
		INSERT INTO accounts (id, type) VALUES ($1, $2)
	`, []byte("test-account"), 1)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Verify data exists
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count accounts: %v", err)
	}

	if count == 0 {
		t.Fatal("Should have at least one account")
	}

	// Reset database
	err = ResetDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to reset database: %v", err)
	}

	// Verify data was deleted
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count accounts after reset: %v", err)
	}

	if count != 0 {
		t.Fatalf("Expected 0 accounts after reset, got %d", count)
	}

	t.Log("Database reset successfully")
}
