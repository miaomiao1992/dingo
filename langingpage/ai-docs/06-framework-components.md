# Front-end Framework Components

Build your Astro website without sacrificing your favorite component framework. Create Astro islands with the UI frameworks of your choice.

## Overview

Astro supports multiple front-end frameworks simultaneously:
- **React** - Most popular UI library
- **Preact** - Lightweight React alternative
- **Vue** - Progressive JavaScript framework
- **Svelte** - Compile-time framework
- **Solid** - Fine-grained reactivity
- **Alpine.js** - Minimal JavaScript framework

**Key benefits:**
- Use familiar component syntax
- Mix multiple frameworks on same page
- Islands Architecture for optimal performance
- Choose the right tool for each component

## Official Framework Integrations

Astro provides official integrations for major frameworks:

| Framework | Package | Best For |
|-----------|---------|----------|
| **React** | `@astrojs/react` | Large ecosystems, complex apps |
| **Preact** | `@astrojs/preact` | Lightweight React alternative |
| **Vue** | `@astrojs/vue` | Progressive enhancement |
| **Svelte** | `@astrojs/svelte` | Minimal bundle size |
| **Solid** | `@astrojs/solid-js` | Fine-grained reactivity |
| **Alpine** | `@astrojs/alpinejs` | Minimal interactive widgets |

