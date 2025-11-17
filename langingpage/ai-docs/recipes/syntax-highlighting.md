# Syntax Highlighting in Astro

Astro comes with built-in support for syntax highlighting in code blocks using Shiki (default) and Prism. This guide covers all methods of highlighting code in your Astro project.

## Overview

Astro provides three ways to add syntax highlighting:

1. **Markdown code blocks** - Automatic highlighting with Shiki (default)
2. **`<Code />` component** - Powered by Shiki, for `.astro` files
3. **`<Prism />` component** - Powered by Prism, alternative highlighter

## Markdown Code Blocks

### Basic Usage

Use triple backticks with a language identifier:

````markdown
```js
// JavaScript code with syntax highlighting
var fun = function lang(l) {
  dateformat.i18n = require('./lang/' + l);
  return true;
};
```
````

**Default behavior:**
- Styled by Shiki
- `github-dark` theme (default)
- Inline styles (no external CSS needed)
- No client-side JavaScript

### Setting a Default Theme

Configure in `astro.config.mjs`:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';

export default defineConfig({
  markdown: {
    shikiConfig: {
      theme: 'dracula',  // Any built-in Shiki theme
    },
  },
});
```

**Popular themes:**
- `github-dark` (default)
- `github-light`
- `dracula`
- `nord`
- `monokai`
- `one-dark-pro`
- `material-theme-palenight`
- `min-light`
- `min-dark`

[See all Shiki themes](https://shiki.style/themes)

### Light and Dark Mode Themes

Configure dual themes for light/dark mode:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';

export default defineConfig({
  markdown: {
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'github-dark',
      },
    },
  },
});
```

**Add dark mode CSS:**

Replace `.shiki` with `.astro-code` from Shiki's documentation:

```css
/* src/styles/global.css */
@media (prefers-color-scheme: dark) {
  .astro-code,
  .astro-code span {
    color: var(--shiki-dark) !important;
    background-color: var(--shiki-dark-bg) !important;
    /* Optional: font styles */
    font-style: var(--shiki-dark-font-style) !important;
    font-weight: var(--shiki-dark-font-weight) !important;
    text-decoration: var(--shiki-dark-text-decoration) !important;
  }
}
```

**Class-based dark mode:**

```css
/* src/styles/global.css */
.dark .astro-code,
.dark .astro-code span {
  color: var(--shiki-dark) !important;
  background-color: var(--shiki-dark-bg) !important;
}
```

### Custom Shiki Theme

Import a custom theme from a local JSON file:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import customTheme from './my-shiki-theme.json';

export default defineConfig({
  markdown: {
    shikiConfig: {
      theme: customTheme,
    },
  },
});
```

**Custom theme format:**

Follow [Shiki's theme schema](https://shiki.style/guide/load-theme#loading-theme)

### CSS Variables Theme

Use CSS variables for maximum customization:

```javascript
// astro.config.mjs
export default defineConfig({
  markdown: {
    shikiConfig: {
      theme: 'css-variables',
    },
  },
});
```

**Define variables:**

```css
/* src/styles/global.css */
:root {
  --astro-code-color-text: #24292e;
  --astro-code-color-background: #ffffff;
  --astro-code-token-constant: #005cc5;
  --astro-code-token-string: #032f62;
  --astro-code-token-comment: #6a737d;
  --astro-code-token-keyword: #d73a49;
  --astro-code-token-parameter: #24292e;
  --astro-code-token-function: #6f42c1;
  --astro-code-token-string-expression: #032f62;
  --astro-code-token-punctuation: #24292e;
  --astro-code-token-link: #032f62;
}
```

**Note:** Use `--astro-code-` prefix instead of `--shiki-`

## `<Code />` Component

Powered by Shiki for use in `.astro` and `.mdx` files.

### Basic Usage

```astro
---
import { Code } from 'astro:components';
---

<!-- JavaScript code -->
<Code code={`const foo = 'bar';`} lang="js" />

<!-- Python code -->
<Code code={`def hello():
    print("Hello, World!")`} lang="python" />
```

### Props

```typescript
interface Props {
  code: string;              // Code to highlight
  lang: string;              // Language
  theme?: string;            // Override theme
  wrap?: boolean;            // Enable word wrapping
  inline?: boolean;          // Inline code
  defaultColor?: boolean;    // Use default colors
  transformers?: any[];      // Shiki transformers
  meta?: string;             // Meta string for transformers
}
```

### Examples

**Custom theme:**

```astro
<Code
  code={`const foo = 'bar';`}
  lang="js"
  theme="dracula"
