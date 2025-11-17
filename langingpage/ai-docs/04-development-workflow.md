# Astro Development Workflow

This guide covers the complete development workflow for Astro projects, from starting the dev server to building for production.

## Overview

Astro provides a streamlined workflow with three primary modes:

1. **Development** - Local dev server with hot reloading
2. **Build** - Production-ready build process
3. **Preview** - Test production build locally

## Development Server

Astro includes a built-in development server for rapid development with live updates.

### Starting the Dev Server

**Command:** `astro dev`

Run via your package manager:

```bash
# npm
npm run dev

# pnpm
pnpm dev

# yarn
yarn dev
```

**Default URL:** `http://localhost:4321/`

### Development Features

The dev server provides:

1. **Live File Watching**
   - Changes to files in `src/` trigger automatic updates
   - Browser updates without manual refresh
   - No server restart required

2. **Dev Toolbar**
   - Islands architecture inspection
   - Accessibility auditing
   - Performance monitoring
   - Other development utilities

3. **Error Reporting**
   - Clear error messages in terminal
   - Browser overlay for errors
   - Stack traces for debugging

4. **Hot Module Replacement (HMR)**
   - Fast updates without full page reload
   - Preserves component state when possible

### Dev Server Configuration

Configure in `astro.config.mjs`:

```javascript
// astro.config.mjs
export default defineConfig({
  server: {
    port: 3000,           // Custom port (default: 4321)
    host: true,           // Listen on all addresses
    open: true,           // Auto-open browser
  },
});
```

**CLI Options:**

```bash
# Custom port
astro dev --port 3000

# Custom host
astro dev --host 0.0.0.0

# Open browser automatically
astro dev --open
```

### Common Dev Server Scenarios

#### Port Already in Use

If port 4321 is busy, Astro automatically tries the next available port:

```
Port 4321 is in use. Trying 4322...
Server running at http://localhost:4322/
```

#### Network Access

To access dev server from other devices on your network:

```bash
astro dev --host
```

Then access via your local IP:
```
http://192.168.1.100:4321/
```

#### Slow Updates

If HMR feels slow:
1. Check for large imports
2. Reduce component complexity
3. Optimize asset sizes
4. Check for expensive computations in component code

## Build Process

The build process creates a production-ready version of your site.

### Building for Production

**Command:** `astro build`

Run via your package manager:

```bash
# npm
npm run build

# pnpm
pnpm build

# yarn
yarn build
```

### Build Output

**Default output directory:** `dist/`

The build creates:

```
dist/
├── index.html           # Static HTML
├── about.html
├── _astro/              # Optimized assets
│   ├── styles.123abc.css
│   └── script.456def.js
└── assets/              # Processed images
    └── hero.webp
```

**What happens during build:**

1. **Compilation**
   - Astro components → HTML
   - TypeScript → JavaScript
   - CSS → Optimized CSS

2. **Optimization**
   - Minification of HTML, CSS, JS
   - Image optimization
   - Asset bundling and hashing

3. **Code Splitting**
   - Automatic code splitting
   - Per-page bundles
   - Shared chunk optimization

4. **Type Checking** (if configured)
   - Validates TypeScript
   - Reports type errors

### Build Configuration

Configure in `astro.config.mjs`:

```javascript
// astro.config.mjs
export default defineConfig({
  build: {
    format: 'directory',     // or 'file'
    assets: '_astro',        // Asset directory name
    assetsPrefix: '/cdn',    // CDN prefix
  },
  outDir: './dist',          // Output directory
});
```

#### Output Formats

**Directory mode (default):**
```
dist/
├── index.html           # /
├── about/
│   └── index.html       # /about/
└── blog/
    ├── index.html       # /blog/
    └── post-1/
        └── index.html   # /blog/post-1/
```

**File mode:**
```
dist/
├── index.html           # /
├── about.html           # /about.html
└── blog/
    ├── index.html       # /blog/
    └── post-1.html      # /blog/post-1.html
```

### Build Performance

**Optimization strategies:**

