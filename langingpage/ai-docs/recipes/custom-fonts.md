# Using Custom Fonts in Astro

This guide shows you how to add web fonts to your Astro project and use them in your components.

## Overview

There are three main approaches to adding custom fonts:

1. **Local font files** - Host fonts in your project
2. **Fontsource** - NPM packages for open-source fonts
3. **Tailwind integration** - Register fonts with Tailwind CSS

## Method 1: Local Font Files

Host custom font files directly in your project.

### Step 1: Add Font Files

Place your font files in the `public/fonts/` directory:

```
public/
└── fonts/
    ├── DistantGalaxy.woff
    ├── DistantGalaxy.woff2
    └── CustomFont.ttf
```

**Recommended formats:**
- `.woff2` - Best compression, modern browsers
- `.woff` - Fallback for older browsers
- `.ttf` / `.otf` - Fallback for very old browsers

### Step 2: Register Font with @font-face

Add a `@font-face` declaration in one of these locations:

**Option A: Global CSS file**

```css
/* src/styles/global.css */
@font-face {
  font-family: 'DistantGalaxy';
  src: url('/fonts/DistantGalaxy.woff2') format('woff2'),
       url('/fonts/DistantGalaxy.woff') format('woff');
  font-weight: normal;
  font-style: normal;
  font-display: swap;
}
```

**Option B: Scoped style block**

```astro
---
// src/layouts/BaseLayout.astro
---

<!DOCTYPE html>
<html>
  <head>
    <!-- ... -->
  </head>
  <body>
    <slot />
  </body>
</html>

<style is:global>
  @font-face {
    font-family: 'DistantGalaxy';
    src: url('/fonts/DistantGalaxy.woff2') format('woff2'),
         url('/fonts/DistantGalaxy.woff') format('woff');
    font-weight: normal;
    font-style: normal;
    font-display: swap;
  }
</style>
```

**Option C: Component-specific**

```astro
---
// src/components/Hero.astro
---

<h1>In a galaxy far, far away...</h1>

<style>
  @font-face {
    font-family: 'DistantGalaxy';
    src: url('/fonts/DistantGalaxy.woff2') format('woff2');
    font-weight: normal;
    font-style: normal;
    font-display: swap;
  }

  h1 {
    font-family: 'DistantGalaxy', sans-serif;
  }
</style>
```

### Step 3: Use the Font

Use the `font-family` name from your `@font-face` declaration:

```astro
---
// src/pages/example.astro
---

<h1>In a galaxy far, far away...</h1>
<p>Custom fonts make my headings much cooler!</p>

<style>
  h1 {
    font-family: 'DistantGalaxy', sans-serif;
  }

  p {
    font-family: system-ui, sans-serif;
  }
</style>
```

### Multiple Font Weights

For fonts with multiple weights:

```css
/* Light weight */
@font-face {
  font-family: 'CustomFont';
  src: url('/fonts/CustomFont-Light.woff2') format('woff2');
  font-weight: 300;
  font-style: normal;
  font-display: swap;
}

/* Regular weight */
@font-face {
  font-family: 'CustomFont';
  src: url('/fonts/CustomFont-Regular.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}

/* Bold weight */
@font-face {
  font-family: 'CustomFont';
  src: url('/fonts/CustomFont-Bold.woff2') format('woff2');
  font-weight: 700;
  font-style: normal;
  font-display: swap;
}
```

**Usage:**

```css
h1 {
  font-family: 'CustomFont', sans-serif;
  font-weight: 700; /* Uses Bold variant */
}

p {
  font-family: 'CustomFont', sans-serif;
  font-weight: 300; /* Uses Light variant */
}
```

## Method 2: Fontsource

Fontsource provides NPM packages for Google Fonts and other open-source fonts.

### Step 1: Find Your Font

