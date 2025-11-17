# Astro DB

## Overview

Astro DB is a fully-managed SQL database designed specifically for the Astro ecosystem. It provides a complete solution for configuring, developing, and querying your data with full TypeScript support, built-in ORM (Drizzle), and seamless integration with both local development and production deployments.

## Key Concepts

### What is Astro DB?

- **Fully-managed SQL database** built on libSQL (SQLite fork)
- **Local-first development** with `.astro/content.db` file
- **Zero configuration** - no Docker or network connection needed for development
- **Built-in Drizzle ORM** with TypeScript type generation
- **Production-ready** with libSQL-compatible database support (Turso, local files, etc.)

### When to Use Astro DB

**Use Astro DB when:**
- Building apps with user-generated content (comments, posts, profiles)
- Managing structured relational data
- Need type-safe database queries
- Want seamless local → production workflow

**Consider alternatives when:**
- Using a headless CMS (see Content Collections or data fetching)
- Building a static site with no user data
- Already have an existing database system

## Installation

### Using astro add

The easiest way to install Astro DB:

```bash
# npm
npx astro add db

# pnpm
pnpm astro add db

# Yarn
yarn astro add db
```

This automatically:
1. Installs `@astrojs/db` integration
2. Creates `db/config.ts` file
3. Updates `astro.config.mjs`

### Manual Installation

```bash
pnpm add @astrojs/db
```

**Configure astro.config.mjs:**

```javascript
import { defineConfig } from 'astro/config';
import db from '@astrojs/db';

export default defineConfig({
  integrations: [db()],
});
```

## Database Configuration

### Define Tables

Create your database schema in `db/config.ts`:

```typescript
// db/config.ts
import { defineDb, defineTable, column } from 'astro:db';

const Comment = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    author: column.text(),
    body: column.text(),
    createdAt: column.date(),
  },
});

export default defineDb({
  tables: { Comment },
});
```

### Column Types

Astro DB supports five column types:

```typescript
import { defineTable, column } from 'astro:db';

const Example = defineTable({
  columns: {
    // String of text
    name: column.text(),

    // Whole integer value
    age: column.number(),

    // True or false value
    isActive: column.boolean(),

    // Date/time (queried as JavaScript Date objects)
    createdAt: column.date(),

    // Untyped JSON object
    metadata: column.json(),
  },
});
```

### Column Options

**Primary Key:**

```typescript
const User = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    email: column.text(),
  },
});
```

**Optional Columns:**

```typescript
const Post = defineTable({
  columns: {
    title: column.text(),
    excerpt: column.text({ optional: true }),
  },
});
```

**Unique Constraint:**

```typescript
const User = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    email: column.text({ unique: true }),
  },
});
```

**Default Values:**

```typescript
const Post = defineTable({
  columns: {
    title: column.text(),
    published: column.boolean({ default: false }),
    views: column.number({ default: 0 }),
  },
});
```

## Table Relationships

### Foreign Keys with References

Define relationships between tables using references:

```typescript
// db/config.ts
import { defineDb, defineTable, column } from 'astro:db';

const Author = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    name: column.text(),
    email: column.text(),
  },
});

const Comment = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    authorId: column.number({ references: () => Author.columns.id }),
    body: column.text(),
  },
});

export default defineDb({
  tables: { Author, Comment },
});
```

### Many-to-Many Relationships

Use junction tables:

```typescript
const Post = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    title: column.text(),
  },
});

const Tag = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    name: column.text(),
  },
});

const PostTag = defineTable({
  columns: {
    postId: column.number({ references: () => Post.columns.id }),
    tagId: column.number({ references: () => Tag.columns.id }),
  },
});

export default defineDb({
  tables: { Post, Tag, PostTag },
});
```

## Seeding Development Data

### Basic Seeding

Create `db/seed.ts` to populate development data:

```typescript
// db/seed.ts
import { db, Author, Comment } from 'astro:db';

export default async function() {
  await db.insert(Author).values([
    { id: 1, name: 'Kasim', email: 'kasim@example.com' },
    { id: 2, name: 'Mina', email: 'mina@example.com' },
  ]);

  await db.insert(Comment).values([
    { authorId: 1, body: 'Hope you like Astro DB!' },
    { authorId: 2, body: 'Enjoy!' },
  ]);
}
```

**How it works:**
- Runs automatically when `astro dev` starts
- Recreates database fresh each time
- Development data only - never affects production

### Seeding with Loops

