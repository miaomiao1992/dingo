# Astro Project Structure

Astro projects follow an **opinionated folder layout** designed for organized development. Understanding this structure is essential for AI agents working with Astro.

## Overview

Every Astro project starts with a well-defined directory structure that separates source code, static assets, and configuration files. Astro processes files in the `src/` directory and outputs a production-ready build.

## Core Directories

### Required: `src/`

**Your project's source code lives here.** This is where Astro processes, optimizes, and bundles files to create the final website shipped to the browser.

**Contains:**
- Pages (define routes)
- Layouts (shared UI structures)
- Astro and framework components
- Styles (CSS/Sass/Less)
- Markdown/MDX content
- Images and assets for processing

**Key Point**: Everything in `src/` is processed by Astro's build system.

### Required: `src/pages/`

**The routing directory.** This subdirectory is **required** - without it, your site has no pages or routes.

**How it works:**
- Supported file types automatically become routes
- File-based routing: `src/pages/about.astro` → `/about`
- Nested routes: `src/pages/blog/post-1.astro` → `/blog/post-1`
- Dynamic routes: `src/pages/blog/[slug].astro` → `/blog/:slug`

**Supported file types:**
- `.astro` - Astro components
- `.md`, `.mdx` - Markdown/MDX content
- `.js`, `.ts` - API endpoints (SSR mode)

**Example structure:**
```
src/pages/
├── index.astro           → /
├── about.astro           → /about
├── blog/
│   ├── index.astro       → /blog
│   ├── [slug].astro      → /blog/:slug (dynamic)
│   └── post-1.astro      → /blog/post-1
└── api/
    └── users.ts          → /api/users (endpoint)
```

### Required: `public/`

**Static assets that bypass Astro's build process entirely.**

**Contains:**
- Fonts
- Icons and favicons
- `robots.txt`
- `manifest.json`
- Images that don't need processing
- Other static files

**Key Behaviors:**
- Files copy unchanged into build output
- No processing, optimization, or bundling
- Directly accessible via URL path
- Useful for files that need exact names/paths

**Example:**
```
public/
├── favicon.ico          → /favicon.ico
├── robots.txt           → /robots.txt
├── fonts/
│   └── custom.woff2     → /fonts/custom.woff2
└── images/
    └── logo.png         → /images/logo.png
```

**When to use `public/` vs `src/`:**

| Use `public/` | Use `src/` |
|---------------|------------|
| Static files that need exact paths | Images that need optimization |
| Files referenced by exact URL | Components and source code |
| Files that bypass build process | Files that need processing |
| robots.txt, manifest.json, favicons | Markdown content, layouts |

## Common Subdirectories (Conventions)

While not required, these conventions organize code effectively:

### `src/components/`

**Reusable UI components** - Both Astro and framework components.

**Example:**
```
src/components/
├── Header.astro         # Static header
├── Footer.astro         # Static footer
├── Card.astro           # Reusable card component
├── Button.jsx           # React button (interactive)
└── ui/
    ├── Input.astro
    └── Modal.svelte     # Svelte modal
```

**Best practice:**
- Keep components small and focused
- Organize by feature or type
- Use framework components only when interactivity is needed

### `src/layouts/`

**Page layout components** - Define shared structures for pages.

**Example:**
```
src/layouts/
├── BaseLayout.astro     # Base HTML structure
├── BlogLayout.astro     # Blog post layout
└── DocsLayout.astro     # Documentation layout
```

**Usage in pages:**
```astro
---
// src/pages/about.astro
import BaseLayout from '../layouts/BaseLayout.astro';
---

<BaseLayout title="About">
  <h1>About Us</h1>
  <p>Content here...</p>
</BaseLayout>
```

### `src/styles/`

**Global and shared styles** - CSS, Sass, or other styling files.

**Example:**
```
src/styles/
├── global.css           # Global styles
├── variables.css        # CSS custom properties
├── reset.css            # CSS reset
└── themes/
    ├── dark.css
    └── light.css
```

**Import in layouts:**
```astro
---
// src/layouts/BaseLayout.astro
import '../styles/global.css';
---
```

### `src/content/` (Content Collections)

**Type-safe content** - Organize Markdown/MDX with schemas.

**Example:**
```
src/content/
├── config.ts            # Collection schemas
└── blog/
    ├── post-1.md
    ├── post-2.md
    └── post-3.md
```

**Content config:**
```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';

const blog = defineCollection({
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
    author: z.string(),
  }),
});

export const collections = { blog };
```

**Benefits:**
- Type safety for frontmatter
- Validation at build time
- Better IDE support

### `src/assets/`

**Optimized assets** - Images and files that need processing.

**Example:**
```
src/assets/
├── hero.png             # Will be optimized
├── profile.jpg          # Will be optimized
└── icons/
    └── logo.svg
```

**Usage:**
```astro
---
import { Image } from 'astro:assets';
import heroImage from '../assets/hero.png';
---

<Image src={heroImage} alt="Hero" />
```

**Benefits:**
- Automatic optimization
- Responsive image generation
- Modern format conversion (WebP, AVIF)

## Configuration Files

### `package.json`

**Dependency management and scripts.**

**Example:**
```json
{
  "name": "my-astro-project",
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview"
  },
  "dependencies": {
    "astro": "^4.0.0"
  }
}
```

