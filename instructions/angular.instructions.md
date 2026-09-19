---
description: 'Angular-specific coding standards and best practices'
applyTo: '**/*.ts, **/*.html, **/*.scss, **/*.css'
---

# Angular Development Instructions

High-quality Angular apps with TypeScript, using Angular Signals for state, per https://angular.dev best practices.

## Project Context
- Latest Angular (standalone components by default)
- TypeScript for type safety, Angular CLI for setup/scaffolding
- Follow the Angular Style Guide (https://angular.dev/style-guide)
- Angular Material or another modern UI library for consistent styling (if specified)

## Development Standards

### Architecture
- Standalone components unless modules are explicitly required
- Organize by feature module/domain for scalability; lazy-load feature modules
- Use Angular's DI system effectively
- Clear separation of concerns (smart vs presentational components)

### TypeScript
- Strict mode in `tsconfig.json`
- Clear interfaces/types for components, services, models
- Type guards and union types for robust checking
- RxJS error handling (`catchError`)
- Typed reactive forms (`FormGroup`, `FormControl`)

### Component Design
- Follow Angular lifecycle-hook best practices
- Angular >= 19: `input()`/`output()`/`viewChild()`/`viewChildren()`/`contentChild()` functions instead of decorators; otherwise decorators
- `OnPush` change detection for performance where applicable
- Keep templates clean, logic in component classes/services
- Directives and pipes for reusable functionality

### Styling
- Component-level CSS encapsulation (default: `ViewEncapsulation.Emulated`)
- SCSS with consistent theming
- Responsive design via CSS Grid/Flexbox/Angular CDK Layout
- Follow Angular Material theming guidelines if used
- Accessibility (a11y): ARIA attributes, semantic HTML

### State Management
- Angular Signals for reactive state in components/services
- `signal()`/`computed()`/`effect()` for reactive updates
- Writable signals for mutable state, computed for derived state
- Loading/error states via signals with proper UI feedback
- `AsyncPipe` for observables in templates when combining signals with RxJS

### Data Fetching
- `HttpClient` with proper typing
- RxJS operators for transformation/error handling
- `inject()` for DI in standalone components
- Caching (`shareReplay` for observables)
- Store API responses in signals for reactive updates
- Global interceptors for consistent API error handling

### Security
- Sanitize user input via Angular's built-in sanitization
- Route guards for authentication/authorization
- `HttpInterceptor` for CSRF protection and auth headers
- Validate forms with reactive forms + custom validators
- Follow Angular security best practices (avoid direct DOM manipulation)

### Performance
- Production builds (`ng build --prod`)
- Lazy-load routes to shrink initial bundle
- `OnPush` + signals for fine-grained reactivity
- `trackBy` in `ngFor` loops
- SSR/SSG via Angular Universal if specified

### Testing
- Unit tests (Jasmine/Karma) for components, services, pipes
- `TestBed` with mocked dependencies
- Test signal-based state updates with Angular's testing utilities
- E2E tests (Cypress/Playwright) if specified
- Mock HTTP via `HttpClientTestingModule`
- High coverage on critical functionality

## Implementation Process
1. Plan project structure and feature modules
2. Define TypeScript interfaces and models
3. Scaffold components/services/pipes via Angular CLI
4. Implement data services and API integrations with signal-based state
5. Build reusable components with clear inputs/outputs
6. Add reactive forms and validation
7. Apply SCSS styling and responsive design
8. Implement lazy-loaded routes and guards
9. Add error handling and loading states via signals
10. Write unit and E2E tests
11. Optimize performance and bundle size

## Additional Guidelines
- Angular naming conventions (`feature.component.ts`, `feature.service.ts`)
- Angular CLI for boilerplate generation
- JSDoc comments on components/services
- WCAG 2.1 accessibility compliance where applicable
- Built-in i18n if specified
- DRY via reusable utilities and shared modules
- Use signals consistently for state management