```typescript
// db/seed.ts
import { db, Post } from 'astro:db';

export default async function() {
  const posts = [];

  for (let i = 1; i <= 100; i++) {
    posts.push({
      id: i,
      title: `Post ${i}`,
      body: `This is the body of post ${i}`,
      published: i % 2 === 0, // Every other post is published
    });
  }

  await db.insert(Post).values(posts);
}
```

### Seeding with Faker

```typescript
// db/seed.ts
import { db, User, Post } from 'astro:db';
import { faker } from '@faker-js/faker';

export default async function() {
  // Seed 50 fake users
  const users = Array.from({ length: 50 }, (_, i) => ({
    id: i + 1,
    name: faker.person.fullName(),
    email: faker.internet.email(),
  }));

  await db.insert(User).values(users);

  // Seed 200 fake posts
  const posts = Array.from({ length: 200 }, (_, i) => ({
    id: i + 1,
    authorId: faker.number.int({ min: 1, max: 50 }),
    title: faker.lorem.sentence(),
    body: faker.lorem.paragraphs(3),
  }));

  await db.insert(Post).values(posts);
}
```

## Production Database Setup

### Using Turso (Recommended)

Turso is the official libSQL database platform, fully compatible with Astro DB.

**1. Install Turso CLI:**

```bash
# macOS/Linux
curl -sSfL https://get.tur.so/install.sh | bash

# Windows
powershell -c "irm get.tur.so/windows | iex"
```

**2. Sign Up/Log In:**

```bash
turso auth signup  # or turso auth login
```

**3. Create Database:**

```bash
turso db create andromeda
```

**4. Get Connection URL:**

```bash
turso db show andromeda
```

Copy the URL value and set as environment variable:

```ini
# .env
ASTRO_DB_REMOTE_URL=libsql://andromeda-houston.turso.io
```

**5. Create Auth Token:**

```bash
turso db tokens create andromeda
```

Copy the token and set as environment variable:

```ini
# .env
ASTRO_DB_REMOTE_URL=libsql://andromeda-houston.turso.io
ASTRO_DB_APP_TOKEN=eyJhbGciOiJF...3ahJpTkKDw
```

**6. Push Schema to Remote:**

```bash
pnpm astro db push --remote
```

### Remote URL Configuration

The `ASTRO_DB_REMOTE_URL` supports multiple options:

**URL Schemes:**

```ini
# In-memory database
ASTRO_DB_REMOTE_URL=memory:

# Local file
ASTRO_DB_REMOTE_URL=file:path/to/database.db

# Remote via libSQL protocol
ASTRO_DB_REMOTE_URL=libsql://your.server.io

# Remote via HTTP
ASTRO_DB_REMOTE_URL=https://your.server.io

# Remote via WebSockets
ASTRO_DB_REMOTE_URL=wss://your.server.io
```

**With Encryption:**

```ini
ASTRO_DB_REMOTE_URL=file:local-copy.db?encryptionKey=your-encryption-key
```

**Embedded Replica (Sync):**

```ini
# In-memory replica that syncs with remote
ASTRO_DB_REMOTE_URL=memory:?syncUrl=libsql%3A%2F%2Fyour.server.io

# Sync every 60 seconds
ASTRO_DB_REMOTE_URL=memory:?syncUrl=libsql%3A%2F%2Fyour.server.io&syncInterval=60
```

### Deployment Configuration

**1. Update Build Command:**

```json
// package.json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build --remote",
    "preview": "astro preview"
  }
}
```

**2. Set Environment Variables:**

Set `ASTRO_DB_REMOTE_URL` and `ASTRO_DB_APP_TOKEN` in your deployment platform (Vercel, Netlify, Cloudflare, etc.).

**3. Deploy:**

Your build command will now connect to the remote database.

## Querying Data

### The Drizzle ORM Client

Astro DB includes a built-in Drizzle ORM client:

```astro
---
import { db } from 'astro:db';
---
```

All database operations use this `db` object.

### SELECT Queries

**Select All Rows:**

```astro
---
import { db, Comment } from 'astro:db';

const comments = await db.select().from(Comment);
---

<h2>Comments ({comments.length})</h2>

{comments.map(({ author, body }) => (
  <article>
    <strong>{author}</strong>
    <p>{body}</p>
  </article>
))}
```

**Select Specific Columns:**

```astro
---
import { db, User } from 'astro:db';

const users = await db.select({
  name: User.name,
  email: User.email,
}).from(User);
---
```

**Select Single Row:**

```astro
---
import { db, User, eq } from 'astro:db';

const user = await db.select()
  .from(User)
  .where(eq(User.id, 1))
  .get(); // Returns single row or undefined
---

{user ? (
  <h1>{user.name}</h1>
) : (
  <p>User not found</p>
)}
```

