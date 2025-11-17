# Markdoc Integration

## Overview

Markdoc is a powerful markup language that extends Markdown with custom tags and components. The `@astrojs/markdoc` integration allows you to use Markdoc files (`.mdoc`) in your Astro project with full support for content collections, Astro components, and advanced templating features.

## Why Markdoc?

### Markdoc vs Standard Markdown

**Standard Markdown:**
- Simple, limited syntax
- No custom components without MDX
- No variables or conditionals
- Basic content structure

**Markdoc:**
- Custom tags for components
- Variables and conditionals
- Functions for dynamic content
- Partials for reusable content
- Type-safe component props
- Better suited for complex documentation

### When to Use Markdoc

**Use Markdoc when:**
- Building documentation sites with complex layouts
- Need reusable content partials
- Want type-safe component attributes
- Migrating existing Markdoc content
- Need conditional content rendering

**Use MDX when:**
- Want JSX syntax familiarity
- Need inline JavaScript expressions
- Prefer React-like component syntax

**Use plain Markdown when:**
- Simple blog posts or articles
- No need for components
- Want maximum compatibility

## Installation

### Automatic Installation

Using the Astro CLI (recommended):

```bash
# npm
npx astro add markdoc

# pnpm
pnpm astro add markdoc

# Yarn
yarn astro add markdoc
```

This automatically:
1. Installs `@astrojs/markdoc`
2. Updates `astro.config.mjs`
3. Creates starter configuration

### Manual Installation

**1. Install Package:**

```bash
pnpm add @astrojs/markdoc
```

**2. Update Config:**

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import markdoc from '@astrojs/markdoc';

export default defineConfig({
  integrations: [markdoc()],
});
```

## VS Code Integration

### Setup Language Extension

**1. Install Extension:**

Search for "Markdoc Language Support" in VS Code extensions.

**2. Create Configuration:**

```json
// markdoc.config.json (project root)
[
  {
    "id": "my-site",
    "path": "src/content",
    "schema": {
      "path": "markdoc.config.mjs",
      "type": "esm",
      "property": "default",
      "watch": true
    }
  }
]
```

This provides:
- Syntax highlighting
- Autocomplete for configured tags
- Error detection
- Schema validation

## Basic Usage

### Content Collections with Markdoc

**1. Create Markdoc Files:**

```
src/
  content/
    docs/
      introduction.mdoc
      quick-start.mdoc
      advanced.mdoc
```

**2. Use `.mdoc` Extension:**

```markdoc
---
title: Getting Started
description: Learn the basics of our framework
---

# Getting Started

Welcome to our documentation! This guide will help you get up and running.

## Installation

Run the following command:

\`\`\`bash
npm install our-framework
\`\`\`
```

**3. Query and Render:**

```astro
---
// src/pages/docs/[...slug].astro
import { getEntry, render } from 'astro:content';

export async function getStaticPaths() {
  return [
    { params: { slug: 'introduction' } },
    { params: { slug: 'quick-start' } },
  ];
}

const { slug } = Astro.params;
const entry = await getEntry('docs', slug);
const { Content } = await render(entry);
---

<article>
  <h1>{entry.data.title}</h1>
  <p>{entry.data.description}</p>
  <Content />
</article>
```

## Variables

### Passing Variables to Content

**From Page Component:**

```astro
---
// src/pages/docs/[slug].astro
import { getEntry, render } from 'astro:content';

const entry = await getEntry('docs', Astro.params.slug);
const { Content } = await render(entry);

const user = {
  name: 'Alice',
  role: 'admin',
};
---

<Content user={user} environment="production" />
```

**In Markdoc File:**

```markdoc
<!-- src/content/docs/dashboard.mdoc -->

# Welcome, {% $user.name %}!

{% if $environment === "production" %}
You are in production mode.
{% /if %}

{% if $user.role === "admin" %}
## Admin Panel

Access your admin dashboard here.
{% /if %}
```

### Global Variables

Set variables available to all Markdoc files:

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  variables: {
    environment: process.env.NODE_ENV,
    version: '1.0.0',
    features: {
      darkMode: true,
      analytics: true,
    },
  },
});
```

**Usage:**

```markdoc
Current version: {% $version %}

