# Benchmark Results for Glob Package

## Bench Command

```bash
go test -bench=. -benchtime 100000000x -benchmem
```

## Test Data

* Benchmark tests were conducted on various patterns and matching scenarios for glob matching.

---

## System Information

* **OS:** linux  
* **Architecture:** amd64  
* **CPU:** Intel(R) Core(TM) i5-6500 CPU @ 3.20GHz  
* **Package:** github.com/FMotalleb/go-tools/matcher/glob  

---

## Benchmark Output

### Compile Benchmarks

| Pattern                                  | Time/op | Bytes/op | Allocations |
|------------------------------------------|---------|----------|-------------|
| `example.com`                            | 196.4 ns | 120 B/op | 4 allocs/op |
| `*.example.com`                          | 272.2 ns | 216 B/op | 5 allocs/op |
| `api-??.example.com`                     | 391.5 ns | 416 B/op | 7 allocs/op |
| `example.{com,net,org}`                  | 490.8 ns | 336 B/op | 10 allocs/op |
| `*.api-??.{example.com,test.net,dev.io}` | 872.8 ns | 936 B/op | 14 allocs/op |

---

### Match Benchmarks

| Test Case                                | Time/op   | Throughput (MB/s) | Bytes/op | Allocations |
|------------------------------------------|-----------|--------------------|----------|-------------|
| **Literal**                              |           |                    |          |             |
| `api.example.com`                        | 16.29 ns  | 920.81 MB/s        | 0 B      | 0 allocs/op |
| `www.example.com`                        | 5.706 ns  | 2628.72 MB/s       | 0 B      | 0 allocs/op |
| **Star Prefix**                          |           |                    |          |             |
| `api.example.com`                        | 35.04 ns  | 428.14 MB/s        | 0 B      | 0 allocs/op |
| `www.api.example.com`                    | 61.26 ns  | 310.14 MB/s        | 0 B      | 0 allocs/op |
| `example.com`                            | 46.92 ns  | 234.45 MB/s        | 0 B      | 0 allocs/op |
| `api.example.net`                        | 74.70 ns  | 200.79 MB/s        | 0 B      | 0 allocs/op |
| **Question Mark**                        |           |                    |          |             |
| `api-01.example.com`                     | 26.05 ns  | 690.96 MB/s        | 0 B      | 0 allocs/op |
| `api-99.example.com`                     | 26.46 ns  | 680.16 MB/s        | 0 B      | 0 allocs/op |
| `api-1.example.com`                      | 17.61 ns  | 965.31 MB/s        | 0 B      | 0 allocs/op |
| `api-100.example.com`                    | 18.93 ns  | 1003.76 MB/s       | 0 B      | 0 allocs/op |
| **Brace Expansion**                      |           |                    |          |             |
| `api.example.com`                        | 23.33 ns  | 643.00 MB/s        | 0 B      | 0 allocs/op |
| `api.example.net`                        | 25.26 ns  | 593.75 MB/s        | 0 B      | 0 allocs/op |
| `api.example.org`                        | 28.40 ns  | 528.21 MB/s        | 0 B      | 0 allocs/op |
| `api.example.io`                         | 18.47 ns  | 758.03 MB/s        | 0 B      | 0 allocs/op |
| **Complex Patterns**                     |           |                    |          |             |
| `subdomain.api-01.example.com`           | 88.44 ns  | 316.62 MB/s        | 0 B      | 0 allocs/op |
| `subdomain.api-99.test.net`              | 90.76 ns  | 275.44 MB/s        | 0 B      | 0 allocs/op |
| `subdomain.api-1.example.com`            | 164.7 ns  | 163.96 MB/s        | 0 B      | 0 allocs/op |
| `api-01.example.com`                     | 98.63 ns  | 182.49 MB/s        | 0 B      | 0 allocs/op |

---

### Special Benchmarks

| Test Case                                | Time/op   | Throughput (MB/s) | Bytes/op | Allocations |
|------------------------------------------|-----------|--------------------|----------|-------------|
| **Zero Alloc Match**                     | 85.85 ns  | N/A                | 0 B      | 0 allocs/op |
| **Input Length**                         |           |                    |          |             |
| Length 13                                | 23.82 ns  | 545.74 MB/s        | 0 B      | 0 allocs/op |
| Length 15                                | 34.99 ns  | 428.68 MB/s        | 0 B      | 0 allocs/op |
| Length 21                                | 67.71 ns  | 310.16 MB/s        | 0 B      | 0 allocs/op |
| Length 31                                | 120.7 ns  | 256.94 MB/s        | 0 B      | 0 allocs/op |
| Length 41                                | 185.9 ns  | 220.54 MB/s        | 0 B      | 0 allocs/op |
| **Worst Case**                           |           |                    |          |             |
| Multiple Stars                           | 79.35 ns  | 189.04 MB/s        | 0 B      | 0 allocs/op |
| Deep Brace Expansion                     | 52.74 ns  | 455.05 MB/s        | 0 B      | 0 allocs/op |
| Long Non-Match                           | 183.7 ns  | 217.74 MB/s        | 0 B      | 0 allocs/op |
| **Parallel Match**                       | 26.64 ns  | N/A                | 0 B      | 0 allocs/op |

---

## Summary

* **Compile Benchmarks**: Show incremental cost for different glob patterns. Simple patterns like `example.com` are compiled very quickly, while more complex patterns like brace expansions and multiple stars incur higher costs.  
* **Match Benchmarks**: All cases show zero allocations, demonstrating excellent efficiency.  
  * Matching literals is the fastest, with throughput exceeding 2 GB/s in some cases.  
  * Patterns with wildcards (`*` and `?`) or brace expansions are slower but remain highly performant.  
* **Worst-Case Scenarios**: Even in edge cases (e.g., long non-matching patterns or deep brace expansions), the performance remains within acceptable limits.  
* **Parallel Matching**: Minimal overhead for parallel execution.

This package demonstrates high performance and zero-alloc efficiency, making it well-suited for real-time glob matching tasks.  
