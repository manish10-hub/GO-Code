# API Design — Interview Cheat Sheet

## 1. REST Basics

Use **REST by default** for resource-oriented APIs.

```text
GET    /users/123       → Read
POST   /users           → Create
PUT    /users/123       → Replace
PATCH  /users/123       → Partial update
DELETE /users/123       → Delete
```

Think **nouns/resources**, not actions:

```text
✅ GET /users/123
❌ GET /getUser?id=123
```

---

## 2. CRUD + Idempotency ⭐

**Idempotent** = making the same request multiple times produces the same final state.

| Method | CRUD           | Idempotent? |
| ------ | -------------- | ----------- |
| GET    | Read           | ✅ Yes       |
| POST   | Create         | ❌ No        |
| PUT    | Replace        | ✅ Yes       |
| PATCH  | Partial update | ⚠️ Depends  |
| DELETE | Delete         | ✅ Yes       |

Example:

```text
PUT /users/123
{"name": "Manish"}
```

Calling it 10 times → user still has `name = Manish`.

POST can create duplicates:

```text
POST /payments
```

For payment/order/booking APIs, use an **Idempotency-Key**:

```text
Idempotency-Key: abc-123
```

Server stores the result for that key and returns the same result on retry.

---

## 3. Path vs Query vs Body ⭐

### Path → identifies required resource

```text
GET /events/123/tickets
```

### Query → optional filtering/sorting/pagination

```text
GET /tickets?event_id=123&status=available
```

### Body → data being created/updated

```text
POST /users

{
  "name": "Manish"
}
```

**Remember:**

> Path = **Which resource?**
> Query = **How/filter?**
> Body = **What data?**

---

## 4. Nested Resources

Use nested resources when the relationship is important/required:

```text
GET /events/123/tickets
```

> Tickets belonging to event 123.

Use query parameters when the relationship is an optional filter:

```text
GET /tickets?event_id=123
```

---

## 5. Pagination ⭐

Never return millions of records.

### Offset pagination

```text
GET /orders?offset=20&limit=10
```

Simple, but can have problems when data changes between requests.

### Cursor pagination

```text
GET /orders?cursor=abc123&limit=10
```

Response:

```json
{
  "orders": [...],
  "next_cursor": "xyz789"
}
```

Better for **large/high-volume/changing datasets**.

**Interview rule:**

> Small/simple → offset.
> Large/changing data → cursor.

---

## 6. Filtering & Sorting

Use query parameters:

```text
GET /orders?status=paid
GET /orders?sort=created_at
GET /orders?status=paid&limit=20
```

Be consistent:

```text
❌ /orders?status=paid
❌ /payments?state=paid

✅ /orders?status=paid
✅ /payments?status=paid
```

Same concept → same parameter name.

---

## 7. HTTP Status Codes

```text
200 → Success
201 → Created
400 → Bad request
401 → Not authenticated
403 → Not authorized
404 → Not found
409 → Conflict
429 → Rate limited
500 → Server error
```

Remember:

```text
401 → Who are you?
403 → I know who you are, but you're not allowed.
```

---

## 8. Authentication vs Authorization

**Authentication:**

> Who are you?

**Authorization:**

> What are you allowed to do?

Example:

```text
JWT → user = 123
       ↓
Can user 123 access /orders/456?
```

Use consistent authentication/authorization across APIs.

---

## 9. API Keys vs JWT

**API Key:**

* Identifies an application/service.
* Common for developer APIs.
* Simple service/application authentication.

**JWT/OAuth/OIDC:**

* Common for user authentication.
* Carries claims such as user ID, role, expiration.

Don't assume API keys are always best for service-to-service; mTLS or OAuth2 client credentials are also common.

---

## 10. API Versioning

Breaking changes should not break existing clients.

```text
/v1/users
/v2/users
```

Keep backward compatibility where possible.

---

## 11. Consistency ⭐

Use the same:

* Naming conventions
* Parameter names
* Response structures
* Error format
* Pagination style

Example:

```text
GET /orders?status=paid
GET /payments?status=paid
```

not:

```text
/orders?status=paid
/payments?state=paid
```

> **Consistency makes APIs predictable.**

---

## 12. Error Response

Use a consistent structure:

```json
{
  "error": {
    "code": "PAYMENT_FAILED",
    "message": "Payment could not be processed"
  }
}
```

