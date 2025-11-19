# Organizing Source vs. Generated Code in Go Projects

Go projects frequently leverage code generation to simplify development, enforce best practices, and integrate with various protocols and data layers. Managing the interplay between manually written source code and automatically generated code is a critical architectural concern. This report will explore common patterns and considerations using prominent Go code generation tools.

### 1. Protocol Buffers (protoc/gRPC) and Buf

**Purpose**: Define language-agnostic data structures and service interfaces for RPC.

**Schema Location**: Protocol Buffer `.proto` files are typically co-located with the services that define them, often under `api/`, `proto/`, or directly within a service's package (e.g., `pkg/myapp/api/v1/`).

**Generated Go Emission**: The Go output is commonly placed in:
- `pkg/pb/`: A dedicated package for all generated protocol buffer code.
- `internal/gen/pb/`: If the generated code is considered an internal implementation detail.
- Directly alongside the `.proto` files within a versioned API path, such as `pkg/myapp/api/v1/`.

**Module/go.mod Considerations**: Generated code should reside within the main module or a separate module if it's intended for reuse by distinct services or external consumers. If in the main module, the `go.mod` file needs to reference the generated packages. If in a separate module, clients will import that module.

**Tooling Patterns**:
- `Makefile`: Used to orchestrate `protoc` or `buf` commands, often involving flags for Go module paths and output directories. Example: `protoc --go_out=. --go-grpc_out=. api/v1/*.proto`.
- `go generate`: Less common for full `protoc` runs due to complexity, but can be used for smaller, self-contained generation tasks within a single package.

**Concrete Example**: Google Cloud Go Libraries (`google.golang.org/genproto`). While a large monorepo, it demonstrates the pattern of storing `.proto` files outside the main Go source, and generating Go code into distinct Go modules within the repository.

### 2. Templ

**Purpose**: Generate Go HTML components from `.templ` files, offering compile-time type safety for web UIs.

**Schema Location**: `.templ` files are typically located alongside `.go` files in the same packages where the components are defined, such as `web/components/` or `views/`.

**Generated Go Emission**: Generated Go files have a `.go` extension (e.g., `component.templ.go`) and reside in the *same directory* as their corresponding `.templ` source files. This is a deliberate design choice by `templ` to keep generated code close to its source and enable seamless Go tooling integration.

**Module/go.mod Considerations**: The generated files become part of the existing Go module. No special `go.mod` entries are required beyond the `templ` dependency itself.

**Tooling Patterns**:
- `go generate`: `templ` strongly advocates for `go:generate`. A common pattern is to add a `//go:generate templ generate` comment to a `main.go` or `doc.go` file at the root of the UI package, triggering generation on `go generate ./...`.
- IDE Integration: `templ` integrates with Go language servers to automatically generate code on file save.

**Concrete Example**: `github.com/a-h/templ/examples`. You'll find `.templ` and `.templ.go` files side-by-side.

### 3. SQLC

**Purpose**: Generate type-safe Go code from SQL queries against a database schema.

**Schema Location**: SQL schema files (`.sql`) defining tables, views, and functions are typically found in `db/schema/` or `migrations/`. SQL query files (`.sql`) are often in `db/queries/`.

**Generated Go Emission**: Output is conventionally placed in `pkg/sqlc/` or `internal/db/`. Some projects might use `internal/gen/sqlc/`.

**Module/go.mod Considerations**: The generated code is part of the main application module and imported as needed.

**Tooling Patterns**:
- `Makefile`: A common target like `make sqlc` runs the `sqlc generate` command.
- `sqlc.yaml`: A configuration file specifies source paths for SQL files, output directories, and package names for the generated Go code.

**Concrete Example**: `github.com/sqlc-dev/sqlc-tutorial`. The examples show `sqlc.yaml` configuring output to `tutorial/db`.

### 4. Ent

**Purpose**: An entity framework for Go, generating code for models, migrations, and queries based on a graph schema.

**Schema Location**: Ent schemas are Go files, typically found in a dedicated `ent/schema/` directory where each file defines an `ent.Schema` struct.

**Generated Go Emission**: Ent generates its code into its own dedicated `ent/` directory, creating subdirectories like `ent/generated/`, `ent/migrate/`, etc. This centralizes all ORM-related code.

**Module/go.mod Considerations**: Similar to `sqlc`, the generated `ent` client and models become an integral part of the main application module.

**Tooling Patterns**:
- `go generate`: `ent` heavily relies on `go:generate`. A `//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema` comment in `ent/generate.go` is standard.
- `Makefile`: Can wrap the `go generate` command or other `ent` specific commands.

**Concrete Example**: `github.com/ent/ent/examples`. The `ent/` directory contains generated code alongside schema definitions.

### 5. Wire

**Purpose**: Compile-time Dependency Injection for Go.

**Schema Location**: Wire doesn't have schema files in the traditional sense. Instead, it relies on Go source files that define provider sets and calls to `wire.Build` to specify dependencies.

**Generated Go Emission**: Generated Go files are typically named `wire_gen.go` and are placed in the *same package* as the `wire.Build` invocations. For example, `cmd/server/wire.go` might have a corresponding `cmd/server/wire_gen.go`.

**Module/go.mod Considerations**: The generated code is local to the package where `wire.Build` is used.

**Tooling Patterns**:
- `go generate`: The `//go:generate wire` comment is standard in the file containing `wire.Build` definitions.

**Concrete Example**: `github.com/google/wire/example`. Shows `wire.go` alongside `wire_gen.go`.

### Common Idioms and Best Practices

-   **`internal/`**: The Go compiler enforces that packages within `internal/` cannot be imported by code outside the module. This is an excellent place for generated code (e.g., `internal/gen/pb`, `internal/db`) that should not be directly consumed by external users of your module.
-   **`pkg/`**: Publicly importable packages, suitable for generated code intended for reuse within the module and potentially by other modules (e.g., `pkg/pb` for shared API contracts).
-   **`cmd/`**: Contains main packages for executables. Generated code for DI (like `wire_gen.go`) is often found here, specific to a particular executable.
-   **`go:generate`**: Widely adopted for triggering code generation, integrated directly into the Go toolchain. It promotes a clear contract between source and generated files.
-   **`Makefile`**: Used for more complex generation orchestrations, especially when multiple steps or custom tool invocations are required.
-   **Clear Separation**: Always ensure a clear distinction between source and generated files, either by directory structure (`gen/` subdirectories) or by file naming conventions (`.pb.go`, `.templ.go`, `.sql.go`, `_gen.go`). This prevents accidental manual edits of generated code.
-   **Git Ignore**: `.gitgnore` rules often include generated files that can be fully regenerated, reducing source control bloat (e.g., `ent/`, `*.pb.go` if output to ephemeral location).

Effective management of generated code is crucial for maintainable and scalable Go projects, ensuring that the benefits of automation outweigh the complexities of a multi-source codebase.