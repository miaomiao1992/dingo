# Similar Tools Research: Output Organization Strategies

## Executive Summary

**Key Finding**: Meta-languages and transpilers use THREE primary organization strategies:

1. **In-Place Transformation** (templ, Scss) - Generate alongside sources
2. **Build Directory Separation** (TypeScript, Rust, Go) - Separate output tree
3. **Hybrid Approach** (Babel, CoffeeScript) - Configurable per project

**Recommendation Preview**: Dingo should adopt **Build Directory Separation** (Strategy 2) as primary, with **In-Place** as fallback for compatibility.

---

## 1. TypeScript: The Gold Standard

### File Organization

**Source Layout**:
```
myproject/
├── src/
│   ├── main.ts
│   ├── utils.ts
│   └── types.ts
├── tsconfig.json
└── package.json
```

**Output Layout** (default: `outDir: "dist"`):
```
myproject/
├── src/           (unchanged, only .ts files)
├── dist/          (generated)
│   ├── main.js
│   ├── main.js.map
│   ├── utils.js
│   ├── utils.js.map
│   └── types.js
└── tsconfig.json
```

**Configuration** (`tsconfig.json`):
```json
{
  "compilerOptions": {
    "outDir": "./dist",           // Output directory
    "sourceMap": true,            // Generate .js.map files
    "inlineSourceMap": false,     // Or embed in .js
    "sourceRoot": "../src",       // For debugger source lookup
    "mapRoot": "./maps",          // Optional: separate map dir
    "rootDir": "./src"            // Input root
  }
}
```

### Key Features

1. **Complete Separation**: Source and output never mixed
2. **Mirror Structure**: `dist/` mirrors `src/` folder hierarchy
3. **Flexible Source Maps**: Inline, separate, or custom directory
4. **Clean .gitignore**: Just `dist/` (one line)
5. **IDE Integration**: VS Code knows to hide `dist/` by default

### Build Command
```bash
tsc              # Uses tsconfig.json
tsc --outDir build  # Override output
```

### LSP Integration

**VS Code TypeScript Language Server**:
- Reads `tsconfig.json` to understand project structure
- Uses `sourceMap` to map `.js` positions back to `.ts`
- Debugger reads source maps to show original `.ts` during debugging
- **Key**: Source maps contain relative paths between output and source

**Source Map Example**:
```json
{
  "version": 3,
  "file": "main.js",
  "sourceRoot": "../src",
  "sources": ["main.ts"],
  "mappings": "AAAA,..."
}
```

### Relevance to Dingo

✅ **Adopt**:
- Separate `out/` or `build/` directory
- Configuration file (already have `dingo.toml`)
- Mirror source structure in output

⚠️ **Consider**:
- TypeScript allows `.ts` and `.js` in same directory (less common)
- Dingo might want to support this for Go interop

---

## 2. Rust: Strict Build Separation

### File Organization

**Source Layout**:
```
myproject/
├── src/
│   ├── main.rs
│   ├── lib.rs
│   └── utils.rs
├── Cargo.toml
└── Cargo.lock
```

**Output Layout** (`cargo build`):
```
myproject/
├── src/           (unchanged)
├── target/        (generated, gitignored)
│   ├── debug/
│   │   ├── myproject      (binary)
│   │   ├── myproject.d    (dependency info)
│   │   └── deps/          (compiled dependencies)
│   └── release/
│       └── myproject      (optimized binary)
└── Cargo.toml
```

### Key Features

1. **Zero Clutter**: `src/` NEVER contains build artifacts
2. **Profile Separation**: `debug/` vs `release/` builds
3. **Intermediate Files Hidden**: All `.o`, `.rlib` in `target/deps/`
4. **Single .gitignore**: Just `/target`
5. **No Source Maps**: Rust keeps debug symbols in binary (DWARF format)

### Build Commands
```bash
cargo build            # → target/debug/
cargo build --release  # → target/release/
cargo clean            # Delete entire target/
```

### Relevance to Dingo

✅ **Adopt**:
- Clear separation (e.g., `build/` directory)
- Easy cleanup (`rm -rf build/`)
- Never modify source tree

⚠️ **Different**:
- Rust compiles to binaries, Dingo transpiles to source
- Dingo `.go` files must be readable/editable (Go interop)

---

## 3. templ: In-Place Generation (Go-Specific)

### File Organization

**templ** generates Go code from HTML-like templates:

**Source Layout**:
```
myproject/
├── home.templ       (template source)
├── user.templ
└── layout.templ
```

**Output Layout** (`templ generate`):
```
myproject/
├── home.templ
├── home_templ.go       (generated)
├── user.templ
├── user_templ.go       (generated)
├── layout.templ
└── layout_templ.go     (generated)
```

### Key Features