/>
```

**Word wrapping:**

```astro
<Code
  code={`const veryLongVariableName = 'This is a very long string that might need wrapping';`}
  lang="js"
  wrap
/>
```

**Inline code:**

```astro
<p>
  The function
  <Code code={`const foo = 'bar';`} lang="js" inline />
  demonstrates a variable declaration.
</p>
```

**Without default colors:**

```astro
<Code
  code={`const foo = 'bar';`}
  lang="js"
  defaultColor={false}
/>
```

### Transformers

**Added in:** Astro v4.11.0

Use Shiki transformers for advanced highlighting features:

```astro
---
import { Code } from 'astro:components';
import {
  transformerNotationFocus,
  transformerMetaHighlight
} from '@shikijs/transformers';

const code = `const foo = 'hello'
const bar = ' world'
console.log(foo + bar) // [!code focus]
`;
---

<Code
  code={code}
  lang="js"
  transformers={[transformerMetaHighlight()]}
  meta="{1,3}"
/>

<style is:global>
  pre.has-focused .line:not(.focused) {
    filter: blur(1px);
  }
</style>
```

**Common transformers:**

```bash
# Install transformers
pnpm add -D @shikijs/transformers
```

```astro
---
import {
  transformerNotationDiff,
  transformerNotationHighlight,
  transformerNotationFocus,
  transformerNotationErrorLevel,
} from '@shikijs/transformers';
---

<!-- Line highlighting -->
<Code
  code={code}
  lang="js"
  transformers={[transformerNotationHighlight()]}
  meta="{1,3-5}"
/>

<!-- Diff highlighting -->
<Code
  code={code}
  lang="js"
  transformers={[transformerNotationDiff()]}
/>
```

**Diff example:**

````markdown
```js
const foo = 'bar'; // [!code --]
const foo = 'baz'; // [!code ++]
```
````

## `<Prism />` Component

Alternative highlighter using Prism.

### Installation

```bash
# Install Prism component
pnpm add @astrojs/prism
```

### Basic Usage

```astro
---
import { Prism } from '@astrojs/prism';
---

<Prism lang="js" code={`const foo = 'bar';`} />
```

### Add Prism Stylesheet

Prism requires external CSS. Choose a theme from [Prism Themes](https://github.com/PrismJS/prism-themes).

**Method 1: Download theme to `public/`**

1. Download `prism-theme.css`
2. Place in `public/styles/prism.css`
3. Link in your layout:

```astro
---
// src/layouts/Layout.astro
---

<html>
  <head>
    <link rel="stylesheet" href="/styles/prism.css" />
  </head>
  <body>
    <slot />
  </body>
</html>
```

**Method 2: Use CDN**

```astro
<html>
  <head>
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css"
    />
  </head>
  <body>
    <slot />
  </body>
</html>
```

**Popular Prism themes:**
- `prism.css` - Default
- `prism-tomorrow.css` - Tomorrow theme
- `prism-twilight.css` - Twilight
- `prism-okaidia.css` - Okaidia
- `prism-coy.css` - Coy

### Switch to Prism for Markdown

Configure Astro to use Prism for all Markdown code blocks:

```javascript
// astro.config.mjs
export default defineConfig({
  markdown: {
    syntaxHighlight: 'prism',
  },
});
```

### Disable Syntax Highlighting

To disable all syntax highlighting:

```javascript
// astro.config.mjs
export default defineConfig({
  markdown: {
    syntaxHighlight: false,
  },
});
```

**Use case:** When using a different highlighting solution (e.g., Expressive Code)

## Community Integrations

### Expressive Code

More advanced code block features:

```bash
pnpm astro add astro-expressive-code
```

**Features:**
- Text markers and annotations
- Line numbers
- Diff highlighting
- Terminal frames
- Copy buttons
- Extensive customization

**Usage:**

````markdown
```js title="example.js" {1,3-5} ins={2} del={6}
const foo = 'bar';
const baz = 'qux';  // Added
const hello = 'world';
const lorem = 'ipsum';
const dolor = 'sit';
const old = 'removed'; // Removed
```
````

[See Expressive Code documentation](https://expressive-code.com/)

## Complete Examples

### Example 1: Blog Post with Code

```astro
---
// src/pages/blog/[slug].astro
import { getEntry, render } from 'astro:content';
import Layout from '../../layouts/Layout.astro';

