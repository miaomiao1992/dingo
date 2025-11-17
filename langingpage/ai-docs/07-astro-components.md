# Astro Components

Astro components are the basic building blocks of any Astro project. They are HTML-only templating components with no client-side runtime and use the `.astro` file extension.

## Overview

**Key characteristics:**
- HTML-only templating (no client-side runtime)
- Server-side rendering (build-time or on-demand)
- Zero JavaScript footprint by default
- Can include any valid HTML
- Support JavaScript in component script
- TypeScript support built-in

**If you know HTML, you already know enough to write your first Astro component.**

## What Can Astro Components Be?

Astro components are extremely flexible:

| Use Case | Example | Location |
|----------|---------|----------|
| **Snippet** | Collection of `<meta>` tags | `src/components/SEO.astro` |
| **UI Element** | Header, button, card | `src/components/Header.astro` |
| **Layout** | Page wrapper with header/footer | `src/layouts/BaseLayout.astro` |
| **Page** | Entire page | `src/pages/about.astro` |

## Server-Side Rendering

**The most important thing to know:** Astro components render on the server, not the client.

```astro
---
// This JavaScript runs on the SERVER only
const expensiveData = await fetchFromDatabase();
const apiKey = import.meta.env.SECRET_API_KEY; // Safe!
---

<!-- This HTML is sent to the browser -->
<div>{expensiveData.title}</div>
```

**Benefits:**
- Fast page loads (HTML only)
- No JavaScript sent to client by default
- Safe to use API keys and sensitive data
- SEO-friendly

**For client-side interactivity:**
- Add `<script>` tags for vanilla JS
- Use framework components with `client:*` directives (see 06-framework-components.md)
- Use server islands with `server:defer` for dynamic content

## Component Structure

Every Astro component has two main parts:

```astro
---
// Component Script (JavaScript/TypeScript)
// Runs on the server only
---

<!-- Component Template (HTML + JS Expressions) -->
<!-- Sent to the browser as HTML -->
```

### 1. Component Script (Frontmatter)

The code fence (`---`) defines the component script:

```astro
---
// Import other components
import Header from './Header.astro';
import Button from './Button.jsx'; // Framework component

// Import data
import data from '../data/products.json';

// Access props
const { title, description } = Astro.props;

// Fetch data (async allowed!)
const posts = await fetch('https://api.example.com/posts')
  .then(r => r.json());

// Define variables
const currentYear = new Date().getFullYear();
---
```

**What you can do:**
- Import Astro components
- Import framework components (React, Vue, etc.)
- Import data files (JSON, etc.)
- Fetch from APIs or databases
- Access component props via `Astro.props`
- Create variables for the template
- Run any JavaScript code

**Important:** All code here runs on the **server only** and is "fenced in" - it never reaches the client.

### 2. Component Template

The template determines HTML output:

```astro
---
const { title } = Astro.props;
const items = [1, 2, 3];
---

<!-- HTML comments supported -->
{/* JSX-style comments also work */}

<!-- Plain HTML -->
<h1>Hello, World!</h1>

<!-- Use variables from script -->
<p>{title}</p>

<!-- JavaScript expressions -->
<ul>
  {items.map(item => <li>{item}</li>)}
</ul>

<!-- Import and use other components -->
<Header />

<!-- Framework components with hydration -->
<ReactComponent client:load />
```

**Template features:**
- Plain HTML
- JavaScript expressions in `{}`
- Component imports
- Astro directives (e.g., `class:list`)
- `<style>` and `<script>` tags

## Component Props

Props allow components to receive data from parents.

### Basic Props

```astro
---
// src/components/Greeting.astro
const { greeting, name } = Astro.props;
---

<h2>{greeting}, {name}!</h2>
```

**Usage:**

```astro
---
import Greeting from './Greeting.astro';
---

<Greeting greeting="Hello" name="World" />
<Greeting greeting="Hi" name="Astro" />
```

### TypeScript Props

Define props with a TypeScript interface:

```astro
---
// src/components/Greeting.astro
interface Props {
  greeting?: string;  // Optional with ?
  name: string;       // Required
}

const { greeting = "Hello", name } = Astro.props;
---

<h2>{greeting}, {name}!</h2>
```

**Benefits:**
- Type checking
- Editor autocomplete
- Automatic error detection
- Documentation

### Default Values

Provide fallback values for props:

```astro
---
const {
  greeting = "Hello",
  name = "Astronaut",
  showImage = false,
} = Astro.props;
---

<h2>{greeting}, {name}!</h2>
{showImage && <img src="/avatar.png" alt={name} />}
```

