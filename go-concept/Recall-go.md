---

### Q4. What is `context` in Go?

- A `context.Context` carries deadlines, cancellation signals, and request-scoped values across API boundaries.
- It's part of the standard library: `context`.

**Main use cases:**
- **Cancellation** → Stop goroutines when they're no longer needed.
ctx, cancel := context.WithCancel(context.Background())
cancel() // Tell worker to stop
- **Timeouts/Deadlines** → Prevent infinite waiting.
ctx, cancel := context.WithTimeout(context.Background(),2*time.Second)
defer cancel()
- **Propagation** → Pass cancellation and values down call chains.
ctx := context.WithValue(context.Background(),RequestID,"abc123")

- Contexts are tree-like → if a parent is canceled, all its children are canceled too. This avoids goroutine leaks and ensures graceful shutdown.

---


### Q9. What are the common channel pitfalls in Go?

1. **Nil Channel**
   - A nil channel is one that hasn't been initialized with `make`.
   - Send → blocks forever. Receive → blocks forever. Close → panic (close of nil channel).
   - **Best practice:** Initialize with `make`, or intentionally use nil to disable a select case.
   - *Use Case* : Imagine two producers.Eventually one finishes.You don't want select to keep reading from its closed channel.
   - Because reading from a closed channel never blocks. So in the Select multiplexing, it might 
   Without Nil channel
   select
     ↓
   Closed channel
     ↓
   Receive immediately
     ↓
   ok == false
     ↓
   Loop
     ↓
   Receive immediately again
     ↓
   Infinite loop

```go
//This pattern lets a select continue processing the remaining active channels without spinning on channels that have already been closed
func main() {
	ch1 := producer("A", 3)
	ch2 := producer("B", 5)

	// Continue until both channels are disabled.
	for ch1 != nil || ch2 != nil {

		select {
		case msg, ok := <-ch1:
			if !ok {
				fmt.Println("Producer A finished")
				ch1 = nil // Disable this select case.
				continue
			}
			fmt.Println(msg)

		case msg, ok := <-ch2:
			if !ok {
				fmt.Println("Producer B finished")
				ch2 = nil // Disable this select case.
				continue
			}
			fmt.Println(msg)
		}
	}

	fmt.Println("All producers finished")
}
```

2. **Reading from a Closed Channel**
   - Allowed in Go. First you get any remaining buffered values; after the buffer is empty, receive returns zero value and `ok = false`.
   - **Best practice:** Use the "comma-ok" idiom (`v, ok := <-ch`) or `for range` to safely drain channels.

3. **Sending on a Closed Channel**
   - Causes panic: `panic: send on closed channel`.
   - **Best practice:** Only the sender should close the channel, and never send after close.

4. **Closing a Channel Twice**
   - Causes panic: `panic: close of closed channel`.
   - **Best practice:** Ensure a channel is closed only once, usually by the producer/coordinator.

5. **Unclear Ownership of Close**
   - A common bug is multiple goroutines trying to close the same channel.
   - **Best practice:** Senders, not receivers, close channels. If multiple senders exist, have a coordinator goroutine responsible for closing.

6. **Leaking Goroutines**
   - Happens when goroutines wait forever on a channel that never gets data or is never closed.
   - **Best practice:** Ensure channels are closed when no more values will be sent. Use timeouts or `context.Context` for cancellation.

---


### Q11. What are strategies to minimize heap allocations in Go?

- Prefer stack allocation; avoid unnecessary heap escapes (`go build -gcflags="-m"`).
- Use value types (`[]Struct`) instead of pointers (`[]*Struct`) where possible.
- Preallocate slices and maps with expected capacity (`make(..., cap)`).
- Reuse buffers and objects instead of repeatedly allocating new ones.:  Use sync.Pool
- Use string interning/deduplication for repeated package names, vendors, or identifiers.
- Store hashes as fixed-size byte arrays (`[16]byte`, `[32]byte`) rather than strings when feasible.
- Reduce total object count by using contiguous data structures (slices/arrays).
- Profile allocations regularly with `pprof`, `go tool trace`, and GC metrics.

