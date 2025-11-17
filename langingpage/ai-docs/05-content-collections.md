# Content Collections

**Added in:** Astro v2.0 | **Content Layer API:** Astro v5.0

Content collections are the best way to manage sets of content in any Astro project. Collections help to organize and query your documents, enable Intellisense and type checking in your editor, and provide automatic TypeScript type-safety for all of your content.

## Overview

Content collections provide:

- **Type Safety** - Zod validation and automatic TypeScript types
- **Organization** - Structured content with predictable shapes
- **Performance** - Caching and scalability for thousands of entries
- **APIs** - Built-in helpers like `getCollection()` and `render()`
- **Flexibility** - Local files or remote data (CMS, databases, APIs)

## What are Content Collections?

A collection is a set of data that is structurally similar. This can be:

- Directory of blog posts (Markdown/MDX)
- JSON file of product items
- Remote data from CMS, database, or API
- Any data representing multiple items of the same shape

### Local Collections Example

```
src/
├── newsletter/          # "newsletter" collection
│   ├── week-1.md       # collection entry
│   ├── week-2.md       # collection entry
│   └── week-3.md       # collection entry
└── authors/            # "authors" collection
    └── authors.json    # single file with all entries
```

### Remote Collections

With an appropriate loader, fetch data from:
- Content Management Systems (CMS)
- Databases
- APIs
- Headless payment systems
- Any external data source

## TypeScript Configuration

Content collections require specific TypeScript settings for Zod validation and type checking.

**Recommended:** Use Astro's strict settings:

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/strict"  // or "strictest"
}
```

**If using `base` template:**

```json
// tsconfig.json
{
  "extends": "astro/tsconfigs/base",
  "compilerOptions": {
    "strictNullChecks": true,  // Required for collections
    "allowJs": true             // Required (included in all templates)
  }
}
```

## Defining Collections

### Collection Config File

Create `src/content.config.ts` to configure your collections:

```typescript
// src/content.config.ts

// 1. Import utilities from `astro:content`
import { defineCollection, z } from 'astro:content';

// 2. Import loader(s)
import { glob, file } from 'astro/loaders';

// 3. Define your collection(s)
const blog = defineCollection({ /* ... */ });
const dogs = defineCollection({ /* ... */ });

// 4. Export a single `collections` object
export const collections = { blog, dogs };
```

**Supported extensions:** `.ts`, `.js`, `.mjs`

### Collection Structure

Each collection needs:

1. **Loader** (required) - Defines the data source
2. **Schema** (optional, highly recommended) - Validates data shape

```typescript
const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    pubDate: z.coerce.date(),
  }),
});
```

## Built-in Loaders

Astro provides two built-in loaders for local content:

### 1. `glob()` Loader

Creates entries from directories of files. One file = one entry.

**Supports:** Markdown, MDX, Markdoc, JSON, YAML, TOML

```typescript
import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({
    pattern: "**/*.md",           // Match all .md files
    base: "./src/data/blog"       // Base directory
  }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
  }),
});
```

**Multiple patterns:**

```typescript
const probes = defineCollection({
  // Load all .md files EXCEPT those starting with "voyager-"
  loader: glob({
    pattern: ['*.md', '!voyager-*'],
    base: 'src/data/space-probes'
  }),
  schema: z.object({
    name: z.string(),
    type: z.enum(['Space Probe', 'Mars Rover', 'Comet Lander']),
    launch_date: z.date(),
    status: z.enum(['Active', 'Inactive', 'Decommissioned']),
  }),
});
```

**Glob patterns use micromatch syntax:**

```typescript
// Common patterns
"**/*.md"        // All .md files recursively
"*.md"           // .md files in base directory only
"blog/*.md"      // .md files in blog/ directory
"!draft-*"       // Exclude files starting with "draft-"
["*.md", "*.mdx"] // Multiple file types
```

### 2. `file()` Loader

Creates multiple entries from a single file. Each entry must have an `id` property.

**Supports:** JSON, YAML, TOML (auto-parsed)

```typescript
import { file } from 'astro/loaders';

