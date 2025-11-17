# Markdown in Astro

Markdown is commonly used to author text-heavy content like blog posts and documentation. Astro includes built-in support for Markdown files with frontmatter YAML/TOML to define custom properties.

## Overview

**Astro's Markdown support:**
- Built-in GitHub Flavored Markdown
- Frontmatter for metadata (YAML or TOML)
- File imports and content collections
- `.astro` component integration
- Automatic page generation from `src/pages/`
- Remark and rehype plugin ecosystem

**For even more features** (components, JSX expressions), use the [@astrojs/mdx integration](https://docs.astro.build/en/guides/integrations-guide/mdx/).

## Organizing Markdown Files

Markdown files can be kept anywhere in `src/`:

```
src/
├── pages/
│   └── blog/
│       ├── post-1.md      # Auto-generates /blog/post-1
│       └── post-2.md      # Auto-generates /blog/post-2
├── content/
│   └── blog/
│       ├── post-1.md      # For content collections
│       └── post-2.md
└── data/
    └── changelog.md       # Import manually
```

**Locations:**
- `src/pages/` - Auto-generates pages
- `src/content/` - For content collections
- Anywhere else - Import manually

## File Imports vs Content Collections

Two ways to use Markdown content:

### Method 1: File Imports

Import Markdown directly in `.astro` components:

```astro
---
// Single file import
import * as post from './posts/great-post.md';

// Multiple files with glob
const posts = Object.values(
  import.meta.glob('./posts/*.md', { eager: true })
);
---

<h1>{post.frontmatter.title}</h1>
<post.Content />
```

**Use when:**
- One-off Markdown files
- Small number of files
- Quick prototypes

### Method 2: Content Collections

Define structured collections with schemas:

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/blog' }),
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
    author: z.string(),
  }),
});

export const collections = { blog };
```

**Use when:**
- Groups of related files
- Need validation and type safety
- Want editor autocomplete
- Building blogs, docs, etc.

**See:** [05-content-collections.md](./05-content-collections.md) for details

## Markdown File Structure

### Basic Markdown File

```markdown
---
title: My Blog Post
author: Jane Doe
pubDate: 2025-01-15
tags: ["astro", "markdown"]
---

# My Blog Post

This is the content written in **Markdown**.

## Section

- List item 1
- List item 2

