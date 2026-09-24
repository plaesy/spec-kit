---
description: "Generate monorepo and project scaffolding with pnpm workspaces, TypeScript, and pre-configured tooling — `/create:template:project`"
---

# `/create:template:project` command

⚡ **Run with**: standard (single-agent; request phases in parallel where possible)

## Usage Format

```bash
/create:template:project --name myapp --type monorepo --preset react-node
/create:template:project --name ui-lib --type package --preset typescript-library
/create:template:project --name frontend --type app --preset react-vite --output ~/projects
```

### Flags Reference

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (required) | Project/monorepo name (kebab-case, 3-214 chars) |
| `--type` | enum | (required) | `monorepo`, `app`, `package` |
| `--preset` | enum | (required) | `react-vite`, `react-next`, `node-ts`, `typescript-library`, `full-stack` |
| `--packageManager` | enum | `pnpm` | `pnpm`, `npm`, `yarn` — must be installed in PATH |
| `--typescript` | boolean | `true` | Include TypeScript configuration with strict mode |
| `--eslint` | boolean | `true` | Include ESLint with @typescript-eslint ruleset |
| `--prettier` | boolean | `true` | Include Prettier code formatter |
| `--husky` | boolean | `true` | Include Husky git hooks for pre-commit |
| `--lintStaged` | boolean | `true` | Include lint-staged to run checks on staged files only |
| `--vite` | boolean | `true` | Use Vite for builds (React/Node apps) instead of webpack |
| `--output` | path | `./` | Destination directory for scaffolding (parent must exist) |
| `--template` | url | (optional) | Remote GitHub template URL with version pin: `owner/repo#v1.0.0` |
| `--skipInstall` | boolean | `false` | Skip automatic `pnpm install` / `npm install` after scaffolding |
| `--strict` | boolean | `true` | Enable pnpm strict node_modules mode to catch phantom dependencies |

## Protocol

### Step 1: Resolve Configuration
- **Validate project name** against npm package naming rules (lowercase, hyphens, no leading dots)
- **Confirm type/preset compatibility**: e.g., `type:app preset:typescript-library` is invalid
- **Check package manager availability**: Run `which pnpm` / `where pnpm.cmd` (Windows); fail loudly if missing and no fallback
- **Verify output directory exists**: If `--output` points to non-existent parent, fail with guidance on creating it first
- **Detect file conflicts**: Scan `--output/{name}/` for existing files; prompt user before overwriting

**Exit condition**: All validation passes or user confirms override

### Step 2: Provider Selection
- **If `--template` provided**: Fetch remote template from GitHub with version pinning
  - Parse format: `owner/repo#v1.0.0` (semver required)
  - Validate manifest `.scaffolding-manifest.json` at repo root matches expected schema
  - Fall back to built-in preset if remote unavailable
- **Otherwise**: Use built-in preset templates stored in `~/.claude/templates/{preset}/`
- **Validate template manifest** includes required fields: `name`, `version`, `packages` (array), `preset`, `features`
- **Fail gracefully**: If preset unavailable, suggest installed presets and exit with helpful message

