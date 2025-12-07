# Golang-bazel-starter - AI Assistant Guide

This document provides guidance for AI assistants (like Claude) working on this codebase.

## Project Overview

This is a production-ready Go + Bazel template featuring:
- Type-safe code generation for message routing
- gRPC service with middleware support
- PostgreSQL with database migrations
- Testcontainers for integration testing
- Compile-time validation of routing configuration
- Kubernetes manifests with Kustomize

## Project Structure

**IMPORTANT**: Kubernetes manifests MUST be created in `k8s/app/` directory, NOT in `dev/` or other locations.

```
k8s/
├── app/                    # All k8s application manifests go here
│   ├── grpcserver/        # Example service
│   │   ├── base/          # Base kustomization
│   │   └── kustomization.yaml
│   ├── registry/          # Container registry
│   │   ├── base/
│   │   └── kustomization.yaml
│   └── common/            # Shared resources
│       ├── namespace/
│       ├── deployment/
│       └── environment/
└── infra/                 # Bazel rules for k8s
    ├── image.bzl
    ├── server.bzl
    └── k8s.bzl
```

## Critical Rules

### 1. Dependency Management

When adding new Go dependencies, **ALWAYS** follow this exact sequence:

```bash
# 1. Update go.mod
go mod tidy -e

# 2. Update Bazel dependencies
bazel mod tidy

# 3. Update BUILD files (NEVER edit BUILD.bazel manually!)
bazel run //:gazelle

# 4. Verify the build
bazel build //...
```

**NEVER manually edit BUILD.bazel files** - always use Gazelle. Read gazelle directives in BUILD files before making changes.

### 2. Code Generation System

This project has TWO code generators that work together:

#### Interface Generator (`tools/codegen/interface-gen`)

**What it does:**
- Reads `golang/generated/routing.yaml`
- Generates Go interfaces in `golang/generated/interfaces/`
- Creates handler interfaces (what each component must implement)
- Creates sendable interfaces (what each component can send to)
- Automatically determines method signatures based on receiver position in chain

**Key behavior:**
- **Last receiver** in a chain → Returns `(result, error)`
- **Intermediate receiver** in a chain → Returns `error` only
- This allows automatic error propagation in middleware chains

#### Messenger Generator (`tools/codegen/messenger-gen`)

**What it does:**
- Reads `golang/generated/routing.yaml`
- Generates messenger implementation in `golang/grpcserver/messenger/`
- Creates routing methods that chain handlers together
- Automatically handles error propagation through middleware

**Key behavior:**
- Only includes handlers that **receive messages** in struct fields
- Handlers that only send (like `AccountApi`) are NOT included
- Generates chaining code for multiple receivers with error handling

### 3. The Routing YAML File

**Location:** `golang/generated/routing.yaml`

This is the **single source of truth** for message routing. It contains:

```yaml
# Configuration for interface generation
interfaces:
  package: interfaces
  imports:
    - 'commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"'
    - 'configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"'

# Configuration for messenger generation
messenger:
  package: messenger
  messenger_name: GrpcMessenger
  imports:
    - 'geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"'
    - 'commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"'
    - 'configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"'

# Handler definitions (all components in the system)
handlers:
  - name: accountRepository
    type: "configrepository.AccountDbRepository"
  - name: AccountApi
    type: "configapi.ConfigurationApi"
  - name: middlewareOne
    type: "middleone.MiddleOne"
  - name: middlewareTwo
    type: "middletwo.MiddleTwo"

# Message routing (who sends what to whom)
routes:
  - source: AccountApi
    messages:
      - message: "*configpb.MiddleOneRequestProto"
        response: "(*configpb.AccountConfigurationProto, error)"
        receivers:
          - middlewareOne

  - source: middlewareOne
    messages:
      - message: "*configpb.MiddleOneRequestProto"
        response: "(*configpb.AccountConfigurationProto, error)"
        receivers:
          - middlewareTwo      # Intermediate - returns error only
          - accountRepository  # Final - returns (result, error)
```

**Important validation:**
- All handlers referenced in `routes` must exist in `handlers` list
- Build will fail if unknown handlers are referenced
- This validation happens at compile time

### 4. Adding New Message Routes

**Complete workflow:**

1. **Update `routing.yaml`**:
   ```yaml
   - source: someSource
     messages:
       - message: "*configpb.NewRequestProto"
         response: "(*configpb.NewResponseProto, error)"
         receivers:
           - someMiddleware
           - finalHandler
   ```

