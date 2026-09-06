---
applyTo: '**/*.ts, **/*.js, **/*.json, **/*.spec.ts, **/*.e2e-spec.ts, **/*.py, **/*.java, **/*.cs, **/*.go, **/*.rb, **/*.php, **/*.sql, **/*.html, **/*.css, **/*.scss'
description: 'The most comprehensive, practical, and engineer-authored performance optimization instructions for all languages, frameworks, and stacks. Covers frontend, backend, and database best practices with actionable guidance, scenario-based checklists, troubleshooting, and pro tips.'
---

# Performance Optimization Best Practices

## General Principles

- **Measure first, optimize second**: profile before optimizing (Chrome DevTools, Lighthouse, New Relic, Datadog, Py-Spy, language profilers) - don't guess
- **Optimize the common case**: focus on frequently-executed paths, not rare edge cases
- **Avoid premature optimization**: write clear code first, optimize only when needed
- **Minimize resource usage**: memory, CPU, network, disk - ask "can this be done with less?"
- **Prefer simplicity**: simple algorithms/data structures are often fastest, don't over-engineer
- **Document performance assumptions**: comment performance-critical or non-obvious optimizations
- **Understand the platform**: performance characteristics differ by language/framework/runtime
- **Automate performance testing**: integrate into CI/CD to catch regressions early
- **Set performance budgets**: define and enforce limits for load time, memory, API latency

---

## Frontend Performance

### Rendering and DOM
- Batch DOM updates (avoid updating in a loop - build a document fragment, append once)
- Virtual DOM frameworks (React/Vue): avoid unnecessary re-renders (`React.memo`, `useMemo`, `useCallback`)
- Stable keys in lists - avoid array indices unless the list is static
- Avoid inline styles (triggers layout thrashing) - prefer CSS classes
- CSS transitions/animations over JS for GPU-accelerated smoothness
- Defer non-critical rendering with `requestIdleCallback`

### Asset Optimization
- Compress images (ImageOptim, Squoosh, TinyPNG), prefer WebP/AVIF
- SVGs for icons - scale well, often smaller than PNGs
- Bundle/minify with Webpack, Rollup, or esbuild; enable tree-shaking
- Long-lived cache headers for static assets, cache-bust on updates
- Lazy-load images (`loading="lazy"`) and JS modules (dynamic imports)
- Font optimization: subset to needed character sets, `font-display: swap`

### Network Optimization
- Reduce HTTP requests: combine files, sprites, inline critical CSS
- Enable HTTP/2 and HTTP/3 for multiplexing and lower latency
- Client-side caching: Service Workers, IndexedDB, localStorage
- Serve static assets from a CDN, use multiple for redundancy
- `defer`/`async` for non-critical JS
- `<link rel="preload">` / `rel="prefetch">` for critical resources

### JavaScript Performance
- Offload heavy computation to Web Workers, don't block the main thread
- Debounce/throttle scroll, resize, and input handlers
- Clean up event listeners/intervals/DOM references - check for detached nodes in dev tools
- Maps/Sets for lookups, TypedArrays for numeric data
- Avoid global variables (memory leaks, unpredictable performance)
- Avoid deep object cloning - shallow copy, or `cloneDeep` only when necessary

### Accessibility and Performance
- Don't over-fire ARIA updates; use semantic HTML for both a11y and performance
- Avoid rapid DOM updates that overwhelm assistive tech

### Framework-Specific Tips
#### React
- `React.memo`, `useMemo`, `useCallback` to avoid unnecessary renders
- Code-split large components (`React.lazy`, `Suspense`)
- Avoid anonymous functions in render (new reference every render)
- `ErrorBoundary` for graceful error handling
- Profile with React DevTools Profiler

#### Angular
- `OnPush` change detection for infrequently-updating components
- Move complex expressions out of templates into the component class
- `trackBy` in `ngFor` for efficient list rendering
- Lazy-load modules/components via Angular Router
- Profile with Angular DevTools

#### Vue
- Computed properties over methods in templates (cached)
- `v-show` for frequent toggling, `v-if` otherwise
- Lazy-load components/routes via Vue Router
- Profile with Vue Devtools

### Common Frontend Pitfalls
Large initial JS bundles; uncompressed/outdated image formats; leaked event listeners; over-reliance on third-party libs for simple tasks; ignoring mobile performance (test on real devices).

### Frontend Troubleshooting
Chrome DevTools Performance tab (slow frames), Lighthouse (actionable audit), WebPageTest (real-world load), Core Web Vitals (LCP, FID, CLS).

---

## Backend Performance

### Algorithm and Data Structure Optimization
- Right data structure for the job: arrays (sequential), hash maps (lookup), trees (hierarchical)
- Efficient algorithms: binary search, quicksort, hash-based where appropriate
- Avoid O(n^2)+ - profile nested loops/recursion, refactor complexity
- Batch processing to reduce overhead (e.g. bulk DB inserts)
- Streaming APIs for large datasets instead of loading everything into memory

