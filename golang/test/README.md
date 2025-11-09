# Test Utilities

A library for managing test infrastructure, including PostgreSQL testcontainers.

## PostgreSQL Testcontainer

This library provides a singleton PostgreSQL testcontainer that can be reused across all tests in your test suite.

### Features

- **Singleton Pattern**: Container is created once and reused across all tests
- **Automatic Cleanup**: Container can be cleaned up after all tests complete
- **Connection Pooling**: Provides a singleton connection pool to the test database
- **Database Reset**: Utility to truncate all tables between tests

### Basic Usage

```go
package mypackage_test

import (
    "context"
    "testing"

    "github.com/berendjan/golang-bazel-starter/golang/test"
)

func TestMyFunction(t *testing.T) {
    ctx := context.Background()

    // Get the singleton connection pool
    pool, err := test.GetPostgresPool(ctx)
    if err != nil {
        t.Fatalf("Failed to get postgres pool: %v", err)
    }

    // Use the pool in your tests
    _, err = pool.Exec(ctx, "INSERT INTO accounts (id, type) VALUES ($1, $2)",
        []byte("test-account"), 1)
    if err != nil {
        t.Fatal(err)
    }

    // Clean up test data
    defer test.ResetDatabase(ctx)
}
```

### Test Suite Setup with TestMain

For optimal performance, use `TestMain` to manage the container lifecycle:

```go
package mypackage_test

import (
    "context"
    "os"
    "testing"

    "github.com/berendjan/golang-bazel-starter/golang/test"
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Start container before tests
    _, err := test.GetPostgresContainer(ctx)
    if err != nil {
        panic(err)
    }

    // Run migrations
    pool, err := test.GetPostgresPool(ctx)
    if err != nil {
        panic(err)
    }

    // Create schema
    _, err = pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS accounts (
            id BYTEA PRIMARY KEY,
            type INTEGER NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        panic(err)
    }

    // Run tests
    code := m.Run()

    // Cleanup
    test.CleanupPostgresContainer(ctx)

    os.Exit(code)
}

func TestAccountCreation(t *testing.T) {
    ctx := context.Background()
    pool, _ := test.GetPostgresPool(ctx)

    // Test code here...

    // Clean up after each test
    defer test.ResetDatabase(ctx)
}
```

### Running with Migrations

You can run your actual database migrations on the test container:

```go
import (
    "github.com/berendjan/golang-bazel-starter/golang/db"
    "github.com/berendjan/golang-bazel-starter/golang/test"
)

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Get container
    container, err := test.GetPostgresContainer(ctx)
    if err != nil {
        panic(err)
    }

    // Get connection details
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    // Build connection string
    connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%d/testdb?sslmode=disable",
        host, port.Int())

    // Run migrations
    err = db.RunMigrations(connStr)
    if err != nil {
        panic(err)
    }

    // Run tests
    code := m.Run()

    // Cleanup
    test.CleanupPostgresContainer(ctx)
    os.Exit(code)
}
```

### Configuration

Customize the container configuration:

```go
// Note: Configuration must be set before first call to GetPostgresContainer
config := &test.PostgresContainerConfig{
    Database: "mydb",
    Username: "myuser",
    Password: "mypass",
    Image:    "postgres:16-alpine",
}

// This would require modifying the library to accept config parameter
```

## Running Tests

```bash
# Run all tests
bazel test //golang/test:all

# Run specific test
bazel test //golang/test:go_default_test

# Run with verbose output
bazel test //golang/test:all --test_output=all
```

## Requirements

- Docker must be running on your machine
- Testcontainers will automatically pull the PostgreSQL image on first run

## Performance Tips

1. **Use TestMain**: Container startup is expensive (~2-5 seconds). Use `TestMain` to create it once.
2. **Reset Instead of Recreate**: Use `ResetDatabase()` to clean up between tests instead of recreating the container.
3. **Parallel Tests**: The singleton pattern means tests sharing the same container cannot run in parallel if they modify data. Use `t.Parallel()` carefully.

## Troubleshooting

**Container won't start:**
- Ensure Docker is running
- Check if port 5432 is available
- Increase startup timeout if needed

**Connection refused:**
- Wait for container to be fully ready (default: waits for "ready to accept connections")
- Check container logs with `docker logs <container-id>`

**Tests are slow:**
- Make sure you're using the singleton pattern (only one container for all tests)
- Use `ResetDatabase()` instead of recreating containers
- Consider using `t.Parallel()` for read-only tests