2. **Regenerate interfaces and messenger**:
   ```bash
   bazel build //golang/generated/interfaces:interfaces
   bazel build //golang/grpcserver/messenger:messenger
   ```

3. **Implement handler methods** - The compiler will tell you what's missing!

4. **Wire up in main.go** if adding new handlers:
   ```go
   newHandler := &newpackage.NewHandler{}
   messenger := messenger.NewGrpcMessenger(
       repo,
       middleware1,
       middleware2,
       newHandler,  // Add new handler
   )
   ```

### 5. Understanding Handler Types

**Three types of handlers:**

1. **Terminal handlers** (repositories):
   - Only receive messages, never send
   - Methods: `Handle(ctx, message) (result, error)`
   - Example: `AccountDbRepository`

2. **Middleware handlers**:
   - Receive and send messages
   - Methods: `Handle(ctx, message, next Sendable) (result, error)` or just `error`
   - Can modify message or return early with error
   - Example: `MiddlewareOne`, `MiddlewareTwo`

3. **API handlers**:
   - Only send messages, never receive
   - NOT included in messenger struct
   - Call messenger methods to initiate message flow
   - Example: `AccountApi`

### 6. Generated Code Structure

**Generated interfaces** (`golang/generated/interfaces/generated_interfaces.go`):
```go
// What a handler must implement
type MiddlewareOneInterface interface {
    HandleMiddleOneRequest(ctx, message, next MiddlewareOneSendable) (*Result, error)
}

// What a handler can send to
type MiddlewareOneSendable interface {
    SendMiddleOneRequestFromMiddlewareOne(ctx, message) (*Result, error)
}
```

**Generated messenger** (`golang/grpcserver/messenger/generated_messenger.go`):
```go
type GrpcMessenger struct {
    accountRepository geninterfaces.AccountRepositoryInterface
    middlewareOne     geninterfaces.MiddlewareOneInterface
    middlewareTwo     geninterfaces.MiddlewareTwoInterface
    // Note: AccountApi is NOT here (it only sends, never receives)
}

// Automatically chains receivers with error handling
func (m *GrpcMessenger) SendFromSource(ctx, msg) (*Result, error) {
    if err := m.middlewareTwo.Handle(ctx, msg, m); err != nil {
        return nil, err
    }
    return m.accountRepository.Handle(ctx, msg)
}
```

### 7. Testing

**Test structure uses TestContainers:**

```go
func TestSomething(t *testing.T) {
    // Create test context with database
    tc := test.NewTestContext(t).
        WithDatabase(test.ConfigDb).
        Build()
    defer tc.Cleanup()

    // Get dependencies
    provider := tc.GetProvider()
    repo := provider.GetAccountRepository()

    // Test
    result, err := repo.HandleSomeRequest(ctx, req)
    assert.NoError(t, err)
}
```

**Important:**
- Database containers are shared across tests for performance
- Each test gets isolated database with migrations applied
- Use `test.TestMiddleOne` for test-specific middleware implementations

### 8. Common Tasks

#### Add a new middleware:

```bash
# 1. Create package
mkdir -p golang/middleware/mynewmiddleware

# 2. Create implementation
# golang/middleware/mynewmiddleware/mynewmiddleware.go

# 3. Add to routing.yaml handlers and routes

# 4. Regenerate
bazel build //golang/generated/interfaces:interfaces
bazel build //golang/grpcserver/messenger:messenger

# 5. Update main.go to instantiate it

# 6. Run Gazelle
bazel run //:gazelle
```

#### Add a new proto message:

```bash
# 1. Update .proto files in proto/

# 2. Build protos
bazel build //proto/...

# 3. Add route in routing.yaml

# 4. Regenerate interfaces/messenger

# 5. Implement handlers
```

#### Debugging generation:

```bash
# Force regeneration (ignore cache)
bazel build --nocache_test_results //golang/generated/interfaces:interfaces

# Check generated code
cat bazel-bin/golang/generated/interfaces/generated_interfaces.go
cat bazel-bin/golang/grpcserver/messenger/generated_messenger.go
```

### 9. Key Design Decisions

**No generics in messenger:**
- Messenger uses interfaces only (no generic type parameters)
- Simpler type signatures: `*GrpcMessenger` instead of `*GrpcMessenger[T1, T2, T3]`
- Same compile-time type safety via interface constraints
- Easier to read and maintain

