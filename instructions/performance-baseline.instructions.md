---
description: "Performance baseline measurement and validation framework for all project types"
applyTo: "**/*.js, **/*.ts, **/*.py, **/*.java, **/*.go, **/*.rs, **/*.cs, **/*.rb"
triggers: ["performance", "optimization", "benchmark", "profiling", "load testing"]
---

# Performance Baseline & Measurement Framework

## Overview
Establish performance baselines and validate optimization improvements with measurable metrics.

---

## Baseline Establishment Protocol

### Step 1: Instrument Code

**JavaScript/TypeScript**:
```javascript
async function measurePerformance(fn, name, iterations = 100) {
  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    await fn();
  }
  const end = performance.now();
  const avgTime = (end - start) / iterations;
  console.log(`${name}: ${avgTime.toFixed(2)}ms`);
}

const baseline = {
  apiResponseTime: 150,  // ms
  databaseQuery: 45,     // ms
  timestamp: new Date().toISOString()
};
```

**Python**:
```python
import time
import statistics

def measure_performance(fn, name, iterations=100):
    times = []
    for _ in range(iterations):
        start = time.perf_counter()
        fn()
        end = time.perf_counter()
        times.append((end - start) * 1000)
    
    avg_time = statistics.mean(times)
    print(f"{name}: {avg_time:.2f}ms")
```

**Go**:
```go
import "time"

func measurePerformance(fn func(), name string, iterations int) float64 {
    var totalTime time.Duration
    for i := 0; i < iterations; i++ {
        start := time.Now()
        fn()
        totalTime += time.Since(start)
    }
    
    avgMs := float64(totalTime.Milliseconds()) / float64(iterations)
    fmt.Printf("%s: %.2fms\n", name, avgMs)
    return avgMs
}
```

### Step 2: Document Baseline

Create `performance-baseline.json`:

```json
{
  "project": "my-app",
  "timestamp": "2026-09-06T12:00:00Z",
  "metrics": {
    "api_response_time": {
      "value": 150,
      "unit": "ms",
      "description": "Average response time for GET /api/users",
      "percentiles": {
        "p50": 145,
        "p95": 180,
        "p99": 210
      }
    },
    "memory_usage": {
      "value": 128,
      "unit": "MB",
      "description": "Memory footprint at startup"
    },
    "throughput": {
      "value": 800,
      "unit": "requests/sec",
      "description": "Max sustained throughput"
    }
  }
}
```

---

## Performance Metrics by Project Type

### Backend/API Projects

| Metric | Target | Tool |
|--------|--------|------|
| **Response Time** | <200ms p95 | Artillery, K6 |
| **Throughput** | >1000 req/s | Vegeta, AB |
| **Database Query** | <100ms | Query profiling |
| **Memory Leak** | Stable after warmup | Node Inspector |
| **CPU Usage** | <70% under load | Performance Monitor |

### Frontend/Web Projects

| Metric | Target | Tool |
|--------|--------|------|
| **Time to Interactive** | <3s | Lighthouse |
| **First Contentful Paint** | <1.5s | PageSpeed Insights |
| **JavaScript Bundle Size** | <200KB gzipped | Webpack Analyzer |
| **Largest Contentful Paint** | <2.5s | Web Vitals |

### Mobile Projects

| Metric | Target | Tool |
|--------|--------|------|
| **Time to First Paint** | <1s | Android Profiler |
| **Frame Rate** | 60fps | Performance Monitor |
| **Memory Usage** | <150MB baseline | Memory Profiler |
| **App Startup Time** | <2s cold start | Built-in tools |

---

## Profiling Tools

### JavaScript/TypeScript
```bash
# Node.js built-in profiler
node --prof app.js

# Clinic.js (comprehensive)
clinic doctor -- node app.js

# K6 (load testing)
k6 run test-script.js

# Artillery (API load testing)
artillery run load-test.yml
```

### Python
```bash
# cProfile (built-in)
python -m cProfile -o stats.prof app.py

# Memory profiler
python -m memory_profiler app.py

# Locust (load testing)
locust -f locustfile.py --host=http://localhost:8000
```

### Go
```bash
# Built-in profiling
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof cpu.prof

# Vegeta (HTTP load testing)
echo "GET http://localhost:8080/" | vegeta attack -duration=30s | vegeta report
```

### Java
```bash
# JMeter (load testing)
jmeter -n -t test.jmx -l results.jtl

# Gatling (performance testing)
mvn gatling:test
```

### Rust
```bash
# Criterion (benchmarking)
cargo bench

# perf (Linux)
perf record ./target/release/app
perf report
```

---

## Optimization Workflow

### Before Optimization
1. Establish baseline metrics (performance-baseline.json)
2. Identify bottlenecks (profiling)
3. Set improvement targets
4. Create regression tests

### During Optimization
1. Apply one optimization at a time
2. Measure impact immediately
3. If regression detected → revert
4. Document each change

### After Optimization
1. Compare against baseline
2. Generate improvement report
3. Validate no regressions
4. Update baseline (new JSON)

---

## Performance Report Template

```
PERFORMANCE OPTIMIZATION REPORT
═══════════════════════════════════

Metric                  BEFORE    AFTER     CHANGE
─────────────────────────────────────────────────
API Response Time       150ms     95ms      -36.7%
Throughput             800 req/s  2400 req/s +200%
Memory Usage           256MB     185MB     -27.7%

OPTIMIZATIONS APPLIED
═════════════════════

1. Database Query Optimization (-60%)
   - Added 3 missing indexes
   - Eliminated 2 N+1 query patterns

2. Caching Implementation (-36.7%)
   - Redis for user sessions
   - Client-side caching headers

3. Bundle Size Optimization (-42%)
   - Code splitting
   - Tree-shaking dead code

REGRESSION TESTING
══════════════════

✅ All unit tests pass (1,247 tests)
✅ Integration tests pass (342 tests)
✅ No performance regressions detected
```

---

## Anti-Patterns (NEVER Do This)

| Pattern | Issue | Fix |
|---------|-------|-----|
| **Optimize without baseline** | No way to measure improvement | Establish baseline first |
| **Premature optimization** | Waste time on non-bottlenecks | Profile first, optimize second |
| **Change multiple things at once** | Can't isolate improvements | One optimization at a time |
| **No regression tests** | Improvements break other things | Regression test suite mandatory |

---

## Production Monitoring

### Key Metrics
```
- Response time: p50, p95, p99
- Error rates
- Resource utilization (CPU, memory, disk)
- Business metrics (conversion, engagement)
```

### Alerting Thresholds
```
- Response time >300ms (p95) → Page team
- Error rate >1% over 5 min → Alert
- CPU usage >80% for 10 min → Auto-scale
- Memory >90% available → Restart service
```

---

*Performance Baseline & Measurement Framework v1.0.0*  
*Framework: Spec Kit Constitutional Development Framework*
