# Deploy Astro to GitHub Pages

## Overview

GitHub Pages is a free static site hosting service that deploys directly from your GitHub repository. Using GitHub Actions, you can automatically build and deploy your Astro site whenever you push changes, making it an excellent choice for documentation sites, portfolios, and project pages.

## Why GitHub Pages?

### Pros

- **Free hosting** for public repositories
- **Automatic deployment** via GitHub Actions
- **Custom domain support** with HTTPS
- **Simple setup** with official Astro action
- **Version control integration** built-in
- **No configuration** needed for basic sites

### Cons

- **Static sites only** (no SSR)
- **100 GB/month bandwidth limit**
- **1 GB storage limit**
- **Public repositories only** for free tier
- **Build time limits** (10 builds/hour)

## Deployment Methods

### Official Astro GitHub Action (Recommended)

The recommended way to deploy uses Astro's official GitHub Action with minimal configuration.

### Manual Deployment

For custom workflows or special requirements.

## Quick Start Deployment

### 1. Create GitHub Workflow

Create `.github/workflows/deploy.yml` in your project:

```yaml
name: Deploy to GitHub Pages

on:
  # Trigger on push to main branch
  push:
    branches: [ main ]
  # Allow manual trigger from Actions tab
  workflow_dispatch:

# Required permissions
permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v5

      - name: Install, build, and upload
        uses: withastro/action@v5
        # with:
          # path: . # Root location of Astro project (optional)
          # node-version: 24 # Node version (optional, defaults to 22)
          # package-manager: pnpm@latest # Package manager (optional, auto-detected)

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

### 2. Configure Astro

Update `astro.config.mjs`:

```javascript
import { defineConfig } from 'astro/config';

export default defineConfig({
  site: 'https://<username>.github.io',
  // Add base if deploying to a repository page
  // base: '/my-repo',
});
```

**Replace `<username>` with your GitHub username.**

### 3. Enable GitHub Pages

- Go to repository → Settings → Pages
- Source: "GitHub Actions"
- Save

### 4. Deploy

```bash
git add .
git commit -m "Add GitHub Pages deployment"
git push origin main
```

Your site will automatically build and deploy to:
- User/Organization page: `https://<username>.github.io`
- Repository page: `https://<username>.github.io/<repo>`

## Repository Types

### User/Organization Page

**Repository name:** `<username>.github.io`

**URL:** `https://<username>.github.io`

**Astro Config:**

```javascript
export default defineConfig({
  site: 'https://<username>.github.io',
  // No base needed
});
```

### Project Repository Page

**Repository name:** Any name (e.g., `my-project`)

**URL:** `https://<username>.github.io/my-project`

**Astro Config:**

```javascript
export default defineConfig({
  site: 'https://<username>.github.io',
  base: '/my-project',
});
```

⚠️ **Important:** The `base` value must match your repository name exactly.

## Using `base` Configuration

### What is `base`?

The `base` option tells Astro that your site is deployed at a subpath (e.g., `/my-repo`) instead of the root (`/`).

### Internal Links with `base`

When `base` is set, prefix all internal links:

```astro
---
// astro.config.mjs: base: '/my-repo'
---

<!-- Correct -->
<a href="/my-repo/about">About</a>
<img src="/my-repo/images/logo.png" alt="Logo" />

<!-- Wrong -->
<a href="/about">About</a>
<img src="/images/logo.png" alt="Logo" />
```

**Use `import.meta.env.BASE_URL` for dynamic paths:**

```astro
---
const base = import.meta.env.BASE_URL;
---

<a href={`${base}about`}>About</a>
<img src={`${base}images/logo.png`} alt="Logo" />
```

### Component Navigation

```astro
---
import { base } from '../config';
const links = [
  { href: `${import.meta.env.BASE_URL}`, label: 'Home' },
  { href: `${import.meta.env.BASE_URL}about`, label: 'About' },
  { href: `${import.meta.env.BASE_URL}blog`, label: 'Blog' },
];
---

<nav>
  {links.map(link => (
    <a href={link.href}>{link.label}</a>
  ))}
</nav>
```

## Custom Domain

### Setup

**1. Add CNAME File:**

Create `public/CNAME` with your domain:

```
example.com
```

Or for subdomain:

```
blog.example.com
```

**2. Configure DNS:**

**Apex Domain (`example.com`):**

Add A records pointing to GitHub Pages IPs:

