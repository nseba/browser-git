# Performance Benchmarks

This directory contains performance benchmarks for browser-git components.

## Running Benchmarks

```bash
# Run all benchmarks
npm run bench

# Run specific benchmark file
npm run bench -- benchmarks/storage-adapters.bench.ts

# Run benchmarks with comparison
npm run bench -- --compare
```

## Benchmark Files

### storage-adapters.bench.ts

Benchmarks for storage adapter implementations:

- **Write Performance**: Measures write speed for small (100B), medium (10KB), and large (100KB) data
- **Read Performance**: Measures read speed across different data sizes
- **Bulk Operations**: Tests performance with 100+ keys
- **Delete Performance**: Measures deletion speed

Adapters tested:

- MemoryAdapter
- IndexedDBAdapter
- LocalStorageAdapter

### filesystem.bench.ts

Benchmarks for filesystem operations:

- **File Operations**: Write/read performance across different file sizes (1KB - 100KB)
- **Directory Operations**: Create/read directories with various nesting levels
- **Stat Operations**: File and directory metadata retrieval
- **Delete Operations**: File and directory tree deletion
- **Bulk Operations**: Performance with 100+ files

## Interpreting Results

Benchmark results show:

- **ops/sec**: Operations per second (higher is better)
- **avg time**: Average execution time (lower is better)
- **min/max**: Fastest and slowest times
- **samples**: Number of iterations run

## Performance Targets

Expected performance targets:

### Storage Adapters

- MemoryAdapter: >10,000 ops/sec for small data
- IndexedDBAdapter: >1,000 ops/sec for small data
- LocalStorageAdapter: >500 ops/sec for small data

### Filesystem

- Small file operations (<1KB): >5,000 ops/sec
- Medium file operations (10KB): >1,000 ops/sec
- Large file operations (100KB): >100 ops/sec
- Directory operations: >1,000 ops/sec

## CI Integration

Benchmarks can be run in CI to track performance over time:

```yaml
- name: Run benchmarks
  run: npm run bench -- --reporter json > benchmarks.json

- name: Store benchmark results
  uses: benchmark-action/github-action-benchmark@v1
  with:
    tool: "vitest"
    output-file-path: benchmarks.json
```

## Adding New Benchmarks

1. Create a new `.bench.ts` file in the benchmarks directory
2. Import `bench` and `describe` from vitest
3. Write benchmark suites:

```typescript
import { bench, describe } from "vitest";

describe("My Component", () => {
  bench("operation name", async () => {
    // Code to benchmark
  });
});
```

4. Run with `npm run bench`

## Notes

- Benchmarks use jsdom environment for browser API compatibility
- Each benchmark runs multiple iterations to get accurate averages
- Results may vary based on hardware and browser
- Use `--compare` to compare against baseline results
