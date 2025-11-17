# Data Fetching in Astro

## Overview

Astro provides powerful data fetching capabilities that work seamlessly at build time and runtime (with SSR). All `.astro` files and framework components have access to the global `fetch()` function for making HTTP requests to APIs, databases, and headless CMS platforms.

## Key Concepts

### When Data is Fetched

**Static (SSG) Mode:**
- Data fetched **once at build time**
- Results baked into static HTML
- No runtime data fetching on the client
- Fast, CDN-cacheable pages

**Server (SSR) Mode:**
- Data fetched **at request time** on the server
- Fresh data for each request
- Dynamic content based on user, time, etc.
- Requires server deployment

### Top-Level Await

Astro supports top-level `await` in component scripts - no need to wrap in async functions:

```astro
---
// ✓ This works! No async wrapper needed
const response = await fetch('https://api.example.com/data');
const data = await response.json();
---

<h1>{data.title}</h1>
```

## fetch() in Astro Components

### Basic Usage

**Simple GET Request:**

```astro
---
// src/components/User.astro
const response = await fetch('https://randomuser.me/api/');
const data = await response.json();
const randomUser = data.results[0];
---

<h1>User</h1>
<h2>{randomUser.name.first} {randomUser.name.last}</h2>
<p>Email: {randomUser.email}</p>
```

**With Error Handling:**

```astro
---
// src/components/Posts.astro
let posts = [];
let error = null;

try {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts');

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  posts = await response.json();
} catch (e) {
  error = e.message;
  console.error('Failed to fetch posts:', e);
}
---

{error ? (
  <p class="error">Failed to load posts: {error}</p>
) : (
  <ul>
    {posts.map(post => (
      <li>
        <h3>{post.title}</h3>
        <p>{post.body}</p>
      </li>
    ))}
  </ul>
)}
```

### Passing Fetched Data as Props

Data fetched in Astro components can be passed to child components:

```astro
---
// src/pages/user-profile.astro
import Contact from '../components/Contact.jsx';
import Location from '../components/Location.astro';

const response = await fetch('https://randomuser.me/api/');
const data = await response.json();
const randomUser = data.results[0];
---

<h1>User Profile</h1>

<!-- Pass to framework component (React) -->
<Contact client:load email={randomUser.email} />

<!-- Pass to Astro component -->
<Location city={randomUser.location.city} />
```

**Framework Component (React):**

```tsx
// src/components/Contact.jsx
export default function Contact({ email }) {
  return (
    <div>
      <a href={`mailto:${email}`}>{email}</a>
    </div>
  );
}
```

**Astro Component:**

```astro
---
// src/components/Location.astro
interface Props {
  city: string;
}

const { city } = Astro.props;
---

<div class="location">
  <span>📍 {city}</span>
</div>
```

### Using Relative URLs with `Astro.url`

Construct URLs to your project's pages and endpoints:

```astro
---
// Fetch from your own API endpoint
const url = new URL('/api/users', Astro.url);
const response = await fetch(url);
const users = await response.json();
---

<ul>
  {users.map(user => (
    <li>{user.name}</li>
  ))}
</ul>
```

## fetch() in Framework Components

Framework components (React, Vue, Svelte, etc.) can also use `fetch()`:

### React Example

```tsx
// src/components/Movies.tsx
import type { FunctionalComponent } from 'preact';

const data = await fetch('https://example.com/movies.json').then((response) =>
  response.json()
);

// Build-time rendered: logs to CLI
// With client:* directive: logs to browser console
console.log(data);

const Movies: FunctionalComponent = () => {
  return (
    <div>
      <h2>Movies</h2>
      <ul>
        {data.map((movie) => (
          <li key={movie.id}>{movie.title}</li>
        ))}
      </ul>
    </div>
  );
};

export default Movies;
```

### Vue Example

```vue
<!-- src/components/Posts.vue -->
<script setup>
const response = await fetch('https://jsonplaceholder.typicode.com/posts');
const posts = await response.json();
</script>

<template>
  <div>
    <h2>Posts</h2>
    <article v-for="post in posts" :key="post.id">
      <h3>{{ post.title }}</h3>
      <p>{{ post.body }}</p>
    </article>
  </div>
</template>
```

### Svelte Example

