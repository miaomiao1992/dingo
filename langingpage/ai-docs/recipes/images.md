# Images in Astro

## Overview

Astro provides powerful built-in image optimization through the `astro:assets` module. The `<Image />` and `<Picture />` components automatically optimize images, generate responsive variants, and maintain aspect ratios.

## Key Concepts

### Image Storage Locations

**1. `src/` Directory (Recommended)**
- Images stored in `src/` can be imported and optimized by Astro
- Enables automatic optimization, resizing, and format conversion
- Required for using `<Image />` and `<Picture />` components
- Best for images that are part of your site's design

```
src/
├── assets/
│   └── images/
│       ├── hero.png
│       └── logo.svg
└── content/
    └── posts/
        └── images/
            └── post-hero.jpg
```

**2. `public/` Directory**
- Images in `public/` are served as-is with no optimization
- Use for images that should not be processed (favicons, robots.txt, etc.)
- Referenced by absolute path: `/favicon.ico`
- No build-time optimization or transformations

```
public/
├── favicon.ico
├── robots.txt
└── social-card.png
```

### When to Use Each Location

**Use `src/` for:**
- Hero images, banners, thumbnails
- Blog post images
- Product photos
- Any image that benefits from optimization

**Use `public/` for:**
- Favicons and manifest icons
- Static assets referenced by external tools
- Images that must maintain exact file names/paths
- Pre-optimized images you don't want Astro to process

## Image Components

### The `<Image />` Component

Generates a single optimized `<img>` element with automatic optimization.

**Basic Usage:**

```astro
---
import { Image } from 'astro:assets';
import myImage from '../assets/hero.png';
---

<Image src={myImage} alt="Hero image" />
```

**With Dimensions:**

```astro
<Image
  src={myImage}
  alt="Hero image"
  width={800}
  height={600}
/>
```

**Props:**
- `src` (required): Image source (import or path)
- `alt` (required): Alt text for accessibility
- `width`: Target width in pixels
- `height`: Target height in pixels
- `format`: Output format ('webp', 'avif', 'png', 'jpg', 'svg')
- `quality`: Quality 1-100 (default: 80)
- `densities`: Array like `[1, 2]` for Retina displays
- `widths`: Array of widths for responsive images

### The `<Picture />` Component

Generates a `<picture>` element with multiple formats and sizes for better browser support.

**Basic Usage:**

```astro
---
import { Picture } from 'astro:assets';
import myImage from '../assets/hero.png';
---

<Picture
  src={myImage}
  formats={['avif', 'webp']}
  alt="Hero image"
/>
```

**Responsive with Multiple Widths:**

```astro
<Picture
  src={myImage}
  widths={[400, 800, 1200]}
  sizes="(max-width: 600px) 400px, (max-width: 900px) 800px, 1200px"
  formats={['avif', 'webp', 'jpg']}
  alt="Responsive hero image"
/>
```

**Generated HTML:**

```html
<picture>
  <source type="image/avif" srcset="/_astro/hero.hash.avif 400w, ..." sizes="...">
  <source type="image/webp" srcset="/_astro/hero.hash.webp 400w, ..." sizes="...">
  <img src="/_astro/hero.hash.jpg" alt="Responsive hero image" loading="lazy">
</picture>
```

## Images in Different File Types

### In `.astro` Components

**Import and Use:**

```astro
---
import { Image } from 'astro:assets';
import localImage from '../assets/hero.png';
---

<Image src={localImage} alt="Local hero" />
```

**Public Folder (No Optimization):**

```astro
<img src="/social-card.png" alt="Social card" />
```

### In Markdown Files

**Relative Paths:**

```markdown
![Alt text](./local-image.png)
```

**Absolute Paths from `public/`:**

```markdown
![Alt text](/public-image.png)
```

**Note:** Standard Markdown syntax doesn't support `<Image />` component. Use MDX for that.

### In MDX Files

**Import and Use Components:**

```mdx
---
title: My Post
---
import { Image } from 'astro:assets';
import rocket from '../assets/rocket.png';

# Welcome

<Image src={rocket} alt="Rocket" width={600} />

![Standard Markdown syntax also works](./local-image.png)
```

### In UI Framework Components

**React Example:**