[Link to Astro](https://astro.build)
```

**Parts:**
1. **Frontmatter** (between `---`) - Metadata
2. **Body** - Markdown content

### Supported Frontmatter Formats

**YAML (default):**

```yaml
---
title: "My Post"
date: 2025-01-15
tags: [astro, markdown]
---
```

**TOML:**

```toml
+++
title = "My Post"
date = 2025-01-15
tags = ["astro", "markdown"]
+++
```

## Using Markdown in Components

### Importing Single File

```astro
---
// src/pages/my-post.astro
import * as greatPost from './posts/great-post.md';
---

<html>
  <head>
    <title>{greatPost.frontmatter.title}</title>
  </head>
  <body>
    <h1>{greatPost.frontmatter.title}</h1>
    <p>By {greatPost.frontmatter.author}</p>
    <greatPost.Content />
  </body>
</html>
```

### Importing Multiple Files

```astro
---
// src/pages/blog/index.astro
const posts = Object.values(
  import.meta.glob('./posts/*.md', { eager: true })
);
---

<h1>All Blog Posts</h1>
<ul>
  {posts.map(post => (
    <li>
      <a href={post.url}>{post.frontmatter.title}</a>
      <p>By {post.frontmatter.author}</p>
    </li>
  ))}
</ul>
```

## Available Properties

### From Imports

When importing Markdown with `import`:

```typescript
{
  // File information
  file: "/home/user/projects/.../file.md",
  url: "/en/guides/markdown-content/",

  // Frontmatter data
  frontmatter: {
    title: "Astro 0.18 Release",
    date: "Tuesday, July 27 2021",
    author: "Matthew Phillips",
    description: "Astro 0.18 is our biggest release.",
  },

  // Render component
  Content: Component,

  // Utility functions
  getHeadings: () => [
    {
      depth: 1,
      text: "Astro 0.18 Release",
      slug: "astro-018-release"
    },
    {
      depth: 2,
      text: "Responsive partial hydration",
      slug: "responsive-partial-hydration"
    }
  ],

  rawContent: () => "# Astro 0.18 Release\nA little over...",

  compiledContent: () => "<h1>Astro 0.18 Release</h1>\n<p>A little over..."
}
```

**Properties:**

| Property | Type | Description |
|----------|------|-------------|
| `file` | `string` | Absolute file path |
| `url` | `string` | Page URL (if in `src/pages/`) |
| `frontmatter` | `object` | All frontmatter properties |
| `Content` | `Component` | Rendered Markdown component |
| `getHeadings()` | `function` | Returns array of headings |
| `rawContent()` | `function` | Returns raw Markdown string |
| `compiledContent()` | `async function` | Returns HTML string |

### From Content Collections

When using `getEntry()` or `getCollection()`:

```astro
---
import { getEntry, render } from 'astro:content';

const post = await getEntry('blog', 'my-post');
const { Content, headings } = await render(post);
---

<h1>{post.data.title}</h1>
<Content />
```

**Properties:**

| Property | Type | Description |
|----------|------|-------------|
| `post.id` | `string` | Entry ID |
| `post.data` | `object` | Validated frontmatter (from schema) |
| `post.body` | `string` | Raw Markdown content |
| `Content` | `Component` | From `render()` function |
| `headings` | `array` | From `render()` function |

## The `<Content />` Component

Render Markdown to HTML using the `<Content />` component:

### From Imports

```astro
---
import { Content as PromoBanner } from '../components/promoBanner.md';
---

<h2>Today's Promo</h2>
<PromoBanner />
```

**Rename Content:**

```astro
---
import { Content as MyContent } from './post.md';
---

<MyContent />
```

### From Content Collections

```astro
---
import { getEntry, render } from 'astro:content';

const product = await getEntry('products', 'shirt');
const { Content } = await render(product);
---

<h2>Product Details</h2>
<p>Sale Ends: {product.data.saleEndDate.toDateString()}</p>
<Content />
```

## Heading IDs and Anchors

Markdown headings automatically get anchor IDs:

```markdown
## Introduction

Content here...

## Conclusion

I can link to [Introduction](#introduction) on the same page.
```

**Generates:**

```html
<h2 id="introduction">Introduction</h2>
<p>Content here...</p>

<h2 id="conclusion">Conclusion</h2>
<p>I can link to <a href="#introduction">Introduction</a> on the same page.</p>
```

**ID generation:**
- Based on [github-slugger](https://github.com/Flet/github-slugger)
- Lowercase, hyphenated
- Special characters removed

**Examples:**

| Heading | Generated ID |
|---------|--------------|
| `## Hello World` | `#hello-world` |
| `## FAQ: Getting Started` | `#faq-getting-started` |
| `## What's New?` | `#whats-new` |

### Accessing Heading Data

```astro
---
import * as post from './post.md';

const headings = await post.getHeadings();
---

<nav>
  <h2>Table of Contents</h2>
  <ul>
    {headings.map(h => (
      <li style={`margin-left: ${(h.depth - 1) * 1}rem`}>
        <a href={`#${h.slug}`}>{h.text}</a>
      </li>
    ))}
  </ul>
</nav>

<post.Content />
```

**Heading structure:**

```typescript
{
  depth: number;  // 1-6 (h1-h6)
  slug: string;   // Anchor ID
  text: string;   // Heading text
}[]
```

### Custom Heading IDs with Plugins

Use `rehype-slug` for custom ID generation:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import rehypeSlug from 'rehype-slug';

export default defineConfig({
  markdown: {
    rehypePlugins: [rehypeSlug],
  },
});
```

**Using Astro's built-in plugin:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import { rehypeHeadingIds } from '@astrojs/markdown-remark';
import { otherPlugin } from 'some/plugin';

export default defineConfig({
  markdown: {
    rehypePlugins: [
      rehypeHeadingIds,  // Must come before plugins that rely on IDs
      otherPlugin,
    ],
  },
});
```

## Markdown Plugins

Astro uses [remark](https://github.com/remarkjs/remark) (parsing) and [rehype](https://github.com/rehypejs/rehype) (rendering) for Markdown processing.

**Default plugins:**
- GitHub Flavored Markdown
- SmartyPants (smart quotes, em-dashes)

### Adding Remark Plugins

Remark plugins transform Markdown AST:

```bash
pnpm add remark-toc
```

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import remarkToc from 'remark-toc';

export default defineConfig({
  markdown: {
    remarkPlugins: [remarkToc],
  },
});
```

**With options:**

```javascript
export default defineConfig({
  markdown: {
    remarkPlugins: [
      [remarkToc, { heading: 'toc', maxDepth: 3 }]
    ],
  },
});
```

### Adding Rehype Plugins

Rehype plugins transform HTML AST:

```bash
pnpm add rehype-accessible-emojis
```

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import { rehypeAccessibleEmojis } from 'rehype-accessible-emojis';

