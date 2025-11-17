# Layouts

Layouts are Astro components used to provide a reusable UI structure, such as a page template. They provide common UI elements shared across pages like headers, navigation bars, and footers.

## Overview

**What layouts provide:**
- Page shell (`<html>`, `<head>`, `<body>` tags)
- Common UI elements (header, footer, navigation)
- Shared styles and scripts
- SEO meta tags
- Consistent page structure

**Key concept:** Layouts are just Astro components with a `<slot />` - there's nothing special about them!

## What Makes a Layout?

Technically, any Astro component can be a layout. However, layouts typically:

1. **Provide a page shell**
   ```astro
   <html>
     <head>...</head>
     <body>...</body>
   </html>
   ```

2. **Include a `<slot />`** for page-specific content
   ```astro
   <main>
     <slot /> <!-- Page content goes here -->
   </main>
   ```

3. **Accept props** for page-specific data
   ```astro
   const { title, description } = Astro.props;
   ```

**But layouts can also:**
- Be partial UI templates (no full page shell)
- Include framework components
- Use client-side scripts
- Import and compose other components

## Directory Organization

**Convention:** Place layouts in `src/layouts/`

```
src/
├── layouts/
│   ├── BaseLayout.astro      # Main site layout
│   ├── BlogLayout.astro       # Blog post layout
│   └── DocsLayout.astro       # Documentation layout
├── pages/
└── components/
```

**Alternative:** Colocate with pages using `_` prefix

```
src/pages/
├── _BlogLayout.astro          # Layout for blog posts
├── index.astro
└── blog/
    └── [slug].astro
```

## Basic Layout Example

### Creating a Layout

```astro
---
// src/layouts/BaseLayout.astro
import Header from '../components/Header.astro';
import Footer from '../components/Footer.astro';

interface Props {
  title: string;
  description?: string;
}

const { title, description } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{title}</title>
    {description && <meta name="description" content={description} />}
  </head>
  <body>
    <Header />

    <main>
      <h1>{title}</h1>
      <slot /> <!-- Page content injected here -->
    </main>

    <Footer />
  </body>
</html>

<style>
  body {
    font-family: system-ui, sans-serif;
    max-width: 1200px;
    margin: 0 auto;
  }

  main {
    padding: 2rem;
  }
</style>
```

### Using a Layout

```astro
---
// src/pages/about.astro
import BaseLayout from '../layouts/BaseLayout.astro';
---

<BaseLayout title="About Us" description="Learn about our company">
  <p>We are a company that does amazing things!</p>
  <p>Founded in 2025, we've been innovating ever since.</p>
</BaseLayout>
```

**Rendered output:**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>About Us</title>
    <meta name="description" content="Learn about our company" />
  </head>
  <body>
    <header>...</header>
    <main>
      <h1>About Us</h1>
      <p>We are a company that does amazing things!</p>
      <p>Founded in 2025, we've been innovating ever since.</p>
    </main>
    <footer>...</footer>
  </body>
</html>
```

## TypeScript with Layouts

Define prop types for type safety and autocomplete:

```astro
---
// src/layouts/BlogLayout.astro
interface Props {
  title: string;
  description: string;
  publishDate: string;
  author: string;
  viewCount?: number;
  tags?: string[];
}

const {
  title,
  description,
  publishDate,
  author,
  viewCount = 0,
  tags = [],
} = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="description" content={description} />
    <title>{title}</title>
  </head>
  <body>
    <header>
      <h1>{title}</h1>
      <p>By {author} on {publishDate}</p>
      {viewCount > 0 && <p>Viewed {viewCount} times</p>}
      {tags.length > 0 && (
        <div class="tags">
          {tags.map(tag => <span class="tag">{tag}</span>)}
        </div>
      )}
    </header>

    <main>
      <slot />
    </main>
  </body>
</html>
```

**Benefits:**
- Type checking at compile time
- Editor autocomplete
- Inline documentation
- Prevents prop errors

## Markdown Layouts

Astro provides special support for Markdown files using the `layout` frontmatter property.

### Basic Markdown Layout

**Markdown file:**

```markdown
---
# src/pages/blog/post-1.md
layout: ../../layouts/MarkdownLayout.astro
title: "My First Post"
author: "Jane Doe"
date: "2025-01-15"
---

This is my blog post content written in Markdown.

## Heading

More content here...
```

**Layout component:**

```astro
---
// src/layouts/MarkdownLayout.astro
const { frontmatter } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{frontmatter.title}</title>
  </head>
  <body>
    <article>
      <header>
        <h1>{frontmatter.title}</h1>
        <p>By {frontmatter.author} on {frontmatter.date}</p>
      </header>

      <!-- Markdown content rendered here -->
      <slot />
    </article>
  </body>