{% if $features.darkMode %}
Dark mode is enabled!
{% /if %}
```

### Accessing Frontmatter

Pass frontmatter as a variable:

```astro
---
import { getEntry, render } from 'astro:content';

const entry = await getEntry('blog', 'post-1');
const { Content } = await render(entry);
---

<Content frontmatter={entry.data} />
```

**In Markdoc:**

```markdoc
---
title: My Blog Post
author: Alice
tags: ['astro', 'markdoc']
---

# {% $frontmatter.title %}

By {% $frontmatter.author %}

Tags: {% $frontmatter.tags.join(', ') %}
```

## Custom Tags (Components)

### Astro Components as Tags

**1. Create Component:**

```astro
---
// src/components/Aside.astro
interface Props {
  type?: 'note' | 'tip' | 'warning' | 'danger';
  title?: string;
}

const { type = 'note', title } = Astro.props;
---

<aside class={`aside aside-${type}`}>
  {title && <h4>{title}</h4>}
  <slot />
</aside>

<style>
  .aside {
    padding: 1rem;
    border-left: 4px solid;
    margin: 1.5rem 0;
  }

  .aside-note { border-color: blue; background: #e3f2fd; }
  .aside-tip { border-color: green; background: #e8f5e9; }
  .aside-warning { border-color: orange; background: #fff3e0; }
  .aside-danger { border-color: red; background: #ffebee; }
</style>
```

**2. Configure Tag:**

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig, component } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  tags: {
    aside: {
      render: component('./src/components/Aside.astro'),
      attributes: {
        type: { type: String },
        title: { type: String },
      },
    },
  },
});
```

**3. Use in Markdoc:**

```markdoc
# Documentation

{% aside type="tip" title="Pro Tip" %}
You can use variables inside custom tags!
{% /aside %}

{% aside type="warning" %}
This is a **warning** with Markdown content.
{% /aside %}
```

### Self-Closing Tags

For components without children:

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    callout: {
      render: component('./src/components/Callout.astro'),
      selfClosing: true,
      attributes: {
        text: { type: String, required: true },
      },
    },
  },
});
```

**Usage:**

```markdoc
{% callout text="Important information!" /%}
```

### Client-Side UI Components

Wrap framework components in Astro components:

**1. Create React Component:**

```tsx
// src/components/Counter.tsx
import { useState } from 'react';

export default function Counter({ initial = 0 }: { initial?: number }) {
  const [count, setCount] = useState(initial);

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </div>
  );
}
```

**2. Create Astro Wrapper:**

```astro
---
// src/components/ClientCounter.astro
import Counter from './Counter';
---

<Counter {...Astro.props} client:load />
```

**3. Configure Tag:**

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    counter: {
      render: component('./src/components/ClientCounter.astro'),
      attributes: {
        initial: { type: Number },
      },
    },
  },
});
```

**4. Use in Markdoc:**

```markdoc
Try this interactive counter:

{% counter initial={5} /%}
```

### Components from NPM Packages

Use components from installed packages:

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig, component } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  tags: {
    tabs: {
      render: component('@astrojs/starlight/components', 'Tabs'),
    },
    tabItem: {
      render: component('@astrojs/starlight/components', 'TabItem'),
      attributes: {
        label: { type: String },
      },
    },
  },
});
```

**Usage:**

```markdoc
{% tabs %}
{% tabItem label="npm" %}
\`\`\`bash
npm install package
\`\`\`
{% /tabItem %}

{% tabItem label="pnpm" %}
\`\`\`bash
pnpm add package
\`\`\`
{% /tabItem %}
{% /tabs %}
```

## Partials

### Creating Reusable Content

**1. Create Partial (with `_` prefix):**

```markdoc
<!-- src/content/docs/_footer.mdoc -->

---

## Need Help?

- [Documentation](https://docs.example.com)
- [Discord Community](https://discord.gg/example)
- [GitHub Issues](https://github.com/example/repo/issues)

---

_Last updated: {% $frontmatter.lastUpdated %}_
```

**2. Use Partial:**

```markdoc
<!-- src/content/docs/getting-started.mdoc -->
---
title: Getting Started
lastUpdated: 2025-11-17
---

# Getting Started

Your content here...

{% partial file="./_footer.mdoc" /%}
```

