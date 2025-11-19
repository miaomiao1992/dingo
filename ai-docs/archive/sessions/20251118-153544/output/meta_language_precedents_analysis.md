## Meta-Language Precedents Analysis

This analysis explores how various successful meta-languages approach parsing to inform Dingo's parser architecture.

### TypeScript

TypeScript builds on JavaScript's parsing by extending its grammar. The TypeScript compiler (`tsc`) has its own parser that can parse ECMAScript (JavaScript) syntax, plus the additional TypeScript syntax like type annotations, interfaces, and enums. It generates an Abstract Syntax Tree (AST) that is a superset of the JavaScript AST. This allows TypeScript to leverage the existing JavaScript ecosystem and toolchain while providing powerful type-checking capabilities. The key is that `tsc` fully understands both JavaScript and TypeScript syntax.

### Kotlin

Kotlin, being a language for the JVM, Android, and web, has a sophisticated parser that generates its own AST. For JVM targets, this AST is then compiled to JVM bytecode. For Kotlin/JS and Kotlin/Native, it uses different backends to generate JavaScript or native binaries. Kotlin's parsing strategy involves defining its own grammar that is distinct from Java but interoperable. It does not extend Java's parser directly but provides seamless integration at the language and tooling level.

### Scala.js

Scala.js compiles Scala code to JavaScript. Its parsing process starts with the standard Scala compiler, which produces a Scala AST. This Scala AST then undergoes a specific transformation phase within Scala.js that targets JavaScript semantics and constructs. Similar to Kotlin, Scala.js doesn't extend a JavaScript parser; it has its own parser (Scala's parser) and then transpiles its AST to JavaScript.

### Borgo

Borgo is a direct precedent for Dingo, aiming to provide Rust-like features for Go. Its approach involves a preprocessor that transforms Rust-like syntax into valid Go code. This is very similar to Dingo's current Stage 1 preprocessor. After preprocessing, it uses the standard `go/parser` to parse the now-valid Go code. This validates Dingo's two-stage approach, where a text-based preprocessor handles the "meta-language" syntax that isn't valid Go, and then the native Go parser takes over.

### CoffeeScript

CoffeeScript, which transpiled to JavaScript, used its own parser written in JavaScript (often using Jison or similar parser generators). It defined its own syntax and grammar entirely separate from JavaScript, then generated equivalent JavaScript. While successful in its time, managing a completely separate parser can be complex and requires maintaining compatibility with the target language's evolving grammar.

### Elm

Elm is a functional language that compiles to JavaScript. It has its own dedicated parser and compiler. The Elm compiler parses Elm code, builds an AST, and then generates highly optimized JavaScript. Like CoffeeScript, it fully owns its language definition and parsing, which provides maximum flexibility but also requires significant effort to build and maintain the parsing infrastructure.

## Key Takeaways and Lessons Learned

1.  **Separate Parser or Superset Parser**: Meta-languages typically either have:
    *   **Their own dedicated parser**: (e.g., Kotlin, Scala.js, CoffeeScript, Elm) that is entirely independent of the target language's parser. This offers maximum control over syntax but requires substantial development and maintenance effort.
    *   **A superset parser**: (e.g., TypeScript) that parses the base language's grammar and extends it with new syntax. This allows leveraging existing language tooling but means the meta-language's syntax must be a consistent extension of the base language.

2.  **Transpilation vs. Direct Parsing**: Languages like Borgo and Dingo use a transpilation approach where the "meta" syntax is converted into the base language's syntax _before_ the base language's official parser is invoked. This avoids the need to modify or extend the base language's (Go's) parser directly, which is often difficult or impossible due to lack of extension points.

3.  **AST Transformation**: Regardless of the initial parsing approach, AST transformation is a common and powerful technique for implementing language features. Once an AST is obtained (either from a custom parser or the base language's parser after preprocessing), features like type-checking, code generation, and semantic analysis are performed by traversing and modifying this AST.

4.  **Balance of Complexity**: Building a custom, high-performance parser is a significant undertaking. Leveraging existing parsers (e.g., `go/parser`) through preprocessing (as Dingo does) is a pragmatic approach that reduces complexity while still allowing for custom syntax.

5.  **Borgo as a strong validation**: Borgo's successful use of a preprocessor followed by `go/parser` for Rust-like features strongly validates Dingo's current two-stage architecture for Go. This approach minimizes the parsing burden by offloading it to the battle-tested `go/parser` once the custom syntax elements are translated.

In summary, directly extending Go's native parser is not typical for meta-languages. Most either build their own parser or, like Borgo and Dingo, use a preprocessor to convert custom syntax into valid Go/JS before passing it to the native parser. This keeps the parsing robust and leverages existing tooling effectively.
