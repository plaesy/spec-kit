---
description: 'ReactJS development standards and best practices'
applyTo: '**/*.jsx, **/*.tsx, **/*.js, **/*.ts, **/*.css, **/*.scss'
---

# ReactJS Development Instructions

High-quality ReactJS apps with modern patterns, hooks, and best practices per https://react.dev.

## Project Context
- Latest React (19+), TypeScript when applicable, functional components with hooks as default
- Follow React's official style guide; modern build tools (Vite, CRA, or custom Webpack)
- Proper component composition and reusability patterns

## Development Standards

### Architecture
- Functional components with hooks as the primary pattern; composition over inheritance
- Organize by feature/domain for scalability
- Separate presentational and container components
- Custom hooks for reusable stateful logic; clear component hierarchies and data flow

### TypeScript Integration
- Interfaces for props/state/component definitions; proper types for event handlers/refs
- Generic components where appropriate; `strict` mode in `tsconfig.json`
- Use React's built-in types (`React.FC`, `React.ComponentProps`); union types for component variants/states

### Component Design
- Single responsibility principle; descriptive, consistent naming
- Prop validation (TypeScript or PropTypes); testable, reusable, small, single-concern components
- Composition patterns (render props, children as functions)

### State Management
- `useState` for local state, `useReducer` for complex state logic, `useContext` for cross-tree sharing
- External state management (Redux Toolkit, Zustand) for complex apps
- Proper state normalization; React Query or SWR for server state

### Hooks and Effects
- `useEffect` with correct dependency arrays (avoid infinite loops); cleanup functions to prevent leaks
- `useMemo`/`useCallback` for performance when needed; custom hooks for reusable logic
- Follow the rules of hooks (top level only); `useRef` for DOM access and mutable values

### Styling
- CSS Modules, Styled Components, or modern CSS-in-JS
- Mobile-first responsive design; BEM (or similar) class naming
- CSS custom properties for theming; consistent spacing/typography/color systems
- Accessibility: ARIA attributes, semantic HTML

### Performance Optimization
- `React.memo` for memoization; code splitting via `React.lazy`+`Suspense`
- Tree shaking, dynamic imports; `useMemo`/`useCallback` judiciously against unnecessary re-renders
- Virtual scrolling for large lists; profile with React DevTools

### Data Fetching
- Modern libraries (React Query, SWR, Apollo Client); proper loading/error/success states
- Handle race conditions and request cancellation; optimistic updates
- Caching strategies; graceful offline/network-error handling

### Error Handling
- Error Boundaries at the component level; proper error states in data fetching
- Fallback UI for errors; appropriate error logging
- Handle async errors in effects/event handlers; meaningful user-facing messages

### Forms and Validation
- Controlled components; validation via Formik/React Hook Form
- Proper submission/error-state handling; form accessibility (labels, ARIA)
- Debounced validation; handle file uploads and complex form scenarios

### Routing
- React Router; nested routes and route protection
- Proper route-param/query-string handling; lazy-loaded route-based code splitting
- Sound navigation patterns/back-button handling; breadcrumbs and nav state

### Testing
- Unit tests via React Testing Library, behavior not implementation details
- Jest as runner/assertions; integration tests for complex interactions
- Mock external deps/API calls; test accessibility and keyboard navigation

### Security
- Sanitize user input (prevent XSS); validate/escape data before rendering
- HTTPS for all external calls; proper auth/authz patterns
- Avoid sensitive data in localStorage/sessionStorage; CSP headers

### Accessibility
- Semantic HTML; proper ARIA attributes/roles; full keyboard navigation
- Alt text for images, descriptive icon text; proper color contrast
- Test with screen readers and a11y tools

## Implementation Process
1. Plan component architecture and data flow
2. Set up project structure/folder organization
3. Define TypeScript interfaces and types
4. Implement core components with styling
5. Add state management and data fetching
6. Implement routing and navigation
7. Add form handling and validation
8. Implement error handling and loading states
9. Add test coverage
10. Optimize performance and bundle size
11. Ensure accessibility compliance
12. Add documentation and code comments

## Additional Guidelines
- Naming: PascalCase components, camelCase functions
- Meaningful commits, clean git history
- Code splitting/lazy loading strategies
- JSDoc for complex components/custom hooks
- ESLint + Prettier for consistent formatting
- Keep dependencies current, audit for vulnerabilities
- Proper env configuration per deployment stage
- React Developer Tools for debugging/perf analysis

## Common Patterns
Higher-Order Components (cross-cutting concerns), render props (composition), compound components (related functionality), Provider pattern (context-based state sharing), Container/Presentational separation, custom hooks (reusable logic extraction).
