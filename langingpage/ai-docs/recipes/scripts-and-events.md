# Scripts and Event Handling in Astro

You can send JavaScript to the browser and add functionality to your Astro components using `<script>` tags. This provides interactivity without requiring a UI framework like React, Svelte, or Vue.

## Overview

**Scripts enable:**
- Event handling (clicks, form submissions, etc.)
- Analytics tracking
- Animations
- Dynamic content updates
- Full JavaScript capabilities in the browser

**Benefits:**
- No framework overhead
- Faster page loads
- Simpler architecture
- Standard JavaScript

## Client-Side Scripts

### Basic Script

Add a `<script>` tag to any `.astro` component:

```astro
---
// src/components/ConfettiButton.astro
---

<button data-confetti-button>Celebrate!</button>

<script>
  // Import from npm package
  import confetti from 'canvas-confetti';

  // Find component DOM elements
  const buttons = document.querySelectorAll('[data-confetti-button]');

  // Add event listeners
  buttons.forEach((button) => {
    button.addEventListener('click', () => confetti());
  });
</script>
```

**What happens:**
1. Script runs on the client (browser)
2. Can import npm packages
3. Has access to the DOM
4. Can add event listeners

## Script Processing

Astro automatically processes `<script>` tags with **no attributes** (except `src`):

### Default Processing Features

```astro
<script>
  // ✅ TypeScript support (all scripts are TypeScript by default)
  const greeting: string = 'Hello';

  // ✅ Import bundling (local files and npm modules)
  import confetti from 'canvas-confetti';
  import { myFunction } from './utils';

  // ✅ Automatic type="module"
  // Scripts become ES modules automatically

  // ✅ Deduplication
  // Script only included once even if component used multiple times

  // ✅ Automatic inlining
  // Small scripts are inlined to reduce HTTP requests
</script>
```

**Processing benefits:**
- TypeScript compilation
- Import resolution
- Module bundling
- Code optimization
- Single inclusion per page

### Unprocessed Scripts

Astro will **not process** scripts with any attribute other than `src`.

**Add `is:inline` to opt out of processing:**

```astro
<script is:inline>
  // Will be rendered exactly as written
  // Not transformed or bundled
  // No TypeScript, no import resolution
  // Duplicated for each component instance
  console.log('Inline script');
</script>
```

**When to use `is:inline`:**
- External CDN scripts
- Scripts in `public/` folder
- Need exact control over output
- Legacy code that can't be bundled

## Including JavaScript Files

### Local Scripts (in `src/`)

Import scripts from your project's `src/` directory:

```astro
---
// src/components/LocalScripts.astro
---

<!-- Relative path to script in src/ -->
<script src="../scripts/local.js"></script>

<!-- TypeScript files work too -->
<script src="./script-with-types.ts"></script>
```

**Processing:**
- Scripts are processed (bundled, TypeScript, etc.)
- Imports resolved
- Single inclusion per page

### External Scripts (CDN or `public/`)

For scripts outside `src/`, use `is:inline`:

```astro
---
// src/components/ExternalScripts.astro
---

<!-- Absolute path to script in public/ -->
<script is:inline src="/my-script.js"></script>

<!-- Full URL to remote server -->
<script is:inline src="https://my-analytics.com/script.js"></script>

<!-- CDN script -->
<script is:inline src="https://cdn.jsdelivr.net/npm/confetti@1.0.0"></script>
```

**Characteristics:**
- Skips Astro processing
- No bundling or optimization
- Loaded directly from URL
- Not dedup licated

## Event Handling

### Standard addEventListener

Astro uses standard HTML and JavaScript (no framework-specific syntax):

```astro
---
// src/components/AlertButton.astro
---

<button class="alert">Click me!</button>

<script>
  // Find all buttons with the alert class
  const buttons = document.querySelectorAll('button.alert');

  // Handle clicks on each button
  buttons.forEach((button) => {
    button.addEventListener('click', () => {
      alert('Button was clicked!');
    });
  });
</script>
```

**Key points:**
- Use `querySelectorAll` to handle multiple instances
- Standard `addEventListener` API
- Works with component used multiple times on page

### Multiple Event Types

```astro
<button id="my-button">Hover or Click</button>

<script>
  const button = document.querySelector('#my-button');

  if (button) {
    // Click event
    button.addEventListener('click', () => {
      console.log('Clicked!');
    });

    // Hover events
    button.addEventListener('mouseenter', () => {
      button.style.background = '#3b82f6';
    });

    button.addEventListener('mouseleave', () => {
      button.style.background = '';
    });
  }
</script>
```