**Pointer usage vs value types:**
- Pointers increase GC work because the collector must traverse every reachable pointer.
- `[]*Struct` creates many heap objects and pointer chains.
- `[]Struct` stores data contiguously, improving cache locality and reducing GC scanning.
- Pointer-free data structures are the cheapest for GC to process.
- Prefer values unless shared mutation or large-copy costs justify pointers.

**`sync.Pool`:**
- Reuses temporary objects/buffers to reduce allocation churn.
- Helps lower GC pressure in high-throughput APIs.
- Useful for JSON buffers, parsers, serializers, and temporary request objects.
- Not a persistent cache; pooled objects may be discarded during GC cycles.

**Slice allocations (cap vs len):**
- Growing slices via repeated `append()` can trigger multiple reallocations and copies.
- Preallocate capacity when the expected size is known.
- `make([]T, 0, n)` avoids repeated backing-array growth.
- Larger upfront capacity reduces allocation frequency and GC overhead.
- Capacity planning is often more important than initial length for performance-sensitive workloads.

**Impact on GC latency at scale:**
- GC latency is driven primarily by the number of live pointer-containing objects.
- Millions of small pointer-rich objects are more expensive than a few large contiguous arrays.
- Fewer allocations + fewer pointers + denser memory layouts = shorter GC pauses and better API performance.
- Prefer dense slices, value types, preallocation, and object reuse for large in-memory datasets.

---

### Q25. How would you design a high-throughput, low-latency event processing pipeline in Go handling 1M events/second?

This question evaluates understanding of batching, lock-free design, and mechanical sympathy.

**Architecture decisions at 1M events/sec:**
1. **Ingest layer** — use UDP or low-latency TCP with `SO_REUSEPORT` for kernel-level load balancing across multiple goroutines bound to different OS threads (`runtime.LockOSThread` + `GOMAXPROCS=NumCPU`).
2. **Zero-copy deserialization** — use protobuf or flatbuffers with pre-allocated byte slices from `sync.Pool`. Avoid JSON (reflection-heavy, allocates heavily).
3. **Lock-free ring buffer** — use a SPSC (single-producer single-consumer) or MPMC ring buffer with cache-line padding to prevent false sharing. Avoid channels for intra-process pipelines — buffered channels add scheduling overhead.
4. **Batching** — accumulate 1,000-10,000 events before flushing to storage. This amortises I/O cost and dramatically improves throughput.
5. **Partitioned processing** — shard events by key hash to dedicated goroutines — eliminates lock contention entirely for most operations.
6. **Backpressure** — when downstream is slow, apply backpressure at ingest (drop with metric / block producer / reject with 429).
7. **Storage** — batch-write to Kafka, Cassandra, or ClickHouse — all optimised for sequential writes.

---

### Q27. Walk through a production Go CPU profiling and optimization playbook.

**1. Locate (Isolate the Bottleneck)**
- Capture live data: Run a 30-second CPU profile during peak load using the native pprof endpoint (`/debug/pprof/profile`) or continuous tools like Pyroscope.
- Visualize the path: Use `go tool pprof` to generate a Flame Graph. Target any single function or execution path consuming >5% CPU.

**2. Replicate (Micro-Benchmark)**
- Isolate suspects: Move high-CPU functions into standard isolated unit tests.
- Measure baseline: Run micro-benchmarks via `go test -bench=. -benchmem` to precisely establish current CPU time per operation and memory allocation footprints.

**3. Optimize (High-Yield Fixes)**
- **Serialization** — swap `encoding/json` for high-performance alternatives like `sonic` or standard protobuf.
- **String allocation** — replace `fmt.Sprintf` with `strings.Builder` or plain concatenation in hot paths.
- **Regex** — move regex compilation out of request flows into `init()` blocks (`regexp.MustCompile`).
- **Memory management** — minimize garbage collection (GC) overhead by reusing byte slices (`[]byte`) and replacing reflection-heavy libraries with code generation (`go generate`).

**4. Validate (Verify and Load Test)**
- Iterate: re-run micro-benchmarks after every code change. If performance doesn't explicitly improve, revert.
- Stress test: simulate realistic production volume using traffic generators like k6 or Vegeta before deploying.

---