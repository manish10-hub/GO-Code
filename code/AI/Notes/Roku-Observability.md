# Roku Senior Software Engineer — Observability
## Interview Preparation Guide
### Tailored to Interviewer: Wiard van Rij (Sr. SWE Observability, Roku | Thanos team member | KubeCon speaker | CNCF contributor)

---

## About Your Interviewer

<cite index="19-1">Wiard van Rij is a Senior Software Engineer in Observability at Roku and a team member on the Thanos project. He is an open-source enthusiast and CNCF community contributor.</cite> <cite index="26-1">His main focus is observability, Kubernetes, and cloud-native challenges, with specific expertise in Thanos — the HA, long-term storage extension for Prometheus.</cite> <cite index="17-1">He actively attends KubeCon and engages with the CNCF community, and has been growing Roku's global observability team.</cite>

**What this means for your interview:**
- He will go deep on **Prometheus internals, Thanos architecture, and TSDB design** — not just surface-level "I've used Prometheus"
- He cares deeply about **open-source contribution and CNCF ecosystem** — showing awareness of upstream projects scores well
- He values **ownership, fast iteration, and practical problem-solving** — every answer should connect to real systems you've built
- He will probe **distributed systems thinking** — expect questions about scale, failure modes, and trade-offs
- **Go is non-negotiable** — 4+ years expected; he will likely ask Go-specific observability implementation questions

---

## Table of Contents