### Form Events

```astro
<form id="contact-form">
  <input name="email" type="email" required />
  <button type="submit">Submit</button>
</form>

<script>
  const form = document.querySelector('#contact-form');

  form?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(form as HTMLFormElement);
    const email = formData.get('email');

    try {
      const response = await fetch('/api/subscribe', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      if (response.ok) {
        alert('Subscribed!');
      }
    } catch (error) {
      console.error('Error:', error);
    }
  });
</script>
```

## Web Components (Custom Elements)

Create reusable components with custom behavior using the Web Components standard.

### Basic Web Component

```astro
---
// src/components/AstroHeart.astro
---

<!-- Wrap elements in custom element -->
<astro-heart>
  <button aria-label="Heart">💜</button> × <span>0</span>
</astro-heart>

<script>
  // Define behavior for custom element
  class AstroHeart extends HTMLElement {
    connectedCallback() {
      let count = 0;

      const heartButton = this.querySelector('button');
      const countSpan = this.querySelector('span');

      // Update count on click
      heartButton?.addEventListener('click', () => {
        count++;
        if (countSpan) {
          countSpan.textContent = count.toString();
        }
      });
    }
  }

  // Register custom element
  customElements.define('astro-heart', AstroHeart);
</script>
```

**Advantages:**

1. **Scoped queries**: `this.querySelector()` only searches within the custom element
2. **Multiple instances**: `connectedCallback()` runs for each instance
3. **Encapsulation**: Each instance has its own state

### Web Component with Cleanup

```astro
<astro-timer>
  <div class="timer">0</div>
</astro-timer>

<script>
  class AstroTimer extends HTMLElement {
    private interval?: number;
    private count = 0;

    connectedCallback() {
      const display = this.querySelector('.timer');

      // Start timer
      this.interval = window.setInterval(() => {
        this.count++;
        if (display) {
          display.textContent = this.count.toString();
        }
      }, 1000);
    }

    // Cleanup when element removed
    disconnectedCallback() {
      if (this.interval) {
        clearInterval(this.interval);
      }
    }
  }

  customElements.define('astro-timer', AstroTimer);
</script>
```

**Lifecycle methods:**
- `connectedCallback()` - Element added to DOM
- `disconnectedCallback()` - Element removed from DOM
- `attributeChangedCallback()` - Attribute changed

## Passing Data from Server to Client

Use `data-*` attributes to pass server-side values to client scripts:

### Basic Data Passing

```astro
---
// src/components/AstroGreet.astro
interface Props {
  message?: string;
}

const { message = 'Welcome, world!' } = Astro.props;
---

<!-- Store message prop as data attribute -->
<astro-greet data-message={message}>
  <button>Say hi!</button>
</astro-greet>

<script>
  class AstroGreet extends HTMLElement {
    connectedCallback() {
      // Read from data attribute
      const message = this.dataset.message;
      const button = this.querySelector('button');

      button?.addEventListener('click', () => {
        alert(message);
      });
    }
  }

  customElements.define('astro-greet', AstroGreet);
</script>
```

**Usage:**

```astro
<!-- Default message -->
<AstroGreet />

<!-- Custom messages -->
<AstroGreet message="Lovely day to build components!" />
<AstroGreet message="Glad you made it! 👋" />
```

### Complex Data (JSON)

```astro
---
interface Props {
  user: {
    name: string;
    email: string;
  };
}

const { user } = Astro.props;
---

<user-profile data-user={JSON.stringify(user)}>
  <div class="name"></div>
  <div class="email"></div>
</user-profile>

<script>
  class UserProfile extends HTMLElement {
    connectedCallback() {
      // Parse JSON data
      const user = JSON.parse(this.dataset.user || '{}');

      const nameEl = this.querySelector('.name');
      const emailEl = this.querySelector('.email');

      if (nameEl) nameEl.textContent = user.name;
      if (emailEl) emailEl.textContent = user.email;
    }
  }

  customElements.define('user-profile', UserProfile);
</script>
```

### Multiple Data Attributes