```tsx
import type { ImageMetadata } from 'astro';
import { Image } from 'astro:assets';
import rocket from '../assets/rocket.png';

interface Props {
  imageSrc: ImageMetadata;
}

export default function MyComponent({ imageSrc }: Props) {
  return (
    <Image src={imageSrc} alt="Rocket" />
  );
}
```

**Usage in Astro:**

```astro
---
import MyComponent from '../components/MyComponent';
import rocket from '../assets/rocket.png';
---

<MyComponent imageSrc={rocket} client:load />
```

## SVG Images

SVGs can be used in multiple ways:

### 1. Import as Component (Recommended)

```astro
---
import Logo from '../assets/logo.svg?component';
---

<Logo />
```

Benefits:
- Can apply CSS classes and styles
- Inline SVG for instant rendering
- Access to SVG DOM for animations

### 2. Import as Source Path

```astro
---
import { Image } from 'astro:assets';
import logoSrc from '../assets/logo.svg';
---

<Image src={logoSrc} alt="Logo" />
```

### 3. Direct Use in Public Folder

```astro
<img src="/logo.svg" alt="Logo" />
```

## Remote Images

### Basic Remote Image Usage

```astro
---
import { Image } from 'astro:assets';
---

<Image
  src="https://example.com/remote-image.jpg"
  alt="Remote image"
  width={800}
  height={600}
  inferSize
/>
```

**Important:**
- Must specify `width` and `height`, or use `inferSize` prop
- `inferSize` fetches image at build time to determine dimensions

### Authorized Remote Images

For images behind authentication (e.g., private CDN):

**1. Configure `image.domains` or `image.remotePatterns`:**

```javascript
// astro.config.mjs
export default defineConfig({
  image: {
    domains: ['cdn.example.com'],
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.example.com',
      },
    ],
  },
});
```

**2. Use Custom Fetch for Authorization:**

```astro
---
import { Image } from 'astro:assets';

// Custom image service
const getAuthorizedImage = async (url: string) => {
  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${import.meta.env.CDN_TOKEN}`,
    },
  });

  const blob = await response.blob();
  return URL.createObjectURL(blob);
};

const imageSrc = await getAuthorizedImage('https://private-cdn.example.com/image.jpg');
---

<Image src={imageSrc} alt="Authorized image" inferSize />
```

## Responsive Images

### Using `widths` and `sizes`

```astro
---
import { Picture } from 'astro:assets';
import hero from '../assets/hero.png';
---

<Picture
  src={hero}
  widths={[400, 800, 1200, 1600]}
  sizes="(max-width: 400px) 400px, (max-width: 800px) 800px, (max-width: 1200px) 1200px, 1600px"
  alt="Responsive hero"
/>
```

### Using `densities` for Retina

```astro
<Image
  src={hero}
  width={800}
  densities={[1, 2]}
  alt="Retina-ready image"
/>
```

Generates:
```html
<img
  srcset="/_astro/hero.hash.jpg 1x, /_astro/hero@2x.hash.jpg 2x"
  src="/_astro/hero.hash.jpg"
  alt="Retina-ready image"
>
```

## Images in Content Collections

### Schema with Image Helper

```typescript
// src/content/config.ts
import { defineCollection, z } from 'astro:content';
import { image } from 'astro:assets';

const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    coverImage: image(), // Image helper
    gallery: z.array(image()).optional(),
  }),
});

export const collections = { blog };
```

### Using Images from Collection Entries

```astro
---
import { getCollection } from 'astro:content';
import { Image } from 'astro:assets';

const posts = await getCollection('blog');
---

{posts.map((post) => (
  <article>
    <Image
      src={post.data.coverImage}
      alt={post.data.title}
      width={600}
    />
    <h2>{post.data.title}</h2>
  </article>
))}
```

### Frontmatter Example

```markdown
---
title: My Blog Post
coverImage: ./cover.jpg
gallery:
  - ./image1.jpg
  - ./image2.jpg
  - ./image3.jpg
---