```svelte
<!-- src/components/Users.svelte -->
<script>
  const response = await fetch('https://jsonplaceholder.typicode.com/users');
  const users = await response.json();
</script>

<div>
  <h2>Users</h2>
  {#each users as user}
    <div class="user">
      <strong>{user.name}</strong>
      <span>{user.email}</span>
    </div>
  {/each}
</div>
```

## GraphQL Queries

Use `fetch()` with POST requests to query GraphQL APIs:

### Basic GraphQL Query

```astro
---
// src/components/Film.astro
const response = await fetch(
  'https://swapi-graphql.netlify.app/.netlify/functions/index',
  {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      query: `
        query getFilm ($id: ID!) {
          film(id: $id) {
            title
            releaseDate
          }
        }
      `,
      variables: {
        id: 'ZmlsbXM6MQ==',
      },
    }),
  }
);

const json = await response.json();
const { film } = json.data;
---

<h1>Fetching information about Star Wars: A New Hope</h1>
<h2>Title: {film.title}</h2>
<p>Year: {film.releaseDate}</p>
```

### GraphQL with Error Handling

```astro
---
// src/components/Repository.astro
const GITHUB_TOKEN = import.meta.env.GITHUB_TOKEN;

let repository = null;
let error = null;

try {
  const response = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${GITHUB_TOKEN}`,
    },
    body: JSON.stringify({
      query: `
        query($owner: String!, $name: String!) {
          repository(owner: $owner, name: $name) {
            name
            description
            stargazerCount
            forkCount
          }
        }
      `,
      variables: {
        owner: 'withastro',
        name: 'astro',
      },
    }),
  });

  const json = await response.json();

  if (json.errors) {
    throw new Error(json.errors[0].message);
  }

  repository = json.data.repository;
} catch (e) {
  error = e.message;
  console.error('GraphQL error:', e);
}
---

{error ? (
  <p class="error">{error}</p>
) : repository ? (
  <div class="repository">
    <h2>{repository.name}</h2>
    <p>{repository.description}</p>
    <div class="stats">
      <span>⭐ {repository.stargazerCount}</span>
      <span>🔱 {repository.forkCount}</span>
    </div>
  </div>
) : (
  <p>Loading...</p>
)}
```

### Reusable GraphQL Client

```typescript
// src/lib/graphql.ts
interface GraphQLRequest {
  query: string;
  variables?: Record<string, any>;
}

interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string }>;
}

export async function graphql<T>(
  endpoint: string,
  { query, variables }: GraphQLRequest,
  headers: Record<string, string> = {}
): Promise<T> {
  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
    body: JSON.stringify({ query, variables }),
  });

  const json: GraphQLResponse<T> = await response.json();

  if (json.errors) {
    throw new Error(json.errors[0].message);
  }

  if (!json.data) {
    throw new Error('No data returned from GraphQL query');
  }

  return json.data;
}
```

**Usage:**

```astro
---
import { graphql } from '../lib/graphql';

interface Film {
  title: string;
  releaseDate: string;
}

const data = await graphql<{ film: Film }>(
  'https://swapi-graphql.netlify.app/.netlify/functions/index',
  {
    query: `
      query getFilm ($id: ID!) {
        film(id: $id) {
          title
          releaseDate
        }
      }
    `,
    variables: { id: 'ZmlsbXM6MQ==' },
  }
);

const { film } = data;
---

<h1>{film.title}</h1>
<p>{film.releaseDate}</p>
```

## Fetching from Headless CMS

### REST API Example (Contentful)

```astro
---
// src/pages/blog/index.astro
const CONTENTFUL_SPACE_ID = import.meta.env.CONTENTFUL_SPACE_ID;
const CONTENTFUL_ACCESS_TOKEN = import.meta.env.CONTENTFUL_ACCESS_TOKEN;

const response = await fetch(
  `https://cdn.contentful.com/spaces/${CONTENTFUL_SPACE_ID}/entries?access_token=${CONTENTFUL_ACCESS_TOKEN}&content_type=blogPost`
);

const data = await response.json();
const posts = data.items;
---

<h1>Blog Posts</h1>