**Community integrations:** Angular, Qwik, Elm, and more in [Astro integrations directory](https://astro.build/integrations)

## Installing Framework Integrations

### Using `astro add` (Recommended)

The easiest way to add framework support:

```bash
# Add React
pnpm astro add react

# Add Vue
pnpm astro add vue

# Add Svelte
pnpm astro add svelte

# Add multiple frameworks
pnpm astro add react vue svelte
```

**What `astro add` does:**
1. Installs the integration package
2. Installs framework peer dependencies
3. Updates `astro.config.mjs`
4. Provides setup instructions

### Manual Installation

Install manually if you need more control:

```bash
# Install React
pnpm add @astrojs/react react react-dom
```

**Configure in `astro.config.mjs`:**

```javascript
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';

export default defineConfig({
  integrations: [react()],
});
```

### Multiple Frameworks

Install and configure multiple frameworks:

```javascript
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import vue from '@astrojs/vue';
import svelte from '@astrojs/svelte';

export default defineConfig({
  integrations: [
    react(),
    vue(),
    svelte(),
  ],
});
```

## Using Framework Components

### Basic Usage

Import and use framework components like Astro components:

```astro
---
// src/pages/static-components.astro
import MyReactComponent from '../components/MyReactComponent.jsx';
import MyVueComponent from '../components/MyVueComponent.vue';
import MySvelteComponent from '../components/MySvelteComponent.svelte';
---

<html>
  <body>
    <h1>Mix frameworks freely!</h1>
    <MyReactComponent />
    <MyVueComponent />
    <MySvelteComponent />
  </body>
</html>
```

**File extensions by framework:**

| Framework | Extension |
|-----------|-----------|
| React | `.jsx`, `.tsx` |
| Preact | `.jsx`, `.tsx` |
| Vue | `.vue` |
| Svelte | `.svelte` |
| Solid | `.jsx`, `.tsx` |

### Static Rendering (Default)

**By default, framework components render as static HTML:**

```astro
---
import Counter from '../components/Counter.jsx';
---

<!-- Renders static HTML, no JavaScript sent to client -->
<Counter />
```

**Benefits:**
- Fast page loads
- No JavaScript overhead
- SEO-friendly
- Good for templating

**Use for:**
- Static content
- Server-side only components
- Non-interactive UI

## Hydrating Interactive Components

To make components interactive, use `client:*` directives.

### Available Directives

| Directive | When JS Loads | Use Case |
|-----------|---------------|----------|
| `client:load` | Immediately on page load | Critical interactive elements |
| `client:idle` | When browser is idle | Non-critical interactions |
| `client:visible` | When scrolled into view | Below-fold content |
| `client:media` | When media query matches | Responsive components |
| `client:only` | Client-side only (no SSR) | Incompatible with SSR |

### Examples

**Load immediately:**

```astro
---
import SearchBar from '../components/SearchBar.jsx';
---

<!-- Critical: Load JS immediately -->
<SearchBar client:load />
```

**Load when idle:**

```astro
---
import ChatWidget from '../components/ChatWidget.jsx';
---

<!-- Non-critical: Load when browser idle -->
<ChatWidget client:idle />
```

**Load when visible:**

```astro
---
import ImageGallery from '../components/ImageGallery.svelte';
---

<!-- Below fold: Load when scrolled to -->
<ImageGallery client:visible />
```

**Load on media query:**

```astro
---
import MobileMenu from '../components/MobileMenu.jsx';
---

<!-- Mobile only: Load on small screens -->
<MobileMenu client:media="(max-width: 768px)" />
```

**Client-only (no SSR):**

```astro
---
import ClientOnlyComponent from '../components/ClientOnly.jsx';
---

<!-- Skip server rendering entirely -->
<ClientOnlyComponent client:only="react" />
```

### How Hydration Works

1. **Server renders** component to static HTML
2. **Browser receives** HTML (fast initial load)
3. **JavaScript loads** according to directive
4. **Component hydrates** and becomes interactive

**Optimization:**
- Framework code shared across components
- Only sent once per framework per page
- Automatic code splitting

## Mixing Frameworks

Use multiple frameworks on the same page:

```astro
---
// src/pages/mixing-frameworks.astro
import MyReactComponent from '../components/MyReactComponent.jsx';
import MySvelteComponent from '../components/MySvelteComponent.svelte';
import MyVueComponent from '../components/MyVueComponent.vue';
import MySolidComponent from '../components/MySolidComponent.jsx';
---

<div>
  <h1>Multiple Frameworks Working Together</h1>

  <!-- React component -->
  <MyReactComponent client:load />

  <!-- Svelte component -->
  <MySvelteComponent client:idle />

  <!-- Vue component -->
  <MyVueComponent client:visible />

  <!-- Solid component -->
  <MySolidComponent client:load />
</div>
```

**Astro automatically:**
- Recognizes framework by file extension
- Loads appropriate runtime
- Manages hydration per component

### Multiple JSX Frameworks

When using both React and Preact (or other JSX frameworks):

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import preact from '@astrojs/preact';

export default defineConfig({
  integrations: [
    react({
      include: ['**/react/*'],  // React components in react/ folder
    }),
    preact({
      include: ['**/preact/*'], // Preact components in preact/ folder
    }),
  ],
});
```

**Directory structure:**

```
src/components/
├── react/
│   └── Counter.jsx       # React component
└── preact/
    └── Button.jsx        # Preact component
```

## Passing Props

### Basic Props

Pass props from Astro to framework components:

```astro
---
import TodoList from '../components/TodoList.jsx';
import Counter from '../components/Counter.svelte';
---

<div>
  <TodoList initialTodos={["learn Astro", "review PRs"]} />
  <Counter startingCount={1} />
</div>
```

### Supported Prop Types

**Serializable types** (can be passed to hydrated components):

```astro
---
import Component from '../components/Component.jsx';

const supportedProps = {
  // Primitives
  string: "hello",
  number: 42,
  boolean: true,

  // Objects and Arrays
  object: { key: "value" },
  array: [1, 2, 3],

  // Special types
  map: new Map([['key', 'value']]),
  set: new Set([1, 2, 3]),
  regex: /pattern/,
  date: new Date(),
  bigint: 9007199254740991n,
  url: new URL('https://example.com'),

  // Typed Arrays
  uint8: new Uint8Array([1, 2, 3]),
  uint16: new Uint16Array([1, 2, 3]),
  uint32: new Uint32Array([1, 2, 3]),

  // Special values
  infinity: Infinity,
};
---

<Component {...supportedProps} client:load />
```

### Non-Supported Prop Types

**Cannot be passed to hydrated components:**

```astro
---
import Component from '../components/Component.jsx';

// ❌ Functions cannot be serialized
const handleClick = () => console.log('clicked');

// ❌ Symbols cannot be serialized
const mySymbol = Symbol('key');

// ❌ Class instances
const instance = new MyClass();
---

<!-- ❌ This won't work for hydrated components -->
<Component onClick={handleClick} client:load />

<!-- ✅ But works for static (server-only) components -->
<Component onClick={handleClick} />
```

**Workaround for functions:**

Define functions inside the component:

```jsx
// Component.jsx
export default function Component({ apiEndpoint }) {
  const handleClick = () => {
    fetch(apiEndpoint).then(/* ... */);
  };

  return <button onClick={handleClick}>Click</button>;
}
```

```astro
<!-- Pass data, define behavior in component -->
<Component apiEndpoint="/api/data" client:load />
```

## Passing Children

Framework components can receive children from Astro:

### Default Slot (Children)

**React, Preact, Solid** use `children` prop:

```astro
---
// src/pages/component-children.astro
import MyReactSidebar from '../components/MyReactSidebar.jsx';
---

<MyReactSidebar>
  <p>Here is a sidebar with some text and a button.</p>
</MyReactSidebar>
```

```jsx
// src/components/MyReactSidebar.jsx
export default function MyReactSidebar({ children }) {
  return (
    <aside>
      {children}
    </aside>
  );
}
```

**Svelte, Vue** use `<slot />`:

```astro
---
import MySidebar from '../components/MySidebar.svelte';
---

<MySidebar>
  <p>Sidebar content here</p>
</MySidebar>
```

```svelte
<!-- src/components/MySidebar.svelte -->
<aside>
  <slot />
</aside>
```

### Named Slots

Pass multiple, named content sections:

```astro
---
// src/pages/named-slots.astro
import MySidebar from '../components/MySidebar.jsx';
---

<MySidebar>
  <h2 slot="title">Menu</h2>
  <p>Main content here</p>
  <ul slot="social-links">
    <li><a href="https://twitter.com/astrodotbuild">Twitter</a></li>
    <li><a href="https://github.com/withastro">GitHub</a></li>
  </ul>
</MySidebar>
```

**React, Preact, Solid** - Slots become props:

```jsx
// src/components/MySidebar.jsx
export default function MySidebar({ title, children, socialLinks }) {
  return (
    <aside>
      <header>{title}</header>
      <main>{children}</main>
      <footer>{socialLinks}</footer>
    </aside>
  );
}
```

**Note:** `kebab-case` slot names become `camelCase` props:
- `slot="social-links"` → `socialLinks` prop

**Svelte, Vue** - Use named `<slot>`:

```svelte
<!-- src/components/MySidebar.svelte -->
<aside>
  <header><slot name="title" /></header>
  <main><slot /></main>
  <footer><slot name="social-links" /></footer>
</aside>
```

**Note:** `kebab-case` preserved in Svelte/Vue

## Nesting Framework Components

### Same Framework

Nest components from the same framework:

```astro
---
import MyReactSidebar from '../components/MyReactSidebar.jsx';
import MyReactButton from '../components/MyReactButton.jsx';
---

<MyReactSidebar>
  <p>Sidebar content</p>
  <div slot="actions">
    <MyReactButton client:idle />
  </div>
</MyReactSidebar>
```

### Mixed Frameworks

Nest components from different frameworks:

```astro
---
import MyReactSidebar from '../components/MyReactSidebar.jsx';
import MySvelteButton from '../components/MySvelteButton.svelte';
import MyVueModal from '../components/MyVueModal.vue';
---

<MyReactSidebar client:load>
  <p>Sidebar with mixed framework children</p>
  <MySvelteButton client:idle />
  <MyVueModal client:visible />
</MyReactSidebar>
```

**Important:** Each component becomes its own island with its own framework runtime.

### Restrictions

**Cannot mix frameworks in component files:**

```jsx
// ❌ WRONG: Cannot import Vue in React component
import MyVueComponent from './MyVueComponent.vue';

export default function MyReactComponent() {
  return <MyVueComponent />; // Won't work!
}
```

**Must mix in Astro files:**

```astro
<!-- ✅ CORRECT: Mix in .astro files -->
---
import MyReactComponent from '../components/MyReactComponent.jsx';
import MyVueComponent from '../components/MyVueComponent.vue';
---

<div>
  <MyReactComponent />
  <MyVueComponent />
</div>
```

## Using Astro Components with Framework Components

### Can I import .astro in framework components?

**No.** Framework components are islands and must be pure framework code.

```jsx
// ❌ WRONG: Cannot import .astro in .jsx
import MyAstroComponent from './MyAstroComponent.astro';

export default function MyReactComponent() {
  return <MyAstroComponent />; // Won't work!
}
```

### Passing Astro Components as Children

**You can** pass Astro components as static children:

```astro
---
// ✅ CORRECT: Pass Astro component as child to framework component
import MyReactComponent from '../components/MyReactComponent.jsx';
import MyAstroComponent from '../components/MyAstroComponent.astro';
---

<MyReactComponent>
  <MyAstroComponent slot="name" />
</MyReactComponent>
```

**How it works:**
- Astro component renders to static HTML
- Static HTML passed as children to framework component
- No Astro runtime needed in browser

## Can I Hydrate Astro Components?

**No.** Astro components cannot be hydrated with `client:*` directives.

```astro
---
import MyAstroComponent from '../components/MyAstroComponent.astro';
---

<!-- ❌ ERROR: Cannot hydrate Astro components -->
<MyAstroComponent client:load />
```

**Why?** Astro components:
- Are HTML-only templating components
- Have no client-side runtime
- Render to static HTML only

### Adding Interactivity to Astro Components

Use `<script>` tags for client-side JavaScript:

```astro
---
// src/components/InteractiveAstro.astro
---

<button id="my-button">Click me</button>

<script>
  // Runs in browser
  document.getElementById('my-button').addEventListener('click', () => {
    alert('Clicked!');
  });
</script>
```

**Or use a framework component instead:**

```astro
---
import InteractiveButton from '../components/InteractiveButton.jsx';
---

<!-- Use framework component for complex interactivity -->
<InteractiveButton client:load />
```

## Framework-Specific Examples

### React Example

```jsx
// src/components/Counter.jsx
import { useState } from 'react';

export default function Counter({ initialCount = 0 }) {
  const [count, setCount] = useState(initialCount);

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
    </div>
  );
}
```

```astro
---
import Counter from '../components/Counter.jsx';
---

