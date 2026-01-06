# Performance Profiling Example

This example demonstrates how to profile the go-linear SDK to identify performance bottlenecks and optimization opportunities.

## What Gets Profiled

This profiles the **SDK** (`pkg/linear`), not the CLI:
- ✅ SDK: API client, transport layer, pagination, iterators
- ❌ CLI: Command-line interface (not performance-critical)

## Quick Start

### File-Based Profiling (Recommended)

Generate profile files for offline analysis:

```bash
# Set your API key
export LINEAR_API_KEY=lin_api_xxx

# Run profiling example (generates cpu.prof, mem.prof, trace.out)
make profile-example

# Or run directly
cd examples/profiling
go run main.go

# Analyze CPU profile
go tool pprof cpu.prof
# Interactive commands:
# - top10: Show top 10 functions by CPU time
# - list funcName: Show annotated source code
# - web: Generate call graph (requires graphviz)

# Analyze memory allocations
go tool pprof -alloc_space mem.prof
# Interactive commands:
# - top10: Show top 10 allocation sites
# - list funcName: Show allocation locations

# View execution trace
go tool trace trace.out
# Opens web UI showing goroutine scheduling, blocking, GC
```

### Live Profiling (Server Mode)

Profile a running application with HTTP endpoints:

```bash
# Start profiling server
make profile-server

# In another terminal, capture profiles:
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30  # CPU
go tool pprof http://localhost:6060/debug/pprof/heap               # Memory
go tool pprof http://localhost:6060/debug/pprof/goroutine          # Goroutines
go tool pprof http://localhost:6060/debug/pprof/block              # Blocking

# Or use web UI:
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

## Profiling SDK Benchmarks

Profile the SDK's benchmark suite:

```bash
# CPU profiling
make profile-cpu
go tool pprof cpu.prof

# Memory profiling
make profile-mem
go tool pprof -alloc_space mem.prof

# Both
make profile-all
```

## Comparing Before/After Changes

Use benchstat to compare performance changes:

```bash
# Save baseline before changes
make benchmark-baseline

# Make your optimization changes...

# Compare with baseline
make benchmark-compare

# Example output:
# name              old time/op  new time/op  delta
# IssueIterator-10  2.45ms ± 2%  1.98ms ± 3%  -19.18% (p=0.000 n=10+10)
```

## What to Look For

### CPU Profile

**Expected hot spots:**
- `transport.RoundTrip` - Network I/O (should dominate)
- `json.Unmarshal` - Response deserialization
- `http.Transport.roundTrip` - HTTP client

**Unexpected hot spots (optimization opportunities):**
- Excessive time in `io.ReadAll` (request body buffering)
- String concatenation or formatting in hot paths
- Reflection or type assertions

**Commands:**
```bash
go tool pprof cpu.prof
(pprof) top10              # Show top 10 functions
(pprof) top10 -cum         # Show by cumulative time
(pprof) list RoundTrip     # Show annotated source
(pprof) web                # Generate call graph (requires graphviz)
(pprof) -http=:8080        # Start web UI
```

### Memory Profile

**Expected allocations:**
- Response body buffers
- JSON deserialization structures
- Pagination buffers

**Unexpected allocations (optimization opportunities):**
- Per-request allocations in transport layer
- String building without `strings.Builder`
- Slice growth without preallocation
- Interface boxing in hot paths

**Commands:**
```bash
# Total allocations (includes freed memory)
go tool pprof -alloc_space mem.prof
(pprof) top10
(pprof) list transport.RoundTrip

# Currently in-use memory
go tool pprof -inuse_space mem.prof
(pprof) top10
```

### Execution Trace

**What to check:**
- Goroutine scheduling patterns
- Channel/mutex blocking
- GC frequency and duration
- Syscall blocking

**Commands:**
```bash
go tool trace trace.out
# Opens web UI with:
# - View trace: Timeline of events
# - Goroutine analysis: Per-goroutine statistics
# - Network blocking profile
# - Synchronization blocking profile
# - Syscall blocking profile
```

## Optimization Workflow

Follow this process from the practices document:

1. **Profile first** - Don't guess where the bottleneck is
   ```bash
   make profile-example
   go tool pprof cpu.prof
   (pprof) top10
   ```

2. **Identify hot spots** - Focus on functions consuming >5% of time

3. **Analyze root cause** - Use `list funcName` to see annotated source

4. **Make targeted changes** - Optimize only identified bottlenecks

5. **Benchmark changes** - Verify improvement
   ```bash
   make benchmark-baseline    # Before optimization
   # Make changes
   make benchmark-compare     # After optimization
   ```

6. **Profile again** - Confirm reduction in hot spot

## Common Optimization Patterns

### Pattern 1: Preallocate Slices

**Before:**
```go
var results []Result
for _, item := range items {
    results = append(results, process(item))  // May reallocate
}
```

**After:**
```go
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))  // No reallocation
}
```

**Profile check:** Look for `runtime.growslice` in allocations

### Pattern 2: Use strings.Builder

**Before:**
```go
s := ""
for _, v := range values {
    s += v  // Allocates new string each iteration
}
```

**After:**
```go
var b strings.Builder
for _, v := range values {
    b.WriteString(v)
}
s := b.String()
```

**Profile check:** Look for string concatenation allocations

### Pattern 3: Buffer Reuse with sync.Pool

**Before:**
```go
func processRequest() {
    buf := new(bytes.Buffer)  // Allocates every call
    // use buf
}
```

**After:**
```go
var bufferPool = sync.Pool{
    New: func() interface{} { return new(bytes.Buffer) },
}

func processRequest() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    // use buf
}
```

**Profile check:** Look for repeated allocations of same type

## Known Hot Paths

Based on profiling, these are the performance-critical paths:

1. **transport.RoundTrip** (transport.go)
   - Request body buffering for retry support
   - Exponential backoff calculation
   - Rate limit header parsing

2. **JSON Deserialization** (internal/graphql)
   - Response unmarshaling (unavoidable cost)
   - Consider using json.Decoder for streaming

3. **Iterator Buffering** (pagination.go)
   - Buffer allocation strategy
   - Page size tuning

## Further Reading

- [Go Performance Profiling Guide](https://go.dev/blog/pprof)
- [Profiling Go Programs](https://go.dev/doc/diagnostics)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)
- Project practices: `~/code/hexproofdev/practices/go/performance-profiling.md`

## Notes

- **Profile production workloads** - Synthetic benchmarks may not reflect real usage
- **Profile with real API calls** - Network latency dominates, but find what you can optimize
- **Don't over-optimize** - Focus on measurable improvements (>10% speedup)
- **Consider costs** - Sometimes "good enough" is better than "perfect but complex"
