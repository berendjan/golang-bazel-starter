# Golang-bazel-starter - AI Assistant Guide

This document provides guidance for AI assistants (like Claude) working on this codebase.

## Project Overview

This is a production-ready Go + Bazel template featuring:
- Type-safe code generation for message routing
- gRPC service with middleware support
- PostgreSQL with database migrations
- Testcontainers for integration testing
- Compile-time validation of routing configuration

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

### 12. Architecture Principles

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



