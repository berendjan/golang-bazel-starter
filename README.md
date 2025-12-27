# Bazel Golang Starter

A production-ready Go + Bazel template with type-safe code generation for gRPC services, middleware, and message routing.

## Features

- **🏗️ Bazel Build System** - Fast, reproducible builds with dependency management
- **🔧 Code Generation** - Type-safe interfaces and message routing from YAML specifications
- **🚀 gRPC Server** - Dual gRPC/HTTP server with automatic service registration
- **🔄 Message Routing** - Declarative middleware chain with compile-time validation
- **🧪 Testing Framework** - Testcontainers integration for database testing
- **📦 Modular Architecture** - Clean separation of concerns with dependency injection

## Quick Start

```bash
# Clone the repository
git clone https://github.com/your-org/golang-bazel-starter
cd golang-bazel-starter

# Update dependencies
go mod tidy && bazel mod tidy

# Build everything
bazel build //...

# Run tests (requires docker running locally)
# On first testrun it is possible that two tests will try to 
# start the testcontainer at the same time at the same port.
# Then just restart the tests.
bazel test //golang/test:test_test

# Start the server
bazel run //golang/grpcserver:grpcserver
```

## Architecture

### Project Structure

```
golang-bazel-starter/
├── golang/
│   ├── config/              # Configuration service
│   │   ├── api/            # gRPC API implementation
│   │   ├── repository/     # Database repositories
│   │   └── db/             # Database migrations
│   ├── middleware/          # Message middleware
│   │   ├── middleone/      # Example middleware
│   │   └── middletwo/      # Example middleware
│   ├── generated/           # Code generation
│   │   ├── routing.yaml    # Message routing specification
│   │   └── interfaces/     # Generated interfaces
│   ├── grpcserver/         # Main gRPC server
│   │   └── messenger/      # Generated messenger
│   ├── framework/          # Shared framework code
│   └── test/               # Testing infrastructure
├── proto/                  # Protocol buffer definitions
├── tools/codegen/          # Code generators
│   ├── interface-gen/      # Interface generator
│   └── messenger-gen/      # Messenger generator
└── BUILD.bazel, MODULE.bazel, etc.
```

### Message Routing System

The core innovation is a **compile-time message routing system** that generates type-safe interfaces and routing logic from a YAML specification.

#### Single Source of Truth: `golang/generated/routing.yaml`

```yaml
# Interface generation configuration
interfaces:
  package: interfaces
  imports:
    - 'commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"'
    - 'configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"'

# Messenger generation configuration
messenger:
  package: messenger
  messenger_name: GrpcMessenger
  imports:
    - 'geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"'
    - 'commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"'
    - 'configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"'

# Handler definitions
handlers:
  - name: accountRepository
    type: "configrepository.AccountDbRepository"
  - name: AccountApi
    type: "configapi.ConfigurationApi"
  - name: middlewareOne
    type: "middleone.MiddleOne"
  - name: middlewareTwo
    type: "middletwo.MiddleTwo"

# Message routing graph
routes:
  - source: AccountApi
    messages:
      - message: "*configpb.MiddleOneRequestProto"
        response: "(*configpb.AccountConfigurationProto, error)"
        receivers:
          - middlewareOne  # First receiver

  - source: middlewareOne
    messages:
      - message: "*configpb.MiddleOneRequestProto"
        response: "(*configpb.AccountConfigurationProto, error)"
        receivers:
          - middlewareTwo   # Intermediate receiver (returns error only)
          - accountRepository  # Final receiver (returns full response)
```

#### What Gets Generated

**1. Type-Safe Interfaces** (`golang/generated/interfaces/`)

```go
// Generated handler interfaces
type AccountRepositoryInterface interface {
    HandleMiddleOneRequest(ctx, message) (*AccountConfigurationProto, error)
}

type MiddlewareOneInterface interface {
    HandleMiddleOneRequest(ctx, message, next MiddlewareOneSendable) (*AccountConfigurationProto, error)
}

type MiddlewareTwoInterface interface {
    // Intermediate receiver - returns error only
    HandleMiddleOneRequest(ctx, message, next MiddlewareTwoSendable) error
}

// Generated sendable interfaces
type AccountApiSendable interface {
    SendMiddleOneRequestFromAccountApi(ctx, message) (*AccountConfigurationProto, error)
}
```

**2. Message Router** (`golang/grpcserver/messenger/`)

