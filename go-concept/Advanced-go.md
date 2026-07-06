# Advanced Go — Interview Preparation
### Organized Beginner → Intermediate → Advanced → Expert

---

## Table of Contents

- [Level 1 — Beginner (Concurrency Foundations)](#level-1--beginner-concurrency-foundations)
- [Level 2 — Intermediate (Channels, Memory, Patterns)](#level-2--intermediate-channels-memory-patterns)
- [Level 3 — Advanced (Memory Model, Lock-Free, Performance)](#level-3--advanced-memory-model-lock-free-performance)
- [Level 4 — Expert (Scheduler, GC Internals, System Design)](#level-4--expert-scheduler-gc-internals-system-design)
- [Level 5 — Architect (Distributed Systems & Production Playbooks)](#level-5--architect-distributed-systems--production-playbooks)

---

## Level 1 — Beginner (Concurrency Foundations)

---

### Q1. What is a goroutine?

- A goroutine is a lightweight, user-space thread managed by the Go runtime.
- Much cheaper than OS threads — creating thousands of goroutines is common.
- Each goroutine has its own stack and is scheduled cooperatively by the Go scheduler (M:N model).

**Stack management:**
- Starts with a very small stack (~2 KB).
- The stack is dynamically grown and shrunk by the runtime (unlike OS threads with fixed ~1–2 MB stacks).
- This makes goroutines memory-efficient and scalable to millions.

**Communication:**
- Goroutines communicate via channels (synchronized queues).
- Synchronization is also possible via mutexes, atomic ops, etc.
- Under the hood, channels use locks and wait queues (`sudog` structs) to park/wake goroutines.

**Garbage collection interaction:**
- The Go GC is goroutine-aware.
- GC can scan goroutine stacks (which are small and segmented) efficiently.
- Stacks shrink when not needed → helps keep memory low.

---

### Q2. What are goroutine leaks, and how do you prevent them?

- A goroutine leak happens when a goroutine is blocked forever (e.g., waiting on a channel that no one writes to).

**Prevention:**
- Use `context.Context` with cancellation.
- Ensure channels are properly closed.
- Monitor goroutine counts with `runtime.NumGoroutine()` in tests.

---

### Q3. What is the Fan-out / Fan-in pattern, and what are worker pools?

**Worker pools:**
- A fixed number of workers (goroutines) pull tasks from a shared channel (job queue).
- Workers process tasks concurrently; results are collected in another channel.
- **Why use it?** To limit concurrency (avoid spawning too many goroutines at once); efficient for CPU/memory-bound workloads.

**Fan-out:**
- One input channel, multiple goroutines consume from it.
- Each goroutine processes messages independently.
- **Use case:** Parallel processing of a stream of tasks.

**Fan-in:**
- Multiple input channels are combined (merged) into one output channel.
- Helps collect results from many producers into a single consumer.
- **Use case:** Multiple goroutines generate results → merge into one stream.

**Example implementation:**

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Worker function
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		fmt.Printf("Worker %d started job %d\n", id, job)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
		fmt.Printf("Worker %d finished job %d\n", id, job)
		results <- job * 2 // Example: processing = multiply by 2
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numJobs := 10
	numWorkers := 3

	jobs := make(chan int, numJobs)    // job queue
	results := make(chan int, numJobs) // results channel

	var wg sync.WaitGroup

	// Fan-out: start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs) // no more jobs

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Fan-in: collect results
	for result := range results {
		fmt.Println("Result:", result)
	}

	fmt.Println("All jobs processed ✅")
}
```

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

### Q5. What are deadlocks and race conditions?

**Deadlock:**
- Happens when goroutines are stuck waiting forever, and no further progress is possible.
- **Why?** Main goroutine tries to send to a channel; no other goroutine is receiving; program hangs. Go runtime detects this and panics.

**How to avoid:**
- Always ensure every send has a receiver.
- Use buffered channels when appropriate.
- Close channels properly to signal completion.

**Race condition:**
- Occurs when two or more goroutines access shared memory at the same time, and at least one of them writes.
- **Problem:** Both goroutines write to a counter at the same time → final value is unpredictable.

**Fixes:**
- Use `sync.Mutex` or `sync/atomic` to protect shared state.

---

## Level 2 — Intermediate (Channels, Memory, Patterns)

---

### Q6. Buffered vs unbuffered channels — when do you use each?

**Use unbuffered when:**
- You need synchronization/ordering guarantees.
- Communication is part of coordination logic (e.g., signals, done channels).

**Use buffered when:**
- You want to decouple producer and consumer speeds.
- You are designing pipelines or worker pools.
- You need throughput over strict sync.

---

### Q7. What are the important rules about closing channels?

- Only the sender should close the channel.
- If multiple goroutines send, usually the coordinator (the one who knows no more values will be sent) closes it.
- Receivers should never close a channel.
- Closing twice panics:
  ```go
  close(ch)
  close(ch) // panic: close of closed channel
  ```
- Sending on a closed channel panics:
  ```go
  close(ch)
  ch <- 10 // panic: send on closed channel
  ```
- Receiving on a closed channel never blocks. It returns:
  - Remaining buffered values first.
  - Then zero-value repeatedly with `ok=false`.

---

### Q8. What is multiplexing with `select`?

- Multiplexing means listening to multiple channels at the same time and acting on whichever is ready first.
- In Go, this is done with the `select` statement.
- `select` waits until one of the case channels is ready.
- If multiple are ready → one is chosen randomly (fairness).
- If none are ready:
  - If there is a `default` → it executes immediately (non-blocking).
  - Otherwise, `select` blocks until something is ready.

**Why multiplex?**
- Goroutines often interact with multiple channels:
  - Waiting on input from multiple sources.
  - Handling multiple outputs.
  - Coordinating cancellation/timeouts.

for {

	select {

	case <-ctx.Done():
		fmt.Println("Stopping worker")
		return
	case <-ticker.C:
		fmt.Println("Heartbeat")

	case job := <-jobs:
		fmt.Println("Processing", job)
	}
}


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

### Q10. What is escape analysis in Go?

- Escape analysis is a compiler technique in Go that determines whether a variable can be safely allocated on the stack or whether it must "escape" to the heap.
- **Stack allocation** → Fast, automatically cleaned up when the function returns.
- **Heap allocation** → Slower, managed by Go's garbage collector (GC).
- Go tries to allocate variables on the stack whenever possible for performance. If the compiler detects that a variable outlives the scope of the function where it's created, it escapes to the heap.

**Example — returned from a function:**

```go
func foo() *int {
    x := 10
    return &x // x escapes to heap, cannot live only on the stack
}
```
- `x` must survive after `foo` returns, so it's on the heap.
- **To check:** `go build -gcflags="-m" file.go`

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

## Level 3 — Advanced (Memory Model, Lock-Free, Performance)

---

### Q12. Explain the "Happens-Before" concept in the Go Memory Model.

- In modern hardware, code does not run sequentially line-by-line the way you wrote it. To maximize CPU pipeline efficiency, compilers and CPUs aggressively reorder assembly instructions, and cores aggressively cache data in local registers rather than flushing them to main RAM.
- The Go Memory Model is a strict contract: "Unless you use a known synchronization primitive, the runtime makes absolutely zero guarantees that Goroutine B will see a change made by Goroutine A."
- The core mechanism used to enforce order is called a **Happens-Before** boundary (a memory barrier/fence). If event X happens before event Y, the Go runtime forces the CPU to flush its local cache pipelines so that the memory changes made by X are structurally visible to the code executing Y.

**Broken code example:**

```go
var ready bool
var data int

go func() {
    data = 42
    ready = true  // PROBLEM: The compiler or CPU can reorder these lines!
}()

for !ready {}     // PROBLEM: Core starvation / infinite cache loop
fmt.Println(data) // PROBLEM: May print 0
```

**Why it fails architecturally:**
- **Instruction reordering:** Because `data` and `ready` are independent variables, the compiler's optimization pass might decide it's faster to execute `ready = true` before `data = 42` in assembly. If the background goroutine gets preempted right between them, the main loop exits, and `fmt.Println(data)` prints 0.
- **Register caching (the infinite loop):** The main thread executing `for !ready {}` runs incredibly fast. The CPU core may load `ready` directly into a hardware register and never read from main RAM again. Even if the background goroutine sets `ready = true` in memory, the main loop's CPU core keeps checking its stale internal register copy, resulting in a permanent infinite loop (CPU thread leak).

---

### Q13. What is the difference between a Data Race and a Race Condition?

- **Data Race (raw memory conflict):** Two concurrent execution threads access the exact same memory location simultaneously without synchronization, and at least one access is a write. This causes raw undefined behavior/memory corruption.
- **Race Condition (flawed business logic):** The code is perfectly synchronized (no data corruption occurs, often protected by mutexes), but the order of execution alters the final outcome incorrectly.
```go
func Purchase(a *Account, amount int) bool {
	a.mu.Lock()
	ok := a.balance >= amount
	a.mu.Unlock()
	if !ok {
		return false
	}
	// Some expensive work
	time.Sleep(time.Second)
	a.mu.Lock()
	a.balance -= amount
	a.mu.Unlock()
	return true
}
```

---

### Q14. What does "lock-free" mean, and how is it designed?

**What does "lock-free" mean?**
- A lock-free data structure avoids mutexes and instead uses atomic operations such as:
  - `atomic.Load`
  - `atomic.Store`
  - `atomic.CompareAndSwap`
  - `atomic.Add`
- The goal is to allow multiple goroutines to make progress without blocking on a lock.

**Design approach:**
- The most common pattern is:
  1. Atomically read the current state.
  2. Compute a new state.
  3. Attempt to update using Compare-And-Swap (CAS).
  4. If another goroutine modified the state meanwhile, retry.

---

### Q15. What is the ABA problem and does it affect Go's `sync/atomic` CAS?

- The ABA problem occurs when a value changes from A to B and back to A. A CAS operation only checks whether the current value equals the expected value and cannot detect intermediate modifications.
- Go's `sync/atomic` CAS operations are susceptible to ABA.
- Garbage collection reduces memory-reuse-related ABA bugs, but logical ABA problems still exist in lock-free algorithms.
- Common mitigations include versioned pointers, tagged counters, and hazard-pointer-style techniques.

---

### Q16. Compare `sync.Mutex`, `sync.RWMutex`, `sync/atomic`, and channels. When would you choose each at architect level?

- **`sync.Mutex`** — exclusive lock. All readers and writers block each other. Use when writes are as frequent as reads, or when the critical section is complex (multiple field updates that must be atomic together).
- **`sync.RWMutex`** — multiple concurrent readers, exclusive writers. Use when reads vastly outnumber writes (e.g., config lookups, caches). Read lock contention is much lower than Mutex. However, RWMutex is slower than Mutex when writes are frequent because it must drain all readers.
- **`sync/atomic`** — lock-free operations on single 32/64-bit values or pointers. Cheapest synchronisation primitive — no OS syscall, no goroutine park. Use for counters, flags, and single-value CAS operations. Cannot atomically update two related values together.
- **`Channels`** — synchronisation through communication. Use when you want to transfer ownership of data between goroutines, pipeline work, broadcast signals (close), or implement fan-out/fan-in. Higher overhead than mutex for simple variable protection but enables compositional designs.

**Architectural guidance:**
- Prefer channels for orchestration and data pipelines.
- Prefer mutexes for protecting shared state that is hard to restructure around ownership.
- Prefer atomic for hot-path counters and flags.
- Avoid RWMutex unless read-to-write ratio is measured to be >10:1.

---

### Q17. How would you implement a concurrent LRU cache in Go? What are the lock granularity options?

- Use a hashmap for lookup and a doubly linked list for eviction ordering.
- The simplest design uses a single mutex protecting both structures.
- For higher concurrency, **sharding** is usually the best tradeoff — each shard maintains its own map and LRU list, reducing contention significantly.
- `RWMutex` often provides limited benefit because even `Get` operations update recency and therefore require writes.
- Fully lock-free LRUs are possible but considerably more complex due to ABA and memory-ordering concerns.

---

### Q18. When does a channel become a bottleneck and how do you diagnose it?

- A Go channel is not just a raw memory pipe; it is a complex structure (`hchan`) protected by an internal mutex lock. Every time a goroutine sends (`ch <-`) or receives (`<-ch`), it must acquire this lock.
- A bottleneck occurs when the time spent waiting to acquire the channel's internal lock exceeds the time spent doing actual work.

**Unbuffered channels — the hard handshake:**
- Unbuffered channels have a capacity of 0. A send cannot complete until a receive is ready to take the data simultaneously.
- **The bottleneck:** If your worker pools are mismatched (e.g., 50 producers sending data, but only 2 consumers processing it), the producers will immediately block. They enter a waiting state, forcing the Go scheduler to constantly context-switch them out.

**Buffered channels — the illusion of speed:**
- Buffered channels hold an in-memory ring buffer. Sending only blocks when the buffer is completely full; receiving only blocks when the buffer is completely empty.
- **The bottleneck:** If the downstream consumer is slow, the buffer quickly fills up. Once full, the buffered channel degrades into an unbuffered channel — every single producer now blocks until a consumer explicitly frees up a slot. Conversely, if producers are too slow, an army of consumers will fight over the empty buffer's single internal lock, causing massive lock contention.

**Architectural fixes:**
- Changing the buffer size (e.g., `make(chan int, 10)` → `make(chan int, 1000)`) is usually just a temporary band-aid. True architectural fixes include:
  - **Fan-Out / Worker Pools** — increase the number of concurrent consumer goroutines to drain the channel faster.
  - **Batching** — instead of sending millions of individual integers over a channel, pass slices of data (e.g., `chan []int`). This amortizes the cost of the channel's internal lock over hundreds of data points.
  - **Lock-Free Alternatives** — if you are just passing simple state changes or counters, bypass channels completely and use `sync/atomic` primitives, which map directly to atomic CPU hardware instructions.

---

### Q19. What is stack growth in Go? How does it affect performance compared to fixed-size thread stacks?

- Unlike traditional languages where threads are allocated a large, fixed-size block of memory for their stack (typically 1 MB–8 MB), Go uses dynamically resizing stacks for goroutines.

**The starting point:**
- Every goroutine starts with a minimal stack allocation of 2 KB (or 4 KB depending on the architecture).

**The stack check:**
- At the prologue of every function call, the compiler inserts a tiny check to see if the current function's frame fits within the remaining space of the goroutine's stack.

**The growth (copying stack):**
- If the space is insufficient, the Go runtime triggers stack growth:
  1. It allocates a new, contiguous memory block that is twice the size of the current stack.
  2. It copies all stack frames, local variables, and arguments from the old stack to the new stack.
  3. It adjusts any active internal pointers pointing to data on the old stack to point to their new memory addresses.
  4. It releases the old stack back to the runtime allocator.

---

### Q20. What is escape analysis used for at an architectural level, and how do you eliminate heap allocations in hot paths?

- Escape analysis is performed by the Go compiler at compile time to determine whether a variable can live on the goroutine's stack (cheap, automatically reclaimed) or must be allocated on the heap (requires GC).

**A variable escapes to the heap when:**
- Its address is stored somewhere that outlives the current stack frame (returned pointer, interface, closure capturing by reference, sent over channel).
- It is too large for the stack (default 8MB limit).
- The compiler cannot determine its size at compile time (dynamic sized allocations).
- It is assigned to an interface (requires addressability).

**Diagnosing:**
- `go build -gcflags="-m -m"` — prints every escape decision and reason.
- `go test -bench=. -gcflags="-m"` — escape analysis during benchmarks.

**Key techniques to prevent escapes:**
1. Prefer value receivers over pointer receivers when the struct is small.
2. Pass buffers in (accept `[]byte`) rather than returning them.
3. Avoid storing pointers in interfaces in hot paths.
4. Pre-allocate slices with exact capacity — slice with cap known at compile time stays on stack.
5. Avoid closures capturing variables by reference in tight loops — use explicit parameters instead.
6. Use `[N]byte` arrays instead of slices for fixed-size data.

---

## Level 4 — Expert (Scheduler, GC Internals, System Design)

---

### Q21. What is goroutine leak? How do you detect and prevent it in a large microservices codebase?

- A goroutine leak occurs when a goroutine is started but never terminates, consuming memory and potentially file descriptors, connections, or other resources indefinitely.
- Unlike threads, goroutines are cheap (2KB initial stack) but not free — thousands of leaked goroutines degrade GC performance and can exhaust memory.

**Common causes:**
- Goroutine blocked forever on a channel receive with no sender.
- Goroutine blocked on a mutex that is never unlocked.
- HTTP server handlers that never return because a context is never cancelled.
- Worker goroutines waiting on a queue that was abandoned.

**Detection:**
- `runtime.NumGoroutine()` — monitor in metrics (sudden growth = leak).
- pprof goroutine profile: `go tool pprof http://localhost:6060/debug/pprof/goroutine`.
- `goleak` library in tests: `goleak.VerifyNone(t)` asserts no leaked goroutines after test.
- Datadog / Prometheus alert on goroutine count rate of change.

**Prevention patterns:**
- Always pass `context.Context` and respect `ctx.Done()`.
- Use done channels or `WaitGroup`s to signal shutdown.
- Set timeouts on all blocking operations.
- Never start a goroutine without a clear lifecycle owner.

---

### Q22. Design a worker pool with dynamic scaling, backpressure, and graceful shutdown for a high-throughput job processing service.

This is a classic architecture question that tests understanding of goroutine lifecycle, backpressure, and graceful degradation.

**Before designing, ask:**
- What is the expected throughput? (100 jobs/sec or 100k jobs/sec?)
- Are jobs CPU-bound or IO-bound?
- Can jobs be retried?
- What is acceptable latency?
- Do producers block when the queue is full, or should jobs be rejected?
- Do we need ordering guarantees?
- Are jobs persisted or in-memory only?

**Key design decisions:**
1. **Bounded job queue** (buffered channel) — defines the backpressure point. When full, producers block or receive a rejection error depending on SLA.
2. **Dynamic worker count** — controlled by a semaphore or by spawning/retiring goroutines based on queue depth.
3. **Graceful shutdown** — close the job channel, drain it, wait for all workers to finish with `sync.WaitGroup`.
4. **Metrics** — expose queue depth, active workers, processing latency for auto-scaling triggers.
5. **Error handling** — dead letter queue for jobs that fail after N retries.

---

### Q23. Explain Go's scheduler (GMP model). How does it affect how you write high-performance concurrent code?

Go uses a user-space scheduler called the GMP model:
- **G** = Goroutine (user-space lightweight thread, 2KB initial stack, grows to 1GB)
- **M** = Machine (OS thread, maps to CPU)
- **P** = Processor (execution context, holds run queue of goroutines)

- `GOMAXPROCS` sets the number of Ps (default = number of logical CPUs). Each P has a local run queue of runnable Gs. Ms are OS threads — Go creates/destroys them dynamically. A G runs on an M with a P attached.

**Key scheduler behaviours:**
- **Work stealing** — idle Ps steal half the run queue from busy Ps — automatic load balancing.
- **Preemption** — Go 1.14+ uses signal-based preemption (SIGURG) so long-running loops don't starve other goroutines.
- **Syscall/cgo** — when G makes a blocking syscall, P is detached from M. A new M (or parked M) is attached to P to keep other Gs running. When the syscall returns, G rejoins the run queue.
- **Network poller** — network I/O is non-blocking; goroutines waiting on network are parked in the netpoller and reinjected when I/O is ready.

**Architectural implications:**
- **CPU-bound work** — set `GOMAXPROCS` to num CPUs; divide work into num-CPU goroutines. More goroutines = more scheduling overhead.
- **I/O-bound work** — goroutine-per-connection is fine because goroutines block cheaply; scheduler handles multiplexing automatically.
- **Avoid syscall-heavy loops** — too many goroutines in syscall exhausts M count (capped at 10,000 by default via `GOMAXTHREADS`).

**Visual model:**

```
┌─────────────────────────────────────────────────────────────┐
  │                     GLOBAL RUN QUEUE                        │
  │  [ G7 ] ──> [ G8 ] ──> [ G9 ] ──> [ G10 ]                  │
  └──────────────────────────────┬──────────────────────────────┘
                                 │ (Fallback if local queues empty)
                                 ▼
 ┌───────────────────────────────────────────────────────────────┐
 │                      GO RUNTIME SYSTEM                        │
 └───────┬───────────────────────────────┬───────────────────────┘
         │                               │
         ▼ (GOMAXPROCS slots)            ▼
┌──────────────────┐            ┌──────────────────┐
│   Processor P1   │            │   Processor P2   │
│ ──────────────── │            │ ──────────────── │
│ Local Run Queue  │            │ Local Run Queue  │
│ [ G2 ][ G3 ][ G4 ]│            │ [ G5 ][ G6 ]     │ <───┐
└────────┬─────────┘            └────────┬─────────┘     │
         │                               │               │ WORK STEALING
         │ (Attached)                    │ (Attached)    │ (P2 steals half
         ▼                               ▼               │  from busy P1)
┌──────────────────┐            ┌──────────────────┐     │
│  OS Thread M1    │            │  OS Thread M2    │ ────┘
└────────┬─────────┘            └────────┬─────────┘
         │                               │
         ▼ (Executing)                   ▼
   ┌───────────┐                   ┌───────────┐
   │    G1     │                   │    G5     │
   └───────────┘                   └───────────┘
         │                               │
         ▼                               ▼
┌──────────────────┐            ┌──────────────────┐
│  CPU Core 0      │            │  CPU Core 1      │
└──────────────────┘            └──────────────────┘
```

---

### Q24. How does Go's garbage collector work? How do you architect services to minimise GC pressure at scale?

Go uses a concurrent, tri-colour mark-and-sweep garbage collector that runs mostly concurrently with the application (since Go 1.5+).

**Key properties:**
- **Tri-colour** — objects are white (not yet seen), grey (seen but children not scanned), black (fully scanned). GC marks all reachable objects then sweeps white (unreachable) objects.
- **Stop-the-world (STW) pauses** — two brief STW pauses per cycle: one to enable write barriers, one to finalise marking. In Go 1.17+, typical pauses are <500 microseconds.
- **Write barrier** — during concurrent marking, writes to pointers are intercepted to maintain invariants (Dijkstra/hybrid barrier).
- **GOGC** — controls the ratio of new allocation to live heap size before triggering GC. Default 100 means GC triggers when heap doubles. Lower = more frequent GC; higher = less GC but larger heap.
- **GOMEMLIMIT** (Go 1.19+) — set a soft memory limit; GC becomes more aggressive as the limit is approached.

**Architectural patterns to reduce GC pressure:**
1. `sync.Pool` — reuse short-lived allocations (buffers, decoders, objects) across goroutines.
2. Avoid pointer-heavy data structures — slices of structs (value types) generate far less GC work than slices of pointers.
3. Pre-allocate slices/maps with known capacity — eliminates mid-operation reallocation.
4. Use arena/slab allocators for hot paths (manually manage memory in performance-critical services).
5. Escape analysis — use `go build -gcflags="-m"` to see what escapes to heap; restructure to keep objects on stack.
6. Batch small allocations — prefer fewer large allocations to many small ones.

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

### Q26. How do you implement distributed tracing across microservices in Go without coupling business logic to observability concerns?

The key principle is to make tracing transparent to business logic using context propagation and middleware patterns.

**Architecture layers:**
1. **Instrumented transport layer** — wrap `http.Client` and `http.Server` (or gRPC interceptors) to inject/extract trace context automatically. Business code just passes `ctx`.
2. **Context propagation** — trace context (trace ID, span ID, baggage) lives in `context.Context`. Never put it in struct fields or global variables.
3. **OpenTelemetry SDK** — use the OTLP exporter to send spans to Jaeger, Honeycomb, or Datadog. The SDK is pluggable — you can swap backends without changing application code.
4. **Automatic instrumentation** — use `otelhttp.NewHandler` for HTTP servers, `otelgrpc` interceptors for gRPC — zero business code change.
5. **Manual spans** only for significant business operations (database queries, external API calls, critical business logic boundaries).
6. **Sampling** — head-based sampling (1% of requests get full trace) vs tail-based sampling (keep traces with errors or high latency). Implement tail-based sampling with a collector like OpenTelemetry Collector.

**Follow-up: How would you implement correlation ID propagation across async Kafka messages in Go?**
- In an async message-driven pipeline using Kafka, `context.Context` remains our delivery vehicle.
- However, because we are crossing a network boundary asynchronously, we cannot pass the context directly in memory.
- Instead, we must serialize the tracking metadata into the Kafka Record Headers, transmit it across the broker, and deserialize it back into a fresh `context.Context` on the consumer side.

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

### Q28. Design a distributed cache in Go with consistent hashing, replication, and automatic failover. What are the key Go-specific implementation decisions?

This tests system design depth combined with Go-specific implementation knowledge.

**Architecture:**
1. **Consistent hashing ring** — use a virtual node ring (100-200 virtual nodes per physical node) to distribute keys evenly. Implemented as a sorted slice of uint32 hashes with binary search for O(log n) lookup.
2. **Node membership** — use a gossip protocol (memberlist library) or a coordinator (etcd) for node discovery and failure detection. Gossip is decentralised; etcd gives stronger consistency.
3. **Replication** — write to N nodes (replication factor), read from R nodes, where R+W > N for strong consistency. Default N=3, W=2, R=2 (like Dynamo).
4. **Failover** — when a node fails, gossip detects it within 1-2 gossip intervals (~500ms). The ring re-routes requests to the next node automatically.
5. **Slab allocator** — pre-allocate fixed-size memory slabs to avoid GC pressure from many small value allocations.
6. **Eviction** — LRU with a doubly-linked list + hash map. Use a sharded design (64 shards) to reduce mutex contention.

**Go-specific decisions:**
- Use `sync.RWMutex` per shard (not global lock) for the hash map.
- Use `sync.Pool` for request/response buffers.
- Use `encoding/gob` or protobuf for node-to-node serialisation (faster than JSON).
- `net.Listener` with goroutine-per-connection model — Go scheduler handles C10K easily.
- `context.Context` for all RPC calls with configurable timeouts.

---

## Level 5 — Architect (Distributed Systems & Production Playbooks)

---

### Q29. What is a Saga Orchestrator? Explain it simply, then describe the Go implementation rules.

**Simple explanation:**
- Imagine you are booking a vacation. To complete the trip, you need to book a Flight, a Hotel, and a Rental Car.
- If the hotel fails to book, you can't just stop — you need to cancel the flight you already paid for, otherwise you lose money.
- In microservices, a Saga Orchestrator is like a specialized travel agent. Instead of services talking randomly to each other, the Orchestrator sits in the middle. It gives orders to the Order, Payment, and Inventory services one by one. If a service succeeds, it moves to the next. If a service fails, the orchestrator acts as a coordinator, going backward and triggering Compensating Transactions (refunds/rollbacks) to clean up the mess.

**1. Core architecture:**
- **The brain (state machine)** — The Orchestrator is a central manager that tracks the transaction's current step (e.g., `PaymentPending`, `InventoryAllocated`). It saves this state into a database (like Postgres) so that if the server crashes, it remembers where it left off.
- **The steps** — Every service action is broken into a basic Go interface with two explicit commands:
  - `Execute(ctx)` — Run the forward action (e.g., charge the card).
  - `Compensate(ctx)` — Run the rollback action (e.g., refund the card).

**2. Key Go implementation rules:**
- **At-least-once delivery (the Outbox Pattern)** — Don't make live API calls inside your main database logic. Instead, save the Saga's current state and a "command message" into a local database table inside a single, safe transaction. A background Go worker reads this table and safely delivers the commands to the other services.
- **The timeout trap** — Every step must have a strict `context.WithTimeout` deadline. If a service hangs and doesn't answer, the orchestrator stops waiting and automatically starts the rollback chain.
- **Idempotency (safety locks)** — Networks fail, meaning messages get sent twice. Downstream services must track unique request keys. If the Payment service receives a duplicate "Refund" command, it must recognize it and return a safe `200 OK` rather than charging or refunding the customer a second time.
- **The critical insight** — Sagas do not offer instant database consistency. For a few seconds, money might be deducted before the stock is verified. Design your application's user interface to gracefully handle these temporary intermediate states (e.g., showing the user a "Processing" screen).

**Follow-up: How do you handle a compensating transaction that also fails?**
- Compensating transactions must be designed to be idempotent and retriable.
- If a compensation fails, persist the saga state and schedule retries through a durable mechanism such as a queue or workflow engine.
- The system remains eventually consistent while retries continue.
- If compensation repeatedly fails or requires external intervention, surface the failure through alerts and operational workflows rather than assuming rollback succeeded.
- This prevents resource leaks and ensures recovery even after process restarts.
- *(This answer demonstrates understanding of: Saga pattern, idempotency, durable state, retries, eventual consistency, and operational recovery strategies.)*

---

### Q30. How do you design for graceful degradation and resilience in a Go microservices architecture?

Resilience is designed in layers — each layer handles a different class of failure.

- **Layer 1 — Timeouts:** every outbound call (HTTP, gRPC, DB, Redis) must have an explicit timeout via `context.WithTimeout`. Never use `http.DefaultClient` (no timeout). Failing to set timeouts is the single most common cause of cascading failure.
- **Layer 2 — Retries with exponential backoff + jitter:** retry transient errors (network blips, 503s) with increasing delay. Jitter prevents thundering herd. Do NOT retry on 400, 409, 401 — these are client errors that won't resolve with retries.
- **Layer 3 — Circuit breaker:** after N consecutive failures, open the circuit and return a fast fallback error for T seconds. Half-open state tests recovery. Use `sony/gobreaker` or `failsafe-go`.
- **Layer 4 — Bulkhead:** isolate resource pools per downstream dependency. If the payment service is slow, it should not exhaust the connection pool used by the inventory service. Implement with separate `http.Client` instances with separate transport pools.
- **Layer 5 — Fallback:** graceful degradation — return stale cached data, a default value, or a partial response rather than an error when possible. Returning degraded results is often better than returning errors.
- **Layer 6 — Health checks and readiness:** implement `/healthz` (liveness) and `/readyz` (readiness) endpoints. Kubernetes uses readiness to stop routing traffic during degradation.

---

### Q31. How would you design the API layer for a multi-tenant SaaS platform in Go, including authentication, authorization, and tenant isolation?

- For a multi-tenant SaaS platform, separate authentication, authorization, and tenant isolation into distinct layers.
- **Authentication** is typically handled through OAuth2/OIDC with JWTs that contain tenant and user identity claims.
- **Authorization** is enforced through RBAC or policy-based controls.
- **Tenant isolation** is implemented at multiple layers:
  - Tenant-aware APIs.
  - Repository methods that always require a tenant context.
  - Database filtering on `tenant_id`.
  - Ideally, database-level protections such as Row-Level Security.
- Caches, queues, logs, and rate limits must also be tenant-aware to prevent cross-tenant data leakage and noisy-neighbor issues.
- The design follows a defense-in-depth approach where a single bug cannot compromise tenant isolation.

---

### Q32. How do you design a Go microservice for optimal operation in Kubernetes, covering resource management, health checks, and graceful shutdown?

- Design Kubernetes-native Go services with explicit resource requests and limits, bounded worker pools, and backpressure mechanisms to ensure predictable resource usage.
- Implement separate startup, readiness, and liveness probes so Kubernetes can accurately manage pod lifecycle without causing unnecessary restarts.
- **For graceful shutdown:** handle SIGTERM, mark the pod unready, stop accepting new work, drain in-flight requests, and flush telemetry before exiting.
- Expose Prometheus metrics, structured logs, and OpenTelemetry traces for observability.
- Use timeouts, retries, circuit breakers, and bulkhead isolation to maintain resilience under failure conditions.
- This ensures the service behaves predictably during deployments, failures, and scaling events.

---

### Q33. What happens when `GOMAXPROCS` > container CPU limit? What are the symptoms?

- When `GOMAXPROCS` exceeds the container CPU quota, the Go scheduler assumes more CPUs are available than Kubernetes actually allows.
- This leads to excessive context switching, CPU throttling, increased latency, and sometimes degraded GC performance.
- **Mitigation:** verify that `GOMAXPROCS` aligns with container CPU limits and monitor CFS throttling metrics to detect this condition.

---

### Q34. How would you implement pod autoscaling based on a custom application metric (e.g., queue depth)?

**Flow:**
```
Application
     ↓
Prometheus Metric
     ↓
Prometheus Adapter
     ↓
Custom Metrics API
     ↓
HPA
```

- CPU is often a poor signal. Example: Queue depth = 50,000, but CPU = 15%.

**Steps:**
1. Export queue depth as an application metric.
2. Expose metrics via a `/metrics` endpoint (Prometheus format).
3. Configure the Prometheus Adapter to surface this metric to the Kubernetes Custom Metrics API.
4. Configure the HPA to scale based on that custom metric.

---

### Q35. Design a structured logging and metrics strategy for a Go microservices platform with 50+ services.

At scale, observability is a product decision, not just an engineering one. The key is consistency across services.

**Structured logging:**
- Use `slog` (Go 1.21 standard library) with JSON handler in production, text handler in development. Consistent key names across all services (`request_id`, `tenant_id`, `duration_ms`, `error`).
- **Log levels** — ERROR for errors requiring human attention, WARN for recoverable anomalies, INFO for request summaries (one log per request at service boundary), DEBUG for development (disabled in production).
- **Correlation IDs** — inject at API gateway, propagate via context, include in every log entry. Enables tracing a request across 10 services in a single log query.
- **Never log sensitive data** — PII, credentials, card numbers. Use field-level redaction middleware.
- **Log sampling** — for very high traffic services, sample INFO logs (1 in 100) to reduce cost while keeping all ERROR and WARN logs.

**Metrics with Prometheus:**
- **Four golden signals** — latency (histogram), traffic (counter), errors (counter by error type), saturation (gauge: queue depth, goroutine count, CPU).
- Use histograms with carefully chosen buckets (.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5) — don't use defaults.
- Business metrics alongside technical: orders per second, payment success rate, active sessions.
- Alert on symptoms (p99 latency > SLO) not causes (CPU > 80%) — cause-based alerts are noisy and miss novel failures.

---

### Q36. How would you diagnose a p99 latency spike that doesn't show up in CPU or memory metrics?

1. Verify p50/p95/p99.
2. Identify the affected endpoint/region/tenant.
3. Open distributed traces.
4. Check downstream dependencies.
5. Check queue depth.
6. Check connection pools.
7. Capture a goroutine profile.
8. Capture a block profile.
9. Capture a mutex profile.
10. Investigate network and retries.

---

### Q37. You are designing a new Go-based platform from scratch for 10M users. Walk through your technology choices, service decomposition, and the first 6 months of architecture decisions.

This is a strategic architecture question. The best answers demonstrate understanding of evolutionary design — starting simple and adding complexity only when justified by real problems.

**Month 1-2 — Monolith first:**
- Start with a modular monolith in Go. Use clear package boundaries (`order`, `payment`, `user` packages) that mirror eventual service boundaries. This gives you the speed of a monolith with the preparedness of microservices.
- Single Postgres database with schema-per-domain. Use `pgx` (not `database/sql`) for performance and features.
- REST API with `chi` or stdlib `net/http`. No gRPC yet — simpler for initial team velocity.
- Deploy to Kubernetes from day one — container discipline matters even in the monolith phase.

**Month 3-4 — First pain points emerge:**
- Identify the first service boundary: usually auth or notification — high-change velocity, different scaling needs, clear interface.
- Extract to a microservice. Set up a service mesh (Istio or Linkerd) for mTLS, observability, and traffic management.
- Introduce Kafka for async events between services. Outbox pattern from the start.

**Month 5-6 — Operational maturity:**
- OpenTelemetry tracing across all services — bake into the base service template.
- Feature flags (LaunchDarkly/Unleash) — decouple deployment from release.
- Chaos engineering — GameDays testing failure scenarios before they happen in production.
- Developer platform — internal service template with all standards pre-wired (logging, metrics, tracing, health checks, graceful shutdown).

**Key principle:** every complexity you add must solve a problem you currently have, not one you might have at 100M users.

---


# Advanced Go — Interview Preparation (Part 2)
### Core Language Fundamentals — Beginner → Expert

This is the companion file to `Advanced-go.md`. That file covers concurrency, GC, and system design in depth. This file fills the gap: slices, defer/panic/recover, interfaces, generics, error handling, maps, strings, and testing — the fundamentals almost every Go interview checks first.

---

## Table of Contents

- [Level 1 — Beginner](#level-1--beginner)
- [Level 2 — Intermediate](#level-2--intermediate)
- [Level 3 — Advanced](#level-3--advanced)
- [Level 4 — Expert](#level-4--expert)

---

## Level 1 — Beginner

---

### Q1. What is the difference between `len` and `cap` on a slice?

- `len` — number of elements currently in the slice.
- `cap` — number of elements the underlying array can hold before a new array must be allocated.
- A slice is a view over an array: `{pointer, len, cap}`.

```go
s := make([]int, 3, 10) // len=3, cap=10
s = append(s, 1)        // len=4, cap=10 — no reallocation, room available
```

---

### Q2. What is the classic slice-aliasing bug?

- Slices share the same backing array when one is derived from another (via re-slicing).
- Modifying one can silently modify the other.

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]    // b shares a's backing array
b[0] = 99      // a is now {1, 99, 3, 4, 5} — surprise!
```

- **Fix:** use `copy()` to get an independent slice when you need isolation.

---

### Q3. What happens when `append()` exceeds capacity?

- If `len == cap`, `append` allocates a **new**, larger backing array (usually 2x for small slices, ~1.25x for large ones) and copies all elements over.
- The original slice variable is unaffected; you must use the returned slice.

```go
s := make([]int, 2, 2)
s2 := append(s, 1) // new array allocated, s and s2 no longer share memory
```

- **Practical pitfall:** appending to a slice passed into a function does not affect the caller's slice unless the function returns the new slice (or capacity was sufficient and same array is reused, which is fragile to rely on).

---

### Q4. Explain `defer`. What order do multiple defers run in?

- `defer` schedules a function call to run when the surrounding function returns (including via `panic`).
- Multiple defers run in **LIFO** order (last-in, first-out) — like a stack.

```go
func demo() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// Output: 3, 2, 1
```

- Common uses: closing files/connections, unlocking mutexes, recovering from panics.

---

### Q5. What is the classic `defer` in a loop pitfall?

- `defer` inside a loop does **not** run until the enclosing function returns — not at the end of each loop iteration.
- This can leak resources (file handles, locks) until the function exits.

```go
func processAll(files []string) {
    for _, f := range files {
        file, _ := os.Open(f)
        defer file.Close() // BUG: all files stay open until processAll returns
    }
}

// Fix: wrap the body in its own function so defer fires per iteration
func processAll(files []string) {
    for _, f := range files {
        func() {
            file, _ := os.Open(f)
            defer file.Close() // closes at the end of each iteration
        }()
    }
}
```

---

### Q6. What is the zero value of common Go types?

| Type | Zero value |
|---|---|
| `int`, `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| pointer, slice, map, channel, func, interface | `nil` |
| struct | every field set to its own zero value |

- Go has no concept of "uninitialized" — every variable always has a usable zero value.

---

## Level 2 — Intermediate

---

### Q7. What is `panic` and `recover`? How do they interact with `defer`?

- `panic` stops normal execution, runs deferred calls up the stack, and crashes the program unless recovered.
- `recover` can only stop a panic if called **inside a deferred function**.

```go
func safeDivide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    result = a / b // panics on b == 0
    return
}
```

- `recover()` called outside a `defer`, or in a deferred function that wasn't directly invoked at panic time, does nothing — returns `nil`.

---

### Q8. Does `recover()` work across goroutines?

- **No.** A panic in one goroutine cannot be recovered by a `defer`/`recover` in a different goroutine.
- An unrecovered panic in any goroutine crashes the **entire program**, not just that goroutine.
- **Rule:** every goroutine that can panic needs its own `defer recover()` at its top level.

```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Println("recovered in goroutine:", r)
        }
    }()
    riskyWork()
}()
```

---

### Q9. Nil interface vs nil pointer — what's the classic gotcha?

- An interface value is internally `{type, value}`. It is only `== nil` when **both** the type and value are nil.
- A non-nil interface holding a nil pointer is **not equal to nil**.
```go
+-------------------+
| Type  = *Person   |
| Value = nil       |
+-------------------+

!= nil

type MyError struct{}
func (e *MyError) Error() string { return "boom" }

func doWork() error {
    var e *MyError = nil
    return e // returns an interface with type=*MyError, value=nil
}

err := doWork()
fmt.Println(err == nil) // false! — classic interview trap
```

- **Fix:** return `nil` directly, not a typed nil pointer, when there is no error.

---

### Q10. What is the empty interface `any` (`interface{}`) and when should you avoid it?

- `any` (alias for `interface{}` since Go 1.18) can hold a value of any type.
- Useful for truly generic containers (e.g., `encoding/json` unmarshalling into unknown shapes).
- **Avoid it when:**
  - You lose compile-time type safety — every read requires a type assertion.
  - It causes heap allocation (boxing) for concrete values assigned to it — bad in hot paths.
  - Generics (Go 1.18+) can usually express the same intent with type safety.

---

### Q11. Type assertion vs type switch — what's the difference?

```go
// Type assertion — checking a single expected type
v, ok := i.(string)
if !ok {
    // i is not a string
}

// Type switch — checking against multiple possible types/ multi branch
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("unknown type")
}
```

- Always use the two-value form (`v, ok := i.(T)`) unless you are certain of the type — the single-value form panics on mismatch.

---

### Q12. How does Go implement error wrapping? What is `errors.Is` vs `errors.As`?

- Since Go 1.13, errors can be wrapped with `%w` in `fmt.Errorf`, preserving the original error in a chain.
- Error wrapping works like a linked list (or a Russian nesting doll). When you wrap an error, you create a new error structure that holds a pointer to the original error inside it.

```go
var ErrNotFound = errors.New("not found")

func getUser(id string) error {
    return fmt.Errorf("getUser failed: %w", ErrNotFound)
}

err := getUser("123")
errors.Is(err, ErrNotFound)   // true — walks the chain comparing by value/Is()
```

- `errors.Is(err, target)` — checks if `target` appears anywhere in the wrap chain (for **sentinel errors**). Used for comparing specific error values.
- `errors.As(err, &target)` — checks if any error in the chain can be assigned to `target`'s type (for **custom error types**), and populates it if so. Used for extracting specific error types.


```go
type ValidationError struct{ Field string }
func (e *ValidationError) Error() string { return "invalid: " + e.Field }

var ve *ValidationError
if errors.As(err, &ve) {
    fmt.Println("bad field:", ve.Field)
}
```

Why does errors.As take a pointer?
Because it needs to populate the variable with the matching error from the chain. because errors.As needs to modify ve.

---

### Q13. What is struct embedding, and how does method promotion work?

- Go has no inheritance, but embedding a struct (or interface) promotes its fields and methods to the outer struct.

```go
type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct {
    Animal // embedded — Name and Speak() are promoted
    Breed string
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Labrador"}
fmt.Println(d.Speak()) // "Rex makes a sound" — promoted method
```

- If the outer struct defines a method with the same name, it **shadows** the embedded one (no overriding/virtual dispatch — it's just shadowing by name resolution).

---

### Q14. Why are Go maps not safe for concurrent use? What happens if you try?

- The built-in `map` type has no internal locking.
- Concurrent read+write (or write+write) from multiple goroutines without synchronization triggers Go's built-in concurrent map detector, which **panics immediately**: `fatal error: concurrent map read and map write`.
- **Fixes:** protect with `sync.Mutex`/`sync.RWMutex`, or use `sync.Map` for specific high-concurrency read-heavy patterns (rarely the best general-purpose choice — sharded maps usually outperform it).

---

## Level 3 — Advanced

---

### Q15. Why is map iteration order randomized in Go? What are the implications?

- Go intentionally randomizes the starting point of map iteration (since Go 1.0) to prevent developers from accidentally relying on insertion order, which the spec never guaranteed. 
- To allow the runtime to optimize and evolve the map implementation without compatibility concerns.
- **Implication:** never assume map iteration order is stable — including across multiple runs of the *same* program.
- If you need ordered output, extract keys into a slice and sort them explicitly.

```go
package main

import (
	"fmt"
	"sort"
)

func main() {

	// Step 1: Create a map
	ages := map[string]int{
		"Charlie": 30,
		"Alice":   25,
		"Bob":     28,
		"David":   35,
	}

	// Step 2: Create a slice to hold all keys
	keys := make([]string, 0, len(ages))

	// Step 3: Copy keys from map into slice
	for key := range ages {
		keys = append(keys, key)
	}

	// Step 4: Sort the keys alphabetically
	sort.Strings(keys)

	// Step 5: Iterate using sorted keys
	for _, key := range keys {
		fmt.Printf("%s -> %d\n", key, ages[key])
	}
}

```

---

### Q16. Why can't slices, maps, or functions be map keys (or compared with `==`)?

- Map keys must be comparable types (`==` must be well-defined). Slices, maps, and functions are **not comparable** — they can only be compared to `nil`.
- Reason: slices/maps are reference-like types whose "equality" would be ambiguous (same backing array? same contents? same length?). Go avoids this ambiguity by disallowing it entirely.
- **Workaround:** use an array (`[N]T`) instead of a slice as a key, or serialize the slice/map to a string/hash.
- arrays are comparable, because Arrays are fixed-length value types, so this comparison is well-defined.

---

### Q17. How are Go strings represented internally? What is the cost of converting between `string` and `[]byte`?

- A Go string is an immutable, read-only sequence of bytes — internally `{pointer, length}`, UTF-8 encoded by convention (not enforced).
- `string` ↔ `[]byte` conversions **copy the underlying data** — they are not free, because strings are immutable and `[]byte` is mutable; sharing memory between them would violate that guarantee.

```go
s := "hello"
b := []byte(s) // copies bytes
s2 := string(b) // copies bytes again
```

- **Hot-path tip:** avoid repeated string↔[]byte conversions in loops; convert once and reuse, or use `strings.Builder` / `bytes.Buffer` to build up content without intermediate allocations.

---

### Q18. What is a `rune` in Go, and why does `len(s)` not always equal the number of "characters"?

- A `rune` is Go's alias for `int32`, representing a single Unicode code point.
- `len(s)` on a string returns the number of **bytes**, not characters — for multi-byte UTF-8 characters, this differs from the visible character count.

```go
s := "héllo"
fmt.Println(len(s))             // 6 (bytes) — "é" is 2 bytes in UTF-8
fmt.Println(len([]rune(s)))     // 5 (runes/characters)

for i, r := range s {
    fmt.Println(i, r) // ranging over a string yields rune values, byte-index positions
}
```

---

### Q19. Generics in Go (1.18+) — what problem do they solve, and what is a type constraint?

- Before generics, writing a function that worked across multiple types required either `interface{}` (losing type safety) or code generation/duplication.
- Generics let you write one function parameterized over a type, with compile-time type safety.

```go
// Constraint: T must support ordering operators
type Ordered interface {
    int | int64 | float64 | string
}

func Max[T Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

Max(3, 5)       // works for int
Max(3.1, 2.9)   // works for float64
Max("a", "b")   // works for string
```

- The standard library provides common constraints in `golang.org/x/exp/constraints` and built-in `comparable`.
- **When to prefer generics over `interface{}`:** when you need type safety and the same logic across multiple concrete types (e.g., generic data structures: stacks, sets, trees). When to prefer interfaces: when behavior (not just storage) varies by type — i.e., you need different method implementations, not just a different stored type.

---

### Q20. What is the difference between a value receiver and a pointer receiver, and when does it matter for interface satisfaction?

```go
type Counter struct{ n int }

func (c Counter) ValueInc()   { c.n++ }   // modifies a copy — no effect on original
func (c *Counter) PointerInc() { c.n++ }  // modifies the original
```

- A **value receiver** method gets a copy of the struct — mutations don't persist.
- A **pointer receiver** method can mutate the original and avoids copying large structs.

**Interface satisfaction gotcha:**
- If a type only has pointer receiver methods, only `*T` (not `T`) satisfies the interface.

```go
type Speaker interface{ Speak() }
func (c *Counter) Speak() {} // pointer receiver only

var s Speaker = Counter{}   // compile error: Counter does not implement Speaker
var s Speaker = &Counter{}  // OK
```

- **Rule of thumb:** if any method on a type needs a pointer receiver (for mutation or to avoid copying), make all methods on that type use pointer receivers for consistency.

---

## Level 4 — Expert

---

### Q21. How do table-driven tests work in Go, and why are they the idiomatic pattern?

- Table-driven tests define a slice of test cases (inputs + expected outputs) and iterate over them with `t.Run` for named subtests.
- Idiomatic because it scales to many cases without duplicating test logic, and failures report exactly which case failed.

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -1, -1, -2},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

---

### Q22. How do you mock dependencies in Go without a mocking framework? What is the interface-based pattern?

- Go favors small interfaces defined at the point of use (consumer-defined interfaces), making mocking trivial without any library.

```go
// Production code depends on a narrow interface
type UserStore interface {
    GetUser(id string) (*User, error)
}

type Service struct{ store UserStore }

// Test — a hand-written fake satisfying the same interface
type fakeStore struct{ user *User; err error }
func (f *fakeStore) GetUser(id string) (*User, error) { return f.user, f.err }

func TestService_GetUser(t *testing.T) {
    svc := Service{store: &fakeStore{user: &User{Name: "Alice"}}}
    // assertions against svc...
}
```

- Libraries like `gomock` or `testify/mock` generate this boilerplate automatically for larger interfaces, but the underlying mechanism is exactly this — substitutable interface implementations.

---

### Q23. How does `httptest` let you test HTTP handlers without starting a real server?

```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
    rec := httptest.NewRecorder() // captures the response in memory

    GetUserHandler(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("got status %d, want 200", rec.Code)
    }
}
```

- `httptest.NewRecorder()` implements `http.ResponseWriter` in memory — no actual socket/port is opened, making tests fast and parallel-safe.
- `httptest.NewServer()` is used instead when you need a *real* listening server (e.g., testing an `http.Client` against it, or testing redirects/cookies end-to-end).

---

### Q24. What is the difference between `go vet`, `golangci-lint`, and the race detector — and when do you run each?

| Tool | Purpose | When to run |
|---|---|---|
| `go vet` | Catches suspicious constructs the compiler allows but are likely bugs (e.g., wrong `Printf` verbs, unreachable code) | Every build / CI |
| `golangci-lint` | Aggregates many linters (style, unused code, complexity, security) into one fast runner | CI, pre-commit |
| `go test -race` | Instruments memory accesses to detect actual **data races** at runtime (not compile time) | CI, especially for concurrency-heavy code — adds runtime overhead so not used in production |

- The race detector only catches races that occur **during the test run** — it cannot prove the absence of races, only catch ones triggered by the executed code path.

---

### Q25. What is `go.sum` and how does it differ from `go.mod`? How does Go ensure reproducible builds?

- `go.mod` declares your module's direct and indirect dependencies and their required **versions**.
- `go.sum` records cryptographic **checksums** of every module version your build touches — both the module zip and its `go.mod` file.
- On every build, Go verifies downloaded modules against `go.sum` — if a checksum doesn't match, the build fails. This protects against a compromised proxy or registry silently serving different code for the same version tag.
- **Reproducibility:** with `go.mod` + `go.sum` committed, any machine running `go build` gets byte-identical dependency code, regardless of when or where the build happens (assuming the Go toolchain version itself is also pinned, e.g. via `go.mod`'s `go` directive or a `toolchain` line).

---

*End of Part 2. Pair this with `Advanced-go.md` (concurrency, GC, system design) for full-spectrum coverage — fundamentals here, deep systems there.*