export default defineConfig({
  markdown: {
    rehypePlugins: [rehypeAccessibleEmojis],
  },
});
```

### Popular Plugins

**Remark plugins:**
- `remark-toc` - Generate table of contents
- `remark-gfm` - GitHub Flavored Markdown (included by default)
- `remark-math` - Math equations support
- `remark-emoji` - Emoji shortcodes

**Rehype plugins:**
- `rehype-slug` - Add IDs to headings
- `rehype-autolink-headings` - Add links to headings
- `rehype-accessible-emojis` - Accessible emoji labels
- `rehype-external-links` - Add target="_blank" to external links

**Browse:**
- [awesome-remark](https://github.com/remarkjs/awesome-remark)
- [awesome-rehype](https://github.com/rehypejs/awesome-rehype)

### Combined Example

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import remarkToc from 'remark-toc';
import remarkMath from 'remark-math';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';

export default defineConfig({
  markdown: {
    remarkPlugins: [
      [remarkToc, { heading: 'contents', maxDepth: 3 }],
      remarkMath,
    ],
    rehypePlugins: [
      rehypeSlug,
      [rehypeAutolinkHeadings, { behavior: 'append' }],
    ],
  },
});
```

## Modifying Frontmatter Programmatically

Add frontmatter properties via remark/rehype plugins:

### Custom Remark Plugin

```javascript
// example-remark-plugin.mjs
export function exampleRemarkPlugin() {
  return function (tree, file) {
    // Add property to frontmatter
    file.data.astro.frontmatter.customProperty = 'Generated property';

    // Calculate reading time
    const text = file.value;
    const wordsPerMinute = 200;
    const wordCount = text.split(/\s+/).length;
    const readingTime = Math.ceil(wordCount / wordsPerMinute);

    file.data.astro.frontmatter.readingTime = readingTime;
  }
}
```

**Use in config:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import { exampleRemarkPlugin } from './example-remark-plugin.mjs';

export default defineConfig({
  markdown: {
    remarkPlugins: [exampleRemarkPlugin],
  },
});
```

**Or for MDX:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';
import { exampleRemarkPlugin } from './example-remark-plugin.mjs';

export default defineConfig({
  integrations: [
    mdx({
      remarkPlugins: [exampleRemarkPlugin],
    }),
  ],
});
```

**Access in component:**

```astro
---
import * as post from './post.md';
---

<p>Reading time: {post.frontmatter.readingTime} min</p>
<p>Custom: {post.frontmatter.customProperty}</p>
```

## Individual Markdown Pages

Files in `src/pages/` automatically become pages:

```markdown
<!-- src/pages/about.md -->
---
title: About Us
description: Learn about our company
---

# About Us

We are a company that does great things!

## Our Mission

To build amazing products.
```

**Generates:** `/about` page

**Features:**
- Auto-route generation
- Automatic `<meta charset="utf-8">` tag
- Basic HTML structure

**Limitations:**
- Minimal styling
- No custom layout by default
- Limited to Markdown features

### Frontmatter `layout` Property

Add a layout wrapper to Markdown pages:

```markdown
<!-- src/pages/blog/post-1.md -->
---
layout: ../../layouts/BlogPostLayout.astro
title: My First Post
author: Jane Doe
date: 2025-01-15
---

# My First Post

Post content here...
```

**Layout component:**

```astro
---
// src/layouts/BlogPostLayout.astro
const { frontmatter } = Astro.props;
---

<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>{frontmatter.title}</title>
  </head>
  <body>
    <article>
      <h1>{frontmatter.title}</h1>
      <p>By {frontmatter.author} on {frontmatter.date}</p>

      <slot /> <!-- Markdown content -->
    </article>
  </body>
</html>
```

**Important:**
- Must include `<meta charset="utf-8">` in layout
- Astro no longer adds it automatically when using `layout` property
- See [08-layouts.md](./08-layouts.md) for more on layouts

## MDX Integration

For components and JSX in Markdown, use MDX:

```bash
pnpm astro add mdx
```

```mdx
---
title: My MDX Post
---

import Button from '../components/Button.astro';
import { Chart } from '../components/Chart.jsx';

# My Post

Regular Markdown content.

<Button>Click me</Button>

<Chart data={[1, 2, 3]} />
```

**Extending Markdown config:**

By default, MDX inherits your Markdown configuration:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';