```
185.199.108.153
185.199.109.153
185.199.110.153
185.199.111.153
```

**Subdomain (`blog.example.com`):**

Add CNAME record:

```
<username>.github.io
```

**3. Update Astro Config:**

```javascript
export default defineConfig({
  site: 'https://example.com',
  // Remove base if using custom domain
});
```

**4. Update Internal Links:**

Remove `/my-repo` prefixes from all links since you're now at the root domain.

```astro
<!-- Before (with base) -->
<a href="/my-repo/about">About</a>

<!-- After (custom domain) -->
<a href="/about">About</a>
```

**5. Configure Custom Domain in GitHub:**

- Go to Settings → Pages
- Custom domain: Enter your domain
- Save

⚠️ **Wait for DNS propagation** (can take 24-48 hours).

**6. Enable HTTPS:**

- GitHub will automatically provision SSL certificate
- Check "Enforce HTTPS" when available

## Advanced Configuration

### Custom Package Manager

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  with:
    package-manager: pnpm@latest
```

**Supported:**
- `npm`
- `pnpm@latest`
- `yarn`
- `bun@latest`

### Custom Build Command

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  with:
    build-cmd: pnpm run build:production
```

### Custom Node Version

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  with:
    node-version: 22
```

### Environment Variables

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  env:
    PUBLIC_API_URL: 'https://api.example.com'
    PUBLIC_SITE_NAME: 'My Astro Site'
```

**Using GitHub Secrets:**

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  env:
    PUBLIC_API_URL: ${{ secrets.PUBLIC_API_URL }}
    PRIVATE_API_KEY: ${{ secrets.PRIVATE_API_KEY }}
```

To add secrets:
- Go to Settings → Secrets and variables → Actions
- New repository secret

### Custom Path

For monorepos:

```yaml
- name: Install, build, and upload
  uses: withastro/action@v5
  with:
    path: ./packages/website
```

## Deployment Branches

### Deploy from Different Branch

```yaml
on:
  push:
    branches: [ production ]
```

### Deploy from Multiple Branches

```yaml
on:
  push:
    branches:
      - main
      - develop
```

### Deploy on Tag

```yaml
on:
  push:
    tags:
      - 'v*'
```

## Manual Deployment

### Without GitHub Action

**1. Build Locally:**

```bash
pnpm build
```

**2. Install `gh-pages`:**

```bash
pnpm add -D gh-pages
```

**3. Add Deploy Script:**

```json
{
  "scripts": {
    "deploy": "gh-pages -d dist"
  }
}
```

**4. Deploy:**

```bash
pnpm run deploy
```

This pushes `dist/` to the `gh-pages` branch.

**5. Configure GitHub Pages:**

- Settings → Pages
- Source: Deploy from a branch
- Branch: `gh-pages` / `root`

## Testing Before Deployment

### Preview Locally

```bash
pnpm build
pnpm preview
```

Simulates production environment at `http://localhost:4321`.

### Test with `base`

```bash
# If base is '/my-repo'
pnpm preview --base /my-repo
```

Or add to `package.json`:

```json
{
  "scripts": {
    "preview:base": "astro preview --base /my-repo"
  }
}
```

## Common Patterns

### Deploy on Pull Request

Preview deployments for PRs:

```yaml
name: Preview Deployment

on:
  pull_request:
    branches: [ main ]

jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: withastro/action@v5
      # Deploy to a preview URL
      # (Requires additional setup with Netlify/Vercel)
```

### Deploy with Cache

Speed up builds:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: 'pnpm'

      - name: Setup pnpm
        uses: pnpm/action-setup@v3
        with:
          version: 9

      - name: Get pnpm store directory
        id: pnpm-cache
        run: echo "pnpm_cache_dir=$(pnpm store path)" >> $GITHUB_OUTPUT

      - name: Cache pnpm dependencies
        uses: actions/cache@v4
        with:
          path: ${{ steps.pnpm-cache.outputs.pnpm_cache_dir }}
          key: ${{ runner.os }}-pnpm-${{ hashFiles('**/pnpm-lock.yaml') }}
          restore-keys: |
            ${{ runner.os }}-pnpm-

      - name: Install dependencies
        run: pnpm install

      - name: Build
        run: pnpm build

      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: './dist'
```

### Deploy with Checks

Run tests before deployment:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - run: pnpm install
      - run: pnpm test

  build:
    needs: test # Only build if tests pass
    runs-on: ubuntu-latest
    steps:
      # ... build steps
```