1. **In-Place Generation**: Output alongside source
2. **Distinct Naming**: `_templ.go` suffix (never collides)
3. **Go Package Integration**: Generated files in same package
4. **LSP Wrapping**: templ LSP wraps gopls, points at `*_templ.go`
5. **Simple .gitignore**: Just `*_templ.go`

### Why In-Place?

**Go's Package Model**:
- All `.go` files in a directory = single package
- Can't split package across directories
- Generated code must be in same directory to share private types

**templ's Choice**: Trade clutter for seamless Go integration.

### Relevance to Dingo

✅ **Similar Use Case**: Both transpile to Go, need gopls integration

⚠️ **Key Difference**:
- templ generates additional files (templates → code)
- Dingo replaces files (`.dingo` → `.go`)
- Dingo doesn't need same-package requirement (different extension)

**Lesson**: In-place works for Go-specific tools, but clutter is real.

---

## 4. Borgo: Rust-Like Syntax → Go

### File Organization

**Borgo** (github.com/borgo-lang/borgo) transpiles `.brg` to `.go`:

**Source Layout**:
```
myproject/
└── src/
    ├── main.brg
    └── utils.brg
```

**Output Layout** (`borgo build`):
```
myproject/
├── src/           (unchanged)
└── go/            (generated)
    ├── main.go
    └── utils.go
```

**OR** (alternate mode):
```
myproject/
└── src/
    ├── main.brg
    ├── main.go    (generated alongside)
    └── utils.brg
    └── utils.go   (generated alongside)
```

### Key Features

1. **Two Modes**: Separate directory OR in-place
2. **Default Separate**: `go/` directory for outputs
3. **Go Package Aware**: Understands Go module structure
4. **Configurable**: `borgo.toml` controls output location

### Configuration (`borgo.toml`)

```toml
[build]
out_dir = "go"        # Output directory (default: "go")
source_maps = true    # Generate .go.map files
```

### Relevance to Dingo

✅ **Most Similar**:
- Both transpile custom syntax → Go
- Both need LSP (Borgo uses gopls proxy)
- Both target Go ecosystem compatibility

✅ **Adopt**:
- Separate `go/` or `build/` directory as default
- Optional in-place mode for compatibility
- Configuration via `dingo.toml`

---

## 5. Babel: Configurable Everything

### File Organization

**Babel** (JavaScript/TypeScript transpiler):

**Configuration** (`.babelrc` or `babel.config.js`):
```json
{
  "presets": ["@babel/preset-env"],
  "sourceMaps": true,
  "outDir": "dist",
  "ignore": ["node_modules"]
}
```

**OR** use CLI:
```bash
babel src --out-dir dist --source-maps
```

### Output Strategies

1. **Separate Directory**: `src/` → `dist/`
2. **In-Place**: `src/foo.js` → `src/foo.compiled.js`
3. **Single File**: Bundle everything to `bundle.js`

### Key Features

1. **Maximum Flexibility**: Every option configurable
2. **Plugin Ecosystem**: Extensible transformations
3. **Source Map Formats**: Inline, separate, both, none
4. **Watch Mode**: Auto-rebuild on changes

### Relevance to Dingo

⚠️ **Complexity Warning**: Babel's flexibility = configuration burden

✅ **Adopt Selectively**:
- Sane defaults (separate directory)
- Override via config (`dingo.toml`)
- Simple CLI for common cases

---

## 6. Go Itself: Build Artifacts

### File Organization

**Go's Approach**:

**Source**:
```
myproject/
└── main.go
```

**Build**:
```bash
go build           # → myproject (binary, in current dir)
go build -o bin/app  # → bin/app
```

**Build Cache** (hidden):
```
~/.cache/go-build/
├── 00/
├── 01/
└── ...
```

### Key Points

1. **No Intermediate Source Files**: Compiles directly to binary
2. **Hidden Build Cache**: User never sees `.o` files
3. **Output Location Flexible**: `-o` flag or current directory
4. **No Clutter**: Source tree stays pristine

### Relevance to Dingo

⚠️ **Different Model**: Go compiles to binary, Dingo to source

✅ **Learn From**:
- Keep source tree clean
- Hidden intermediate artifacts
- Flexible output with `-o` flag

---

## 7. CoffeeScript: In-Place with Convention

### File Organization

**CoffeeScript** (JavaScript transpiler):

**Default** (in-place):
```
myproject/
├── app.coffee
├── app.js         (generated)
└── utils.coffee
    └── utils.js   (generated)
```

**With `--output`**:
```bash
coffee --compile --output js/ src/
```

**Output**:
```
myproject/
├── src/
│   └── app.coffee
└── js/
    └── app.js
```

### Key Features

