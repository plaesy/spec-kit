---
description: 'JavaScript/TypeScript monorepo scaffolding standards — workspace layout, lockfile discipline, dependency graph hygiene, and the pnpm-strict pitfalls that surface only at install or build time.'
applyTo: '**/package.json,**/pnpm-workspace.yaml,**/turbo.json,**/nx.json,**/.npmrc'
---

# Monorepo Scaffolding (Node/pnpm/TypeScript)

Each item names a pitfall to avoid, then the pattern that prevents it.

1. **Avoid** directory layers inside a monorepo → **use** a flat `apps/` (deployables) + `packages/` (libraries) split to enforce dependency direction and prevent cycles.

2. **Avoid** multiple lockfiles in a monorepo → **use** a single `pnpm-lock.yaml` or `yarn.lock` at root for determinism across installations.

3. **Avoid** phantom dependencies via npm hoisting → **use** pnpm strict mode (`node-linker: strict` in `.npmrc`) to catch them at runtime immediately.

4. **Avoid** circular dependency hell → **use** a unidirectional graph (shared-utils → ui → apps).

5. **Avoid** inconsistent TypeScript configs per package → **use** `extends: "../../tsconfig.base.json"` with overrides only for preset-specific options.

6. **Avoid** ESLint parser misconfiguration → **use** `@typescript-eslint/parser` with `parserOptions.project: './tsconfig.base.json'` for type-aware linting.

7. **Avoid** skipping the workspace protocol in dependencies → **use** `"@myapp/ui": "workspace:*"` (pnpm) to reference live source, not a published version.

8. **Avoid** missing or incomplete pre-commit hooks → **use** Husky + lint-staged to run ESLint/Prettier on staged files only, no full codebase scan.

9. **Avoid** leaving conditional exports undeclared → **use** the `exports` field in `package.json` with `{ "browser": "...", "require": "...", "import": "..." }` conditions.

10. **Avoid** over-reliance on root `node_modules` → **use** Nx or Turborepo for distributed caching and dependency graph analysis.

11. **Avoid** mixing pnpm strict mode with unhandled peer dependencies → **use** `peerDependencies` for transitive deps and list them in `peerDependenciesMeta.*.optional = true` when optional.

12. **Avoid** unvalidated remote templates → **use** GitHub releases with semver tags (`owner/repo#v1.2.3`) and a validated manifest.

13. **Avoid** skipping post-scaffolding validation → run `pnpm install && pnpm lint --max-warnings=0 && pnpm build` right after scaffolding to catch config errors before writing any feature code.

14. **Avoid** forgetting `.npmrc` strict mode setup → **use** a committed `.npmrc` with `strict-peer-dependencies=true` and `node-linker=strict` for determinism.

## Verifying a Scaffold

- `pnpm-workspace.yaml` globs every package under `apps/**` and `packages/**`
- Every `package.json` parses as valid JSON and declares `name`, `version`, `type`, `exports`
- `pnpm install && pnpm lint --max-warnings=0 && pnpm build` all pass from a clean checkout
