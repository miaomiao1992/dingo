# Style Rendered Markdown with Tailwind Typography

You can use Tailwind's Typography plugin to style rendered Markdown from sources such as Astro's content collections with beautiful, pre-designed typographic defaults.

## Overview

**Problem:** Rendered Markdown loses styling by default.

**Solution:** Use `@tailwindcss/typography` plugin to apply professional typography styles to Markdown content with a single class.

## Prerequisites

An Astro project that:
- Has Tailwind's Vite plugin installed
- Uses Astro's content collections (or other Markdown sources)

## Installation

### Step 1: Install Tailwind (if not already installed)

```bash
# Using pnpm
pnpm astro add tailwind

# Using npm
npm run astro add tailwind

# Using bun
bun astro add tailwind
```

### Step 2: Install Typography Plugin

```bash
# pnpm
pnpm add -D @tailwindcss/typography

# npm
npm install -D @tailwindcss/typography

# bun
bun add -d @tailwindcss/typography
```

### Step 3: Configure Tailwind

Add the plugin to your CSS file:

```css
// src/styles/global.css
@import 'tailwindcss';
@plugin '@tailwindcss/typography';
```

**Alternative (older Tailwind config):**

```javascript
// tailwind.config.mjs
export default {
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
```

## Create a Reusable Component

### Basic Prose Component

Create a wrapper component that applies typography styles:

```astro
---
// src/components/Prose.astro
---
<div class="prose">
  <slot />
</div>
```

**Usage:**

```astro
---
import Prose from '../components/Prose.astro';
import { getEntry, render } from 'astro:content';

const entry = await getEntry('blog', 'my-post');
const { Content } = await render(entry);
---

<Prose>
  <Content />
</Prose>
```

### Styled Prose Component

Add Tailwind modifiers for customization:

```astro
---
// src/components/Prose.astro
---
<div
  class="prose dark:prose-invert
  prose-h1:font-bold prose-h1:text-xl
  prose-a:text-blue-600 prose-p:text-justify
  prose-img:rounded-xl prose-headings:underline">
  <slot />
</div>
```

**What this does:**
- `prose` - Applies base typography styles
- `dark:prose-invert` - Dark mode support
- `prose-h1:font-bold` - Makes all `<h1>` bold
- `prose-h1:text-xl` - Makes all `<h1>` extra large
- `prose-a:text-blue-600` - Blue links
- `prose-p:text-justify` - Justified paragraphs
- `prose-img:rounded-xl` - Rounded images
- `prose-headings:underline` - Underlined headings

## Element Modifiers

The Typography plugin uses element modifiers to target specific HTML elements.

### Syntax

```
prose-[element]:class-to-apply
```

### Common Element Modifiers

```css
/* Headings */
prose-h1:...
prose-h2:...
prose-h3:...
prose-h4:...
prose-headings:...     /* All headings */

/* Text */
prose-p:...            /* Paragraphs */
prose-strong:...       /* Bold text */
prose-em:...           /* Italic text */
prose-code:...         /* Inline code */
prose-pre:...          /* Code blocks */
prose-blockquote:...   /* Blockquotes */

/* Links */
prose-a:...            /* Links */

/* Lists */
prose-ul:...           /* Unordered lists */
prose-ol:...           /* Ordered lists */
prose-li:...           /* List items */

/* Media */
prose-img:...          /* Images */
prose-video:...        /* Videos */
prose-figure:...       /* Figures */
prose-figcaption:...   /* Figure captions */

/* Tables */
prose-table:...
prose-thead:...
prose-tr:...
prose-th:...
prose-td:...

/* Other */
prose-hr:...           /* Horizontal rules */
prose-lead:...         /* Lead paragraph */
```

## Customization Examples

### Example 1: Blog Post Style

```astro
---
// src/components/BlogProse.astro
---
<div class="prose lg:prose-xl
  prose-headings:font-serif prose-headings:font-bold
  prose-p:text-gray-700 prose-p:leading-relaxed
  prose-a:text-indigo-600 prose-a:no-underline hover:prose-a:underline
  prose-code:text-pink-600 prose-code:bg-gray-100 prose-code:px-1 prose-code:rounded
  prose-img:rounded-lg prose-img:shadow-lg
  prose-blockquote:border-l-indigo-500 prose-blockquote:italic">
  <slot />
</div>
```