<Counter initialCount={5} client:load />
```

### Vue Example

```vue
<!-- src/components/Counter.vue -->
<script setup>
import { ref } from 'vue';

const props = defineProps({
  initialCount: {
    type: Number,
    default: 0
  }
});

const count = ref(props.initialCount);
</script>

<template>
  <div>
    <p>Count: {{ count }}</p>
    <button @click="count++">Increment</button>
  </div>
</template>
```

```astro
---
import Counter from '../components/Counter.vue';
---

<Counter :initial-count="5" client:load />
```

### Svelte Example

```svelte
<!-- src/components/Counter.svelte -->
<script>
  export let initialCount = 0;
  let count = initialCount;
</script>

<div>
  <p>Count: {count}</p>
  <button on:click={() => count++}>
    Increment
  </button>
</div>
```

```astro
---
import Counter from '../components/Counter.svelte';
---

<Counter initialCount={5} client:load />
```

### Solid Example

```jsx
// src/components/Counter.jsx (Solid)
import { createSignal } from 'solid-js';

export default function Counter(props) {
  const [count, setCount] = createSignal(props.initialCount || 0);

  return (
    <div>
      <p>Count: {count()}</p>
      <button onClick={() => setCount(count() + 1)}>
        Increment
      </button>
    </div>
  );
}
```

```astro
---
import Counter from '../components/Counter.jsx';
---