Browse the [Fontsource catalog](https://fontsource.org/) and find your font.

**Examples:**
- `@fontsource/inter`
- `@fontsource/roboto`
- `@fontsource/twinkle-star`

### Step 2: Install the Package

```bash
# pnpm
pnpm add @fontsource/twinkle-star

# npm
npm install @fontsource/twinkle-star

# yarn
yarn add @fontsource/twinkle-star

# bun
bun add @fontsource/twinkle-star
```

**Tip:** Find the exact package name in the "Quick Installation" section on each font's Fontsource page. Package names start with `@fontsource/` or `@fontsource-variable/`.

### Step 3: Import in Your Layout

Import the font package in a common layout to make it available site-wide:

```astro
---
// src/layouts/BaseLayout.astro
import '@fontsource/twinkle-star';

const { title } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    <slot />
  </body>
</html>
```

**What this does:**
- Automatically adds `@font-face` rules
- Sets up all necessary font files
- No manual CSS configuration needed

### Step 4: Use the Font

Use the font name as shown on its Fontsource page:

```css
h1 {
  font-family: 'Twinkle Star', cursive;
}

p {
  font-family: 'Inter', sans-serif;
}
```

### Importing Specific Weights

Import only the weights you need for better performance:

```astro
---
// Import only 400 and 700 weights
import '@fontsource/inter/400.css';
import '@fontsource/inter/700.css';
---
```

**Default import vs specific:**

```astro
---
// Imports all weights (larger bundle)
import '@fontsource/inter';

// Imports only specific weights (smaller bundle)
import '@fontsource/inter/400.css';
import '@fontsource/inter/700.css';
---
```

### Variable Fonts

For variable fonts, use the `-variable` package:

```bash
pnpm add @fontsource-variable/inter
```

```astro
---
import '@fontsource-variable/inter';
---

<style>
  body {
    font-family: 'InterVariable', sans-serif;
    font-variation-settings: 'wght' 450; /* Custom weight */
  }
</style>
```

### Preloading Fonts

For critical fonts, preload them to improve rendering times:

```astro
---
// src/layouts/BaseLayout.astro
import '@fontsource/inter/400.css';
import '@fontsource/inter/700.css';
---

<!DOCTYPE html>
<html>
  <head>
    <link
      rel="preload"
      href="/node_modules/@fontsource/inter/files/inter-latin-400-normal.woff2"
      as="font"
      type="font/woff2"
      crossorigin
    />
  </head>
  <body>
    <slot />
  </body>
</html>
```

**See:** [Fontsource guide to preloading fonts](https://fontsource.org/docs/getting-started/preload)

## Method 3: Fonts with Tailwind CSS

Register fonts with Tailwind for use with utility classes.

### Step 1: Install Font

Use either local fonts or Fontsource (skip the final CSS step):

```bash
# Option A: Fontsource
pnpm add @fontsource/inter

# Option B: Local fonts
# Place in public/fonts/
```

### Step 2: Register @font-face

**For local fonts:**

```css
/* src/styles/global.css */
@font-face {
  font-family: 'Inter';
  src: url('/fonts/Inter-Regular.woff2') format('woff2');
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
```

**For Fontsource:**

```astro
---
// src/layouts/BaseLayout.astro
import '@fontsource/inter';
---
```

### Step 3: Configure Tailwind Theme

Add the font to your Tailwind configuration:

```css
/* src/styles/global.css */
@import 'tailwindcss';

@theme {
  --font-sans: 'Inter', 'sans-serif';
  --font-display: 'Playfair Display', 'serif';
}
```

**Or in config file (older method):**

```javascript
// tailwind.config.mjs
export default {
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        display: ['Playfair Display', 'serif'],
      },
    },
  },
}
```

### Step 4: Use with Tailwind Classes

```astro
<h1 class="font-sans">Uses Inter font</h1>
<h2 class="font-display">Uses Playfair Display font</h2>
```

**Custom font utilities:**

```css
@theme {
  --font-sans: 'Inter', 'sans-serif';
  --font-serif: 'Merriweather', 'serif';
  --font-mono: 'Fira Code', 'monospace';
  --font-display: 'Playfair Display', 'serif';
}
```

```astro
<p class="font-sans">Body text</p>
<code class="font-mono">Code snippet</code>
<h1 class="font-display">Fancy heading</h1>
```

## Font Display Strategies

The `font-display` property controls how fonts are displayed while loading:

```css
@font-face {
  font-family: 'MyFont';
  src: url('/fonts/MyFont.woff2') format('woff2');
  font-display: swap; /* Recommended */
}
```

| Value | Behavior | Use Case |
|-------|----------|----------|
| `swap` | Show fallback immediately, swap when loaded | **Recommended**: Best UX |
| `block` | Hide text briefly, then show font | Critical branding fonts |
| `fallback` | Brief block, then fallback if not loaded | Balance performance/design |
| `optional` | Use font if cached, else fallback | Fastest, least important fonts |
| `auto` | Browser decides | Default behavior |

**Recommendation:** Use `swap` for best user experience.

## Performance Best Practices

### 1. Use Modern Formats

```css
@font-face {
  font-family: 'MyFont';
  /* Prefer woff2 (best compression) */
  src: url('/fonts/MyFont.woff2') format('woff2'),
       url('/fonts/MyFont.woff') format('woff'); /* Fallback */
  font-display: swap;
}
```

### 2. Subset Fonts

Only include characters you need:

```css
@font-face {
  font-family: 'MyFont';
  src: url('/fonts/MyFont-Latin.woff2') format('woff2');
  unicode-range: U+0000-00FF, U+0131, U+0152-0153; /* Latin only */
  font-display: swap;
}
```

**Tools:**
- [Font Squirrel Webfont Generator](https://www.fontsquirrel.com/tools/webfont-generator)
- [Glyphhanger](https://github.com/zachleat/glyphhanger)

### 3. Preload Critical Fonts

```astro
<head>
  <!-- Preload the most important font -->
  <link
    rel="preload"
    href="/fonts/MyFont-Regular.woff2"
    as="font"
    type="font/woff2"
    crossorigin
  />
</head>
```

**Only preload:**
- Fonts used above the fold
- Primary body text font
- Main heading font

### 4. Self-Host Google Fonts

Instead of using Google CDN:

```astro
---
// Use Fontsource instead of Google CDN
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/700.css';
---
```

**Benefits:**
- Better performance
- GDPR compliance
- No external requests
- Better caching control

## Complete Examples

### Example 1: Local Font with Multiple Weights

```astro
---
// src/layouts/BaseLayout.astro
const { title } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>
  </head>
  <body>
    <slot />
  </body>
</html>

<style is:global>
  /* Light weight */
  @font-face {
    font-family: 'CustomFont';
    src: url('/fonts/CustomFont-Light.woff2') format('woff2');
    font-weight: 300;
    font-display: swap;
  }

  /* Regular weight */
  @font-face {
    font-family: 'CustomFont';
    src: url('/fonts/CustomFont-Regular.woff2') format('woff2');
    font-weight: 400;
    font-display: swap;
  }

  /* Bold weight */
  @font-face {
    font-family: 'CustomFont';
    src: url('/fonts/CustomFont-Bold.woff2') format('woff2');
    font-weight: 700;
    font-display: swap;
  }

  body {
    font-family: 'CustomFont', system-ui, sans-serif;
  }
</style>
```

### Example 2: Fontsource with Preload

```astro
---
// src/layouts/BaseLayout.astro
import '@fontsource/inter/400.css';
import '@fontsource/inter/700.css';

const { title } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>{title}</title>

    <!-- Preload critical font -->
    <link
      rel="preload"
      href="/node_modules/@fontsource/inter/files/inter-latin-400-normal.woff2"
      as="font"
      type="font/woff2"
      crossorigin
    />
  </head>
  <body>
    <slot />
  </body>
</html>

<style is:global>
  body {
    font-family: 'Inter', system-ui, sans-serif;
  }
</style>
```

### Example 3: Tailwind with Multiple Fonts

```astro
---
// src/layouts/BaseLayout.astro
import '@fontsource/inter/400.css';
import '@fontsource/inter/700.css';
import '@fontsource/playfair-display/700.css';
---

<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
  </head>
  <body>
    <slot />
  </body>
</html>
```

```css
/* src/styles/global.css */
@import 'tailwindcss';

@theme {
  --font-sans: 'Inter', 'system-ui', 'sans-serif';
  --font-display: 'Playfair Display', 'Georgia', 'serif';
}
```

```astro
<!-- Usage -->
<h1 class="font-display text-4xl">Elegant Heading</h1>
<p class="font-sans">Body text with Inter</p>
```

## Troubleshooting

### Font Not Loading

**Check:**
1. File path is correct (`/fonts/...` not `../fonts/...`)
2. Font files are in `public/` directory
3. `@font-face` is in global scope or imported
4. Font name matches exactly (case-sensitive)

### CORS Errors

**Solution:** Fonts must be on same domain or have CORS headers.

```javascript
// astro.config.mjs
export default defineConfig({
  vite: {
    server: {
      headers: {
        'Access-Control-Allow-Origin': '*',
      },
    },
  },
});
```

### Font Flashing (FOUT)

**Solution:** Use `font-display: swap` and preload:

```css
@font-face {
  font-family: 'MyFont';
  src: url('/fonts/MyFont.woff2') format('woff2');
  font-display: swap; /* Prevents invisible text */
}
```

### Large Bundle Size

**Solutions:**
1. Import only needed weights
2. Use variable fonts
3. Subset fonts to needed characters
4. Use `woff2` format (best compression)

## Best Practices for AI Agents

### 1. Prefer Fontsource for Open-Source Fonts

```astro
<!-- ✅ Good: Fontsource (automatic setup) -->
---
import '@fontsource/inter/400.css';
---

<!-- ❌ Avoid: Manual setup for Google Fonts -->
<link href="https://fonts.googleapis.com/..." />
```

### 2. Use font-display: swap

```css
/* ✅ Good: swap for best UX */
@font-face {
  font-family: 'MyFont';
  src: url('/fonts/MyFont.woff2') format('woff2');
  font-display: swap;
}

/* ❌ Avoid: block (causes invisible text) */
@font-face {
  font-family: 'MyFont';
  src: url('/fonts/MyFont.woff2') format('woff2');
  font-display: block;
}
```

### 3. Import in Layout for Site-Wide Availability

```astro
<!-- ✅ Good: Import in base layout -->
---
// src/layouts/BaseLayout.astro
import '@fontsource/inter';
---

<!-- ❌ Avoid: Import in every component -->
---
// src/components/Header.astro
import '@fontsource/inter'; // Redundant!
---
```

### 4. Preload Only Critical Fonts

```astro
<!-- ✅ Good: Preload only main font -->
<link rel="preload" href="/fonts/MainFont.woff2" as="font" />

<!-- ❌ Avoid: Preloading too many fonts -->
<link rel="preload" href="/fonts/Font1.woff2" as="font" />
<link rel="preload" href="/fonts/Font2.woff2" as="font" />
<link rel="preload" href="/fonts/Font3.woff2" as="font" />
```

### 5. Provide Fallback Fonts

```css
/* ✅ Good: Fallback stack */
h1 {
  font-family: 'CustomFont', system-ui, sans-serif;
}

/* ❌ Avoid: No fallback */
h1 {
  font-family: 'CustomFont';
}
```

## Resources

- [MDN Web Fonts Guide](https://developer.mozilla.org/en-US/docs/Learn/CSS/Styling_text/Web_fonts)
- [Fontsource Documentation](https://fontsource.org/docs)
- [Font Squirrel Webfont Generator](https://www.fontsquirrel.com/tools/webfont-generator)
- [Tailwind Custom Fonts](https://tailwindcss.com/docs/font-family#customizing-your-theme)
- [Web Font Loading Patterns](https://www.zachleat.com/web/comprehensive-webfonts/)

---

**Last Updated**: 2025-11-17
**Purpose**: AI agent guide for using custom fonts in Astro
**See Also**: recipes/tailwind-typography.md for styling Markdown content