### Partial with Import Alias

```markdoc
{% partial file="@/content/docs/_footer.mdoc" /%}
```

### Organizing Partials

```
src/
  content/
    _partials/
      footer.mdoc
      toc.mdoc
      cta.mdoc
    docs/
      guide.mdoc
      api.mdoc
```

**Usage:**

```markdoc
{% partial file="../_partials/cta.mdoc" /%}
```

## Syntax Highlighting

### Using Shiki (Recommended)

**1. Configure Shiki:**

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig } from '@astrojs/markdoc/config';
import shiki from '@astrojs/markdoc/shiki';

export default defineMarkdocConfig({
  extends: [
    shiki({
      theme: 'dracula',
      wrap: true,
      langs: ['astro', 'typescript', 'bash'],
    }),
  ],
});
```

**2. Use in Markdoc:**

````markdoc
```typescript
interface User {
  name: string;
  email: string;
}

const user: User = {
  name: 'Alice',
  email: 'alice@example.com',
};
```
````

### Using Prism

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig } from '@astrojs/markdoc/config';
import prism from '@astrojs/markdoc/prism';

export default defineMarkdocConfig({
  extends: [prism()],
});
```

See [recipes/syntax-highlighting.md](./syntax-highlighting.md) for Prism stylesheet configuration.

## Custom Nodes

### Override Default Markdown Elements

**Blockquotes:**

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig, nodes, component } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  nodes: {
    blockquote: {
      ...nodes.blockquote,
      render: component('./src/components/Quote.astro'),
    },
  },
});
```

**Component:**

```astro
---
// src/components/Quote.astro
---
<blockquote class="fancy-quote">
  <slot />
</blockquote>

<style>
  .fancy-quote {
    border-left: 4px solid var(--accent);
    padding-left: 1.5rem;
    font-style: italic;
    color: var(--text-muted);
  }
</style>
```

### Custom Headings

**1. Configure Heading Node:**

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  nodes: {
    heading: {
      ...nodes.heading,
      render: component('./src/components/Heading.astro'),
    },
  },
});
```

**2. Create Component:**

```astro
---
// src/components/Heading.astro
interface Props {
  level: 1 | 2 | 3 | 4 | 5 | 6;
  id: string;
}

const { level, id } = Astro.props;
const Tag = `h${level}` as any;
---

<Tag id={id} class={`heading-${level}`}>
  <a href={`#${id}`} class="anchor-link">
    <slot />
  </a>
</Tag>

<style>
  .anchor-link {
    color: inherit;
    text-decoration: none;
  }

  .anchor-link:hover {
    text-decoration: underline;
  }
</style>
```

**Props received:**
- `level`: Number 1-6
- `id`: Generated from heading text (e.g., `"level-3-heading"`)

### Custom Images

**1. Override Image Node:**

```astro
---
// src/components/MarkdocImage.astro
import { Image } from 'astro:assets';

interface Props {
  src: ImageMetadata | string;
  alt: string;
}

const { src, alt } = Astro.props;
---

{typeof src === 'string' ? (
  <img src={src} alt={alt} />
) : (
  <Image src={src} alt={alt} />
)}
```

**2. Configure Node:**

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  nodes: {
    image: {
      ...nodes.image,
      render: component('./src/components/MarkdocImage.astro'),
    },
  },
});
```

**Usage:**

```markdoc
<!-- Local images optimized with <Image /> -->
![Cat photo](./cat.jpg)

<!-- Remote images use <img> -->
![Dog photo](https://example.com/dog.jpg)
```

### Custom Image Tag

For more control (width, height, captions):

**1. Create Component:**

```astro
---
// src/components/Figure.astro
import { Image } from 'astro:assets';

interface Props {
  src: ImageMetadata | string;
  alt: string;
  width: number;
  height: number;
  caption?: string;
}

const { src, alt, width, height, caption } = Astro.props;
---

<figure>
  <Image {src} {alt} {width} {height} />
  {caption && <figcaption>{caption}</figcaption>}
</figure>
```

