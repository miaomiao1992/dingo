# MDX Integration

## Overview

MDX is a powerful format that combines Markdown with JSX, allowing you to use variables, expressions, and components directly in your Markdown content. The `@astrojs/mdx` integration enables full MDX support in Astro with content collections, frontmatter, and all standard Astro features.

## Why MDX?

### MDX vs Standard Markdown vs Markdoc

**Standard Markdown:**
- ✓ Simple, universal syntax
- ✓ Fast parsing
- ✗ No components without integrations
- ✗ Limited dynamic features

**MDX:**
- ✓ JSX syntax (familiar to React developers)
- ✓ Import and use components inline
- ✓ JavaScript expressions in content
- ✓ Export data from content files
- ✗ More complex syntax
- ✗ Requires JavaScript knowledge

**Markdoc:**
- ✓ Custom tag syntax (more Markdown-like)
- ✓ Type-safe attributes
- ✓ Variables and conditionals
- ✗ Less familiar syntax
- ✗ Smaller ecosystem

### When to Use MDX

**Use MDX when:**
- Coming from React/JSX ecosystem
- Need inline JavaScript expressions
- Want to export data from content files
- Prefer component syntax like `<Component />`
- Need complex interactive documentation

**Use Markdoc when:**
- Building documentation sites
- Want type-safe tag attributes
- Prefer `{% tag %}` syntax over `<Component />`
- Need reusable partials

**Use plain Markdown when:**
- Simple blog or article content
- No need for components
- Maximum compatibility

## Installation

### Automatic Installation

Using the Astro CLI (recommended):

```bash
# npm
npx astro add mdx

# pnpm
pnpm astro add mdx

# Yarn
yarn astro add mdx
```

This automatically:
1. Installs `@astrojs/mdx`
2. Updates `astro.config.mjs`
3. Enables `.mdx` file support

### Manual Installation

**1. Install Package:**

```bash
pnpm add @astrojs/mdx
```

**2. Update Config:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';

export default defineConfig({
  integrations: [mdx()],
});
```

## Editor Integration

### VS Code

**Install Extension:**

Search for "MDX" in VS Code extensions or install the [official MDX extension](https://marketplace.visualstudio.com/items?itemName=unifiedjs.vscode-mdx).

Features:
- Syntax highlighting
- IntelliSense for components
- Error detection
- Format on save

### Other Editors

Use the [MDX language server](https://github.com/mdx-js/mdx-analyzer) for:
- Neovim
- Sublime Text
- Emacs
- Other LSP-compatible editors

## Basic Usage

### Content Collections with MDX

**1. Configure Collection:**

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({ pattern: "**/*.{md,mdx}", base: "./src/blog" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.coerce.date(),
    author: z.string(),
  }),
});

export const collections = { blog };
```

**2. Create MDX Files:**

```mdx
---
title: Getting Started with MDX
description: Learn how to use MDX in Astro
pubDate: 2025-11-17
author: Alice
---

import { Code } from 'astro:components';
import Callout from '../../components/Callout.astro';

# {frontmatter.title}

By **{frontmatter.author}** on {frontmatter.pubDate.toLocaleDateString()}

<Callout type="tip">
  MDX lets you use components directly in your Markdown!
</Callout>

Here's some code:

<Code code={`console.log('Hello, MDX!');`} lang="javascript" />
```

**3. Render in Page:**

```astro
---
// src/pages/blog/[...slug].astro
import { getEntry, render } from 'astro:content';

export async function getStaticPaths() {
  return [
    { params: { slug: 'getting-started' } },
  ];
}

const { slug } = Astro.params;
const entry = await getEntry('blog', slug);
const { Content } = await render(entry);
---

<article>
  <h1>{entry.data.title}</h1>
  <p>{entry.data.description}</p>
  <Content />
</article>
```

## Exported Variables

MDX files can export JavaScript variables for use both within the file and when imported elsewhere.

