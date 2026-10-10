---
description: 'Vue development standards - Composition API, Single-File Components, reactivity, props/events, state, testing, and performance'
applyTo: '**/*.vue,**/vite.config.*,**/vue.config.js,**/nuxt.config.*'
---

# Vue Development Instructions

Apply `.plaesy/instructions/ui-ux-design-principles.md` (hierarchy, feedback,
affordance, navigation) to every component you build — this file covers
framework mechanics, not interaction design.

Component-first frontend work: declarative templates plus a reactivity system, per
the [Vue introduction](https://vuejs.org/guide/introduction.html) (retrieved
2026-09-27). Vue 2 reached end of life on 2023-12-31, so Vue 2 migration is out of
scope for new work unless the project is explicitly a Vue 2 codebase.

Source (retrieved 2026-09-27): [Vue Guide — Introduction](https://vuejs.org/guide/introduction.html)
for the two core features (declarative rendering, reactivity), Single-File
Components, and the Options API / Composition API choice. The guide also notes an
LLM-optimised Markdown version of each page at the same path plus `.md`. Vue's
official libraries are Vue Router and Pinia; anything else in a Vue project is a
third-party dependency and gets the same scrutiny as any other.

## Project Context

- Single-File Components (`*.vue`) are the recommended authoring format when a build
  setup exists: one file holds the template, the logic, and the scoped styles
- Two API styles exist and both are fully supported; Vue's own guidance is Options
  API for progressive enhancement / no build step, and Composition API + SFCs for
  full applications
- TypeScript is a first-class option via the Composition API (Vue maintains a
  dedicated TS guide and `vue-tsc`); adopt it if the project already does
- Build tooling is the project's choice (Vite is the common one); do not migrate a
  working build to a different tool inside a feature change

## Development Standards

### Choosing an API style

- Pick one style per component and per file; do not mix Options API and Composition
  API inside one component
- `<script setup>` is a compile-time hint: it makes Composition API usable with less
  boilerplate and puts top-level bindings directly in the template's scope
- Migration between the two styles is a deliberate, separate change — never a
  side-effect of an unrelated edit

### Components and props

- One component, one job; if the name needs "and" (e.g. `UserListAndForm`), split it
- Props are the component's declared input; declare their types. Use the
  `defineProps` macro in `<script setup>`
- Declare a default for every object/array prop — a missing default means a shared
  mutable literal between instances
- Props flow one way (parent → child). Child → parent is emitted events, named for
  what happened (`submit`, `remove-item`), not for the button that was clicked
- Fallthrough attributes arrive on the root element; a component with multiple roots
  must declare `inheritAttrs: false` and bind attributes explicitly
- `v-model` on a component is sugar for a `modelValue` prop plus an
  `update:modelValue` event — a custom `v-model:title` needs the matching
  `title`/`update:title` pair
- Slots for composition; avoid prop-drilling data through five levels when
  `provide`/`inject` fits

### Reactivity

- `ref` for primitives and for anything the template must unwrap; `reactive` for a
  group of related fields
- Never destructure a `reactive` object in a way that loses reactivity; use `toRefs`
  when the individual refs are needed
- `computed` for derived state — a value derived from other state belongs in a
  computed property, not in a watcher that writes state back
- `watch` only for side effects (fetch on a change, log, persist). Watching a value
  and assigning it in the callback is a computed property written the long way
- Side effects need teardown: cancel in-flight requests, clear timers, and dispose
  observers in `onUnmounted`/`onBeforeUnmount`, or the component leaks
- Keep the state close to where it is used; lift only when a second consumer exists

### State management

- Local state: `ref`/`reactive` inside the component
- Shared client state: a store (Pinia) — one store per domain concept, actions for
  anything with a side effect, getters for derived values
- Server state (fetched collections, caching, invalidation) belongs in a data layer
  that handles request cancellation, loading and error states; do not hand-roll it
  into components
- Never mutate another component's state; communicate via props, events, or a store
  action

### Templates and styling

- Semantic HTML first; ARIA only where semantics cannot express the intent
- `v-if` vs `v-show` by cost, not by preference: `v-show` keeps the node in the DOM
- Always give `v-for` a stable `key` — an array index as key is a bug the moment the
  list reorders
- Component-scoped styles plus design tokens; no hardcoded colours or spacing values
  in components
- Async components need explicit `loading` and `error` states
- Form input bound with `v-model`; validation errors announced and associated with
  their field

### Performance

- `computed` caches; a method called in a template re-runs on every re-render
- `v-once` / `v-memo` for genuinely static subtrees; list virtualisation for long
  lists
- Code-split routes and heavy components (dynamic `import()`); an async component
  needs a `Suspense` boundary or an explicit fallback
- Measure with the Vue devtools performance panel before optimising; most cost is in
  fewer, larger components rather than in micro-optimising render functions

### Testing

- Component tests assert on rendered output and emitted events, not on internals
- Mount with the project's test utils (`@vue/test-utils`); query by role/label first,
  by test id only as a last resort
- Every component with a branch gets a test for each branch: default, populated,
  loading, error, empty
- Test reactive side effects through the rendered outcome, and assert that teardown
  runs (no leaked timers/requests after unmount)

### Security

- Template interpolation escapes, but `v-html` does not — it is a raw-HTML sink.
  Sanitise before it ever reaches `v-html`
- Never put secrets in client-side state; anything in the bundle is public, including
  `import.meta.env` values shipped to the client
- Validate on the server as well — the client is a convenience, not a boundary
- Guard routes in the router and re-check authorisation server-side per request
- Third-party scripts/plugins are supply-chain risk: pin versions, review what they
  do in the browser

## Implementation Process

1. Confirm the component's contract: props in, events out, slots
2. Decide the API style for the file and stay inside it
3. Write the template with semantic markup and stable keys
4. Add state (`ref`/`reactive`), then `computed` for everything derived
5. Add `watch` only for the side effects that genuinely need it, with teardown
6. Style from the token set; keep styles scoped
7. Cover default / loading / error / empty states
8. Test the branches and the emitted events
9. Check accessibility: roles, labels, keyboard order, focus
10. Split and lazy-load if the component grew past one screen's worth of template

## Common Pitfalls to Avoid

Mutating a prop; mixing Options and Composition API in one component; watching a
value and assigning it back instead of computing it; using the array index as a
`v-for` key; side effects without cleanup on unmount; `v-html` on untrusted input;
a shared mutable default on an object/array prop; a computed property replaced by a
method call in the template; a store mutated directly from a component; forgetting
loading and error states; a `:key` on a `v-if` branch that resets the whole subtree
needlessly.

## Usage Example

```vue

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({
  userId: { type: String, required: true },
  tags: { type: Array, default: () => [] },
})

const user = ref(null)
const error = ref('')
let controller = null

const displayName = computed(() => user.value?.name ?? 'Unknown user')

watch(() => props.userId, load, { immediate: true })

async function load(id) {
  controller?.abort()
  controller = new AbortController()
  error.value = ''
  try {
    const res = await fetch(`/api/users/${encodeURIComponent(id)}`, {
      signal: controller.signal,
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    user.value = await res.json()
  } catch (err) {
    if (err.name !== 'AbortError') error.value = 'Could not load user'
  }
}

function onFocus() { console.log('focused', displayName.value) }

onMounted(() => window.addEventListener('focus', onFocus))
onUnmounted(() => {
  controller?.abort()
  window.removeEventListener('focus', onFocus)
})
</script>

<template>
  <p v-if="error" role="alert">{{ error }}</p>
  <p v-else-if="user">{{ displayName }} ({{ tags.length }} tags)</p>
  <p v-else>Loading…</p>
</template>
```

## References

- Introduction — https://vuejs.org/guide/introduction.html
- Guide index (Reactivity Fundamentals, Components, Single-File Components,
  Testing, Performance, Security, TypeScript) — https://vuejs.org/guide/