### Filtering with WHERE

**Basic Filters:**

```astro
---
import { db, Post, eq, gt, like } from 'astro:db';

// Exact match
const publishedPosts = await db.select()
  .from(Post)
  .where(eq(Post.published, true));

// Greater than
const recentPosts = await db.select()
  .from(Post)
  .where(gt(Post.views, 100));

// Text search
const astroPosts = await db.select()
  .from(Post)
  .where(like(Post.title, '%Astro%'));
---
```

**Multiple Conditions (AND):**

```astro
---
import { db, Post, and, eq, gt } from 'astro:db';

const posts = await db.select()
  .from(Post)
  .where(
    and(
      eq(Post.published, true),
      gt(Post.views, 100)
    )
  );
---
```

**Multiple Conditions (OR):**

```astro
---
import { db, Post, or, eq, like } from 'astro:db';

const posts = await db.select()
  .from(Post)
  .where(
    or(
      like(Post.title, '%Astro%'),
      like(Post.title, '%DB%')
    )
  );
---
```

### Ordering and Limiting

```astro
---
import { db, Post, desc, asc } from 'astro:db';

// Order by views descending
const topPosts = await db.select()
  .from(Post)
  .orderBy(desc(Post.views));

// Order by title ascending
const alphabeticalPosts = await db.select()
  .from(Post)
  .orderBy(asc(Post.title));

// Limit results
const latestPosts = await db.select()
  .from(Post)
  .orderBy(desc(Post.createdAt))
  .limit(10);

// Offset and limit (pagination)
const page2Posts = await db.select()
  .from(Post)
  .orderBy(desc(Post.createdAt))
  .limit(10)
  .offset(10);
---
```

### JOIN Queries

**Inner Join:**

```astro
---
import { db, eq, Comment, Author } from 'astro:db';

const comments = await db.select()
  .from(Comment)
  .innerJoin(Author, eq(Comment.authorId, Author.id));
---

{comments.map(({ Author, Comment }) => (
  <article>
    <strong>{Author.name}</strong>
    <p>{Comment.body}</p>
  </article>
))}
```

**Left Join:**

```astro
---
import { db, eq, Post, Comment } from 'astro:db';

const postsWithComments = await db.select()
  .from(Post)
  .leftJoin(Comment, eq(Post.id, Comment.postId));
---
```

### Aggregation

**Count:**

```astro
---
import { db, count, Comment } from 'astro:db';

const result = await db.select({
  count: count(),
}).from(Comment);

const commentCount = result[0].count;
---

<p>Total comments: {commentCount}</p>
```

**Group By:**

```astro
---
import { db, count, Comment } from 'astro:db';

const authorCommentCounts = await db.select({
  authorId: Comment.authorId,
  commentCount: count(),
})
  .from(Comment)
  .groupBy(Comment.authorId);
---

{authorCommentCounts.map(({ authorId, commentCount }) => (
  <p>Author {authorId}: {commentCount} comments</p>
))}
```

## INSERT Operations

### Basic Insert

```astro
---
import { db, Comment } from 'astro:db';

if (Astro.request.method === 'POST') {
  const formData = await Astro.request.formData();
  const author = formData.get('author');
  const body = formData.get('body');

  if (typeof author === 'string' && typeof body === 'string') {
    await db.insert(Comment).values({ author, body });
  }
}
---

<form method="POST">
  <label>
    Author: <input name="author" required />
  </label>
  <label>
    Comment: <textarea name="body" required></textarea>
  </label>
  <button type="submit">Submit</button>
</form>
```

### Insert Multiple Rows

```astro
---
import { db, Comment } from 'astro:db';

await db.insert(Comment).values([
  { author: 'Alice', body: 'Great post!' },
  { author: 'Bob', body: 'Thanks for sharing!' },
  { author: 'Charlie', body: 'Very helpful!' },
]);
---
```

### Insert with Returning

```astro
---
import { db, Comment } from 'astro:db';

const newComments = await db.insert(Comment)
  .values({ author: 'Alice', body: 'Hello!' })
  .returning();

// newComments contains the inserted row(s) with generated IDs
console.log(newComments[0].id);
---
```

### Using Astro Actions

```typescript
// src/actions/index.ts
import { db, Comment } from 'astro:db';
import { defineAction } from 'astro:actions';
import { z } from 'astro:schema';

export const server = {
  addComment: defineAction({
    input: z.object({
      author: z.string(),
      body: z.string(),
    }),
    handler: async (input) => {
      const updatedComments = await db
        .insert(Comment)
        .values(input)
        .returning();

      return updatedComments;
    },
  }),
};
```

