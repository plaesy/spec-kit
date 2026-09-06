---
applyTo: '**/*.ts, **/*.js, **/*.json, **/*.spec.ts, **/*.e2e-spec.ts'
description: 'Next.js development standards and best practices for building scalable Node.js server-side applications'
---

# Next.js Best Practices for LLMs (2025)

_Last updated: July 2025_

Latest authoritative best practices for building/structuring/maintaining Next.js apps - for LLMs and developers, for code quality, maintainability, scalability.

---

## 1. Project Structure & Organization

- Use the `app/` directory (App Router) for all new projects, not legacy `pages/`
- Top-level folders: `app/` (routing/layouts/pages/route handlers), `public/` (static assets), `lib/` (shared utilities/API clients/logic), `components/` (reusable UI), `contexts/` (React context providers), `styles/`, `hooks/` (custom hooks), `types/` (TS definitions)
- Colocate files (components/styles/tests) near where used, avoid deep nesting
- Route groups: parentheses (e.g. `(admin)`) to group routes without affecting the URL
- Private folders: prefix `_` (e.g. `_internal`) to opt out of routing
- Feature folders for large apps (e.g. `app/dashboard/`, `app/auth/`)
- `src/` (optional): separate source code from config files

## 2.1. Server and Client Component Integration (App Router)

**Never use `next/dynamic` with `{ ssr: false }` inside a Server Component** - unsupported, causes a build/runtime error.

**Correct approach**: move client-only logic/UI into a dedicated Client Component (`'use client'` at top), import it directly into the Server Component (no `next/dynamic` needed). Composing multiple client-only elements (e.g. navbar + profile dropdown)? Put them all in one Client Component.

**Example:**

```tsx
// Server Component
import DashboardNavbar from '@/components/DashboardNavbar';

export default async function DashboardPage() {
  // ...server logic...
  return (
    <>
      <DashboardNavbar /> {/* This is a Client Component */}
      {/* ...rest of server-rendered page... */}
    </>
  );
}
```

**Why**: Server Components can't use client-only features or SSR-disabled dynamic imports. Client Components can render inside Server Components, not vice versa.

**Summary**: move client-only UI into a Client Component, import directly into the Server Component - never `next/dynamic` with `{ ssr: false }` in a Server Component.

---

## 2. Component Best Practices

- **Types**: Server Components (default) for data fetching/heavy logic/non-interactive UI; Client Components (`'use client'` at top) for interactivity/state/browser APIs
- **When to create**: reused UI pattern, complex/self-contained page section, or improves readability/testability
- **Naming**: `PascalCase` for component files/exports (`UserCard.tsx`), `camelCase` for hooks (`useUser.ts`), `snake_case`/`kebab-case` for static assets, context providers as `XyzProvider`
- **File naming**: file name matches component name; single-export files default-export; multiple related components use an `index.ts` barrel
- **Location**: shared components in `components/`, route-specific inside the relevant route folder
- **Props**: TypeScript interfaces, explicit types and default values
- **Testing**: co-locate tests with components (`UserCard.test.tsx`)

## 3. Naming Conventions (General)

- Folders: `kebab-case` (`user-profile/`)
- Files: `PascalCase` components, `camelCase` utilities/hooks, `kebab-case` static assets
- Variables/functions: `camelCase`
- Types/interfaces: `PascalCase`
- Constants: `UPPER_SNAKE_CASE`

## 4. API Routes (Route Handlers)

- Prefer API Routes over Edge Functions unless ultra-low latency/geographic distribution is needed
- Location: `app/api/` (e.g. `app/api/users/route.ts`)
- HTTP methods: export async functions named after verbs (`GET`, `POST`, etc.)
- Request/Response: Web `Request`/`Response` APIs, `NextRequest`/`NextResponse` for advanced features
- Dynamic segments: `[param]` (e.g. `app/api/users/[id]/route.ts`)
- Validation: always validate/sanitize input (`zod`, `yup`)
- Error handling: appropriate HTTP status codes and messages
- Auth: protect sensitive routes via middleware or server-side session checks

## 5. General Best Practices

- **TypeScript**: everywhere, `strict` mode in `tsconfig.json`
- **ESLint & Prettier**: enforced, official Next.js ESLint config
- **Env vars**: secrets in `.env.local`, never committed
- **Testing**: Jest, React Testing Library, or Playwright - cover all critical logic/components
- **Accessibility**: semantic HTML, ARIA attributes, test with screen readers
- **Performance**: built-in Image/Font optimization; Suspense + loading states for async data; avoid large client bundles, keep most logic in Server Components
- **Security**: sanitize all user input, HTTPS in production, secure HTTP headers
- **Documentation**: clear README and code comments, document public APIs/components

# Avoid Unnecessary Example Files
Don't create example/demo files (e.g. `ModalExample.tsx`) unless the user specifically requests a live example, Storybook story, or explicit documentation component - keep the repo clean and production-focused by default.

# Always use the latest documentation and guides
- For every Next.js-related request, start by searching the most current Next.js documentation/guides/examples
- If available, use `resolve_library_id` to resolve the package/library name in docs, and `get_library_docs` for up-to-date documentation