{posts.map((post) => (
  <article>
    <h2>{post.fields.title}</h2>
    <p>{post.fields.excerpt}</p>
    <a href={`/blog/${post.fields.slug}`}>Read more</a>
  </article>
))}
```

### GraphQL Example (Hygraph/GraphCMS)

```astro
---
// src/pages/products/index.astro
const GRAPHCMS_ENDPOINT = import.meta.env.GRAPHCMS_ENDPOINT;

const response = await fetch(GRAPHCMS_ENDPOINT, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: `
      query {
        products {
          id
          name
          description
          price
          image {
            url
          }
        }
      }
    `,
  }),
});

const { data } = await response.json();
const { products } = data;
---

<div class="products">
  {products.map((product) => (
    <div class="product">
      <img src={product.image.url} alt={product.name} />
      <h3>{product.name}</h3>
      <p>{product.description}</p>
      <span class="price">${product.price}</span>
    </div>
  ))}
</div>
```

### WordPress REST API

```astro
---
// src/pages/blog/index.astro
const WP_URL = 'https://yoursite.com/wp-json/wp/v2';

const response = await fetch(`${WP_URL}/posts?_embed`);
const posts = await response.json();
---

<h1>Blog Posts</h1>

{posts.map((post) => (
  <article>
    <h2 set:html={post.title.rendered} />
    <div set:html={post.excerpt.rendered} />
    <a href={`/blog/${post.slug}`}>Read more</a>
  </article>
))}
```

## Dynamic Routes with Fetched Data

Generate pages dynamically based on API data:

```astro
---
// src/pages/users/[id].astro
export async function getStaticPaths() {
  const response = await fetch('https://jsonplaceholder.typicode.com/users');
  const users = await response.json();

  return users.map((user) => ({
    params: { id: user.id.toString() },
    props: { user },
  }));
}

const { user } = Astro.props;
---

<h1>{user.name}</h1>
<p>Email: {user.email}</p>
<p>Phone: {user.phone}</p>
<p>Website: {user.website}</p>
```

### With Additional Data Fetching

```astro
---
// src/pages/posts/[slug].astro
export async function getStaticPaths() {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts');
  const posts = await response.json();

  return posts.map((post) => ({
    params: { slug: post.id.toString() },
    props: { postId: post.id },
  }));
}

const { postId } = Astro.props;

// Fetch additional data for this specific post
const [postRes, commentsRes] = await Promise.all([
  fetch(`https://jsonplaceholder.typicode.com/posts/${postId}`),
  fetch(`https://jsonplaceholder.typicode.com/posts/${postId}/comments`),
]);

const post = await postRes.json();
const comments = await commentsRes.json();
---

<article>
  <h1>{post.title}</h1>
  <p>{post.body}</p>

  <h2>Comments ({comments.length})</h2>
  {comments.map((comment) => (
    <div class="comment">
      <strong>{comment.name}</strong>
      <p>{comment.body}</p>
    </div>
  ))}
</article>
```

## Environment Variables

Store sensitive API keys and tokens securely:

```ini
# .env
CONTENTFUL_SPACE_ID=abc123
CONTENTFUL_ACCESS_TOKEN=secret_token
GITHUB_TOKEN=ghp_secret
```

**Usage:**

```astro
---
const CONTENTFUL_SPACE_ID = import.meta.env.CONTENTFUL_SPACE_ID;
const CONTENTFUL_ACCESS_TOKEN = import.meta.env.CONTENTFUL_ACCESS_TOKEN;

const response = await fetch(
  `https://cdn.contentful.com/spaces/${CONTENTFUL_SPACE_ID}/entries?access_token=${CONTENTFUL_ACCESS_TOKEN}`
);
---
```

**TypeScript Support:**

```typescript
// src/env.d.ts
interface ImportMetaEnv {
  readonly CONTENTFUL_SPACE_ID: string;
  readonly CONTENTFUL_ACCESS_TOKEN: string;
  readonly GITHUB_TOKEN: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
```

## Caching and Performance

### Build-Time Caching

In static mode, data is fetched once and cached in the build output:

```astro
---
// This runs ONCE at build time
const response = await fetch('https://api.example.com/data');
const data = await response.json();
---

<!-- Data is baked into HTML -->
<div>{JSON.stringify(data)}</div>
```

### Parallel Fetching

Use `Promise.all()` for concurrent requests:

```astro
---
// Sequential (slow)
const users = await fetch('/api/users').then(r => r.json());
const posts = await fetch('/api/posts').then(r => r.json());

// Parallel (fast)
const [users, posts] = await Promise.all([
  fetch('/api/users').then(r => r.json()),
  fetch('/api/posts').then(r => r.json()),
]);
---
```

### Reusable Data Fetchers

```typescript
// src/lib/api.ts
export async function getUsers() {
  const response = await fetch('https://jsonplaceholder.typicode.com/users');
  if (!response.ok) {
    throw new Error(`Failed to fetch users: ${response.status}`);
  }
  return response.json();
}

export async function getPost(id: number) {
  const response = await fetch(`https://jsonplaceholder.typicode.com/posts/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch post ${id}: ${response.status}`);
  }
  return response.json();
}
```

**Usage:**

```astro
---
import { getUsers, getPost } from '../lib/api';

