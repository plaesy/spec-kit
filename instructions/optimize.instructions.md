---
description: "Performance optimization and code refactoring"
---

# `/optimize` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Comprehensive code and design optimization: improve performance, design quality, code efficiency, and user experience

## Protocol
**Optimize ALL aspects systematically**:
1. Performance Analysis - Collect baseline metrics, identify bottlenecks
2. Design Review - Assess UI/UX improvements (if applicable)
3. Code Quality - Refactor for efficiency, readability, maintainability
4. Planning - Prioritize optimizations by impact vs effort
5. Implementation - Apply optimizations incrementally with validation
6. Validation - Measure improvements against baseline

**MANDATORY NEXT STEP**: After `/optimize` completes, run `/assess` to verify all improvements worked and no regressions introduced

## Validation Checklist (Before Running)
- ✅ Code passes quality gates (assess score ≥80)
- ✅ Tests passing (all tests green)
- ✅ Baseline metrics recorded (before optimization)
- ✅ Production traffic representable in tests

## Performance Targets
- **Response Time** <200ms (APIs)
- **Throughput** >1000 req/s
- **Memory Usage** <512MB baseline
- **CPU Utilization** <70% average
- **Database Query** <100ms

## Anti-Patterns (NEVER Do These)
- ❌ **Never optimize prematurely** - Measure baseline first
- ❌ **Never skip regression testing** - Test all paths after optimization
- ❌ **Never change behavior for speed** - Correctness > Performance
- ❌ **Never optimize without metrics** - All improvements must be measurable

## Error Recovery
**If optimization fails:**
- Report which optimization caused regression
- Revert change + document why it failed
- Try different approach

**If metrics unavailable:**
- Use profiling tools (Python cProfile, Node.js clinic, Go pprof)
- Report: "Metrics from [tool]"

## Completion Format
✅ Optimization Complete
- Performance improvements: [metrics with % change]
- Design improvements: [if UI: consistency, tokens, dark mode, accessibility]
- Code improvements: [refactoring, duplication removed, type safety]
- No regressions: [Y/N - confirmed via testing]
- **MANDATORY NEXT**: Run `/assess` to verify all improvements + no new issues

## Optimization Priorities (All Projects)

### Performance Optimization
1. **Database** - Queries, indexes, connection pooling (backend)
2. **Caching** - Application-level, distributed, CDN
3. **Algorithms** - Complexity reduction, better data structures
4. **I/O Operations** - Async processing, batching
5. **Memory Management** - Object pooling, GC tuning
6. **Component Performance** - Reduce re-renders, lazy load, code-split (frontend)
7. **Image Optimization** - WebP, responsive, lazy load (frontend)

### Design Optimization (If UI/Frontend)
1. **Component Reuse** - Consolidate components, increase design system adoption
2. **Design Tokens** - Standardize colors, spacing, typography
3. **Visual Consistency** - Ensure consistent patterns across UI
4. **Dark Mode** - Optimize color contrast and readability
5. **Animation** - 60fps target, optimize transitions, reduce jank
6. **Accessibility** - WCAG AA compliance, keyboard navigation, screen reader

### Code Quality Optimization (All Projects)
1. **Refactoring** - Simplify complex functions, improve readability
2. **Duplication Removal** - Consolidate similar code patterns
3. **Type Safety** - Add type annotations, reduce any types (TypeScript/Python)
4. **Error Handling** - Improve error messages, edge case handling
5. **Testing** - Increase coverage, improve test quality
6. **Documentation** - Update code comments, clarify complex logic

## Context7 Protocol (Required)
**For optimization efforts**:
1. `mcp__context7__resolve-library-id` for technologies being optimized
2. `mcp__context7__get-library-docs` for optimization techniques
3. **Citation**: "Optimization based on Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"

## Optimization Areas

### Database Optimizations
- **Query Optimization**: Add indexes, optimize queries, fix N+1 problems
- **Connection Pooling**: Optimize pool size and reuse
- **Caching**: Query result caching, object caching
- **Schema**: Appropriate normalization and data types

### Application Optimizations
- **Algorithm Improvements**: Reduce complexity, better data structures
- **Caching**: Multi-level caching strategy (memory, Redis, CDN)
- **Async Processing**: Non-blocking I/O, queue-based processing
- **Memory Management**: Object pooling, stream processing

### Network Optimizations
- **Compression**: Gzip, Brotli for responses
- **Batching**: Combine multiple operations
- **CDN**: Static content delivery
- **Data Serialization**: Efficient formats (Protocol Buffers, MessagePack)