```go
// Generated messenger with type-safe routing
type GrpcMessenger struct {
    accountRepository geninterfaces.AccountRepositoryInterface
    middlewareOne     geninterfaces.MiddlewareOneInterface
    middlewareTwo     geninterfaces.MiddlewareTwoInterface
}

// Automatically chains middleware with error handling
func (m *GrpcMessenger) SendMiddleOneRequestFromMiddlewareOne(ctx, message) (*AccountConfigurationProto, error) {
    // Call intermediate receiver, check error
    if err := m.middlewareTwo.HandleMiddleOneRequest(ctx, message, m); err != nil {
        return nil, err
    }
    // Call final receiver, return result
    return m.accountRepository.HandleMiddleOneRequest(ctx, message)
}
```

#### Key Features

- **Compile-time type safety** - Invalid routes won't compile
- **Automatic middleware chaining** - Intermediate handlers return `error`, final handlers return `(result, error)`
- **No reflection** - All routing resolved at compile time
- **Single source of truth** - Change `routing.yaml`, both interfaces and messenger update
- **Unknown handler validation** - Build fails if routes reference non-existent handlers

## Code Generators

### Interface Generator (`tools/codegen/interface-gen`)

Generates Go interfaces from `routing.yaml`:
- Handler interfaces (what each component must implement)
- Sendable interfaces (what each component can send)
- Automatically determines method signatures based on receiver position

### Messenger Generator (`tools/codegen/messenger-gen`)

Generates message routing implementation:
- Creates messenger struct with handler fields
- Generates routing methods with automatic middleware chaining
- Handles error propagation through middleware chain

### Usage

Both generators run automatically during build:

```bash
# Regenerate interfaces
bazel build //golang/generated/interfaces:interfaces

# Regenerate messenger
bazel build //golang/grpcserver/messenger:messenger

# Or regenerate all
bazel run //:gazelle && bazel build //...
```

## Development Workflow

### Adding a New Message Route

1. **Update `routing.yaml`**:
   ```yaml
   - source: middlewareOne
     messages:
       - message: "*configpb.NewMessageProto"
         response: "(*configpb.NewResponseProto, error)"
         receivers:
           - middlewareTwo
           - accountRepository
   ```

2. **Regenerate code**:
   ```bash
   bazel build //golang/generated/interfaces:interfaces
   bazel build //golang/grpcserver/messenger:messenger
   ```

3. **Implement handler methods** - Compiler will tell you what's missing!

### Adding New Middleware

1. **Create middleware package**:
   ```bash
   mkdir -p golang/middleware/mynewmiddleware
   ```

2. **Implement the generated interface**:
   ```go
   type MyNewMiddleware struct{}

   func (m *MyNewMiddleware) HandleSomeRequest(ctx, req, next) error {
       // Your middleware logic
       log.Printf("Processing: %+v", req)
       return nil  // Continue to next handler
   }
   ```

3. **Add to `routing.yaml`**:
   ```yaml
   handlers:
     - name: myNewMiddleware
       type: "mynewmiddleware.MyNewMiddleware"

   routes:
     - source: someSource
       messages:
         - message: "*configpb.SomeRequest"
           receivers:
             - myNewMiddleware
             - accountRepository
   ```

4. **Update main.go to wire it up**:
   ```go
   myMiddleware := &mynewmiddleware.MyNewMiddleware{}
   messenger := messenger.NewGrpcMessenger(repo, myMiddleware, ...)
   ```

### Dependency Management

When adding new Go packages:

```bash
# Add to go.mod
go mod tidy -e

# Update Bazel dependencies
bazel mod tidy

# Update BUILD files (do NOT edit manually! (except for ones that have '# gazelle:ignore' directive))
bazel run //:gazelle

# Build and verify
bazel build //...
```

**Important**: Never manually edit `BUILD.bazel` files - always use Gazelle! (except for generation files starting with `# gazelle:ignore`)

## Testing

### Unit Tests

```bash
# Run all tests
bazel test //...

# Run specific test
bazel test //golang/test:test_test

# Run with verbose output
bazel test //golang/test:test_test --test_output=all
```

### Test Infrastructure

The project includes a robust testing framework with:

- **TestContainers** - Spin up real PostgreSQL databases for tests
- **Test Context Providers** - Dependency injection for tests
- **Shared test databases** - Reuse containers across tests
- **Database migrations** - Automatic schema setup per test

Example test structure:

```go
func TestSomething(t *testing.T) {
    // Build test context with database
    tc := test.NewTestContext(t).
        WithDatabase(test.ConfigDb).
        Build()
    defer tc.Cleanup()

    // Get dependencies
    provider := tc.GetProvider()
    accountRepo := provider.GetAccountRepository()

    // Test your code
    result, err := accountRepo.HandleSomeRequest(ctx, req)
    assert.NoError(t, err)
}
```

## Building and Running

### Development

```bash
# Build everything
bazel build //...

# Run server in development mode
bazel run //golang/grpcserver:grpcserver

# The server runs on:
# - gRPC: localhost:50051
# - HTTP: localhost:8080
```

### Production

```bash
# Build optimized binary
bazel build -c opt //golang/grpcserver:grpcserver

# Binary is at:
# bazel-bin/golang/grpcserver/grpcserver_/grpcserver

# Run production binary
./bazel-bin/golang/grpcserver/grpcserver_/grpcserver
```

### Container Images

The project includes Bazel rules for building and pushing OCI container images.

#### Prerequisites

Start a local Docker registry (v3) on port 5001:

```bash
# Start a local registry (runs on localhost:5001)
docker run -d -p 5001:5000 --name registry registry:3

# Verify it's running
curl http://localhost:5001/v2/_catalog
```

#### Building Images

```bash
# Build the container image
bazel build //golang/grpcserver:grpcserver_image

# Build cross-compiled binary for Linux
bazel build //golang/grpcserver:grpcserver_cross
```

#### Pushing Images

```bash
# Push image to local registry (both sha256 tag and latest tag)
bazel run //golang/grpcserver:grpcserver_push

# Push only sha256 tag
bazel run //golang/grpcserver:grpcserver_localhost_push_sha256_tag

# Push only latest tag
bazel run //golang/grpcserver:grpcserver_localhost_push_latest_tag
```

#### Verifying Pushed Images

```bash
# List all repositories in registry
curl http://localhost:5001/v2/_catalog

# List tags for specific image
curl http://localhost:5001/v2/grpcserver/tags/list

# Get image manifest
curl http://localhost:5001/v2/grpcserver/manifests/latest
```

#### Running Docker Images Locally

```bash
# Pull and run from local registry
docker run --rm -p 25000:25000 -p 26000:26000 localhost:5001/grpcserver:latest

# Or run with specific tag
docker run --rm -p 25000:25000 -p 26000:26000 localhost:5001/grpcserver:<sha256-tag>

# Run in detached mode
docker run -d -p 25000:25000 -p 26000:26000 --name grpcserver localhost:5001/grpcserver:latest

# View logs
docker logs grpcserver

# Stop and remove
docker stop grpcserver && docker rm grpcserver
```

#### Image Build Details

The `go_binary` macro in `k8s/infra/server.bzl` automatically:
1. Builds the Go binary
2. Cross-compiles for Linux AMD64
3. Packages into a distroless base image
4. Exposes ports 25000 (gRPC) and 26000 (HTTP)
5. Creates push targets with both sha256 and latest tags

## Customization

### Adapting for Your Project

1. **Update package paths**:
   - Find and replace `github.com/berendjan/golang-bazel-starter` with your module path
   - Update in `go.mod`, `MODULE.bazel`, and `routing.yaml`

2. **Replace example services**:
   - Remove/rename `config` service
   - Update proto definitions in `proto/`
   - Regenerate routing config

3. **Customize middleware**:
   - Remove example middleware (`middleone`, `middletwo`)
   - Add your own middleware implementations
   - Update `routing.yaml` with your message flow

4. **Update database migrations**:
   - Modify migrations in `golang/config/db/migrations/`
   - Update repository implementations

## Advanced Features

### Generic-Free Messenger

The messenger uses **interface-only** design (no generics) for:
- Simpler type signatures
- Easier to read and maintain
- Same compile-time type safety
- Zero runtime overhead

### Automatic Error Handling

Middleware chains automatically handle errors:
```go
// Generated code handles this automatically:
if err := middleware1.Handle(ctx, msg, m); err != nil {
    return nil, err  // Early return on error
}
if err := middleware2.Handle(ctx, msg, m); err != nil {
    return nil, err
}
return repository.Handle(ctx, msg)  // Final handler
```

### Compile-Time Validation

The code generators validate:
- All handlers referenced in routes exist
- All message types are consistent
- Receiver chains are valid
- Interface implementations are complete

## Troubleshooting

### Build Issues