</html>
```

### TypeScript with Markdown Layouts

Use the `MarkdownLayoutProps` helper for type safety:

```astro
---
// src/layouts/BlogPostLayout.astro
import type { MarkdownLayoutProps } from 'astro';

type Props = MarkdownLayoutProps<{
  // Define frontmatter props here
  title: string;
  author: string;
  date: string;
  tags?: string[];
  description?: string;
}>;

// Now frontmatter, url, and other properties are type-safe
const { frontmatter, url, headings } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="canonical" href={new URL(url, Astro.site).pathname} />
    <title>{frontmatter.title}</title>
    {frontmatter.description && (
      <meta name="description" content={frontmatter.description} />
    )}
  </head>
  <body>
    <article>
      <h1>{frontmatter.title}</h1>
      <p>By {frontmatter.author}</p>
      <p>Published: {frontmatter.date}</p>

      <!-- Table of contents from headings -->
      {headings.length > 0 && (
        <nav class="toc">
          <h2>Table of Contents</h2>
          <ul>
            {headings.map(h => (
              <li style={`margin-left: ${(h.depth - 1) * 1}rem`}>
                <a href={`#${h.slug}`}>{h.text}</a>
              </li>
            ))}
          </ul>
        </nav>
      )}

      <slot />

      {frontmatter.tags && (
        <footer>
          <p>Tags: {frontmatter.tags.join(', ')}</p>
        </footer>
      )}
    </article>
  </body>
</html>
```

### Markdown Layout Props

Markdown layouts receive these props via `Astro.props`:

| Property | Type | Description |
|----------|------|-------------|
| `file` | `string` | Absolute file path (e.g., `/home/user/.../file.md`) |
| `url` | `string` | Page URL (e.g., `/en/guides/markdown`) |
| `frontmatter` | `object` | All frontmatter from the Markdown document |
| `frontmatter.file` | `string` | Same as top-level `file` |
| `frontmatter.url` | `string` | Same as top-level `url` |
| `headings` | `array` | List of headings with metadata |
| `rawContent()` | `function` | Returns raw Markdown as string |
| `compiledContent()` | `async function` | Returns compiled HTML as string |

**Headings structure:**

```typescript
{
  depth: number;  // 1-6 for h1-h6
  slug: string;   // URL fragment
  text: string;   // Heading text
}[]
```

**Example using props:**

```astro
---
const { frontmatter, headings, url, rawContent } = Astro.props;

// Get reading time estimate
const wordCount = rawContent().split(/\s+/).length;
const readingTime = Math.ceil(wordCount / 200); // Assume 200 words/min
---

<article>
  <h1>{frontmatter.title}</h1>
  <p>Reading time: {readingTime} min</p>
  <p>Canonical URL: {url}</p>

  <!-- Headings preview -->
  <ul>
    {headings.map(h => <li>{h.text}</li>)}
  </ul>

  <slot />
</article>
```

## MDX Layouts

MDX files support layouts in two ways:

### 1. Frontmatter Layout Property

```mdx
---
# src/pages/post.mdx
layout: ../../layouts/BaseLayout.astro
title: 'My MDX Post'
---

MDX content here...
```

**Note:** This automatically passes `frontmatter` and `headings` props to the layout, just like Markdown files.

### 2. Manual Import (Recommended)

Import and use the layout component directly for more control:

```mdx
---
# src/pages/post.mdx
title: 'My MDX Post'
publishDate: '2025-01-15'
---

import BaseLayout from '../../layouts/BaseLayout.astro';

export function fancyJsHelper() {
  return "MDX can do this!";
}

<BaseLayout
  title={frontmatter.title}
  date={frontmatter.publishDate}
  fancyJsHelper={fancyJsHelper}
>
  # Welcome to my MDX blog!

  This content will be injected into the layout's `<slot />`.

  {fancyJsHelper()}
</BaseLayout>
```

**Layout receives props:**

```astro
---
// src/layouts/BaseLayout.astro
const { title, date, fancyJsHelper } = Astro.props;
---

<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    <h1>{title}</h1>
    <p>Published: {date}</p>
    <slot />
    <p>{fancyJsHelper()}</p>
  </body>
