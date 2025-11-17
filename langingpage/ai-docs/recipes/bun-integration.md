# Using Bun with Astro

Bun is an all-in-one JavaScript runtime & toolkit designed for speed. It can replace Node.js as your runtime and package manager for Astro projects.

## Overview

**Benefits of using Bun:**
- Faster package installation
- Faster script execution
- Built-in test runner
- All-in-one toolkit (runtime, bundler, test runner)

**Caution:** Using Bun with Astro may reveal rough edges. Some integrations may not work as expected. Always consult [Bun's official documentation for Astro](https://bun.sh/guides/ecosystem/astro) for current compatibility.

## Prerequisites

Install Bun locally on your machine. See [Bun's installation instructions](https://bun.sh/docs/installation).

**Verify installation:**

```bash
bun --version
```

## Creating a New Astro Project with Bun

Use the `bun create astro` command:

```bash
bun create astro my-astro-project-using-bun
```

**CLI wizard steps:**
1. Choose template (recommended: "Empty" or "Blog" for starters)
2. Install dependencies? → **Yes** (recommended)
3. Initialize git repository? → Your choice
4. TypeScript setup → Recommended: "Strict" or "Strictest"

## Install Dependencies

If you skipped dependency installation during setup:

```bash
bun install
```

**What this does:**
- Installs all dependencies from `package.json`
- Creates `bun.lockb` (Bun's lockfile)
- Much faster than npm/pnpm/yarn

## Add Bun Types

For full TypeScript support with Bun runtime APIs:

```bash
bun add -d @types/bun
```

**This enables:**
- TypeScript autocomplete for Bun APIs
- Type checking for Bun-specific features
- Better IDE support

## CLI Command Flags

### Using Integrations

Add Astro integrations using the `astro add` command:

```bash
# Add React
bun astro add react

# Add Tailwind CSS
bun astro add tailwind

# Add multiple integrations
bun astro add react tailwind sitemap
```

**Common integrations:**
- `react` - React support
- `vue` - Vue support
- `svelte` - Svelte support
- `tailwind` - Tailwind CSS
- `mdx` - MDX support
- `sitemap` - Sitemap generation
- `partytown` - Third-party script optimization

### Using Templates

Start from official examples or GitHub repos:

```bash
# Use official Astro example
bun create astro@latest --template <example-name>

# Use GitHub repository
bun create astro@latest --template <github-username>/<github-repo>
```

**Popular templates:**
```bash
# Blog template
bun create astro@latest --template blog

# Portfolio template
bun create astro@latest --template portfolio

# Docs site (Starlight)
bun create astro@latest --template starlight

# Minimal template
bun create astro@latest --template minimal
```

## Development Workflow

### Start Dev Server

```bash
bun run dev
```

**Default:** http://localhost:4321/

**With options:**

```bash
# Custom port
bun run dev --port 3000

# Open browser automatically
bun run dev --open

# Network access
bun run dev --host
```

### Build for Production

```bash
bun run build
```

**Output:** `dist/` directory with optimized production files

### Preview Production Build

```bash
bun run preview
```

**Use case:** Test production build locally before deployment

### Other Commands

```bash
# Type checking
bun astro check

# Sync content collections
bun astro sync

# CLI help
bun astro --help
```

## Testing with Bun

Bun ships with a built-in, Jest-compatible test runner.

### Create Test Files

**Example test:**

```typescript
// src/utils/math.test.ts
import { expect, test, describe } from 'bun:test';
import { add, subtract } from './math';

describe('Math utilities', () => {
  test('add', () => {
    expect(add(1, 2)).toBe(3);
  });

  test('subtract', () => {
    expect(subtract(5, 3)).toBe(2);
  });
});
```

### Run Tests

```bash
# Run all tests
bun test

# Watch mode
bun test --watch

# Specific file
bun test src/utils/math.test.ts
```

### Testing Astro Components

For component testing, you can use Bun's test runner with other testing libraries:

```bash
# Install testing utilities
bun add -d @testing-library/dom
```

**Example component test:**

```typescript
// src/components/Button.test.ts
import { expect, test } from 'bun:test';
import { render } from './test-utils'; // Custom render helper

test('Button renders with text', () => {
  const { container } = render('<button>Click me</button>');
  expect(container.querySelector('button')).toBeTruthy();
  expect(container.textContent).toBe('Click me');
});
```

## Package.json Scripts

Your `package.json` should have these scripts:

```json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview",
    "astro": "astro",
    "test": "bun test"
  }
}
```

**Running scripts:**

```bash
# Development
bun run dev
# or shorthand
bun dev

# Build
bun run build

# Preview
bun run preview

# Test
bun run test
```

## Performance Comparison

**Package installation:**

| Manager | Time (typical) |
|---------|----------------|
| npm | ~30s |
| pnpm | ~15s |
| yarn | ~20s |
| **bun** | **~2s** |

**Script execution:**

Bun's runtime is typically 2-3x faster than Node.js for most scripts.

## Common Issues and Solutions

### Issue: Integration Not Working

**Solution:**
1. Check [Bun's compatibility tracker](https://github.com/oven-sh/bun/issues)
2. Try with Node.js to isolate issue
3. Report to Bun or integration maintainer

### Issue: Native Modules

Some npm packages with native Node.js bindings may not work.

**Solution:**
- Check if Bun has built-in replacement
- Use alternative packages
- Fall back to Node.js for specific commands

### Issue: Environment Variables

Bun uses `.env` files differently than Node.js.

**Solution:**
Bun automatically loads `.env` files without additional packages.

```bash
# .env
PUBLIC_API_URL=https://api.example.com
```

```typescript
// Access in code
const apiUrl = process.env.PUBLIC_API_URL;
```

## Migrating Existing Project to Bun

### Step 1: Install Bun

```bash
curl -fsSL https://bun.sh/install | bash
```

### Step 2: Install Dependencies

```bash
# Remove existing node_modules and lockfiles
rm -rf node_modules package-lock.json yarn.lock pnpm-lock.yaml

# Install with Bun
bun install
```

### Step 3: Add Bun Types

```bash
bun add -d @types/bun
```

### Step 4: Update Scripts

Replace npm/pnpm/yarn commands with `bun`:

```bash
# Before
npm run dev
npm run build

# After
bun run dev
bun run build
```

### Step 5: Test Thoroughly

```bash
# Test dev server
bun run dev

# Test build
bun run build
bun run preview

# Test all integrations work
```

## Best Practices for AI Agents

### 1. Check Compatibility First

Before using Bun, verify:
- Project dependencies support Bun
- Required integrations work with Bun
- No native Node.js modules that won't work

### 2. Use Bun for Speed

Bun excels at:
- Package installation (very fast)
- Running dev server
- Running builds
- Running tests

### 3. Fall Back When Needed

If an integration doesn't work:
```bash
# Use Node.js for specific command
node --run dev  # or npm run dev
```

### 4. Keep Lockfile

Commit `bun.lockb` to version control for reproducible builds.

### 5. Use Bun's Built-in Features

Leverage Bun's built-ins:
- Test runner (no Jest needed)
- `.env` loading (no dotenv needed)
- Fast transpilation
- Built-in bundler

## Official Resources

- [Bun Documentation](https://bun.sh/docs)
- [Build an app with Astro and Bun](https://bun.sh/guides/ecosystem/astro)
- [Bun GitHub Issues](https://github.com/oven-sh/bun/issues)

## Community Resources

- [Building a Cloudflare Pages site with Bun](https://blog.cloudflare.com/using-bun-with-astro-and-cloudflare-pages/)
- [Using Bun with Astro and Cloudflare Pages](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/#use-bun)

## Quick Reference

```bash
# Create new project
bun create astro my-project

# Install dependencies
bun install

# Add types
bun add -d @types/bun

# Development
bun run dev

# Build
bun run build

# Preview
bun run preview

# Test
bun test

# Add integration
bun astro add [integration]
```

## When to Use Bun

### Use Bun When:

✅ Performance is critical
✅ You want faster package installs
✅ You need built-in test runner
✅ Working on greenfield projects
✅ Dependencies are Bun-compatible

### Consider Alternatives When:

❌ Critical dependencies don't work with Bun
❌ Team is unfamiliar with Bun
❌ Production environment requires Node.js
❌ Using native Node.js modules heavily

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent guide for using Bun with Astro
**Status**: Bun is in active development; compatibility may improve over time
