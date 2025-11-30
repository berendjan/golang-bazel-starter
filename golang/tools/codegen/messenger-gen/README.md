# Messenger Code Generator

A compile-time code generator for type-safe message routing in Go, inspired by Rust's declarative macros.

## Overview

This tool generates boilerplate code for routing messages between handlers in a type-safe manner. It reads a YAML specification and produces Go code with:

- Type-safe interfaces for sending messages
- A messenger struct that holds handler instances
- Routing methods that forward messages to configured handlers
- Support for multiple receivers per message (chain of responsibility pattern)

## How It Works

1. You define handlers and message routes in a YAML file
2. The Bazel genrule executes this generator during build
3. Generated Go code is compiled into your binary
4. You get compile-time type safety for message routing

## YAML Specification

```yaml
package: main                    # Go package name for generated code
messenger_name: MyMessenger      # Name of the messenger struct

imports:                         # Go imports (use quotes appropriately)
  - '"github.com/your/pkg"'
  - 'alias "github.com/pkg"'

handlers:                        # Handler instances in the messenger
  - name: repository             # Field name in messenger struct
    type: "*Repository"          # Go type (must support Handle method)

routes:                          # Message routing definitions
  - source: "Service"            # Source type (for generated sender name)
    message: "CreateRequest"     # Message type
    response: "error"            # Response type (optional)
    receivers:                   # Handlers that process this message
      - repository               # Must match a handler name
```

## Generated Code Structure

### Interfaces

```go
type Sender[M any, R any] interface {
    Send(ctx context.Context, message M, messenger *YourMessenger) (R, error)
}

type MessengerRoute[M any, R any] interface {
    Route(ctx context.Context, message M) (R, error)
}
```

### Messenger Struct

```go
type YourMessenger struct {
    Repository *Repository
    // ... other handlers
}
```

### Routing Logic

For each route, generates:
- A sender type implementing `Sender[MessageType, ResponseType]`
- A routing method that calls handlers in sequence
- Proper error handling and result propagation

## Integration with Bazel

In your BUILD.bazel:

```python
genrule(
    name = "generate_messenger",
    srcs = ["messenger.yaml"],
    outs = ["generated_messenger.go"],
    cmd = "$(location //golang/tools/codegen/messenger-gen) -input=$(location messenger.yaml) -output=$@",
    tools = ["//golang/tools/codegen/messenger-gen"],
)

go_library(
    name = "messenger",
    srcs = [":generate_messenger"],
    deps = [
        # Add deps for imported packages
    ],
)
```

## Handler Interface

Your handler types must implement:

```go
func (h *YourHandler) Handle(ctx context.Context, message MessageType) (ResponseType, error) {
    // Process message
    return response, nil
}
```

## CLI Usage

```bash
messenger-gen -input=spec.yaml -output=generated.go
```

Flags:
- `-input`: Path to YAML specification file (required)
- `-output`: Path to output Go file (required)

## Features

- **Type Safety**: Compile-time verification of message types
- **Multiple Receivers**: Route one message to multiple handlers
- **Error Handling**: Stops routing on first error
- **Result Propagation**: Returns result from first handler (if multiple)
- **Clean Separation**: Generated code separate from business logic

## Example

See `golang/grpcserver/messenger/messenger.yaml` for a complete example.