</html>
```

**Important:** When using manual import, you **must** include `<meta charset="utf-8">` in your layout, as Astro won't add it automatically to MDX pages.

## Nesting Layouts

Layouts can be nested to create flexible, composable page templates.

### Pattern: Base + Specific Layouts

**Base layout** (site-wide structure):

```astro
---
// src/layouts/BaseLayout.astro
interface Props {
  title: string;
  description?: string;
}

const { title, description } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{title} | My Site</title>
    {description && <meta name="description" content={description} />}

    <!-- Global styles -->
    <link rel="stylesheet" href="/global.css" />
  </head>
  <body>
    <header>
      <nav>
        <a href="/">Home</a>
        <a href="/blog">Blog</a>
        <a href="/about">About</a>
      </nav>
    </header>

    <main>
      <slot />
    </main>

    <footer>
      <p>&copy; 2025 My Site</p>
    </footer>
  </body>
</html>
```

**Blog-specific layout** (wraps base layout):

```astro
---
// src/layouts/BlogPostLayout.astro
import BaseLayout from './BaseLayout.astro';

interface Props {
  title: string;
  author: string;
  publishDate: string;
  tags: string[];
}

const { title, author, publishDate, tags } = Astro.props;
---

<BaseLayout title={title} description={`Blog post by ${author}`}>
  <article class="blog-post">
    <header>
      <h1>{title}</h1>
      <div class="meta">
        <p>By {author} on {publishDate}</p>
        <div class="tags">
          {tags.map(tag => (
            <a href={`/tags/${tag}`} class="tag">{tag}</a>
          ))}
        </div>
      </div>
    </header>

    <div class="content">
      <slot />
    </div>

    <footer>
      <p>Thanks for reading!</p>
    </footer>
  </article>
</BaseLayout>

<style>
  .blog-post {
    max-width: 65ch;
    margin: 0 auto;
  }

  .meta {
    color: #666;
    margin-bottom: 2rem;
  }

  .tags {
    display: flex;
    gap: 0.5rem;
  }

  .tag {
    background: #f0f0f0;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    text-decoration: none;
  }
</style>
```

**Usage:**

```astro
---
// src/pages/blog/my-post.astro
import BlogPostLayout from '../../layouts/BlogPostLayout.astro';
---

<BlogPostLayout
  title="My Amazing Post"
  author="Jane Doe"
  publishDate="2025-01-15"
  tags={["astro", "web-dev"]}
>
  <p>This is my blog post content!</p>
  <p>It gets wrapped in the BlogPostLayout,</p>
  <p>which is wrapped in the BaseLayout.</p>
</BlogPostLayout>
```

### Pattern: Markdown with Nested Layouts

```markdown
---
# src/pages/blog/post.md
layout: ../../layouts/BlogPostLayout.astro
title: "My Post"
author: "Jane Doe"
publishDate: "2025-01-15"
tags: ["astro", "markdown"]
---

Blog post content here...
```

```astro
---
// src/layouts/BlogPostLayout.astro
import BaseLayout from './BaseLayout.astro';

const { frontmatter } = Astro.props;
---

<BaseLayout title={frontmatter.title}>
  <article>
    <h1>{frontmatter.title}</h1>
    <p>By {frontmatter.author}</p>
    <slot />
  </article>
</BaseLayout>
```

**Result:** Markdown content → BlogPostLayout → BaseLayout → Final HTML

## Layout Patterns

### Pattern 1: SEO-Optimized Layout

```astro
---
// src/layouts/SEOLayout.astro
interface Props {
  title: string;
  description: string;
  image?: string;
  article?: boolean;
  publishDate?: string;
}

const {
  title,
  description,
  image = '/default-og.png',
  article = false,
  publishDate,
} = Astro.props;

const canonicalURL = new URL(Astro.url.pathname, Astro.site);
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />

    <!-- Primary Meta Tags -->
    <title>{title}</title>
    <meta name="title" content={title} />
    <meta name="description" content={description} />
    <link rel="canonical" href={canonicalURL} />

    <!-- Open Graph / Facebook -->
    <meta property="og:type" content={article ? 'article' : 'website'} />
    <meta property="og:url" content={canonicalURL} />
    <meta property="og:title" content={title} />
    <meta property="og:description" content={description} />
    <meta property="og:image" content={new URL(image, Astro.site)} />
    {article && publishDate && (
      <meta property="article:published_time" content={publishDate} />
    )}

    <!-- Twitter -->
    <meta property="twitter:card" content="summary_large_image" />
    <meta property="twitter:url" content={canonicalURL} />
    <meta property="twitter:title" content={title} />
    <meta property="twitter:description" content={description} />
    <meta property="twitter:image" content={new URL(image, Astro.site)} />
  </head>
  <body>
    <slot />
  </body>