<Counter initialCount={5} client:load />
```

## Best Practices for AI Agents

### 1. Choose the Right Framework

```astro
<!-- ✅ Good: Use framework for complex interactivity -->
<ComplexDataGrid client:load />

<!-- ❌ Avoid: Framework for simple static content -->
<SimpleCard client:load />  <!-- Use Astro component instead -->
```

### 2. Use Appropriate Hydration Directive

```astro
<!-- ✅ Good: Match directive to use case -->
<SearchBar client:load />        <!-- Critical -->
<Comments client:visible />      <!-- Below fold -->
<ChatWidget client:idle />       <!-- Non-critical -->

<!-- ❌ Avoid: Everything with client:load -->
<Footer client:load />           <!-- Should be static -->
```

### 3. Pass Only Serializable Props to Hydrated Components

```astro
<!-- ✅ Good: Serializable props -->
<Component data={data} count={5} client:load />

<!-- ❌ Bad: Function props to hydrated component -->
<Component onClick={handleClick} client:load />

<!-- ✅ Good: Function props to static component -->
<Component onClick={handleClick} />
```

### 4. Mix Frameworks in Astro Files Only

```astro
<!-- ✅ Good: Mix in .astro -->
<ReactComponent />
<VueComponent />

<!-- ❌ Bad: Mix in .jsx -->
<!-- Cannot import Vue in React file -->
```

### 5. Keep Framework Islands Small

```astro
<!-- ✅ Good: Small interactive island -->
<StaticContent>
  <InteractiveButton client:idle />