**Usage in Component:**

```astro
---
import { actions } from 'astro:actions';

const result = Astro.getActionResult(actions.addComment);
---

<form method="POST" action={actions.addComment}>
  <input name="author" required />
  <textarea name="body" required></textarea>
  <button type="submit">Add Comment</button>
</form>

{result?.data && <p>Comment added successfully!</p>}
```

## UPDATE Operations

```astro
---
import { db, Post, eq } from 'astro:db';

// Update a single post
await db.update(Post)
  .set({ published: true })
  .where(eq(Post.id, 1));

// Update multiple posts
await db.update(Post)
  .set({ views: 0 })
  .where(eq(Post.published, false));

// Update with returning
const updatedPosts = await db.update(Post)
  .set({ featured: true })
  .where(eq(Post.views, 1000))
  .returning();
---
```

## DELETE Operations

**Basic Delete:**

```typescript
// src/pages/api/comments/[id].ts
import type { APIRoute } from 'astro';
import { db, Comment, eq } from 'astro:db';

export const DELETE: APIRoute = async (ctx) => {
  await db.delete(Comment).where(eq(Comment.id, ctx.params.id));

  return new Response(null, { status: 204 });
};
```

**Delete Multiple Rows:**

```astro
---
import { db, Post, lt } from 'astro:db';

// Delete old unpublished posts
const oneYearAgo = new Date();
oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

await db.delete(Post).where(
  and(
    eq(Post.published, false),
    lt(Post.createdAt, oneYearAgo)
  )
);
---
```

## Batch Transactions

For multiple queries, use batching to improve performance:

```typescript
// db/seed.ts
import { db, Author, Comment } from 'astro:db';

export default async function() {
  const queries = [];

  // Seed 100 comments in a single network request
  for (let i = 0; i < 100; i++) {
    queries.push(
      db.insert(Comment).values({ body: `Test comment ${i}` })
    );
  }

  await db.batch(queries);
}
```

## Managing Schema Changes

### Pushing Schema Changes

**To Remote Database:**

```bash
pnpm astro db push --remote
```

This command:
1. Verifies changes are safe
2. Suggests fixes for conflicts
3. Applies changes to remote database

**Force Reset (Destructive):**

```bash
pnpm astro db push --remote --force-reset
```

⚠️ **Warning:** This destroys all production data!

### Renaming Tables Safely

**1. Mark Old Table as Deprecated:**

```typescript
// db/config.ts
const Comment = defineTable({
  deprecated: true,
  columns: {
    author: column.text(),
    body: column.text(),
  },
});
```

**2. Create New Table:**

```typescript
const Feedback = defineTable({
  columns: {
    author: column.text(),
    body: column.text(),
  },
});

export default defineDb({
  tables: { Comment, Feedback },
});
```

**3. Push Changes:**

```bash
pnpm astro db push --remote
```

**4. Update Code to Use New Table:**

Migrate data and update all references.

**5. Remove Old Table:**

```typescript
// db/config.ts
export default defineDb({
  tables: { Feedback }, // Comment removed
});
```

**6. Push Again:**

```bash
pnpm astro db push --remote
```

## TypeScript Support

### Auto-Generated Types

Astro automatically generates TypeScript types for your tables:

```astro
---
import { db, Comment, type Comment as CommentType } from 'astro:db';

// Type-safe query
const comments: CommentType[] = await db.select().from(Comment);

// TypeScript knows the shape
comments.forEach(comment => {
  console.log(comment.author); // ✓ Type-checked
  console.log(comment.invalid); // ✗ TypeScript error
});
---
```

### Type-Safe Insert

```typescript
import { db, Comment } from 'astro:db';

// ✓ Type-checked
await db.insert(Comment).values({
  author: 'Alice',
  body: 'Great post!',
});

// ✗ TypeScript error: missing required field
await db.insert(Comment).values({
  author: 'Alice',
  // Missing 'body'
});
```

## Best Practices

### 1. Use Transactions for Related Operations

```typescript
await db.batch([
  db.insert(User).values({ id: 1, name: 'Alice' }),
  db.insert(Post).values({ authorId: 1, title: 'First Post' }),
]);
```

### 2. Enable SSR for Dynamic Data

```javascript
// astro.config.mjs
export default defineConfig({
  output: 'server', // or 'hybrid'
  adapter: node(),
  integrations: [db()],
});
```

### 3. Validate Input with Zod