`code` → machine-readable
`message` → human-readable

---

## 13. REST vs GraphQL vs gRPC

```text
REST     → Resource-oriented APIs / default choice
GraphQL  → Flexible client-specific data requirements
gRPC     → Internal service-to-service communication
```

---

# ⭐ 30-Second Revision

```text
REST → resources, not actions

GET    → Read      → Idempotent
POST   → Create    → Not idempotent
PUT    → Replace   → Idempotent
PATCH  → Partial   → Depends
DELETE → Delete    → Idempotent

Path  → Resource identity
Query → Filter/sort/pagination
Body  → Create/update data

Offset  → Simple pagination
Cursor  → Large/changing datasets

POST + retry → Idempotency-Key

401 → Authentication
403 → Authorization
429 → Rate limit

Consistency → Same naming, parameters, responses, errors

REST → default
GraphQL → flexible data fetching
gRPC → service-to-service
```





# Data Modeling – System Design Interview Notes

## 1. What is Data Modeling?

Data modeling = deciding:

* What entities do we store?
* What fields do they have?
* How are entities related?
* How will we query the data?
* How will the data scale?

**Interview goal:** Don't create a perfect schema. Create a **simple schema that supports the main requirements and query patterns**.

---

## 2. Choose the Database

### Relational DB / SQL — Default choice

Examples: PostgreSQL, MySQL

Use when you need:

* Structured data
* Relationships / joins
* Transactions
* Strong consistency
* ACID guarantees

**Default interview answer:** Start with PostgreSQL unless requirements suggest otherwise.

### Document DB

Examples: MongoDB, Firestore

Use when:

* Schema changes frequently
* Data is naturally nested
* Different records have different structures

Trade-off: fewer joins, but more denormalization and update complexity.

### Key-Value

Examples: Redis, DynamoDB

Use for:

* Cache
* Sessions
* Feature flags
* Simple `key → value` lookups

Very fast, but limited querying.

### Wide-Column

Examples: Cassandra, HBase

Use for:

* Very high write volume
* Time-series data
* Event/telemetry workloads

Model heavily around **query/access patterns**.

### Graph DB

Examples: Neo4j

Use when relationship traversal is the core problem.

**Interview tip:** Don't choose Graph DB just because the problem contains relationships. SQL is often sufficient.

---

# 3. Start With Requirements

Three things drive your data model:

### 1. Data Volume

How much data will we store?

### 2. Access Patterns ⭐

How will we query the data?

Ask:

> "What queries/API endpoints do we need to support?"

### 3. Consistency

Which data requires strong consistency vs eventual consistency?

Example:

* Payment → strong consistency
* Activity feed → eventual consistency may be acceptable

**Key rule:**

> Access patterns should drive schema, indexes, and partitioning.

---

# 4. Entities, Keys & Relationships

Example:

```text
users
  id (PK)
  username
  email

posts
  id (PK)
  user_id (FK)
  content
  created_at

comments
  id (PK)
  post_id (FK)
  user_id (FK)
  content
```

### Primary Key

Uniquely identifies a record.

Prefer:

```text
user_id
post_id
order_id
```

instead of business data such as email.

**Why?**

Business attributes can change; IDs should remain stable.

---

## 5. Relationships

### One-to-Many (1:N)

```text
User → Posts
Post → Comments
```

One user can have many posts.

### Many-to-Many (N:M)

```text
Users ↔ Posts
```

Example: users liking posts.

Usually represented using a junction table:

```text
likes
  user_id
  post_id
```

### One-to-One (1:1)

Less common.

Often ask whether both tables can simply be merged.

---

# 6. Foreign Keys

Example:

```text
posts.user_id → users.id
```

Foreign keys provide **referential integrity**.

They prevent things like:

```text
post.user_id = 999
```

when user 999 doesn't exist.

At very large scale, some systems enforce integrity at the application level instead of using DB foreign keys to reduce write overhead.

---

# 7. Constraints

Common constraints:

```text
PRIMARY KEY
FOREIGN KEY
UNIQUE
NOT NULL
CHECK
```

Example:

```sql
email UNIQUE
price CHECK (price > 0)
```

Use constraints to protect data correctness.

---

# 8. Indexing ⭐⭐⭐

An index allows the DB to find records without scanning the entire table.

**Important interview rule:**

> Design indexes based on your queries, not just your columns.