</html>
```

### Pattern 2: Multi-Slot Layout

```astro
---
// src/layouts/DashboardLayout.astro
const { title } = Astro.props;
---

<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    <div class="dashboard">
      <aside class="sidebar">
        <slot name="sidebar" />
      </aside>

      <header>
        <h1>{title}</h1>
        <slot name="actions" />
      </header>

      <main>
        <slot />
      </main>

      <footer>
        <slot name="footer">
          <!-- Fallback footer -->
          <p>Default footer content</p>
        </slot>
      </footer>
    </div>
  </body>
</html>
```

**Usage:**

```astro
<DashboardLayout title="My Dashboard">
  <nav slot="sidebar">
    <a href="/dashboard">Home</a>
    <a href="/settings">Settings</a>
  </nav>

  <div slot="actions">
    <button>New Item</button>
  </div>

  <!-- Default slot -->
  <p>Main dashboard content</p>

  <!-- Omit footer slot to use fallback -->
</DashboardLayout>
```

### Pattern 3: Conditional Layout Features

```astro
---
// src/layouts/PageLayout.astro
interface Props {
  title: string;
  showBreadcrumbs?: boolean;
  showTOC?: boolean;
  wide?: boolean;
}

const {
  title,
  showBreadcrumbs = false,
  showTOC = false,
  wide = false,
} = Astro.props;
---

<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    {showBreadcrumbs && (
      <nav class="breadcrumbs">
        <a href="/">Home</a> / <span>{title}</span>
      </nav>
    )}

    <div class:list={['container', { wide }]}>
      {showTOC && (
        <aside class="toc">
          <slot name="toc" />
        </aside>
      )}

      <main>
        <slot />
      </main>
    </div>
  </body>
</html>
```

## Best Practices for AI Agents

### 1. Use TypeScript for Layout Props

```astro
<!-- ✅ Good: Type-safe layout props -->
---
interface Props {
  title: string;
  description?: string;
}
const { title, description } = Astro.props;
---

<!-- ❌ Avoid: Untyped props -->
---
const { title, description } = Astro.props;
---
```

### 2. Provide Sensible Defaults

```astro
<!-- ✅ Good: Default values -->
---
const {
  title,
  description = "Default description",
  showSidebar = true,
} = Astro.props;
---

<!-- ❌ Avoid: No defaults (may be undefined) -->
---
const { title, description, showSidebar } = Astro.props;
---
```

### 3. Use Nested Layouts for Shared Structure

```astro
<!-- ✅ Good: Nested layouts -->
<BaseLayout>
  <BlogLayout>
    <slot />
  </BlogLayout>
</BaseLayout>

<!-- ❌ Avoid: Duplicating base structure in each layout -->
```

### 4. Make Layouts Flexible with Slots

```astro
<!-- ✅ Good: Multiple named slots -->
<slot name="header" />
<slot />
<slot name="footer" />

<!-- ❌ Avoid: Hard-coded structure only -->
<header>Hard-coded header</header>
<slot />
```

### 5. Include Essential Meta Tags

```astro
<!-- ✅ Good: Essential tags -->
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{title}</title>
</head>

<!-- ❌ Avoid: Missing essential tags -->
<head>
  <title>{title}</title>
</head>
```

## Quick Reference

```astro
<!-- Basic layout -->
---
const { title } = Astro.props;
---
<html>
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    <slot />
  </body>
</html>

<!-- TypeScript layout -->
---
interface Props {
  title: string;
  description?: string;
}
const { title, description } = Astro.props;
---

<!-- Markdown layout -->
---
const { frontmatter } = Astro.props;
---
<html>
  <head>
    <title>{frontmatter.title}</title>
  </head>
  <body>
    <h1>{frontmatter.title}</h1>
    <slot />
  </body>
</html>

<!-- Nested layouts -->
---
import BaseLayout from './BaseLayout.astro';
---
<BaseLayout>
  <article>
    <slot />
  </article>
</BaseLayout>

<!-- Multiple slots -->
<slot name="header" />
<slot />  <!-- default -->
<slot name="footer" />
```

## Resources

- [Astro Layouts Guide](https://docs.astro.build/en/guides/layouts/)
- [Markdown & MDX Guide](https://docs.astro.build/en/guides/markdown-content/)
- [MarkdownLayoutProps Type](https://docs.astro.build/en/guides/typescript/#markdownlayoutprops)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro layout components
**See Also**:
- 07-astro-components.md for component basics
- 05-content-collections.md for Markdown/MDX content