### Export from MDX

```mdx
<!-- src/blog/posts/post-1.mdx -->
export const title = 'My First MDX Post';
export const tags = ['astro', 'mdx', 'tutorial'];
export const metadata = {
  readingTime: 5,
  difficulty: 'beginner',
};

# {title}

This post is tagged: {tags.join(', ')}

Reading time: {metadata.readingTime} minutes
```

### Import Exported Variables

```astro
---
// src/pages/blog/index.astro
const posts = await import.meta.glob('./posts/*.mdx', { eager: true });
const postList = Object.values(posts);
---

<ul>
  {postList.map(post => (
    <li>
      <h2>{post.title}</h2>
      <p>Tags: {post.tags.join(', ')}</p>
    </li>
  ))}
</ul>
```

### Available Properties

When importing an MDX file, these properties are available:

```typescript
interface MDXInstance {
  // File system path
  file: string; // /home/user/project/src/blog/post.mdx

  // URL (if page/content entry)
  url?: string; // /blog/post

  // Frontmatter data
  frontmatter: {
    title: string;
    // ... other frontmatter fields
  };

  // Async function returning headings
  getHeadings: () => Promise<{
    depth: number;
    slug: string;
    text: string;
  }[]>;

  // Component to render content
  Content: AstroComponent;

  // Any exported variables
  [key: string]: any;
}
```

**Example:**

```astro
---
import { Content, frontmatter, getHeadings } from './post.mdx';

const headings = await getHeadings();
---

<h1>{frontmatter.title}</h1>

<!-- Table of Contents -->
<nav>
  <ul>
    {headings.map(({ slug, text, depth }) => (
      <li style={`margin-left: ${(depth - 1) * 1}rem`}>
        <a href={`#${slug}`}>{text}</a>
      </li>
    ))}
  </ul>
</nav>

<Content />
```

## Frontmatter in MDX

### Basic Frontmatter

```mdx
---
title: My Blog Post
author: Alice
date: 2025-11-17
tags: [astro, mdx]
featured: true
---

# {frontmatter.title}

Written by {frontmatter.author} on {frontmatter.date}

{frontmatter.featured && (
  <div class="badge">⭐ Featured Post</div>
)}
```

### Accessing Frontmatter

Within the MDX file, use `frontmatter` object:

```mdx
---
title: Dynamic Title
showTOC: true
---

# {frontmatter.title}

{frontmatter.showTOC && (
  <aside>
    Table of contents will appear here
  </aside>
)}
```

### TypeScript Support

```typescript
// src/content/config.ts
import { z, defineCollection } from 'astro:content';

const blog = defineCollection({
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.date(),
    author: z.string(),
    tags: z.array(z.string()),
    featured: z.boolean().default(false),
  }),
});
```

## Using Components in MDX

### Astro Components

**1. Import Component:**

```mdx
---
title: Component Example
---

import Callout from '../../components/Callout.astro';
import CodeBlock from '../../components/CodeBlock.astro';

# Using Astro Components

<Callout type="info">
  This is an **info** callout with Markdown formatting!
</Callout>

<CodeBlock lang="javascript">
  {`const greeting = 'Hello, MDX!';
console.log(greeting);`}
</CodeBlock>
```

**2. Component with Props:**

```astro
---
// src/components/Callout.astro
interface Props {
  type?: 'info' | 'tip' | 'warning' | 'danger';
}

const { type = 'info' } = Astro.props;
---

<aside class={`callout callout-${type}`}>
  <slot />
</aside>

