# System Design Trade-offs — A Comprehensive Engineering Guide

> **Purpose:** This guide is built for engineering decision-making, not interview memorisation. Every trade-off is explained from the perspective of *why* a choice is made in production, what you gain, what you give up, and how real systems have navigated it.

---

## Table of Contents

1. [Vertical vs Horizontal Scaling](#1-vertical-vs-horizontal-scaling)
2. [SQL vs NoSQL](#2-sql-vs-nosql)
3. [Strong Consistency vs Eventual Consistency](#3-strong-consistency-vs-eventual-consistency)
4. [CAP Theorem](#4-cap-theorem)
5. [ACID vs BASE](#5-acid-vs-base)
6. [Synchronous vs Asynchronous Communication](#6-synchronous-vs-asynchronous-communication)
7. [REST vs gRPC vs GraphQL](#7-rest-vs-grpc-vs-graphql)
8. [Polling vs WebSockets vs Server-Sent Events](#8-polling-vs-websockets-vs-server-sent-events)
9. [Monolith vs Microservices](#9-monolith-vs-microservices)
10. [Stateful vs Stateless Services](#10-stateful-vs-stateless-services)
11. [Load Balancer Strategies](#11-load-balancer-strategies)
12. [Caching Strategies](#12-caching-strategies)
13. [Read Replicas vs Sharding](#13-read-replicas-vs-sharding)
14. [Partitioning Strategies](#14-partitioning-strategies)
15. [Message Queues vs Event Streaming](#15-message-queues-vs-event-streaming)
16. [Push vs Pull Architecture](#16-push-vs-pull-architecture)
17. [Batch Processing vs Stream Processing](#17-batch-processing-vs-stream-processing)
18. [CDN vs Edge Computing](#18-cdn-vs-edge-computing)
19. [Object Storage vs Block Storage vs File Storage](#19-object-storage-vs-block-storage-vs-file-storage)
20. [Leader-Follower vs Multi-Leader vs Leaderless Replication](#20-leader-follower-vs-multi-leader-vs-leaderless-replication)
21. [Database Indexing Trade-offs](#21-database-indexing-trade-offs)
22. [Optimistic vs Pessimistic Locking](#22-optimistic-vs-pessimistic-locking)
23. [API Rate Limiting Algorithms](#23-api-rate-limiting-algorithms)
24. [Circuit Breakers, Retries, and Timeouts](#24-circuit-breakers-retries-and-timeouts)
25. [Idempotency](#25-idempotency)
26. [Service Discovery](#26-service-discovery)
27. [Distributed Transactions — 2PC, Saga, Outbox](#27-distributed-transactions--2pc-saga-outbox)
28. [Consensus Algorithms — Raft and Paxos](#28-consensus-algorithms--raft-and-paxos)
29. [Distributed Locking](#29-distributed-locking)
30. [Cache Invalidation Strategies](#30-cache-invalidation-strategies)
31. [Event Sourcing vs CRUD](#31-event-sourcing-vs-crud)
32. [CQRS](#32-cqrs)
33. [Data Compression vs Latency](#33-data-compression-vs-latency)
34. [Security vs Performance](#34-security-vs-performance)
35. [Cost vs Reliability](#35-cost-vs-reliability)
36. [Latency vs Throughput](#36-latency-vs-throughput)
37. [Availability vs Durability](#37-availability-vs-durability)

---

## 1. Vertical vs Horizontal Scaling

### What it is

**Vertical scaling (scale up):** make the existing machine bigger — more CPU cores, more RAM, faster SSD.

**Horizontal scaling (scale out):** add more machines and distribute the load across them.

### The core trade-off

Vertical scaling is simple — your application changes nothing, you just buy a bigger server. Horizontal scaling is powerful — theoretically unlimited, but your application must be designed for it (stateless, partitioned, distributed-aware).

### When to choose each

**Choose vertical when:**
- Your workload is inherently single-threaded or difficult to parallelize (e.g., a legacy database engine)
- You are early-stage and operational simplicity matters more than scale
- The bottleneck is a stateful component (e.g., a primary database) where horizontal scaling is complex
- You are under a burst load that is temporary — upgrading instance class is faster than re-architecting

**Choose horizontal when:**
- Traffic is unpredictable and you need elastic scaling (add/remove machines based on load)
- You need geographic distribution (multiple regions)
- Your application is stateless or you have solved state distribution
- You have hit the vertical ceiling (the largest available instance is still not enough)

### Advantages and disadvantages

| | Vertical Scaling | Horizontal Scaling |
|---|---|---|
| **Simplicity** | High — no code changes | Low — requires distributed design |
| **Cost at scale** | Expensive — premium for largest instances | Cheaper — commodity hardware |
| **Ceiling** | Hard limit (largest machine available) | Practically unlimited |
| **Failure risk** | Single point of failure | Redundant — one machine fails, others continue |
| **Latency** | Low — no network hops between components | Higher — network calls between nodes |
| **Operational complexity** | Low | High — service discovery, load balancing, distributed state |

### Impact on system properties

- **Scalability:** Horizontal wins — no theoretical ceiling
- **Availability:** Horizontal wins — no single point of failure
- **Performance:** Vertical wins at small scale — no distributed overhead
- **Cost:** Vertical cheaper early; horizontal cheaper at large scale
- **Operational complexity:** Vertical much simpler

### Real-world examples

- **Instagram (early):** ran on very few large vertical instances (PostgreSQL on massive servers) until horizontal sharding became necessary at 100M+ users
- **YouTube:** moved from vertical MySQL instances to horizontally sharded Vitess to handle billions of rows
- **Amazon DynamoDB:** built from the ground up for horizontal scaling — no vertical scaling path exists; you provision capacity across partitions
- **Netflix:** horizontally scales stateless microservices on AWS; each service can independently scale to thousands of instances

### Common interview questions

- "Your database is becoming a bottleneck. How do you scale it?"
- "When would you NOT horizontally scale a service?"
  I wouldn't horizontally scale if the bottleneck isn't the service itself. 
  CPU Isn't the Bottleneck, 
  Shared External Rate Limits- Suppose your service calls:Payment Gateway limited to: 1000 requests/sec , Whether your service has: 2 instances,20 instances.200 instances, you're still limited by the downstream service.
  Some work simply cannot be parallelised.
  Strong Consistency Requirements
  If the true bottleneck is a shared dependency, coordination overhead, or sequential work, adding more instances provides little or no benefit
  I'd first use metrics and profiling to identify the actual bottleneck, then decide whether to optimise the service, redesign the architecture, or scale a different component."
- "What prevents horizontal scaling of a traditional RDBMS?"

### Typical mistakes

- Jumping to horizontal scaling too early — most teams don't need it at <1M users
- Forgetting that horizontal scaling requires solving distributed state, distributed transactions, and data consistency — the hard problems are not in adding machines, they are in the data layer
- Treating read replicas as "horizontal scaling" of a database — they help with reads but the write path remains single-node

### Decision framework

```
Is the component stateless?
  YES → horizontal scales well, start there
  NO  → vertical first, then solve state distribution explicitly

Is the bottleneck the CPU/RAM of a single process?
  YES → vertical (quick win)
  NO (it is I/O, network, lock contention) → vertical alone won't help

Do you need geographic redundancy?
  YES → horizontal is required

Can you afford to redesign for distribution?
  NO → vertical first, defer horizontal
  YES → design for horizontal from the start (saves painful migration later)
```

---

## 2. SQL vs NoSQL

### What it is

**SQL (relational databases):** data stored in tables with predefined schemas, relationships enforced by foreign keys, queried with SQL. Examples: PostgreSQL, MySQL, Oracle.

**NoSQL:** umbrella term for non-relational databases. Includes document stores (MongoDB), key-value stores (Redis, DynamoDB), wide-column stores (Cassandra, HBase), and graph databases (Neo4j).

### The core trade-off

SQL gives you structure, relationships, and strong consistency (ACID) at the cost of rigid schema and limited horizontal scalability. NoSQL gives you flexibility, horizontal scalability, and high write throughput at the cost of weaker consistency guarantees and limited join/query capability.

### When to choose SQL

- Your data has clear relationships and you need joins (e.g., users → orders → products → inventory)
- You need ACID transactions (financial systems, inventory management, booking systems)
- Your schema is known and stable — not changing every sprint
- Your team knows SQL and relational modelling — underestimated advantage
- Read-heavy workloads with complex queries — SQL query planners are decades ahead of NoSQL

### When to choose NoSQL

- **Document store (MongoDB, Firestore):** hierarchical or variable-structure data; when the "document" is the natural unit (product catalog, user profiles with optional fields)
- **Key-value (Redis, DynamoDB):** simple lookup by key with extremely low latency; sessions, caches, rate limiting counters
- **Wide-column (Cassandra, HBase):** time-series data, write-heavy workloads, data with a clear partition key; sensor data, activity feeds, audit logs
- **Graph (Neo4j):** deeply connected data where relationship traversal is the primary query pattern; social graphs, recommendation engines, fraud detection

### Advantages and disadvantages

| | SQL | NoSQL |
|---|---|---|
| **Schema** | Enforced, consistent | Flexible, per-document |
| **Relationships** | First-class (joins, FK) | Application-side |
| **Transactions** | ACID, multi-table | Limited (often single-document) |
| **Horizontal scale** | Hard (sharding is painful) | Built-in (DynamoDB, Cassandra) |
| **Query flexibility** | Very high (SQL) | Limited to access patterns |
| **Consistency** | Strong by default | Configurable (eventual by default) |
| **Operational maturity** | Decades of tooling | Younger, varies by product |

### Impact on system properties

- **Consistency:** SQL strongly wins — NoSQL often trades consistency for availability/partition tolerance
- **Scalability:** NoSQL wins for write-heavy horizontal scale
- **Flexibility:** NoSQL wins for schema evolution
- **Queryability:** SQL wins — ad-hoc queries, aggregations, analytics

### Real-world examples

- **Airbnb:** PostgreSQL for core booking/payment data (needs ACID); Elasticsearch for search; Redis for caching
- **Twitter:** MySQL for user accounts and tweets; Manhattan (internal key-value) for social graph
- **Netflix:** Cassandra for activity history and viewing records — billions of writes/day, time-series access pattern; MySQL for billing
- **Uber:** Schemaless (built on MySQL) for trips and driver data; Cassandra for geospatial data
- **Amazon:** DynamoDB for shopping cart (write-heavy, simple key-value access); Aurora (MySQL-compatible) for inventory and orders

### Common interview questions

- "Design a URL shortener — which database would you use?"
- "When would you use MongoDB over PostgreSQL?"
- "How does Cassandra achieve horizontal scalability?"

### Typical mistakes

- Using MongoDB "because it's flexible" and ending up with inconsistent, un-queryable data
- Using SQL for a time-series workload and then struggling with table bloat and slow inserts
- Treating NoSQL as automatically faster — a well-tuned PostgreSQL often outperforms MongoDB for the same workload
- Not modelling access patterns before choosing a NoSQL database — Cassandra's data model is designed around queries, not entities; getting this wrong produces the wrong schema at the start

### Decision framework

```
Do you need multi-entity ACID transactions?
  YES → SQL (PostgreSQL)

Is your primary query pattern: "look up by key"?
  YES → Key-value (DynamoDB, Redis)

Is your data hierarchical with variable structure?
  YES → Document (MongoDB, Firestore)

Is your workload write-heavy (millions of writes/sec) with time-series or partitioned access?
  YES → Wide-column (Cassandra)

Is relationship traversal your primary operation?
  YES → Graph (Neo4j)

Otherwise → SQL (PostgreSQL) is the right default
```

---

## 3. Strong Consistency vs Eventual Consistency

### What it is

**Strong consistency:** after a write completes, every subsequent read from any node returns that written value. What you wrote is immediately visible everywhere.

**Eventual consistency:** after a write completes, reads may temporarily return stale data. Given enough time with no new writes, all nodes will converge to the same value.

### The core trade-off

Strong consistency requires coordination between nodes before confirming a write — this coordination takes time (network round trips) and reduces availability (if coordination fails, the write fails). Eventual consistency removes that coordination requirement — writes are fast and always available, but readers may see stale data for some window of time.

### When to choose strong consistency

- **Financial transactions:** a bank account balance must be the same everywhere immediately. A double-spend is unacceptable even for milliseconds.
- **Inventory management:** if the last item in stock is sold, that must be visible to every concurrent shopper immediately
- **Authentication and authorization:** a user's revoked access must be effective immediately, not "eventually"
- **Leader election / distributed coordination:** the system must agree on one leader

### When to choose eventual consistency

- **Social media feeds:** if a new post takes 500ms to appear for all followers, no one cares
- **Product view counts, like counts:** approximate numbers are acceptable; exact counts are not worth the coordination cost
- **Shopping cart:** a user's cart is personal; stale data for a few seconds is acceptable
- **DNS:** propagation of DNS changes takes minutes to hours globally — classic eventual consistency
- **Search indexes:** newly created content taking seconds to appear in search results is acceptable

### Advantages and disadvantages

| | Strong Consistency | Eventual Consistency |
|---|---|---|
| **Correctness** | Guaranteed — no stale reads | Stale reads possible during propagation |
| **Latency** | Higher — coordination overhead | Lower — no coordination |
| **Availability** | Lower — coordination can fail | Higher — always accepts writes |
| **Write throughput** | Lower — serialized through coordinator | Higher — parallel writes possible |
| **Complexity** | Lower for the application | Higher — app must handle stale reads |

### Real-world examples

- **Google Spanner:** provides external consistency (strongest form of strong consistency) across globally distributed nodes using TrueTime (atomic clocks + GPS). Used for Google's AdWords financial data.
- **Amazon S3:** eventual consistency for overwrite PUTs and DELETEs for years (became strongly consistent in December 2020 after architectural changes)
- **Amazon DynamoDB:** offers both — eventually consistent reads (cheaper, 50% the cost) or strongly consistent reads (more expensive, single-leader read)
- **Cassandra:** tunable consistency — you choose how many replicas must acknowledge a write/read. `QUORUM` gives you "almost strong" consistency; `ONE` gives you eventual
- **Zookeeper/etcd:** strong consistency for distributed coordination (leader election, service discovery)

### Common interview questions

- "In your design, does the user need to see their own write immediately?"
- "How does DynamoDB handle consistency, and what is the cost difference?"
- "What is read-your-own-writes consistency, and when do you need it?"

### Typical mistakes

- Assuming eventual consistency is always faster — the consistency level affects latency, but a poorly designed eventual system can be slower than a well-tuned strongly consistent one
- Not accounting for the "convergence window" — in high-write systems, eventual consistency can mean data is stale for seconds or minutes, not milliseconds
- Conflating "eventual consistency" with "no guarantees" — eventual consistency is a specific guarantee: the system *will* converge; it just takes time

---

## 4. CAP Theorem

### What it is

The CAP theorem states that a distributed system can guarantee at most two of the following three properties simultaneously:

- **C — Consistency:** every read returns the most recent write (or an error)
- **A — Availability:** every request receives a response (not an error), though it may not be the most recent data
- **P — Partition tolerance:** the system continues operating even when network partitions occur (some nodes cannot communicate with others)

### The critical insight

**In practice, P is not optional.** Network partitions *will* happen in any distributed system — network cables fail, switches crash, cloud availability zones lose connectivity. Therefore, the real choice is always between C and A when a partition occurs:

- **CP systems:** when a partition occurs, the system refuses to answer (returns error/timeout) rather than risk returning stale data
- **AP systems:** when a partition occurs, the system returns potentially stale data rather than return an error

### Real classification of common systems

| System | Classification | Reasoning |
|---|---|---|
| PostgreSQL (single node) | CA | Not distributed — partitions not applicable |
| HBase, Zookeeper, etcd | CP | Prefer consistency; refuse writes during partition |
| Cassandra, CouchDB | AP | Prefer availability; serve stale data during partition |
| DynamoDB | AP (configurable) | Default eventual; strong reads are CP |
| Google Spanner | CP (effectively CA) | Minimises partition probability via hardware |
| MongoDB | CP (primary reads) | Primary node is consistent; unavailable if primary unreachable |

### Typical mistakes

- Treating CAP as a design-time database selection tool rather than a partition-time behaviour description
- Forgetting that PACELC (an extension of CAP) is more practically useful — it asks: what do you trade when there is *no* partition? (latency vs consistency)
- Claiming a system is "CA" in a distributed deployment — CA is only possible for non-distributed (single-node) systems

---

## 5. ACID vs BASE

### What it is

**ACID** — the consistency model of relational databases:
- **Atomicity:** a transaction either fully succeeds or fully fails — no partial state
- **Consistency:** a transaction brings the database from one valid state to another
- **Isolation:** concurrent transactions behave as if they were sequential
- **Durability:** once committed, data survives crashes

**BASE** — the consistency model of many NoSQL systems:
- **Basically Available:** the system guarantees availability (response to every request)
- **Soft state:** the state may change over time even without new input (convergence happening)
- **Eventually consistent:** the system will converge to consistent state given time

### When each matters

ACID is mandatory for: financial systems, inventory, booking (where double-booking is catastrophic), medical records.

BASE is acceptable for: social media, analytics, content delivery, recommendation systems, activity feeds — anywhere where eventual correctness is fine and a few seconds of inconsistency is invisible to the user.

### Real-world examples

- **PayPal, Stripe:** ACID (PostgreSQL, MySQL) — money cannot be in two places at once
- **Twitter timeline:** BASE (eventual consistency) — a tweet appearing 200ms late for some users is acceptable
- **Amazon shopping cart:** BASE (Dynamo paper) — Bezos famously chose availability over consistency: "show the customer their cart even if it's slightly stale; don't show them an error"

---

## 6. Synchronous vs Asynchronous Communication

### What it is

**Synchronous:** the caller sends a request and blocks until it receives a response. The two parties must both be available at the same time.

**Asynchronous:** the caller sends a message and continues without waiting. The receiver processes the message at its own pace. The two parties do not need to be available simultaneously.

### The core trade-off

Synchronous is simple to reason about — you know the result immediately, errors propagate naturally, and the flow is sequential. Asynchronous is more resilient and scalable — the caller is not blocked, the receiver can work at its own pace, and the system continues functioning if the receiver is temporarily down. The cost is complexity: you must handle eventual delivery, out-of-order processing, idempotency, and partial failures.

### When to choose synchronous

- The caller needs the result immediately to continue (e.g., login returns a token; the UI needs it to proceed)
- The operation is fast (sub-100ms) — the blocking cost is negligible
- Simplicity is paramount — early-stage systems where async complexity is not yet justified
- Error handling must be immediate — you cannot proceed if the downstream fails (e.g., payment authorisation)

### When to choose asynchronous

- The operation is slow (email sending, video transcoding, report generation, ML inference)
- The caller and receiver scale independently — high write bursts should not overwhelm the processor
- Decoupling is needed — the sender should not care if the receiver is temporarily down
- Fan-out — one event triggers multiple independent downstream actions (order placed → send email AND update inventory AND charge payment AND notify warehouse)
- The operation is non-critical-path (analytics, audit logging, notification sending)

### Real-world examples

- **Stripe payment:** synchronous for authorisation (must know if the card was approved before showing "success"); asynchronous for webhook delivery to merchants (eventual delivery is fine)
- **Uber:** synchronous for driver matching (user must wait for result); asynchronous for surge pricing recalculation, trip analytics, email receipts
- **Netflix:** asynchronous for video transcoding — when you upload a video, a message is queued; encoding workers pick it up over minutes or hours; the upload API returns immediately

### Typical mistakes

- Making everything asynchronous for the sake of decoupling — async complexity (idempotency, ordering, retry logic) is expensive; only pay for it when you need it
- Not implementing idempotency in async systems — message delivery is "at least once"; if processing is not idempotent, you get duplicate side effects
- Synchronous chains across microservices — if Service A calls B synchronously, B calls C synchronously, and C calls D synchronously, your effective latency is A+B+C+D and your availability is A×B×C×D

---

## 7. REST vs gRPC vs GraphQL

### What it is

**REST (Representational State Transfer):** HTTP-based API design style. Resources are identified by URLs. Operations map to HTTP verbs (GET, POST, PUT, DELETE). Payloads are typically JSON.

**gRPC:** Google's RPC framework. Uses HTTP/2 for transport and Protocol Buffers for serialization. Strongly typed, contract-first. Supports bidirectional streaming.

**GraphQL:** a query language for APIs developed by Meta. Clients specify exactly what data they want. Single endpoint. Reduces over-fetching and under-fetching.

### When to choose each

**REST when:**
- Public API consumed by external developers — REST is universally understood
- Simple CRUD operations — REST's resource model maps naturally
- Browser-to-server communication — REST over HTTP/1.1 works everywhere
- Cacheability is important — HTTP caching (CDN, browser cache) works naturally with GET requests

**gRPC when:**
- Internal microservice-to-microservice communication — performance and type safety matter more than universality
- Streaming is needed — bidirectional streaming is built in
- Polyglot environments — protobuf generates clients for Go, Java, Python, etc. automatically
- Low latency and high throughput — binary encoding and HTTP/2 multiplexing are significantly faster than JSON over HTTP/1.1

**GraphQL when:**
- Multiple client types (web, iOS, Android) with different data needs — each client fetches only what it needs
- Rapid frontend iteration — frontend teams can query new data without waiting for backend API changes
- Aggregating data from multiple sources — GraphQL resolver pattern handles federation well
- The over-fetching/under-fetching problem is real in your system

### Performance comparison

```
Same payload (user object with 10 fields):

JSON/REST:    ~200 bytes, HTTP/1.1
protobuf/gRPC: ~50 bytes, HTTP/2 multiplexed

Latency at 10,000 RPS (internal network):
REST JSON:  ~5ms median
gRPC proto: ~1ms median

Throughput:
gRPC is typically 5-10x higher RPS for equivalent hardware
```

### Real-world examples

- **Stripe, Twilio:** REST for public APIs — external developers know REST; no tooling required
- **Google internal:** gRPC everywhere internally — all Google services communicate via gRPC
- **Meta (Facebook):** invented GraphQL; uses it for mobile apps — crucial for reducing data transfer on mobile networks where bandwidth is limited
- **Netflix:** GraphQL federation — multiple teams own parts of the graph; the gateway aggregates them for the client

### Typical mistakes

- Using REST for internal microservice communication when gRPC would give 5-10x throughput improvement
- Using GraphQL without understanding its N+1 query problem — a single GraphQL query can trigger hundreds of database queries if resolvers are not batched (DataLoader pattern)
- Exposing gRPC directly to browser clients — browsers cannot natively speak gRPC (HTTP/2 trailers not supported); use gRPC-Web or a REST gateway

---

## 8. Polling vs WebSockets vs Server-Sent Events

### What it is

**Polling:** the client repeatedly makes HTTP requests to check for new data. Short polling checks on a timer. Long polling holds the connection open until new data arrives or a timeout occurs.

**WebSockets:** a persistent, full-duplex (bidirectional) TCP connection between client and server. Either side can send messages at any time.

**Server-Sent Events (SSE):** a persistent HTTP connection where the server can push data to the client. Unidirectional — server to client only.

### When to choose each

**Polling when:**
- Update frequency is low (once per minute or less) — the overhead of maintaining a persistent connection outweighs the polling cost
- Simplicity is paramount — polling is the easiest to implement, debug, and cache
- You cannot use persistent connections (corporate firewalls, restrictive proxies)
- The data does not need to be real-time — eventual polling is acceptable

**WebSockets when:**
- True bidirectional real-time communication — chat apps, multiplayer games, collaborative editing (Google Docs)
- High frequency updates — stock tickers, live sports scores, real-time dashboards
- The client also sends messages to the server frequently — not just receiving

**SSE when:**
- The server pushes updates but the client does not need to send messages — live feed, notifications, progress updates
- You want real-time without the complexity of WebSockets — SSE is simpler, uses plain HTTP, and auto-reconnects
- You need event streaming in a browser — SSE works through HTTP/1.1 (unlike gRPC streaming)

### Comparison

| | Polling | WebSockets | SSE |
|---|---|---|---|
| **Direction** | Client → Server (request/response) | Bidirectional | Server → Client |
| **Connection** | New per request | Persistent TCP | Persistent HTTP |
| **Latency** | High (polling interval) | Very low | Very low |
| **Overhead** | High (many HTTP requests) | Low after handshake | Low after handshake |
| **Firewall-friendly** | Yes | Sometimes blocked | Yes (plain HTTP) |
| **Auto-reconnect** | N/A | Manual | Built-in |
| **Browser support** | Universal | Universal | Universal except IE |

### Real-world examples

- **Slack:** WebSockets for real-time messaging — bidirectional (user types, server confirms, server pushes new messages)
- **GitHub Actions log tailing:** SSE — server pushes log lines as they are produced; client only receives
- **Uber driver location:** WebSockets — driver app sends location, server pushes nearby trip requests
- **Traditional banking dashboards:** polling — balance updates once per minute is acceptable; simpler than persistent connections

---

## 9. Monolith vs Microservices

### What it is

**Monolith:** a single deployable unit containing all application logic. Everything runs in the same process. The entire application is deployed together.

**Microservices:** the application is decomposed into small, independently deployable services. Each service owns its domain, has its own database, and is deployed independently.

### The core trade-off

A monolith is simpler to build, test, debug, and operate. A microservices architecture enables independent scaling, independent deployment, and team autonomy — at the cost of significant distributed systems complexity.

### When to choose monolith

- You are starting — a monolith lets you iterate fast and find the right domain boundaries before committing to service cuts
- Your team is small (<10 engineers) — the operational overhead of microservices (service discovery, distributed tracing, independent CI/CD) is not justified
- Your domain is not well understood — premature service decomposition creates the wrong boundaries, which are expensive to change later
- A modular monolith (single deployable, clearly separated modules) can give most of the organisational benefits without distributed systems complexity

**The "Modular Monolith" is often the correct middle ground:**
- Single deployment unit
- Strict module/package boundaries (no cross-module DB access, clean interfaces)
- Can be decomposed into services later when boundaries are clear

### When to choose microservices

- Teams are large (50+ engineers across 5+ teams) and independent deployment velocity is a bottleneck — team A should not be blocked by team B's release cycle
- Different components have radically different scaling requirements (the video encoding service needs GPU instances; the user profile API needs low-latency CPU instances)
- Domain boundaries are clear and stable — you know where the service cuts should be
- You have the operational maturity (Kubernetes, distributed tracing, service mesh, centralized logging) to run distributed systems

### Advantages and disadvantages

| | Monolith | Microservices |
|---|---|---|
| **Deployment complexity** | Low — single artifact | High — many artifacts, coordination |
| **Development speed (early)** | High — no network calls, simple debugging | Low — distributed testing is slow |
| **Independent scaling** | No — scale the whole app | Yes — scale each service independently |
| **Technology diversity** | Low — one language/stack | High — each service can use different tech |
| **Operational complexity** | Low | Very high — service discovery, tracing, mesh |
| **Failure isolation** | Low — one bug can crash everything | High — one service failure is contained |
| **Data consistency** | Easy — single database, ACID | Hard — distributed transactions, eventual consistency |
| **Team autonomy** | Low — shared codebase creates coupling | High — teams own their service end-to-end |

### Real-world examples

- **Amazon (circa 2002):** Jeff Bezos mandated the "two-pizza team" and service decomposition — teams must expose data only through service APIs. This is the origin of AWS (internal services productised).
- **Netflix:** moved from a DVD-era monolith to 700+ microservices. Famous for Chaos Engineering (Chaos Monkey) to test resilience.
- **Shopify:** runs a modular monolith — single Rails application, very strict module boundaries. They explicitly chose NOT to decompose into microservices despite massive scale, citing operational simplicity and fast developer onboarding.
- **Stack Overflow:** largely a monolith serving billions of requests per month — proves that microservices are not required for scale when the monolith is well-optimized.

### Typical mistakes

- Microservices-first on a greenfield project — most projects should start as a monolith
- Wrong service boundaries — splitting by technical layer (a "database service", a "validation service") rather than by domain creates chatty, tightly coupled distributed systems with all the downsides of both
- Distributed monolith — services that are "micro" in name but tightly coupled in data (sharing databases) or deployment (must deploy together)
- Underestimating operational overhead — distributed tracing, service mesh, multi-service local development environment, distributed testing

---

## 10. Stateful vs Stateless Services

### What it is

**Stateless service:** each request contains all information needed to process it. The service holds no session state between requests. Any instance can serve any request.

**Stateful service:** the service maintains state across requests. A specific instance may hold data (cache, session, connection) that subsequent requests from the same client depend on.

### The core trade-off

Stateless services are trivially horizontally scalable — add more instances, put them behind a load balancer, any instance handles any request. Stateful services are harder to scale — you must route the same client to the same instance (sticky sessions) or replicate state across instances, both of which add complexity.

### Impact on scalability and availability

- **Stateless:** trivially scalable; if an instance dies, any other instance picks up the request; no data is lost
- **Stateful:** scaling requires state migration or replication; if an instance dies, in-flight state is lost unless replicated

### Real-world examples

- **HTTP stateless by design:** REST is stateless — each request carries its own authentication (JWT, API key)
- **Kubernetes:** the control plane (kube-apiserver) is stateless — all state lives in etcd; apiserver instances can be replaced without ceremony
- **WebSocket servers:** inherently stateful — a client's connection is pinned to one server instance; horizontal scaling requires a message broker (Redis Pub/Sub) to relay messages between instances
- **Database connection pools:** stateful — PgBouncer maintains connections; scaling it is more complex than scaling a stateless service

---

## 11. Load Balancer Strategies

### The strategies

**Round robin:** requests are distributed sequentially across instances. Simple, even distribution when requests are uniform.

**Weighted round robin:** instances get traffic proportional to their weight. Useful when instances have different capacities (a 32-core machine gets 2x the weight of a 16-core machine).

**Least connections:** new requests go to the instance with the fewest active connections. Better than round robin when requests have variable processing time.

**IP hash / sticky sessions:** the same client IP always routes to the same instance. Required for stateful services.

**Random:** random instance selection. Surprisingly effective at scale — randomness averages out.

**Resource-based:** route based on CPU/memory reported by instances. Requires health metric reporting from instances.

### When to use which

| Strategy | Best for |
|---|---|
| Round robin | Stateless services, uniform request duration |
| Least connections | Long-lived connections (WebSockets, streaming) |
| IP hash | Stateful services needing session affinity |
| Weighted | Mixed-capacity instance pools |
| Resource-based | Variable-cost workloads (ML inference) |

### Real-world: Layer 4 vs Layer 7

- **Layer 4 (TCP):** fast, no HTTP parsing, operates on IP/port; used for raw performance (AWS NLB)
- **Layer 7 (HTTP):** can route based on URL path, headers, cookies; used for application-aware routing (AWS ALB, nginx, Envoy)

---

## 12. Caching Strategies

### The strategies in detail

**Cache-aside (lazy loading):**
- App checks cache first; on miss, fetches from DB, stores in cache, returns to caller
- Pro: only caches what is actually needed; cache failure doesn't break reads
- Con: first request always slow (cold miss); stale data possible if TTL is long

**Write-through:**
- Every write goes to cache AND DB synchronously before returning to caller
- Pro: cache is always fresh; no stale reads after a write
- Con: write latency doubles; cache fills with data that may never be read

**Write-behind (write-back):**
- Write to cache only; asynchronous worker flushes to DB later
- Pro: write latency is very low
- Con: data loss risk if cache crashes before flush; DB may be stale

**Read-through:**
- Cache handles misses automatically by fetching from DB (cache library, not application, manages this)
- Pro: application code is simpler
- Con: first request always slow; requires cache library support

**Refresh-ahead:**
- Cache proactively refreshes entries before they expire based on predicted access
- Pro: no cold miss for frequently accessed items
- Con: wastes resources refreshing items that won't be accessed; complex to implement

### When to use which

| Strategy | Use when |
|---|---|
| Cache-aside | Read-heavy, cache size must be controlled, simple to start |
| Write-through | Write operations are rare, consistency critical |
| Write-behind | Write-heavy, can tolerate brief DB lag |
| Refresh-ahead | Predictable access patterns, latency SLA critical |

### Real-world examples

- **Twitter:** cache-aside with Redis for tweet content and user profiles
- **Facebook TAO:** write-through cache for social graph — reads are dominant, data must be fresh after writes
- **Memcached at Facebook:** cache-aside at massive scale — 80% of read traffic served from cache

---

## 13. Read Replicas vs Sharding

### What it is

**Read replicas:** copies of the primary database that serve read traffic. Writes go only to the primary; replicas replicate asynchronously.

**Sharding (horizontal partitioning):** splitting the dataset across multiple database instances. Each shard holds a subset of rows. Reads AND writes are distributed.

### When to use each

**Read replicas when:**
- Read traffic is the bottleneck (read:write ratio is high — 10:1 or more)
- Write volume is manageable by a single primary
- Slight read staleness (replication lag) is acceptable

**Sharding when:**
- Write volume has exceeded what a single primary can handle
- Dataset is too large to fit on one machine
- You need geographic data locality (shard by region)

### The complexity cost

Read replicas are operationally simple — most databases support them natively (PostgreSQL streaming replication, MySQL replication). Sharding is operationally very complex:
- Cross-shard queries (join across shards) must be done in application code
- Rebalancing shards when you add nodes is painful without consistent hashing
- Distributed transactions across shards require 2PC or Sagas

### Real-world examples

- **Instagram:** started with read replicas; eventually needed sharding as write volume grew past single-PostgreSQL capacity
- **Pinterest:** sharded MySQL — sharded by user ID; each shard owns all data for a subset of users
- **Vitess (YouTube/Slack):** a sharding middleware for MySQL that automates many of the hard parts of sharding

---

## 14. Partitioning Strategies

### What it is

How you decide which shard/node a piece of data lives on. The choice of partition key and strategy has profound impact on data distribution and query efficiency.

### Strategies

**Range partitioning:**
- Data is partitioned by value range (user IDs 1-1M on shard 1, 1M-2M on shard 2)
- Pro: range queries are efficient — "all orders for user IDs 1-100" hits one shard
- Con: hotspots — if recent data is all in one range (time-series), one shard gets all writes

**Hash partitioning:**
- A hash function maps the key to a shard
- Pro: even distribution, no hotspots
- Con: range queries are now cross-shard (must fan out to all shards)

**Consistent hashing:**
- Nodes and keys are mapped to a ring; each key is owned by the nearest node clockwise
- Pro: when nodes are added/removed, only the data that must move does so (minimises reshuffling)
- Con: can still have hotspots without virtual nodes; complexity
- **Used by:** Amazon DynamoDB, Apache Cassandra, Riak, Memcached

**Directory-based partitioning:**
- A lookup service stores which shard each key lives on
- Pro: flexible — can move data freely by updating the directory
- Con: the directory itself becomes a bottleneck and single point of failure

### Real-world examples

- **Cassandra:** consistent hashing with virtual nodes (vnodes) — each physical node owns many virtual nodes on the ring, giving even distribution
- **DynamoDB:** consistent hashing for partition distribution; internally manages partition splitting/merging automatically
- **MySQL with Vitess:** range-based sharding by user ID — historical decision at YouTube that is now hard to change

---

## 15. Message Queues vs Event Streaming

### What it is

**Message queues (RabbitMQ, ActiveMQ, SQS):**
- Messages are produced and consumed once — after a consumer processes a message, it is removed from the queue
- Designed for task distribution — "do this job exactly once"
- Consumer pull-based; queue manages delivery

**Event streaming (Kafka, Kinesis, Pulsar):**
- Events are immutable records appended to a log — they are not deleted after consumption
- Multiple independent consumers can each read the full event stream from any point
- Events are retained for a configurable period (days, weeks, forever)
- Designed for event-driven architectures where the log is the source of truth

### When to choose each

**Message queue when:**
- Work distribution: "process this order", "send this email" — each task done by exactly one worker
- Simple producer-consumer with no need for replay
- Tasks have variable processing time and you need back-pressure
- Operational simplicity is important

**Event streaming when:**
- Multiple independent consumers need the same events (order created → email service AND inventory service AND analytics service)
- Event replay is needed — being able to rebuild a service's state from the beginning of the log
- Audit trail — the log is the record of everything that happened
- Event sourcing / CQRS architectures
- Stream processing (aggregations, joins, windowed computations on the event stream)

### Real-world examples

- **Uber:** Kafka for all trip events — 1 trillion messages per day; used for real-time analytics, surge pricing, driver matching, and billing
- **LinkedIn (Kafka's birthplace):** Kafka was built at LinkedIn to handle activity stream data — page views, likes, messages — as a centralized commit log
- **Stripe:** SQS for webhook delivery (task queue — send this webhook once); Kafka internally for event sourcing and fraud detection
- **Netflix:** Kafka for all server-side events; consumers include recommendation engines, A/B testing, billing

### Key distinction — "exactly once" vs "at least once"

Message queues typically provide at-least-once delivery. Making processing exactly-once requires idempotency on the consumer side. Kafka now has transactional APIs that enable exactly-once semantics within the Kafka ecosystem.

---

## 16. Push vs Pull Architecture

### What it is

**Push:** the sender proactively delivers data to the receiver. The receiver must be ready to accept it.

**Pull:** the receiver proactively fetches data from the sender when it is ready.

### The core trade-off

Push minimises latency — the receiver gets data the moment it is available. But if the receiver is slow or overwhelmed, push creates back-pressure problems — the sender must handle buffer management, rate limiting, and consumer failures.

Pull gives the consumer control — it fetches at its own pace, applying natural back-pressure. But it introduces polling overhead and latency (the consumer only gets data when it asks).

### When to choose each

**Push when:**
- Real-time is critical — notifications, alerts, live collaboration
- The producer knows when the consumer should receive data — event-driven triggers
- Receiver capacity is not a concern (the push rate is manageable)

**Pull when:**
- Consumers have variable processing speed — Kafka's consumer model is pull-based; each consumer fetches when it is ready
- Batch processing — consumers pull large batches for efficient processing
- Back-pressure is needed — slow consumers naturally slow down without overflowing buffers

### Real-world examples

- **Email delivery (SMTP push):** the sending server pushes email to the receiving server
- **Kafka consumers (pull):** Kafka consumers pull from the broker at their own pace; if a consumer is slow, the broker does not care — messages wait in the log
- **Webhooks (push):** Stripe pushes payment events to merchant servers; the merchant must be available
- **RSS (pull):** clients poll for new content; the server does not push

---

## 17. Batch Processing vs Stream Processing

### What it is

**Batch processing:** data is collected over a period of time and processed as a group. The result is computed on historical, static data.

**Stream processing:** data is processed as it arrives, continuously and in real-time (or near-real-time).

### When to choose each

**Batch processing when:**
- The computation requires the full dataset (sorting, global aggregations, joins across all data)
- Latency is not critical — daily reports, monthly billing, ETL pipelines
- Computational efficiency matters — batch systems amortise setup cost over large data volumes
- Data arrives in bulk — log files, database dumps, S3 objects

**Stream processing when:**
- Real-time decisions are needed — fraud detection, surge pricing, real-time recommendations
- Windowed aggregations on incoming data — "sum of transactions in the last 5 minutes"
- Data freshness is critical — dashboards, alerting, operational monitoring

### Systems

- **Batch:** Apache Spark, Hadoop MapReduce, dbt, SQL batch queries
- **Stream:** Apache Flink, Kafka Streams, Google Dataflow, Apache Storm

### Lambda Architecture

Many real systems use both:
- **Batch layer:** reprocesses all historical data periodically for accuracy
- **Speed layer:** processes recent data in real-time for freshness
- **Serving layer:** merges both views for queries

### Real-world examples

- **Netflix recommendations:** batch (nightly training of recommendation models) + stream (real-time feature updates from current viewing session)
- **Uber surge pricing:** stream (needs sub-second pricing updates based on current demand)
- **Google Analytics:** batch for historical reports; stream for real-time "active users" counter

---

## 18. CDN vs Edge Computing

### What it is

**CDN (Content Delivery Network):** a geographically distributed network of servers that cache and serve static content (images, JS, CSS, video) from locations close to end users. Examples: Cloudflare, Akamai, AWS CloudFront.

**Edge computing:** running actual computation (code execution, not just caching) at edge nodes close to users. Examples: Cloudflare Workers, AWS Lambda@Edge.

### CDN — what it gives you

- Dramatically reduced latency for static content — a user in Mumbai gets content from a Mumbai edge node, not a US origin server
- Reduced origin server load — 90%+ of static asset requests never reach your origin
- DDoS protection — attacks are absorbed at the edge before reaching your infrastructure
- TLS termination at the edge — handshakes happen close to the user

### Edge computing — what it gives you

Everything CDN gives you, plus:
- Logic execution at the edge (A/B testing, personalisation, request routing, auth validation)
- Dynamic content generation without a round-trip to origin
- Real-time content transformation (image resizing, header manipulation)

### When to use each

**CDN when:** you serve static assets, media, or content that rarely changes. This is the default for any web application.

**Edge computing when:**
- You need personalisation without a round-trip to origin (serve user A different content from user B at the edge)
- Auth/JWT validation at the edge (reject unauthenticated requests before they hit your origin)
- Geographic routing logic (route users to the nearest regional cluster)
- Low-latency applications (multiplayer games, financial trading systems)

### Real-world examples

- **Netflix:** serves video content from CDN nodes worldwide — the video bytes never leave the edge; Netflix's servers only provide the manifest
- **Cloudflare Workers:** Shopify uses Workers to run storefront logic at the edge — reducing latency for e-commerce pages from ~500ms to ~50ms globally

---

## 19. Object Storage vs Block Storage vs File Storage

### What it is

**Object storage:** stores data as objects (blob + metadata + unique key). No hierarchy — a flat namespace with a key. Examples: AWS S3, Google Cloud Storage.

**Block storage:** raw block-level storage attached to a compute instance, presented as a disk. Examples: AWS EBS, Google Persistent Disk.

**File storage (NAS/shared file system):** a hierarchical file system mounted by multiple clients simultaneously. Examples: AWS EFS, NFS, CIFS.

### When to use each

**Object storage when:**
- Storing large, unstructured data — images, videos, backups, logs, ML training data
- Data is accessed by multiple services — S3 is accessible from anywhere with a URL
- Durability over performance — S3 provides 11 9s durability
- Cost efficiency at scale — object storage is 10-100x cheaper than block storage per GB

**Block storage when:**
- A database is running on the instance — databases need raw block-level access for I/O performance
- Low-latency random I/O is needed — SSDs attached as EBS volumes
- The OS disk of a compute instance — every cloud VM has an EBS volume as its boot disk

**File storage when:**
- Multiple instances need to share the same filesystem simultaneously (e.g., a web farm where all nodes need access to the same uploaded files)
- Legacy applications that use POSIX file system APIs (open, read, write, seek)
- Shared configuration or code volumes across a Kubernetes cluster

---

## 20. Leader-Follower vs Multi-Leader vs Leaderless Replication

### What it is

**Leader-follower (single-leader, primary-replica):** one node accepts all writes (the leader). Changes are replicated to follower nodes. Followers can serve reads.

**Multi-leader (multi-master):** multiple nodes can accept writes independently. Writes must be reconciled across leaders (conflict resolution).

**Leaderless (Dynamo-style):** no designated leader. Writes go to multiple nodes simultaneously. Reads also contact multiple nodes. Quorum determines validity.

### When to choose each

**Leader-follower when:**
- Strong consistency on reads is needed (read from leader)
- Conflict-free write ordering is needed (single writer)
- Simple operations — most SQL databases use this model
- Read scaling via replicas is the goal

**Multi-leader when:**
- Geographic distribution — each region has a leader that handles local writes with low latency; inter-region replication happens asynchronously
- Offline clients that sync — CouchDB, collaborative apps like Google Docs (conceptually)
- High write availability when cross-region write latency is unacceptable

**Leaderless when:**
- Maximum write availability — writes succeed even if some nodes are down (as long as W nodes are reachable)
- Tunable consistency — you choose the trade-off per request
- No single point of failure for writes
- Examples: Cassandra, DynamoDB (original design), Riak

### The conflict problem in multi-leader

When two leaders accept writes to the same key concurrently, they have conflicting values. Resolution strategies:
- **Last write wins (LWW):** the write with the highest timestamp wins — simple but loses data
- **Merge:** the application merges both values (works for CRDTs, sets)
- **Custom resolution logic:** application decides which value wins based on business rules

### Real-world examples

- **PostgreSQL:** leader-follower; streaming replication to standby
- **CockroachDB:** multi-leader within a region with Raft; geo-distributed with inter-region replication
- **Cassandra:** leaderless with tunable quorum; W + R > N for strong consistency
- **DynamoDB:** internally managed leaderless with eventual consistency by default

---

## 21. Database Indexing Trade-offs

### What it is

An index is a data structure (typically a B-tree or hash table) that allows the database to find rows without scanning the entire table.

### The core trade-off

Every index speeds up reads at the cost of slower writes and more storage. When a row is inserted or updated, every index on that table must also be updated. A table with 10 indexes has slower INSERT/UPDATE/DELETE than a table with 1 index.

### Types and when to use each

**B-tree index (default in PostgreSQL, MySQL):**
- Supports equality, range, sorting, prefix matching
- Use on columns used in WHERE, ORDER BY, JOIN conditions
- Most general-purpose index type

**Hash index:**
- Only equality checks — very fast for exact-match lookups
- Does not support range queries or sorting
- Use for high-cardinality equality-only columns (UUID lookups)

**Composite index:**
- Index on multiple columns together
- Must match the "left prefix" rule — index on (a, b, c) helps queries on a; on (a, b); on (a, b, c); but NOT on (b) alone
- Column ordering in the index matters enormously

**Partial index:**
- Index only a subset of rows matching a condition
- E.g., index only active orders — dramatically smaller index if most orders are completed
- PostgreSQL supports this; underused in practice

**Covering index:**
- An index that contains all columns needed by a query — the query is served entirely from the index without touching the table
- Eliminates the "table heap fetch" step — significant performance gain for read-heavy queries

### Common mistakes

- Over-indexing — every column indexed "just in case" — write performance collapses; index maintenance time dominates
- Under-indexing — full table scans on million-row tables; query timeouts
- Wrong column order in composite indexes — index on (status, created_at) helps "WHERE status = 'active' ORDER BY created_at" but not "WHERE created_at > '2024-01-01'"
- Not using EXPLAIN ANALYZE to verify that indexes are actually being used

---

## 22. Optimistic vs Pessimistic Locking

### What it is

Both are strategies for preventing lost updates when multiple processes modify the same data concurrently.

**Pessimistic locking:** lock the row before reading it. Other processes block until the lock is released. Assumes conflicts are likely.

**Optimistic locking:** read the row without locking. Track the version. When writing, verify the version has not changed. If it has (conflict detected), retry. Assumes conflicts are rare.

### When to use each

**Pessimistic when:**
- Conflicts are frequent — multiple users are likely editing the same record simultaneously
- The cost of retry (on conflict) is high — long, expensive operations
- The locked resource is held for a short time — lock contention is brief

**Optimistic when:**
- Conflicts are rare — most writes will succeed without conflict
- High read:write ratio — you do not want reads to block
- Distributed systems where holding a database lock across a network call is dangerous (lock-holder crashes while holding the lock)
- Long-running operations — holding a pessimistic lock for seconds is unacceptable

### Implementation

```sql
-- Pessimistic (SELECT FOR UPDATE in PostgreSQL)
BEGIN;
SELECT * FROM orders WHERE id = 123 FOR UPDATE; -- other transactions block here
UPDATE orders SET status = 'processing' WHERE id = 123;
COMMIT;

-- Optimistic (version column)
-- Read:
SELECT id, amount, version FROM orders WHERE id = 123;
-- returns version = 7

-- Update (only succeeds if version has not changed):
UPDATE orders SET amount = 500, version = 8
WHERE id = 123 AND version = 7;
-- if 0 rows affected → conflict → retry
```

### Real-world examples

- **E-commerce inventory:** pessimistic locking for the final "decrement stock" step when multiple users check out simultaneously — conflicts are frequent at this moment, and a lost update (overselling) is catastrophic
- **Collaborative document editing:** optimistic locking with operational transforms (OT) — conflicts are expected but merges handle them
- **Database ORMs (ActiveRecord, Hibernate):** optimistic locking via version column is a built-in feature

---

## 23. API Rate Limiting Algorithms

### Why rate limiting matters

Without rate limiting: one misbehaving client (intentional or accidental) can consume all available resources, degrading or crashing the service for all other clients.

### The algorithms

**Fixed window counter:**
- Count requests per time window (100 requests per minute)
- Reset counter at window boundary
- **Problem:** a burst at the boundary (50 requests at 00:59, 50 at 01:00) allows 100 requests in a 2-second span
- Simple, low memory, but exploitable at boundaries

**Sliding window log:**
- Store the timestamp of every request in a sorted set
- Count requests in the last N seconds from the current time
- Accurate — no boundary problem
- **Problem:** memory proportional to request count per user

**Sliding window counter:**
- Approximate sliding window using two fixed windows weighted by overlap
- Good balance of accuracy and memory efficiency

**Token bucket:**
- Each user has a bucket of tokens; tokens are added at a fixed rate; each request consumes a token
- Allows bursts up to bucket capacity (good for APIs where occasional bursts are acceptable)
- **Used by:** AWS API Gateway, Stripe

**Leaky bucket:**
- Requests enter a queue; they are processed at a fixed rate regardless of input rate
- Smooths bursts — no spike can exceed the processing rate
- **Used for:** traffic shaping, network QoS

### Real-world examples

- **Stripe:** token bucket — allows short bursts (useful for checkout spikes) with sustained rate limits
- **GitHub API:** fixed window with a 5000 requests/hour limit for authenticated clients
- **Twitter API:** sliding window — 300 requests per 15 minutes

---

## 24. Circuit Breakers, Retries, and Timeouts

### Why all three are needed

These three patterns work together to prevent cascading failures in distributed systems. They are not optional at production scale.

### Timeouts

**Always set explicit timeouts on every outbound call.** Without a timeout, a slow downstream service holds your goroutine/thread indefinitely, exhausting your connection pool and effectively taking down your service.

- Connect timeout: how long to wait for TCP connection establishment
- Request timeout: how long to wait for the full response
- Rule of thumb: set timeout to the 99th percentile latency of the dependency + 50% headroom

### Retries

Retry transient failures (network blips, 503s, 429s with Retry-After). Do NOT retry:
- 4xx client errors (400, 401, 403, 404) — retrying won't fix these
- Non-idempotent operations without idempotency keys — retrying a payment without idempotency causes double charges
- When the downstream is clearly overloaded — retrying without backoff amplifies the problem

**Exponential backoff with jitter:** each retry waits longer (1s, 2s, 4s, 8s), with random jitter added to prevent thundering herd (all retrying clients hitting the downstream simultaneously after the same wait).

### Circuit breaker

After N consecutive failures, "open" the circuit — all subsequent calls fail immediately without contacting the downstream. After a cooldown, allow one test request ("half-open"). If it succeeds, close the circuit.

**States:**
```
CLOSED → normal, requests flow through
OPEN   → failing fast, no requests sent downstream
HALF-OPEN → test request allowed; success → CLOSED, failure → OPEN
```

**Why it matters:** without a circuit breaker, a failing downstream service causes all callers to wait for timeout before failing. With N services calling it, the entire system stalls waiting for timeouts. The circuit breaker makes failures fast and visible immediately.

### Real-world examples

- **Netflix Hystrix:** the canonical circuit breaker library — wraps every microservice call; falls back to cached data or default response when the circuit is open
- **Envoy/Istio:** implements circuit breaking and retry at the service mesh level, removing the need for every service to implement it in application code

---

## 25. Idempotency

### What it is

An operation is idempotent if performing it multiple times produces the same result as performing it once.

### Why it is critical

In distributed systems, you cannot know if a request succeeded. A timeout means "I did not get a response" — not "the operation failed." The server may have processed the request and the response was lost in the network. Retrying a non-idempotent operation (like charging a credit card) produces duplicate side effects.

### Implementing idempotency

**Idempotency key:** the client generates a unique key per operation and sends it with the request. The server stores the key and the result. If the same key is seen again, the server returns the stored result without re-executing.

```
POST /payments
Idempotency-Key: order-123-payment-attempt-1
Body: { amount: 9999, currency: "INR", card: "..." }

First call:  → process payment, store result with key, return 200
Second call: → look up key, find result, return 200 (no duplicate charge)
```

**At the database level:** use unique constraints on business keys (not auto-increment IDs). `INSERT INTO payments (order_id, ...) ON CONFLICT (order_id) DO NOTHING` is idempotent.

### Real-world examples

- **Stripe:** every API endpoint accepts an `Idempotency-Key` header; payment intent creation is safe to retry
- **AWS APIs:** most AWS operations are designed to be idempotent (create-if-not-exists semantics)

---

## 26. Service Discovery

### What it is

In a dynamic microservices environment, service instances come and go (due to scaling, crashes, deployments). Service discovery is the mechanism by which services find each other's current network addresses.

### Two models

**Client-side discovery:**
- Each service client queries a registry (Consul, etcd, Eureka) to get the list of available instances
- The client implements load balancing — chooses which instance to call
- Pro: client has full control; no extra network hop
- Con: every client must implement discovery and load balancing logic; coupling to registry

**Server-side discovery (service mesh / load balancer):**
- The client calls a fixed address (a load balancer or service mesh sidecar)
- The load balancer/sidecar performs discovery and routing transparently
- Pro: clients are simple — they call a fixed DNS name; discovery is infrastructure
- Con: extra network hop; the load balancer is a potential bottleneck

### Real-world examples

- **Kubernetes DNS:** server-side; services get a stable DNS name (`my-service.namespace.svc.cluster.local`); Kubernetes handles pod discovery behind the service IP
- **Consul (HashiCorp):** client-side; services register themselves; clients query Consul for healthy instances
- **Istio/Envoy service mesh:** server-side via sidecar proxy; all traffic goes through the sidecar, which knows the current instance list

---

## 27. Distributed Transactions — 2PC, Saga, Outbox

### The problem

A distributed transaction must atomically update data across multiple services/databases. If service A succeeds but service B fails, you have partial state — a fundamental challenge.

### Two-Phase Commit (2PC)

**Phase 1 (prepare):** coordinator asks all participants "can you commit?" Each participant locks its resources and replies yes/no.

**Phase 2 (commit/abort):** if all said yes, coordinator sends commit. If any said no, coordinator sends abort.

**Problems with 2PC:**
- **Blocking:** if the coordinator crashes after phase 1, participants are locked indefinitely waiting
- **Latency:** two network round trips across all participants before completion
- **Availability:** any participant unavailability blocks the entire transaction

**When to use 2PC:** within a single database system (PostgreSQL's internal 2PC) or within a single organisation's tightly coupled systems where availability cost is acceptable. Rarely appropriate across independently operated microservices.

### Saga Pattern

Break the distributed transaction into a sequence of local transactions. Each step has a compensating transaction to undo its effect if a later step fails.

**Choreography Saga:** each service publishes events; downstream services listen and react. No central coordinator.
- Pro: loosely coupled, simple
- Con: hard to track overall state; event ordering can cause issues

**Orchestration Saga:** a central orchestrator sends commands to each service in sequence. If a step fails, it sends compensating commands in reverse.
- Pro: clear visibility of the transaction's overall state
- Con: orchestrator is a centralised component

### Outbox Pattern

When a service must atomically save data to its database AND publish a message/event, the Outbox Pattern solves the dual-write problem.

Instead of writing to DB and then writing to Kafka (which can fail between the two):
1. Write the DB row AND an "outbox" row in the same local ACID transaction
2. A separate process reads the outbox table and publishes events to Kafka
3. On successful publish, mark the outbox row as processed

This guarantees: either both the DB write and the event happen, or neither does.

### Real-world examples

- **Uber:** Saga pattern for trip lifecycle (create trip → reserve driver → charge payment → dispatch)
- **Amazon:** Saga for order fulfillment — each step (payment, warehouse allocation, shipping) has a compensating transaction (refund, re-stock, cancel shipment)
- **Stripe:** Outbox pattern to guarantee webhook delivery even if Kafka is temporarily unavailable

---

## 28. Consensus Algorithms — Raft and Paxos

### What it is

Consensus algorithms allow a cluster of nodes to agree on a single value (leader, log entry, configuration) even when some nodes are unavailable or slow — the foundation of all strongly consistent distributed systems.

### Paxos

The original consensus algorithm (Lamport, 1989). Provably correct but notoriously difficult to understand and implement. The full protocol (multi-Paxos, leader election, membership changes) has many subtle edge cases.

**Used by:** Google Chubby, Google Spanner

### Raft

Designed specifically to be understandable (Ongaro & Ousterhout, 2014). Same guarantees as Paxos but with clearer separation of concerns (leader election, log replication, safety). The dominant choice for new systems.

**Key properties:**
- One leader at a time handles all writes
- Leader replicates log entries to a majority (quorum) of followers
- A leader is elected when the current leader is unavailable (detected via heartbeat timeout)
- A node only wins election if its log is at least as up-to-date as a majority of nodes

**Used by:** etcd (Kubernetes' backing store), CockroachDB, Consul, TiKV

### Quorum

Both algorithms require a **majority quorum** to make progress. In a 5-node cluster, 3 nodes must agree. This means:
- Can tolerate 2 node failures out of 5
- Cannot proceed if fewer than 3 nodes are reachable (chooses consistency over availability — CP)

---

## 29. Distributed Locking

### What it is

Preventing multiple nodes from executing a critical section simultaneously when there is no shared database lock available.

### Approaches

**Redis-based lock (Redlock):**
- SET key value NX PX timeout — atomic set-if-not-exists with expiry
- The holder has the lock for the TTL duration
- Must renew before expiry if the operation takes longer

**Problems with Redis locking:**
- If the lock holder pauses (GC pause, network hiccup) past the TTL, another holder acquires the lock while the first still thinks it holds it — two holders simultaneously
- Redis replication is asynchronous — a lock acquired on the primary may not be replicated before the primary crashes; the secondary gives the lock to another client

**ZooKeeper/etcd-based lock:**
- Uses strongly consistent consensus (Raft/ZAB) — no split-brain
- A node creates an ephemeral node; if the holder crashes, ZooKeeper automatically deletes it
- More operationally complex; higher latency than Redis

### Fencing tokens

To handle the "lock holder pauses" problem, use a fencing token:
- Each lock acquisition gets a monotonically increasing token
- The downstream resource rejects requests with older tokens

```
Client A acquires lock, gets token 33
Client A pauses (GC), lock expires
Client B acquires lock, gets token 34
Client B writes with token 34 → accepted
Client A resumes, tries to write with token 33 → rejected (34 > 33)
```

---

## 30. Cache Invalidation Strategies

### The problem

Cached data becomes stale when the underlying source changes. Invalidating cache too aggressively wastes the cache's performance benefits. Not invalidating enough means serving stale data.

### Strategies

**TTL (Time-to-Live):** entries expire after a fixed duration. Simple, no coordination needed. The cache is eventually consistent — stale for at most TTL seconds.

**Event-driven invalidation:** when source data changes, publish an event that invalidates the relevant cache entries. Fresher than TTL but requires event infrastructure and coordination.

**Write-through invalidation:** on every write to the source, simultaneously invalidate or update the cache. Zero stale window. Requires cache to be tightly coupled to write path.

**Cache stampede prevention:** when a popular key expires, many concurrent requests miss and all try to populate the cache simultaneously — overloading the database.
- Solution: probabilistic early expiry (refresh before TTL, with probability increasing as TTL approaches)
- Solution: mutex/lock — only one request fetches from DB; others wait

### Real-world examples

- **Facebook Memcached:** uses lease tokens to prevent stampedes — on a cache miss, the requesting client gets a lease; other clients wait rather than all fetching simultaneously
- **Redis TTL:** nearly universal — every entry has a TTL; staleness is bounded by the TTL

---

## 31. Event Sourcing vs CRUD

### What it is

**CRUD:** create, read, update, delete. The current state of an entity is stored and updated in place. History is lost unless explicitly audited.

**Event Sourcing:** instead of storing current state, store the sequence of events that led to that state. Current state is derived by replaying events. History is inherent.

### When to choose event sourcing

- **Audit trail is a first-class requirement:** financial systems, medical records, compliance systems — you need to know not just what the state is but how it got there
- **Time travel queries:** "what was the state of this account on January 1st?" is trivially answered by replaying events up to that date
- **Event-driven downstream consumers:** other services can build their own views by consuming the event stream
- **Undo/redo functionality:** since every change is an event, undo is just not applying the last event

### The cost of event sourcing

- **Query complexity:** "find all orders with status=active" requires either replaying all events or maintaining a read model (which leads naturally to CQRS)
- **Event schema evolution:** changing the shape of historical events while keeping them replayable is hard
- **Storage growth:** the event log grows indefinitely (snapshots can help — periodically storing the current state to avoid replaying the full history)
- **Operational complexity:** significant; event sourcing is not a default choice

### Real-world examples

- **Banking ledgers:** the canonical event sourcing system — every transaction is an event; the balance is the sum of all transactions
- **Version control (Git):** commits are events; HEAD is current state derived from replaying all commits
- **LinkedIn Feed:** event sourced — every action (like, share, comment) is an event; the feed is derived from the event stream
- **Kafka as event store:** many systems use Kafka's log as a lightweight event store — durable, ordered, replayable

---

## 32. CQRS

### What it is

**Command Query Responsibility Segregation:** separate the write model (commands that change state) from the read model (queries that return data).

### Why it matters

A single data model optimised for writes (normalised, ACID-safe) is often poorly suited for reads (denormalised, aggregated, joined). CQRS lets each side evolve independently.

### The pattern

```
Write path:
  Command → Validation → Aggregate → Event published → Write DB updated

Read path:
  Event consumed → Read model updated (denormalised, optimised for queries)
  Query → Read model → Fast response
```

### When to use CQRS

- High read:write ratio where read performance is critical — the read model can be a highly denormalised table or even a search index
- Complex read queries that are expensive to compute from the normalised write model
- Event sourcing — CQRS and event sourcing are natural companions (the event stream feeds both the write model's aggregate and the read model's projection)
- Scale — write and read services can scale independently

### Common mistake

Using CQRS as the default pattern for every service. It is genuinely complex — two models to maintain, eventual consistency between them, more moving parts. Use it when the query/write mismatch problem is real and painful, not speculatively.

---

## 33. Data Compression vs Latency

### The trade-off

Compression reduces data size (network bandwidth, storage cost) at the cost of CPU time to compress/decompress. The break-even depends on whether your bottleneck is bandwidth or CPU.

### When compression wins

- **High latency, low bandwidth networks:** mobile networks, cross-ocean links — compression dramatically reduces transfer time even accounting for CPU cost
- **Storage cost is significant:** storing 10PB of compressed logs vs uncompressed is a meaningful cost difference
- **CPU is cheap, bandwidth is expensive:** true in many cloud environments where egress bandwidth costs money

### When compression loses

- **Low latency, high bandwidth networks:** internal data center links — data transfers in microseconds; compression CPU cost exceeds transfer time savings
- **Already-compressed data:** JPEG, MP4, ZIP — compressing again wastes CPU with minimal size reduction
- **Latency-critical hot paths:** adding 1ms of compression CPU time to a 2ms internal RPC is unacceptable

### Algorithms

| Algorithm | Speed | Compression ratio | Use case |
|---|---|---|---|
| gzip | Medium | Good | HTTP responses, log archival |
| zstd | Fast | Very good | Kafka messages, storage |
| lz4 | Very fast | Moderate | Real-time, internal data |
| snappy | Very fast | Moderate | Kafka (default), in-memory |
| brotli | Slow | Excellent | Static web assets (CDN) |

---

## 34. Security vs Performance

### The trade-off

Security controls (encryption, authentication, authorisation, input validation) consume CPU, add latency, and increase operational complexity. Removing them improves performance but creates risk.

### Key examples

**TLS encryption:**
- TLS handshake: ~1ms overhead on first connection; ~0ms with session resumption (TLS 1.3)
- Encryption/decryption of payload: negligible on modern hardware with AES-NI instructions
- **Decision:** always use TLS — the performance cost is negligible; the risk of not using it is not

**JWT validation:**
- Validating a JWT signature on every request: ~0.1ms
- Can be cached (rate-limit the signature validation for the same token)
- **Decision:** always validate; cache the result for the TTL of the token

**Input validation and sanitisation:**
- Cost: trivial
- Skipping it: SQL injection, XSS, remote code execution
- **Decision:** never skip this

**Authorisation checks:**
- Centralised policy (OPA sidecar): ~1ms per request
- **Decision:** always enforce; the latency is acceptable; the alternative is privilege escalation

### Where the real trade-off exists

- **mTLS in service mesh:** encrypts and authenticates every internal service call — adds CPU and latency overhead that is measurable at millions of RPS. Trade-off is real; many systems use mTLS selectively (external-facing services) rather than universally.
- **Field-level encryption:** encrypting individual database columns — adds significant read overhead (every query must decrypt). Use only for the most sensitive fields (SSN, card numbers).

---

## 35. Cost vs Reliability

### The trade-off

Reliability costs money: redundant hardware, cross-region replication, hot standby databases, over-provisioned capacity buffers. Every additional 9 of availability roughly 10x the infrastructure cost.

### The SLA/SLO framework

Define your requirements before choosing architecture:
- 99% availability = 87 hours downtime/year — acceptable for internal tools
- 99.9% availability = 8.7 hours/year — typical for B2C products
- 99.99% availability = 52 minutes/year — financial services, healthcare
- 99.999% availability = 5 minutes/year — critical infrastructure; extremely expensive

### Where to invest in reliability

Not every component needs the same reliability level. Map your system's critical path:
- Payment processing: 99.99% (every minute of downtime is lost revenue)
- Recommendation engine: 99% (degraded experience, not lost transactions)
- Admin dashboard: 95% (internal tool, acceptable downtime)

### Typical mistakes

- Building for 99.999% availability everywhere when only one component is truly business-critical
- Not accounting for the "cost" of operations in the reliability equation — a complex HA system that requires constant human intervention is less reliable than a simpler system, not more

---

## 36. Latency vs Throughput

### What it is

**Latency:** how long a single operation takes (milliseconds per request).

**Throughput:** how many operations the system handles per unit time (requests per second).

### The relationship

They are related but not the same. A system can have:
- **High throughput, high latency:** a batch system processes millions of records per hour but each record takes minutes
- **Low throughput, low latency:** a single-threaded system handles one request at a time with sub-millisecond response — low latency, but throughput is 1 RPS

**Little's Law:** `L = λ × W`
- L = average number of requests in the system (queue + processing)
- λ = throughput (requests per second)
- W = average latency (seconds)

As latency increases under load, the system fills up with in-flight requests, which increases queue depth, which increases latency further — a feedback loop that causes service degradation.

### Optimising for each

**To reduce latency:** reduce the critical path — fewer serial operations, caching, precomputation, moving logic closer to data.

**To increase throughput:** parallelism — more workers, more partitions, batching, async processing.

### Real-world examples

- **Database query:** adding an index reduces latency; adding read replicas increases throughput
- **Kafka:** high throughput by design (sequential disk writes, batching); individual message latency is not Kafka's strength
- **Redis:** sub-millisecond latency by design; single-threaded for certain operations (limits throughput)

---

## 37. Availability vs Durability

### What it is

**Availability:** the system is operational and responding to requests. Measured as uptime percentage.

**Durability:** data that has been committed is not lost, even in the face of hardware failures.

### These are different

A system can be:
- **Available but not durable:** the database accepts writes but does not fsync to disk — if power is lost, recently written data is gone. Redis without persistence: high availability, low durability.
- **Durable but not available:** the database is down for maintenance but all data is safe on disk. Data is preserved; requests fail.
- **Both:** a fully replicated, synchronously persisted database — the goal of most critical systems
- **Neither:** a volatile in-memory cache with no persistence and a single node

### The trade-off

Increasing durability often reduces availability:
- **Synchronous replication:** every write must be acknowledged by all replicas before returning success — durability is strong, but if any replica is slow or down, availability suffers
- **Async replication:** high availability (writes succeed immediately), but if the primary crashes before replicating, data is lost — lower durability

### Real-world examples

- **AWS S3:** 11 9s durability (99.999999999%) — data is stored across multiple AZs with synchronous replication; losing S3 data is practically impossible. Availability is 99.99% — occasionally S3 is temporarily unavailable for reads/writes.
- **Redis without persistence:** very high availability (fast, in-memory); zero durability (restart = data loss)
- **PostgreSQL with synchronous_commit = on:** durable (fsync before ack); potentially lower availability under write load
- **PostgreSQL with synchronous_commit = off:** higher availability/throughput; risk of losing last few transactions if the server crashes

---

*This guide is a living reference. Engineering decisions are always contextual — the best choice for a 10-engineer startup serving 10,000 users is rarely the right choice for a 1,000-engineer company serving 1 billion users. Start simple, measure, and evolve.*