Post content here...
```

## Image Formats

### Supported Formats

**Input Formats:**
- JPEG, PNG, GIF
- WebP, AVIF
- TIFF
- SVG (pass-through or component)

**Output Formats:**
- `jpg` - Good compatibility, larger file size
- `png` - Transparency support, larger file size
- `webp` - Modern format, good compression, wide support
- `avif` - Best compression, growing support
- `svg` - Vector graphics, infinite scaling

### Format Conversion

```astro
---
import { Image } from 'astro:assets';
import pngImage from '../assets/hero.png';
---

<!-- Convert PNG to WebP -->
<Image
  src={pngImage}
  format="webp"
  alt="Converted to WebP"
/>

<!-- Multiple formats with Picture -->
<Picture
  src={pngImage}
  formats={['avif', 'webp', 'jpg']}
  alt="Multi-format image"
/>
```

## Image Quality and Optimization

### Quality Settings

```astro
<Image
  src={myImage}
  quality={90}  // 1-100, default is 80
  alt="High quality image"
/>
```

**Quality Guidelines:**
- `60-70`: Thumbnails, small previews
- `80`: Default, good balance
- `90-100`: Hero images, important visuals

### Performance Optimization

```astro
---
import { Image } from 'astro:assets';
import hero from '../assets/hero.jpg';
---

<!-- Lazy loading (default) -->
<Image src={hero} alt="Hero" />

<!-- Eager loading for above-the-fold images -->
<Image src={hero} alt="Hero" loading="eager" />

<!-- Decode async for better rendering -->
<Image src={hero} alt="Hero" decoding="async" />
```

## TypeScript Support

### Image Metadata Type

```typescript
import type { ImageMetadata } from 'astro';

interface Props {
  image: ImageMetadata;
  alt: string;
}

const { image, alt } = Astro.props;
```

### Props Interface with Images

```typescript
interface BlogPost {
  title: string;
  coverImage: ImageMetadata;
  thumbnails: ImageMetadata[];
}
```

## Best Practices

### 1. Always Provide Alt Text

```astro
<!-- Good -->
<Image src={hero} alt="Team celebrating project launch" />

<!-- Bad -->
<Image src={hero} alt="" />
```

### 2. Use Appropriate Formats

```astro
<!-- Photos: WebP/AVIF -->
<Picture
  src={photo}
  formats={['avif', 'webp', 'jpg']}
  alt="Photo"
/>

<!-- Logos: SVG when possible -->
<Logo />  <!-- SVG component -->

<!-- Icons: SVG or PNG with transparency -->
<Image src={icon} format="webp" alt="Icon" />
```

### 3. Optimize for Responsive Design

```astro
<!-- Mobile-first sizing -->
<Picture
  src={hero}
  widths={[375, 768, 1024, 1920]}
  sizes="(max-width: 375px) 375px, (max-width: 768px) 768px, (max-width: 1024px) 1024px, 1920px"
  alt="Responsive hero"
/>
```

### 4. Lazy Load Below-the-Fold Images

```astro
<!-- Above the fold: eager -->
<Image src={hero} alt="Hero" loading="eager" />

<!-- Below the fold: lazy (default) -->
<Image src={feature} alt="Feature" />
<Image src={gallery1} alt="Gallery 1" />
<Image src={gallery2} alt="Gallery 2" />
```

### 5. Store Images Logically

```
src/
├── assets/
│   ├── images/
│   │   ├── common/        # Shared images (logos, icons)
│   │   ├── hero/          # Hero/banner images
│   │   └── ui/            # UI elements
│   └── icons/             # SVG icons
└── content/
    └── blog/
        └── post-name/
            ├── index.md
            └── images/    # Post-specific images
```

### 6. Use Image Helper in Content Collections

```typescript
// Always use image() helper for type safety
const blog = defineCollection({
  schema: z.object({
    coverImage: image(),  // ✓ Type-safe
    // coverImage: z.string(),  // ✗ No validation
  }),
});
```

## Common Patterns

### Hero Image Component

```astro
---
// components/Hero.astro
import { Picture } from 'astro:assets';
import type { ImageMetadata } from 'astro';

interface Props {
  image: ImageMetadata;
  alt: string;
  title: string;
}

const { image, alt, title } = Astro.props;
---