1. **Static Site Generation (SSG)**
   - Pre-renders all pages at build time
   - Fastest possible page loads
   - Default mode

2. **Incremental Static Regeneration (ISR)**
   - Update specific pages without full rebuild
   - Requires adapter support

3. **Parallel Processing**
   - Builds pages in parallel
   - Faster builds for large sites

**Monitoring build:**

```bash
# Verbose output
astro build --verbose

# Debug mode
astro build --debug
```

### Type Checking During Build

Configure TypeScript strictness:

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/strict",  // or "strictest"
  "compilerOptions": {
    "strictNullChecks": true
  }
}
```

Build will fail if type errors are found.

**Skip type checking:**

```bash
astro build --no-check
```

## Preview Mode

Preview lets you test the production build locally before deployment.

### Running Preview

**Command:** `astro preview`

Run via your package manager:

```bash
# npm
npm run preview

# pnpm
pnpm preview

# yarn
yarn preview
```

**Default URL:** `http://localhost:4321/`

### Preview Configuration

```javascript
// astro.config.mjs
export default defineConfig({
  preview: {
    port: 8080,           // Custom port
    host: true,           // Listen on all addresses
  },
});
```

### Preview vs Development

| Feature | Development | Preview |
|---------|-------------|---------|
| **HMR** | Yes | No |
| **File Watching** | Yes | No |
| **Optimization** | Minimal | Full |
| **Source Maps** | Yes | Optional |
| **Build Output** | None | Uses `dist/` |
| **Speed** | Fast updates | Static serving |

**Important:** Preview serves the **last build**. Changes require rebuilding:

```bash
# 1. Make changes to source files
# 2. Rebuild
pnpm build
# 3. Preview again
pnpm preview
```

### Exiting Preview

Press `Ctrl + C` in terminal to stop preview server.

## CLI Commands Reference

### Core Commands

```bash
# Start dev server
astro dev

# Build for production
astro build

# Preview production build
astro preview

# Check for issues
astro check
```

### Advanced Commands

```bash
# Add integration
astro add [integration]

# Sync content collections
astro sync

# Get CLI help
astro --help

# Get command-specific help
astro build --help
```

### Integration Management

```bash
# Add React
astro add react

# Add Tailwind CSS
astro add tailwind

# Add multiple integrations
astro add react tailwind

# Add SSR adapter
astro add vercel
```

**What `astro add` does:**

1. Installs the integration package
2. Updates `astro.config.mjs`
3. Installs peer dependencies
4. Provides setup instructions

### Content Collection Sync

```bash
# Generate TypeScript types for content collections
astro sync
```

**When to run:**
- After defining new content collections
- After changing content schemas
- After git pull (if schemas changed)

**What it does:**
- Generates `.astro/` directory with types
- Creates type definitions for collections
- Validates content against schemas

## Complete Workflow Example

### Typical Development Session

```bash
# 1. Start development
pnpm dev

# 2. Make changes to files in src/
# 3. Browser auto-updates

# 4. Add new integration
pnpm astro add react

# 5. Continue development
# Changes hot-reload automatically

# 6. When ready to deploy
pnpm build

# 7. Test production build
pnpm preview

# 8. If issues found
# Make fixes, then rebuild
pnpm build
pnpm preview

# 9. Deploy
# (Deployment varies by platform)
```

### Working with Content Collections

```bash
# 1. Define collection schema
# Edit src/content/config.ts

# 2. Sync types
pnpm astro sync

# 3. Add content files
# Create .md files in src/content/[collection]/

# 4. Use in pages
# Query collections in .astro files

# 5. Build and deploy
pnpm build
```

## Environment Variables

### Development vs Production

```javascript
// Different values per environment
if (import.meta.env.DEV) {
  // Development-only code
  console.log('Dev mode');
}

if (import.meta.env.PROD) {
  // Production-only code
  enableAnalytics();
}
```

### Environment Files

```
.env                # All environments
.env.development    # Dev only
.env.production     # Production only
```

**Example `.env`:**