Example:

API:

```text
GET /users/{id}/posts
```

Useful index:

```text
posts(user_id)
```

If we frequently need recent posts:

```text
posts(user_id, created_at)
```

### Example

Query:

```sql
SELECT *
FROM posts
WHERE user_id = ?
ORDER BY created_at DESC;
```

Good index:

```text
(user_id, created_at)
```

Always explain **why** you need the index.

---

# 9. Normalization vs Denormalization

## Normalization

Store data in one place.

```text
users
  id
  username

posts
  id
  user_id
  content
```

### Advantages

* Less duplication
* Easier updates
* Better consistency

### Disadvantage

May require joins → slower reads at scale.

---

## Denormalization

Duplicate data to make reads faster.

```text
posts
  id
  user_id
  username
  content
```

### Advantages

* Faster reads
* Fewer joins

### Disadvantages

* Duplicate data
* More storage
* Updates become harder
* Risk of inconsistent data

### Interview rule ⭐

> Start normalized. Denormalize only when there is a clear performance/read requirement.

Common cases:

* Analytics
* Search
* Read-heavy systems
* Precomputed views
* Event/audit snapshots

Often you can keep the DB normalized and use a **cache/materialized view** for the denormalized read model.

---

# 10. Sharding / Partitioning

When one database cannot handle the data volume, distribute data across multiple machines.

Example:

```text
Shard 1 → users A–F
Shard 2 → users G–M
Shard 3 → users N–Z
```

### Choose shard key based on access pattern

If most queries are:

```text
Get posts by user_id
```

then:

```text
shard key = user_id
```

This keeps a user's posts together.

### Avoid

Cross-shard queries whenever possible.

They require:

```text
Query multiple shards
        ↓
Merge results
        ↓
Return response
```

which is expensive and complex.

### Hot shard

Be careful with time-based sharding:

```text
2026-09 → Shard 1
2026-10 → Shard 2
```

If all new writes go to the latest shard, it can become a **hot shard**.

---

# 11. Query Pattern → Data Model

This is one of the most important interview concepts.

Think:

```text
API
 ↓
Query
 ↓
Index
 ↓
Schema
 ↓
Partition/Sharding
```

Example:

Requirement:

> Get a user's recent posts.

API:

```text
GET /users/{user_id}/posts
```

Query:

```sql
WHERE user_id = ?
ORDER BY created_at DESC
```

Schema:

```text
posts(
    id,
    user_id,
    content,
    created_at
)
```

Index:

```text
(user_id, created_at)
```

Possible shard key:

```text
user_id
```

Everything is driven by the same access pattern.

---

# 12. Quick Database Selection

| Requirement                    | Good Choice |
| ------------------------------ | ----------- |
| Transactions / relationships   | SQL         |
| Flexible nested documents      | Document DB |
| Simple key lookup / cache      | Key-Value   |
| Massive writes / time-series   | Wide-Column |
| Complex relationship traversal | Graph DB    |

**Interview default:** PostgreSQL + Redis/cache is often a very reasonable starting point.

---

# 13. Interview Process

When asked to design the data model:

### Step 1 — Identify entities

```text
User
Post
Comment
Like
```

### Step 2 — Define important fields

```text
id
created_at
user_id
...
```

### Step 3 — Define relationships

```text
User → Posts
Post → Comments
User ↔ Post (Likes)
```

### Step 4 — Choose database

Usually:

```text
PostgreSQL
```

unless requirements suggest otherwise.

### Step 5 — Identify important queries

```text
Get user's posts
Get recent posts
Get comments
```

### Step 6 — Add indexes

```text
posts(user_id)
posts(user_id, created_at)
```

### Step 7 — Discuss normalization

Start normalized.

Denormalize only if required for performance.

### Step 8 — Discuss scaling

If necessary:

```text
Read replicas
Caching
Partitioning
Sharding
```

---

# 30-Second Revision

Remember these **7 points**:

```text
1. Start with requirements.
2. Identify entities and relationships.
3. Use stable system-generated primary keys.
4. Design indexes from query/access patterns.
5. Start normalized.
6. Denormalize for clear read-performance needs.
7. Shard based on the main access pattern.
```

### Golden Rule ⭐

> **Don't design the database first. Design it around how the system needs to read, write, and scale the data.**




# Caching — System Design Interview Notes