**Intermediate vs final receivers:**
- Based on position in `receivers` list
- Allows type-safe error-only returns for middleware
- Automatic error propagation through chains

**Single routing.yaml:**
- Both generators read the same file
- Generator-specific config in `interfaces:` and `messenger:` sections
- Shared handlers and routes
- Prevents configuration drift

### 10. Files to Never Manually Edit

❌ **DO NOT EDIT:**
- Any `BUILD.bazel` file (use `bazel run //:gazelle` instead)
- `golang/generated/interfaces/generated_interfaces.go` (generated)
- `golang/grpcserver/messenger/generated_messenger.go` (generated)
- Any file with `// Code generated` header

✅ **OK TO EDIT:**
- `golang/generated/routing.yaml` (source of truth)
- Handler implementations (`middleware/*`, `config/repository/*`)
- `main.go` (wiring)
- Tests
- Proto definitions

### 11. Troubleshooting

**"unknown handler" error:**
- Handler referenced in routes doesn't exist in handlers list
- Fix: Add handler to `routing.yaml` handlers section

**"does not implement interface" error:**
- Handler missing required method
- Check generated interface for exact signature needed
- Remember: intermediate receivers return `error`, final receivers return `(result, error)`

**Import errors after regeneration:**
- Run `go mod tidy -e && bazel mod tidy && bazel run //:gazelle`
- Rebuild: `bazel build //...`

**BUILD file changes not working:**
- Never edit BUILD files manually
- Always use: `bazel run //:gazelle`

### 12. Container Image Infrastructure (`k8s/infra`)

The project includes custom Bazel rules for building and deploying container images.

#### Key Files

**`k8s/infra/image.bzl`**:
- `image()` macro - Builds OCI images with automatic tagging
- `build_sha265_tag` rule - Extracts 7-character sha256 tag from image digest
- Automatically creates push targets for multiple repositories

**`k8s/infra/server.bzl`**:
- `go_binary()` wrapper macro - Extends standard go_binary with containerization
- Automatically cross-compiles for Linux AMD64
- Creates image with distroless base
- Exposes ports 25000 (gRPC) and 26000 (HTTP)

**`k8s/infra/k8s.bzl`**:
- `deploy_targets()` - Generates k8s manifests for multiple clusters
- `target()` - Defines a deployment target configuration
- `cluster()` - Defines cluster configurations (dev, staging, prod)
- Integrates with Kustomize for overlay management

#### Local Registry Setup

**IMPORTANT**: You need a local Docker registry (v3) running on port 5001:

```bash
# Start local registry
docker run -d -p 5001:5000 --name registry registry:3

# Verify it's running
curl http://localhost:5001/v2/_catalog
```

#### How the `go_binary` Macro Works

When you use the `go_binary` macro from `k8s/infra/server.bzl`:

```go
load("//k8s/infra:server.bzl", "go_binary")

go_binary(
    name = "grpcserver",
    embed = [":grpcserver_lib"],
    visibility = ["//visibility:public"],
)
```

It automatically creates these targets:

1. **`grpcserver`** - Standard Go binary
2. **`grpcserver_cross`** - Cross-compiled Linux AMD64 binary
3. **`grpcserver_tar`** - Tarball of the cross-compiled binary
4. **`grpcserver_image`** - OCI image with distroless base
5. **`grpcserver_remote_tag`** - File containing 7-char sha256 tag
6. **`grpcserver_localhost_push_sha256_tag`** - Push with sha256 tag
7. **`grpcserver_localhost_push_latest_tag`** - Push with latest tag
8. **`grpcserver_push`** - Push both tags (multirun)

#### Building and Pushing Images

```bash
# Build the image
bazel build //golang/grpcserver:grpcserver_image

# Push to local registry (both sha256 and latest tags)
bazel run //golang/grpcserver:grpcserver_push

# Push only specific tag
bazel run //golang/grpcserver:grpcserver_localhost_push_sha256_tag
bazel run //golang/grpcserver:grpcserver_localhost_push_latest_tag
```

#### Verifying Images

```bash
# List all images in registry
curl http://localhost:5001/v2/_catalog

# List tags for specific image
curl http://localhost:5001/v2/grpcserver/tags/list

# Expected output:
# {"name":"grpcserver","tags":["latest","a1b2c3d"]}

# Get image manifest
curl http://localhost:5001/v2/grpcserver/manifests/latest
```