### Complex Props

Props can be any JavaScript type:

```astro
---
interface Props {
  // Primitives
  title: string;
  count: number;
  isActive: boolean;

  // Objects
  user: {
    name: string;
    email: string;
  };

  // Arrays
  tags: string[];
  items: Array<{ id: number; name: string }>;

  // Functions (for server-only components)
  onClick?: () => void;

  // Optional props
  description?: string;
}

const { title, user, tags, items } = Astro.props;
---

<h1>{title}</h1>
<p>Author: {user.name}</p>
<ul>
  {tags.map(tag => <li>{tag}</li>)}
</ul>
```

**Usage:**

```astro
<BlogPost
  title="My Post"
  user={{ name: "Alice", email: "alice@example.com" }}
  tags={["astro", "typescript"]}
  items={[
    { id: 1, name: "Item 1" },
    { id: 2, name: "Item 2" },
  ]}
/>
```

## Slots

Slots allow components to accept child content.

### Default Slot

The `<slot />` element renders child content:

```astro
---
// src/components/Card.astro
const { title } = Astro.props;
---

<div class="card">
  <h2>{title}</h2>
  <slot />  <!-- Children go here -->
</div>
```

**Usage:**

```astro
---
import Card from './Card.astro';
---

<Card title="My Card">
  <p>This content goes in the slot!</p>
  <button>Click me</button>
</Card>
```

**Rendered output:**

```html
<div class="card">
  <h2>My Card</h2>
  <p>This content goes in the slot!</p>
  <button>Click me</button>
</div>
```

### Named Slots

Multiple slots with specific names:

```astro
---
// src/components/Layout.astro
const { title } = Astro.props;
---

<div class="layout">
  <header>
    <slot name="header" />
  </header>

  <main>
    <h1>{title}</h1>
    <slot />  <!-- Default slot -->
  </main>

  <footer>
    <slot name="footer" />
  </footer>
</div>
```

**Usage:**

```astro
---
import Layout from './Layout.astro';
---

<Layout title="My Page">
  <nav slot="header">
    <a href="/">Home</a>
    <a href="/about">About</a>
  </nav>

  <p>This goes in the default slot (main content)</p>

  <p slot="footer">Copyright 2025</p>
</Layout>
```

**Rendered output:**

```html
<div class="layout">
  <header>
    <nav>
      <a href="/">Home</a>
      <a href="/about">About</a>
    </nav>
  </header>

  <main>
    <h1>My Page</h1>
    <p>This goes in the default slot (main content)</p>
  </main>

  <footer>
    <p>Copyright 2025</p>
  </footer>
</div>
```

### Slots with Fragments

Pass multiple elements to a slot without wrapper `<div>`:

```astro
---
import CustomTable from './CustomTable.astro';
---

<CustomTable>
  <Fragment slot="header">
    <tr><th>Name</th><th>Age</th></tr>
  </Fragment>

  <Fragment slot="body">
    <tr><td>Alice</td><td>30</td></tr>
    <tr><td>Bob</td><td>25</td></tr>
  </Fragment>
</CustomTable>
```

**Component definition:**

```astro
---
// src/components/CustomTable.astro
---

<table>
  <thead><slot name="header" /></thead>
  <tbody><slot name="body" /></tbody>
</table>
```

### Fallback Content

Provide default content when slot is empty:

```astro
---
// src/components/Card.astro
---

<div class="card">
  <slot name="image">
    <!-- Fallback: shown if no image slot provided -->
    <img src="/default-image.png" alt="Default" />
  </slot>

  <slot>
    <!-- Fallback: shown if no default content provided -->
    <p>No content provided</p>
  </slot>
</div>
```

**Usage:**

```astro
<!-- No image slot: uses fallback -->
<Card>
  <p>Custom content</p>
</Card>

<!-- Provides image slot: no fallback -->
<Card>
  <img src="/custom.png" alt="Custom" slot="image" />
  <p>Custom content</p>
</Card>
```

**Important:** Fallback content only displays when no matching `slot="name"` is provided. Empty slots (`<div slot="image"></div>`) won't trigger fallback.

### Transferring Slots

Pass slots from one component to another:

```astro
---
// src/layouts/BaseLayout.astro
---

<html>
  <head>
    <slot name="head" />
  </head>
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
  <!-- Transfer the 'head' slot to parent -->
  <slot name="head" slot="head" />

  <!-- Transfer default slot to parent -->
  <article>
    <slot />
  </article>
</BaseLayout>
```

**Usage:**