</StaticContent>

<!-- ❌ Avoid: Large unnecessary island -->
<EntirePageWrapper client:load>
  <MostlyStaticContent />
</EntirePageWrapper>
```

## Common Patterns

### Pattern 1: Progressive Enhancement

```astro
---
import EnhancedForm from '../components/EnhancedForm.jsx';
---

<!-- Works without JavaScript -->
<form method="POST" action="/api/submit">
  <input name="email" type="email" required />
  <button type="submit">Subscribe</button>
</form>

<!-- Enhanced with JavaScript -->
<EnhancedForm client:idle />
```

### Pattern 2: Conditional Framework Loading

```astro
---
import MobileMenu from '../components/MobileMenu.jsx';
import DesktopMenu from '../components/DesktopMenu.astro';
---

<!-- Static desktop menu -->
<DesktopMenu />

<!-- Interactive mobile menu (only loads on mobile) -->
<MobileMenu client:media="(max-width: 768px)" />
```

### Pattern 3: Shared State Management

```jsx
// src/stores/cart.js
import { atom } from 'nanostores';

export const cartItems = atom([]);
```

```jsx
// React component
import { useStore } from '@nanostores/react';
import { cartItems } from '../stores/cart';

export default function CartButton() {
  const $cartItems = useStore(cartItems);
  return <button>Cart ({$cartItems.length})</button>;
}
```

```svelte
<!-- Svelte component -->
<script>
  import { cartItems } from '../stores/cart';
</script>

<button>Cart ({$cartItems.length})</button>
```

## Troubleshooting

### Framework Not Recognized

**Issue:** Component not rendering correctly.

**Solution:** Check file extension and framework installation:

```bash
# Verify framework is installed
pnpm list @astrojs/react
pnpm list @astrojs/vue

# Check astro.config.mjs
```

### Props Not Working

**Issue:** Props undefined in component.

**Solution:** Ensure prop types are serializable for hydrated components.

### Hydration Not Working

**Issue:** Component not interactive.

**Solution:** Add `client:*` directive:

```astro
<!-- Add directive for interactivity -->
<Component client:load />
```

## Quick Reference

```astro
<!-- Install framework -->
pnpm astro add react

<!-- Import and use -->
---
import Component from '../components/Component.jsx';
---

<!-- Static (default) -->
<Component />

<!-- Hydration directives -->
<Component client:load />       <!-- Load immediately -->
<Component client:idle />       <!-- Load when idle -->
<Component client:visible />    <!-- Load when visible -->
<Component client:media="..." /> <!-- Load on media query -->
<Component client:only="react" /> <!-- Client-only -->

<!-- Pass props -->
<Component data={data} count={5} />

<!-- Pass children -->
<Component>
  <p>Children content</p>
</Component>

<!-- Named slots -->
<Component>
  <h2 slot="title">Title</h2>
  <p>Default content</p>
</Component>
```

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for using framework components in Astro
**See Also**: 02-islands-architecture.md for Islands concepts
