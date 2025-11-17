# Astro Best Practices Checklist

This checklist helps AI agents validate implementations against Astro community best practices documented in the ai-docs/ modules.

## Quick Reference Guide

Use this checklist during:
- **Planning Phase**: Ensure architectural decisions align with Astro principles
- **Implementation Phase**: Validate code as you write
- **QA Testing Phase**: Systematic verification of best practices
- **Fix Application Phase**: Ensure fixes follow correct patterns

---

## Core Principles Validation

**Reference**: [ai-docs/01-why-astro.md](./01-why-astro.md)

### Server-First Architecture
- [ ] Components render on server by default
- [ ] No client-side JavaScript unless explicitly needed
- [ ] HTML is delivered fully rendered
- [ ] Client hydration is opt-in, not default

**CRITICAL Violations**:
- Sending unnecessary JavaScript to client
- Using framework components without `client:*` directives when static would work
- Over-hydrating components that don't need interactivity

### Zero JS by Default
- [ ] Static components use `.astro` files
- [ ] Framework components only when interactivity is required
- [ ] No global JavaScript bundles unless necessary
- [ ] `<script>` tags are intentional and minimal

**CRITICAL Violations**:
- Importing framework components without checking if `.astro` would suffice
- Adding client-side libraries when server-side would work

### Content-Focused Design
- [ ] Content loads fast (< 1s LCP target)
- [ ] Images are optimized using `<Image />` component
- [ ] Critical CSS is inlined
- [ ] Fonts are preloaded/optimized

**CRITICAL Violations**:
- Using `<img>` tags instead of `<Image />`
- Blocking render with unoptimized assets
- Not using content collections for structured content

---

## Islands Architecture

**Reference**: [ai-docs/02-islands-architecture.md](./02-islands-architecture.md)

### Island Usage Patterns
- [ ] Interactive components are isolated islands
- [ ] Correct `client:*` directive for each use case:
  - `client:load` - Critical interactive elements (rare)
  - `client:idle` - Important but not critical (common)
  - `client:visible` - Below-fold interactions (preferred)
  - `client:media` - Responsive behavior
  - `client:only` - Framework-specific SSR issues
- [ ] Static siblings remain server-rendered
- [ ] No unnecessary island nesting

**CRITICAL Violations**:
- Using `client:load` when `client:visible` would work
- Making entire page an island when only small parts need interactivity
- Not using islands when interactivity is needed

### State Management
- [ ] Shared state uses framework stores (nano-stores, zustand, etc.)
- [ ] Simple state is component-local
- [ ] Server state passed via props
- [ ] URL state for shareable/bookmarkable data

**MEDIUM Issues**:
- Complex prop drilling (should use stores)
- Re-implementing state that exists in URL

---

## Project Structure

**Reference**: [ai-docs/03-project-structure.md](./03-project-structure.md)

### Directory Organization
- [ ] `src/pages/` - Routes only (minimal logic)
- [ ] `src/components/` - Reusable components
- [ ] `src/layouts/` - Page layouts
- [ ] `src/content/` - Content collections
- [ ] `public/` - Static assets (non-optimized)
- [ ] `src/assets/` - Optimized assets (images, etc.)

**CRITICAL Violations**:
- Files in wrong directories (breaks Astro conventions)
- Routes outside `src/pages/`
- Putting optimizable images in `public/`

### File Naming
- [ ] `.astro` for Astro components
- [ ] Framework extension when needed (`.tsx`, `.vue`, `.svelte`)
- [ ] `layout.astro` or descriptive names for layouts
- [ ] Lowercase with hyphens for URLs: `blog-post.astro`
- [ ] PascalCase for component files: `BlogCard.astro`

**MINOR Issues**:
- Inconsistent naming conventions
- Unclear component names

---

## Component Development

**Reference**: [ai-docs/07-astro-components.md](./07-astro-components.md), [ai-docs/06-framework-components.md](./06-framework-components.md)

