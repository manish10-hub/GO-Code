# System Design Notes

---

## 1. Request Flow Through Production

```
Client → DNS → CDN → LB → API Gateway → Service → Cache → DB
```

### DNS
- Resolves URL to IP
- You can return different IPs based on region or health

### CDN
- Images, CSS, JS — CDN serves content from edge locations
- These requests never touch your origin if cache hit

**Why it matters:**
- If assets are slow, it might be CDN or DNS, not your app
- Offloading static traffic here keeps your origin capacity focused on business logic

### LB (Load Balancer)
- Acts as the public entry point
- Terminate TLS: decryption happens here, so inner traffic can be plain HTTP inside a private network
- Distributed request across backend instances based on health & policy
- Can enforce simple IP blocks, basic rate limiting or WAF rules
- Can inject a trace ID header into your observability stack

---

## 2. API Gateway

- Authentication & Authorization — validate tokens, session, API key
- Fail fast for unauthorized users, so deeper services do not waste work
- Routing & versioning — apply version-based routing, tenant-based routing or canary routing
- Map path like `/orders` to the correct internal service
- Input validation, rate limiting per user, per key
- Start a trace, attach correlation IDs

---

## 3. Service Layer

- HTTP or gRPC server receives the request
- Add logging, metrics, tracing, feature flags
- Business logic
- If needed, call other services synchronously
- Each call has its own timeouts, retries & circuit breaker

---

## 4. Cache

- For reads, service checks a distributed cache using a key schema like `user_id`
- Hot data like user profiles, product details or session data
- Cache strategies: Cache-aside — on miss, read from DB, then populate cache
- Sometimes local in-process cache for tiny, ultra-hot data

### Database (DB)
- Some common patterns, like read replicas for scaling
- Strict constraints & transactions where consistency matters
- Partial updates & eventual consistency where speed & volume matter more

---

## 5. External Systems

- Payment Gateway
- SMS & Email providers
- Wrap all outbound calls with timeout, retries, circuit breaker if a provider is consistently failing

---

## 6. Async Work

- Not everything should happen in the main request, so service publishes event like `OrderCreated`; enqueue jobs like sending emails
- Workers or consumers pull messages at their own pace
- Perform heavier or bulk operations

---

## 7. Observability Layer

- **Logs** → structured logs at each boundary, include trace ID, user, correlation ID
- **Metrics** → request rate, error rate, latency per endpoint
- **Trace ID** → trace ID created at the edge & passed through headers
  - Each internal call creates spans with timing & tags
  - When something is slow, you can see exactly which hop caused the delay

---

## 8. Scaling — Trade-offs

### Comparison Table

```
┌─────────────────────┬──────────────────────────┬───────────────────────────┐
│ Property            │ Vertical Scaling          │ Horizontal Scaling        │
├─────────────────────┼──────────────────────────┼───────────────────────────┤
│ Approach            │ Make one node stronger    │ Add more nodes            │
│ Code changes        │ Change code often         │ Usually some changes in   │
│                     │                           │ app & design              │
│ Complexity          │ Low                       │ Higher                    │
│ Cost curve          │ Grows steeply at high end │ More linear               │
│ Design              │ Database easily on,       │ Easier to design          │
│                     │ quick performance fixes   │ App servers, large scale  │
│ Failure             │ Single point of failure   │ Fault tolerant            │
└─────────────────────┴──────────────────────────┴───────────────────────────┘
```

### Vertical Scaling
- No need to add sharding, coordination or complex routing
- Strong consistency is easy
- Low latency
- Bigger instances often cost a lot more per unit capacity
- Eventually you hit a limit
- Single point of failure
- **Tip:** For stateful things like DB, scale vertically first until you are close to limits or cost feels painful

### Horizontal Scaling
- App layer — stateless services
- Scale horizontally from easily on

---

## 9. Ask These Questions (When Deciding)

- Is my problem today a single hot machine that is still within what a bigger box can fix?
- Is this component stateless or easy to make stateless?
  - Yes → scale horizontally
  - No → probably a DB, try vertical & schema or query optimization
- Do I and my team have the capacity to operate a complex distributed setup?

---

## 10. Caching

> Caching is a technique used to temporarily store copies of data in high speed storage layer.

### Benefits
- Improve performance
- Reduce load on backend systems
- Increased scalability
- Cost efficiency — by reducing load on backend systems
- Enhanced user experience

### Types of Caching