<style>
  .callout {
    padding: 1rem;
    border-radius: 0.5rem;
    margin: 1.5rem 0;
  }
  .callout-info { background: #e3f2fd; }
  .callout-tip { background: #e8f5e9; }
  .callout-warning { background: #fff3e0; }
  .callout-danger { background: #ffebee; }
</style>
```

### Framework Components

**React Example:**

```mdx
---
title: Interactive Components
---

import Counter from '../../components/Counter.tsx';
import Form from '../../components/Form.tsx';

# Interactive Examples

## Counter Component

<Counter client:load initial={5} />

## Form Component

<Form client:visible />
```

**Component:**

```tsx
// src/components/Counter.tsx
import { useState } from 'react';

interface Props {
  initial?: number;
}

export default function Counter({ initial = 0 }: Props) {
  const [count, setCount] = useState(initial);

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>+</button>
      <button onClick={() => setCount(count - 1)}>-</button>
    </div>
  );
}
```

### Multiple Imports

```mdx
import { Image } from 'astro:assets';
import { Code } from 'astro:components';
import Tabs from '../../components/Tabs.astro';
import Tab from '../../components/Tab.astro';
import hero from '../../assets/hero.png';

<Image src={hero} alt="Hero image" />

<Tabs>
  <Tab label="JavaScript">
    <Code code={`console.log('Hello');`} lang="js" />
  </Tab>
  <Tab label="TypeScript">
    <Code code={`console.log('Hello');`} lang="ts" />
  </Tab>
</Tabs>
```

## Custom Component Mapping

Override default HTML elements with custom components.

### Basic Mapping

**1. Create Custom Component:**

```astro
---
// src/components/CustomBlockquote.astro
---
<blockquote class="fancy-quote">
  <span class="quote-mark">"</span>
  <slot />
</blockquote>

<style>
  .fancy-quote {
    border-left: 4px solid var(--accent);
    padding-left: 2rem;
    position: relative;
    font-style: italic;
  }

  .quote-mark {
    position: absolute;
    left: 0.5rem;
    font-size: 3rem;
    color: var(--accent);
    opacity: 0.3;
  }
</style>
```

**2. Map in MDX File:**

```mdx
import CustomBlockquote from '../../components/CustomBlockquote.astro';

export const components = {
  blockquote: CustomBlockquote
};

# Regular Markdown

> This blockquote will use the CustomBlockquote component automatically!

Regular paragraph text.

> Another fancy blockquote.
```

### Multiple Element Mappings

```mdx
import CustomHeading from '../../components/CustomHeading.astro';
import CustomLink from '../../components/CustomLink.astro';
import CustomCode from '../../components/CustomCode.astro';

export const components = {
  h1: CustomHeading,
  h2: CustomHeading,
  a: CustomLink,
  code: CustomCode,
};

# This is a custom H1

## This is a custom H2

[This link](https://example.com) uses CustomLink

Inline `code` uses CustomCode
```

### Available HTML Elements

You can override any of these elements:

```javascript
export const components = {
  // Headings
  h1: CustomH1,
  h2: CustomH2,
  h3: CustomH3,
  h4: CustomH4,
  h5: CustomH5,
  h6: CustomH6,

  // Text
  p: CustomParagraph,
  strong: CustomStrong,
  em: CustomEmphasis,
  code: CustomCode,
  pre: CustomPre,

  // Links
  a: CustomLink,

  // Lists
  ul: CustomUnorderedList,
  ol: CustomOrderedList,
  li: CustomListItem,

  // Other
  blockquote: CustomBlockquote,
  img: CustomImage,
  hr: CustomHR,
  table: CustomTable,
};
```

## Passing Components to Imported MDX

When rendering imported MDX content, pass custom components via the `components` prop.

### Direct Import

```astro
---
// src/pages/docs.astro
import { Content, components as mdxComponents } from '../content/intro.mdx';
import CustomHeading from '../components/CustomHeading.astro';
---

<!-- Override h1, keep other exported components -->
<Content components={{
  ...mdxComponents,
  h1: CustomHeading
}} />
```

### Content Collections

```astro
---
// src/pages/blog/[slug].astro
import { getEntry, render } from 'astro:content';
import CustomHeading from '../../components/CustomHeading.astro';
import CustomBlockquote from '../../components/CustomBlockquote.astro';

const entry = await getEntry('blog', Astro.params.slug);
const { Content } = await render(entry);
---

<article>
  <Content components={{
    h1: CustomHeading,
    h2: CustomHeading,
    blockquote: CustomBlockquote,
  }} />
</article>
```

### Dynamic Component Selection

```astro
---
const entry = await getEntry('blog', Astro.params.slug);
const { Content } = await render(entry);

// Use different components based on frontmatter
const components = entry.data.fancy
  ? {
      blockquote: FancyBlockquote,
      h1: FancyHeading,
    }
  : {
      blockquote: SimpleBlockquote,
      h1: SimpleHeading,
    };
---

<Content components={components} />
```

## Configuration

### Inheriting Markdown Config

By default, MDX inherits your Markdown configuration:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';

export default defineConfig({
  markdown: {
    syntaxHighlight: 'shiki',
    shikiConfig: { theme: 'nord' },
    remarkPlugins: [remarkPlugin1],
    rehypePlugins: [rehypePlugin1],
  },
  integrations: [
    mdx({
      // Inherits all markdown config by default
    }),
  ],
});
```

### Override Specific Options

```javascript
// astro.config.mjs
export default defineConfig({
  markdown: {
    syntaxHighlight: 'prism',
    remarkPlugins: [remarkPlugin1],
    gfm: true,
  },
  integrations: [
    mdx({
      // Override: use Shiki instead of Prism
      syntaxHighlight: 'shiki',
      shikiConfig: { theme: 'dracula' },

      // Override: different remark plugins
      remarkPlugins: [remarkPlugin2],

      // Override: disable GFM
      gfm: false,

      // Inherited: rehypePlugins from markdown config
    }),
  ],
});
```

### Disable Markdown Config Extension

```javascript
// astro.config.mjs
export default defineConfig({
  markdown: {
    remarkPlugins: [remarkPlugin1],
  },
  integrations: [
    mdx({
      extendMarkdownConfig: false,
      // Now using default MDX config
      // remarkPlugin1 is NOT applied
    }),
  ],
});
```

## Remark and Rehype Plugins

### Using Remark Plugins

Remark plugins transform Markdown AST.

**Install Plugin:**

```bash
pnpm add remark-gfm remark-toc
```

**Configure:**

```javascript
// astro.config.mjs
import remarkGfm from 'remark-gfm';
import remarkToc from 'remark-toc';

export default defineConfig({
  integrations: [
    mdx({
      remarkPlugins: [
        remarkGfm,
        [remarkToc, { heading: 'contents', maxDepth: 3 }],
      ],
    }),
  ],
});
```

### Using Rehype Plugins

Rehype plugins transform HTML AST.

**Install Plugin:**

```bash
pnpm add rehype-autolink-headings rehype-slug
```

**Configure:**

```javascript
// astro.config.mjs
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';

export default defineConfig({
  integrations: [
    mdx({
      rehypePlugins: [
        rehypeSlug,
        [rehypeAutolinkHeadings, { behavior: 'wrap' }],
      ],
    }),
  ],
});
```

### Popular Plugins

**Remark:**
- `remark-gfm` - GitHub Flavored Markdown
- `remark-toc` - Generate table of contents
- `remark-math` - Math support
- `remark-emoji` - Emoji shortcodes

**Rehype:**
- `rehype-slug` - Add IDs to headings
- `rehype-autolink-headings` - Add links to headings
- `rehype-accessible-emojis` - Accessible emoji
- `rehype-external-links` - Handle external links

## Recma Plugins

Recma plugins modify the JavaScript output (ESTree).

```javascript
// astro.config.mjs
export default defineConfig({
  integrations: [
    mdx({
      recmaPlugins: [
        // Plugin to inject variables
        () => (tree) => {
          // Modify JavaScript AST
        },
      ],
    }),
  ],
});
```

## Optimization

### Enable MDX Optimization

Optimize MDX output for faster builds and rendering:

```javascript
// astro.config.mjs
export default defineConfig({
  integrations: [
    mdx({
      optimize: true,
    }),
  ],
});
```

**Benefits:**
- Faster builds
- Smaller output
- Better runtime performance

**Trade-off:**
- May generate unescaped HTML
- Some dynamic features might break

### Ignore Custom Components

When using custom components via the `components` prop, exclude them from optimization:

```javascript
// astro.config.mjs
export default defineConfig({
  integrations: [
    mdx({
      optimize: {
        ignoreElementNames: ['h1', 'h2', 'blockquote'],
      },
    }),
  ],
});
```

**Example:**

```astro
---
import { Content } from '../content.mdx';
import CustomHeading from '../components/CustomHeading.astro';
---

<!-- h1 won't be optimized because it's in ignoreElementNames -->
<Content components={{ h1: CustomHeading }} />
```

## JavaScript Expressions

Use JavaScript expressions directly in MDX:

### Inline Expressions

```mdx
---
title: JavaScript in MDX
---

export const name = 'Alice';
export const items = ['Astro', 'MDX', 'React'];

# Hello, {name}!

## My favorite tools:

{items.map(item => `- ${item}`).join('\n')}

## Current date:

{new Date().toLocaleDateString()}

## Conditional content:

{name === 'Alice' ? (
  <p>Welcome back, Alice!</p>
) : (
  <p>Hello, guest!</p>
)}
```

### Complex Expressions

```mdx
export const posts = [
  { title: 'Post 1', likes: 10 },
  { title: 'Post 2', likes: 25 },
  { title: 'Post 3', likes: 5 },
];

export const totalLikes = posts.reduce((sum, post) => sum + post.likes, 0);

# Blog Statistics

Total posts: {posts.length}
Total likes: {totalLikes}
Average likes: {(totalLikes / posts.length).toFixed(1)}

## Top Posts

{posts
  .sort((a, b) => b.likes - a.likes)
  .slice(0, 2)
  .map(post => (
    <div key={post.title}>
      <h3>{post.title}</h3>
      <p>{post.likes} likes</p>
    </div>
  ))
}
```

## Best Practices

### 1. Use Frontmatter for Metadata

```mdx
---
title: Best Practices
date: 2025-11-17
tags: [mdx, best-practices]
---

# {frontmatter.title}

Published: {frontmatter.date}
```

### 2. Keep Imports at Top

```mdx
<!-- Good -->
import Component1 from './Component1.astro';
import Component2 from './Component2.astro';

# Content

<Component1 />

<!-- Bad -->
# Content

import Component1 from './Component1.astro';
<Component1 />
```

### 3. Export Reusable Data

```mdx
export const author = {
  name: 'Alice',
  avatar: '/avatars/alice.jpg',
  bio: 'Software Engineer',
};

# About {author.name}

{author.bio}
```

### 4. Use TypeScript for Complex Logic

Instead of complex expressions in MDX, create utility functions:

```typescript
// src/utils/mdx-helpers.ts
export function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function calculateReadingTime(text: string): number {
  const wordsPerMinute = 200;
  const words = text.split(/\s+/).length;
  return Math.ceil(words / wordsPerMinute);
}
```

```mdx
import { formatDate, calculateReadingTime } from '../utils/mdx-helpers';

export const content = `Your long article content...`;

Published: {formatDate(new Date(frontmatter.date))}
Reading time: {calculateReadingTime(content)} minutes
```

### 5. Component Composition

```mdx
import Card from './Card.astro';
import Button from './Button.astro';

<Card>
  <h3>Card Title</h3>
  <p>Card content</p>
  <Button>Click me</Button>
</Card>
```

## Common Patterns

### Blog Post Template

```mdx
---
title: My Blog Post
author: Alice
date: 2025-11-17
tags: [astro, mdx]
image: ./hero.jpg
---

import { Image } from 'astro:assets';
import AuthorBio from '../../components/AuthorBio.astro';
import TagList from '../../components/TagList.astro';
import RelatedPosts from '../../components/RelatedPosts.astro';
import heroImage from './hero.jpg';

<Image src={heroImage} alt={frontmatter.title} />

# {frontmatter.title}

<AuthorBio author={frontmatter.author} />

<TagList tags={frontmatter.tags} />

{/* Your content here */}

---

<RelatedPosts tags={frontmatter.tags} />
```

### Documentation Page

```mdx
---
title: API Reference
section: Getting Started
order: 1
---

import { Code } from 'astro:components';
import Tabs from '../../components/Tabs.astro';
import Callout from '../../components/Callout.astro';

# {frontmatter.title}

<Callout type="info">
  This is the API reference for version 2.0
</Callout>

## Installation

<Tabs>
  <div label="npm">
    <Code code={`npm install package`} lang="bash" />
  </div>
  <div label="pnpm">
    <Code code={`pnpm add package`} lang="bash" />
  </div>
</Tabs>
```

### Interactive Tutorial

```mdx
---
title: Interactive Tutorial
difficulty: beginner
---

import Counter from '../../components/Counter.tsx';
import Quiz from '../../components/Quiz.tsx';
import CodeEditor from '../../components/CodeEditor.tsx';

# {frontmatter.title}

## Step 1: Understanding State

Try this interactive counter:

<Counter client:load />

## Step 2: Quiz

<Quiz
  question="What happens when you click the button?"
  options={['Count increases', 'Count decreases', 'Nothing']}
  correct={0}
  client:visible
/>

## Step 3: Code Challenge

<CodeEditor
  initialCode="function greet(name) {\n  // Your code here\n}"
  client:idle
/>
```

## Troubleshooting

### Issue: Components Not Rendering

**Problem:** Components show as text instead of rendering.

**Solution:**
1. Check import path is correct
2. Ensure component is properly exported
3. Verify JSX syntax (use `<Component />` not `{% component %}`)

### Issue: Frontmatter Not Accessible

**Problem:** `frontmatter` is undefined.

**Solution:**
- Access via `frontmatter.property`, not just `property`
- Ensure frontmatter is in YAML format at top of file

### Issue: Syntax Errors in MDX

**Common mistakes:**
```mdx
<!-- Wrong: HTML comments -->
{/* Correct: JSX comments */}

<!-- Wrong: Unclosed JSX -->
<Component>

<!-- Correct: Self-closing or properly closed -->
<Component />
<Component>content</Component>
```

### Issue: Custom Components Not Optimized

**Solution:**

Add to `ignoreElementNames`:

```javascript
mdx({
  optimize: {
    ignoreElementNames: ['CustomComponent'],
  },
})
```

## Quick Reference

### Create MDX File

```mdx
---
title: My Post
---

import Component from './Component.astro';

# {frontmatter.title}

<Component prop="value" />
```

### Render MDX

```astro
---
import { Content } from '../content.mdx';
---

<Content />
```

### Export Variables

```mdx
export const title = 'My Post';
export const tags = ['astro', 'mdx'];
```

### Custom Components

```mdx
import Custom from './Custom.astro';

export const components = {
  h1: Custom
};
```

### Pass Components

```astro
<Content components={{ h1: CustomHeading }} />
```

## See Also

- [09-markdown.md](../09-markdown.md) - Markdown in Astro overview
- [recipes/markdoc.md](./markdoc.md) - Markdoc alternative
- [05-content-collections.md](../05-content-collections.md) - Content Collections
- [Official MDX Docs](https://mdxjs.com/) - Complete MDX reference
- [Astro MDX Integration](https://docs.astro.build/en/guides/integrations-guide/mdx/) - Official Astro docs

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to using MDX in Astro projects
**Key Concepts**: JSX in Markdown, component imports, frontmatter, custom element mapping, content collections