const dogs = defineCollection({
  loader: file("src/data/dogs.json"),
  schema: z.object({
    id: z.string(),      // Required!
    breed: z.string(),
    temperament: z.array(z.string()),
  }),
});
```

**Example data file:**

```json
// src/data/dogs.json
[
  {
    "id": "poodle",
    "breed": "Poodle",
    "temperament": ["Intelligent", "Active"]
  },
  {
    "id": "labrador",
    "breed": "Labrador Retriever",
    "temperament": ["Friendly", "Outgoing"]
  }
]
```

### Parser Function

Use custom parsers for unsupported formats or nested data:

**CSV Example:**

```typescript
import { file } from 'astro/loaders';
import { parse as parseCsv } from 'csv-parse/sync';

const cats = defineCollection({
  loader: file("src/data/cats.csv", {
    parser: (text) => parseCsv(text, {
      columns: true,
      skipEmptyLines: true
    })
  })
});
```

**Nested JSON Example:**

```json
// src/data/pets.json
{
  "dogs": [{...}, {...}],
  "cats": [{...}, {...}]
}
```

```typescript
const dogs = defineCollection({
  loader: file("src/data/pets.json", {
    parser: (text) => JSON.parse(text).dogs
  })
});

const cats = defineCollection({
  loader: file("src/data/pets.json", {
    parser: (text) => JSON.parse(text).cats
  })
});
```

## Custom Loaders

### Inline Loaders

Simple async function that returns entries:

```typescript
const countries = defineCollection({
  loader: async () => {
    const response = await fetch("https://restcountries.com/v3.1/all");
    const data = await response.json();

    // Must return array with id property
    return data.map((country) => ({
      id: country.cca3,
      ...country,
    }));
  },
  schema: z.object({
    name: z.object({
      common: z.string(),
    }),
    population: z.number(),
  }),
});
```

**When inline loader runs:**
- Clears the store
- Reloads all entries
- Stores in collection

**Use for:** Simple data fetching without manual control

### Loader Objects

For advanced control, use the Content Loader API to create loader objects.

**Features:**
- Incremental updates
- Manual store control
- Conditional clearing
- NPM distribution

See [Content Loader API documentation](https://docs.astro.build/en/reference/loader-reference/) for building custom loaders.

**Community loaders:** Available in the [Astro integrations directory](https://astro.build/integrations)

## Schemas

Schemas enforce consistent data structure using Zod validation.

### Basic Schema

```typescript
import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    tags: z.array(z.string()),
  })
});
```

### Zod Data Types Cheatsheet

```typescript
import { z, defineCollection } from 'astro:content';

defineCollection({
  schema: z.object({
    // Primitives
    isDraft: z.boolean(),
    title: z.string(),
    sortOrder: z.number(),

    // Objects
    image: z.object({
      src: z.string(),
      alt: z.string(),
    }),

    // Defaults
    author: z.string().default('Anonymous'),

    // Enums
    language: z.enum(['en', 'es']),

    // Arrays
    tags: z.array(z.string()),

    // Optional
    footnote: z.string().optional(),

    // Dates
    // YAML dates (without quotes) are Date objects
    publishDate: z.date(),

    // Transform string to Date
    updatedDate: z.string().transform((str) => new Date(str)),

    // Validation
    authorContact: z.string().email(),
    canonicalURL: z.string().url(),
  })
})
```

**See:** [Zod documentation](https://zod.dev/) for complete reference

### Schema Benefits

1. **Type Safety** - Automatic TypeScript interfaces
2. **Validation** - Build-time error checking
3. **IntelliSense** - IDE autocomplete for data properties
4. **Documentation** - Self-documenting data structure

**Restart dev server or sync** (`s` + `enter` in dev mode) after schema changes.

## Collection References

Reference other collection entries using `reference()`:

```typescript
import { defineCollection, reference, z } from 'astro:content';
import { glob } from 'astro/loaders';

const blog = defineCollection({
  loader: glob({ pattern: '**/[^_]*.md', base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    // Reference single author
    author: reference('authors'),
    // Reference array of related posts
    relatedPosts: z.array(reference('blog')),
  })
});