1. **Default In-Place**: Simple for small projects
2. **Optional Separation**: Use flag for larger projects
3. **Naming Convention**: `.coffee` vs `.js` never collides
4. **Source Map Support**: `.js.map` files

### Relevance to Dingo

✅ **Lesson**: Start simple (in-place), add complexity as needed

⚠️ **Dated**: CoffeeScript's in-place default is now seen as messy (TypeScript's approach won)

---

## Strategy Comparison Matrix

| Tool       | Default Output | Source Maps | Config File | Go Integration | Verdict for Dingo |
|------------|----------------|-------------|-------------|----------------|-------------------|
| TypeScript | Separate `dist/` | Separate or inline | `tsconfig.json` | N/A | ⭐ Best model |
| Rust       | Separate `target/` | In binary (DWARF) | `Cargo.toml` | N/A | ✅ Clean separation |
| templ      | In-place `*_templ.go` | N/A | N/A | ✅ Native | ✅ Go-specific OK |
| Borgo      | Separate `go/` | Separate `.go.map` | `borgo.toml` | ✅ Native | ⭐⭐ Most similar |
| Babel      | Configurable | Configurable | `.babelrc` | N/A | ⚠️ Too complex |
| Go         | Binary only | In binary | N/A | ✅ Native | ⚠️ Different model |
| CoffeeScript | In-place | Separate | N/A | N/A | ❌ Outdated |

---

## Recommendations for Dingo

### Primary Strategy: Build Directory Separation (Like TypeScript/Borgo)

**Adopt**:
```
myproject/
├── src/
│   ├── main.dingo
│   └── utils.dingo
├── build/              (generated, gitignored)
│   ├── main.go
│   └── utils.go
├── maps/               (or inline in build/)
│   ├── main.go.map
│   └── utils.go.map
└── dingo.toml
```

**Configuration** (`dingo.toml`):
```toml
[build]
out_dir = "build"           # Output directory (default: "build")
source_maps_dir = "maps"    # Or "inline" or "build" (default: "build")

[build.organization]
mode = "separate"           # "separate" or "adjacent" (default: "separate")
```

**CLI**:
```bash
dingo build main.dingo              # → build/main.go
dingo build --out-dir=gen main.dingo  # → gen/main.go
dingo build --adjacent main.dingo   # → main.go (in-place)
```

### Secondary Strategy: In-Place Compatibility (Like templ)

**For Legacy/Simple Projects**:
```bash
dingo build --adjacent main.dingo
```

**Output**:
```
myproject/
├── main.dingo
├── main.go         (generated)
└── main.go.map     (generated)
```

**Use Cases**:
- Quick prototyping
- Single-file scripts
- Backwards compatibility
- Users who prefer in-place

### LSP Integration Updates

**Current** (adjacent files):
```
foo.dingo → foo.go (same directory)
```

**New** (build directory):
```
src/foo.dingo → build/foo.go
```

**LSP Changes Needed**:
1. Read `dingo.toml` to discover `out_dir`
2. Map `src/foo.dingo` → `build/foo.go` using config
3. Load source maps from configured location
4. Update gopls workspace to include `build/` directory

---

## Key Insights

### What Works Well

1. **Separate Directories**: TypeScript, Rust, Borgo all use this (proven pattern)
2. **Configuration Files**: `tsconfig.json`, `Cargo.toml`, `borgo.toml` (user expectations)
3. **Mirror Structure**: Output mirrors source tree (intuitive)
4. **Simple .gitignore**: One directory to ignore (easy)

### What to Avoid

1. **In-Place Default**: CoffeeScript's approach is outdated
2. **Excessive Configuration**: Babel's complexity burden
3. **No Configuration**: Go's build simplicity doesn't apply to transpilers
4. **Inline-Only Source Maps**: Clutters generated code

### Dingo-Specific Considerations

1. **Go Ecosystem**: Generated `.go` must be valid, readable, and in correct package
2. **LSP Wrapping**: gopls needs to resolve imports correctly
3. **Module Structure**: Go's `go.mod` expects certain directory layouts
4. **Interop**: Users might want to mix `.dingo` and hand-written `.go`

---

## Conclusion

**Recommended Path Forward**:

1. **Adopt TypeScript/Borgo model**: Separate build directory by default
2. **Keep templ's in-place as option**: For compatibility and simplicity
3. **Configuration via dingo.toml**: Existing file, natural fit
4. **Update LSP**: Read config, map paths correctly
5. **Migration**: Pre-v1.0 breaking change is acceptable

**Benefits**:
- Clean source trees (like TypeScript)
- Easy .gitignore (like Rust)
- Go ecosystem compatible (like Borgo)
- Flexible for different workflows (like Babel, but simpler)

**Next Steps**:
1. Design `dingo.toml` build configuration schema
2. Plan LSP path mapping logic
3. Define migration strategy for existing projects