## 1. Why Caching?

Use caching when:

* Reads are very frequent.
* Database queries are slow or expensive.
* Database CPU/load is becoming a bottleneck.
* Low latency is required.

```text
Client
  ↓
Application
  ↓
Cache ── HIT ──→ Return data
  │
  MISS
  ↓
Database
  ↓
Cache
  ↓
Return data
```

**Main benefits:**

* Lower latency
* Reduce database load
* Higher read throughput

**Main cost:**

* Stale data
* Cache invalidation
* Cache failures
* Additional complexity

---

# 2. Where Can We Cache?

### External Cache

Examples: Redis, Memcached

```text
Application → Redis → Database
```

* Shared by multiple application instances.
* Supports TTL and eviction policies.
* **Default caching answer in system design interviews.**

### CDN

Used mainly for static/media content.

```text
User → CDN Edge → Origin Server
```

* Content is served from a location close to the user.
* Reduces latency and origin-server load.
* Good for images, videos, static files.

### Client-Side Cache

Cache on:

* Browser
* Mobile device
* Local storage
* Client libraries

Very fast, but backend has less control over freshness.

### In-Process Cache

Cache directly inside application memory.

```text
Application
   │
   ├── Local Cache
   │
   └── Redis
```

* Faster than Redis because there is no network call.
* Each application instance has its own cache.
* Good for small, frequently accessed data.
* Use as an optimization, not usually as the primary shared cache.

---

# 3. Cache Architectures

## Cache-Aside / Lazy Loading ⭐

**Most important pattern for interviews.**

```text
Application
    │
    ▼
  Cache
    │
 ┌──┴──┐
HIT   MISS
 │      │
 ▼      ▼
Return  DB
        │
        ▼
      Cache
        │
        ▼
      Return
```

Steps:

1. Check cache.
2. Cache hit → return data.
3. Cache miss → read database.
4. Store result in cache.
5. Return result.

**Pros:** Simple, cache only requested data.

**Cons:** First request after a miss is slower.

---

## Write-Through

```text
Application
     │
     ▼
   Cache
     │
     ▼
 Database
```

Application writes to cache → cache synchronously writes to DB → response returned.

**Pros:** Cache and DB are updated together.

**Cons:**

* Slower writes.
* Can cache data that is never read.
* Still has consistency/dual-write failure scenarios.

---

## Write-Behind / Write-Back

```text
Application
     │
     ▼
   Cache
     │
     │ async
     ▼
 Database
```

Cache acknowledges the write first and writes to DB asynchronously.

**Pros:** Very fast writes, good write throughput.

**Cons:** Data can be lost if cache fails before flushing.

Use when **eventual consistency/data-loss risk is acceptable**.

---

## Read-Through

```text
Application
     │
     ▼
   Cache
     │
   MISS
     ▼
 Database
```

The application talks only to the cache.

On a miss, the cache itself loads data from the database.

**Less common for application-level Redis caching.**

CDNs are a common example of read-through behavior.

---

# 4. Cache Eviction Policies

When cache memory is full, we need to decide what to remove.

### LRU — Least Recently Used ⭐

Remove the item that hasn't been accessed for the longest time.

```text
Oldest                         Newest
  ↓                              ↓
[A] ↔ [B] ↔ [C] ↔ [D]
 ↑                              ↑
LRU                            MRU
```

**Safe/default answer for most interviews.**

---

### LFU — Least Frequently Used

Remove the item accessed the fewest times.

```text
A → 10 accesses
B → 2 accesses
C → 7 accesses

Evict → B
```

Useful when some keys remain consistently popular.

---

### FIFO

Remove the item that entered the cache first.

```text
[A] → [B] → [C]

Evict A
```

Simple but ignores access frequency.

---

### TTL — Time To Live

TTL is **not an eviction policy by itself**.

It specifies how long an entry remains valid.

```text
user:123 → value
TTL = 10 minutes
```

After 10 minutes → entry expires.

Often combined with LRU/LFU.

---

# 5. Cache Consistency

Problem:

```text
Database → NEW value
Cache    → OLD value
```

Users may receive stale data.

### Common solution: Invalidate on Write

```text
1. Update DB
2. Delete cache entry
```

Next read:

```text
Cache MISS
   ↓
Read DB
   ↓
Populate Cache
```

Other approaches:

* Short TTL
* Accept eventual consistency
* Update cache after DB write