const authors = defineCollection({
  loader: glob({ pattern: '**/[^_]*.json', base: "./src/data/authors" }),
  schema: z.object({
    name: z.string(),
    portfolio: z.string().url(),
  })
});

export const collections = { blog, authors };
```

**Using references in content:**

```markdown
---
title: "Welcome to my blog"
author: ben-holmes              # references authors/ben-holmes.json
relatedPosts:
- about-me                      # references blog/about-me.md
- my-year-in-review             # references blog/my-year-in-review.md
---
```

**Querying referenced data:**

```astro
---
import { getEntry, getEntries } from 'astro:content';

const blogPost = await getEntry('blog', 'welcome');

// Resolve single reference
const author = await getEntry(blogPost.data.author);

// Resolve array of references
const relatedPosts = await getEntries(blogPost.data.relatedPosts);
---

<h1>{blogPost.data.title}</h1>
<p>Author: {author.data.name}</p>

<h2>Related Posts:</h2>
{relatedPosts.map(post => (
  <a href={`/blog/${post.id}`}>{post.data.title}</a>
))}
```

## Custom Entry IDs

By default, entry `id` is generated from filename in URL-friendly format.

**Override with custom slug:**

```markdown
---
title: My Blog Post
slug: my-custom-id/supports/slashes
---
```

```json
{
  "title": "My Category",
  "slug": "my-custom-id/supports/slashes",
  "description": "Your category description here."
}
```

**Use cases:**
- Custom URL structure
- Migration from other systems
- Human-readable slugs

## Querying Collections

Astro provides two main query functions:

### `getCollection()`

Fetches entire collection, returns array of entries.

```typescript
import { getCollection } from 'astro:content';

// Get all entries
const allBlogPosts = await getCollection('blog');
```

**Returns:** Array of `CollectionEntry` objects with:
- `id` - Unique identifier
- `data` - Validated data object (frontmatter/properties)
- `body` - Raw uncompiled body (Markdown/MDX/Markdoc only)

### `getEntry()`

Fetches single entry by ID.

```typescript
import { getEntry } from 'astro:content';

// Get single entry
const poodleData = await getEntry('dogs', 'poodle');
```

### Sorting Collections

**Important:** Collection order is non-deterministic. Always sort manually:

```astro
---
import { getCollection } from 'astro:content';