```
┌────────────────────────────────────────────────────────┐
│ Type                │ Description                      │
├────────────────────────────────────────────────────────┤
│ In-Memory Cache     │ Like RAM                         │
│ Distributed Cache   │ (across services)                │
│ Database Cache      │                                  │
│ Client-Side Cache   │ In form of cookies, localStorage │
│ CDN                 │                                  │
└────────────────────────────────────────────────────────┘
```

---

## 11. Caching Strategies

### 1. Cache Aside (Lazy Loading)

```
APP --read--> [Cache]
                 |
              miss/hit
                 |
              [DB]
```

- App checks cache first, if miss then app retrieves from DB
- **Flow:** app checks cache → if miss, then app retrieves from DB

**Cache Aside — Lazy Loading**
1. App checks cache, if it misses, then app retrieves cache → DB
2. If them the DB, then load into the cache
   - To avoid stale data, we can set a TTL for cache data
   - It is perfect for system where reads are like e-commerce website, product data updates are infrequent (price, description) but read much more frequently

### 2. Write Through

```
APP --write--> [Cache] --write--> [DB]
```

- Every write operation is executed on both (Cache & DB)
- This is synchronous process, both are updated as a single request from the cache
- Ideal for consistency — critical systems

**Write Through:**
- First written to the cache & then asynchronously written to the DB
- This strategy focuses on minimizing write latency, cache as a Primary Storage
- Reduces write latency, there is a risk of data loss

### 3. Write Around
- This strategy, where data is written directly to the DB
- Cache is only updated when the data is requested, at which point the latest during a read operation, cache aside strategy is used to load the data into the cache

### 4. Write Back (Write Behind)

```
APP --write--> [Cache]
                  |
               async
                  |
              [DB] (later)
```

- First written to the cache & then asynchronously written to the DB
- This strategy focuses on minimizing write latency, cache acts as Primary Storage
- Reduces write latency, there is a risk of data loss
- This ensures that only frequently accessed data resides inside the cache
- Write are relatively faster because they only target the DB
- This strategy is best used in write-heavy systems, where data is frequently written, but not for read
- Use TTL
- This is ideal for write-heavy scenarios, but immediate consistency with DB is not critical
- Such as logging systems & social media feeds

---

## 12. Cache Eviction Policies

```
┌──────┬──────────────────────────────────────┐
│ Policy │ Description                        │
├──────┼──────────────────────────────────────┤
│ LRU  │ Least Recently Used                  │
│ LFU  │ Least Frequently Used                │
│ FIFO │ First In First Out                   │
│ TTL  │ Time To Live                         │
└──────┴──────────────────────────────────────┘
```

---

## 13. Cache Challenges & Considerations

1. **Cache Coherence** — ensure cache data remains consistent; how to update/remove stale data
2. **Cache Invalidation** — when & how to update cache when cache is empty
3. **Cold Start** — handling scenarios where many cache conditions out
4. **Cache Eviction Policies** — managing situations where many cache conditions out
5. **Cache Stampede** — request attempt to rebuild the cache once

---

## 14. Where Should I Cache?

### Cache Layers

```
Client → Edge → Service → DB
```

#### Client
- Lives on Browser, mobile app to avoid N/W call
- Good for anything that is safe to reuse for the same user & does not change on every request
- User profile that rarely changes
- If you want to cut round trips completely (yield)
- You can tolerate slightly stale data at user level

#### Edge Caching (CDN, Reverse Proxy)
- Good for content that is shared across many users
- Like product catalog pages, public content like blog posts, marketing pages
- Use when you want to take load off your origin & improve global latency

#### Service Layer Caching
- Good for expensive computations or data that many requests inside your system reuse
- Example like pre-computed user permissions for this request city, so the cache
- Lets your app calculate temperature of every city once for any other request, which doesn't change for every second

#### Database Layer Caching
- DB query store in Cache

---

## 15. Distributed Caching

**Distributed Caching:**
- Cache data is stored across multiple nodes. This allows the cache to scale horizontally & accommodate the needs of large-scale apps

**Why?**
1. Scalability
2. Fault tolerance
3. Load balancing

---

## 16. Rate Limiting

> Rate limiting: we will be able to support multiple different implementations of the rate limiter interface

```
┌────────────────────────────────────────────────────────┐
│ Rate Limiter stores:                                   │
│ - boolean setRateLimit (long userId, string IpAddress, │
│   string serviceName, Date acquireTime)                │
│ - return true if request should be rate limited        │
└────────────────────────────────────────────────────────┘
```

