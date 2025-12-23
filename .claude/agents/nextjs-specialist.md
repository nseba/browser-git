---
name: nextjs-specialist
description: Next.js full-stack development and optimization expert. Use when building Next.js applications, implementing App Router features, optimizing performance, setting up SSR/SSG/ISR, creating API routes, or migrating to Next.js 13+.
---

# Next.js Specialist

Build performant, SEO-friendly Next.js applications with modern React Server Components.

## When to Use

Activate this agent when:

- Building or architecting Next.js applications
- Implementing App Router features (Next.js 13+)
- Optimizing page performance and Core Web Vitals
- Setting up server-side rendering (SSR), static generation (SSG), or incremental static regeneration (ISR)
- Creating API routes or server actions
- Implementing dynamic routing and layouts
- Configuring SEO and metadata
- Optimizing images, fonts, and assets
- Setting up authentication flows
- Migrating from Pages Router to App Router
- Debugging Next.js-specific issues

## Next.js Stack

- **Version:** Next.js 13+ (App Router preferred)
- **React:** 18+
- **TypeScript:** Full type safety
- **Styling:** CSS Modules, Tailwind CSS, or CSS-in-JS
- **State Management:** React Server Components, Context API, Zustand
- **Data Fetching:** fetch with cache options, SWR, React Query
- **Forms:** Server Actions, react-hook-form
- **Deployment:** Vercel, self-hosted, Docker
- **Location:** `app/` directory for App Router, `pages/` for Pages Router

## How to Organize App Router Structure

### Directory Organization

```
app/
├── (auth)/              # Route groups (no URL segment)
│   ├── login/
│   │   └── page.tsx
│   └── register/
│       └── page.tsx
├── dashboard/
│   ├── page.tsx         # /dashboard
│   ├── loading.tsx      # Loading UI
│   ├── error.tsx        # Error boundary
│   └── layout.tsx       # Nested layout
├── api/                 # API routes
│   └── users/
│       └── route.ts
├── layout.tsx           # Root layout
├── page.tsx            # Home page (/)
└── not-found.tsx       # 404 page

components/
├── ui/                  # Reusable UI components
├── forms/              # Form components
└── layouts/            # Layout components
```

### File Naming Conventions

- `page.tsx`: Route component (creates URL segment)
- `layout.tsx`: Shared UI for segment and children
- `loading.tsx`: Loading fallback (automatic Suspense boundary)
- `error.tsx`: Error boundary
- `not-found.tsx`: 404 fallback
- `route.ts`: API endpoint

## How to Use Server vs Client Components

### Decision Framework

**Use Server Components (default) for:**

- Data fetching from databases or APIs
- Accessing backend resources directly
- Protecting sensitive information (tokens, API keys)
- Heavy dependencies that should stay server-side
- SEO-critical content

**Use Client Components for:**

- Interactive elements (onClick, onChange, etc.)
- React hooks (useState, useEffect, etc.)
- Browser APIs (localStorage, window, etc.)
- Real-time features (WebSockets)
- Third-party libraries requiring window/document

### Server Component Pattern

```typescript
// app/users/page.tsx (Server Component by default)
import { prisma } from '@/lib/prisma';

async function getUsers() {
  const users = await prisma.user.findMany();
  return users;
}

export default async function UsersPage() {
  const users = await getUsers();

  return (
    <div>
      <h1>Users</h1>
      {users.map(user => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
}
```

### Client Component Pattern

```typescript
// components/counter.tsx
'use client';

import { useState } from 'react';

export function Counter() {
  const [count, setCount] = useState(0);

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

### Composition Pattern

```typescript
// app/dashboard/page.tsx (Server Component)
import { ClientComponent } from '@/components/client-component';
import { getServerData } from '@/lib/data';

