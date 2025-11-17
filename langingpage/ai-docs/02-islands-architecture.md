# Astro Islands Architecture

## Overview

Astro pioneered **Islands Architecture**, a frontend pattern that renders most pages as static HTML with interactive "islands" of JavaScript added only where needed. This approach avoids heavy JavaScript payloads that slow down modern web applications.

**Key Concept**: Instead of shipping entire frameworks to the client, Astro delivers static HTML and hydrates only the specific components that need interactivity.

## Historical Context

The term **"component island"** originated with Etsy's Katie Sylor-Miller in 2019. Jason Miller (creator of Preact) expanded the concept in 2020, describing it as:

> *"Render HTML pages on the server, and inject placeholders around dynamic regions that can then be hydrated on the client into small self-contained widgets."*

This technique, also called **partial or selective hydration**, contrasts with single-page applications (SPAs) that hydrate entire sites as monolithic JavaScript bundles.

### Traditional SPA vs Islands Architecture

**Traditional SPA:**
```
Server → Client: Full JavaScript bundle
Client: Hydrate entire page
Result: Heavy initial load, slow Time to Interactive (TTI)
```

**Islands Architecture:**
```
Server → Client: Static HTML + minimal JS for interactive components only
Client: Hydrate individual islands independently
Result: Fast initial load, instant interactivity where needed
```

## What Is an Island?

An island is an **enhanced UI component** within an otherwise static HTML page. Astro supports two types:

### 1. Client Islands

Interactive JavaScript components hydrated separately from the page. By default, Astro renders all components as static HTML/CSS, stripping client-side JavaScript. You activate interactivity using `client:*` directives.

#### Basic Example

```astro
---
// Static by default - no JavaScript sent to client
import Header from '../components/Header.astro';
import Footer from '../components/Footer.astro';

// Interactive components - JavaScript will be sent
import InteractiveCounter from '../components/Counter.jsx';
import VideoPlayer from '../components/VideoPlayer.svelte';
---

<html>
  <body>
    <!-- Static: No JS -->
    <Header />

    <!-- Island: Hydrated on page load -->
    <InteractiveCounter client:load />

    <!-- Island: Hydrated when visible -->
    <VideoPlayer client:visible />

    <!-- Static: No JS -->
    <Footer />
  </body>
</html>
```

#### Client Directives (Loading Strategies)

Astro provides several strategies for controlling when islands load:

| Directive | When It Loads | Use Case |
|-----------|---------------|----------|
| `client:load` | Immediately on page load | Critical interactive elements (search bar, mobile menu) |
| `client:idle` | When browser is idle | Non-critical elements (chat widget, newsletter signup) |
| `client:visible` | When enters viewport | Below-fold content (image gallery, comments section) |
| `client:media` | When media query matches | Responsive components (mobile-only nav) |
| `client:only` | Skip server rendering entirely | Client-only components (no SSR support) |

**Example with Different Strategies:**

```astro
---
import SearchBar from '../components/SearchBar.jsx';
import ChatWidget from '../components/Chat.jsx';
import ImageGallery from '../components/Gallery.jsx';
import MobileMenu from '../components/MobileMenu.jsx';
---

<!-- Critical: Load immediately -->
<SearchBar client:load />

<!-- Non-critical: Load when browser idle -->
<ChatWidget client:idle />

<!-- Below fold: Load when user scrolls to it -->
<ImageGallery client:visible />

<!-- Mobile only: Load on small screens -->
<MobileMenu client:media="(max-width: 768px)" />
```

### 2. Server Islands

Dynamic server-rendered components that process expensive operations independently. Use the `server:defer` directive to defer expensive server-side rendering.

**Example:**

```astro
---
import Avatar from '../components/Avatar.astro';
import ProductList from '../components/ProductList.astro';
---

<!-- Deferred server rendering with fallback -->
<Avatar server:defer>
  <div slot="fallback">Loading avatar...</div>
</Avatar>

<!-- Main content renders immediately -->
<ProductList />
```

**Benefits:**
- Separates expensive server logic from main rendering
- Enables aggressive caching of static content
- Displays fallback content while personalized data loads
- Works with any hosting infrastructure

## Benefits of Islands Architecture

### Client Islands Benefits

