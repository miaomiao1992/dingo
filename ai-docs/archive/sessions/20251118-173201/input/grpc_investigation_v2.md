Investigate common idiomatic patterns for organizing protoc-generated gRPC stubs in Go projects using *internal search capabilities*. Focus on:
1. Where original `.proto` schema files live (e.g., `api/proto`, `proto`).
2. Where generated Go files are emitted (e.g., `pkg/pb`, `internal/gen`, `api/proto`).
3. `go.mod` / module boundary considerations when committing generated code.
4. Tooling ergonomics (e.g., `makefiles`, `go:generate` directives, `buf.gen.yaml`).

Provide concrete repository examples (URLs) and directory paths from *real-world open-source projects*. Examples should include projects from `google.golang.org/grpc`, `buf.build`, or general `grpc-go` examples. Limit response to ~200 words.