```astro
---
const { apiKey, endpoint, timeout = 5000 } = Astro.props;
---

<api-client
  data-api-key={apiKey}
  data-endpoint={endpoint}
  data-timeout={timeout}
>
  <button>Fetch Data</button>
</api-client>

<script>
  class ApiClient extends HTMLElement {
    connectedCallback() {
      const { apiKey, endpoint, timeout } = this.dataset;

      const button = this.querySelector('button');

      button?.addEventListener('click', async () => {
        try {
          const response = await fetch(endpoint!, {
            headers: { 'Authorization': `Bearer ${apiKey}` },
            signal: AbortSignal.timeout(Number(timeout)),
          });

          const data = await response.json();
          console.log(data);
        } catch (error) {
          console.error('Error:', error);
        }
      });
    }
  }

  customElements.define('api-client', ApiClient);
</script>
```

## Combining with UI Frameworks

If using UI framework components, scripts may run before framework elements are ready.

### Problem

```astro
---
import ReactButton from './ReactButton.jsx';
---

<ReactButton client:load />

<script>
  // ❌ May not find React button yet
  const button = document.querySelector('.react-button');
</script>
```

### Solution: Use Custom Elements

```astro
---
import ReactButton from './ReactButton.jsx';
---

<interactive-wrapper>
  <ReactButton client:load />
</interactive-wrapper>

<script>
  class InteractiveWrapper extends HTMLElement {
    connectedCallback() {
      // ✅ Runs when element is ready
      // Wait a bit for React to hydrate
      setTimeout(() => {
        const button = this.querySelector('.react-button');
        // Now safe to interact with button
      }, 100);
    }
  }

  customElements.define('interactive-wrapper', InteractiveWrapper);
</script>
```

## Common Patterns

### Pattern 1: Toggle Visibility

```astro
<button data-toggle="menu">Menu</button>
<nav id="menu" style="display: none;">
  <a href="/">Home</a>
  <a href="/about">About</a>
</nav>

<script>
  const toggle = document.querySelector('[data-toggle="menu"]');
  const menu = document.querySelector('#menu');

  toggle?.addEventListener('click', () => {
    if (menu) {
      const isHidden = menu.style.display === 'none';
      menu.style.display = isHidden ? 'block' : 'none';
    }
  });
</script>
```

### Pattern 2: Dark Mode Toggle

```astro
<button id="theme-toggle">Toggle Theme</button>

<script>
  const toggle = document.querySelector('#theme-toggle');

  toggle?.addEventListener('click', () => {
    const isDark = document.documentElement.classList.toggle('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
  });

  // Set initial theme from localStorage
  const theme = localStorage.getItem('theme');
  if (theme === 'dark') {
    document.documentElement.classList.add('dark');
  }
</script>
```

### Pattern 3: Smooth Scroll

```astro
<nav>
  <a href="#section1" data-smooth>Section 1</a>
  <a href="#section2" data-smooth>Section 2</a>
</nav>

<script>
  const links = document.querySelectorAll('a[data-smooth]');

  links.forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();

      const href = link.getAttribute('href');
      if (!href) return;

      const target = document.querySelector(href);
      target?.scrollIntoView({ behavior: 'smooth' });
    });
  });
</script>
```

### Pattern 4: Form Validation

```astro
<form id="signup-form">
  <input name="email" type="email" required />
  <div class="error" style="display: none;"></div>
  <button type="submit">Sign Up</button>
</form>

<script>
  const form = document.querySelector('#signup-form');
  const errorDiv = form?.querySelector('.error');

  form?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(form as HTMLFormElement);
    const email = formData.get('email');

    // Validation
    if (!email || !/\S+@\S+\.\S+/.test(email as string)) {
      if (errorDiv) {
        errorDiv.textContent = 'Please enter a valid email';
        errorDiv.style.display = 'block';
      }
      return;
    }

    // Submit
    try {
      const response = await fetch('/api/signup', {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        window.location.href = '/success';
      }
    } catch (error) {
      if (errorDiv) {
        errorDiv.textContent = 'An error occurred';
        errorDiv.style.display = 'block';
      }
    }
  });
</script>
```

### Pattern 5: Analytics Tracking

```astro
<button data-track="cta-click" data-label="Sign Up">
  Sign Up
</button>

<script>
  const trackedElements = document.querySelectorAll('[data-track]');

  trackedElements.forEach(element => {
    element.addEventListener('click', () => {
      const event = element.getAttribute('data-track');
      const label = element.getAttribute('data-label');

      // Send to analytics
      if (typeof gtag !== 'undefined') {
        gtag('event', event, {
          'event_label': label,
        });
      }
    });
  });
</script>
```