<section class="hero">
  <Picture
    src={image}
    widths={[375, 768, 1024, 1920]}
    sizes="100vw"
    formats={['avif', 'webp']}
    alt={alt}
    loading="eager"
  />
  <div class="hero-content">
    <h1>{title}</h1>
  </div>
</section>

<style>
  .hero {
    position: relative;
    width: 100%;
    height: 60vh;
  }

  .hero img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .hero-content {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: white;
    text-align: center;
  }
</style>
```

### Image Gallery Component

```astro
---
// components/Gallery.astro
import { Image } from 'astro:assets';
import type { ImageMetadata } from 'astro';

interface Props {
  images: ImageMetadata[];
}

const { images } = Astro.props;
---

<div class="gallery">
  {images.map((image) => (
    <div class="gallery-item">
      <Image
        src={image}
        width={400}
        height={300}
        alt="Gallery image"
        loading="lazy"
      />
    </div>
  ))}
</div>

<style>
  .gallery {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
  }

  .gallery-item img {
    width: 100%;
    height: auto;
    object-fit: cover;
    border-radius: 0.5rem;
  }
</style>
```

### Avatar Component with Fallback

```astro
---
// components/Avatar.astro
import { Image } from 'astro:assets';
import type { ImageMetadata } from 'astro';
import defaultAvatar from '../assets/default-avatar.png';

interface Props {
  avatar?: ImageMetadata;
  name: string;
  size?: number;
}

const { avatar = defaultAvatar, name, size = 48 } = Astro.props;
---

<div class="avatar">
  <Image
    src={avatar}
    width={size}
    height={size}
    alt={`${name}'s avatar`}
    loading="lazy"
  />
</div>

<style>
  .avatar img {
    border-radius: 50%;
    object-fit: cover;
  }
</style>
```

## Troubleshooting

### Issue: "Could not find image"

**Problem:** Import path is incorrect.

**Solution:**
```astro
<!-- Wrong -->
<Image src="./assets/hero.png" alt="Hero" />

<!-- Right -->
---
import hero from '../assets/hero.png';
---
<Image src={hero} alt="Hero" />
```

### Issue: Build Fails with Large Images

**Problem:** Images too large for optimization.

**Solution:**
1. Resize images before adding to project
2. Use external image service for very large images
3. Increase Node memory: `NODE_OPTIONS=--max-old-space-size=4096 npm run build`

### Issue: Remote Images Not Loading

**Problem:** Domain not authorized.

**Solution:**
```javascript
// astro.config.mjs
export default defineConfig({
  image: {
    domains: ['images.unsplash.com', 'cdn.example.com'],
  },
});
```

### Issue: Images in Content Collections Not Working

**Problem:** Not using `image()` helper in schema.

**Solution:**
```typescript
// Wrong
schema: z.object({
  cover: z.string(),
})

// Right
import { image } from 'astro:assets';

schema: z.object({
  cover: image(),
})
```

## Quick Reference

### Import Image

```astro
---
import { Image, Picture } from 'astro:assets';
import myImage from '../assets/image.png';
---
```

### Basic Image

```astro
<Image src={myImage} alt="Description" />
```

### Picture with Formats

```astro
<Picture
  src={myImage}
  formats={['avif', 'webp']}
  alt="Description"
/>
```

### Responsive Image

```astro
<Picture
  src={myImage}
  widths={[400, 800, 1200]}
  sizes="(max-width: 600px) 400px, (max-width: 900px) 800px, 1200px"
  alt="Description"
/>
```

### Remote Image

```astro
<Image
  src="https://example.com/image.jpg"
  alt="Description"
  inferSize
/>
```

### SVG as Component

```astro
---
import Logo from '../assets/logo.svg?component';
---
<Logo />
```

### Content Collection Image

```typescript
import { image } from 'astro:assets';

defineCollection({
  schema: z.object({
    coverImage: image(),
  }),
});
```

## See Also

- [05-content-collections.md](../05-content-collections.md) - Using images in collections
- [07-astro-components.md](../07-astro-components.md) - Component patterns
- [09-markdown.md](../09-markdown.md) - Images in Markdown/MDX
- [recipes/custom-fonts.md](./custom-fonts.md) - Similar asset optimization patterns

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to image optimization and usage in Astro
**Key Concepts**: Image component, Picture component, responsive images, remote images, content collections
