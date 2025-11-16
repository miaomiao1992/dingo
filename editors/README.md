# Editor Support for Dingo

This directory contains editor plugins and extensions for the Dingo language.

## Available Editors

### Visual Studio Code

Full-featured syntax highlighting and language support.

**Location:** `vscode/`

**Quick Start:**
```bash
cd vscode
ln -s "$(pwd)" "$HOME/.vscode/extensions/dingo-language"
```

See `vscode/INSTALL.md` for detailed instructions.

## Planned Support

- **Vim/Neovim** - Tree-sitter grammar and syntax files
- **Emacs** - Major mode with syntax highlighting
- **IntelliJ IDEA** - Plugin for JetBrains IDEs
- **Sublime Text** - TextMate grammar
- **Helix** - Tree-sitter support

## Contributing

Want to add support for your favorite editor? See the VS Code implementation as a reference for:

- TextMate grammar format (`vscode/syntaxes/dingo.tmLanguage.json`)
- Language configuration (`vscode/language-configuration.json`)
- Example files (`vscode/examples/`)

All editors should support:
- Syntax highlighting for Dingo keywords, types, and operators
- Bracket matching and auto-closing pairs
- Comment toggling
- Basic indentation rules

## Language Server (Future)

A Language Server Protocol (LSP) implementation (`dingo-lsp`) is planned that will provide:

- Autocomplete
- Go to definition
- Hover documentation
- Diagnostics
- Refactoring

This will work with any LSP-compatible editor (VS Code, Vim, Emacs, etc.).