### Concurrency and Parallelism
- Async I/O (async/await, callbacks, event loops) to avoid blocking threads
- Thread/worker pools to manage concurrency, avoid resource exhaustion
- Locks/semaphores/atomics to avoid race conditions
- Batch network/DB calls to reduce round trips
- Backpressure in queues/pipelines to avoid overload

### Caching
- In-memory caches (Redis, Memcached) for hot data
- Cache invalidation: TTL, event-based, or manual - stale cache is worse than none
- Distributed caching for multi-server setups - watch consistency
- Lock/request coalescing to prevent cache-stampede/thundering herd
- Don't cache everything - some data is too volatile/sensitive

### API and Network
- Minimize payloads: JSON, gzip/Brotli compression, no unnecessary data
- Paginate large result sets; cursors for real-time data
- Rate limiting against abuse/overload
- Connection pooling for DBs and external services
- HTTP/2, gRPC, or WebSockets for high-throughput/low-latency

### Logging and Monitoring
- Minimize logging in hot paths (slows critical code)
- Structured (JSON/key-value) logs for easier parsing
- Monitor latency, throughput, error rates, resource usage (Prometheus, Grafana, Datadog)
- Alert on performance regressions and resource exhaustion

### Language/Framework-Specific Tips
#### Node.js
- Async APIs, never `fs.readFileSync` in production (blocks event loop)
- Clustering/worker threads for CPU-bound tasks
- Limit concurrent open connections
- Streams for large file/network data
- Profile with `clinic.js`, `node --inspect`, Chrome DevTools

#### Python
- Built-in structures (`dict`, `set`, `deque`) for speed
- Profile with `cProfile`, `line_profiler`, `Py-Spy`
- `multiprocessing`/`asyncio` for parallelism
- Avoid GIL bottlenecks in CPU-bound code - C extensions or subprocesses
- `lru_cache` for memoization

#### Java
- Efficient collections (`ArrayList`, `HashMap`)
- Profile with VisualVM, JProfiler, YourKit
- Thread pools (`Executors`) for concurrency
- Tune JVM heap/GC (`-Xmx`, `-Xms`, `-XX:+UseG1GC`)
- `CompletableFuture` for async

#### .NET
- `async/await` for I/O-bound ops
- `Span<T>`/`Memory<T>` for efficient memory access
- Profile with dotTrace, Visual Studio Profiler, PerfView
- Pool objects/connections where appropriate
- `IAsyncEnumerable<T>` for streaming

### Common Backend Pitfalls
Sync/blocking I/O in web servers; no connection pooling; over-caching sensitive/volatile data; unhandled errors in async code; no monitoring/alerting on regressions.

### Backend Troubleshooting
Flame graphs (CPU usage), distributed tracing (OpenTelemetry, Jaeger, Zipkin) for cross-service latency, heap dumps/memory profilers for leaks, log slow queries/API calls.

---

## Database Performance

### Query Optimization
- Index frequently queried/filtered/joined columns; monitor and drop unused indexes
- Avoid `SELECT *` - select only needed columns
- Parameterized queries (prevents injection, improves plan caching)
- Analyze/optimize query plans (`EXPLAIN`)
- Avoid N+1 queries - joins or batch queries
- `LIMIT`/`OFFSET` or cursors for large tables

### Schema Design
- Normalize to reduce redundancy; denormalize for read-heavy workloads if needed
- Efficient data types, appropriate constraints
- Partition large tables for scalability/manageability
- Archive/purge old data regularly
- Foreign keys for integrity - weigh performance trade-off in high-write scenarios

### Transactions
- Keep transactions short to reduce lock contention
- Use the lowest isolation level that meets consistency needs
- Avoid long-running transactions (block other ops, increase deadlocks)

### Caching and Replication
- Read replicas for read-heavy workloads - monitor replication lag
- Cache query results (Redis, Memcached) for frequent queries
- Choose write-through vs write-behind per consistency needs
- Shard to distribute data across servers

### NoSQL Databases
- Design schema for actual access patterns
- Avoid hot partitions - distribute writes/reads evenly
- Watch for unbounded array/document growth
- Shard and replicate for scalability/availability
- Understand eventual vs strong consistency, choose appropriately

### Common Database Pitfalls
Missing/unused indexes; `SELECT *` in production; unmonitored slow queries; ignored replication lag; no data archiving.

### Database Troubleshooting
Slow query logs, `EXPLAIN` for query plans, cache hit/miss ratios, DB-specific monitoring (pg_stat_statements, MySQL Performance Schema).

---

## Code Review Checklist for Performance

- [ ] Any obvious algorithmic inefficiencies (O(n^2)+)?
- [ ] Data structures appropriate for their use?
- [ ] Unnecessary/repeated computation?
- [ ] Caching used appropriately, invalidation handled correctly?
- [ ] DB queries optimized, indexed, free of N+1?
- [ ] Large payloads paginated/streamed/chunked?
- [ ] Memory leaks or unbounded resource usage?
- [ ] Network requests minimized, batched, retried on failure?
- [ ] Assets optimized, compressed, served efficiently?
- [ ] Blocking operations in hot paths?
- [ ] Logging in hot paths minimized and structured?
- [ ] Performance-critical paths documented and tested?
- [ ] Automated tests/benchmarks for performance-sensitive code?
- [ ] Alerts for performance regressions?
- [ ] Anti-patterns present (SELECT *, blocking I/O, global variables)?

