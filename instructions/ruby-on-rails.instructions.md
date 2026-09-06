---
description: 'Ruby on Rails coding conventions and guidelines'
applyTo: '**/*.rb'
---

# Ruby on Rails

## General Guidelines

- Follow RuboCop Style Guide; use `rubocop`/`standardrb`/`rufo` for consistent formatting
- snake_case variables/methods, CamelCase classes/modules
- Short, focused methods - early returns, guard clauses, private methods to reduce complexity
- Meaningful names over short/generic ones
- Comment only when necessary, not the obvious
- Single Responsibility Principle for classes/methods/modules
- Composition over inheritance - extract reusable logic into modules/services
- Thin controllers - business logic in models/services/command-query objects
- "Fat model, skinny controller" applied thoughtfully with clean abstractions
- Extract business logic into service objects for reuse/testability
- Partials or view components to reduce view duplication
- `unless` for negative conditions, avoid pairing with `else`
- Avoid deep conditional nesting - guard clauses and method extraction
- Safe navigation (`&.`) instead of multiple `nil` checks
- `.present?`/`.blank?`/`.any?` over manual nil/empty checks
- RESTful conventions in routing/controller actions
- Rails generators for consistent resource scaffolding
- Strong parameters to whitelist attributes securely
- Enums and typed attributes for model clarity and validations
- DB-agnostic migrations, avoid raw SQL where possible
- Indexes on foreign keys and frequently queried columns
- `null: false`/`unique: true` at the DB level, not just in models
- `find_each` for large-dataset iteration (memory)
- Scope queries in models or use query objects for clarity/reuse
- `before_action` sparingly - no business logic in callbacks
- `Rails.cache` for expensive computations or frequently accessed data
- `Rails.root.join(...)` for file paths, never hardcoded
- `class_name`/`foreign_key` in associations for explicit relationships
- Secrets/config out of the codebase - `Rails.application.credentials` or ENV vars
- Isolated unit tests for models/services/helpers
- Request/system tests for end-to-end logic
- ActiveJob for non-blocking operations (emails, API calls)
- `FactoryBot` (RSpec) or fixtures (Minitest) for clean test data
- No `puts` for debugging - `byebug`, `pry`, or logger utilities
- Document complex code paths/methods with YARD or RDoc

## App Directory Structure

- `app/services` - service objects encapsulating business logic
- `app/forms` - form objects for validation/submission logic
- `app/serializers` - JSON serializers for API responses
- `app/policies` - authorization policies for resource access
- `app/graphql` - GraphQL schemas, queries, mutations
- `app/validators` - custom validators
- `app/queries` - complex ActiveRecord queries, isolated for reuse/testability
- `app/types` - custom data types/coercion logic extending ActiveModel type behavior

## Commands

- `rails generate` - new models, controllers, migrations
- `rails db:migrate` - apply migrations
- `rails db:seed` - populate initial data
- `rails db:rollback` - revert last migration
- `rails console` - REPL
- `rails server` - dev server
- `rails test` - run test suite
- `rails routes` - list defined routes
- `rails assets:precompile` - compile assets for production


## API Development Best Practices

- `resources` routing for RESTful conventions
- Namespaced routes (e.g. `/api/v1/`) for versioning/forward compatibility
- `ActiveModel::Serializer` or `fast_jsonapi` for consistent response serialization
- Proper HTTP status codes per response (200 OK, 201 Created, 422 Unprocessable Entity, etc.)
- `before_action` filters for loading/authorizing resources, not business logic
- Pagination (`kaminari`/`pagy`) for large-dataset endpoints
- Rate limiting/throttling on sensitive endpoints (`rack-attack` or middleware)
- Structured JSON error responses with codes, messages, details
- Sanitize/whitelist input via strong parameters
- Custom serializers/presenters to decouple internal logic from response formatting
- Avoid N+1 queries - `includes` for eager loading
- Background jobs for non-blocking tasks (emails, external API syncs)
- Log request/response metadata for debugging/observability/auditing
- Document endpoints via OpenAPI (Swagger), `rswag`, or `apipie-rails`
- CORS headers (`rack-cors`) for cross-origin access when needed
- Never expose sensitive data in API responses/error messages

## Frontend Development Best Practices

- `app/javascript` for JS packs/modules/frontend logic (Rails 6+, Webpacker/esbuild)
- Structure JS by component/domain, not file type
- Hotwire (Turbo + Stimulus) for real-time updates, minimal JS in Rails-native apps
- Stimulus controllers to bind behavior to HTML declaratively
- SCSS modules, Tailwind, or BEM under `app/assets/stylesheets`
- Extract repetitive markup into partials/components
- Semantic HTML, accessibility (a11y) best practices throughout
- No inline JS/styles - separate `.js`/`.scss` files
- Optimize assets (images/fonts/icons) via the asset pipeline/bundlers for caching/compression
- `data-*` attributes to bridge frontend interactivity with Rails HTML and Stimulus
- Test frontend with system tests (Capybara) or Cypress/Playwright integration tests
- Environment-specific asset loading (avoid unnecessary prod scripts/styles)
- Follow a design system/component library for consistent, scalable UI
- Optimize time-to-first-paint via lazy loading, Turbo Frames, deferred JS

## Testing Guidelines

- Unit tests for models: `test/models` (Minitest) or `spec/models` (RSpec)
- Fixtures (Minitest) or `FactoryBot` (RSpec) for clean, consistent test data
- Controller specs under `test/controllers` or `spec/requests` for RESTful API behavior
- `before` blocks (RSpec) / `setup` (Minitest) for common test-data init
- No hitting external APIs in tests - `WebMock`, `VCR`, or `stub_request`
- System tests (Minitest) or feature specs with Capybara (RSpec) for full user flows
- Isolate slow/expensive tests (external services, file uploads) into separate types/tags
- `SimpleCov` for coverage tracking
- No `sleep` in tests - `perform_enqueued_jobs` (Minitest) or `ActiveJob::TestHelper` (RSpec)
- Database cleaning (`rails test:prepare`, `DatabaseCleaner`, transactional fixtures) between tests
- Test background jobs via `ActiveJob::TestHelper` or `have_enqueued_job` matchers
- Consistent CI test runs (GitHub Actions, CircleCI)
- Custom matchers (RSpec) / custom assertions (Minitest) for reusable, expressive tests
- Tag tests by type (`:model`, `:request`, `:feature`) for targeted runs
- Avoid brittle tests - don't rely on specific timestamps/randomized data/order unless necessary
- Integration tests for end-to-end flows across model/view/controller
- Keep tests fast, reliable, as DRY as production code