**There is no perfect caching strategy. Choose based on freshness requirements.**

---

# 6. Cache Stampede / Thundering Herd ⭐

Problem occurs when a popular key expires.

```text
              Cache MISS
                  │
        ┌─────────┼─────────┐
        ↓         ↓         ↓
      Request   Request   Request
        │         │         │
        └─────────┼─────────┘
                  ↓
              Database
```

Hundreds/thousands of requests may simultaneously hit the DB.

### Solutions

**Request coalescing / Singleflight**

Only one request rebuilds the cache.

```text
1000 requests
      │
      ▼
  One DB request
      │
      ▼
  Populate cache
      │
      ▼
 Other requests receive result
```

**Cache warming**

Refresh popular keys before they expire.

**Probabilistic early expiration**

Expire some entries slightly earlier to avoid synchronized expiration.

---

# 7. Hot Keys ⭐

A hot key receives extremely high traffic.

Example:

```text
user:popular-celebrity
        │
        ▼
  Millions of requests
        │
        ▼
   One Redis shard
        │
        ▼
     Bottleneck
```

Even if overall Redis traffic is fine, one key can overload a single node/shard.

### Solutions

* Replicate hot keys across multiple cache nodes.
* Load balance reads.
* Use an in-process cache.
* Rate-limit abusive traffic.

---

# 8. Cache Failure

What happens if Redis goes down?

```text
Application
     │
     X
   Redis
     │
     ▼
 Database
```

If every request falls directly to DB, the DB may get overloaded.

Possible solutions:

* Database fallback
* Circuit breaker
* In-process cache
* Rate limiting
* Request coalescing

**Always think about what happens when the cache fails.**

---

# 9. Cache Key Design

Good cache keys should be:

* Unique
* Predictable
* Easy to construct

Examples:

```text
user:123:profile

product:456

trending:posts:global
```

For multiple dimensions:

```text
user:123:feed:page:1
```

---

# 10. When Should You Introduce Caching?

Don't immediately say:

> "Let's add Redis."

First identify the bottleneck.

### Good reasons

```text
High read traffic
      ↓
Database overloaded
      ↓
Many repeated reads
      ↓
Add cache
```

Or:

```text
Expensive computation
      ↓
Same result requested repeatedly
      ↓
Cache computed result
```

Or:

```text
DB latency = 50ms
Required API latency = <10ms
      ↓
Use cache
```

---

# 11. Interview Approach ⭐

When introducing caching, follow this order:

```text
1. Identify bottleneck
        ↓
2. Decide what to cache
        ↓
3. Design cache key
        ↓
4. Choose cache architecture
        ↓
5. Choose eviction + TTL
        ↓
6. Handle invalidation
        ↓
7. Handle cache failure
        ↓
8. Handle stampede / hot keys if relevant
```

### Example interview answer

> "The system is read-heavy and the database is becoming the bottleneck. I'll use Redis as an external cache. I'll use cache-aside: check Redis first, and on a miss read from the database and populate Redis. I'll use LRU with a TTL to control memory and staleness. When data changes, I'll invalidate the corresponding cache entry. For very hot keys, I'll consider request coalescing or an in-process cache."

---

# 12. Quick Comparison

| Pattern       | Read               | Write                     | Main Use              |
| ------------- | ------------------ | ------------------------- | --------------------- |
| Cache-Aside   | App checks cache   | App writes DB             | ⭐ Default             |
| Read-Through  | Cache handles read | —                         | Specialized           |
| Write-Through | —                  | Cache → DB synchronously  | Fresh data            |
| Write-Behind  | —                  | Cache → DB asynchronously | High write throughput |

---

# 13. Last-Minute Revision

```text
CACHE
│
├── Why?
│   ├── Reduce latency
│   ├── Reduce DB load
│   └── Increase throughput
│
├── Where?
│   ├── Redis / Memcached
│   ├── CDN
│   ├── Client
│   └── In-process
│
├── Patterns
│   ├── Cache-Aside ⭐
│   ├── Read-Through
│   ├── Write-Through
│   └── Write-Behind
│
├── Eviction
│   ├── LRU ⭐
│   ├── LFU
│   ├── FIFO
│   └── TTL
│
└── Problems
    ├── Stale data
    ├── Invalidation
    ├── Cache failure
    ├── Stampede
    └── Hot keys
```