**Citation**: Best practices from [pnpm 9 Workspaces Guide](https://pristren.com/blog/pnpm-9-workspaces-guide/) recommend symlink-based node_modules over npm's hoisting to prevent phantom dependencies.

### Step 3: Generate Structure
- **Create root directory** with populated `package.json`:
  - Set `"name"`, `"version"` (default: "0.0.1"), `"private": true`
  - Include `pnpm.workspaces` array or `workspaces` field per package manager
  - Add shared scripts: `"lint"`, `"format"`, `"build"`, `"test"`, `"dev"`
- **Generate `pnpm-workspace.yaml`** (or `workspaces` in root `package.json` for npm/yarn):
  ```yaml
  packages:
    - 'apps/*'
    - 'packages/*'
  ```
- **Output `tsconfig.base.json`** with path aliases:
  ```json
  {
    "compilerOptions": {
      "strict": true,
      "baseUrl": ".",
      "paths": {
        "@myapp/*": ["packages/*/src"],
        "@myapp/ui": ["packages/ui/src"]
      }
    }
  }
  ```
- **Generate `.eslintrc.base.js`** extending `typescript-eslint/recommended-type-checked`:
  - Include parser: `@typescript-eslint/parser`
  - Set `parserOptions.project: './tsconfig.base.json'` for type-aware linting
- **Create `.prettierrc.json`** with standard config: `semi: true`, `tabWidth: 2`, `singleQuote: false`
- **Initialize `.husky/` directory** with pre-commit hook calling `lint-staged` on staged files
- **For each app/package in preset**:
  - Create `{name}/package.json` with `exports` field for conditional entrypoints
  - Use `workspace:*` protocol (pnpm) for internal dependencies
  - Add `src/` directory with `index.ts` placeholder
  - Generate `{name}/tsconfig.json` extending `../../tsconfig.base.json`
- **Generate `README.md`** with setup instructions, workspace overview, and command reference

**Citation**: [JavaScript Monorepos 2026: Best Practices](https://www.pkgpulse.com/guides/javascript-monorepos-2026-best-practices-pitfalls) emphasizes separating `apps/` (deployable) from `packages/` (shared libs) to enforce dependency layering.

### Step 4: Index & Manifest
- **Write `.scaffolding-manifest.json`** with template metadata:
  ```json
  {
    "preset": "react-node",
    "version": "0.0.1",
    "generatedAt": "2026-09-25T12:00:00Z",
    "packageManager": "pnpm",
    "packages": ["apps/frontend", "apps/backend", "packages/ui", "packages/shared"]
  }
  ```
- **Output file tree summary** to stdout for user confirmation
- **Create `.gitignore`** excluding: `node_modules/`, `dist/`, `.turbo/`, `.nx/`, `.env.local`, `*.log`
- **List generated commands**: `cd {name} && pnpm install && pnpm dev` (if not skipped)

### Step 5: Validate Output
- **Syntax validation**: Parse all generated `package.json` files as valid JSON
- **Path alias verification**: Ensure `tsconfig.base.json` paths map to actual `packages/*/src/` directories
- **Workspace protocol check**: Confirm internal dependencies use `workspace:*` (pnpm) or equivalent for selected package manager
- **ESLint config validation**: Verify `@typescript-eslint/parser` present and `parserOptions.project` points to `tsconfig.base.json`
- **Prettier config validation**: Ensure `.prettierrc.json` contains standard properties (semi, tabWidth, etc.)
- **Hook validation**: Confirm `.husky/pre-commit` script exists and references `lint-staged`
- **Report validation results**: Pass/fail with actionable error messages

## Programmatic Invocation

For use by `/implement` skill after scaffolding phase:

```javascript
const scaffoldProject = async (options) => {
  const {
    name,              // project name (kebab-case)
    type,              // 'monorepo' | 'app' | 'package'
    preset,            // 'react-vite' | 'node-ts' | 'typescript-library'
    packageManager = 'pnpm',
    output = './',
    skipInstall = false,
    features = []      // optional: ['husky', 'prettier', 'eslint']
  } = options;

  // Validate input
  validateProjectName(name);
  validatePreset(type, preset);

  // Load template
  const template = await loadTemplate(preset);

  // Generate files (returns Map of path → content)
  const files = template.generate({
    name,
    packageManager,
    features
  });

  // Write to disk
  const outputPath = `${output}/${name}`;
  for (const [path, content] of files) {
    await writeFile(`${outputPath}/${path}`, content);
  }

  // Create manifest
  const manifest = {
    preset,
    version: '0.0.1',
    generatedAt: new Date().toISOString(),
    packages: extractPackageNames(files)
  };
  await writeFile(`${outputPath}/.scaffolding-manifest.json`, JSON.stringify(manifest, null, 2));

  // Install dependencies if not skipped
  if (!skipInstall) {
    await installDependencies(outputPath, packageManager);
  }

  return {
    success: true,
    manifestPath: `${outputPath}/.scaffolding-manifest.json`,
    rootDir: outputPath
  };
};
```

## Anti-Patterns (NEVER Do These)

1. **Flat monorepo without directory layers** — Avoid: all packages at root. Use: `apps/` (deployables) and `packages/` (libraries) to enforce dependency direction and prevent cycles.

2. **Multiple lockfiles in monorepo** — Avoid: separate `package-lock.json` per package. Use: single `pnpm-lock.yaml` or `yarn.lock` at root for determinism across installations.

3. **Phantom dependencies via npm hoisting** — Avoid: relying on `@babel/core` being accessible without listing as dependency. Use: pnpm strict mode (`node-linker: strict` in `.npmrc`) to catch at runtime immediately.

4. **Circular dependency hell** — Avoid: packages/ui depending on apps/frontend. Use: unidirectional graph (shared-utils → ui → apps).

5. **Inconsistent TypeScript configs per package** — Avoid: custom `tsconfig.json` duplicating root settings. Use: `extends: "../../tsconfig.base.json"` with overrides only for preset-specific options.

6. **ESLint parser misconfiguration** — Avoid: default parser for TypeScript files. Use: `@typescript-eslint/parser` with `parserOptions.project: './tsconfig.base.json'` for type-aware linting.

7. **Skipping workspace protocol in dependencies** — Avoid: `"@myapp/ui": "^1.0.0"` for internal packages. Use: `"@myapp/ui": "workspace:*"` (pnpm) to reference live source, not published version.

8. **Missing or incomplete pre-commit hooks** — Avoid: committing unformatted code. Use: Husky + lint-staged to run ESLint/Prettier on staged files only, no full codebase scan.

9. **Conditional exports not declared** — Avoid: bundler confusion between ESM/CJS/browser/Node. Use: `exports` field in `package.json` with `{ "browser": "...", "require": "...", "import": "..." }` conditions.

10. **Over-reliance on root `node_modules`** — Avoid: installing all deps at root for 50+ packages. Use: Nx or Turborepo for distributed caching and dependency graph analysis.

11. **Mixing pnpm strict mode with unhandled peer dependencies** — Avoid: strict mode breaking legitimate peer dependency patterns. Use: mark transitive deps as `peerDependencies` and list in `peerDependenciesMeta.*.optional = true` when optional.

12. **Unvalidated remote templates** — Avoid: pulling templates from untrusted URLs without version pinning. Use: GitHub releases with semver tags (`owner/repo#v1.2.3`) and validated manifest.

13. **No post-scaffolding validation** — Avoid: assuming generated config is correct without testing. Use: `/implement` phase running `pnpm install && pnpm lint --max-warnings=0 && pnpm build` to catch config errors.

14. **Forgetting `.npmrc` strict mode setup** — Avoid: inconsistent pnpm behavior across CI/local machines. Use: committed `.npmrc` with `strict-peer-dependencies=true` and `node-linker=strict` for determinism.

## Success Criteria

- [x] Generated directory structure matches selected preset (file tree, package names, layer separation)
- [x] All `package.json` files parse as valid JSON and include required fields (`name`, `version`, `type`, `exports`)
- [x] `pnpm-workspace.yaml` correctly globs all workspace packages under `apps/**` and `packages/**`
- [x] TypeScript paths aliases in `tsconfig.base.json` resolve to actual `packages/*/src/` directories
- [x] ESLint config extends `typescript-eslint/recommended-type-checked` with proper `@typescript-eslint/parser` setup
- [x] Prettier config applied consistently at root with optional package-level overrides
- [x] `.husky/pre-commit` hook invokes `lint-staged` on staged files only (if Husky enabled)
- [x] `.gitignore` excludes standard build artifacts, node_modules, and cache directories
- [x] `.scaffolding-manifest.json` written with correct preset, version, timestamp, and package inventory
- [x] Internal package dependencies use `workspace:*` protocol (pnpm) or equivalent for selected package manager
- [x] File conflicts detected and user prompted for confirmation before overwriting
- [x] Remote template (if specified) fetched with version pinning and validated against manifest schema

## Integration Notes

- **Input from `/assess`**: Tech stack preferences (React, Node, monorepo size) inform preset selection
- **Output to `/implement`**: `.scaffolding-manifest.json` passed to installation/build phases with package metadata
- **CI/CD readiness**: Template includes Turborepo/Nx caching config for parallel execution in GitHub Actions/GitLab CI
- **Design system support**: Workspace structure accommodates component library in `packages/ui` with Storybook auto-discovery

## References

- [pnpm 9 Workspaces Guide](https://pristren.com/blog/pnpm-9-workspaces-guide/) — Symlink-based node_modules, version catalogs, 40-70% disk savings
- [JavaScript Monorepos 2026: Best Practices](https://www.pkgpulse.com/guides/javascript-monorepos-2026-best-practices-pitfalls) — Layering strategy, phantom dependency prevention
- [Nx with Vite Integration](https://nx.dev/nx-api/vite) — Modern build tooling for monorepo scaffolding
- [Node.js Packages exports/imports](https://nodejs.org/api/packages.html) — Conditional exports specification
- [Configure TypeScript monorepo with Turborepo](https://valcker.medium.com/unpacking-turborepo-configure-typescript-monorepo-with-eslint-prettier-and-webstorm-960bf9e65e60) — End-to-end monorepo setup patterns