### `astro.config.mjs`

**Astro project configuration.**

**Supported formats:** `.js`, `.mjs`, `.cjs`, `.ts` (`.mjs` recommended)

**Example:**
```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwind from '@astrojs/tailwind';

export default defineConfig({
  integrations: [react(), tailwind()],
  site: 'https://example.com',
  base: '/my-site',
  output: 'static', // or 'server' for SSR
});
```

**Common options:**
- `integrations` - Add framework support, tools
- `site` - Production site URL
- `base` - Base path for deployment
- `output` - Static or SSR mode
- `server` - Dev server configuration

### `tsconfig.json`

**TypeScript configuration.**

**Example:**
```json
{
  "extends": "astro/tsconfigs/base",
  "compilerOptions": {
    "jsx": "react-jsx",
    "jsxImportSource": "react"
  }
}
```

**Astro presets:**
- `astro/tsconfigs/base` - Recommended base
- `astro/tsconfigs/strict` - Stricter checking
- `astro/tsconfigs/strictest` - Strictest checking

## Full Project Structure Example

```
my-astro-project/
├── public/
│   ├── favicon.ico
│   ├── robots.txt
│   └── fonts/
│       └── custom.woff2
├── src/
│   ├── assets/
│   │   └── hero.png
│   ├── components/
│   │   ├── Header.astro
│   │   ├── Footer.astro
│   │   └── ui/
│   │       ├── Button.jsx
│   │       └── Card.astro
│   ├── content/
│   │   ├── config.ts
│   │   └── blog/
│   │       ├── post-1.md
│   │       └── post-2.md
│   ├── layouts/
│   │   ├── BaseLayout.astro
│   │   └── BlogLayout.astro
│   ├── pages/
│   │   ├── index.astro
│   │   ├── about.astro
│   │   └── blog/
│   │       ├── index.astro
│   │       └── [slug].astro
│   └── styles/
│       ├── global.css
│       └── variables.css
├── astro.config.mjs
├── package.json
├── tsconfig.json
└── README.md
```

## Best Practices for AI Agents

### 1. Respect the Structure

Always place files in the correct directories:

```astro
<!-- Good: Component in components/ -->
src/components/Button.astro

<!-- Bad: Component in pages/ (unless it's a page) -->
src/pages/Button.astro
```

### 2. Use Appropriate Directories

**Decision tree:**

- **Reusable component?** → `src/components/`
- **Page layout?** → `src/layouts/`
- **New route/page?** → `src/pages/`
- **Static file (exact path needed)?** → `public/`
- **Image needing optimization?** → `src/assets/`
- **Markdown content with schema?** → `src/content/`
- **Global styles?** → `src/styles/`

### 3. Organize Components

Group related components:

```
src/components/
├── layout/
│   ├── Header.astro
│   ├── Footer.astro
│   └── Sidebar.astro
├── ui/
│   ├── Button.astro
│   ├── Input.astro
│   └── Modal.astro
└── features/
    ├── Newsletter.astro
    └── ContactForm.jsx
```

### 4. Use Content Collections

For structured content, always use Content Collections:

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';

const blog = defineCollection({
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
    tags: z.array(z.string()),
  }),
});
```

### 5. Optimize Asset Location

**Use `src/assets/` for:**
- Images that need optimization
- Responsive images
- Modern format conversion

**Use `public/` for:**
- Favicons
- robots.txt, manifest.json
- Files with specific paths (e.g., `/logo.png`)

## File Naming Conventions

### Components

```
// Use PascalCase for component files
Header.astro
Button.astro
VideoPlayer.jsx
```

### Pages

```
// Use kebab-case for page files
index.astro
about.astro
contact-us.astro
blog-post.astro

// Dynamic routes use [brackets]
[slug].astro
[id].astro
[...path].astro
```

### Styles

```
// Use kebab-case for style files
global.css
reset.css
dark-theme.css
```

## Common Patterns

### Pattern 1: Feature-Based Organization

```
src/components/
├── blog/
│   ├── PostCard.astro
│   ├── PostList.astro
│   └── Author.astro
├── shop/
│   ├── ProductCard.astro
│   └── Cart.jsx
└── shared/
    ├── Button.astro
    └── Input.astro
```

### Pattern 2: Layout Composition

```astro
---
// src/layouts/BaseLayout.astro
---
<!DOCTYPE html>
<html>
  <head>...</head>
  <body>
    <slot />
  </body>
</html>
```

```astro
---
// src/layouts/BlogLayout.astro
import BaseLayout from './BaseLayout.astro';
---

<BaseLayout>
  <article>
    <slot />
  </article>
</BaseLayout>
```

### Pattern 3: API Routes (SSR)

```typescript
// src/pages/api/users.ts
export async function get() {
  const users = await fetchUsers();
  return {
    body: JSON.stringify(users),
  };
}
```

## Key Takeaways for AI Agents

1. **`src/pages/` is required** - No pages = no routes
2. **`public/` bypasses processing** - Use for static files
3. **`src/` gets processed** - Use for everything else
4. **Follow conventions** - Use common subdirectories for organization
5. **Content Collections** - Use for type-safe content
6. **Organize logically** - Group related components and files
7. **Name consistently** - PascalCase for components, kebab-case for pages

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro project structure
**Next Module**: Development Workflow (04-development-workflow.md)
