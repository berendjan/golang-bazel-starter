# Golang-bazel-starter

## Dependency Management

When changing dependencies in go code do the following:
After adding a new golang library:

```bash
# update go.mod
go mod tidy -e

# update bazel go deps
bazel mod tidy

# update BUILD files
bazel run //:gazelle

# build target with bazel
```

Do not manually update BUILD.bazel files, always do
```bash
bazel run //:gazelle
```
Read the gazelle directives in BUILD.bazel

## Message Routing Code Generator

This project includes a compile-time code generator for type-safe message routing, similar to Rust's declarative macros.

### Location
- Generator tool: `tools/codegen/messenger-gen/`
- Example spec: `golang/grpcserver/messenger/messenger.yaml`

### How it works

1. Define a YAML specification with handlers and message routes
2. Bazel genrule runs the generator at build time
3. Generated code provides type-safe message routing

### YAML Specification Format

```yaml
package: main
messenger_name: GrpcMessenger

# Go imports needed for the generated code
imports:
  - 'alias "github.com/your/package"'
  - '"github.com/another/package"'

# Handlers that process messages
handlers:
  - name: handler_field_name
    type: "*YourHandlerType"

# Routes define which messages go to which handlers
routes:
  - source: "SourceType"
    message: "MessageType"
    response: "ResponseType"  # optional, omit for no response
    receivers:
      - handler_field_name
```

### Generated Code

The generator creates:
- `Sender[M, R]` interface for type-safe message sending
- `MessengerRoute[M, R]` interface for routing
- Messenger struct with handler fields
- Constructor function
- Routing methods that call handlers in sequence

### Usage Example

After generation, use like this:

```go
// Create messenger with your handlers
messenger := NewGrpcMessenger(
    myRepository,
    myApi,
)

// Send messages (type-safe!)
sender := GrpcService_CreateAccountRequest_Sender{}
result, err := sender.Send(ctx, myMessage, messenger)
```

### Handler Requirements

Your handler types must implement a `Handle` method:

```go
func (h *YourHandler) Handle(ctx context.Context, message MessageType) (ResponseType, error) {
    // Handle the message
    return result, nil
}
```

### Building Generated Code

```bash
# Build the generator
bazel build //tools/codegen/messenger-gen

# Build messenger (automatically runs generator)
bazel build //golang/grpcserver/messenger

# The generator runs automatically as part of the build
```