---

## Advanced Topics

### Profiling and Benchmarking
- Language-specific profilers (Chrome DevTools, Py-Spy, VisualVM, dotTrace) to find bottlenecks
- Microbenchmarks for critical paths (`benchmark.js`, `pytest-benchmark`, JMH)
- A/B or canary releases to measure real-world impact
- CI/CD performance testing (k6, Gatling, Locust)

### Memory Management
- Release resources (files, sockets, DB connections) promptly
- Object pooling for frequently created/destroyed objects
- Monitor heap usage/GC, tune settings for the workload
- Leak detection tools (Valgrind, LeakCanary, Chrome DevTools)

### Scalability
- Stateless services, sharding/partitioning, load balancers for horizontal scaling
- Cloud auto-scaling with sensible thresholds
- Identify/address single points of failure
- Idempotent operations, retries, circuit breakers for distributed systems

### Security and Performance
- Hardware-accelerated, well-maintained crypto libraries
- Efficient input validation - avoid regexes in hot paths
- Rate limiting against DoS without harming legitimate users

### Mobile Performance
- Lazy-load features, defer heavy work, minimize initial bundle size
- Responsive images, compressed assets for mobile bandwidth
- SQLite/Realm or platform-optimized storage
- Profile with Android Profiler, Instruments (iOS), Firebase Performance Monitoring

### Cloud and Serverless
- Minimize dependencies, keep functions warm (cold starts)
- Tune memory/CPU allocation for serverless functions
- Managed caching/queues/DBs for scalability
- Monitor and optimize cloud cost as a performance metric

---

## Practical Examples

### Example 1: Debouncing User Input in JavaScript
```javascript
// BAD: Triggers API call on every keystroke
input.addEventListener('input', (e) => {
  fetch(`/search?q=${e.target.value}`);
});

// GOOD: Debounce API calls
let timeout;
input.addEventListener('input', (e) => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    fetch(`/search?q=${e.target.value}`);
  }, 300);
});
```

### Example 2: Efficient SQL Query
```sql
-- BAD: Selects all columns and does not use an index
SELECT * FROM users WHERE email = 'user@example.com';

-- GOOD: Selects only needed columns and uses an index
SELECT id, name FROM users WHERE email = 'user@example.com';
```

### Example 3: Caching Expensive Computation in Python
```python
# BAD: Recomputes result every time
result = expensive_function(x)

# GOOD: Cache result
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_function(x):
    ...
result = expensive_function(x)
```

### Example 4: Lazy Loading Images in HTML
```html
<!-- BAD: Loads all images immediately -->
<img src="large-image.jpg" />

<!-- GOOD: Lazy loads images -->
<img src="large-image.jpg" loading="lazy" />
```

### Example 5: Asynchronous I/O in Node.js
```javascript
// BAD: Blocking file read
const data = fs.readFileSync('file.txt');

// GOOD: Non-blocking file read
fs.readFile('file.txt', (err, data) => {
  if (err) throw err;
  // process data
});
```

### Example 6: Profiling a Python Function
```python
import cProfile
import pstats

def slow_function():
    ...

cProfile.run('slow_function()', 'profile.stats')
p = pstats.Stats('profile.stats')
p.sort_stats('cumulative').print_stats(10)
```

### Example 7: Using Redis for Caching in Node.js
```javascript
const redis = require('redis');
const client = redis.createClient();

function getCachedData(key, fetchFunction) {
  return new Promise((resolve, reject) => {
    client.get(key, (err, data) => {
      if (data) return resolve(JSON.parse(data));
      fetchFunction().then(result => {
        client.setex(key, 3600, JSON.stringify(result));
        resolve(result);
      });
    });
  });
}
```

---

## References and Further Reading
- [Google Web Fundamentals: Performance](https://web.dev/performance/)
- [MDN Web Docs: Performance](https://developer.mozilla.org/en-US/docs/Web/Performance)
- [OWASP: Performance Testing](https://owasp.org/www-project-performance-testing/)
- [Microsoft Performance Best Practices](https://learn.microsoft.com/en-us/azure/architecture/best-practices/performance)
- [PostgreSQL Performance Optimization](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [MySQL Performance Tuning](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [Node.js Performance Best Practices](https://nodejs.org/en/docs/guides/simple-profiling/)
- [Python Performance Tips](https://docs.python.org/3/library/profile.html)
- [Java Performance Tuning](https://www.oracle.com/java/technologies/javase/performance.html)
- [.NET Performance Guide](https://learn.microsoft.com/en-us/dotnet/standard/performance/)
- [WebPageTest](https://www.webpagetest.org/)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)
- [k6 Load Testing](https://k6.io/)
- [Gatling](https://gatling.io/)
- [Locust](https://locust.io/)
- [OpenTelemetry](https://opentelemetry.io/)
- [Jaeger](https://www.jaegertracing.io/)
- [Zipkin](https://zipkin.io/)