```astro
---
import BlogLayout from '../layouts/BlogLayout.astro';
---

<BlogLayout>
  <title slot="head">My Blog Post</title>
  <h1>Post Title</h1>
  <p>Post content...</p>
</BlogLayout>
```

**Result:** Both `head` and default slots are transferred through `BlogLayout` to `BaseLayout`.

## Accessing Slot Content

Use `Astro.slots` to check and access slots programmatically:

```astro
---
const hasHeader = Astro.slots.has('header');
const headerContent = await Astro.slots.render('header');
---

{hasHeader ? (
  <header>{headerContent}</header>
) : (
  <p>No header provided</p>
)}

<slot />
```

**Useful methods:**
- `Astro.slots.has(name)` - Check if slot exists
- `Astro.slots.render(name)` - Render slot as string

## HTML Components

Astro supports `.html` files as components.

### Basic HTML Component

```html
<!-- src/components/Banner.html -->
<div class="banner">
  <h2>Welcome!</h2>
  <slot></slot>
</div>
```

**Usage:**

```astro
---
import Banner from './Banner.html';
---

<Banner>
  <p>This is the banner content</p>
</Banner>
```

### Limitations

HTML components have restrictions:

❌ **Cannot use:**
- Frontmatter (component script)
- Server-side imports
- Dynamic expressions
- TypeScript

✅ **Can use:**
- Plain HTML
- `<slot />` elements
- Assets from `public/` folder

**Scripts are inlined:**

```html
<!-- src/components/Widget.html -->
<div id="widget">Widget</div>

<script>
  // Treated as if it has is:inline directive
  document.getElementById('widget').addEventListener('click', () => {
    alert('Clicked!');
  });
</script>
```

### When to Use HTML Components

**Use HTML components when:**
- Migrating from plain HTML sites
- Need guaranteed static output
- Ensuring no dynamic features

**Use Astro components when:**
- Need imports or data fetching
- Want dynamic content
- Need TypeScript
- Want server-side logic

## Component Patterns

### Pattern 1: Layout Components

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
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width" />
    <title>{title}</title>
    {description && <meta name="description" content={description} />}
    <slot name="head" />
  </head>
  <body>
    <slot />
  </body>
</html>
```

### Pattern 2: Reusable UI Components

```astro
---
// src/components/Button.astro
interface Props {
  variant?: 'primary' | 'secondary';
  size?: 'sm' | 'md' | 'lg';
  href?: string;
}

const { variant = 'primary', size = 'md', href } = Astro.props;

const Tag = href ? 'a' : 'button';
---

<Tag
  class:list={['btn', `btn-${variant}`, `btn-${size}`]}
  href={href}
>
  <slot />
</Tag>

<style>
  .btn {
    padding: 0.5rem 1rem;
    border-radius: 0.25rem;
    font-weight: 600;
  }
  .btn-primary {
    background: #3b82f6;
    color: white;
  }
  .btn-secondary {
    background: #6b7280;
    color: white;
  }
  .btn-sm { font-size: 0.875rem; }
  .btn-md { font-size: 1rem; }
  .btn-lg { font-size: 1.125rem; }
</style>
```

**Usage:**

```astro
<Button variant="primary" size="lg">Click me</Button>
<Button variant="secondary" href="/about">Learn more</Button>
```

### Pattern 3: Container Components

```astro
---
// src/components/Container.astro
interface Props {
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl';
}

const { maxWidth = 'lg' } = Astro.props;

const widths = {
  sm: 'max-w-screen-sm',
  md: 'max-w-screen-md',
  lg: 'max-w-screen-lg',
  xl: 'max-w-screen-xl',
};
---

<div class:list={['mx-auto px-4', widths[maxWidth]]}>
  <slot />
</div>
```

### Pattern 4: Conditional Rendering

```astro
---
// src/components/Alert.astro
interface Props {
  type?: 'info' | 'warning' | 'error' | 'success';
  title?: string;
  dismissible?: boolean;
}

const { type = 'info', title, dismissible = false } = Astro.props;
---

<div class:list={['alert', `alert-${type}`]}>
  {title && <h3>{title}</h3>}
  <div class="content">
    <slot />
  </div>
  {dismissible && (
    <button class="close" aria-label="Close">×</button>
  )}
</div>
```

### Pattern 5: Data Fetching Component

```astro
---
// src/components/UserProfile.astro
interface Props {
  userId: string;
}

const { userId } = Astro.props;

// Fetch data at build time or server render time
const user = await fetch(`https://api.example.com/users/${userId}`)
  .then(r => r.json())
  .catch(() => null);
---

