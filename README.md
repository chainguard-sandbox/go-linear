# go-linear

Go client library for the Linear API, automatically synchronized with the official [Linear TypeScript SDK](https://github.com/linear/linear).

[![Go Reference](https://pkg.go.dev/badge/github.com/eslerm/go-linear.svg)](https://pkg.go.dev/github.com/eslerm/go-linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/eslerm/go-linear)](https://goreportcard.com/report/github.com/eslerm/go-linear)

> **Work in Progress**: This SDK is under active development. API may change until v1.0.0.

## Installation

```bash
go get github.com/eslerm/go-linear
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/eslerm/go-linear/pkg/linear"
)

func main() {
    client, err := linear.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get authenticated user
    viewer, err := client.Viewer(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Hello, %s!\n", viewer.Name)
}
```

## Features

- ✅ Type-safe GraphQL operations via [genqlient](https://github.com/Khan/genqlient)
- ✅ Context support for cancellation and timeouts
- ✅ Structured error types
- ✅ Follows Go best practices (stdlib-first, minimal dependencies)
- 🚧 Comprehensive API coverage (in progress)
- 🚧 Automatic pagination with iterators (in progress)
- 🚧 Rate limit handling (in progress)

## Development

### Setup

```bash
# Clone repository
git clone https://github.com/eslerm/go-linear
cd go-linear

# Install development tools and dependencies
make dev
```

### Common Tasks

```bash
make help          # Show all available commands
make fmt           # Format code
make lint          # Run linters
make test          # Run tests
make check         # Run all checks (use before commit)
```

## Upstream Synchronization

This SDK is kept in sync with the official [Linear TypeScript SDK](https://github.com/linear/linear):

- Schema synced from upstream `schema.graphql`
- CI monitors for updates weekly
- Version tracking maintained

## License

MIT

## Credits

This SDK follows patterns from the official [Linear TypeScript SDK](https://github.com/linear/linear).