### UX/Design Optimizations (Frontend Projects Only)
- **Component Performance**: Reduce re-renders, memoization, lazy loading
- **Animation Performance**: Target 60fps, optimize transitions, respect prefers-reduced-motion
- **Image Optimization**: WebP, responsive images, lazy loading, proper sizing
- **Design Token Efficiency**: Audit CSS file size, consolidate color/spacing definitions
- **Accessibility Performance**: WCAG AA compliance verification, screen reader performance
- **Component Reuse**: Increase design system adoption, reduce duplicate components
- **Dark Mode Optimization**: Ensure light/dark themes have equal performance

## Implementation Strategy

### Impact vs Effort Matrix
```
High Impact, Low Effort    → Immediate Priority (Cache queries, add indexes)
High Impact, High Effort   → Plan Next Sprint (Refactor architecture)
Low Impact, Low Effort     → Quick Wins (Code cleanup, minor tweaks)
Low Impact, High Effort    → Defer or Skip (Premature optimization)
```

### Risk Mitigation
- **Incremental Deployment**: Roll out optimizations gradually
- **Comprehensive Testing**: Unit, integration, performance testing
- **Monitoring**: Real-time performance monitoring
- **Rollback Plans**: Quick revert capability

## Need Help Choosing Optimization Strategy?

When facing competing optimization priorities, call **`@nara`**:
- Database optimization vs caching vs algorithm improvements (which first)?
- Component performance vs image optimization vs design system adoption?
- Animation performance vs accessibility performance trade-off?
- Should we optimize for time-to-first-byte or interaction speed?
- Rewrite vs incremental optimization (time vs risk)?

**Example:**
```
@nara Bottleneck analysis shows: 500ms API response, 80MB JS bundle, 60fps animation target.
Limited time: which optimization gives most user impact?
```

---

## Progress Format
```
Analysis: Identifying bottlenecks...
Optimizing: [component]...
Validating: Measuring improvements...
```

## Validation Protocol
**Pre-Optimization Baseline**:
- Record current performance metrics
- Document resource usage patterns
- Create performance test suite

**Post-Optimization Validation**:
- Measure improvements against baseline
- Validate no performance regressions
- Confirm functional correctness
- Update performance documentation

## Completion Format (Console Output Only)

```
✅ Optimization Complete (30 seconds)

📊 Performance Improvements:
- Response Time: 850ms → 240ms (↓72%)
- Throughput: 800 req/s → 2400 req/s (↑200%)
- Memory Usage: 580MB → 320MB (↓45%)
- CPU Utilization: 75% → 42% (↓44%)

🔧 Optimizations Applied:
✓ Database: Added 3 indexes, removed 2 n+1 queries
✓ Caching: Implemented Redis caching layer
✓ Algorithm: Optimized sorting from O(n²) → O(n log n)
✓ Async: Converted 4 blocking calls to async

📈 Impact:
- 15 slow endpoints now < 200ms
- Memory footprint reduced by 45%
- Peak CPU usage dropped 33%
- Overall system throughput 3x

💾 Next Steps:
- Deploy optimizations: `git push && deploy`
- Monitor metrics: Set alerts for regressions
- Save details: `--report` flag to create optimization.md
```

**No files created by default.** For detailed report:
```bash
/optimize --report   # Creates optimization.md with metrics + code changes
/optimize --metrics  # Creates metrics.json for monitoring systems
```

## Example Optimizations

### Database Index
```sql
-- Add composite index for frequent query pattern
CREATE INDEX idx_user_status_created
ON users(status, created_at)
WHERE status IN ('active', 'pending');
```

### Multi-Level Caching
```javascript
async function getUserData(userId) {
  // L1: In-memory cache
  if (memoryCache.has(userId)) return memoryCache.get(userId);

  // L2: Redis cache
  const redisData = await redis.get(`user:${userId}`);
  if (redisData) {
    memoryCache.set(userId, redisData);
    return redisData;
  }

  // L3: Database
  const dbData = await database.getUser(userId);
  await redis.setex(`user:${userId}`, 3600, dbData);
  memoryCache.set(userId, dbData);
  return dbData;
}
```

### Async Processing
```javascript
async function processLargeDataset(data) {
  const chunks = chunkArray(data, 1000);
  const promises = chunks.map(chunk =>
    queue.add('process-chunk', { chunk })
  );
  await Promise.all(promises);
  return await queue.getJobResults();
}
```

## Monitoring
**Real-time Metrics**:
- Response time percentiles (p50, p95, p99)
- Error rates and types
- Resource utilization (CPU, memory, disk, network)
- Business metrics (conversion, engagement)

**Alerting Thresholds**:
- Response time > 500ms for 95th percentile
- Error rate > 1% over 5 minutes
- CPU usage > 80% for 10 minutes
- Memory usage > 90% of available

---

**Follow shared protocols**: `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`