## Troubleshooting

### Issue: 404 on Deployment

**Cause:** Base URL misconfigured.

**Solution:**

For repository page `https://user.github.io/repo`:

```javascript
// astro.config.mjs
export default defineConfig({
  site: 'https://user.github.io',
  base: '/repo',
});
```

### Issue: Assets Not Loading

**Cause:** Missing `base` prefix in asset paths.

**Solution:**

Use `import.meta.env.BASE_URL`:

```astro
<img src={`${import.meta.env.BASE_URL}images/logo.png`} />
```

Or import assets:

```astro
---
import logo from '../assets/logo.png';
---
<img src={logo} alt="Logo" />
```

### Issue: Custom Domain Not Working

**Cause:** DNS not configured correctly.

**Solution:**

1. Check CNAME file in `public/`
2. Verify DNS records with `dig` or `nslookup`:
   ```bash
   dig example.com +short
   ```
3. Wait for DNS propagation (24-48 hours)

### Issue: Build Fails

**Cause:** Missing dependencies or build errors.

**Solution:**

1. Check Actions tab for error logs
2. Test build locally:
   ```bash
   pnpm build
   ```
3. Ensure all dependencies in `package.json`
4. Verify lockfile is committed (`pnpm-lock.yaml`, `package-lock.json`, etc.)

### Issue: Old Version Deployed

**Cause:** Browser cache or GitHub Pages cache.

**Solution:**

1. Hard refresh browser: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)
2. Check deployment timestamp in Actions tab
3. Clear GitHub Pages cache (re-deploy)

## Performance Optimization

### Enable Compression

GitHub Pages automatically serves gzip/brotli compressed files if they exist.

```javascript
// astro.config.mjs
export default defineConfig({
  vite: {
    build: {
      assetsInlineLimit: 0, // Don't inline assets
      minify: 'esbuild',
      cssMinify: true,
    },
  },
});
```

### Optimize Images

Use Astro's Image component:

```astro
---
import { Image } from 'astro:assets';
import hero from '../assets/hero.png';
---

<Image src={hero} alt="Hero" width={800} />
```

See [recipes/images.md](./images.md) for more details.

### Reduce Bundle Size

```bash
# Analyze bundle
pnpm build
du -sh dist/*

# Remove unused dependencies
pnpm prune
```

## Monitoring

### GitHub Pages Status

Check [GitHub Status](https://www.githubstatus.com/) for service issues.

### Deployment History

- Go to Actions tab
- View all deployment runs
- Check logs for each deployment

### Analytics

Add analytics to track visitors:

```astro
---
// src/components/Analytics.astro
---
<script>
  // Google Analytics, Plausible, etc.
</script>
```

## Best Practices

### 1. Commit Lockfile

Always commit your package manager's lockfile:
- `package-lock.json` (npm)
- `pnpm-lock.yaml` (pnpm)
- `yarn.lock` (Yarn)
- `bun.lockb` (Bun)

### 2. Use Environment Variables

```yaml
env:
  PUBLIC_API_URL: ${{ secrets.PUBLIC_API_URL }}
```

### 3. Test Locally First

```bash
pnpm build
pnpm preview
```

### 4. Use Semantic Versioning

Tag releases:

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 5. Add Status Badge

Add to `README.md`:

```markdown
![Deploy Status](https://github.com/<username>/<repo>/actions/workflows/deploy.yml/badge.svg)
```

## Quick Reference

### Basic Workflow

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches: [ main ]

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: withastro/action@v5

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/deploy-pages@v4
        id: deployment
```

### Repository Page Config

```javascript
// astro.config.mjs
export default defineConfig({
  site: 'https://<username>.github.io',
  base: '/my-repo',
});
```

### Custom Domain Config

```javascript
// astro.config.mjs
export default defineConfig({
  site: 'https://example.com',
});
```

```
// public/CNAME
example.com
```

## See Also

- [04-development-workflow.md](../04-development-workflow.md) - Build and preview
- [recipes/deploy-aws.md](./deploy-aws.md) - AWS deployment alternative
- [GitHub Pages Docs](https://docs.github.com/en/pages)
- [Astro Deployment Guide](https://docs.astro.build/en/guides/deploy/github/)
- [withastro/action](https://github.com/withastro/action) - Official GitHub Action

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to deploying Astro sites on GitHub Pages
**Key Concepts**: GitHub Actions, base URL configuration, custom domains, CI/CD, static hosting