#### Image Customization

To customize the image for a new service:

1. **Use the `go_binary` macro** in your BUILD file:
   ```python
   load("//k8s/infra:server.bzl", "go_binary")

   go_binary(
       name = "myservice",
       embed = [":myservice_lib"],
       visibility = ["//visibility:public"],
   )
   ```

2. **Push to multiple registries** by modifying the `repositories` parameter in `k8s/infra/image.bzl`:
   ```python
   image(
       name = name,
       srcs = [":%s" % cross_name],
       base = "@distroless_base",
       entrypoint = ["/%s" % cross_name],
       exposed_ports = ["25000", "26000"],
       repositories = [
           "localhost:5001",
           "gcr.io/my-project",  # Add production registry
       ],
   )
   ```

3. **Different ports** - Modify `exposed_ports` in `k8s/infra/server.bzl`:
   ```python
   exposed_ports = [
       "8080",   # HTTP
       "9090",   # gRPC
   ],
   ```

#### Kubernetes Deployment

The `k8s/infra/k8s.bzl` rules generate Kubernetes manifests:

```bash
# Print generated k8s resources for dev cluster
bazel run //k8s/app:myapp-dev-print

# This uses Kustomize under the hood to:
# 1. Find the appropriate overlay (dev/staging/prod)
# 2. Apply cluster-specific substitutions
# 3. Generate final YAML
```

The `deploy_targets()` macro looks for kustomization files in this order:
1. `{dir}/{cluster_path}/{environment}/kustomization.yaml`
2. `{dir}/{cluster_path}/kustomization.yaml`
3. `{dir}/{environment}/kustomization.yaml`
4. `{dir}/kustomization.yaml`

#### Important Notes

- **Always use registry v3**: The push commands expect Docker registry v3 API
- **SHA256 tags are immutable**: Each image gets a unique 7-character tag from its digest
- **Latest tag is mutable**: Points to the most recently pushed image
- **Local registry is ephemeral**: Restarting removes all images (fine for development)
- **Cross-compilation is automatic**: `go_binary` macro handles Linux AMD64 builds

### 13. Database Migrations with dbmate

