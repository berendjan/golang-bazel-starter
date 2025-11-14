package test_test

import (
	"context"
	"testing"

	"github.com/berendjan/golang-bazel-starter/golang/test"
)

// // TestBuilderWithMultipleDatabases demonstrates using the builder to create multiple databases
// func TestBuilderWithMultipleDatabases(t *testing.T) {
// 	ctx := context.Background()

// 	// Create test context with multiple databases
// 	tc, err := test.NewTestContextBuilder().
// 		WithDatabase("main", configDb.MigrationsFS).
// 		WithDatabase("analytics", configDb.MigrationsFS). // Reusing same migrations for demo
// 		Build(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to create test context: %v", err)
// 	}
// 	defer func() {
// 		if err := tc.CleanUp(ctx); err != nil {
// 			t.Logf("Warning: cleanup failed: %v", err)
// 		}
// 	}()

// 	// Use named database clients
// 	mainDB := tc.Database("main")
// 	analyticsDB := tc.Database("analytics")

// 	// Insert test data into main database
// 	_, err = mainDB.Exec(ctx,
// 		"INSERT INTO accounts (id, type, created_at, updated_at) VALUES ($1, $2, $3, $4)",
// 		[]byte("main-account"),
// 		1,
// 		time.Now(),
// 		time.Now(),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to insert into main database: %v", err)
// 	}

// 	// Insert test data into analytics database
// 	_, err = analyticsDB.Exec(ctx,
// 		"INSERT INTO accounts (id, type, created_at, updated_at) VALUES ($1, $2, $3, $4)",
// 		[]byte("analytics-account"),
// 		2,
// 		time.Now(),
// 		time.Now(),
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to insert into analytics database: %v", err)
// 	}

// 	// Verify data is isolated in each database
// 	var mainCount, analyticsCount int

// 	err = mainDB.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&mainCount)
// 	if err != nil {
// 		t.Fatalf("Failed to query main database: %v", err)
// 	}

// 	err = analyticsDB.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&analyticsCount)
// 	if err != nil {
// 		t.Fatalf("Failed to query analytics database: %v", err)
// 	}

// 	if mainCount != 1 {
// 		t.Errorf("Expected 1 account in main database, got %d", mainCount)
// 	}

// 	if analyticsCount != 1 {
// 		t.Errorf("Expected 1 account in analytics database, got %d", analyticsCount)
// 	}

// 	t.Log("Multiple databases test completed successfully")
// }

// TestBuilderWithServers demonstrates using the builder to create servers
func TestBuilderWithServers(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Send request to server with client

	// Access servers
	// httpServer := tc.Server("http")
	// grpcServer := tc.Server("grpc")

	// if httpServer.grpcPort != 8080 {
	// 	t.Errorf("Expected HTTP server gRPC port 8080, got %d", httpServer.grpcPort)
	// }

	// if grpcServer.grpcPort != 9090 {
	// 	t.Errorf("Expected gRPC server gRPC port 9090, got %d", grpcServer.grpcPort)
	// }

	// expectedHTTPURL := "http://localhost:8080"
	// if httpServer.baseURL != expectedHTTPURL {
	// 	t.Errorf("Expected HTTP URL %s, got %s", expectedHTTPURL, httpServer.baseURL)
	// }

	t.Log("Server configuration test completed successfully")
}

// // TestBuilderNoDatabases demonstrates that builder works with only servers
// func TestBuilderNoDatabases(t *testing.T) {
// 	ctx := context.Background()

// 	tc, err := test.NewTestContextBuilder().
// 		WithServer("api").
// 		Build(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to create test context: %v", err)
// 	}
// 	defer func() {
// 		if err := tc.CleanUp(ctx); err != nil {
// 			t.Logf("Warning: cleanup failed: %v", err)
// 		}
// 	}()

// 	// server := tc.Server("api")
// 	// if server.grpcPort != 8000 {
// 	// 	t.Errorf("Expected server on gRPC port 8000, got %d", server.grpcPort)
// 	// }