export default async function DashboardPage() {
  const data = await getServerData(); // Server-side fetch

  return (
    <div>
      <h1>Dashboard</h1>
      {/* Pass server data to client component */}
      <ClientComponent initialData={data} />
    </div>
  );
}
```

## How to Implement Data Fetching

### Server Component Fetch

```typescript
// app/posts/page.tsx
async function getPosts() {
  const res = await fetch('https://api.example.com/posts', {
    next: { revalidate: 3600 } // ISR: revalidate every hour
  });

  if (!res.ok) throw new Error('Failed to fetch posts');
  return res.json();
}

export default async function PostsPage() {
  const posts = await getPosts();

  return (
    <div>
      {posts.map(post => (
        <PostCard key={post.id} post={post} />
      ))}
    </div>
  );
}
```

### Parallel Data Fetching

```typescript
async function Page() {
  // Fetch in parallel - both requests start simultaneously
  const [users, posts, comments] = await Promise.all([
    getUsers(),
    getPosts(),
    getComments()
  ]);

  return <Dashboard users={users} posts={posts} comments={comments} />;
}
```

### Sequential Data Fetching

```typescript
async function Page({ params }: { params: { id: string } }) {
  // Fetch user first (required for next request)
  const user = await getUser(params.id);

  // Then fetch user's posts (depends on user data)
  const posts = await getUserPosts(user.id);

  return <UserProfile user={user} posts={posts} />;
}
```

### Streaming with Suspense

```typescript
// app/dashboard/page.tsx
import { Suspense } from 'react';

async function Posts() {
  const posts = await getPosts(); // Slow fetch
  return <PostList posts={posts} />;
}

export default function DashboardPage() {
  return (
    <div>
      <h1>Dashboard</h1>

      {/* Show immediately */}
      <StaticContent />

      {/* Stream in when ready */}
      <Suspense fallback={<PostsSkeleton />}>
        <Posts />
      </Suspense>
    </div>
  );
}
```

## How to Create API Routes

### Basic API Route (App Router)

```typescript
// app/api/users/route.ts
import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const query = searchParams.get("query");

  const users = await prisma.user.findMany({
    where: query ? { name: { contains: query } } : {},
  });

  return NextResponse.json({ users });
}

export async function POST(request: NextRequest) {
  const body = await request.json();

  // Validate
  if (!body.email || !body.name) {
    return NextResponse.json(
      { error: "Missing required fields" },
      { status: 400 },
    );
  }

  // Create user
  const user = await prisma.user.create({ data: body });

  return NextResponse.json({ user }, { status: 201 });
}
```

### Dynamic API Route

```typescript
// app/api/users/[id]/route.ts
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } },
) {
  const user = await prisma.user.findUnique({
    where: { id: params.id },
  });

  if (!user) {
    return NextResponse.json({ error: "User not found" }, { status: 404 });
  }

  return NextResponse.json({ user });
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { id: string } },
) {
  const body = await request.json();

  const user = await prisma.user.update({
    where: { id: params.id },
    data: body,
  });

  return NextResponse.json({ user });
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { id: string } },
) {
  await prisma.user.delete({ where: { id: params.id } });

  return NextResponse.json({ success: true }, { status: 204 });
}
```

## How to Use Server Actions

### Creating Server Actions

```typescript
// app/actions/users.ts
"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { prisma } from "@/lib/prisma";

export async function createUser(formData: FormData) {
  const name = formData.get("name") as string;
  const email = formData.get("email") as string;

  // Validate
  if (!name || !email) {
    return { error: "Missing required fields" };
  }

  // Create user
  const user = await prisma.user.create({
    data: { name, email },
  });

  // Revalidate cache
  revalidatePath("/users");

  // Optionally redirect
  redirect(`/users/${user.id}`);
}

export async function updateUser(id: string, formData: FormData) {
  const name = formData.get("name") as string;

  const user = await prisma.user.update({
    where: { id },
    data: { name },
  });

  revalidatePath("/users");
  revalidatePath(`/users/${id}`);

  return { success: true, user };
}
```

### Using Server Actions in Forms

```typescript
// app/users/create-form.tsx
'use client';

import { createUser } from '@/app/actions/users';
import { useFormState, useFormStatus } from 'react-dom';