### Example 2: Documentation Style

```astro
---
// src/components/DocsProse.astro
---
<div class="prose prose-slate max-w-none
  prose-headings:scroll-mt-28 prose-headings:font-display
  prose-lead:text-slate-500
  prose-a:font-semibold prose-a:text-sky-500
  prose-pre:rounded-xl prose-pre:bg-slate-900 prose-pre:shadow-lg
  prose-code:font-mono prose-code:text-sm
  prose-img:rounded-lg
  prose-hr:border-slate-200">
  <slot />
</div>
```

### Example 3: Dark Mode Support

```astro
---
// src/components/Prose.astro
---
<div class="prose
  dark:prose-invert
  prose-headings:text-gray-900 dark:prose-headings:text-gray-100
  prose-p:text-gray-800 dark:prose-p:text-gray-200
  prose-a:text-blue-600 dark:prose-a:text-blue-400
  prose-code:text-pink-600 dark:prose-code:text-pink-400
  prose-code:bg-gray-100 dark:prose-code:bg-gray-800
  prose-pre:bg-gray-900 dark:prose-pre:bg-black
  prose-blockquote:border-gray-300 dark:prose-blockquote:border-gray-700">
  <slot />
</div>
```

## Size Modifiers

Typography plugin provides size variants:

```astro
<!-- Small text -->
<div class="prose-sm">
  <slot />
</div>

<!-- Base (default) -->
<div class="prose">
  <slot />
</div>

<!-- Large -->
<div class="prose-lg">
  <slot />
</div>

<!-- Extra large -->
<div class="prose-xl">
  <slot />
</div>

<!-- 2X large -->
<div class="prose-2xl">
  <slot />
</div>

<!-- Responsive sizes -->
<div class="prose md:prose-lg lg:prose-xl">
  <slot />
</div>
```

## Color Themes

Built-in color themes:

```astro
<!-- Slate (default) -->
<div class="prose prose-slate">
  <slot />
</div>

<!-- Gray -->
<div class="prose prose-gray">
  <slot />
</div>

<!-- Zinc -->
<div class="prose prose-zinc">
  <slot />
</div>

<!-- Neutral -->
<div class="prose prose-neutral">
  <slot />
</div>

<!-- Stone -->
<div class="prose prose-stone">
  <slot />
</div>
```

## Maximum Width

Control content width:

```astro
<!-- Default max-width -->
<div class="prose">
  <slot />
</div>

<!-- No max-width (full width) -->
<div class="prose max-w-none">
  <slot />
</div>

<!-- Custom max-width -->
<div class="prose max-w-4xl">
  <slot />
</div>

<!-- Responsive max-width -->
<div class="prose max-w-full md:max-w-3xl lg:max-w-4xl">
  <slot />
</div>
```

## Complete Example

### Layout Component

```astro
---
// src/layouts/BlogPostLayout.astro
import Prose from '../components/Prose.astro';

interface Props {
  title: string;
  pubDate: Date;
  author: string;
}

const { title, pubDate, author } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>{title}</title>
  </head>
  <body class="bg-white dark:bg-gray-900">
    <article class="mx-auto max-w-4xl px-4 py-12">
      <header class="mb-8">
        <h1 class="text-4xl font-bold mb-2">{title}</h1>
        <p class="text-gray-600 dark:text-gray-400">
          By {author} on {pubDate.toLocaleDateString()}
        </p>
      </header>

      <Prose>
        <slot />
      </Prose>
    </article>
  </body>
</html>
```

### Page Using Layout

```astro
---
// src/pages/blog/[id].astro
import BlogPostLayout from '../../layouts/BlogPostLayout.astro';
import { getEntry, render } from 'astro:content';

const { id } = Astro.params;
const post = await getEntry('blog', id);

if (!post) {
  return Astro.redirect('/404');
}

const { Content } = await render(post);
---

<BlogPostLayout
  title={post.data.title}
  pubDate={post.data.pubDate}
  author={post.data.author}>
  <Content />
</BlogPostLayout>
```

## Advanced Customization

### Custom Prose Component with Props