**2. Configure Tag:**

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    figure: {
      render: component('./src/components/Figure.astro'),
      attributes: {
        src: { type: String, required: true },
        alt: { type: String, required: true },
        width: { type: Number, required: true },
        height: { type: Number, required: true },
        caption: { type: String },
      },
    },
  },
});
```

**Usage:**

```markdoc
{% figure
  src="./hero.png"
  alt="Hero image"
  width={800}
  height={600}
  caption="Our amazing product"
/%}
```

## Advanced Configuration

### Functions

Create custom functions for dynamic content:

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  functions: {
    getCountryEmoji: {
      transform(parameters) {
        const [country] = Object.values(parameters);
        const countryToEmojiMap = {
          japan: '🇯🇵',
          spain: '🇪🇸',
          france: '🇫🇷',
          usa: '🇺🇸',
        };
        return countryToEmojiMap[country] ?? '🏳';
      },
    },
    uppercase: {
      transform(parameters) {
        const [text] = Object.values(parameters);
        return text.toUpperCase();
      },
    },
  },
});
```

**Usage:**

```markdoc
Hello from {% getCountryEmoji("japan") %}!

# {% uppercase("important notice") %}
```

### Validation

Add validation to tag attributes:

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    button: {
      render: component('./src/components/Button.astro'),
      attributes: {
        variant: {
          type: String,
          default: 'primary',
          matches: ['primary', 'secondary', 'danger'],
          errorLevel: 'critical',
        },
        size: {
          type: String,
          default: 'medium',
          matches: ['small', 'medium', 'large'],
        },
      },
    },
  },
});
```

**This enforces:**
- Only valid values accepted
- Default values applied
- Build-time errors for invalid values

### Change Root Element

By default, Markdoc wraps content in `<article>`. Change or remove it:

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  nodes: {
    document: {
      ...nodes.document,
      render: null, // Remove wrapper
      // or
      render: 'section', // Use <section> instead
    },
  },
});
```

## Integration Options

### allowHTML

Enable HTML markup in Markdoc:

```javascript
// astro.config.mjs
export default defineConfig({
  integrations: [
    markdoc({
      allowHTML: true,
    }),
  ],
});
```

**Enables:**

```markdoc
# My Document

<div class="custom-wrapper">
  Regular Markdoc content with **bold** text.

  {% aside type="tip" %}
  Mixed HTML and Markdoc!
  {% /aside %}
</div>
```

⚠️ **Security Warning:** Only enable if you trust the content source. HTML can include `<script>` tags (XSS risk).

### ignoreIndentation

Allow arbitrary indentation for better readability:

```javascript
// astro.config.mjs
export default defineConfig({
  integrations: [
    markdoc({
      ignoreIndentation: true,
    }),
  ],
});
```

**Enables:**

```markdoc
{% tabs %}
  {% tabItem label="Step 1" %}
    {% aside type="tip" %}
      This is deeply nested but readable!
    {% /aside %}
  {% /tabItem %}
{% /tabs %}
```

## TypeScript Support

### Type-Safe Configuration

```typescript
// markdoc.config.mts
import { defineMarkdocConfig, component } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  tags: {
    aside: {
      render: component('./src/components/Aside.astro'),
      attributes: {
        // Type-checked attributes
        type: {
          type: String,
          default: 'note',
          matches: ['note', 'tip', 'warning', 'danger'],
        } as const,
      },
    },
  },
});
```

### Component Props Types

```astro
---
// src/components/Aside.astro
interface Props {
  type: 'note' | 'tip' | 'warning' | 'danger';
  title?: string;
}

const { type, title } = Astro.props;
---
```

## Best Practices

### 1. Organize Partials

```
src/
  content/
    _partials/
      alerts/
        info.mdoc
        warning.mdoc
      navigation/
        footer.mdoc
        sidebar.mdoc
    docs/
      ...
```

### 2. Use TypeScript for Config

```typescript
// markdoc.config.mts (note the .mts extension)
import { defineMarkdocConfig } from '@astrojs/markdoc/config';
```

### 3. Validate Attribute Types

```javascript
tags: {
  button: {
    attributes: {
      variant: {
        type: String,
        matches: ['primary', 'secondary'],
        errorLevel: 'critical',
      },
    },
  },
},
```