export default defineConfig({
  markdown: {
    remarkPlugins: [remarkPlugin1],
    gfm: true,
  },
  integrations: [
    mdx({
      // Inherits markdown config by default
      // Add MDX-specific plugins:
      remarkPlugins: [mdxPlugin],
    })
  ]
});
```

**Override Markdown config:**

```javascript
integrations: [
  mdx({
    // Disable inheritance
    extendMarkdownConfig: false,

    // MDX gets its own separate config
    remarkPlugins: [mdxOnlyPlugin],
    gfm: false,
  })
]
```

## Fetching Remote Markdown

### Not Recommended: Direct Fetch

Astro's Markdown processor is **not available** for remote Markdown.

**Manual approach** (loses Astro features):

```astro
---
// src/pages/remote-example.astro
import { marked } from 'marked'; // External package

const response = await fetch(
  'https://raw.githubusercontent.com/wiki/adam-p/markdown-here/Markdown-Cheatsheet.md'
);
const markdown = await response.text();
const content = marked.parse(markdown);
---

<article set:html={content} />
```

**Limitations:**
- No Astro Markdown settings applied
- No remark/rehype plugins
- No frontmatter processing
- Manual HTML parsing

### Recommended: Content Collections Loader

Use a custom loader for remote Markdown:

```typescript
// src/content/config.ts
import { defineCollection } from 'astro:content';

const blog = defineCollection({
  loader: async () => {
    const response = await fetch('https://api.example.com/posts');
    const posts = await response.json();

    return posts.map(post => ({
      id: post.slug,
      ...post.frontmatter,
      body: post.markdown,
    }));
  },
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
  }),
});
```

**Benefits:**
- Type safety
- Validation
- Works with Astro's content APIs
- Can use `render()` function

## Best Practices for AI Agents

### 1. Use Content Collections for Structured Content

```astro
<!-- ✅ Good: Content collections for blogs -->
---
import { getCollection } from 'astro:content';
const posts = await getCollection('blog');
---

<!-- ❌ Avoid: File imports for many related files -->
---
const posts = Object.values(
  import.meta.glob('./posts/*.md', { eager: true })
);
---
```

### 2. Define Schemas for Type Safety

```typescript
// ✅ Good: Schema with validation
const blog = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/blog' }),
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
    tags: z.array(z.string()),
  }),
});

// ❌ Avoid: No schema (no type safety)
const blog = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/blog' }),
});
```

### 3. Use Layouts for Markdown Pages

```markdown
<!-- ✅ Good: Layout for consistent structure -->
---
layout: ../../layouts/PostLayout.astro
title: My Post
---

<!-- ❌ Avoid: Bare Markdown page (no styling) -->
---
title: My Post
---
```

### 4. Access Headings for TOC

```astro
<!-- ✅ Good: Generate TOC from headings -->
---
const { Content, headings } = await render(post);
---

<nav>
  {headings.map(h => (
    <a href={`#${h.slug}`}>{h.text}</a>
  ))}
</nav>
<Content />

<!-- ❌ Avoid: Manual TOC (gets out of sync) -->
<nav>
  <a href="#intro">Introduction</a>
  <a href="#conclusion">Conclusion</a>
</nav>
```

### 5. Use Plugins for Common Tasks

```javascript
// ✅ Good: Use plugins for TOC
remarkPlugins: [
  [remarkToc, { heading: 'contents' }]
]

// ❌ Avoid: Manual TOC generation in each file
```

## Quick Reference

```astro
<!-- Import single file -->
---
import * as post from './post.md';
---
<h1>{post.frontmatter.title}</h1>
<post.Content />

<!-- Import multiple files -->
---
const posts = Object.values(
  import.meta.glob('./posts/*.md', { eager: true })
);
---
{posts.map(post => <li>{post.frontmatter.title}</li>)}

<!-- Content collections -->
---
import { getCollection, render } from 'astro:content';
const posts = await getCollection('blog');
const post = posts[0];
const { Content, headings } = await render(post);
---
<h1>{post.data.title}</h1>
<Content />

<!-- Table of contents -->
<ul>
  {headings.map(h => (
    <li style={`margin-left: ${h.depth}rem`}>
      <a href={`#${h.slug}`}>{h.text}</a>
    </li>
  ))}
</ul>
```

## Resources

- [Markdown Guide](https://www.markdownguide.org/)
- [GitHub Flavored Markdown Spec](https://github.github.com/gfm/)
- [awesome-remark plugins](https://github.com/remarkjs/awesome-remark)
- [awesome-rehype plugins](https://github.com/rehypejs/awesome-rehype)
- [MDX Documentation](https://mdxjs.com/)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Markdown in Astro
**See Also**:
- 05-content-collections.md for collection details
- 08-layouts.md for Markdown layouts
- recipes/syntax-highlighting.md for code blocks