```astro
---
// src/components/Prose.astro
interface Props {
  size?: 'sm' | 'base' | 'lg' | 'xl' | '2xl';
  theme?: 'slate' | 'gray' | 'zinc' | 'neutral' | 'stone';
  maxWidth?: boolean;
}

const {
  size = 'base',
  theme = 'slate',
  maxWidth = true,
} = Astro.props;

const sizeClass = size === 'base' ? 'prose' : `prose-${size}`;
const themeClass = `prose-${theme}`;
const widthClass = maxWidth ? '' : 'max-w-none';
---

<div class={`prose ${sizeClass} ${themeClass} ${widthClass} dark:prose-invert`}>
  <slot />
</div>
```

**Usage:**

```astro
<!-- Small, gray theme, full width -->
<Prose size="sm" theme="gray" maxWidth={false}>
  <Content />
</Prose>

<!-- Large, zinc theme, with max-width -->
<Prose size="lg" theme="zinc">
  <Content />
</Prose>
```

## Styling Specific Content

### Only Style Blog Content

```astro
---
import Prose from '../components/Prose.astro';
import { getEntry, render } from 'astro:content';

const blogPost = await getEntry('blog', 'my-post');
const { Content: BlogContent } = await render(blogPost);

const docsPage = await getEntry('docs', 'getting-started');
const { Content: DocsContent } = await render(docsPage);
---

<!-- Blog with prose -->
<Prose>
  <BlogContent />
</Prose>

<!-- Docs without prose (custom styling) -->
<div class="custom-docs-styles">
  <DocsContent />
</div>
```

## Best Practices for AI Agents

### 1. Always Wrap Markdown Content

```astro
<!-- Good: Wrapped in Prose -->
<Prose>
  <Content />
</Prose>

<!-- Bad: Unstyled Markdown -->
<Content />
```

### 2. Use Consistent Prose Component

Create one reusable component rather than inline classes:

```astro
<!-- Good: Reusable component -->
<Prose>
  <Content />
</Prose>

<!-- Avoid: Inline prose classes everywhere -->
<div class="prose prose-lg...">
  <Content />
</div>
```

### 3. Customize for Content Type

Different content types need different styles:

```astro
<BlogProse>     <!-- Blog posts -->
<DocsProse>     <!-- Documentation -->
<ArticleProse>  <!-- Articles -->
```

### 4. Support Dark Mode

Always include dark mode variant:

```astro
<div class="prose dark:prose-invert">
  <slot />
</div>
```

### 5. Make It Responsive

Use responsive size modifiers:

```astro
<div class="prose md:prose-lg lg:prose-xl">
  <slot />
</div>
```

## Common Issues and Solutions

### Issue: Styles Not Applying

**Solution:** Ensure Typography plugin is loaded:

```css
// src/styles/global.css
@import 'tailwindcss';
@plugin '@tailwindcss/typography';
```

### Issue: Dark Mode Not Working

**Solution:** Add `dark:prose-invert`:

```astro
<div class="prose dark:prose-invert">
  <slot />
</div>
```

### Issue: Content Too Wide

**Solution:** Prose has default max-width. Remove if needed:

```astro
<!-- Keep default width -->
<div class="prose">
  <slot />
</div>

<!-- Full width -->
<div class="prose max-w-none">
  <slot />
</div>
```

## Resources

- [Tailwind Typography Documentation](https://tailwindcss.com/docs/typography-plugin)
- [Tailwind Typography GitHub](https://github.com/tailwindlabs/tailwindcss-typography)
- [Customizing Typography Styles](https://github.com/tailwindlabs/tailwindcss-typography#customization)

## Quick Reference

```astro
<!-- Basic usage -->
<div class="prose">
  <Content />
</div>

<!-- With dark mode -->
<div class="prose dark:prose-invert">
  <Content />
</div>

<!-- Responsive sizes -->
<div class="prose md:prose-lg lg:prose-xl">
  <Content />
</div>

<!-- Custom element styles -->
<div class="prose
  prose-headings:font-bold
  prose-a:text-blue-600
  prose-img:rounded-lg">
  <Content />
</div>

<!-- Full width -->
<div class="prose max-w-none">
  <Content />
</div>
```

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent guide for styling Markdown with Tailwind Typography
**See Also**: 05-content-collections.md for Content Collections basics