// Sort by date (newest first)
const posts = (await getCollection('blog')).sort(
  (a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf()
);
---
```

**Common sorts:**

```typescript
// Alphabetical by title
posts.sort((a, b) => a.data.title.localeCompare(b.data.title));

// By custom order field
posts.sort((a, b) => a.data.sortOrder - b.data.sortOrder);

// By date (oldest first)
posts.sort((a, b) => a.data.pubDate.valueOf() - b.data.pubDate.valueOf());
```

### Filtering Collections

Use filter callback to query specific entries:

```typescript
import { getCollection } from 'astro:content';

// Filter by data property
const publishedPosts = await getCollection('blog', ({ data }) => {
  return data.draft !== true;
});

// Filter by environment
const posts = await getCollection('blog', ({ data }) => {
  return import.meta.env.PROD ? data.draft !== true : true;
});

// Filter by nested directory
const englishDocs = await getCollection('docs', ({ id }) => {
  return id.startsWith('en/');
});
```

**Filter use cases:**
- Draft posts (dev only)
- Published/unpublished content
- Language-specific content
- Category filtering
- Date ranges

## Using Collections in Templates

### Display List of Posts

```astro
---
// src/pages/blog/index.astro
import { getCollection } from 'astro:content';

const posts = (await getCollection('blog'))
  .sort((a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf());
---

<h1>My Blog</h1>
<ul>
  {posts.map(post => (
    <li>
      <a href={`/blog/${post.id}`}>
        {post.data.title}
      </a>
      <time>{post.data.pubDate.toDateString()}</time>
    </li>
  ))}
</ul>
```

### Render Entry Content

Use `render()` to convert Markdown/MDX to HTML:

```astro
---
// src/pages/blog/[id].astro
import { getEntry, render } from 'astro:content';

const { id } = Astro.params;
const entry = await getEntry('blog', id);

if (!entry) {
  return Astro.redirect('/404');
}

const { Content, headings } = await render(entry);
---

<article>
  <h1>{entry.data.title}</h1>
  <p>Published: {entry.data.pubDate.toDateString()}</p>

  <!-- Rendered HTML content -->
  <Content />

  <!-- Table of contents from headings -->
  <aside>
    <h2>Table of Contents</h2>
    <ul>
      {headings.map(heading => (
        <li><a href={`#${heading.slug}`}>{heading.text}</a></li>
      ))}
    </ul>
  </aside>
</article>
```

**`render()` returns:**
- `Content` - Component with rendered HTML
- `headings` - Array of heading objects

### Passing as Props

Type-safe prop passing using `CollectionEntry`:

```astro
---
// src/components/BlogCard.astro
import type { CollectionEntry } from 'astro:content';

interface Props {
  post: CollectionEntry<'blog'>;
}

const { post } = Astro.props;
---

<article>
  <h2>{post.data.title}</h2>
  <p>{post.data.description}</p>
  <a href={`/blog/${post.id}`}>Read more</a>
</article>
```

**Usage:**

```astro
---
import BlogCard from '../components/BlogCard.astro';
import { getCollection } from 'astro:content';

const posts = await getCollection('blog');
---

{posts.map(post => <BlogCard post={post} />)}
```

## Generating Routes

Collections don't auto-generate routes. Create dynamic routes manually.

### Static Routes (SSG - Default)

Use `getStaticPaths()` to generate pages at build time:

```astro
---
// src/pages/blog/[id].astro
import { getCollection, render } from 'astro:content';

// 1. Generate path for every entry
export async function getStaticPaths() {
  const posts = await getCollection('blog');
  return posts.map(post => ({
    params: { id: post.id },
    props: { post },
  }));
}

// 2. Get entry from props
const { post } = Astro.props;
const { Content } = await render(post);
---

<h1>{post.data.title}</h1>
<Content />
```

**Generates:** One HTML page per entry at build time

**Slug with slashes:**

If using custom slugs with `/` characters:

```astro
// src/pages/blog/[...slug].astro
export async function getStaticPaths() {
  const posts = await getCollection('blog');
  return posts.map(post => ({
    params: { slug: post.id },  // Can contain slashes
    props: { post },
  }));
}
```

### Server Routes (SSR)

For on-demand rendering, query on each request:

```astro
---
// src/pages/blog/[id].astro
import { getEntry, render } from "astro:content";

// 1. Get slug from request
const { id } = Astro.params;

if (id === undefined) {
  return Astro.redirect("/404");
}

// 2. Query for entry
const post = await getEntry("blog", id);

// 3. Redirect if not found
if (post === undefined) {
  return Astro.redirect("/404");
}

// 4. Render entry
const { Content } = await render(post);
---

<h1>{post.data.title}</h1>
<Content />
```

**Use cases:**
- Dynamic content that changes frequently
- Personalized content
- Large sites (faster builds)

## JSON Schemas

**Added in:** Astro v4.13.0

Astro auto-generates JSON Schema files for editor support.

### Generated Schemas

Located in `.astro/collections/`:

```
.astro/
└── collections/
    ├── authors.schema.json
    └── posts.schema.json
```

### Using in JSON Files

Reference schema directly:

```json
// src/data/authors/armand.json
{
  "$schema": "../../../.astro/collections/authors.schema.json",
  "name": "Armand",
  "skills": ["Astro", "Starlight"]
}
```

**Benefits:**
- IntelliSense in editor
- Type checking
- Validation errors

### VS Code JSON Configuration

Apply schema to all files in directory:

```json
// .vscode/settings.json
{
  "json.schemas": [
    {
      "fileMatch": ["/src/data/authors/**"],
      "url": "./.astro/collections/authors.schema.json"
    }
  ]
}
```

### YAML Support (VS Code)

Install: [Red Hat YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

**Single file:**

```yaml
# yaml-language-server: $schema=../../../.astro/collections/authors.schema.json
name: Armand
skills:
  - Astro
  - Starlight
```

**All files in directory:**

```json
// .vscode/settings.json
{
  "yaml.schemas": {
    "./.astro/collections/authors.schema.json": [
      "/src/content/authors/*.yml"
    ]
  }
}
```

## When to Use Collections

### Use Collections When:

✅ **Multiple related files** with shared structure
- Blog posts with same frontmatter
- Product listings with same properties
- Team member profiles

✅ **Remote content** from CMS/database
- Take advantage of collection APIs
- Use built-in or custom loaders

✅ **Large-scale content** (thousands of entries)
- Performance and caching at scale
- Type-safe querying

✅ **Type safety is important**
- Zod validation
- TypeScript autocomplete
- Build-time error checking

### Don't Use Collections When:

❌ **Single or few unique pages**
- Use page components directly (`src/pages/about.astro`)

❌ **Static files not processed by Astro**
- PDFs, downloads → use `public/` directory

❌ **SDK with incompatible data source**
- Prefer native SDK if it doesn't offer a loader

❌ **Real-time data requirements**
- Collections update at build time only
- Use fetch() for live data with SSR

## Best Practices for AI Agents

### 1. Always Define Schemas

```typescript
// Good: Type-safe with validation
const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
  }),
});