```typescript
import { z } from 'astro:schema';

const commentSchema = z.object({
  author: z.string().min(1).max(100),
  body: z.string().min(10).max(1000),
});

const input = commentSchema.parse(formData);
await db.insert(Comment).values(input);
```

### 4. Index Foreign Keys

```typescript
const Comment = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    postId: column.number({
      references: () => Post.columns.id,
      index: true, // Improves JOIN performance
    }),
    body: column.text(),
  },
});
```

### 5. Use Environment Variables

```ini
# .env (local)
ASTRO_DB_REMOTE_URL=libsql://localhost:8080
ASTRO_DB_APP_TOKEN=local-token

# .env.production (deploy platform)
ASTRO_DB_REMOTE_URL=libsql://prod.turso.io
ASTRO_DB_APP_TOKEN=prod-token
```

## Common Patterns

### Pagination

```astro
---
export async function getStaticPaths() {
  const pageSize = 20;
  const totalPosts = await db.select({ count: count() }).from(Post);
  const pageCount = Math.ceil(totalPosts[0].count / pageSize);

  return Array.from({ length: pageCount }, (_, i) => ({
    params: { page: (i + 1).toString() },
  }));
}

const { page } = Astro.params;
const pageNumber = parseInt(page);
const pageSize = 20;

const posts = await db.select()
  .from(Post)
  .orderBy(desc(Post.createdAt))
  .limit(pageSize)
  .offset((pageNumber - 1) * pageSize);
---
```

### User Authentication

```astro
---
import { db, User, eq } from 'astro:db';
import bcrypt from 'bcryptjs';

if (Astro.request.method === 'POST') {
  const formData = await Astro.request.formData();
  const email = formData.get('email') as string;
  const password = formData.get('password') as string;

  const user = await db.select()
    .from(User)
    .where(eq(User.email, email))
    .get();

  if (user && await bcrypt.compare(password, user.passwordHash)) {
    // Authentication successful
  }
}
---
```

### Dynamic Search

```astro
---
import { db, Post, or, like } from 'astro:db';

const query = Astro.url.searchParams.get('q') || '';
let results = [];

if (query) {
  results = await db.select()
    .from(Post)
    .where(
      or(
        like(Post.title, `%${query}%`),
        like(Post.body, `%${query}%`)
      )
    );
}
---
```

## Troubleshooting

### Issue: Types Not Generated

**Solution:**
```bash
# Restart dev server
pnpm astro dev
```

### Issue: Remote Connection Fails

**Check:**
1. Environment variables set correctly
2. Network connectivity to remote database
3. Auth token is valid

```bash
# Test connection
turso db shell [database-name]
```

### Issue: Schema Push Fails

**Solution:**
```bash
# See what changes will be made
pnpm astro db push --remote --dry-run

# If safe, apply
pnpm astro db push --remote

# If breaking changes needed
pnpm astro db push --remote --force-reset
```

## Quick Reference

### Setup

```bash
pnpm astro add db
```

### Define Table

```typescript
import { defineTable, column } from 'astro:db';

const Comment = defineTable({
  columns: {
    id: column.number({ primaryKey: true }),
    body: column.text(),
  },
});
```

### Seed Data

```typescript
// db/seed.ts
import { db, Comment } from 'astro:db';

export default async function() {
  await db.insert(Comment).values([
    { body: 'Hello!' },
  ]);
}
```

### Query

```astro
---
import { db, Comment, eq } from 'astro:db';

const comments = await db.select().from(Comment);
const comment = await db.select()
  .from(Comment)
  .where(eq(Comment.id, 1))
  .get();
---
```

### Insert

```astro
---
await db.insert(Comment).values({ body: 'New comment' });
---
```

### Update

```astro
---
await db.update(Comment)
  .set({ body: 'Updated' })
  .where(eq(Comment.id, 1));
---
```

### Delete

```astro
---
await db.delete(Comment).where(eq(Comment.id, 1));
---
```

## See Also

- [05-content-collections.md](./05-content-collections.md) - Alternative for static content
- [10-data-fetching.md](./10-data-fetching.md) - Fetching from external APIs
- [04-development-workflow.md](./04-development-workflow.md) - Environment variables
- [Turso Documentation](https://docs.turso.tech/) - Official libSQL database platform
- [Drizzle ORM Docs](https://orm.drizzle.team/) - Complete ORM reference

---

**Last Updated**: 2025-11-17
**Purpose**: Comprehensive guide to Astro DB - a fully-managed SQL database for Astro
**Key Concepts**: libSQL, Drizzle ORM, table definitions, seeding, remote databases, Turso, type safety