The project uses [dbmate](https://github.com/amacneil/dbmate) for database migrations in both production and tests.

#### Migration File Structure

Migrations are stored in `db/config/migrations/` using dbmate format:

```sql
-- migrate:up
CREATE TABLE IF NOT EXISTS accounts (
    id BYTEA PRIMARY KEY,
    type INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- migrate:down
DROP TABLE IF EXISTS accounts;
```

**Important**:
- Do NOT include `CREATE DATABASE` or `DROP DATABASE` statements
- Database creation is handled by PostgreSQL cluster bootstrap in Kubernetes
- Database creation is handled by test framework in tests

#### The dbmate Macro (`k8s/infra/dbmate.bzl`)

A reusable Bazel macro for building dbmate migration images:

```python
load("//k8s/infra:dbmate.bzl", "dbmate_image")

# Export migration files
filegroup(
    name = "migrations",
    srcs = glob(["migrations/*.sql"]),
    visibility = ["//visibility:public"],
)

# Create dbmate image with migrations
dbmate_image(
    name = "dbmate",
    migrations = [":migrations"],
    repositories = ["registry.localhost"],
)
```

**What it creates:**
- `{name}_image` - Multi-platform OCI image (linux/amd64, linux/arm64)
- `{name}_remote_tag` - SHA256 tag file
- `{name}_push` - Multirun target to push both sha256 and latest tags

**Parameters:**
- `name` - Base name for targets (e.g., "dbmate")
- `migrations` - List of migration files or filegroup label
- `package_dir` - Directory in image where migrations are stored (default: "/migrations")
- `repositories` - List of registries to push to (default: ["registry.localhost"])

#### Running Migrations in Kubernetes

Migrations run as a Kubernetes Job that executes once during deployment:

```yaml
# k8s/app/dbmate/base/job.yaml
containers:
- name: dbmate
  image: dbmate
  command: ["dbmate"]
  args: ["up"]
  env:
  - name: DATABASE_URL
    value: "postgres://dbmate@app-postgres-rw.app-namespace.svc.cluster.local:5432/config?sslmode=verify-full&sslcert=/mnt/client-certs/tls.crt&sslkey=/mnt/client-certs/tls.key&sslrootcert=/mnt/postgres-ca/ca.crt"
  - name: DBMATE_MIGRATIONS_DIR
    value: "/migrations"
```

**Key points:**
- Uses SSL certificate authentication
- Runs as the `dbmate` PostgreSQL user (created during cluster bootstrap)
- `dbmate` user has `CREATEROLE` privilege to create other database users

#### Running Migrations in Tests

Tests use a custom dbmate runner in `golang/test/dbmate.go` with SQL replacement support:

```go
// Replace hardcoded database names with test database names
replacements := map[string]string{
    "config": "config_83892444",  // Dynamic test database name
}

err := RunDbmateMigrations(ctx, dbURL, migrationsDir, replacements)
```

**How it works:**
1. Test framework creates databases with dynamic names (e.g., `config_83892444`)
2. Migration SQL references hardcoded names (e.g., `config` in GRANT statements)
3. `RunDbmateMigrations` replaces all occurrences before execution
4. Same migration files work in both production and tests

**Test migration flow:**
```go
func createDatabase(ctx, testID, config) (*TestDBContext, error) {
    dbName := fmt.Sprintf("%s_%s", config.database, testID)

    // Replace database name in SQL before running migrations
    replacements := map[string]string{
        string(config.database): dbName,
    }

    err := RunDbmateMigrations(ctx, dbURL, config.migrationsDir, replacements)
}
```

#### Adding New Migrations

**1. Create migration file** in `db/config/migrations/`:
```bash
# Timestamp format: YYYYMMDDHHMMSS
touch db/config/migrations/20250107120000_add_user_profiles.sql
```

**2. Write migration SQL:**
```sql
-- migrate:up
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id BYTEA PRIMARY KEY,
    display_name TEXT NOT NULL
);

-- migrate:down
DROP TABLE IF EXISTS user_profiles;
```

**3. Test locally:**
```bash
# Build and push image
bazel run //db/config:dbmate_push

# Run in Kubernetes
kubectl delete job dbmate -n app-namespace  # Delete old job
kubectl apply -k k8s/app/dbmate/dev

# Watch migration job
kubectl logs -f job/dbmate -n app-namespace
```

**4. Verify tests still pass:**
```bash
bazel test //golang/test:test_test
```

#### Creating Migrations for New Databases

To add migrations for a new database (e.g., "analytics"):

**1. Create directory structure:**
```
db/analytics/
├── BUILD.bazel
└── migrations/
    └── 20250107120000_initial_schema.sql
```

**2. Create BUILD.bazel:**
```python
load("//k8s/infra:dbmate.bzl", "dbmate_image")

filegroup(
    name = "migrations",
    srcs = glob(["migrations/*.sql"]),
    visibility = ["//visibility:public"],
)

dbmate_image(
    name = "dbmate_analytics",
    migrations = [":migrations"],
    repositories = ["registry.localhost"],
)
```

**3. Create Kubernetes Job:**
```yaml
# k8s/app/dbmate-analytics/base/job.yaml
# Similar to dbmate job but with different database URL
```

#### Troubleshooting

**Migration files not found in tests:**
- Ensure `data = ["//db/config:migrations"]` in test BUILD file
- Check migrations path is relative to Bazel runfiles: `../../db/config/migrations`

**Permission errors in Kubernetes:**
- Verify `dbmate` user has correct privileges in PostgreSQL cluster
- Check `postInitApplicationSQL` in `k8s/app/postgres/base/cluster.yaml`

**Database name mismatch in tests:**
- Add replacement mapping in test context provider
- Ensure all database references use the replacement system

### 14. Architecture Principles

1. **Compile-time validation** - Invalid routing won't build
2. **No reflection** - All routing resolved at compile time
3. **Type safety** - Compiler enforces correct handler implementations
4. **Single source of truth** - `routing.yaml` drives everything
5. **Clean separation** - Handlers don't know about routing, messenger handles it
6. **Testability** - Easy to swap implementations via interfaces

---

When making changes to this codebase:
1. Always regenerate after changing `routing.yaml`
2. Always run Gazelle after adding packages
3. Never manually edit BUILD or generated files
4. Test changes with `bazel test //...`
5. Check that both interface-gen and messenger-gen produce valid code
6. When working with containers, ensure local registry is running on port 5001
7. When adding database migrations, ensure tests still pass and migration works in Kubernetes