{user ? (
  <div class="profile">
    <h2>{user.name}</h2>
    <p>{user.email}</p>
    <img src={user.avatar} alt={user.name} />
  </div>
) : (
  <p>User not found</p>
)}
```

## Best Practices for AI Agents

### 1. Use Astro Components for Static Content

```astro
<!-- ✅ Good: Static header -->
<Header />

<!-- ❌ Avoid: Framework component for static content -->
<ReactHeader client:load />
```

### 2. Define Props with TypeScript

```astro
<!-- ✅ Good: Type-safe props -->
---
interface Props {
  title: string;
  count?: number;
}
const { title, count = 0 } = Astro.props;
---

<!-- ❌ Avoid: No types -->
---
const { title, count } = Astro.props;
---
```

### 3. Use Named Slots for Flexibility

```astro
<!-- ✅ Good: Named slots for clear structure -->
<Layout>
  <nav slot="header">...</nav>
  <p>Content</p>
  <footer slot="footer">...</footer>
</Layout>

<!-- ❌ Avoid: Single slot with complex logic -->
<Layout>
  <div>...</div> <!-- Where does this go? -->
</Layout>
```

### 4. Provide Fallback Content

```astro
<!-- ✅ Good: Fallback for missing content -->
<slot name="image">
  <img src="/default.png" alt="Default" />
</slot>

<!-- ❌ Avoid: No fallback (could be empty) -->
<slot name="image" />
```

### 5. Keep Components Focused

```astro
<!-- ✅ Good: Single responsibility -->
<Button variant="primary">Click me</Button>

<!-- ❌ Avoid: Component doing too much -->
<MegaComponent
  showHeader={true}
  showFooter={true}
  fetchData={true}
  enableAnalytics={true}
/>
```

## Common Patterns

### Pattern: SEO Component

```astro
---
// src/components/SEO.astro
interface Props {
  title: string;
  description: string;
  image?: string;
  canonicalURL?: string;
}

const { title, description, image, canonicalURL } = Astro.props;
const socialImage = image || '/default-og.png';
---

<title>{title}</title>
<meta name="description" content={description} />

{/* Open Graph */}
<meta property="og:title" content={title} />
<meta property="og:description" content={description} />
<meta property="og:image" content={socialImage} />

{/* Twitter */}
<meta name="twitter:card" content="summary_large_image" />
<meta name="twitter:title" content={title} />
<meta name="twitter:description" content={description} />
<meta name="twitter:image" content={socialImage} />

{/* Canonical */}
{canonicalURL && <link rel="canonical" href={canonicalURL} />}
```

### Pattern: Wrapper with Variants

```astro
---
// src/components/Section.astro
interface Props {
  variant?: 'default' | 'dark' | 'accent';
  spacing?: 'sm' | 'md' | 'lg';
}

const { variant = 'default', spacing = 'md' } = Astro.props;
---

<section
  class:list={[
    'section',
    `section-${variant}`,
    `spacing-${spacing}`
  ]}
>
  <slot />
</section>

<style>
  .section { width: 100%; }
  .section-default { background: white; }
  .section-dark { background: #1a1a1a; color: white; }
  .section-accent { background: #3b82f6; color: white; }

  .spacing-sm { padding: 2rem 0; }
  .spacing-md { padding: 4rem 0; }
  .spacing-lg { padding: 8rem 0; }
</style>
```

## Quick Reference

```astro
<!-- Component structure -->
---
// Component script (server-only)
const { prop } = Astro.props;
---
<!-- Template (HTML + expressions) -->

<!-- Props -->
interface Props {
  title: string;
  optional?: number;
}

<!-- Default slot -->
<div>
  <slot />
</div>

<!-- Named slots -->
<header><slot name="header" /></header>
<slot />  <!-- default -->
<footer><slot name="footer" /></footer>

<!-- Fallback content -->
<slot>
  <p>Default content if slot empty</p>
</slot>

<!-- Using slots -->
<Component>
  <div slot="header">Header content</div>
  <p>Default slot content</p>
</Component>

<!-- Conditional rendering -->
{condition && <div>Shown if true</div>}
{condition ? <div>True</div> : <div>False</div>}

<!-- Lists -->
{items.map(item => <li>{item}</li>)}
```

## Resources

- [Astro Syntax Reference](https://docs.astro.build/en/reference/astro-syntax/)
- [Component Props Reference](https://docs.astro.build/en/reference/api-reference/#astroprops)
- [Astro.slots Reference](https://docs.astro.build/en/reference/api-reference/#astroslots)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro component development
**See Also**:
- 06-framework-components.md for framework integration
- 02-islands-architecture.md for Islands concepts