```env
PUBLIC_API_URL=https://api.example.com
SECRET_KEY=abc123  # Not exposed to client
```

**Usage:**

```astro
---
// Server-side only
const secretKey = import.meta.env.SECRET_KEY;

// Client-side accessible (PUBLIC_ prefix)
const apiUrl = import.meta.env.PUBLIC_API_URL;
---
```

## Debugging

### Development Debugging

**Browser DevTools:**
- Inspect rendered HTML
- Check network requests
- View console logs

**Astro DevTools:**
- Islands inspection
- Accessibility audits
- Performance metrics

**VS Code Debugging:**

```json
// .vscode/launch.json
{
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "Astro Dev",
      "runtimeExecutable": "npm",
      "runtimeArgs": ["run", "dev"],
      "console": "integratedTerminal"
    }
  ]
}
```

### Build Debugging

**Verbose output:**

```bash
astro build --verbose
```

**Debug mode:**

```bash
DEBUG=* astro build
```

**Common build issues:**

1. **Type errors**
   - Check `tsconfig.json`
   - Run `astro check`

2. **Missing dependencies**
   - Run `pnpm install`
   - Check `package.json`

3. **Asset 404s**
   - Verify paths (absolute vs relative)
   - Check `public/` vs `src/assets/`

4. **Slow builds**
   - Reduce image sizes
   - Check for circular dependencies
   - Optimize third-party scripts

## Performance Monitoring

### Build Performance

**Track build time:**

```bash
time pnpm build
```

**Analyze bundle size:**

```bash
# Install bundle analyzer
pnpm add -D rollup-plugin-visualizer

# Configure in astro.config.mjs
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig({
  vite: {
    plugins: [visualizer()],
  },
});

# Build and view stats
pnpm build
```

### Dev Server Performance

**Monitor dev server:**
- Watch for slow HMR updates
- Check terminal for warnings
- Use browser performance tools

## Best Practices for AI Agents

### 1. Always Use Package Manager Scripts

```bash
# Good: Use configured scripts
pnpm dev
pnpm build
pnpm preview

# Avoid: Direct CLI (may not respect project config)
astro dev
```

### 2. Check Before Building

```bash
# Validate TypeScript and content
pnpm astro check

# Then build
pnpm build
```

### 3. Preview Before Deploying

```bash
# Always test production build locally
pnpm build
pnpm preview

# Verify:
# - All pages load
# - Assets load correctly
# - Forms work
# - Links function
```

### 4. Use Environment Variables Properly

```astro
---
// Good: PUBLIC_ prefix for client-side
const apiUrl = import.meta.env.PUBLIC_API_URL;

// Bad: Exposes secret to client
const secret = import.meta.env.SECRET_KEY;  // Only use server-side!
---
```

### 5. Monitor Build Output

Watch for:
- Warnings about large bundles
- Type errors
- Missing dependencies
- Asset optimization opportunities

## Common Commands Summary

```bash
# Development
pnpm dev                    # Start dev server
pnpm dev --port 3000       # Custom port
pnpm dev --host            # Network access

# Building
pnpm build                 # Production build
pnpm build --verbose       # Verbose output
pnpm astro check           # Type check

# Preview
pnpm preview               # Preview build
pnpm preview --port 8080   # Custom port

# Integrations
pnpm astro add [name]      # Add integration
pnpm astro sync            # Sync content types

# Utilities
pnpm astro --help          # CLI help
pnpm astro check           # Check project
```

## Troubleshooting

### Dev Server Won't Start

**Check:**
1. Port availability
2. Node version (use latest LTS)
3. Dependencies installed (`pnpm install`)
4. Config file syntax

### Build Fails

**Check:**
1. TypeScript errors (`pnpm astro check`)
2. Missing dependencies
3. Invalid imports
4. Content collection schemas

### Preview Shows Old Content

**Solution:**
```bash
# Rebuild first
pnpm build

# Then preview
pnpm preview
```

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro development workflow
**Next Module**: Component Development (05-components.md)