## Golden Rules

1. **Don't cache everything.**
2. **Start with cache-aside for most interview problems.**
3. **Use Redis for shared application-level caching.**
4. **Use CDN for static/media content.**
5. **Use LRU as the safe default eviction policy.**
6. **Use TTL when data should expire.**
7. **Always discuss cache invalidation.**
8. **Think about cache failure and DB overload.**
9. **For hot keys, consider replication or local caching.**
10. **Introduce caching only after identifying the bottleneck.**





# Sharding: System Design Interview Notes

## 1. Why do we need sharding?

Sharding is used when **one database can no longer handle the required scale**.

Typical bottlenecks:

* Storage capacity
* Write throughput
* Read throughput
* CPU / memory
* Connection limits

```text
Single Database
      |
      |  Scale limit
      v
+-----------+-----------+-----------+
|  Shard 1  |  Shard 2  |  Shard 3  |
+-----------+-----------+-----------+
```

### Key idea

> Sharding = splitting data across multiple machines so storage and traffic can scale horizontally.

**Don't shard too early.** First prove that a single database or read replicas cannot meet the requirements.

---

## 2. Partitioning vs Sharding

### Partitioning

Data is split into smaller pieces **inside one database machine**.

```text
Database
  |
  +-- Partition 1
  +-- Partition 2
  +-- Partition 3
```

### Sharding

Data is split across **multiple database machines**.

```text
             Application
                  |
        +---------+---------+
        |         |         |
     Shard 1   Shard 2   Shard 3
     DB-1      DB-2      DB-3
```

### Types of partitioning

**Horizontal partitioning**

* Split rows.
* Same columns, fewer rows per partition.

**Vertical partitioning**

* Split columns.
* Same rows, fewer columns per partition.

### Remember

> Partitioning = usually within one database.
> Sharding = data distributed across multiple machines.

---

## 3. Two important sharding decisions

When designing sharding, answer two questions:

### 1. What to shard by?

Choose the **shard key**.

Example:

```text
user_id
order_id
tenant_id
```

### 2. How to distribute?

Choose the distribution strategy:

```text
Range-based
Hash-based
Directory-based
```

---

## 4. Choosing a good shard key

A good shard key should have:

### High cardinality

Many unique values.

Good:

```text
user_id → millions of values
```

Bad:

```text
is_premium → true / false
```

### Even distribution

Data and traffic should be spread across shards.

Bad example:

```text
country

US    → 70%
India → 10%
UK    → 5%
...
```

One shard may become much larger/hotter.

### Align with query patterns

The most common queries should ideally hit **one shard**.

Example:

```text
Query:
Get orders for user 123

Shard key:
user_id

Result:
Only one shard needs to be queried.
```

### Interview rule

> Choose the shard key based on **access patterns + distribution + cardinality**.

---

# 5. Sharding strategies

## A. Range-Based Sharding

Data is divided into ranges.

```text
Shard 1 → user_id 1 - 1M
Shard 2 → user_id 1M - 2M
Shard 3 → user_id 2M - 3M
```

### Advantages

* Simple
* Good for range queries
* Easy to understand

### Problems

Can create hotspots.

Example:

```text
created_at

Shard 1 → Old data
Shard 2 → Old data
Shard 3 → Recent data  ← HOT
```

If most writes are for new records, the latest shard gets most traffic.

### Best suited for

* Natural ranges
* Multi-tenant systems
* Queries that frequently access ranges

---

## B. Hash-Based Sharding ⭐ Default

Hash the shard key and use the result to select a shard.

```text
shard = hash(user_id) % N
```

Example:

```text
hash(user_id) % 4

User 42  → Shard 2
User 99  → Shard 3
User 123 → Shard 1
```

### Advantages

* Usually good distribution
* Reduces hotspots caused by uneven key ranges
* Simple routing

### Major problem

Changing the number of shards can move lots of data.

```text
hash(user_id) % 4

       ↓ add shard

hash(user_id) % 5
```

Many records may map to different shards.

### Solution

Use **consistent hashing / resharding techniques** to reduce data movement.

### Interview default

> Unless there is a reason to choose something else, hash-based sharding is usually a good default.

---

## C. Directory-Based Sharding

Maintain a mapping of key → shard.

```text
user_to_shard

User 15  → Shard 1
User 87  → Shard 4
User 204 → Shard 2
```