**Where does the rate limiter go?**
- The amount of data that we need to store is too big to put on load balancer
- Perform rate limiting locally on each service

### Option 1 — Rate Limiting on each Service (Local)

```
       ┌──────┐
Client →  LB  →  [RL] [Ser]
                  [RL] [Ser]
```

**Pros:**
1. No extra N/W call required to perform rate limiting
2. Fewer types of component to manage

**Cons:**
1. Application & rate limiting scaling are tightly coupled
2. If we get the spam coupled can use servers N/W call (means LB to service N/W call)
   — bandwidth

### Option 2 — Dedicated Distributed Rate Limiter

```
        ┌────┐        ┌────┐ ┌────┐
Client → LB  → check → RL  │ RL  │
                      └────┘ └────┘
                      ↗
                   [Svc]
```

**Pros:**
- Shields application server from large bursts of N/W requests
- Scale up or down independently

**Rate Limiting Caching:**
- Rate limiting limits now require an extra N/W call to API request
- Since our rate limits now require an extra N/W call, it would potentially be very useful to avoid this for users who submit lots of requests
- Write back cache on the load balancer once user reaches certain threshold
- Ideal: write back cache to the load balancer

```
0 → 1 → 2 → 3 → 4 → 5
         ↓
      [LB] [Cache]
              ↓
           [RL]
         (limit is ok)
```

---

## 17. Rate Limiting — Database Choice

> Recall: we want both read & write to be as fast as possible & we don't have that much data to store → No time to access disk

**So Redis / Memcached:**
- Can take advantage of existing data structure
- Replication also fairly easy to set up

---

## 18. Rate Limiting — Replication Choice

We need our rate limiter to be fault tolerant. If it goes down, we can call any endpoint.

### 1. Multi-Leader / Leaderless
- Could increase write throughput, lower write load on each system
- Have to wait for anti-entropy for counts to synchronize
- Could use CRDT

### 2. Single Leader Replication
- Count of requests on master are accurate, can do all reads on master
- If we have enough shards should be able to have reasonable R/W throughput

```
┌─────────────────────────────────────────────────────┐
│           Single Leader Replication                 │
│                                                     │
│  [Master] ──replicate──→ [Replica 1]               │
│     │                    [Replica 2]               │
│     │                    [Replica 3]               │
│  All writes                                         │
│  All reads (accurate counts)                        │
└─────────────────────────────────────────────────────┘
```

---

## Summary Diagrams

### Full Request Flow

```
┌────────┐    ┌─────┐    ┌─────┐    ┌────┐    ┌─────────────┐    ┌─────────┐    ┌───────┐    ┌────┐
│ Client │───▶│ DNS │───▶│ CDN │───▶│ LB │───▶│ API Gateway │───▶│ Service │───▶│ Cache │───▶│ DB │
└────────┘    └─────┘    └─────┘    └────┘    └─────────────┘    └─────────┘    └───────┘    └────┘
                                                                        │
                                                                        ▼
                                                               ┌─────────────────┐
                                                               │ External Systems │
                                                               │ Payment/SMS/Email│
                                                               └─────────────────┘
                                                                        │
                                                                        ▼
                                                               ┌─────────────────┐
                                                               │   Async Queue   │
                                                               │ Worker/Consumer │
                                                               └─────────────────┘
```

### Caching Strategy Decision

```
Need to cache?
      │
      ▼
Is data shared across many users?
  YES ──→ Edge Cache (CDN)
  NO  ──→ Is it expensive to compute?
              YES ──→ Service Layer Cache
              NO  ──→ Is it a DB query?
                          YES ──→ DB Cache / Redis
                          NO  ──→ Client-Side Cache
```

### Scaling Decision

```
Performance problem?
         │
         ▼
Is component stateless?
  YES ──→ Scale horizontally (add more nodes)
  NO  ──→ Is it a DB?
              YES ──→ Scale vertically first
                      then shard when limits hit
              NO  ──→ Try to make it stateless first
```

### Rate Limiter Architecture

```
Option 1 — Local (simple, coupled)
  Client ──→ LB ──→ [Service + RL]
                    [Service + RL]
                    [Service + RL]

Option 2 — Distributed (independent scale)
  Client ──→ LB ──→ RL cluster ──→ Service cluster
                    [RL1][RL2]     [Svc1][Svc2]
                       │
                    Redis/Memcached
                    (shared counters)
```