# Installing the Dingo VS Code Extension

## Quick Install (Development)

### Option 1: Symlink (Recommended for Development)

This allows you to edit the extension and see changes immediately:

```bash
# From the dingo repository root
cd editors/vscode

# Create symlink to VS Code extensions folder
ln -s "$(pwd)" "$HOME/.vscode/extensions/dingo-language"

# Reload VS Code
# Command Palette → "Developer: Reload Window"
```

### Option 2: Copy Extension

```bash
# From the dingo repository root
cp -r editors/vscode ~/.vscode/extensions/dingo-language

# Reload VS Code
```

### Option 3: Package and Install

```bash
# From editors/vscode directory
cd editors/vscode

# Install dependencies (if not already installed)
npm install

# Build the grammar
npm run build

# Package extension
npm run package

# Install the .vsix file
code --install-extension dingo-*.vsix
```

## Verify Installation

1. Open VS Code
2. Create a new file with `.dingo` extension
3. Verify syntax highlighting is active
4. Check the language in the bottom-right corner (should show "Dingo")

## Testing

### Test with Example Files

```bash
# Open example files in VS Code
code editors/vscode/examples/result.dingo
code editors/vscode/examples/lambdas.dingo
code editors/vscode/examples/enums.dingo
```

### Use Scope Inspector

1. Open a `.dingo` file
2. Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`)
3. Type: `Developer: Inspect Editor Tokens and Scopes`
4. Click on any token to see its scope and theme rules

## Troubleshooting

### Extension Not Loading

1. Check extension is in the right place:
   ```bash
   ls ~/.vscode/extensions/
   ```
   Should see `dingo-language/` folder

2. Check VS Code output:
   - View → Output
   - Select "Extensions" from dropdown
   - Look for Dingo-related errors

3. Reload VS Code:
   - Command Palette → "Developer: Reload Window"

### No Syntax Highlighting

1. Verify file extension is `.dingo`
2. Check language mode (bottom-right corner)
3. Manually select language: Click language → Type "Dingo"
4. Rebuild grammar:
   ```bash
   cd editors/vscode
   npm run build
   ```

### Grammar Not Updating

If you edit the grammar but don't see changes:

1. Rebuild: `npm run build`
2. Reload VS Code: Command Palette → "Developer: Reload Window"
3. Check console for errors: Help → Toggle Developer Tools

## Uninstalling

### If Symlinked

```bash
rm ~/.vscode/extensions/dingo-language
```

### If Copied

```bash
rm -rf ~/.vscode/extensions/dingo-language
```

### If Installed via .vsix

```bash
code --uninstall-extension dingo-language
```

## Next Steps

- Read `MAINTENANCE.md` for how to maintain and update the extension
- See `README.md` for features and usage examples
- Check `examples/` folder for syntax examples