- [Section 1 — Observability Foundations (Beginner)](#section-1--observability-foundations-beginner)
- [Section 2 — Prometheus & Metrics Deep Dive (Intermediate)](#section-2--prometheus--metrics-deep-dive-intermediate)
- [Section 3 — Thanos & Long-Term Storage (Advanced)](#section-3--thanos--long-term-storage-advanced)
- [Section 4 — Logs, Traces & OpenTelemetry (Intermediate–Advanced)](#section-4--logs-traces--opentelemetry-intermediateadvanced)
- [Section 5 — Data Pipelines & Storage at Scale (Advanced)](#section-5--data-pipelines--storage-at-scale-advanced)
- [Section 6 — Go for Observability Systems (Advanced)](#section-6--go-for-observability-systems-advanced)
- [Section 7 — Kubernetes & Cloud-Native Observability (Intermediate–Advanced)](#section-7--kubernetes--cloud-native-observability-intermediateadvanced)
- [Section 8 — System Design Scenarios (Expert)](#section-8--system-design-scenarios-expert)
- [Section 9 — Open Source, CNCF & Cost Management (Advanced)](#section-9--open-source-cncf--cost-management-advanced)
- [Section 10 — Behavioural & Culture Fit](#section-10--behavioural--culture-fit)

---

## Section 1 — Observability Foundations (Beginner)

---

### Q1. What is observability and how is it different from monitoring?

**Answer:**

**Monitoring** is about tracking known failure modes — you define dashboards and alerts for conditions you have already anticipated. It answers: "Is the system in a state I expected it might reach?"
Is something wrong?

**Observability** is the ability to understand the internal state of a system purely from its external outputs — metrics, logs, and traces. It answers questions you did not think to ask before the system broke. A highly observable system lets you debug novel failure modes without deploying new instrumentation.
Why is it wrong?

The three pillars:
- **Metrics** — numeric measurements over time (request rate, error rate, latency, saturation). Cheap to store and query, but no context per individual request
- **Logs** — discrete, structured events with rich context. Expensive at scale, but invaluable for debugging specific failures
- **Traces** — the path of a single request through a distributed system. Shows which service and which operation in that service caused the latency

**The practical distinction at Roku scale:** ingesting terabytes of observability data daily means you cannot treat monitoring and observability as the same thing. You must be deliberate about what you collect, how long you retain it, and how you query it efficiently — this is exactly the problem Thanos, Loki, and Tempo solve at scale.

Monitoring is primarily about detecting known failure conditions through predefined metrics and alerts, whereas observability is about understanding the internal state and behavior of a system using telemetry such as metrics, logs, and traces. Monitoring tells us that something is wrong; observability helps us determine why it is wrong and trace the problem to its root cause.

---

### Q2. What are the four golden signals and why do they matter?

**Answer:**

The four golden signals (from Google's SRE book) are the minimum set of metrics that give you a complete picture of a service's health:

1. **Latency** — time to serve a request. Distinguish successful request latency from error latency separately (errors that fail fast can make average latency look good while users are experiencing failures).

2. **Traffic** — demand placed on the system (requests per second, queries per second, events per second). Baseline for capacity planning and anomaly detection.

3. **Errors** — rate of failed requests. Both explicit (HTTP 5xx) and implicit (HTTP 200 with wrong content, or timeouts) errors matter.

4. **Saturation** — how "full" the service is — the constrained resource (CPU, memory, disk I/O, network). Most services degrade before reaching 100% saturation.

**Why they matter for an observability platform:** when Roku's observability platform is ingesting hundreds of millions of active series, these four signals apply to the observability infrastructure itself. Latency of metric queries, traffic volume to the ingest pipeline, error rates on scrapes, and disk saturation on TSDB nodes are all first-class concerns.

---

### Q3. What is a time-series database (TSDB) and what makes it different from a general-purpose database?

**Answer:**

A time-series database is optimised for storing and querying sequences of values indexed by time. Key differences from a general-purpose relational or document database:

**Write pattern:** TSDBs receive extremely high write rates (hundreds of thousands of samples per second) that are always append-only and time-ordered. No random writes, no updates to existing data — only new samples appended to existing series.

**Compression:** sequential numeric values with small deltas between timestamps compress extremely well. Prometheus TSDB uses delta-of-delta encoding for timestamps (the delta between consecutive deltas is usually 0 or 1) and XOR encoding for float64 values — achieving 1-2 bytes per sample versus the 8 bytes a float64 would normally occupy. This is the Gorilla compression algorithm from Facebook's paper.

**Query pattern:** the most common queries are time-range scans of specific series — "give me the last hour of CPU usage for this host". TSDBs optimise for this with column-oriented storage where all values for a series are stored together.

**Retention and compaction:** older data is less useful at full resolution. TSDBs implement downsampling (reduce 1-second resolution to 5-minute resolution for data older than 2 weeks) and compaction (merge many small blocks into fewer large blocks for efficient query).

**Not good at:** ad-hoc queries across many series with no time filter, joins across different metric names, or transactional consistency — this is why TSDBs are purpose-built, not general.

---

### Q4. Explain the RED method vs USE method for monitoring microservices.

**Answer:**

**RED method (for services — requests):**
- **R**ate — requests per second
- **E**rrors — failed requests per second
- **D**uration — distribution of request durations (latency percentiles)

Best for: user-facing services, APIs, microservices — anything that serves requests. Directly correlates to user experience.

**USE method (for resources — infrastructure):**
- **U**tilization — percentage of time the resource is busy
- **S**aturation — amount of work the resource cannot service (queue depth)
- **E**rrors — error events from the resource

Best for: CPUs, memory, disks, network interfaces — the infrastructure components that services run on.

**How they complement each other at Roku:** RED tells you "the streaming API's latency jumped to 2 seconds." USE tells you "the disk on the TSDB node is 95% saturated." Together they let you trace user-visible degradation back to a specific infrastructure bottleneck. A complete observability platform needs both layers instrumented.

---

## Section 2 — Prometheus & Metrics Deep Dive (Intermediate)

---

### Q5. Explain Prometheus's data model. What is a metric, label, and time series?

**Answer:**

Prometheus's data model is deceptively simple and enormously powerful.

**Metric:** a named measurement. Example: `http_requests_total`. The name follows the convention `namespace_subsystem_name_unit`.

**Label:** a key-value pair that adds dimensions to a metric. Example: `http_requests_total{method="POST", status="200", handler="/api/v1/stream"}`. Labels are what make Prometheus's data model multidimensional.

**Time series:** the combination of a metric name and a unique set of labels defines one time series. Each time series has an independent sequence of (timestamp, float64) samples.

```
http_requests_total{method="GET",  status="200"} → [1716000000, 42], [1716000015, 43], ...
http_requests_total{method="POST", status="200"} → [1716000000, 12], [1716000015, 14], ...
http_requests_total{method="GET",  status="500"} → [1716000000, 1],  [1716000015, 2], ...
```

**Critical understanding — high cardinality:**
Every unique combination of label values creates a new time series. `user_id` as a label with 10 million unique users creates 10 million time series — this is the "cardinality explosion" problem that is the #1 operational challenge at Roku scale (hundreds of millions of active series). The observability platform must enforce label cardinality limits, detect cardinality explosions, and help developers understand the cost of their instrumentation choices.

---

### Q6. What are the four Prometheus metric types and when do you use each?

**Answer:**

**Counter:**
- Monotonically increasing value; only ever goes up (or resets to 0 on restart)
- Use for: total requests, total errors, total bytes sent, total events processed
- Always use `rate()` or `increase()` in queries — the raw counter value is rarely useful
- Example: `http_requests_total`, `go_gc_cycles_total`

**Gauge:**
- Current value that can go up or down
- Use for: current memory usage, current queue depth, current active connections, temperature
- Can be directly used in queries without `rate()`
- Example: `process_resident_memory_bytes`, `kafka_consumer_lag`

**Histogram:**
- Samples observations into configurable buckets; tracks sum, count, and bucket counts
- Use for: request latency distribution, request/response sizes
- Enables `histogram_quantile()` — compute p50, p95, p99 latency from bucketed data
- **Trade-off:** bucket boundaries must be chosen upfront; wrong buckets lose resolution
- Example: `http_request_duration_seconds_bucket{le="0.1"}` counts requests <= 100ms

**Summary:**
- Like histogram but computes quantiles client-side and exports pre-calculated quantiles
- Use for: when you need accurate quantiles but cannot choose bucket boundaries in advance
- **Trade-off:** cannot be aggregated across instances (you cannot sum two summaries' p99s)
- Prefer histograms for microservices where you have multiple instances

**Practical recommendation:** prefer Histograms for latency at Roku scale — they aggregate across instances (crucial when you have hundreds of pods) and the Native Histograms feature (Prometheus 2.40+) removes the bucket-boundary problem.

---

### Q7. Explain PromQL — how does it work, and what are the key concepts?

**Answer:**

PromQL (Prometheus Query Language) is a functional query language for time-series data.

**Instant vector:** a set of time series, each with one sample at a single point in time.
```promql
http_requests_total   # returns current value of all matching series
```

**Range vector:** a set of time series, each with a range of samples over a time window.
```promql
http_requests_total[5m]   # returns last 5 minutes of samples for each series
```

**Key functions:**

```promql
# rate() — per-second rate of increase over a range (for counters)
rate(http_requests_total[5m])

# irate() — instant rate based on last two samples (more responsive, noisier)
irate(http_requests_total[1m])

# increase() — total increase over a range
increase(http_requests_total[1h])

# histogram_quantile() — compute quantile from histogram buckets
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# sum by() — aggregate across a label dimension
sum by (status) (rate(http_requests_total[5m]))

# without — aggregate dropping specific labels
sum without (instance) (rate(http_requests_total[5m]))

# topk() — top N series by value
topk(5, rate(http_requests_total[5m]))
```

**Staleness handling:** Prometheus considers a time series stale if no sample has been received within 5 minutes. Queries on stale series return no data. This is important for understanding gaps in graphs.

**Recording rules:** pre-compute expensive PromQL expressions and store the result as a new time series. Crucial at Roku scale — computing `sum(rate(http_requests_total[5m]))` across 100M series in real-time is expensive; pre-computing it as a recording rule and querying that is 100x cheaper.

---

### Q8. What is Prometheus's TSDB storage format? How does it organise data on disk?

**Answer:**

Prometheus TSDB (the storage engine, also used by Thanos) organises data into **blocks** on disk.

**Block structure:**
```
data/
├── 01BKGV7JBM69T2G1BGBGM6KB12/   ← block directory (ULID)
│   ├── chunks/
│   │   ├── 000001                  ← chunk files (raw compressed samples)
│   │   └── 000002
│   ├── index                       ← inverted index (label→series→chunk mapping)
│   ├── meta.json                   ← block metadata (time range, stats)
│   └── tombstones                  ← marks for deleted series
└── wal/                            ← Write-Ahead Log (recent data, not yet compacted)
    ├── 00000001
    └── 00000002
```

**WAL (Write-Ahead Log):**
- All incoming samples are first written to the WAL — a sequential append-only log
- WAL provides durability: if Prometheus crashes, it can replay the WAL to recover recent data
- WAL segments are 128MB each; Prometheus keeps the last 2 hours in the WAL

**Chunks:**
- Once the WAL has enough data, it is compacted into a block
- Each chunk holds 120 samples for one series (typically 2 hours of data at 1-minute scrape interval)
- Samples are compressed using Gorilla encoding (delta-of-delta for timestamps, XOR for values)
- Each chunk file is memory-mapped for efficient random access without reading the whole file into memory

**Index:**
- An inverted index maps label names → label values → series IDs → chunk positions
- This is how `http_requests_total{status="500"}` efficiently finds all matching series without scanning all chunks

**Compaction:**
- Prometheus periodically merges small blocks into larger blocks (2h → 6h → 1 day → 2 weeks)
- Larger blocks improve query performance (fewer files to open) and enable better compression across larger time ranges
- This is where Thanos takes over for long-term retention — it uploads completed 2-hour blocks to object storage

---

### Q9. What are recording rules and alerting rules in Prometheus? Why are they important at scale?

**Answer:**

**Recording rules:**
Pre-compute expensive PromQL expressions and store the result as a new, cheaper-to-query time series.

```yaml
groups:
  - name: streaming_aggregations
    interval: 30s
    rules:
      - record: job:http_requests_total:rate5m
        expr: sum by (job, status) (rate(http_requests_total[5m]))
      - record: job:http_request_duration_p99:rate5m
        expr: histogram_quantile(0.99,
                sum by (job, le) (rate(http_request_duration_seconds_bucket[5m])))
```

**Why they matter at Roku scale:**
- Without recording rules: every dashboard refresh computes `sum(rate(...))` across hundreds of millions of active series — this hits every TSDB block, saturates Thanos Query nodes
- With recording rules: the aggregation is pre-computed every 30s; dashboard queries hit a tiny set of pre-aggregated series — 100-1000x cheaper
- Mandatory for dashboards that aggregate over many high-cardinality series

**Alerting rules:**
```yaml
groups:
  - name: platform_alerts
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          / sum(rate(http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Error rate above 1% for 5 minutes"
```

- `for: 5m` — the condition must be true for 5 consecutive minutes before firing, preventing flapping alerts
- Alerting rules are evaluated in Prometheus; in a Thanos setup with HA Prometheus pairs, they are evaluated in Thanos Ruler to avoid duplicate alerts

---

## Section 3 — Thanos & Long-Term Storage (Advanced)

*(This section is the most important for Wiard's interview — he is a Thanos team member)*

---

### Q10. What problem does Thanos solve, and explain its core architecture?

**Answer:**

**The Prometheus limitations Thanos solves:**

1. **No global view:** when you have 50 Kubernetes clusters, each with its own Prometheus, you cannot query metrics across all clusters from a single place. You would have to query each Prometheus separately and mentally merge results.

2. **Local storage only:** Prometheus stores data on local disk. A pod with 200GB of disk cannot retain more than a few weeks of data at high cardinality.

3. **No HA without duplicates:** running two Prometheus instances for HA is correct, but when you query both you get duplicate results because both scrape the same targets.

4. **No cross-cluster aggregation:** federation (Prometheus scraping another Prometheus) is fragile and creates hierarchical bottlenecks.

**Thanos architecture — the components:**

```
┌─────────────────────────────────────────────────────────┐
│  Prometheus + Thanos Sidecar (per cluster)              │
│  ┌──────────────┐    ┌─────────────────────────────┐   │
│  │  Prometheus  │───▶│  Thanos Sidecar             │   │
│  │  (local TSDB)│    │  - Exposes StoreAPI         │   │
│  └──────────────┘    │  - Uploads 2h blocks to S3  │   │
│                      └─────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                              │ uploads blocks
                              ▼
                    ┌─────────────────┐
                    │  Object Storage  │  (S3, GCS, etc)
                    │  (long-term)     │
                    └─────────────────┘
                              │ reads blocks
                              ▼
              ┌───────────────────────────┐
              │  Thanos Store Gateway     │  serves historical data via StoreAPI
              └───────────────────────────┘
                              │
              ┌───────────────────────────┐
              │  Thanos Compactor         │  downsample + compact S3 blocks
              └───────────────────────────┘

              ┌───────────────────────────┐
              │  Thanos Query             │  fan-out to all StoreAPIs
              │  (global query frontend)  │  deduplicate HA replicas
              └───────────────────────────┘
                              │
              ┌───────────────────────────┐
              │  Thanos Query Frontend    │  caching, query splitting, retry
              └───────────────────────────┘
                              │
              ┌───────────────────────────┐
              │  Thanos Ruler             │  recording + alerting rules globally
              └───────────────────────────┘
```

**The key insight — StoreAPI:** Thanos defines a gRPC interface called StoreAPI that all components implement. The Sidecar, Store Gateway, and Ruler all speak StoreAPI. The Query component fans out to all StoreAPI endpoints, merges and deduplicates the results. This is what makes Thanos horizontally composable.

**Deduplication:** two HA Prometheus instances scrape the same targets and produce near-identical data. Thanos Query deduplicates using the `replica` external label — when it sees two series that are identical except for their `replica` label, it merges them into one, picking the value that exists in either.

---

### Q11. How does Thanos handle long-term storage? Explain the Sidecar and Store Gateway.

**Answer:**

**Thanos Sidecar:**
- Runs as a sidecar container alongside Prometheus (same Pod in Kubernetes)
- Two responsibilities:
  1. Exposes Prometheus's current (in-memory + recent on-disk) data via StoreAPI — this is how Thanos Query accesses recent data without downloading blocks
  2. Watches Prometheus's data directory; when a 2-hour block is finalized (flushed from WAL and compacted), the Sidecar uploads it to object storage (S3, GCS) with the Prometheus external labels attached

**Why 2-hour blocks?** Prometheus compacts WAL data into 2-hour blocks. This is the unit of work Thanos uploads — small enough for reasonable upload latency, large enough to get good compression.

**Thanos Store Gateway:**
- Reads blocks from object storage on demand
- Implements StoreAPI — Thanos Query treats it identically to a Sidecar
- Index-aware: downloads only the index and metadata files from S3 at startup; downloads chunk data lazily only when a specific time range is queried
- Caches the index in memory (and optionally in memcached/Redis) to avoid repeated S3 HEAD/GET calls for index files — critical for query latency

**Block metadata:**
Each block in S3 has a `meta.json`:
```json
{
  "ulid": "01BKGV7JBM69T2G1BGBGM6KB12",
  "minTime": 1716000000000,
  "maxTime": 1716007200000,
  "stats": { "numSamples": 47200332, "numSeries": 196421, "numChunks": 3862 },
  "compaction": { "level": 1, "sources": ["01BKGV7JBM69T2G1BGBGM6KB12"] },
  "thanos": {
    "labels": { "cluster": "prod-us-west-2", "replica": "0" },
    "downsample": { "resolution": 0 }
  }
}
```

---

### Q12. Explain Thanos Compactor — what does it do and why is it critical?

**Answer:**

Thanos Compactor runs as a single instance (it must be single-instance to avoid concurrent block modification) and performs two critical functions:

**1. Horizontal compaction:**
- Prometheus uploads 2-hour blocks to S3. After a while, querying a month of data means the Store Gateway must open hundreds of 2-hour blocks.
- The Compactor merges small blocks into larger blocks (2h → 1d → 2w) following the same compaction rules as Prometheus itself.
- Result: fewer blocks to open per query → better query performance for historical data.

**2. Downsampling:**
- Raw data (resolution 0 = original scrape interval) is kept for the configured retention period (e.g., 2 weeks)
- After 2 weeks, the Compactor creates 5-minute resolution downsampled blocks (1 sample per 5 minutes instead of 1 per 15 seconds) — 20x storage reduction
- After 1 month, the Compactor creates 1-hour resolution downsampled blocks — 240x storage reduction from original
- Thanos Query automatically selects the appropriate resolution based on the query's time range — a 1-year graph uses 1h resolution; a 30-minute graph uses raw resolution

**Why it is critical:**
- Without compaction: object storage fills with millions of tiny 2-hour blocks; Store Gateway query performance degrades as it opens thousands of files
- Without downsampling: a query for "last 1 year of CPU usage" downloads terabytes of raw data; with downsampling it downloads megabytes
- The Compactor also handles **block deletion** — after retention period expires, it marks blocks for deletion (soft delete first, then hard delete after a delay to allow rollback)

**Operational consideration:** the Compactor must never run as multiple instances (it would corrupt blocks). Use a Kubernetes Deployment with `replicas: 1` or a CronJob pattern.

---

### Q13. How does Thanos Query handle deduplication and fan-out? What are the performance implications?

**Answer:**

**Fan-out:**
When a PromQL query arrives at Thanos Query, it fans out the request to all registered StoreAPI endpoints simultaneously:
- Sidecar for cluster A (recent data, 0–2 hours)
- Sidecar for cluster B (recent data, 0–2 hours)
- Store Gateway (historical data, all clusters, 2h → months)

Each endpoint filters data by time range and label matchers and returns matching chunks. Thanos Query merges the streaming results from all endpoints.

**Deduplication:**
When two HA Prometheus instances (differentiated by `replica` label) both have data for the same series, Thanos Query deduplicates:
```
series A from replica=0: [10, 11, 12, _, 14]   (sample at t=3 missing due to scrape failure)
series A from replica=1: [10, 11, 12, 13, 14]   (all samples present)
merged result:            [10, 11, 12, 13, 14]   (gaps filled from the other replica)
```
This is how Thanos provides HA Prometheus without running separate HA infrastructure.

**Performance implications:**
- Fan-out multiplies the cost of every query by the number of StoreAPI endpoints
- A query across 50 clusters hits 50 Sidecars + 1 Store Gateway = 51 parallel gRPC calls
- Each endpoint streams chunk data; Thanos Query must buffer and merge it in memory
- For Roku-scale (hundreds of millions of series, multiple clusters): Thanos Query Frontend adds caching (results cache in memcached), query splitting (split a 30-day query into 30 1-day sub-queries), and query retries to make this manageable

**The Query Frontend:**
- Splits large time-range queries into smaller sub-queries that can be cached independently
- Caches query results in Redis/memcached — identical queries within the cache TTL return instantly
- This is the key to making dashboards with 30-day time ranges work without hitting every block in S3 on every dashboard refresh

---

## Section 4 — Logs, Traces & OpenTelemetry (Intermediate–Advanced)

---

### Q14. How does Grafana Loki work? How is it different from Elasticsearch?

**Answer:**

**Loki's core design principle:** "like Prometheus, but for logs."

**What Loki indexes:** only the labels attached to log streams (not the log line content). A log stream is defined by a unique set of label key-value pairs:
```
{app="roku-api", env="prod", pod="roku-api-7f9d-4xk2p"}
```

**What Elasticsearch indexes:** every word/token in every log message — a full-text inverted index.

**Implications:**

| | Loki | Elasticsearch |
|---|---|---|
| **Index size** | Tiny (only labels) | Large (full text index) |
| **Ingest cost** | Low | High (indexing CPU) |
| **Storage cost** | Low (chunks compressed like Prometheus) | High |
| **Query for known labels** | Very fast | Fast |
| **Full-text search** | Slower (requires scanning chunk data) | Fast |
| **Operational complexity** | Low | Very high (shards, replicas, mapping) |

**Loki's storage:**
Log lines are compressed and stored in chunks, similar to Prometheus TSDB. Chunks are flushed to object storage (S3). The index maps label combinations → chunk locations. This is the same object-storage-backed architecture as Thanos — cheap, durable, horizontally scalable.

**LogQL:** Loki's query language:
```logql
# Filter logs by label, then by content
{app="roku-api", env="prod"} |= "error" | json | level="ERROR"

# Extract a metric from logs (count errors per minute)
sum(rate({app="roku-api"} |= "error" [1m])) by (pod)
```

**When to choose Loki over Elasticsearch:** when you control how logs are produced and can attach meaningful labels; when cost at terabyte-per-day scale is a concern; when the team is already running Prometheus/Thanos (reuse of operational knowledge). Use Elasticsearch when you need full-text search across unstructured log content you do not control.

---

### Q15. Explain distributed tracing — what is a trace, span, and context propagation?

**Answer:**

**Trace:** a complete record of a single request's journey through a distributed system — from the user's browser/device to the API gateway to backend services to databases and back.

**Span:** a named, timed operation within a trace. Every function call, database query, HTTP call, or Kafka message that you want to measure is a span.

**Trace structure:**
```
Trace: "Watch stream initiation" (total: 280ms)
  ├── Span: "API Gateway" (0-5ms)
  ├── Span: "Auth Service" (5-20ms)
  │     └── Span: "DB: validate token" (10-18ms)
  ├── Span: "Content Service" (20-120ms)
  │     ├── Span: "DB: get channel metadata" (22-45ms)
  │     └── Span: "Cache: get stream URL" (45-48ms)
  └── Span: "CDN redirect" (120-280ms)
```

**Context propagation:** the mechanism by which trace context (trace ID, span ID, sampling decision) is passed across service boundaries — through HTTP headers (W3C TraceContext standard: `traceparent` header), Kafka message headers, or gRPC metadata.

Without context propagation, each service creates isolated spans that cannot be joined into a coherent trace. The trace ID is the single key that links all spans across all services into one trace.

**Tempo (Grafana's tracing backend):**
- Stores traces at object-storage scale — same architectural pattern as Loki and Thanos
- Indexes only the trace ID and some metadata (service name, operation name, duration)
- Retrieves full trace by trace ID without needing a full-text index
- Designed for cost-efficient storage of billions of spans

---

### Q16. What is OpenTelemetry? How does it differ from OpenTracing and OpenMetrics?

**Answer:**

**OpenTelemetry (OTel):** a CNCF project that provides vendor-neutral APIs, SDKs, and a collector for metrics, logs, and traces. It is the merger and successor of OpenTracing (traces) and OpenCensus (metrics + traces).

**The three parts of OpenTelemetry:**

1. **API:** language-specific interfaces for instrumentation. You instrument your code against the OTel API — no vendor lock-in.
2. **SDK:** the implementation of the API — handles sampling, batching, and export. Pluggable backends.
3. **Collector:** a standalone agent/gateway that receives telemetry, processes it (batching, filtering, routing), and exports it to backends (Prometheus, Jaeger, Tempo, Loki, Datadog, etc.)

```
Your Go service
  → OTel SDK (auto-instrumented: net/http, database/sql, gRPC)
  → OTLP (OpenTelemetry Protocol) via gRPC or HTTP
  → OTel Collector
  → [Prometheus | Tempo | Loki] backends
```

**OpenTracing:** was a specification for distributed tracing APIs only. Deprecated; OpenTelemetry is its successor.

**OpenMetrics:** an IETF standard that extends the Prometheus exposition format with explicit type definitions, created_timestamps, and unit metadata. Prometheus 2.x supports OpenMetrics format natively.

**Why OTel matters at Roku:** with hundreds of services emitting terabytes of telemetry daily, having all services use a consistent API means you can:
- Switch backends without re-instrumenting services
- Add a new backend (e.g., add ClickHouse for long-term log analytics) without touching service code
- Centrally control sampling rates at the Collector, not in service code
- Correlate logs, metrics, and traces using the same resource attributes and trace context

---

## Section 5 — Data Pipelines & Storage at Scale (Advanced)

---

### Q17. The job description mentions pre-aggregation, downsampling, and sampling. Explain each and when you would use them.

**Answer:**

**Pre-aggregation:**
Computing aggregations ahead of time (recording rules in Prometheus/Thanos terms) so that dashboards and alerts query pre-computed data rather than running expensive aggregations on raw data at query time.

```promql
# This is expensive at query time across 100M series:
sum by (region) (rate(http_requests_total[5m]))

# Pre-compute it as a recording rule:
record: region:http_requests:rate5m
expr: sum by (region) (rate(http_requests_total[5m]))

# Dashboard now queries cheap single-dimension series:
region:http_requests:rate5m
```

**When to use:** for any aggregation that is queried frequently (dashboards refreshing every 30s), especially aggregations that reduce high-cardinality series to low-cardinality summaries.

**Downsampling:**
Reducing resolution of historical data — replacing N samples in a time window with a single summary (min, max, sum, count) to reduce storage and query cost for long time ranges.

- Raw data (15s scrape interval) → keep for 2 weeks
- 5-minute downsampled data → keep for 1 month
- 1-hour downsampled data → keep indefinitely

**When to use:** for metrics where high-resolution historical data is not valuable (nobody needs 15-second resolution for 6-month-old CPU metrics) but trends are valuable.

**Sampling (for traces):**
Since collecting every trace for every request is prohibitively expensive at Roku scale (billions of requests/day), only a fraction of traces are collected.

- **Head sampling:** decide at the entry point of a request whether to trace it (random % chance). Simple but may miss rare error traces.
- **Tail sampling:** collect all spans, buffer them, then make the sampling decision after the trace is complete — keep all error traces, all slow traces, and a random sample of normal traces. More accurate but requires buffering (OTel Collector tail sampling processor).

**When to use which:** head sampling for baseline coverage (sample 1% of all requests); tail sampling to ensure you always capture slow or error traces regardless of the sampling rate.

---

### Q18. What is Parquet and why is it relevant for observability data pipelines?

**Answer:**

Parquet is a columnar storage format designed for analytics workloads — querying large datasets efficiently when you only need a subset of columns.

**Row-oriented vs columnar:**
```
Row store (traditional):
  row1: [timestamp=1716000000, service="api",  latency=45.2, status=200]
  row2: [timestamp=1716000001, service="auth", latency=12.1, status=200]
  row3: [timestamp=1716000002, service="api",  latency=890, status=500]

  Query "average latency for service=api" → must read all columns of all rows

Columnar (Parquet):
  timestamp column: [1716000000, 1716000001, 1716000002, ...]
  service column:   ["api", "auth", "api", ...]
  latency column:   [45.2, 12.1, 890, ...]
  status column:    [200, 200, 500, ...]

  Query "average latency for service=api" → read only service and latency columns
  → 50-90% less I/O for wide tables
```

**Why observability pipelines use Parquet:**
- Log analytics at terabyte-per-day scale: you rarely need all fields. Querying "all errors in the last hour for service X" only needs timestamp, service, and level columns.
- ClickHouse, Apache Spark, DuckDB, and AWS Athena all natively read Parquet — enabling ad-hoc analytics on stored observability data without loading it into a specialized database.
- Parquet's encoding (dictionary encoding, run-length encoding, delta encoding) combined with ZSTD compression achieves 5-20x compression for typical log/trace data.

**Roku's observability pipeline use case:** ingesting terabytes of logs/traces daily → store in Parquet on S3 → query with ClickHouse or Athena for cost analytics, retention analysis, and capacity planning without loading data into Elasticsearch.

---

### Q19. Design a high-throughput metrics ingestion pipeline capable of handling 1 million samples/second. What are the bottlenecks and how do you address them?

**Answer:**

1M samples/second is approximately what a large-scale Prometheus deployment with 1000+ scrape targets at high cardinality would generate.

**Architecture:**

```
Prometheus instances (scrapers)
        │ Remote Write (OTLP or Prometheus remote write protocol)
        ▼
Load Balancer / Kafka (ingestion buffer)
        │ Partitioned by metric name hash
        ▼
Ingest Workers (stateless, horizontally scalable)
  - Parse samples
  - Validate label names/values
  - Cardinality enforcement (reject series exceeding tenant limits)
  - Write to TSDB / forward to storage backend
        │
        ▼
Storage Backend
  - Cortex/Mimir (horizontally scalable Prometheus-compatible TSDB)
  - Or Thanos Receive (remote write endpoint for Thanos)
```

**Bottleneck 1 — Ingest bandwidth:**
- 1M samples/sec × 50 bytes per sample (wire format) = 50 MB/s raw ingestion bandwidth
- Mitigation: use snappy compression on remote write (Prometheus remote write v2 compresses)

**Bottleneck 2 — WAL write throughput:**
- All samples must be durably written to WAL before acknowledgment
- Mitigation: batch writes across multiple goroutines; async WAL writes with careful ordering; TSDB's WAL is single-writer, so at this scale you need multiple TSDB instances (Mimir/Cortex sharding)

**Bottleneck 3 — Label cardinality:**
- 1M samples/sec with high cardinality = millions of series in memory simultaneously
- Mitigation: enforce per-tenant cardinality limits at the ingest layer; reject or relabel series that exceed thresholds

**Bottleneck 4 — Memory:**
- Each active series requires in-memory state (current chunk, head series metadata)
- 1M active series × ~1 KB per series = ~1 GB memory minimum — manageable, but cardinality explosions multiply this
- Mitigation: active series limit enforcement, chunk encoding efficiency, memory-mapped chunks

**Bottleneck 5 — Object storage upload latency:**
- Thanos Sidecar uploads 2-hour blocks to S3; large blocks take minutes to upload
- Mitigation: streaming multipart upload, background goroutine for uploads that does not block ingestion

---

## Section 6 — Go for Observability Systems (Advanced)

---

### Q20. How do you implement a Prometheus metrics exporter in Go? What are the best practices?

**Answer:**

```go
package main

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics at package level (not in handlers — avoid re-registration)
var (
    // Counter — total requests by method and status
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "roku",
            Subsystem: "api",
            Name:      "requests_total",
            Help:      "Total HTTP requests by method and status code.",
        },
        []string{"method", "status", "handler"},
    )

    // Histogram — request duration distribution
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "roku",
            Subsystem: "api",
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration distribution.",
            // Buckets tuned for Roku API latency profile
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "handler"},
    )

    // Gauge — current active connections
    activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "roku",
        Subsystem: "api",
        Name:      "active_connections",
        Help:      "Number of currently active HTTP connections.",
    })
)

// Middleware that records metrics for every request
func instrumentedHandler(handler string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        activeConnections.Inc()
        defer activeConnections.Dec()

        // Wrap ResponseWriter to capture status code
        rw := &responseWriter{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(rw, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(rw.statusCode)

        httpRequestsTotal.WithLabelValues(r.Method, status, handler).Inc()
        httpRequestDuration.WithLabelValues(r.Method, handler).Observe(duration)
    })
}

func main() {
    // Expose /metrics endpoint
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

**Best practices:**
1. Use `promauto` for automatic registration — avoids forgetting to call `prometheus.MustRegister()`
2. Define metrics as package-level variables — not inside functions (causes "already registered" panics on restart or test reruns)
3. Keep label cardinality low — never use user IDs, request IDs, or URLs with path parameters as labels
4. Always set `Help` text — it appears in `/metrics` and helps teams understand what the metric measures
5. Use `MustRegister` (panics on duplicate) rather than `Register` (returns error) — you want to catch registration errors at startup, not silently at runtime
6. Custom registries for libraries: avoid polluting the default registry in shared libraries; pass a `prometheus.Registerer` interface instead

---

### Q21. How would you implement a high-performance metric ingestion pipeline in Go? What concurrency patterns would you use?

**Answer:**

```go
// Ingestion pipeline with batching and bounded concurrency

type MetricSample struct {
    Name      string
    Labels    map[string]string
    Value     float64
    Timestamp int64
}

type IngestionPipeline struct {
    input    chan MetricSample
    batches  chan []MetricSample
    storage  Storage
    wg       sync.WaitGroup
}

func NewIngestionPipeline(storage Storage, workers int) *IngestionPipeline {
    p := &IngestionPipeline{
        input:   make(chan MetricSample, 10000),  // bounded input buffer = backpressure
        batches: make(chan []MetricSample, 100),
        storage: storage,
    }

    // Stage 1: batcher — accumulate samples into batches for efficient writes
    p.wg.Add(1)
    go p.batcher()

    // Stage 2: parallel writers — write batches to storage
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go p.writer()
    }

    return p
}

func (p *IngestionPipeline) batcher() {
    defer p.wg.Done()
    batch := make([]MetricSample, 0, 1000)
    ticker := time.NewTicker(100 * time.Millisecond) // flush every 100ms max

    for {
        select {
        case sample, ok := <-p.input:
            if !ok {
                if len(batch) > 0 {
                    p.batches <- batch
                }
                close(p.batches)
                return
            }
            batch = append(batch, sample)
            if len(batch) >= 1000 {  // flush when batch is full
                p.batches <- batch
                batch = make([]MetricSample, 0, 1000)
            }

        case <-ticker.C:
            if len(batch) > 0 {
                p.batches <- batch
                batch = make([]MetricSample, 0, 1000)
            }
        }
    }
}

func (p *IngestionPipeline) writer() {
    defer p.wg.Done()
    for batch := range p.batches {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        if err := p.storage.WriteBatch(ctx, batch); err != nil {
            // structured logging + increment error counter
            log.Error("batch write failed", "size", len(batch), "err", err)
        }
        cancel()
    }
}
```

**Key concurrency patterns used:**
- Bounded channels for backpressure (input overflow → block producers → slow down upstream)
- Stage-based pipeline (batcher → writers) for separation of concerns
- `sync.WaitGroup` for clean shutdown
- Per-operation context with timeout to prevent blocking forever on storage
- Pre-allocated batch slices to reduce GC pressure (`sync.Pool` is another option for very high throughput)

---

## Section 7 — Kubernetes & Cloud-Native Observability (Intermediate–Advanced)

---

### Q22. How does Prometheus service discovery work in Kubernetes? What are the common discovery mechanisms?

**Answer:**

Prometheus uses the Kubernetes API to dynamically discover scrape targets — no static configuration needed.

**Discovery mechanisms:**

**`kubernetes_sd_configs` with role types:**

```yaml
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod         # discover all pods with scrape annotations
    relabel_configs:
      # Only scrape pods with annotation prometheus.io/scrape: "true"
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      # Use the pod's annotation for the scrape port
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: (\d+)
        replacement: ${1}
```

**Available roles:**
- `pod` — discovers all pods; uses pod annotations to configure scraping
- `service` — discovers services (scrapes the service endpoint, not individual pods)
- `endpoints` — discovers all endpoints behind services (each pod behind the service)
- `node` — discovers Kubernetes nodes (for node-level metrics via Node Exporter)
- `ingress` — discovers ingress resources (for blackbox probing)

**ServiceMonitor (Prometheus Operator):**
When using the Prometheus Operator, you define `ServiceMonitor` CRDs instead of editing Prometheus config directly:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: roku-api-monitor
spec:
  selector:
    matchLabels:
      app: roku-api
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

The Prometheus Operator watches ServiceMonitor resources and automatically regenerates Prometheus scrape configuration. This is the Kubernetes-native way to manage Prometheus scraping at scale — no manual Prometheus config editing.



In Kubernetes, application instances, pods, and IP addresses are ephemeral—they scale up and down, crash, and move across nodes constantly. Hardcoding static target IP addresses in Prometheus configuration is impossible.

Prometheus solves this through native Kubernetes Service Discovery (kubernetes_sd_configs), which continuously queries the Kubernetes API server to detect targets dynamically in real time.

How Prometheus Service Discovery Works (The Lifecycle)
┌─────────────────┐       Watch API       ┌──────────────────────┐
│  Kubernetes     │ ───────────────────>  │  Prometheus Engine   │
│  API Server     │  (Nodes, Pods, etc)   │ (kubernetes_sd_config)│
└─────────────────┘                       └──────────┬───────────┘
                                                     │
                                                     ▼
┌─────────────────┐    Scrape Metrics     ┌──────────────────────┐
│ Target Pod /    │ <───────────────────  │   Relabel Configs    │
│ Node Endpoint   │    http://.../metrics │ (Filter & Add Labels)│
└─────────────────┘                       └──────────────────────┘
RBAC & Authentication: Prometheus runs inside (or outside) the cluster using a Kubernetes ServiceAccount bound to a ClusterRole with get, list, and watch permissions for nodes, pods, services, and endpoints.

API Watching: Prometheus establishes long-lived Watch streams with the Kubernetes API server (/api/v1/...). When a pod or service is created, updated, or destroyed, Kubernetes broadcasts an event to Prometheus instantly.

Internal Metadata Extraction: Prometheus generates an initial target with hidden metadata labels starting with double underscores (e.g., __meta_kubernetes_pod_name, __meta_kubernetes_pod_annotation_prometheus_io_scrape).

Target Relabeling (relabel_configs): Prometheus runs the raw discovery metadata through relabeling rules to:

Filter out unwanted targets (e.g., only keep pods with prometheus.io/scrape: "true").

Rewrite target addresses (e.g., setting the target port to match prometheus.io/port).

Attach standard labels (e.g., setting namespace, pod_name, node_name).

Scraping: The target enters Prometheus' active target pool, and the scraper begins fetching /metrics at the configured scrape_interval.

---

### Q23. How do you set up distributed tracing with OpenTelemetry on Kubernetes? Walk through the components.

**Answer:**

```
Services (auto-instrumented with OTel SDK)
        │ OTLP gRPC :4317
        ▼
OTel Collector DaemonSet (one per node, for low-latency local collection)
  - Receives OTLP
  - Batches spans
  - Applies tail sampling (keep errors + slow traces + 1% of normal)
        │ OTLP gRPC
        ▼
OTel Collector Deployment (central, for processing and routing)
  - Applies resource detection (adds k8s.cluster.name, k8s.namespace.name)
  - Routes: traces → Tempo, metrics → Prometheus, logs → Loki
        │
        ├─ Traces → Tempo (object-storage backed)
        ├─ Metrics → Prometheus Remote Write
        └─ Logs → Loki
```

**OTel Collector configuration (tail sampling example):**

```yaml
processors:
  tail_sampling:
    decision_wait: 10s        # wait 10s before deciding (collect complete trace)
    num_traces: 50000         # buffer 50k traces in memory
    expected_new_traces_per_sec: 1000
    policies:
      - name: errors-policy
        type: status_code
        status_code: {status_codes: [ERROR]}   # always keep error traces
      - name: slow-traces-policy
        type: latency
        latency: {threshold_ms: 1000}          # keep traces > 1s
      - name: base-rate-policy
        type: probabilistic
        probabilistic: {sampling_percentage: 1} # keep 1% of everything else
```

**Auto-instrumentation in Go:**

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer(ctx context.Context) (func(), error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint("otel-collector:4317"),
        otlptracegrpc.WithInsecure(), // use TLS in production
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.AlwaysSample()), // collector handles sampling
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("roku-streaming-api"),
            attribute.String("k8s.cluster.name", os.Getenv("CLUSTER_NAME")),
        )),
    )
    otel.SetTracerProvider(tp)
    return func() { tp.Shutdown(ctx) }, nil
}
```

---

## Section 8 — System Design Scenarios (Expert)

---

### Q24. Design Roku's observability platform from scratch. It must handle terabytes of data daily and hundreds of millions of active series. Walk through your architecture decisions.

**Answer:**

*This is the most likely system design question from Wiard given his Thanos background and Roku's stated scale.*

**Requirements clarification:**
- Scale: terabytes of ingestion/day, 100M+ active series, multiple Kubernetes clusters
- Retention: metrics 13 months, logs 30-90 days, traces 7-30 days
- Query latency: dashboards <5s, alert evaluation <30s
- Cost: must be optimized — object storage for long-term, not block storage
- Multi-tenant: many teams, cost attribution needed
- Open-source preferred (CNCF ecosystem)

**Architecture:**

```
─────────────────── INGESTION LAYER ───────────────────
K8s Cluster A       K8s Cluster B       K8s Cluster C
Prometheus + Sidecar  Prometheus + Sidecar  Prometheus + Sidecar
OTel Collector DaemonSet per cluster

─────────────────── TRANSPORT ─────────────────────────
Metrics: Thanos Receive (remote write) or Sidecar upload
Logs:    Promtail → Loki Distributor
Traces:  OTel Collector → Tempo Distributor

─────────────────── STORAGE LAYER ─────────────────────
Metrics:  Thanos (blocks on S3) + Mimir for HA ingest
          - Raw: 2 weeks
          - 5m downsampled: 6 months
          - 1h downsampled: 13 months

Logs:     Loki (chunks on S3)
          - Hot: 7 days (SSD-backed index + S3 chunks)
          - Cold: 90 days (S3 only)

Traces:   Grafana Tempo (blocks on S3)
          - 30 days retention

Long-term analytics: ClickHouse
          - Log aggregations, cost analytics, usage reports

─────────────────── QUERY LAYER ───────────────────────
Thanos Query Frontend → Thanos Query → Store Gateway + Sidecars
Loki Query Frontend → Loki Querier → S3 chunks
Tempo Query → S3 blocks
Grafana (unified UI for all three)

─────────────────── DEVELOPER PLATFORM ────────────────
- Self-service onboarding (ServiceMonitor auto-provisioning)
- Cardinality explorer (which teams are using most series?)
- Cost attribution dashboard ($ per team per day)
- Alert management (Alertmanager with routing by team)
- SLO dashboards (Sloth or Pyrra for SLO generation)
```

**Key decisions:**

**Decision 1 — Thanos vs Cortex/Mimir for metrics:**
- Thanos: simpler architecture, excellent for multi-cluster global view, Sidecar model
- Mimir: purpose-built for high ingest (1M+ samples/sec), horizontally scalable ingest, but more complex
- At Roku scale (hundreds of millions of series), Mimir's distributed ingest is more appropriate than Thanos's Receive component

**Decision 2 — ClickHouse for analytics:**
- Prometheus/Thanos is optimized for alerting and operational dashboards (recent data, small time ranges)
- Long-term analytics ("which services caused the most alerts last quarter?") are better served by ClickHouse's columnar analytics — orders of magnitude cheaper than Thanos for cross-series aggregations over months

**Decision 3 — Cardinality governance:**
- At 100M+ series, cardinality explosions (one team adding `user_id` as a label) can OOM a Prometheus in minutes
- Build cardinality limits per namespace/team, enforced at ingest
- Developer-facing tooling: "your metric has 50,000 unique series — here's why and how to reduce it"

---

### Q25. Your Thanos Query is returning incorrect or incomplete results. How do you debug it?

**Answer:**

This is a realistic operational question that directly maps to Wiard's Thanos expertise.

**Step 1 — Identify the symptom precisely:**
- Missing data (gaps in graphs) — different from incorrect data (wrong values)
- Which time range? Recent data (Sidecar issue) or historical (Store Gateway issue)?
- Which clusters? One cluster missing or all?

**Step 2 — Check component health:**
```bash
# Check all StoreAPI endpoints registered with Thanos Query
curl http://thanos-query:10902/api/v1/stores

# Expected: list all Sidecars, Store Gateways, Rulers
# If a Sidecar is missing → that cluster's recent data is unavailable
```

**Step 3 — Inspect the Store Gateway:**
```bash
# Check block metadata — are blocks from all clusters visible?
kubectl logs -l app=thanos-store-gateway | grep "loaded block"

# Check S3 for recent blocks from the missing cluster
aws s3 ls s3://roku-thanos-metrics/cluster-a/ --recursive | grep meta.json | tail
```

**Step 4 — Check Thanos Sidecar:**
```bash
# Is the Sidecar able to reach Prometheus?
kubectl logs -l app=prometheus -c thanos-sidecar | grep "upload"

# Is the Sidecar uploading blocks successfully?
# Metrics to check:
thanos_objstore_bucket_operations_total{operation="upload", result="success"}
thanos_objstore_bucket_operations_total{operation="upload", result="error"}
```

**Step 5 — Deduplication issues:**
- If you see duplicate data points (values appear twice), the `replica` external label may be missing or misconfigured
- Check that all Prometheus instances have `external_labels: { replica: "0" }` / `{ replica: "1" }` set differently
- Query Thanos with deduplication disabled to see raw data: add `?dedup=false` to Thanos Query

**Step 6 — Compactor issues (gaps in historical data):**
```bash
# Check if Compactor is running and processing blocks
kubectl logs -l app=thanos-compactor | grep "compaction"

# Metrics to check:
thanos_compact_iterations_total
thanos_compact_block_cleanup_failures_total
```

---

## Section 9 — Open Source, CNCF & Cost Management (Advanced)

---

### Q26. How do you approach contributing to an open-source observability project like Thanos or Prometheus? What makes a good contribution?

**Answer:**

*This question is very likely given Wiard is a Thanos team member.*

**Starting points for meaningful contribution:**

1. **Fix issues marked "good first issue" or "help wanted"** — these are explicitly curated for new contributors and have context in the issue
2. **Documentation and examples** — often overlooked but extremely valuable; Wiard himself has given talks explaining Thanos to the community — documentation contributions align with this
3. **Write tests for untested code paths** — storage systems have complex failure modes; adding tests for edge cases is always welcome
4. **Improve error messages** — observability tools are debugged by users; better errors directly reduce support burden

**What makes a good contribution:**
- **Small, focused PRs** — a 2000-line PR changing multiple things will sit in review for weeks; a 50-line PR fixing one specific issue gets merged fast
- **Tests included** — no test, no merge in quality open-source projects
- **Design discussion first** — for significant changes, open an issue or RFC before writing code; validates the approach before investment
- **Respect the project's patterns** — read the existing code; match the style, error handling approach, and logging conventions
- **Respond to review comments promptly** — reviewers' time is limited; a PR that goes silent dies

**For Roku's observability team specifically:**
- Roku operates Thanos at scale that most contributors never reach — real production insights (e.g., "the Compactor is slow when there are >10,000 blocks because of this specific code path") are extremely valuable bug reports and PRs
- Contributing back solutions to problems Roku has already solved internally benefits both the project and Roku's reputation in the CNCF community

---

### Q27. How do you implement cost attribution for observability in a multi-team organization?

**Answer:**

Cost attribution is a developer-facing capability explicitly mentioned in the Roku job description.

**What "observability cost" means:**
- Prometheus TSDB memory: proportional to active series count
- Storage: proportional to series count × retention period × samples per second
- Query compute: proportional to series count accessed per query × query frequency

**Tracking cost by team/namespace:**
```promql
# Active series per Kubernetes namespace (proxy for team)
prometheus_tsdb_head_series

# With external label enrichment during scrape, you can tag by namespace:
sum by (namespace) (
  count by (namespace, __name__) (
    {__name__!=""}
  )
)
```

**Building the cost attribution system:**

1. **Series count per team:** cardinality explorer — query the Prometheus API for series count per job/namespace: `GET /api/v1/query?query=count by (job) ({__name__=~".+"})`

2. **Storage cost per team:** track block sizes per external label set; assign S3 storage costs proportionally

3. **Query cost per team:** instrument Thanos Query Frontend to attribute query CPU time per tenant (by `X-Scope-OrgID` header or namespace label in the query)

4. **Cost dashboards in Grafana:**
   - Top 10 teams by series count (trend over 30 days)
   - Top 10 metrics by cardinality (which metric names have the most unique label combinations?)
   - Storage cost trend per team
   - Query cost per team (CPU-seconds consumed by their dashboards/alerts)

5. **Developer-facing tooling:**
   - Self-service cardinality explorer: "show me the cardinality breakdown of my metrics"
   - Budget alerts: "your team's series count has grown 50% in the last week — here's what changed"
   - Cost projections: "if this metric growth continues, your monthly storage cost will be $X"

---

## Section 10 — Behavioural & Culture Fit

---

### Q28. Tell me about a time you significantly improved the performance of an observability system. What was the bottleneck and how did you find and fix it?

**Answer (structure using your IBM experience):**

**Situation:** the distributed correlation-ID and request-tracing layer I built across 12 Go microservices at IBM was generating so many spans that the trace storage backend was overwhelmed, and query latency for traces was exceeding 30 seconds.

**Finding the bottleneck:**
- Profiled the trace ingest pipeline — identified that every internal function call was being instrumented as a span, generating 500-2000 spans per single request
- For a service handling 10,000 requests per second, this meant 5-20 million spans per second — far beyond what the backend could handle
- Also identified that 99% of traces were successful requests with sub-10ms latency — not interesting for debugging

**The fix:**
1. **Reduced span verbosity:** instrumented only service boundary crossings, database calls, and external API calls — reduced from 500 spans to 8-12 spans per request (98% reduction)
2. **Implemented tail-based sampling:** collect all spans for 10 seconds, then: keep all error traces (100%), keep all traces >500ms (100%), keep 0.1% of normal fast traces
3. **Pre-aggregated trace statistics:** for normal fast traces that were sampled away, emit a counter metric instead (request count, error count, duration histogram) — you lose individual trace details but retain aggregate statistics cheaply

**Result:** ingest volume reduced by 95%; trace query latency dropped from 30s to <1s; operational cost of the tracing backend reduced by 80%.

---

### Q29. How do you approach building a culture of observability in an engineering organization where teams do not currently instrument their services?

**Answer:**

**The wrong approach: mandate and enforce.**
Telling teams they must add 50 Prometheus metrics to their service creates resistance and checkbox compliance — teams add metrics because they must, not because they understand the value.

**The right approach: make observability the path of least resistance.**

1. **Provide a zero-friction instrumentation library:**
   - A shared Go library that adds 80% of what teams need (HTTP request metrics, gRPC metrics, go runtime metrics, structured logging) in 3 lines of code
   - Auto-instruments `net/http` and `google.golang.org/grpc` using OTel's auto-instrumentation
   - Teams get useful dashboards immediately without writing any instrumentation code

2. **Auto-generate dashboards:**
   - When a service deploys (Kubernetes Deployment), auto-generate a Grafana dashboard for the standard metrics the shared library emits
   - Teams see value before they write a single line of instrumentation

3. **Use observability to solve a real problem for a team:**
   - Find a team with a production incident they struggled to debug
   - Show them how distributed traces would have led them to the root cause in 5 minutes instead of 2 hours
   - That team becomes an internal advocate — word of mouth beats any mandate

4. **Celebrate good instrumentation publicly:**
   - "Team X's new rate limiter service has SLO dashboards, structured logs, and trace context propagation — great example for the org"
   - Recognition drives adoption faster than requirements

5. **Cost transparency:**
   - Show teams the storage cost of their metrics: "your service has 50,000 unique series; here's the monthly cost; here's how to reduce it to 1,000 series with label consolidation"
   - Engineers respond to concrete feedback about concrete impact

---

### Q30. What excites you about working on observability infrastructure specifically, as opposed to product features?

**Answer (personalise this to your experience — use your IBM background):**

The thing that genuinely excites me about observability infrastructure is that it is a force multiplier for every other engineering team. When I built the distributed correlation-ID system at IBM across 12 microservices, I reduced mean time to debug from 4+ hours to under 20 minutes. That improvement did not just help me — it made every engineer who debugged a production issue that day faster. Infrastructure work scales in a way that product feature work typically does not.

There is also a fascinating technical depth to observability systems specifically. The problem of ingesting terabytes of metrics and logs per day, retaining them cost-efficiently across months, and still answering queries in milliseconds is a genuinely hard distributed systems problem. The solutions — TSDB compression, erasure-coded object storage, columnar formats, smart downsampling — are intellectually interesting in a way that most product features are not.

And the feedback loop is immediate and honest. If the observability platform is slow, every dashboard in the company is slow. If it drops data, teams know it. You cannot hide behind unclear requirements or subjective user feedback — the operational reality is visible immediately to the people who use the tools.

At Roku's scale — hundreds of millions of active series, terabytes per day — the problems get genuinely harder and more interesting. That is exactly the kind of challenge I want to work on.

---

## Quick Reference — What Wiard Will Probe Deepest

Based on his background and Roku's stated problems:

| Topic | Depth Expected | Key Terms to Know |
|---|---|---|
| **Thanos architecture** | Very deep | Sidecar, Store Gateway, Compactor, Query, Query Frontend, StoreAPI, deduplication |
| **Prometheus internals** | Deep | TSDB, WAL, chunks, blocks, recording rules, cardinality |
| **PromQL** | Intermediate–deep | rate, histogram_quantile, recording rules, label operations |
| **OpenTelemetry** | Intermediate | Collector, OTLP, tail sampling, auto-instrumentation |
| **Go for observability** | Deep | client_golang, custom collectors, concurrency patterns |
| **Kubernetes** | Intermediate | ServiceMonitor, Prometheus Operator, service discovery |
| **Cost management** | Intermediate | Cardinality, downsampling, attribution |
| **Loki/Tempo** | Intermediate | Object-storage backed, LogQL, distributed tracing |
| **System design** | Deep | Scale, bottlenecks, trade-offs at 100M+ series |
| **Open source** | Know the community | CNCF, contribution process, upstream projects |

---

*This guide is tailored to Wiard van Rij's specific background as a Thanos team member, KubeCon speaker, and CNCF community contributor. The depth on Thanos (Q10–Q13) and Go observability patterns (Q20–Q21) are the sections most likely to determine the outcome of this interview. Practise explaining the Thanos architecture diagram from memory — it will almost certainly come up.*



##########################


Based on the provided job description for the Senior Software Engineer, Observability position in Bengaluru, India, your interview loop will strongly prioritize high-throughput distributed systems, Go/Golang optimization, and deep knowledge of open-source observability ecosystems.Expect questions targeting Roku's massive scale—which ingests terabytes of data daily and manages hundreds of millions of active series.
🐹 Go / Golang Optimization (Deep Dive)Roku specifies that 4+ years of Go experience is mandatory, as their entire observability ecosystem is built on it. Interviewers will test your ability to write memory-safe, hyper-efficient Go code.Goroutine & Channel Internals: Explain how the Go scheduler (M:N model) works. How would you handle a memory leak caused by unbuffered channels or abandoned goroutines in a heavy ingestion pipeline?Memory & GC Tuning: Given a service processing terabytes of time-series data daily, how do you minimize garbage collection (GC) pauses? How do sync.Pool or manual memory allocations help avoid heap escapes?Concurrency Primitives: Walk through how you would safely design a thread-safe, high-throughput batching queue using the sync package or channels to aggregate incoming metric data.
📈 Metrics, Logs, & Traces at ScaleThe role explicitly mandates extending and overhauling core storage layers and query paths for open-source tools.TSDB & Columnar Layouts: How do Time-Series Databases (TSDB) handle data layouts differently than columnar formats like Apache Parquet? Under what workloads would you use ClickHouse over an inverted-index store?Pre-aggregation & Downsampling: The JD mentions implementing downsampling and pre-aggregation. How would you design a distributed pipeline to downsample 100 million active series from 1-second resolution to 1-minute resolution without losing statistical anomalies?OpenTelemetry Context Propagation: Explain how context propagation works across distributed systems using OpenTelemetry or OpenTracing. How do you propagate a trace_id seamlessly across asynchronous Kafka topics or gRPC calls?🛠️ Ecosystem & Cloud Infrastructure (Kubernetes & CNCF)You will be evaluated on your integration knowledge with CNCF-supported projects and cloud platforms.Prometheus Scraping vs. Pushing: Prometheus natively uses a pull-based architecture. How does this model scale when dealing with hundreds of thousands of ephemeral Kubernetes pods, and when does it make sense to transition to an active pushing strategy or Cortex/Thanos?Cardinality Explosions: How do you handle a production incident where a developer deploys code with a high-cardinality label (e.g., embedding a user_id inside a Prometheus metric label), threatening to crash your cluster's memory?Automated Developer Tooling: The job description mentions automating onboarding, dashboards, and alerting. How would you design a Kubernetes Operator or custom CRD to automatically provision Grafana dashboards and Alertmanager rules for a newly deployed microservice?
🧠 Behavioral & "The Roku Culture"Roku evaluates alignment with its unique cultural values: high ownership, independent thinking, keeping egos in check, and doing "more for less cost" with smaller, elite teams.Ownership: Describe a time you noticed a critical structural flaw or an extreme cost inefficiency in a core infrastructure component. How did you take ownership to fix it without a direct mandate?Pragmatic Problem Solving: Roku prides itself on a "penchant for action" and "pragmatic innovation". Tell me about a time you had to balance building a "perfect" custom solution versus leveraging an existing open-source tool to unblock your consumers quickly.Cost vs. Performance: Give an example of a project where you successfully optimized a system to drastically cut cloud infrastructure costs (AWS/GCP) while preserving or improving query latency.




##################


# Roku Observability Interview — Questions & Answers

## Priority Legend

* 🔴 **Must Know**
* 🟠 **Important**
* 🟢 **Good to Know**

---

# 1. Observability Fundamentals

## 🔴 Q1. What is Observability?

**Answer:**

Observability is the ability to understand the internal state of a system by looking at the data it produces, primarily:

* Metrics
* Logs
* Traces

Monitoring generally tells us **that something is wrong**, while observability helps us understand **why it is wrong**.

For example:

```text
Monitoring:
API latency > 2 seconds

Observability:
API latency increased
        ↓
Trace shows DB call = 1.8 sec
        ↓
Logs show DB connection pool exhausted
        ↓
Metrics show DB connections = 100/100
```

So monitoring detects the symptom, while observability helps investigate the cause.

---

## 🔴 Q2. What are the three pillars of observability?

**Answer:**

### Metrics

Numerical measurements over time.

Examples:

```text
http_requests_total
cpu_usage
request_latency
```

### Logs

Discrete events describing what happened.

Example:

```text
ERROR database connection timeout
```

### Traces

Follow a request across multiple services.

```text
API
 ↓
Service A
 ↓
Service B
 ↓
Database
```

They complement each other.

---

## 🔴 Q3. Metrics vs Logs vs Traces?

|          | Metrics       | Logs            | Traces                |
| -------- | ------------- | --------------- | --------------------- |
| Data     | Numeric       | Events          | Request journey       |
| Best for | Trends/alerts | Detailed events | Distributed debugging |
| Storage  | TSDB          | Log store       | Trace store           |
| Example  | CPU = 80%     | DB timeout      | Request A → B → DB    |

---

## 🔴 Q4. What is cardinality?

**Answer:**

Cardinality is the number of unique combinations of labels for a metric.

For example:

```text
http_requests_total{
    service="payment",
    region="us-east",
    user_id="123"
}
```

If `user_id` can have millions of values, this can create millions of time series.

High cardinality can cause:

* High memory consumption
* More storage
* Higher ingestion cost
* Slower queries

**Important interview point:**

> Labels should represent bounded dimensions, not unbounded values such as user IDs or request IDs.

---

## 🔴 Q5. What are RED metrics?

**Answer:**

RED stands for:

* **Rate** — requests per second
* **Errors** — failed requests
* **Duration** — latency

Example:

```text
Rate     → 1000 requests/sec
Errors   → 2%
Duration → p99 = 500ms
```

RED is particularly useful for monitoring request-driven services.

---

## 🟠 Q6. What are the USE metrics?

**Answer:**

USE stands for:

* **Utilization**
* **Saturation**
* **Errors**

It is commonly used for infrastructure resources.

Example:

```text
CPU utilization = 80%
CPU saturation  = high
CPU errors      = 0
```

---

# 2. Prometheus

## 🔴 Q7. What is Prometheus?

**Answer:**

Prometheus is an open-source monitoring and time-series database system.

It typically works like this:

```text
Application
     ↓
/metrics endpoint
     ↓
Prometheus
     ↓
TSDB
     ↓
PromQL
     ↓
Grafana
```

Prometheus primarily uses a **pull/scrape model**.

---

## 🔴 Q8. Explain Prometheus's data model.

**Answer:**

A Prometheus metric is identified by:

```text
metric name + labels
```

Example:

```text
http_requests_total{
    method="GET",
    status="200",
    service="payment"
}
```

Each unique label combination represents a separate **time series**.

For example:

```text
http_requests_total{method="GET"}
http_requests_total{method="POST"}
```

are two different time series.

---

## 🔴 Q9. What is a time series?

**Answer:**

A time series is a sequence of timestamped samples for one unique metric + label combination.

Example:

```text
http_requests_total{service="payment"}

10:00 → 100
10:01 → 150
10:02 → 180
```

The metric and labels identify the series; the samples represent its values over time.

---

## 🔴 Q10. What happens when you add a label?

**Answer:**

Suppose:

```text
http_requests_total
```

has one series.

Add:

```text
region
```

with three possible values:

```text
us-east
us-west
eu-west
```

You potentially get:

```text
1 × 3 = 3 series
```

With multiple labels, cardinality can grow multiplicatively.

That's why uncontrolled labels can become dangerous.

---

# 3. Prometheus Metric Types

## 🔴 Q11. What are the four Prometheus metric types?

**Answer:**

### Counter

Only increases, except when the process restarts.

Example:

```text
http_requests_total
```

### Gauge

Can increase or decrease.

Examples:

```text
temperature
memory_usage
```

### Histogram

Measures distributions and creates buckets.

Useful for latency measurements.

Example:

```text
http_request_duration_seconds_bucket
```

### Summary

Calculates quantiles on the client side.

---

## 🔴 Q12. Counter vs Gauge?

**Answer:**

Counter:

```text
Requests:
10 → 20 → 30 → 40
```

Gauge:

```text
Memory:
40 → 70 → 30 → 80
```

Use a **Counter** for events that accumulate.

Use a **Gauge** for current state.

---

## 🔴 Q13. What happens if a Counter reaches maximum integer value?

**Answer:**

In practice, Prometheus counters are normally represented using floating-point values, so you don't generally hit a normal Go `int64` maximum in the same way an application integer would.

More importantly, counters are expected to **reset when the process restarts**.

PromQL functions such as:

```text
rate()
increase()
```

are designed to handle counter resets.

---

# 4. PromQL

## 🔴 Q14. What is PromQL?

**Answer:**

PromQL is Prometheus's query language.

It allows you to:

* Select time series
* Filter labels
* Aggregate data
* Calculate rates
* Calculate quantiles
* Compare series

Example:

```promql
rate(http_requests_total[5m])
```

This calculates approximately the per-second request rate over the last five minutes.

---

## 🔴 Q15. `rate()` vs `irate()`?

**Answer:**

`rate()` calculates the average per-second increase over a range.

```promql
rate(http_requests_total[5m])
```

It is generally preferred for alerting and stable measurements.

`irate()` uses the most recent samples and is more sensitive to short-term changes.

---

## 🔴 Q16. What does `increase()` do?

**Answer:**

It calculates the total increase of a counter over a time range.

```promql
increase(http_requests_total[1h])
```

Conceptually:

> How many requests happened during the last hour?

---

## 🔴 Q17. How do you calculate request rate by service?

**Answer:**

```promql
sum by (service) (
    rate(http_requests_total[5m])
)
```

This calculates the request rate for each service.

---

## 🔴 Q18. How do you calculate p99 latency from a Histogram?

**Answer:**

A typical PromQL query is:

```promql
histogram_quantile(
    0.99,
    sum by (le) (
        rate(http_request_duration_seconds_bucket[5m])
    )
)
```

The important concept is that Histogram buckets are aggregated and `histogram_quantile()` estimates the requested percentile.

---

# 5. Prometheus TSDB

## 🔴 Q19. How does Prometheus store data?

**Answer:**

Conceptually:

```text
Incoming samples
       ↓
     WAL
       ↓
    Head
       ↓
   Chunks
       ↓
 Compaction
       ↓
   Blocks
       ↓
Persistent storage
```

The Head contains recent data.

Data is eventually compacted into persistent blocks.

---

## 🔴 Q20. What is WAL?

**Answer:**

WAL means **Write-Ahead Log**.

Before data is considered safely persisted into the normal TSDB structures, it is written to the WAL.

Purpose:

> Provide durability and allow recovery after a crash.

---

## 🔴 Q21. What is compaction?

**Answer:**

Compaction combines smaller blocks into larger blocks.

Benefits include:

* Reducing the number of files
* Improving query efficiency
* Reducing redundant data
* Improving storage organization

---

## 🔴 Q22. Why is high cardinality dangerous for Prometheus?

**Answer:**

Each unique label combination creates a separate time series.

More series means:

```text
More memory
    +
More indexes
    +
More samples
    +
More storage
    +
More query work
```

At large scale, this becomes a major architectural concern.

---

# 6. Distributed Systems

## 🔴 Q23. How would you design a metrics system handling TBs of data per day?

**Answer:**

A reasonable high-level design is:

```text
Applications
     ↓
OTel / Prometheus agents
     ↓
Load Balancer
     ↓
Ingestion layer
     ↓
Kafka / Queue
     ↓
Processing
     ↓
Sharding
     ↓
TSDB / Columnar storage
     ↓
Query layer
     ↓
API / Grafana
```

Then discuss:

* Horizontal scaling
* Sharding
* Replication
* Backpressure
* Rate limiting
* Compression
* Retention
* Sampling
* Downsampling
* Query caching

---

## 🔴 Q24. How would you handle backpressure?

**Answer:**

Suppose:

```text
Producer = 100K events/sec
Consumer = 50K events/sec
```

The queue will grow.

Potential solutions include:

* Bounded queues
* Rate limiting
* Load shedding
* Sampling
* Backpressure propagation
* Horizontal scaling
* Prioritization
* Temporary buffering

For observability, low-priority telemetry may be dropped while preserving critical signals.

---

## 🔴 Q25. What happens if the storage system becomes slower?

**Answer:**

Don't allow ingestion memory to grow indefinitely.

Use:

```text
Ingestion
   ↓
Bounded queue
   ↓
Backpressure
   ↓
Rate limiting / sampling
   ↓
Storage
```

Also monitor:

* Queue depth
* Processing latency
* Storage latency
* Error rate
* Dropped samples

---

## 🔴 Q26. How would you shard time-series data?

**Answer:**

Possible approaches include:

### By metric

```text
hash(metric)
```

### By labels

```text
hash(metric + labels)
```

### By tenant/service

```text
tenant → shard
```

The important concern is avoiding hot partitions and maintaining balanced load.

---

# 7. Sampling / Downsampling / Pre-Aggregation

## 🔴 Q27. What is sampling?

**Answer:**

Sampling means keeping only a subset of the original data.

For example:

```text
1 million traces
      ↓
10% sampling
      ↓
100,000 traces
```

This reduces:

* Storage
* Network
* Processing
* Cost

---

## 🔴 Q28. What is downsampling?

**Answer:**

Downsampling reduces the resolution of historical data.

Example:

```text
Raw:
1 sample / second

After downsampling:
1 sample / minute
```

You might store:

```text
average
min
max
p95
```

for each time window.

---

## 🔴 Q29. What is pre-aggregation?

**Answer:**

Instead of repeatedly calculating an aggregation over millions of raw series, calculate and store aggregated results ahead of time.

Example:

```text
100,000 raw series
       ↓
pre-aggregation
       ↓
service-level metric
```

This makes queries much faster.

---

# 8. OpenTelemetry

## 🔴 Q30. What is OpenTelemetry?

**Answer:**

OpenTelemetry is a vendor-neutral observability framework for generating, collecting and exporting telemetry.

It supports:

* Metrics
* Logs
* Traces

A common architecture is:

```text
Application
    ↓
OTel SDK
    ↓
OTel Collector
    ↓
Backend
```

---

## 🔴 Q31. What is the OpenTelemetry Collector?

**Answer:**

The Collector is a standalone telemetry processing component.

It can:

```text
Receive
   ↓
Process
   ↓
Transform
   ↓
Sample
   ↓
Export
```

It allows centralized telemetry processing instead of embedding all processing logic inside applications.

---

## 🔴 Q32. What are receivers, processors and exporters?

**Answer:**

### Receiver

Receives telemetry.

Examples:

```text
OTLP
Kafka
Prometheus
```

### Processor

Modifies or processes telemetry.

Examples:

```text
batch
filter
sampling
memory limiter
```

### Exporter

Sends telemetry somewhere else.

Examples:

```text
Kafka
OTLP backend
Prometheus
```

---

# 9. Distributed Tracing

## 🔴 Q33. What is a trace?

**Answer:**

A trace represents the complete journey of a request through a distributed system.

Example:

```text
Trace ID: ABC

API
 ├── Auth
 ├── Payment
 │    └── Database
 └── Notification
```

---

## 🔴 Q34. What is a Span?

**Answer:**

A Span represents one operation within a trace.

```text
Trace
 ├── API span
 ├── DB span
 ├── Redis span
 └── Kafka span
```

Each span normally has:

* Trace ID
* Span ID
* Parent Span ID
* Start time
* Duration
* Attributes
* Status

---

## 🔴 Q35. How does context propagation work?

**Answer:**

The trace context is propagated from one service to another.

Conceptually:

```text
Service A
   ↓
HTTP headers
   ↓
Service B
   ↓
HTTP headers
   ↓
Service C
```

This allows all spans to be associated with the same trace.

---

# 10. Go

## 🔴 Q36. How would you design a worker pool in Go?

**Answer:**

A typical architecture is:

```text
Producer
   ↓
Buffered channel
   ↓
Workers
 ┌─┴─┐
 W1 W2 W3
 └─┬─┘
   ↓
Results
```

Important considerations:

* Bounded queue
* Worker count
* Context cancellation
* Graceful shutdown
* Backpressure
* Error handling

---

## 🔴 Q37. Channel vs Mutex?

**Answer:**

Use a **mutex** when protecting shared state.

Use a **channel** when goroutines need to communicate or transfer ownership of work/data.

A useful principle is:

> Don't use channels just because Go has channels. Use the simplest synchronization mechanism that clearly expresses the design.

---

## 🔴 Q38. How do you gracefully shut down a Go service?

**Answer:**

Typical flow:

```text
SIGTERM
   ↓
Stop accepting new work
   ↓
Cancel context
   ↓
Workers finish current work
   ↓
Close resources
   ↓
Exit
```

Common tools include:

```go
context.Context
sync.WaitGroup
```

with appropriate timeouts.

---

## 🔴 Q39. How would you investigate high CPU in Go?

**Answer:**

Use Go's profiling tools, especially:

```text
pprof
```

Look at:

* CPU profile
* Goroutine profile
* Heap profile
* Mutex profile
* Block profile

Then identify where CPU is being consumed.

---

## 🔴 Q40. How would you investigate high memory?

**Answer:**

Look at:

* Heap profile
* Allocation rate
* Object lifetime
* Goroutine count
* Caches
* Unbounded queues
* Large buffers

Then use:

```text
go tool pprof
```

to identify allocation hotspots.

---

# 11. Kubernetes

## 🟠 Q41. How would you deploy an observability platform on Kubernetes?

**Answer:**

Separate components based on their workload:

```text
DaemonSet
   ↓
Node-level collectors

Deployment
   ↓
Ingestion services

StatefulSet
   ↓
Stateful storage

Service
   ↓
Internal communication
```

Then configure:

* Resource requests/limits
* HPA
* PDB
* Probes
* Affinity
* Anti-affinity

---

## 🟠 Q42. Deployment vs StatefulSet?

**Answer:**

A Deployment is primarily used for stateless workloads.

A StatefulSet provides stable:

* Pod identity
* Network identity
* Persistent storage association
* Ordered operations

A TSDB/storage component may need StatefulSet semantics, whereas an ingestion API can often be a Deployment.

---

# 12. ClickHouse / Columnar Storage

## 🟠 Q43. Why is columnar storage useful for observability?

**Answer:**

Suppose a table contains:

```text
timestamp
service
host
status
latency
message
```

A query might only need:

```text
timestamp
latency
```

Columnar storage can read only those columns instead of the entire row.

Advantages:

* Efficient analytical queries
* Better compression
* Reduced I/O
* Column pruning

---

## 🟠 Q44. Row-oriented vs column-oriented storage?

**Answer:**

### Row-oriented

```text
A: timestamp, host, CPU, memory
B: timestamp, host, CPU, memory
```

### Column-oriented

```text
timestamp: A, B, C...
host:      A, B, C...
CPU:       A, B, C...
memory:    A, B, C...
```

Row storage is often better for transactional access.

Column storage is often better for analytical workloads.

---

# 13. System Design

## 🔴 Q45. Design a large-scale observability platform.

**Answer:**

### Requirements

```text
TB/day
100M+ series
Low query latency
High availability
Multi-tenant
Cost efficient
```

### Architecture

```text
Applications
      ↓
Collectors
      ↓
Load Balancer
      ↓
Ingestion
      ↓
Kafka
      ↓
Processing
      ↓
 ┌────┴────┐
 ↓         ↓
Metrics   Logs/Traces
 ↓         ↓
TSDB      Columnar Store
      ↓
Query Layer
      ↓
API
      ↓
Grafana
```

### Scaling

Use horizontal sharding.

### Reliability

Use replication and failover.

### Cost

Use:

* Sampling
* Downsampling
* Compression
* Tiered storage
* Retention policies

### Query performance

Use:

* Caching
* Pre-aggregation
* Indexing
* Query optimization

---

# 14. Troubleshooting

## 🔴 Q46. Latency suddenly increases. How do you investigate?

**Answer:**

Follow the telemetry:

```text
Metrics
  ↓
Identify affected service
  ↓
Check RED metrics
  ↓
Traces
  ↓
Find slow dependency
  ↓
Logs
  ↓
Find exact error
  ↓
Infrastructure metrics
```

Example:

```text
p99 latency ↑
       ↓
Trace
       ↓
DB query slow
       ↓
DB CPU ↑
       ↓
Connection pool exhausted
```

---

## 🔴 Q47. Prometheus memory suddenly increases. What would you check?

**Answer:**

First suspect a **cardinality explosion**.

Check:

* Number of active series
* New metrics
* New labels
* Label cardinality
* Scrape targets
* Ingestion rate

A dangerous metric could be:

```text
request_count{request_id="UUID"}
```

because every request can create a new time series.

---

# 15. Architecture / Design Thinking

## 🔴 Q48. How do you design for failure?

**Answer:**

For every component, ask:

```text
What happens if it dies?
What happens if it becomes slow?
What happens if the network fails?
What happens if data is duplicated?
What happens if data is lost?
```

Then introduce appropriate mechanisms:

* Retries
* Timeouts
* Circuit breakers
* Replication
* Queues
* Idempotency
* Health checks
* Failover

---

## 🔴 Q49. What is idempotency and why is it important?

**Answer:**

An operation is idempotent if performing it multiple times has the same effect as performing it once.

For example:

```text
PUT /resource/123
```

can be designed to be idempotent.

In distributed systems, retries can cause duplicate operations, so idempotency is extremely important.

---

# 16. Questions About Your Experience

These questions are particularly important because you should connect the theoretical concepts to your actual work.

## 🔴 Q50. Tell me about a difficult distributed-systems problem you solved.

**Answer approach:**

Use your **Ceph monitor timeout** problem.

Structure the answer as:

```text
Problem
   ↓
Investigation
   ↓
Root cause
   ↓
Design options
   ↓
Implementation
   ↓
Result
```

Discuss:

* Distributed systems
* Timeout
* Failure handling
* Parallel communication
* Performance

---

## 🔴 Q51. Tell me about a production race condition you found.

**Answer approach:**

Use your **DFC / ImageArchive / MZR race condition**.

For example:

```text
Zone A
   ↓
Claims IA
   ↓
Creates DFC Action

Zone B
   ↓
Sees IA
   ↓
Waits

JIT CronJob
   ↓
Incorrectly interprets state
   ↓
Marks IA failed
```

Then explain the state-machine fix.

This demonstrates:

* Distributed coordination
* Race-condition debugging
* Kubernetes
* Asynchronous workflows
* State management

---

## 🔴 Q52. Tell me about a performance improvement.

**Answer approach:**

Use your Ceph timeout optimization.

Structure:

```text
Before:
Random monitor selection
       ↓
600 sec timeout
       ↓
VSI failure

After:
Parallel/local monitor communication
       ↓
6–9 ms connection
```

Focus on:

* How you measured the problem
* Root cause
* Options considered
* Implementation
* Trade-offs
* Final result

---

# 17. Behavioral Questions

## 🔴 Q53. Why are you interested in Observability?

**Suggested answer:**

> My background is primarily in distributed cloud systems, Go, Kubernetes and storage. What interests me about Observability is that it brings these areas together at very large scale. I'm particularly interested in the storage, ingestion and query side of observability rather than just dashboards or monitoring. The opportunity to work on high-throughput telemetry systems and solve distributed-systems and performance problems is a natural extension of what I've been doing.

---

## 🔴 Q54. Why should we hire you if you don't have deep Observability experience?

**Suggested answer:**

> While Observability is a newer domain for me, the underlying engineering problems are not. I have extensive experience building distributed cloud systems using Go and Kubernetes, working with storage systems, troubleshooting production issues and optimizing performance. I've also worked on request correlation, tracing, logging and context propagation. I'm therefore bringing strong systems-engineering fundamentals and applying them to a new domain, while actively building my depth in Prometheus, OpenTelemetry and observability systems.

---

# 18. The 20 Questions You Must Know

If you become overwhelmed during the 3–4 day preparation period, focus on these first.

1. What is observability?
2. Monitoring vs observability?
3. Metrics vs logs vs traces?
4. What is cardinality?
5. What is a Prometheus time series?
6. Explain Prometheus's data model.
7. Counter vs Gauge vs Histogram vs Summary.
8. `rate()` vs `irate()` vs `increase()`.
9. How does Prometheus TSDB work?
10. What is WAL?
11. What is compaction?
12. How would you handle 100M time series?
13. How would you design a TB/day metrics pipeline?
14. What is backpressure?
15. What is OpenTelemetry?
16. What does the OTel Collector do?
17. What are Trace/Span/Trace ID/Span ID?
18. How would you troubleshoot high latency?
19. How would you design a highly available observability platform?
20. Tell me about a difficult distributed-systems problem you solved.

---

# Final Preparation Strategy

With only **3–4 days**, don't try to become an expert in every tool.

Focus on this hierarchy:

```text
                 Roku Interview
                       │
        ┌──────────────┼──────────────┐
        ↓              ↓              ↓
  Observability   Distributed      System Design
   Concepts        Systems
        │              │
        ↓              ↓
   Prometheus        Go
        │              │
        ↓              ↓
       TSDB       Kubernetes
        │
        ↓
 OpenTelemetry
```

## Highest Priority

1. **Prometheus + PromQL**
2. **Prometheus TSDB**
3. **Observability fundamentals**
4. **Distributed systems**
5. **Large-scale system design**
6. **OpenTelemetry**
7. **Go concurrency/performance**
8. **Kubernetes**
9. **Columnar storage / ClickHouse**

The goal is **not** to memorize every tool.

The goal is to demonstrate:

> **"I understand distributed systems, storage, Go and Kubernetes deeply, and I can apply those skills to building observability infrastructure at very large scale."**

That is the strongest positioning for your existing background.



Coding


Q1. Two Sum
Q2. Top K Frequent Elements
Q3. Longest Substring Without Repeating Characters
Q4. Minimum Window Substring
Q5. Longest Substring With At Most K Distinct Characters
Q13. Merge Intervals
Q22. Search in Rotated Sorted Array
Q23. Find First and Last Position
Q25. Kth Largest Element
Q30. LRU Cache

| #       | Problem                                                  | One-line explanation                                                                                          |
| ------- | -------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **Q1**  | **Two Sum**                                              | Find two numbers in an array whose sum equals a given target and return their indices.                        |
| **Q2**  | **Top K Frequent Elements**                              | Find the `K` elements that occur most frequently in an array.                                                 |
| **Q3**  | **Longest Substring Without Repeating Characters**       | Find the longest substring containing no duplicate characters.                                                |
| **Q4**  | **Minimum Window Substring**                             | Find the smallest substring of `S` that contains all characters of `T` with the required frequencies.         |
| **Q5**  | **Longest Substring With At Most K Distinct Characters** | Find the longest substring containing no more than `K` distinct characters.                                   |
| **Q13** | **Merge Intervals**                                      | Merge all overlapping intervals and return the resulting non-overlapping intervals.                           |
| **Q22** | **Search in Rotated Sorted Array**                       | Find a target in a sorted array that has been rotated, using `O(log n)` binary search.                        |
| **Q23** | **Find First and Last Position**                         | Find the first and last index of a target value in a sorted array.                                            |
| **Q25** | **Kth Largest Element**                                  | Find the `K`th largest element in an unsorted array without necessarily sorting the entire array.             |
| **Q30** | **LRU Cache**                                            | Design a cache that supports `Get` and `Put` in `O(1)` while evicting the least recently used item when full. |



Two Sum                         → Hash Map
Top K Frequent                  → Hash Map + Heap/Bucket
Longest Substring               → Sliding Window
Minimum Window                  → Sliding Window + Frequency Map
At Most K Distinct              → Sliding Window + Frequency Map
Merge Intervals                 → Sort + Merge
Rotated Sorted Array            → Binary Search
First & Last Position           → Binary Search
Kth Largest                     → Heap / Quickselect
LRU Cache                       → Hash Map + Doubly Linked List