```bash
# Clean build cache
bazel clean

# Deep clean
bazel clean --expunge

# Rebuild from scratch
bazel build --nobuild
bazel build //...
```

### Regeneration Issues

```bash
# Force regeneration
bazel build --nocache_test_results //golang/generated/interfaces:interfaces
bazel build --nocache_test_results //golang/grpcserver/messenger:messenger

# Update all BUILD files
bazel run //:gazelle
```

### Import Errors

If you see undefined imports after adding packages:

```bash
# Update dependencies
go mod tidy
bazel mod tidy

# Regenerate BUILD files
bazel run //:gazelle

# Rebuild
bazel build //...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `bazel test //...`
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Frontend Development

The project includes a React frontend built with Vite, Tailwind CSS v4, and managed through Bazel.

### Features

- **React 18** with TypeScript
- **Tailwind CSS v4** with native Vite plugin (no PostCSS needed)
- **Ory Kratos** integration for authentication
- **Vite** for fast dev server and bundler
- **ESLint** with TypeScript support
- **Vitest** for unit testing

### Project Structure

```
golang-bazel-starter/
├── package.json            # Root workspace dependencies (eslint, vitest, etc.)
├── pnpm-lock.yaml          # Lock file for all npm packages
├── pnpm-workspace.yaml     # Defines frontend/react as workspace package
├── eslint.config.mjs       # ESLint v9 flat config (must be at root)
├── frontend/
│   ├── react/              # React application
│   │   ├── src/            # Source files (.tsx, .ts)
│   │   │   ├── pages/      # Page components (Login, Register, Dashboard)
│   │   │   └── lib/        # Kratos client and utilities
│   │   ├── public/         # Static assets
│   │   ├── package.json    # React app dependencies
│   │   ├── tsconfig.json   # TypeScript config
│   │   ├── vite.config.mjs # Vite configuration (ESM for Tailwind plugin)
│   │   └── BUILD.bazel     # Bazel build rules
│   └── tools/
│       ├── lint/           # ESLint Bazel configuration
│       ├── vitest/         # Test runner config
│       └── pnpm            # Bazel-managed pnpm script
```

### Quick Start

```bash
# Install npm dependencies
./frontend/tools/pnpm install

# Start development server with hot reload
ibazel run //frontend/react:start

# Build production bundle (linting runs automatically)
bazel build //frontend/react:build

# Run tests
bazel test //frontend/react/src:test

# Run linting only
bazel test //frontend/react/src:lint
```

### Adding npm Dependencies

```bash
# Go to the webpage directory
cd frontend/react

# Add a dependency
../tools/pnpm add tailwindcss

# Add a dev dependency
../tools/pnpm add -D @types/lodash

# Then add to BUILD.bazel deps (in ts_project that needs it)
# Example: "//frontend/react:node_modules/tailwindcss"
```

### Removing npm Dependencies

```bash
# Go to the webpage directory
cd frontend/react

# Remove a dependency
../tools/pnpm remove tailwindcss

# Remove from BUILD.bazel deps as well
```

### Adding New Pages/Components

1. **Create component file** in `frontend/react/src/`:
   ```tsx
   // frontend/react/src/MyPage.tsx
   export function MyPage() {
     return <div>My New Page</div>;
   }
   ```

2. **Import and use** in your app:
   ```tsx
   // frontend/react/src/App.tsx
   import { MyPage } from './MyPage';
   ```

3. **Build and test**:
   ```bash
   bazel build //frontend/react:build
   ibazel run //frontend/react:start
   ```

### Production Build

```bash
# Build optimized bundle (linting runs automatically)
bazel build //frontend/react:build

# Output is at: bazel-bin/frontend/react/dist/
```

### Linting

ESLint v9 with TypeScript support runs automatically during `bazel build`. Config is at root (`eslint.config.mjs`).

```bash
# Run linting only
bazel test //frontend/react/src:lint
```

### Testing

Uses Vitest for unit testing:

```bash
# Run tests
bazel test //frontend/react/src:test

# Run with verbose output
bazel test //frontend/react/src:test --test_output=all
```

## Resources

- [Bazel Documentation](https://bazel.build/)
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Vite](https://vitejs.dev/)
- [React](https://react.dev/)

## Support

- Issues: [GitHub Issues](https://github.com/your-org/golang-bazel-starter/issues)
- Discussions: [GitHub Discussions](https://github.com/your-org/golang-bazel-starter/discussions)