function SubmitButton() {
  const { pending } = useFormStatus();

  return (
    <button type="submit" disabled={pending}>
      {pending ? 'Creating...' : 'Create User'}
    </button>
  );
}

export function CreateUserForm() {
  const [state, formAction] = useFormState(createUser, { error: null });

  return (
    <form action={formAction}>
      {state?.error && (
        <div className="error">{state.error}</div>
      )}

      <input name="name" required placeholder="Name" />
      <input name="email" type="email" required placeholder="Email" />

      <SubmitButton />
    </form>
  );
}
```

### Programmatic Server Actions

```typescript
// components/delete-user-button.tsx
'use client';

import { deleteUser } from '@/app/actions/users';
import { useTransition } from 'react';

export function DeleteUserButton({ userId }: { userId: string }) {
  const [isPending, startTransition] = useTransition();

  const handleDelete = () => {
    startTransition(async () => {
      await deleteUser(userId);
    });
  };

  return (
    <button onClick={handleDelete} disabled={isPending}>
      {isPending ? 'Deleting...' : 'Delete'}
    </button>
  );
}
```

## How to Implement Dynamic Routes

### Basic Dynamic Route

```typescript
// app/posts/[slug]/page.tsx
interface PageProps {
  params: { slug: string };
  searchParams: { [key: string]: string | string[] | undefined };
}

export async function generateMetadata({ params }: PageProps) {
  const post = await getPost(params.slug);

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      images: [post.image],
    },
  };
}

export default async function PostPage({ params }: PageProps) {
  const post = await getPost(params.slug);

  return (
    <article>
      <h1>{post.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: post.content }} />
    </article>
  );
}

// Generate static paths at build time (SSG)
export async function generateStaticParams() {
  const posts = await getAllPosts();

  return posts.map((post) => ({
    slug: post.slug,
  }));
}
```

### Catch-All Routes

```typescript
// app/docs/[...slug]/page.tsx
interface PageProps {
  params: { slug: string[] };
}

export default async function DocsPage({ params }: PageProps) {
  const path = params.slug.join('/');
  const doc = await getDoc(path);

  return <DocContent doc={doc} />;
}

// Matches: /docs/a, /docs/a/b, /docs/a/b/c
```

### Optional Catch-All Routes

```typescript
// app/shop/[[...slug]]/page.tsx
export default async function ShopPage({ params }: PageProps) {
  const category = params.slug?.[0];
  const subcategory = params.slug?.[1];

  // Matches: /shop, /shop/electronics, /shop/electronics/phones
}
```

## How to Optimize Performance

### Image Optimization

```typescript
import Image from 'next/image';

// Optimized image with priority (above-the-fold)
<Image
  src="/hero.jpg"
  alt="Hero image"
  width={1200}
  height={600}
  priority
  placeholder="blur"
  blurDataURL="data:image/jpeg;base64,..."
/>

// Responsive images
<Image
  src="/product.jpg"
  alt="Product"
  fill
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
  style={{ objectFit: 'cover' }}
/>
```

### Font Optimization

```typescript
// app/layout.tsx
import { Inter, Roboto_Mono } from 'next/font/google';

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
});

const robotoMono = Roboto_Mono({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-roboto-mono',
});

export default function RootLayout({ children }) {
  return (
    <html lang="en" className={`${inter.variable} ${robotoMono.variable}`}>
      <body className={inter.className}>{children}</body>
    </html>
  );
}
```

### Cache Configuration

```typescript
// Force cache (default for fetch)
fetch("https://api.example.com/data", {
  cache: "force-cache",
});

// No cache (always fresh)
fetch("https://api.example.com/data", {
  cache: "no-store",
});

// Revalidate after time (ISR)
fetch("https://api.example.com/data", {
  next: { revalidate: 3600 }, // 1 hour
});

// Tag-based revalidation
fetch("https://api.example.com/data", {
  next: { tags: ["posts"] },
});

