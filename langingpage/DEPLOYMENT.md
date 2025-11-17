# Dingo Landing Page - Deployment Guide

This guide covers deploying the Dingo landing page to GitHub Pages with custom domain (dingolang.com).

## Table of Contents

- [Overview](#overview)
- [Deployment Architecture](#deployment-architecture)
- [Initial Setup](#initial-setup)
- [DNS Configuration](#dns-configuration)
- [Testing Procedures](#testing-procedures)
- [Future Deployments](#future-deployments)
- [Troubleshooting](#troubleshooting)

## Overview

The Dingo landing page is deployed using:

- **Platform**: GitHub Pages (static site hosting)
- **CI/CD**: GitHub Actions (automated build & deploy)
- **Domain**: dingolang.com (custom domain with HTTPS)
- **Framework**: Astro (static site generation)
- **Package Manager**: pnpm

### Deployment Flow

```
Push to main → GitHub Actions → Build → Deploy → Live at dingolang.com
```

## Deployment Architecture

### GitHub Actions Workflow

Located at `.github/workflows/deploy.yml`:

- **Trigger**: Push to `main` branch or manual dispatch
- **Jobs**:
  1. **Build**: Uses `withastro/action@v5` to install dependencies and build
  2. **Deploy**: Uses `actions/deploy-pages@v4` to publish to GitHub Pages

### Key Configuration

**Monorepo Path**: The workflow specifies `path: ./langingpage` because this project is in a subdirectory of the Dingo monorepo.

**Package Manager**: Set to `pnpm@latest` for automatic detection and use.

**Permissions**: Required for GitHub Pages deployment:
- `contents: read`
- `pages: write`
- `id-token: write`

## Initial Setup

### 1. Verify GitHub Pages Source

⚠️ **Critical**: Ensure GitHub Pages is configured to deploy from GitHub Actions (not branch).

1. Navigate to repository settings:
   - URL: https://github.com/MadAppGang/dingo/settings/pages
   - Or: Settings → Pages (left sidebar)

2. Under "Build and deployment":
   - **Source** dropdown must be set to: **"GitHub Actions"**
   - If showing "Deploy from a branch": Click dropdown and select "GitHub Actions"

3. Verify selection:
   - Should show: "Use a workflow from your repository"
   - No branch selection dropdowns visible

**Why This Matters**: GitHub Pages defaults to "Deploy from a branch" for new repositories. If source is not set to "GitHub Actions", your workflow will run successfully but the site won't deploy.

**Reference**: See `ai-docs/recipes/deploy-github-pages.md` (Quick Start Step 3)

---

### 2. Configure Custom Domain

⚠️ **One-time configuration required**:

1. In the same Pages settings page:
   - **Custom domain**: Enter `dingolang.com`
   - Click "Save"
   - Wait for DNS check (will show "DNS check unsuccessful" until DNS is configured)

2. **Configure HTTPS** (complete AFTER DNS propagation):
   - **Important**: Do NOT enable "Enforce HTTPS" yet
   - GitHub Pages must verify DNS before allowing HTTPS enforcement
   - After DNS propagates (24-48 hours) and certificate provisions (~15 min):
     1. Return to Settings → Pages
     2. Verify "HTTPS" section shows certificate provisioned
     3. Check "Enforce HTTPS" checkbox
   - See "HTTPS Certificate" section below for DNS verification steps

---

### 3. Verify Workflow Files

Ensure these files exist:

- `.github/workflows/deploy.yml` - Deployment workflow
- `langingpage/public/CNAME` - Custom domain configuration
- `langingpage/astro.config.mjs` - Astro configuration with `site: 'https://dingolang.com'`

## DNS Configuration

### Required DNS Records

Configure these A records at your domain registrar (e.g., Cloudflare, Namecheap, etc.):

```
Record Type: A
Host/Name: @ (or blank for apex domain)
Value/Points to: 185.199.108.153
TTL: 3600 (1 hour) or Auto

Record Type: A
Host/Name: @ (or blank for apex domain)
Value/Points to: 185.199.109.153
TTL: 3600 (1 hour) or Auto

Record Type: A
Host/Name: @ (or blank for apex domain)
Value/Points to: 185.199.110.153
TTL: 3600 (1 hour) or Auto

Record Type: A
Host/Name: @ (or blank for apex domain)
Value/Points to: 185.199.111.153
TTL: 3600 (1 hour) or Auto
```

### DNS Provider Examples

#### Cloudflare

1. Log into Cloudflare Dashboard
2. Select your domain
3. Go to DNS → Records
4. Add A records with GitHub Pages IPs (as above)
5. Set Proxy status to "Proxied" or "DNS only"

#### Namecheap

1. Log into Namecheap
2. Domain List → Manage → Advanced DNS
3. Add A records with GitHub Pages IPs (as above)
4. Set TTL to Automatic

### Verify DNS Propagation

After configuring DNS, check propagation:

```bash
# Check DNS resolution
dig dingolang.com +short

# Should return GitHub Pages IPs:
# 185.199.108.153
# 185.199.109.153
# 185.199.110.153
# 185.199.111.153

# Alternative check
nslookup dingolang.com

# Check multiple DNS servers
dig @8.8.8.8 dingolang.com +short  # Google DNS
dig @1.1.1.1 dingolang.com +short  # Cloudflare DNS
```

⏱️ **Timeline**: DNS propagation can take 24-48 hours globally.

### HTTPS Certificate Status

After DNS propagates, GitHub Pages automatically provisions a Let's Encrypt certificate (~15 minutes). Check certificate status:

1. Visit Settings → Pages
2. Look for "HTTPS" section
3. Should show: "✓ Your site is published at https://dingolang.com"
4. If certificate is ready, "Enforce HTTPS" checkbox becomes enabled

**Troubleshooting**:
- If "Enforce HTTPS" is disabled: DNS not fully propagated yet, wait longer
- If certificate fails: Verify DNS A records are correct (see DNS Configuration section)

**Important**: GitHub Pages blocks enabling "Enforce HTTPS" until:
1. DNS fully propagates (24-48 hours)
2. SSL certificate is provisioned (~15 minutes after DNS propagation)
3. Do not attempt to enable HTTPS before these steps complete

## Testing Procedures

### Pre-Deployment Testing (Local)

Before pushing changes to main:

```bash
# Navigate to landing page directory
cd langingpage

# Test build
pnpm build

# Verify output
ls -la dist/

# Preview production build
pnpm preview
# Visit http://localhost:4321

# Check browser console for errors
# Verify all pages load correctly
# Test navigation links
```

### First Deployment Verification

After pushing to main:

1. **Monitor GitHub Actions**:
   - Go to: https://github.com/MadAppGang/dingo/actions
   - Watch "Deploy to GitHub Pages" workflow
   - Verify build job succeeds (green checkmark)
   - Verify deploy job succeeds
   - Note deployment URL from output

2. **Test Temporary URL** (before DNS propagates):
   - Visit: https://madappgang.github.io (if user page)
   - Or: https://madappgang.github.io/dingo (if repository page)
   - Verify site loads correctly
   - Check assets (CSS, JS, images)
   - Check browser console for errors

3. **Test Custom Domain** (after DNS propagates):
   - Visit: https://dingolang.com
   - Verify HTTPS works (green lock icon)
   - Check certificate is valid (click lock icon)
   - Verify assets load correctly
   - Test all navigation links

### Performance Verification

Run a Lighthouse audit:

```bash
# In Chrome DevTools
# 1. Open DevTools (F12)
# 2. Go to Lighthouse tab
# 3. Select "Performance" category
# 4. Click "Analyze page load"
```

**Performance Targets**:
- Performance score: > 90
- LCP (Largest Contentful Paint): < 2.5s
- FID (First Input Delay): < 100ms
- CLS (Cumulative Layout Shift): < 0.1

## Future Deployments

### Automatic Deployment

Every push to `main` triggers automatic deployment:

```bash
# Make changes to landing page
cd langingpage
# Edit files...

# Test locally
pnpm build
pnpm preview

# Commit and push
git add .
git commit -m "Update landing page content"
git push origin main

# GitHub Actions automatically:
# 1. Builds the site
# 2. Deploys to GitHub Pages
# 3. Site updates at https://dingolang.com
```

### Manual Deployment

Trigger deployment without pushing:

**Via GitHub UI**:
1. Go to Actions tab
2. Select "Deploy to GitHub Pages" workflow
3. Click "Run workflow"
4. Select branch (main)
5. Click "Run workflow"

**Via GitHub CLI** (`gh`):
```bash
gh workflow run deploy.yml
```

### Deployment Timeline

- **Build time**: ~1-3 minutes (depends on site size)
- **Deployment time**: ~30 seconds
- **Total**: ~2-4 minutes from push to live

### Cache Invalidation

Browser caching may show old version after deployment:

**Hard Refresh**:
- Mac: Cmd + Shift + R
- Windows/Linux: Ctrl + Shift + R

**Check Deployment Timestamp**:
- Go to Actions tab
- Check latest workflow run timestamp
- Compare with browser's loaded version

## Troubleshooting

### Issue: Build Fails in GitHub Actions

**Symptoms**: Red X on workflow run, build job fails

**Common Causes**:
1. Missing dependencies
2. TypeScript/Astro errors
3. Missing lockfile

**Solution**:

```bash
# Test build locally first
cd langingpage
pnpm build

# Check for errors
pnpm astro check

# Verify lockfile is committed
git status pnpm-lock.yaml

# If lockfile missing, create it
pnpm install
git add pnpm-lock.yaml
git commit -m "Add pnpm lockfile"
git push
```

**Check GitHub Actions logs**:
1. Go to Actions tab
2. Click failed workflow run
3. Click "build" job
4. Expand steps to see error details

### Issue: 404 Error on Deployment

**Symptoms**: Site shows "404 - File not found"

**Cause**: Base URL misconfigured (not applicable for custom domain at root)

**Solution**:

Verify `astro.config.mjs`:

```javascript
export default defineConfig({
  site: 'https://dingolang.com',
  // NO base configuration for custom domain at root
});
```

### Issue: Assets Not Loading

**Symptoms**: Images/CSS/JS return 404 errors

**Cause 1**: Assets not in `dist/` after build

**Solution**:
```bash
# Check build output
pnpm build
ls -la dist/

# Verify assets directory exists
ls -la dist/assets/
```

**Cause 2**: Incorrect asset paths

**Solution**:
```astro
<!-- Prefer importing assets -->
---
import logo from '../assets/logo.png';
---
<img src={logo} alt="Logo" />

<!-- Or use import.meta.env.BASE_URL -->
<img src={`${import.meta.env.BASE_URL}images/logo.png`} alt="Logo" />
```

### Issue: Custom Domain Not Working

**Symptoms**: https://dingolang.com shows error or doesn't load

**Checklist**:

1. **Verify CNAME file exists**:
   ```bash
   cat langingpage/public/CNAME
   # Should output: dingolang.com
   ```

2. **Check DNS records**:
   ```bash
   dig dingolang.com +short
   # Should show GitHub Pages IPs
   ```

3. **Verify GitHub Pages settings**:
   - Go to Settings → Pages
   - Custom domain should show "dingolang.com"
   - Should show "DNS check successful" (after propagation)

4. **Wait for DNS propagation**:
   - Can take 24-48 hours
   - Use https://dnschecker.org to check global propagation

5. **Check CNAME file in build output**:
   ```bash
   pnpm build
   ls langingpage/dist/CNAME
   # File should exist
   cat langingpage/dist/CNAME
   # Should output: dingolang.com
   ```

### Issue: HTTPS Certificate Issues

**Symptoms**: "Not Secure" warning, certificate errors

**Solution**:

1. **Wait for DNS to propagate** (24-48 hours)
2. GitHub auto-provisions Let's Encrypt certificate
3. Check GitHub Pages settings:
   - Should show "HTTPS" section
   - Enable "Enforce HTTPS" when available
4. If certificate fails to provision:
   - Remove custom domain in GitHub settings
   - Wait 5 minutes
   - Re-add custom domain
   - Wait 15 minutes for re-provisioning

### Issue: Old Version Showing After Deployment

**Symptoms**: Changes not visible on live site

**Solution**:

1. **Check deployment succeeded**:
   - Go to Actions tab
   - Verify latest workflow completed successfully
   - Check timestamp

2. **Clear browser cache**:
   - Hard refresh: Cmd+Shift+R (Mac) or Ctrl+Shift+R (Windows)
   - Or clear site data in browser settings

3. **Verify build output**:
   ```bash
   pnpm build
   # Check dist/ contains latest changes
   ```

4. **Check GitHub Pages cache**:
   - Sometimes GitHub Pages caches old version
   - Re-deploy to force update:
     ```bash
     git commit --allow-empty -m "Force redeploy"
     git push
     ```

### Issue: Workflow Permission Errors

**Symptoms**: "Resource not accessible by integration" error

**Solution**:

1. Go to repository Settings → Actions → General
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Enable "Allow GitHub Actions to create and approve pull requests"
5. Save

## Build Optimization

Current optimizations in `astro.config.mjs`:

```javascript
vite: {
  build: {
    assetsInlineLimit: 0, // Don't inline assets (better caching)
    minify: 'esbuild',    // Fast minification
    cssMinify: true,      // Minify CSS
  },
}
```

### Check Build Size

```bash
# Build and check size
pnpm build
du -sh langingpage/dist/*

# Detailed size breakdown
du -h langingpage/dist/ | sort -h
```

### Performance Monitoring

Track Core Web Vitals over time:

1. Use Lighthouse CI for automated tracking
2. Monitor GitHub Actions build times
3. Track bundle size trends
4. Use browser DevTools Performance tab

## Quick Reference

### Common Commands

```bash
# Development
pnpm dev              # Start dev server (http://localhost:4321)
pnpm build            # Build for production
pnpm preview          # Preview production build

# Testing
pnpm astro check      # Type check and validate
dig dingolang.com +short  # Check DNS

# Deployment
git push origin main  # Trigger automatic deployment
gh workflow run deploy.yml  # Manual deployment (requires gh CLI)
```

### Important URLs

- **Live Site**: https://dingolang.com
- **GitHub Actions**: https://github.com/MadAppGang/dingo/actions
- **GitHub Pages Settings**: https://github.com/MadAppGang/dingo/settings/pages
- **DNS Checker**: https://dnschecker.org

### Key Files

- `.github/workflows/deploy.yml` - Deployment workflow
- `langingpage/astro.config.mjs` - Astro configuration
- `langingpage/public/CNAME` - Custom domain configuration
- `langingpage/package.json` - Dependencies and scripts

## Security Best Practices

1. **Never commit secrets**: Use GitHub Secrets for API keys
2. **Keep dependencies updated**: Run `pnpm update` regularly
3. **Enable HTTPS**: Always use "Enforce HTTPS" in GitHub Pages
4. **Review Actions logs**: Check for security warnings
5. **Use branch protection**: Protect `main` branch (optional)

## Next Steps

After successful deployment:

1. ✅ Verify site is live at https://dingolang.com
2. ✅ Run Lighthouse audit and document baseline
3. ✅ Test on multiple browsers (Chrome, Firefox, Safari)
4. ✅ Test on mobile devices
5. ⏳ Set up analytics (Google Analytics, Plausible, etc.)
6. ⏳ Monitor uptime and performance
7. ⏳ Consider adding preview deployments for PRs (future)

## Support

- **Astro Docs**: https://docs.astro.build
- **GitHub Pages Docs**: https://docs.github.com/en/pages
- **Astro GitHub Action**: https://github.com/withastro/action
- **Dingo Project**: https://github.com/MadAppGang/dingo

---

**Last Updated**: 2025-11-17
**Deployment Type**: GitHub Pages with Custom Domain
**Framework**: Astro (Static Site Generation)
**Status**: Production Ready