const { slug } = Astro.params;
const post = await getEntry('blog', slug);
const { Content } = await render(post);
---

<Layout title={post.data.title}>
  <article>
    <h1>{post.data.title}</h1>
    <!-- Markdown with highlighted code blocks -->
    <Content />
  </article>
</Layout>
```

**Markdown content:**

````markdown
---
title: "Using Async/Await in JavaScript"
---

Here's how to use async/await:

```js
async function fetchData() {
  try {
    const response = await fetch('https://api.example.com/data');
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error:', error);
  }
}
```
````

### Example 2: Dynamic Code Examples

```astro
---
import { Code } from 'astro:components';

const examples = [
  {
    title: 'JavaScript',
    code: `const greeting = 'Hello, World!';
console.log(greeting);`,
    lang: 'js',
  },
  {
    title: 'Python',
    code: `greeting = 'Hello, World!'
print(greeting)`,
    lang: 'python',
  },
];
---

<div class="examples">
  {examples.map(example => (
    <div>
      <h3>{example.title}</h3>
      <Code code={example.code} lang={example.lang} />
    </div>
  ))}
</div>
```

### Example 3: API Reference with Code

```astro
---
import { Code } from 'astro:components';

const apiExample = `fetch('https://api.example.com/users', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer TOKEN',
    'Content-Type': 'application/json',
  }
})
.then(response => response.json())
.then(data => console.log(data));`;
---

<section>
  <h2>API Usage Example</h2>
  <Code code={apiExample} lang="js" theme="github-dark" />
</section>
```

## Best Practices for AI Agents

### 1. Use Shiki for Markdown

```javascript
// Good: Default Shiki for Markdown
export default defineConfig({
  markdown: {
    shikiConfig: {
      theme: 'github-dark',
    },
  },
});
```

### 2. Use `<Code />` for Dynamic Content

```astro
<!-- Good: Component for dynamic code -->
<Code code={dynamicCode} lang="js" />

<!-- Avoid: Markdown for dynamic content -->
```

### 3. Configure Dark Mode

```javascript
// Good: Support both modes
shikiConfig: {
  themes: {
    light: 'github-light',
    dark: 'github-dark',
  },
}
```

### 4. Always Specify Language

````markdown
<!-- Good: Language specified -->
```js
const foo = 'bar';
```

<!-- Bad: No language -->
```
const foo = 'bar';
```
````

### 5. Use Transformers for Advanced Features

```astro
<!-- Good: Line highlighting with transformers -->
<Code
  code={code}
  lang="js"
  transformers={[transformerNotationHighlight()]}
  meta="{1,3-5}"
/>
```

## Common Issues and Solutions

### Issue: Theme Not Applying

**Solution:** Check theme name is correct and supported by Shiki.

```javascript
// Check available themes
import { bundledThemes } from 'shiki';
console.log(Object.keys(bundledThemes));
```

### Issue: Dark Mode Not Working

**Solution:** Ensure CSS uses `.astro-code` class:

```css
/* Correct */
.astro-code { }

/* Incorrect */
.shiki { }
```

### Issue: Code Block Too Wide

**Solution:** Add overflow styling:

```css
.astro-code {
  overflow-x: auto;
  max-width: 100%;
}
```

## Quick Reference

```javascript
// astro.config.mjs - Basic theme
export default defineConfig({
  markdown: {
    shikiConfig: {
      theme: 'dracula',
    },
  },
});

// astro.config.mjs - Dual themes
export default defineConfig({
  markdown: {
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'github-dark',
      },
    },
  },
});
```

```astro
<!-- Code component -->
<Code code={`const foo = 'bar';`} lang="js" />

<!-- With theme -->
<Code code={code} lang="js" theme="dracula" />

<!-- Inline -->
<Code code={code} lang="js" inline />

<!-- With wrapping -->
<Code code={code} lang="js" wrap />
```

## Resources

- [Shiki Documentation](https://shiki.style/)
- [Shiki Themes](https://shiki.style/themes)
- [Shiki Transformers](https://shiki.style/packages/transformers)
- [Prism Themes](https://github.com/PrismJS/prism-themes)
- [Expressive Code](https://expressive-code.com/)
- [Astro Markdown Config Reference](https://docs.astro.build/en/reference/configuration-reference/#markdown-options)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent guide for syntax highlighting in Astro
**See Also**: 05-content-collections.md for rendering Markdown content