const users = await getUsers();
const firstPost = await getPost(1);
---
```

## TypeScript Support

### Type-Safe Fetching

```typescript
// src/types/api.ts
export interface User {
  id: number;
  name: string;
  email: string;
  phone: string;
  website: string;
}

export interface Post {
  id: number;
  userId: number;
  title: string;
  body: string;
}
```

**Usage:**

```astro
---
import type { User, Post } from '../types/api';

const response = await fetch('https://jsonplaceholder.typicode.com/users');
const users: User[] = await response.json();

const postResponse = await fetch('https://jsonplaceholder.typicode.com/posts/1');
const post: Post = await postResponse.json();
---

<h1>Users</h1>
{users.map((user: User) => (
  <div>{user.name} - {user.email}</div>
))}
```

### Generic Fetch Wrapper

```typescript
// src/lib/fetch.ts
export async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, options);

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
}
```

**Usage:**

```astro
---
import { fetchJSON } from '../lib/fetch';
import type { User } from '../types/api';

const users = await fetchJSON<User[]>('https://jsonplaceholder.typicode.com/users');
---
```

## Best Practices

### 1. Error Handling

Always handle fetch errors gracefully:

```astro
---
let data = null;
let error = null;

try {
  const response = await fetch('https://api.example.com/data');

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  data = await response.json();
} catch (e) {
  error = e.message;
  console.error('Fetch error:', e);
}
---

{error ? (
  <div class="error">
    <p>Failed to load data: {error}</p>
  </div>
) : data ? (
  <div>{JSON.stringify(data)}</div>
) : (
  <p>Loading...</p>
)}
```

### 2. Environment Variables for Secrets

Never hardcode API keys:

```astro
---
// ✗ Bad: hardcoded token
const response = await fetch('https://api.example.com/data', {
  headers: { 'Authorization': 'Bearer hardcoded_token' }
});

// ✓ Good: from environment
const API_TOKEN = import.meta.env.API_TOKEN;
const response = await fetch('https://api.example.com/data', {
  headers: { 'Authorization': `Bearer ${API_TOKEN}` }
});
---
```

### 3. Parallel Requests

Fetch independent data concurrently:

```astro
---
// ✓ Parallel
const [users, posts, comments] = await Promise.all([
  fetch('/api/users').then(r => r.json()),
  fetch('/api/posts').then(r => r.json()),
  fetch('/api/comments').then(r => r.json()),
]);
---
```

### 4. Type Safety

Use TypeScript for API responses:

```astro
---
import type { ApiResponse } from '../types/api';

const response = await fetch('https://api.example.com/data');
const data: ApiResponse = await response.json();