### Advantages

* Very flexible
* Can move specific users/tenants
* Can isolate hot users

### Disadvantages

* Extra lookup on every request
* Directory becomes a critical dependency
* More system complexity

### Interview usage

Usually don't choose this as the first solution unless you specifically need dynamic shard placement.

---

# 6. Hotspots / Load Imbalance

A hotspot occurs when one shard receives much more traffic than others.

```text
Shard 1 → 10K requests
Shard 2 → 12K requests
Shard 3 → 500K requests  ← HOT
Shard 4 → 11K requests
```

### Common causes

#### Celebrity problem

One user is extremely popular.

```text
Taylor Swift's user_id
        |
        v
Millions of requests
        |
        v
One shard becomes HOT
```

Hashing doesn't completely solve this because the problem is the **traffic concentration on one key**.

#### Time-based sharding

```text
Shard 1 → old data
Shard 2 → old data
Shard 3 → today's data ← HOT
```

### Solutions

* Move hot keys to dedicated shards
* Use compound shard keys
* Spread a user's data across multiple shards
* Dynamically split shards
* Monitor shard CPU, latency and request volume

---

# 7. Cross-Shard Queries

This is one of the biggest problems with sharding.

Suppose:

```text
64 shards
```

Query:

```text
"Give me top 10 popular posts globally"
```

If posts are distributed across all shards:

```text
Application
    |
    +---- Shard 1
    +---- Shard 2
    +---- Shard 3
    ...
    +---- Shard 64
             |
             v
         Aggregate
```

The application must:

1. Query all shards.
2. Wait for responses.
3. Merge results.
4. Calculate the final result.

### Problems

* More network calls
* Higher latency
* More application complexity
* Failure of one shard can affect the query

### How to reduce cross-shard queries

**1. Cache results**

Good for:

* Trending posts
* Leaderboards
* Aggregate statistics

**2. Denormalize**

Keep frequently related data together.

**3. Precompute**

Use background jobs to calculate expensive global results.

**4. Accept the cost**

If the query is rare, querying all shards may be acceptable.

### Interview rule

> Common queries should ideally be single-shard queries.

---

# 8. Cross-Shard Transactions

Transactions become difficult when related data lives on different shards.

Example:

```text
User A → Shard 1
User B → Shard 2
```

Money transfer:

```text
Shard 1: Deduct $100
Shard 2: Add $100
```

A normal database transaction cannot atomically cover both databases.

---

## Option 1: Avoid cross-shard transactions ⭐

Choose the shard key so related data stays together.

```text
user_id
   |
   +-- Profile
   +-- Account
   +-- Transactions
```

All live on the same shard.

**This is usually the best solution.**

---

## Option 2: Two-Phase Commit (2PC)

```text
Coordinator
    |
    +---- Prepare → Shard 1
    |
    +---- Prepare → Shard 2
    |
    +---- Commit  → Shard 1
    |
    +---- Commit  → Shard 2
```

### Problem

* Slow
* Complex
* Failure-prone
* Can block during failures

Most large-scale systems prefer avoiding 2PC where possible.

---

## Option 3: Saga

Break the transaction into independent steps.

Example:

```text
1. Deduct money from A
2. Add money to B

If step 2 fails:
3. Compensate step 1
   → Refund A
```

Saga provides **eventual consistency** using compensating actions.

---

# 9. Consistency Across Shards

Sharding can make strong consistency more difficult.

Possible approaches:

### Strong consistency

Use distributed transactions when required.

### Eventual consistency

Accept temporary inconsistency.

Example:

```text
Follower count

Shard 1 → 1000
Shard 2 → 998

After some time:

Shard 1 → 1000
Shard 2 → 1000
```

For many applications, eventual consistency is acceptable.

### Interview rule

> First design the system to avoid cross-shard transactions. Use Saga or eventual consistency when cross-shard coordination is unavoidable.

---

# 10. Resharding

As data grows, existing shards may become too large.

```text
Before:

Shard 1
Shard 2
Shard 3

        ↓ Add capacity

Shard 1
Shard 2
Shard 3
Shard 4
Shard 5
```

The system needs to move some data to new shards.

### Problem with simple hash

```text
hash(key) % 3
```

Changing to:

```text
hash(key) % 5
```

can remap many keys.

### Better approach

Use:

* Consistent hashing
* Online resharding
* Database-managed partition movement
* Rebalancing

### Interview question

**"What happens when we add a new shard?"**

Good answer:

> We need a resharding strategy. With simple modulo hashing, many keys move. I'd use consistent hashing or an online resharding mechanism to minimize data movement and keep the system available.

---

# 11. Modern databases

You usually don't implement sharding yourself.

Examples:

| Database  | Sharding / Partitioning               |
| --------- | ------------------------------------- |
| Cassandra | Partition key + token ranges          |
| DynamoDB  | Partition key                         |
| MongoDB   | Sharding with chunks                  |
| Vitess    | MySQL sharding layer                  |
| Citus     | PostgreSQL distributed/sharding layer |

The exact internal mechanism differs between databases.

### Interview answer

You can say:

> "I'd use DynamoDB with `user_id` as the partition key."

or:

> "I'd use Vitess to shard MySQL by `user_id`."

You generally don't need to implement the internals unless the interviewer asks.

---

# 12. When should I introduce sharding?

Don't say:

> "We'll use sharding because the system is large."

Instead:

```text
Requirements
     ↓
Capacity estimation
     ↓
Single DB capacity
     ↓
Is it enough?
     |
   YES → Start with single DB
     |
    NO
     ↓
Consider sharding
```

### Typical reasons

* Database storage is too large
* Write throughput is too high
* Read throughput is too high
* Single machine CPU/memory is insufficient

### Interview golden rule

> **Prove that you need sharding before introducing it.**

---

# 13. Interview approach

When discussing sharding, follow this order:

```text
1. Identify the database bottleneck
        ↓
2. Estimate required capacity
        ↓
3. Choose shard key
        ↓
4. Explain why the key fits query patterns
        ↓
5. Choose distribution strategy
        ↓
6. Discuss hotspots
        ↓
7. Discuss cross-shard queries
        ↓
8. Discuss cross-shard transactions
        ↓
9. Explain resharding
```

---

# 14. Example Interview Answer

### Question

**How would you shard a social media application?**

### Answer

> Most queries are user-centric, so I would shard by `user_id`. I would use hash-based sharding to distribute users evenly across shards. This allows profile, posts and other user-specific queries to remain on a single shard.
>
> The main trade-off is that global queries such as trending posts may need to query multiple shards. I would handle those using caching and background precomputation. For very popular users, I would consider isolating hot keys or using a compound key. I would also plan for online resharding as the system grows.

---

# 15. Quick Comparison

| Strategy  | Distribution | Strength               | Main Problem            |
| --------- | ------------ | ---------------------- | ----------------------- |
| Range     | Value ranges | Simple + range queries | Hotspots                |
| Hash      | Hash(key)    | Even distribution      | Resharding              |
| Directory | Lookup table | Very flexible          | Extra dependency/lookup |

---

# 16. Sharding Mental Model

```text
                    SHARDING
                       |
        +--------------+--------------+
        |              |              |
     Shard Key      Strategy       Challenges
        |              |              |
   user_id         Range          Hotspots
   order_id        Hash           Cross-shard
   tenant_id       Directory      Transactions
                                  Resharding
```

---

# 17. Last-Minute Revision

```text
Sharding
|
+-- Why?
|   +-- Storage limit
|   +-- Read throughput
|   +-- Write throughput
|
+-- Shard Key
|   +-- High cardinality
|   +-- Even distribution
|   +-- Matches query patterns
|
+-- Strategies
|   +-- Range
|   +-- Hash ⭐ Default
|   +-- Directory
|
+-- Problems
|   +-- Hotspots
|   +-- Cross-shard queries
|   +-- Cross-shard transactions
|   +-- Resharding
|
+-- Solutions
    +-- Good shard key
    +-- Consistent hashing
    +-- Caching
    +-- Denormalization
    +-- Precomputation
    +-- Saga
```

## 5 Things to Remember

1. **Don't shard too early.** Prove the single DB is the bottleneck.
2. **Shard key is critical:** high cardinality + even distribution + query aligned.
3. **Hash-based sharding is the common default.**
4. **Minimize cross-shard queries and transactions.**
5. **Always have a resharding strategy.**

### One-line interview summary

> **Shard when one database cannot scale; choose a high-cardinality, evenly distributed key that matches access patterns, then address hotspots, cross-shard operations, and resharding.**