### 4. Use Semantic Component Names

```javascript
// Good
tags: {
  callout: { ... },
  codeBlock: { ... },
  inlineCode: { ... },
}

// Avoid
tags: {
  c1: { ... },
  comp: { ... },
  thing: { ... },
}
```

### 5. Leverage Frontmatter

```markdoc
---
title: API Reference
version: 2.0
experimental: true
---

# {% $frontmatter.title %}

{% if $frontmatter.experimental %}
⚠️ This feature is experimental.
{% /if %}
```

## Common Patterns

### Documentation Site

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    tabs: { render: component('./src/components/Tabs.astro') },
    tab: { render: component('./src/components/Tab.astro') },
    apiReference: { render: component('./src/components/APIReference.astro') },
    codeGroup: { render: component('./src/components/CodeGroup.astro') },
  },
  nodes: {
    heading: {
      ...nodes.heading,
      render: component('./src/components/DocsHeading.astro'),
    },
  },
});
```

### Blog with Callouts

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  tags: {
    callout: {
      render: component('./src/components/Callout.astro'),
      attributes: {
        type: {
          type: String,
          matches: ['info', 'success', 'warning', 'error'],
        },
      },
    },
    tweet: {
      render: component('./src/components/TweetEmbed.astro'),
      attributes: {
        id: { type: String, required: true },
      },
    },
  },
});
```

### Multi-Language Support

```javascript
// markdoc.config.mjs
export default defineMarkdocConfig({
  variables: {
    locale: 'en',
  },
  functions: {
    t: {
      transform(parameters) {
        const [key] = Object.values(parameters);
        const translations = {
          en: { greeting: 'Hello' },
          es: { greeting: 'Hola' },
          fr: { greeting: 'Bonjour' },
        };
        return translations[this.variables.locale]?.[key] ?? key;
      },
    },
  },
});
```

**Usage:**

```markdoc
{% t("greeting") %}, World!
```

## Troubleshooting

### Issue: Components Not Rendering

**Check:**
1. Component path is correct in `markdoc.config.mjs`
2. Tag is properly configured with `render` property
3. Attributes match component props

### Issue: VS Code Extension Not Working

**Solution:**
1. Ensure `markdoc.config.json` exists in project root
2. Restart VS Code
3. Check extension is enabled

### Issue: HTML Not Rendering

**Solution:**
Enable `allowHTML` option:

```javascript
// astro.config.mjs
markdoc({ allowHTML: true })
```

### Issue: Indentation Breaking Layout

**Solution:**
Enable `ignoreIndentation` option:

```javascript
// astro.config.mjs
markdoc({ ignoreIndentation: true })
```

## Quick Reference

### Create Markdoc File

```markdoc
---
title: My Doc
---

# Heading

Regular **Markdown** content.

{% customTag attribute="value" %}
Content here
{% /customTag %}
```

### Configure Tag

```javascript
// markdoc.config.mjs
import { defineMarkdocConfig, component } from '@astrojs/markdoc/config';

export default defineMarkdocConfig({
  tags: {
    customTag: {
      render: component('./src/components/Custom.astro'),
      attributes: {
        attribute: { type: String },
      },
    },
  },
});
```

### Render in Page

```astro
---
import { getEntry, render } from 'astro:content';

const entry = await getEntry('docs', 'my-doc');
const { Content } = await render(entry);
---

<Content />
```

### Use Variables

```astro
<Content myVar="value" />
```

```markdoc
{% if $myVar === "value" %}
Content here
{% /if %}
```

### Use Partial

```markdoc
{% partial file="./_footer.mdoc" /%}
```

## See Also

- [05-content-collections.md](../05-content-collections.md) - Content Collections overview
- [09-markdown.md](../09-markdown.md) - Standard Markdown in Astro
- [recipes/syntax-highlighting.md](./syntax-highlighting.md) - Code highlighting setup
- [Official Markdoc Docs](https://markdoc.dev/) - Complete Markdoc reference
- [Astro Markdoc Integration](https://docs.astro.build/en/guides/integrations-guide/markdoc/) - Official Astro docs

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to using Markdoc in Astro projects
**Key Concepts**: Markdoc tags, custom components, variables, partials, content collections, type safety