### Astro Components
- [ ] Frontmatter script for server-side logic
- [ ] TypeScript for props interface
- [ ] Slots for composition (default + named)
- [ ] Scoped styles by default
- [ ] HTML components for simple wrappers

**CRITICAL Violations**:
- Client-side code in frontmatter (runs at build time!)
- Not typing props with TypeScript
- Using framework components when `.astro` would work

### Framework Components
- [ ] Only used when interactivity is required
- [ ] Appropriate `client:*` directive specified
- [ ] Props are serializable (no functions unless `client:only`)
- [ ] Children passed correctly for each framework

**CRITICAL Violations**:
- No `client:*` directive (component won't hydrate)
- Using wrong directive (performance issue)
- Passing non-serializable props

---

## Content Management

**Reference**: [ai-docs/05-content-collections.md](./05-content-collections.md)

### Content Collections (Astro 5.x)
- [ ] Using Content Layer API with appropriate loader
- [ ] Schemas defined with Zod validation
- [ ] TypeScript types generated automatically
- [ ] Querying with `getCollection()` or `getEntry()`
- [ ] References for relationships between collections

**CRITICAL Violations**:
- Using old `src/content/config.ts` pattern (deprecated in 5.x)
- Not defining schemas (loses type safety)
- Manual data loading instead of collections

### Markdown/MDX
- [ ] Frontmatter YAML for metadata
- [ ] Layouts specified for consistent design
- [ ] `<Content />` component for rendering
- [ ] Remark/Rehype plugins for transformations

**MEDIUM Issues**:
- Not using layouts (duplicated markup)
- Missing frontmatter validation

---

## Layouts

**Reference**: [ai-docs/08-layouts.md](./08-layouts.md)

### Layout Patterns
- [ ] Base layout with HTML structure, meta tags
- [ ] Default slot for page content
- [ ] Named slots for optional sections
- [ ] TypeScript props for configuration
- [ ] Layout nesting for inheritance

**CRITICAL Violations**:
- Missing `<!DOCTYPE html>` and `<html>` tags in base layout
- Duplicated `<head>` content across pages
- Not using layouts (SEO and consistency issues)

### SEO & Meta
- [ ] Meta tags in layout `<head>`
- [ ] Dynamic titles and descriptions via props
- [ ] Open Graph tags for social sharing
- [ ] Canonical URLs
- [ ] Structured data (JSON-LD)

**CRITICAL Violations**:
- Missing or duplicate `<title>` tags
- No meta descriptions

---

## Recipes & Common Patterns

**References**: [ai-docs/recipes/](./recipes/)

### Images
**Reference**: [ai-docs/recipes/images.md](./recipes/images.md)

- [ ] Using `<Image />` for local images
- [ ] Using `<Picture />` for art direction
- [ ] Images in `src/assets/` (not `public/`)
- [ ] Alt text for accessibility
- [ ] Responsive formats (WebP, AVIF)

**CRITICAL Violations**:
- Using `<img>` for local images (not optimized)
- Missing alt text
- Large unoptimized images

### Custom Fonts
**Reference**: [ai-docs/recipes/custom-fonts.md](./recipes/custom-fonts.md)

- [ ] Fonts in `public/fonts/` or via Fontsource
- [ ] `@font-face` with proper paths
- [ ] `font-display: swap` for performance
- [ ] Preloading critical fonts
- [ ] Subset fonts when possible

**MEDIUM Issues**:
- Not using `font-display`
- Loading all font weights/styles unnecessarily

### Syntax Highlighting
**Reference**: [ai-docs/recipes/syntax-highlighting.md](./recipes/syntax-highlighting.md)

- [ ] Using built-in Shiki (default)
- [ ] Theme configuration in `astro.config.mjs`
- [ ] `<Code />` component for inline code
- [ ] Dual themes for light/dark mode

**MINOR Issues**:
- Not configuring theme
- Using Prism when Shiki is better

### Scripts & Events
**Reference**: [ai-docs/recipes/scripts-and-events.md](./recipes/scripts-and-events.md)

- [ ] `<script>` for client-side behavior
- [ ] Understanding `is:inline` vs bundled scripts
- [ ] Event delegation for dynamic content
- [ ] Web Components for encapsulated interactivity
- [ ] Passing server data via `data-*` attributes

**CRITICAL Violations**:
- Using `<script>` in frontmatter (runs at build time)
- Not understanding script processing
- Large inline scripts (should be bundled)

### Tailwind Typography
**Reference**: [ai-docs/recipes/tailwind-typography.md](./recipes/tailwind-typography.md)

- [ ] `@tailwindcss/typography` installed
- [ ] `prose` classes for content
- [ ] Customization for brand colors
- [ ] Dark mode support

**MINOR Issues**:
- Not using typography plugin for long-form content

---

## Development Workflow

**Reference**: [ai-docs/04-development-workflow.md](./04-development-workflow.md)

### Build & Dev
- [ ] Using `pnpm dev` for development
- [ ] Using `pnpm build` before deployment
- [ ] Testing with `pnpm preview`
- [ ] No build errors or warnings
- [ ] Environment variables properly configured

**CRITICAL Violations**:
- Build errors (deployment will fail)
- Using development mode in production

---

## Performance Checklist

### Core Web Vitals Targets
- [ ] LCP (Largest Contentful Paint) < 2.5s
- [ ] FID (First Input Delay) < 100ms
- [ ] CLS (Cumulative Layout Shift) < 0.1

### Optimization Techniques
- [ ] Lazy load below-fold images
- [ ] Use `client:visible` for below-fold interactions
- [ ] Preload critical assets
- [ ] Minimize client-side JavaScript
- [ ] Use content collections for static data

**CRITICAL Violations**:
- LCP > 4s
- Large JavaScript bundles (> 200KB)
- Not using Islands Architecture

---

## Accessibility Checklist

### Basic Requirements
- [ ] Semantic HTML elements
- [ ] Alt text for all images
- [ ] ARIA labels when needed
- [ ] Keyboard navigation works
- [ ] Focus indicators visible
- [ ] Color contrast meets WCAG AA (4.5:1)

**CRITICAL Violations**:
- Missing alt text
- Non-semantic HTML (div soup)
- Inaccessible interactive elements

---

## Security Checklist

### Common Issues
- [ ] No XSS vulnerabilities (sanitize user input)
- [ ] Environment variables not exposed to client
- [ ] HTTPS in production
- [ ] CSP headers configured
- [ ] No sensitive data in client bundles

**CRITICAL Violations**:
- XSS vulnerabilities
- Exposed API keys/secrets
- Not validating user input

---

## QA Testing Severity Guide

Use this guide to classify issues during QA testing:

### CRITICAL
- Violates core Astro principles (server-first, zero JS, content-focused)
- Breaks functionality or deployment
- Security vulnerabilities
- Accessibility violations (WCAG A)
- Performance > 4s LCP

### MEDIUM
- Suboptimal patterns (wrong directory, non-standard naming)
- Missing optimization (should use `<Image />`, wrong `client:*` directive)
- Accessibility issues (WCAG AA)
- Performance 2.5s-4s LCP

### MINOR
- Style/convention inconsistencies
- Opportunities for refinement
- Documentation improvements
- Performance optimizations beyond requirements

---

## AI Agent Workflow

### During Planning
1. Read [ai-docs/INDEX.md](./INDEX.md) first
2. Consult relevant modules for the task
3. Reference this checklist for decisions
4. Document which ai-docs modules influenced the plan

### During Implementation
1. Keep this checklist open
2. Consult ai-docs on-demand for specific patterns
3. Validate as you write code
4. Document ai-docs references in code comments

### During QA Testing
1. Systematically check each section
2. Classify issues by severity (CRITICAL/MEDIUM/MINOR)
3. Reference specific ai-docs modules for each issue
4. Include ai-docs links in QA reports

### During Fixes
1. Consult ai-docs module referenced in issue
2. Apply the correct pattern from ai-docs
3. Re-validate against checklist
4. Document which ai-docs module was used

---

## Common Anti-Patterns

### ❌ Using React/Vue when .astro would work
**Why**: Sends unnecessary JavaScript to client
**Fix**: Use `.astro` components for static content
**Reference**: [ai-docs/07-astro-components.md](./07-astro-components.md)

### ❌ Putting images in public/ instead of src/assets/
**Why**: Images don't get optimized
**Fix**: Move to `src/assets/`, use `<Image />`
**Reference**: [ai-docs/recipes/images.md](./recipes/images.md)

### ❌ Using client:load everywhere
**Why**: Loads all JavaScript immediately (bad performance)
**Fix**: Use `client:visible` or `client:idle`
**Reference**: [ai-docs/02-islands-architecture.md](./02-islands-architecture.md)

### ❌ Not using Content Collections
**Why**: Loses type safety and performance benefits
**Fix**: Set up Content Layer with appropriate loader
**Reference**: [ai-docs/05-content-collections.md](./05-content-collections.md)

### ❌ Duplicating layout markup
**Why**: Maintenance burden, SEO issues
**Fix**: Create layouts with slots
**Reference**: [ai-docs/08-layouts.md](./08-layouts.md)

### ❌ Client-side code in frontmatter
**Why**: Frontmatter runs at BUILD time, not runtime
**Fix**: Use `<script>` tags for client code
**Reference**: [ai-docs/recipes/scripts-and-events.md](./recipes/scripts-and-events.md)

---

## Quick Decision Trees

### Should I use a framework component?
```
Is interactivity needed?
├─ NO → Use .astro component
└─ YES → Is it simple (click, toggle)?
         ├─ YES → Try <script> or Web Component first
         └─ NO (complex state) → Use framework component
```

### Which client: directive?
```
When does it need to be interactive?
├─ Immediately (above fold, critical) → client:load
├─ After page loads (important but not critical) → client:idle
├─ When scrolled into view (below fold) → client:visible ⭐ PREFERRED
├─ Based on screen size → client:media="(min-width: 768px)"
└─ SSR breaks the component → client:only="react"
```

### Where should images go?
```
Is the image...
├─ Optimized by Astro? → src/assets/ + <Image />
├─ Static favicon/robots.txt? → public/
└─ Remote (CDN)? → <Image src="https://..." />
```

---

## Pre-Deployment Checklist

Before marking implementation complete:

- [ ] All files in correct directories per [ai-docs/03-project-structure.md](./03-project-structure.md)
- [ ] All components follow patterns from [ai-docs/07-astro-components.md](./07-astro-components.md) or [ai-docs/06-framework-components.md](./06-framework-components.md)
- [ ] Islands use correct `client:*` directives per [ai-docs/02-islands-architecture.md](./02-islands-architecture.md)
- [ ] Images optimized per [ai-docs/recipes/images.md](./recipes/images.md)
- [ ] Layouts implemented per [ai-docs/08-layouts.md](./08-layouts.md)
- [ ] Content collections configured per [ai-docs/05-content-collections.md](./05-content-collections.md)
- [ ] Build succeeds: `pnpm build`
- [ ] Preview works: `pnpm preview`
- [ ] No console errors in browser
- [ ] Core principles validated per [ai-docs/01-why-astro.md](./01-why-astro.md)

---

**Last Updated**: 2025-11-17
**Purpose**: Quick reference for AI agents implementing Astro best practices
**Companion to**: ai-docs/ modules (detailed explanations and examples)
**Usage**: Consult during planning, implementation, QA, and fixes
