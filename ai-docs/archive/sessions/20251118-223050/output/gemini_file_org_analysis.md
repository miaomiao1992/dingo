# Dingo File & Module Organization Recommendations

## 1. Current Repository Observations
- **Source layout:** Root directories such as `/Users/jack/mag/dingo/cmd`, `/pkg`, and `/internal` follow standard Go packaging, while `.dingo` transpiler inputs primarily live under `/tests/golden/`, paired with `.go.golden` outputs for golden testing (e.g., `error_prop_01_simple.dingo` ↔ `error_prop_01_simple.go.golden`).
- **Generated artifacts:** Golden outputs are committed next to their sources, and tooling support (CLI, LSP, VS Code extension) is separated into `/cmd/dingo`, `/cmd/dingo-lsp`, and `/editors/`.
- **Implication:** The core repo already distinguishes authored Dingo sources, generated Go, and tooling, but end-user guidance on real project layout (where to place transpiled Go) is still open.

## 2. Reference Patterns from Go Ecosystems
### Protobuf/gRPC & Buf
- `grpc-go` example services store `.proto` definitions with the service (e.g., `examples/helloworld/helloworld/helloworld.proto`) and emit `helloworld.pb.go` + `helloworld_grpc.pb.go` in the same import path, keeping API versioning via directories like `api/v1` ([github.com/grpc/grpc-go](https://github.com/grpc/grpc-go/tree/master/examples/helloworld/helloworld)).
- Buf starter templates keep schemas in `proto/` and direct generated Go into `gen/go/<module>` via `buf.gen.yaml`, often under `internal/gen` to hide implementation ([github.com/bufbuild/buf-tour](https://github.com/bufbuild/buf-tour/tree/main/module-1)).
- Tooling: Make targets (`make proto`), `buf generate`, and language-specific `go_package` options ensure generated packages line up with `go.mod` import paths.

### templ
- Each `.templ` file (e.g., `examples/counter/counter.templ`) emits a sibling `counter.templ.go`, keeping generated code co-located for seamless `go build` ([github.com/a-h/templ/examples](https://github.com/a-h/templ/tree/main/examples)).
- Tooling relies on `//go:generate templ generate` or `templ generate ./...`, so developers rarely need a separate module; generated files stay checked in (or regenerated on demand) but always share the same package namespace.

### sqlc
- SQL schema files live in `db/schema` or `db/migrations` and queries in `db/queries`; `sqlc.yaml` maps them into Go packages like `internal/db` or `pkg/sqlc` (see [github.com/sqlc-dev/sqlc/tree/main/examples/chinook](https://github.com/sqlc-dev/sqlc/tree/main/examples/chinook)).
- Generated Go is typically committed because it expresses database contract types; Make targets (`make sqlc`) keep regeneration deterministic.

### ent (entgo)
- Author-written schemas reside in `ent/schema/` while running `go generate ./ent` produces the entire `ent/` client, migrations, and predicate helpers (e.g., [github.com/ent/ent/tree/master/examples/start](https://github.com/ent/ent/tree/master/examples/start)).
- `ent/generate.go` hosts `//go:generate go run entgo.io/ent/cmd/ent generate ./ent/schema`, ensuring a single command syncs generated code with schemas.

### wire
- Dependency injection specs (`wire.Build`) live in files such as `cmd/server/wire.go`, and generated `wire_gen.go` stays in the same directory/package (see [github.com/google/wire/tree/master/_examples/struct](https://github.com/google/wire/tree/master/_examples/struct)).
- `//go:generate wire` comments make DI outputs reproducible without separate modules.

## 3. Common Idioms Across Tools
1. **Directory pairing:** Keep authored schemas/templates (.proto, .templ, .sql, ent schema Go files) adjacent to generated results, often within the same package or under `internal/gen` when outputs should remain private.
2. **Versioned API paths:** Use `api/<feature>/vN` (Protobuf) or `internal/db/<version>` (sqlc) so regeneration doesn’t disturb stable imports.
3. **`go:generate` hooks:** Every tool adds a one-line directive that documents how to rebuild generated code; Makefiles simply orchestrate multiple generators.
4. **Module boundaries:** Public contracts (e.g., pb clients) go under `pkg/` so downstream modules can import them; private plumbing lives under `internal/` to avoid accidental dependencies.
5. **Naming conventions:** Suffixes like `.pb.go`, `.templ.go`, `_gen.go` warn contributors not to edit by hand and simplify `.gitignore` patterns when regeneration is cheap.

## 4. Recommendations Tailored to Dingo
1. **Adopt a dual-root layout for user projects:**
   - Author `.dingo` in `src/` (or feature folders) and emit Go into `internal/gen/dingo` by default, mirroring Buf/sqlc’s private-output pattern. Provide a flag (e.g., `--out pkg/dingo`) when users want reusable packages.
   - Maintain `.dingo` next to any `.go` shims (like `main.go`) to mirror templ/wire co-location, enabling IDEs to reason about both files.
2. **Codify output naming:** Generate `<file>.dingo.go` (or `.gen.go`) so Go tooling can watch for deletions and so contributors instantly recognize derivative files. Update golden tests to follow the same suffix for clarity when they move out of `/tests/golden` and into real sample apps.
3. **Introduce `go:generate dingo build` entrypoints:** Ship scaffolding comments (e.g., `//go:generate dingo build ./src --out internal/gen/dingo`) in starter templates. This matches ent/sqlc habits and gives editors a canonical regeneration path.
4. **Offer Buf-style config (`dingo.gen.toml`):** Let teams declare multiple source/output mappings—e.g., `[[targets]] src="apps/auth" out="internal/gen/auth"`—so complex repos can segregate generated Go per service the way buf/sqlc split API surfaces.
5. **Clarify module strategy in docs:**
   - Recommend keeping generated Go inside the same module when it is an implementation detail, but allow publishing a secondary module (e.g., `github.com/org/project-dingo-api`) when transpiled code represents a public contract, similar to how Protobuf APIs expose `pkg/pb`.
   - Document when to commit outputs (default: yes, for reproducibility) versus rely on CI regeneration.
6. **Tooling ergonomics:**
   - Extend `cmd/dingo` with `dingo fmt` / `dingo clean` commands that mirror protoc/sqlc Make targets (e.g., `make dingo` invoking `dingo build ./src`).
   - For IDE parity, align VS Code tasks with templ’s auto-regeneration: watch `.dingo` files and re-run `dingo build` into the configured `internal/gen` directory.
7. **Testing guidance:** Encourage a `tests/golden/<feature>` template but emphasize that production repos should keep `.dingo` + transpiled `.go` under the service directory (e.g., `services/billing/src/invoice.dingo` & `services/billing/internal/gen/invoice.dingo.go`).

## 5. Next Steps for Dingo Documentation
- Add a "Project Layout" section to `README.md` (or `docs/architecture/`) with concrete trees similar to gRPC/sqlc examples above.
- Provide new starter repo(s) that match these patterns, ensuring `go.mod` includes both handwritten Go (entrypoints) and generated Go (from `dingo build`).
- Mirror Buf’s reproducibility story by pinning Dingo CLI versions in Makefiles or `tools.go`, so regeneration stays deterministic across CI and local machines.

Applying these conventions will make Dingo-generated projects feel familiar to Go developers who already rely on protoc, templ, sqlc, ent, and wire, lowering adoption friction while keeping module hygiene tight.