## Best Practices for AI Agents

### 1. Use querySelectorAll for Multiple Instances

```astro
<!-- ✅ Good: Handles multiple instances -->
<script>
  const buttons = document.querySelectorAll('.alert-button');
  buttons.forEach(button => {
    button.addEventListener('click', () => alert('Hello!'));
  });
</script>

<!-- ❌ Avoid: querySelector only gets first instance -->
<script>
  const button = document.querySelector('.alert-button');
  button?.addEventListener('click', () => alert('Hello!'));
</script>
```

### 2. Prefer Web Components for Reusable Logic

```astro
<!-- ✅ Good: Web component (reusable, scoped) -->
<astro-counter>
  <button>Count: 0</button>
</astro-counter>

<script>
  class AstroCounter extends HTMLElement {
    connectedCallback() {
      let count = 0;
      const button = this.querySelector('button');
      button?.addEventListener('click', () => {
        count++;
        button.textContent = `Count: ${count}`;
      });
    }
  }
  customElements.define('astro-counter', AstroCounter);
</script>

<!-- ❌ Avoid: Global script for component-specific logic -->
<button id="counter">Count: 0</button>
<script>
  let count = 0;
  document.querySelector('#counter')?.addEventListener('click', ...);
</script>
```

### 3. Use data-* Attributes for Server-to-Client Data

```astro
<!-- ✅ Good: data-* attributes -->
<my-component data-config={JSON.stringify(config)}>

<script>
  class MyComponent extends HTMLElement {
    connectedCallback() {
      const config = JSON.parse(this.dataset.config || '{}');
    }
  }
</script>

<!-- ❌ Avoid: Global variables -->
<script is:inline>
  window.config = {/* ... */};
</script>
```

### 4. Add TypeScript Types

```astro
<script>
  // ✅ Good: Typed
  const button = document.querySelector<HTMLButtonElement>('.my-button');

  if (button) {
    button.disabled = true;  // TypeScript knows this is valid
  }

  // ❌ Avoid: Untyped
  const button = document.querySelector('.my-button');
  button.disabled = true;  // Might be null!
</script>
```

### 5. Clean Up Event Listeners

```astro
<script>
  // ✅ Good: Cleanup in disconnectedCallback
  class MyComponent extends HTMLElement {
    private controller = new AbortController();

    connectedCallback() {
      document.addEventListener('click', this.handler, {
        signal: this.controller.signal
      });
    }

    disconnectedCallback() {
      this.controller.abort(); // Removes listener
    }

    private handler = () => { /* ... */ };
  }

  // ❌ Avoid: No cleanup (memory leak)
  class BadComponent extends HTMLElement {
    connectedCallback() {
      document.addEventListener('click', () => {});
      // Never removed!
    }
  }
</script>
```

## Quick Reference

```astro
<!-- Basic script -->
<script>
  // Processed: TypeScript, imports, bundled
  console.log('Hello');
</script>

<!-- Inline script (unprocessed) -->
<script is:inline>
  // Not processed, rendered as-is
  console.log('Raw');
</script>

<!-- External script (processed) -->
<script src="../scripts/local.ts"></script>

<!-- External script (unprocessed) -->
<script is:inline src="/public-script.js"></script>

<!-- Event handling -->
<button class="my-btn">Click</button>
<script>
  const buttons = document.querySelectorAll('.my-btn');
  buttons.forEach(btn => {
    btn.addEventListener('click', () => alert('Clicked!'));
  });
</script>

<!-- Web component -->
<my-component data-value="hello">
  <button>Click</button>
</my-component>

<script>
  class MyComponent extends HTMLElement {
    connectedCallback() {
      const value = this.dataset.value;
      const button = this.querySelector('button');
      button?.addEventListener('click', () => alert(value));
    }
  }
  customElements.define('my-component', MyComponent);
</script>
```

## Resources

- [MDN: Custom Elements](https://developer.mozilla.org/en-US/docs/Web/Web_Components/Using_custom_elements)
- [web.dev: Reusable Web Components](https://web.dev/custom-elements-v1/)
- [Astro Script Processing](https://docs.astro.build/en/guides/client-side-scripts/)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent guide for scripts and event handling in Astro
**See Also**: 06-framework-components.md for framework-based interactivity