// 	t.Log("Server-only test completed successfully")
// }

// // TestDatabaseURL demonstrates getting database connection URLs
// func TestDatabaseURL(t *testing.T) {
// 	ctx := context.Background()

// 	tc, err := test.NewTestContextBuilder().
// 		WithDatabase("db1", configDb.MigrationsFS).
// 		WithDatabase("db2", configDb.MigrationsFS).
// 		Build(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to create test context: %v", err)
// 	}
// 	defer func() {
// 		if err := tc.CleanUp(ctx); err != nil {
// 			t.Logf("Warning: cleanup failed: %v", err)
// 		}
// 	}()

// 	// Get database URLs
// 	db1URL := tc.DatabaseURL("db1")
// 	db2URL := tc.DatabaseURL("db2")

// 	if db1URL == "" {
// 		t.Error("db1 URL is empty")
// 	}

// 	if db2URL == "" {
// 		t.Error("db2 URL is empty")
// 	}

// 	if db1URL == db2URL {
// 		t.Error("Database URLs should be different for different databases")
// 	}

// 	t.Logf("db1 URL: %s", db1URL)
// 	t.Logf("db2 URL: %s", db2URL)
// }

// // TestConcurrentContextCreation verifies that multiple tests can safely create contexts concurrently
// // and they all share the same container
// func TestConcurrentContextCreation(t *testing.T) {
// 	ctx := context.Background()
// 	const numGoroutines = 5

// 	var wg sync.WaitGroup
// 	wg.Add(numGoroutines)

// 	// Track any errors
// 	errChan := make(chan error, numGoroutines)

// 	for i := 0; i < numGoroutines; i++ {
// 		go func(id int) {
// 			defer wg.Done()

// 			tc, err := test.NewTestContextBuilder().
// 				WithDatabase("concurrent", configDb.MigrationsFS).
// 				Build(ctx)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}
// 			defer func() {
// 				if err := tc.CleanUp(ctx); err != nil {
// 					t.Logf("Warning: cleanup failed for goroutine %d: %v", id, err)
// 				}
// 			}()

// 			// Verify the database works
// 			db := tc.Database("concurrent")
// 			var count int
// 			err = db.QueryRow(ctx, "SELECT COUNT(*) FROM accounts").Scan(&count)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}

// 			if count != 0 {
// 				t.Errorf("Goroutine %d: expected empty database, got %d accounts", id, count)
// 			}
// 		}(i)
// 	}

// 	wg.Wait()
// 	close(errChan)

// 	// Check for errors
// 	for err := range errChan {
// 		if err != nil {
// 			t.Errorf("Concurrent test failed: %v", err)
// 		}
// 	}

// 	t.Log("Concurrent context creation completed successfully")
// }

// // TestSharedContainerReuse verifies that the container is reused across tests
// func TestSharedContainerReuse(t *testing.T) {
// 	ctx := context.Background()

// 	// Create first context
// 	tc1, err := test.NewTestContextBuilder().
// 		WithDatabase("first", configDb.MigrationsFS).
// 		Build(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to create first test context: %v", err)
// 	}
// 	defer tc1.CleanUp(ctx)

// 	// Create second context
// 	tc2, err := test.NewTestContextBuilder().
// 		WithDatabase("second", configDb.MigrationsFS).
// 		Build(ctx)
// 	if err != nil {
// 		t.Fatalf("Failed to create second test context: %v", err)
// 	}
// 	defer tc2.CleanUp(ctx)

// 	// Both contexts should reference the same underlying container
// 	// We can verify they use the same host and port
// 	db1URL := tc1.DatabaseURL("first")
// 	db2URL := tc2.DatabaseURL("second")

// 	t.Logf("First context DB URL: %s", db1URL)
// 	t.Logf("Second context DB URL: %s", db2URL)

// 	// The URLs should have the same host:port but different database names
// 	// This confirms they share the same container
// 	if db1URL == db2URL {
// 		t.Error("Database URLs should be different (different database names)")
// 	}

// 	t.Log("Shared container reuse verified")
// }