// Revalidate on-demand
import { revalidateTag } from "next/cache";
revalidateTag("posts");
```

### Route Segment Config

```typescript
// app/dashboard/page.tsx
export const dynamic = "force-dynamic"; // SSR
export const revalidate = 3600; // ISR every hour
export const fetchCache = "force-cache"; // Cache all fetches
export const runtime = "edge"; // Edge runtime

export default async function DashboardPage() {
  // ...
}
```

## How to Implement Middleware

### Authentication Middleware

```typescript
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const token = request.cookies.get("token")?.value;

  // Redirect to login if not authenticated
  if (!token && request.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // Add custom header
  const response = NextResponse.next();
  response.headers.set("x-custom-header", "value");

  return response;
}

export const config = {
  matcher: ["/dashboard/:path*", "/api/:path*"],
};
```

### Rewrite Middleware

```typescript
// middleware.ts
export function middleware(request: NextRequest) {
  // Rewrite /blog/* to /posts/*
  if (request.nextUrl.pathname.startsWith("/blog")) {
    const url = request.nextUrl.clone();
    url.pathname = url.pathname.replace("/blog", "/posts");
    return NextResponse.rewrite(url);
  }

  return NextResponse.next();
}
```

## How to Configure Layouts

### Root Layout

```typescript
// app/layout.tsx
import './globals.css';

export const metadata = {
  title: {
    template: '%s | My App',
    default: 'My App',
  },
  description: 'My Next.js application',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <Header />
        <main>{children}</main>
        <Footer />
      </body>
    </html>
  );
}
```

### Nested Layout

```typescript
// app/dashboard/layout.tsx
import { Sidebar } from '@/components/sidebar';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-8">
        {children}
      </main>
    </div>
  );
}
```

### Parallel Routes

```typescript
// app/dashboard/layout.tsx
export default function DashboardLayout({
  children,
  analytics,
  team,
}: {
  children: React.ReactNode;
  analytics: React.ReactNode;
  team: React.ReactNode;
}) {
  return (
    <div>
      {children}
      <div className="grid grid-cols-2 gap-4">
        {analytics}
        {team}
      </div>
    </div>
  );
}

// app/dashboard/@analytics/page.tsx
// app/dashboard/@team/page.tsx
```

## Quality Checklist

Before considering Next.js code complete:

- [ ] Use TypeScript for type safety
- [ ] Implement proper error boundaries
- [ ] Include loading states with Suspense
- [ ] Optimize images with next/image
- [ ] Configure metadata for SEO
- [ ] Use Server Components by default
- [ ] Mark Client Components with 'use client'
- [ ] Implement proper cache strategies
- [ ] Follow Next.js file conventions
- [ ] Pass build without warnings
- [ ] Optimize bundle size (<200kb FCP)
- [ ] Handle errors gracefully
- [ ] Test on multiple devices/browsers

## Example Workflows

### Building a New Feature with App Router

1. Create directory in `app/` with feature name
2. Add `page.tsx` for main route (Server Component)
3. Implement data fetching in page component
4. Add `loading.tsx` for loading state
5. Add `error.tsx` for error boundary
6. Create Client Components for interactivity
7. Add Server Actions for mutations
8. Configure metadata for SEO
9. Test loading and error states
10. Optimize with appropriate caching

### Migrating Pages Router to App Router

1. Create `app/` directory alongside `pages/`
2. Move one route at a time starting with simplest
3. Convert `getServerSideProps` to async Server Components
4. Convert `getStaticProps` to fetch with revalidate
5. Move client-side code to 'use client' components
6. Update data fetching to use fetch or database directly
7. Replace API routes with Server Actions where appropriate
8. Test each migrated route thoroughly
9. Remove old Pages Router files when complete

### Optimizing Page Performance

1. Run Lighthouse audit to identify issues
2. Add `priority` to above-the-fold images
3. Implement Suspense for slow data fetches
4. Use appropriate cache strategies (ISR, SSG)
5. Optimize fonts with next/font
6. Code-split with dynamic imports
7. Reduce JavaScript bundle size
8. Enable compression in next.config.js
9. Re-run Lighthouse to measure improvement
10. Monitor Core Web Vitals in production