// TypeScript ensures type safety
console.log(data.items.length); // ✓ Type-checked
---
```

### 5. Reusable Fetchers

Extract common patterns:

```typescript
// src/lib/cms.ts
export async function fetchFromCMS<T>(endpoint: string): Promise<T> {
  const CMS_URL = import.meta.env.CMS_URL;
  const CMS_TOKEN = import.meta.env.CMS_TOKEN;

  const response = await fetch(`${CMS_URL}${endpoint}`, {
    headers: {
      'Authorization': `Bearer ${CMS_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(`CMS error: ${response.status}`);
  }

  return response.json();
}
```

## Common Patterns

### Pagination Component

```astro
---
// src/pages/posts/[page].astro
export async function getStaticPaths() {
  const response = await fetch('https://jsonplaceholder.typicode.com/posts');
  const posts = await response.json();

  const pageSize = 10;
  const pageCount = Math.ceil(posts.length / pageSize);

  return Array.from({ length: pageCount }, (_, i) => ({
    params: { page: (i + 1).toString() },
    props: {
      posts: posts.slice(i * pageSize, (i + 1) * pageSize),
      currentPage: i + 1,
      totalPages: pageCount,
    },
  }));
}

const { posts, currentPage, totalPages } = Astro.props;
---

<div class="posts">
  {posts.map(post => (
    <article>
      <h2>{post.title}</h2>
      <p>{post.body}</p>
    </article>
  ))}
</div>

<nav class="pagination">
  {currentPage > 1 && (
    <a href={`/posts/${currentPage - 1}`}>← Previous</a>
  )}
  <span>Page {currentPage} of {totalPages}</span>
  {currentPage < totalPages && (
    <a href={`/posts/${currentPage + 1}`}>Next →</a>
  )}
</nav>
```

### Search Results Page

```astro
---
// src/pages/search.astro
const query = Astro.url.searchParams.get('q') || '';

let results = [];

if (query) {
  const response = await fetch(
    `https://api.example.com/search?q=${encodeURIComponent(query)}`
  );
  results = await response.json();
}
---

<form method="get">
  <input type="search" name="q" value={query} placeholder="Search..." />
  <button type="submit">Search</button>
</form>

{query && (
  <div class="results">
    <h2>Results for "{query}" ({results.length})</h2>
    {results.map(result => (
      <div class="result">
        <h3>{result.title}</h3>
        <p>{result.excerpt}</p>
      </div>
    ))}
  </div>
)}
```

## Troubleshooting

### Issue: Data Not Updating

**Problem:** Changes to API don't reflect in site.

**Solution:** In static mode, rebuild the site:
```bash
pnpm build
```

For dynamic data, enable SSR:
```javascript
// astro.config.mjs
export default defineConfig({
  output: 'server',
});
```

### Issue: CORS Errors

**Problem:** Browser blocks cross-origin requests.

**Solution:** Fetch happens on the server (no CORS in build), but if using `client:*` directives:
```astro
---
// Server-side (no CORS)
const data = await fetch('https://api.example.com/data').then(r => r.json());
---

<!-- Pass to client component as props -->
<Component client:load data={data} />
```

### Issue: Environment Variables Undefined

**Problem:** `import.meta.env.VAR` is undefined.

**Solution:**
1. Create `.env` file in project root
2. Prefix with `PUBLIC_` for client-side access:
```ini
PUBLIC_API_URL=https://api.example.com
SECRET_TOKEN=secret_value
```

```astro
---
// Server-side: both work
const url = import.meta.env.PUBLIC_API_URL;
const token = import.meta.env.SECRET_TOKEN;
---

<script>
  // Client-side: only PUBLIC_ works
  const url = import.meta.env.PUBLIC_API_URL;
</script>
```

## Quick Reference

### Basic Fetch

```astro
---
const response = await fetch('https://api.example.com/data');
const data = await response.json();
---
```

### GraphQL Query

```astro
---
const response = await fetch('https://api.example.com/graphql', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: `{ users { name } }`,
  }),
});
const { data } = await response.json();
---
```

### Parallel Fetching

```astro
---
const [users, posts] = await Promise.all([
  fetch('/api/users').then(r => r.json()),
  fetch('/api/posts').then(r => r.json()),
]);
---
```

### With Environment Variables

```astro
---
const API_KEY = import.meta.env.API_KEY;
const response = await fetch('https://api.example.com/data', {
  headers: { 'Authorization': `Bearer ${API_KEY}` },
});
---
```

## See Also

- [04-development-workflow.md](./04-development-workflow.md) - Environment variables
- [05-content-collections.md](./05-content-collections.md) - Alternative to API fetching for content
- [03-project-structure.md](./03-project-structure.md) - Where to put API utilities
- [Official CMS Guides](https://docs.astro.build/en/guides/cms/) - Integrating headless CMS

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to data fetching in Astro with fetch(), GraphQL, and headless CMS
**Key Concepts**: fetch(), top-level await, GraphQL, headless CMS, dynamic routes, environment variables
