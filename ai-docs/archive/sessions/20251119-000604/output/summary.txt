# Package-Wide Unqualified Import Inference Architecture

## 1. Proposed Architecture
- **Components**
  - `pkg/preprocessor/importtracker.PackageIndex`: central object built per Go package before file transpilation. Tracks local symbols and drives exclusion list used by `ImportTracker`.
  - `pkg/preprocessor/importscanner.Scanner`: lightweight parser that tokenizes `.dingo` files (pre-preprocessor) to collect top-level function names, method receivers, and exported helpers. Avoids full AST by scanning for `fn`/`func`/`let` declarations.
  - `cmd/dingo/build.PackagePlanner`: orchestrator that, given a package directory, (1) discovers `.dingo` files, (2) materializes the `PackageIndex`, (3) hands it to the existing preprocessing pipeline.
  - `pkg/preprocessor/pipeline.Context`: extended to carry `*PackageIndex` reference so every processor (especially `ImportTracker`) can consult shared metadata.
- **Data Flow**
  1. Planner collects file list → invokes `PackageIndex.Build(files)`.
  2. `importscanner.Scanner` runs on each file, producing `FileSymbols` (functions, methods, alias imports).
  3. `PackageIndex` merges into `PackageSymbols` map keyed by identifier, with source file + span metadata.
  4. Preprocessor receives `PackageIndex` via context; `ImportTracker` consults `index.HasLocalSymbol(name)` before auto-qualifying an unqualified identifier.
- **Integration Points**
  - `pkg/preprocessor/preprocessor.go`: update `ProcessPackage` (or equivalent) to accept `PackageIndex`.
  - `pkg/preprocessor/import_tracker.go` (new or existing) uses `PackageIndex` to seed exclusion list.
  - `pkg/generator/markers.go` (if they emit imports) references index to avoid re-qualifying local helpers inserted by generators.

_Illustrative flow_
```
PackagePlanner → PackageIndex.Build → {FileSymbols}
                           ↓
                     Preprocessor Context
                           ↓
                   ImportTracker decisions
```

## 2. Caching Strategy
| Level | Mechanism | Stored Data | Invalidation |
|---|---|---|---|
| In-memory (per build run) | `PackageIndex` keyed by package path | `map[string]FileSymbols`, hash of file contents | Discard when build exits or package path changes |
| On-disk (watch / daemon) | `pkg/cache/indexcache` (new) storing JSON (package path → file hash → symbol-summary) | Function names, last-modified timestamps, digest of relevant regions | Invalidate entry when file hash differs, file removed, or config toggles import inference |
| IDE / watch mode | `sync.Map` referencing latest `PackageIndex` along with `generation` counter | Enables zero-copy reuse across concurrent file builds | Replace only affected `FileSymbols`; bump generation for dependents |
- Use `xxhash` (already dependency) or `fnv` for 64-bit digest of file content to avoid reading entire file for diff detection after first scan.
- Cache file also records Dingo compiler version; bump schema on breaking changes.

## 3. Incremental Build Handling
- **Changed Files**: When watcher reports file change, recompute `FileSymbols` only for that file, update `PackageIndex`, and mark `dirtySymbols` set. Preprocessor receives updated index for subsequent builds.
- **File Deletions / Renames**: Remove associated entries; re-run import resolution for dependents to ensure stale exclusions removed.
- **Concurrency**: Use RCU-style pattern—reader goroutines (per-file transpilation) read immutable snapshot (`PackageIndex` with `version`). Mutations create new snapshot while prior builds finish.
- **Partial Packages**: When only subset of files requested (e.g., `dingo build file.dingo`), still provide full package index from cache; if missing, opportunistically load previous on-disk cache to avoid scanning rest unless needed.
- **Watch Mode Budget**: Pipeline ensures `BuildIndex` ≤ 200 ms for 50 files by:
  - Streaming scanner (single pass, no allocations beyond slices)
  - Parallel file scanning with worker pool respecting CPU cores but capped (min(4, runtime.GOMAXPROCS))
  - Early bailout if no unqualified imports found previously (flag in cache) → optionally skip rebuild.

## 4. Performance Analysis
- **Complexity**: O(total lines) per package with small constant factor; scanning uses `bufio.Scanner`/`byte` search to identify tokens.
- **Estimated Timing**:
  - 10 files × 200 LOC: ~40 µs/line → ~80 ms total scan.
  - 50 files × 300 LOC: ~400 µs/line aggregated with 4 workers → ~350 ms (within <500 ms budget).
- **Hot Paths**: Regex for declaration detection, string allocations when capturing identifiers. Use reusable byte slices + `[]byte` → string conversion only once per symbol.
- **Memory**: `PackageIndex` roughly (#symbols * 64 bytes). With 500 symbols ≈ 32 KB.
- **Instrumentation**: Add `trace.Logger` hooks to emit `package_index_scan_duration` metrics (tied into existing telemetry).

## 5. Implementation Plan
1. **Scaffolding** (`pkg/preprocessor/importtracker/index.go`): define `PackageIndex`, `FileSymbols`, builder API, thread-safe snapshot semantics.
2. **Scanner** (`pkg/preprocessor/importtracker/scanner.go`): streaming tokenizer for `.dingo` files; unit tests covering nested blocks, pattern matching, multi-line signatures.
3. **Planner Integration** (`cmd/dingo/build/package.go` + `pkg/preprocessor/preprocessor.go`): compute index before invoking preprocessors; extend pipeline context struct.
4. **ImportTracker Update** (`pkg/preprocessor/preprocessor.go` or `pkg/preprocessor/imports.go`): consult `PackageIndex` when placing auto-qualified imports and when generating fallback alias set.
5. **Caching Layer** (`pkg/cache/packageindex`): read/write JSON snapshots, include compiler version + feature flag; integrate with watch daemon.
6. **Incremental Hooks** (`cmd/dingo/watch/watch.go`): on change event, rescan file, update cache, broadcast new snapshot to worker goroutines.
7. **Performance Harness** (`tests/perf/import_inference_test.go`): benchmark scanning vs package size; ensure <500 ms.
8. **Documentation** (`features/import_inference.md`): describe behavior + edge cases.

## 6. Trade-offs & Edge Cases
- **False Positives**: Macros or generated code might look like functions but aren't; mitigate via simple grammar (require `fn`/`func` keyword start + identifier + `(`). Provide escape hatch comment (`// dingo:no-import-scan` block).
- **Generated Files**: Skip files under `tests/golden/` or `*.go.golden`; rely on file extension filter + user config.
- **Nested Modules / Replace Directories**: Index keyed by module root path; ensure we don't cross module boundaries (stop scanning when encountering nested go.mod unless `--recursive`).
- **Tooling Integration**: Ensure `dingo-lsp` receives package index via shared daemon memory; fall back to synchronous rebuild when LSP opened file belongs to cold package.
- **Large Files**: Scanner handles >1MB files by chunking; ensure `bufio.Scanner` split size increased.
- **Consistency**: Need deterministic ordering; keep symbol map insertion order stable (use `[]Symbol` + map for membership) for reproducible builds.