1. **Performance Gains**
   - Minimal JavaScript shipping (only what's needed)
   - Faster page loads and Time to Interactive (TTI)
   - Reduced bundle sizes

2. **Parallel Loading**
   - Islands load independently without blocking dependencies
   - One slow island doesn't delay others

3. **Per-Component Control**
   - Fine-grained control over when JavaScript executes
   - Optimize for user experience (critical vs non-critical)

4. **Fast-by-Default Architecture**
   - Default to static HTML
   - Opt-in to JavaScript complexity

### Server Islands Benefits

1. **Performance Optimization**
   - Separates expensive operations from initial page load
   - Allows static content to render immediately

2. **Caching Efficiency**
   - Aggressive caching of static content
   - Dynamic content loaded separately

3. **Progressive Enhancement**
   - Fallback content displayed during loading
   - Better perceived performance

## Framework Flexibility

One of Astro's superpowers is **framework agnosticism**. You can mix multiple frameworks on the same page:

```astro
---
import ReactCounter from '../components/Counter.jsx';
import VueCalendar from '../components/Calendar.vue';
import SvelteChart from '../components/Chart.svelte';
import SolidButton from '../components/Button.tsx';
---

<div>
  <!-- Mix frameworks freely -->
  <ReactCounter client:load />
  <VueCalendar client:visible />
  <SvelteChart client:idle />
  <SolidButton client:load />
</div>
```

**Why This Matters:**
- Use the best tool for each component
- Migrate gradually between frameworks
- Leverage existing component libraries
- Team members can use preferred frameworks

## Island Isolation and State Sharing

By default, islands are **isolated** and run independently. However, you can share state between islands:

### Using Framework-Specific Stores

**Example with Nano Stores (Framework-Agnostic):**

```javascript
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

## Best Practices

### 1. Default to Static

**Always start with static components** and only add interactivity when needed:

```astro
<!-- Good: Static by default -->
<Card>
  <h2>Product Name</h2>
  <p>Description</p>
</Card>

<!-- Only add interactivity when necessary -->
<AddToCartButton client:idle />
```

### 2. Choose the Right Loading Strategy

**Decision Tree:**

- **Immediately visible and critical?** → `client:load`
- **Below the fold?** → `client:visible`
- **Non-critical interaction?** → `client:idle`
- **Responsive/conditional?** → `client:media`

### 3. Minimize Island Size

Keep islands small and focused:

```astro
<!-- Bad: Entire section is an island -->
<ProductSection client:load>
  <Header />
  <Image />
  <Description />
  <AddToCartButton />
</ProductSection>

<!-- Good: Only interactive part is an island -->
<ProductSection>
  <Header />
  <Image />
  <Description />
  <AddToCartButton client:idle />
</ProductSection>
```

### 4. Avoid Nested Islands

Don't nest `client:*` directives unnecessarily:

```astro
<!-- Bad: Redundant nesting -->
<ParentComponent client:load>
  <ChildComponent client:load />
</ParentComponent>

<!-- Good: Single directive -->
<ParentComponent client:load>
  <ChildComponent />
</ParentComponent>
```

### 5. Use Server Islands for Expensive Operations

Defer expensive server-side work:

```astro
<!-- Personalized recommendations (expensive) -->
<Recommendations server:defer>
  <div slot="fallback">Loading recommendations...</div>
</Recommendations>

<!-- Static content renders immediately -->
<StaticContent />
```

## Common Patterns

### Pattern 1: Progressive Enhancement

Build with HTML first, enhance with JavaScript:

```astro
---
import EnhancedForm from '../components/EnhancedForm.jsx';
---

<!-- Works without JavaScript -->
<form method="POST" action="/api/submit">
  <input name="email" type="email" required />
  <button type="submit">Subscribe</button>
</form>

<!-- Enhanced with JavaScript when available -->
<EnhancedForm client:idle />
```

### Pattern 2: Lazy-Loaded Modals/Dialogs

Only load when needed:

```astro
---
import Modal from '../components/Modal.jsx';
---

<!-- Modal loaded only when visible -->
<Modal client:visible>
  <p>This content loads when modal appears</p>
</Modal>
```

### Pattern 3: Responsive Components

Different components for different screen sizes:

```astro
---
import DesktopNav from '../components/DesktopNav.astro';
import MobileNav from '../components/MobileNav.jsx';
---

<!-- Static desktop nav -->
<DesktopNav />

<!-- Interactive mobile nav (only on mobile) -->
<MobileNav client:media="(max-width: 768px)" />
```

### Pattern 4: Third-Party Widgets

Defer non-critical third-party code:

```astro
---
import ChatWidget from '../components/ChatWidget.jsx';
import AnalyticsDashboard from '../components/Analytics.jsx';
---

<!-- Load chat when browser is idle -->
<ChatWidget client:idle />

<!-- Load analytics when scrolled into view -->
<AnalyticsDashboard client:visible />
```

## Performance Impact

### Metrics

Islands Architecture typically results in:

- **40% faster load times** compared to traditional SPAs
- **90% less JavaScript** shipped to the client
- **Better Core Web Vitals** scores (LCP, FID, CLS)
- **Improved SEO** due to faster rendering

### Example Comparison

**Traditional SPA:**
```
Bundle size: 500KB JavaScript
Time to Interactive: 3.5s
First Contentful Paint: 2.1s
```

**Astro Islands:**
```
Bundle size: 50KB JavaScript (only for islands)
Time to Interactive: 0.8s
First Contentful Paint: 0.4s
```

## Key Takeaways for AI Agents

When working with Astro Islands:

1. **Default to static** - No `client:*` directive means no JavaScript to client
2. **Explicit opt-in** - Use directives only when interactivity is needed
3. **Choose wisely** - Select the right loading strategy for each island
4. **Keep islands small** - Minimize the amount of code that hydrates
5. **Mix frameworks freely** - Use the best tool for each component
6. **Share state properly** - Use framework-agnostic stores like nano-stores
7. **Progressive enhancement** - Build base functionality without JavaScript

## Framework Integration Examples

### React

```astro
---
import { Counter } from '../components/Counter.jsx';
---

<Counter client:load />
```

### Vue

```astro
---
import Calendar from '../components/Calendar.vue';
---

<Calendar client:visible />
```

### Svelte

```astro
---
import Chart from '../components/Chart.svelte';
---

<Chart client:idle />
```

### Solid

```astro
---
import Button from '../components/Button.tsx';
---

<Button client:load />
```

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent reference for Astro Islands Architecture
**Next Module**: Component Development (03-components.md)