// Avoid: No validation, no type safety
const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
});
```

### 2. Sort Collections Explicitly

```typescript
// Good: Explicit sorting
const posts = (await getCollection('blog')).sort(
  (a, b) => b.data.pubDate.valueOf() - a.data.pubDate.valueOf()
);

// Bad: Relying on default order (non-deterministic)
const posts = await getCollection('blog');
```

### 3. Filter Draft Content Appropriately

```typescript
// Good: Hide drafts in production
const posts = await getCollection('blog', ({ data }) => {
  return import.meta.env.PROD ? data.draft !== true : true;
});
```

### 4. Use References for Related Content

```typescript
// Good: Type-safe references
schema: z.object({
  author: reference('authors'),
  relatedPosts: z.array(reference('blog')),
})

// Avoid: String IDs (no validation)
schema: z.object({
  authorId: z.string(),
  relatedPostIds: z.array(z.string()),
})
```

### 5. Handle Missing Entries

```astro
---
const post = await getEntry('blog', id);

// Good: Explicit error handling
if (!post) {
  return Astro.redirect('/404');
}
---
```

### 6. Choose Right Loader

```typescript
// Use glob() for: One file = one entry
loader: glob({ pattern: "**/*.md", base: "./src/data/blog" })

// Use file() for: Multiple entries in one file
loader: file("src/data/products.json")

// Use inline loader for: Simple remote data
loader: async () => {
  const res = await fetch('https://api.example.com/posts');
  return res.json();
}
```

## Common Patterns

### Pattern 1: Blog with Authors

```typescript
// src/content.config.ts
const authors = defineCollection({
  loader: glob({ pattern: '*.json', base: './src/data/authors' }),
  schema: z.object({
    name: z.string(),
    bio: z.string(),
    avatar: z.string(),
  }),
});

const blog = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/data/blog' }),
  schema: z.object({
    title: z.string(),
    pubDate: z.date(),
    author: reference('authors'),
  }),
});

export const collections = { authors, blog };
```

### Pattern 2: Multi-Language Content

```typescript
const docs = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/data/docs' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
  }),
});

// Query by language
const englishDocs = await getCollection('docs', ({ id }) => {
  return id.startsWith('en/');
});

const spanishDocs = await getCollection('docs', ({ id }) => {
  return id.startsWith('es/');
});
```

### Pattern 3: Products from CMS

```typescript
const products = defineCollection({
  loader: async () => {
    const response = await fetch('https://cms.example.com/api/products');
    const data = await response.json();
    return data.products.map(product => ({
      id: product.sku,
      ...product,
    }));
  },
  schema: z.object({
    name: z.string(),
    price: z.number(),
    inStock: z.boolean(),
  }),
});
```

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro Content Collections
**Next Module**: Component Development (06-components.md)
