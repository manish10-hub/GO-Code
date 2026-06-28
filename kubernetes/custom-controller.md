# Kubernetes Operators, Controllers, CRDs & client-go
### Interview Q&A — Organized Beginner → Expert

---

## Table of Contents

- [Level 1 — Beginner](#level-1--beginner)
- [Level 2 — Intermediate](#level-2--intermediate)
- [Level 3 — Advanced](#level-3--advanced)
- [Level 4 — Expert](#level-4--expert)

---

## Level 1 — Beginner

---

### Q1. What is a Kubernetes Operator?

A Controller is any piece of code that runs a reconciliation loop:
Deployment controller, ReplicaSet controller, Namespace controller — none of these are "Operators," they're just controllers for resources Kubernetes already understands natively.

**Operator — a controller for your application, with domain knowledge**

An Operator is a Controller that manages a Custom Resource (a resource type you defined yourself via a CRD) and encodes operational knowledge specific to that application — the kind of knowledge a human SRE would otherwise apply manually.

- All Operators are Controllers.
- Not all Controllers are Operators.

The distinguishing feature: an Operator knows how to run a specific application well — not just "keep N replicas running."

**The basic flow:**

```
User creates CR
        ↓
API Server stores CR
        ↓
Operator Watch receives event
        ↓
Reconcile()
        ↓
Create/Update/Delete resources
        ↓
Update Status
        ↓
Requeue if needed
```

---

### Q2. Why do we need Operators? Can we not just use Deployments and Services?

Deployments and Services are great for stateless applications. But stateful applications — databases, message queues, search engines — have complex operational requirements that plain Kubernetes primitives cannot express:

- A database needs ordered startup (primary before replicas)
- Scaling a database requires copying data, not just starting a new pod
- Backup needs to be triggered at specific times with specific logic
- Upgrades need to happen in a specific sequence to avoid data corruption

---

### Q3. What is the control loop (reconciliation loop)? Explain it simply.

The control loop is the heartbeat of every Operator. It follows one simple pattern repeated forever:

```
Observe  →  Compare  →  Act
```

---

### Q4. What are some real-world examples of popular Operators?

- **etcd Operator** — manages etcd cluster lifecycle, scaling, and backup
- **Prometheus Operator** — manages Prometheus instances, alerting rules, and scrape configs via ServiceMonitor CRDs
- **cert-manager** — automates TLS certificate issuance and renewal
- **Strimzi** — manages Apache Kafka clusters on Kubernetes
- **Rook** — manages Ceph storage clusters

---

### Q5. What is a Custom Resource Definition (CRD)?

A CRD is a way to extend the Kubernetes API with your own resource types. Once you create a CRD, you can use kubectl to create, read, update, and delete instances of that resource — called Custom Resources (CRs) — just like you would with built-in resources like Pods or Deployments.

---

### Q6. What are the main components of a CRD manifest?

```
CRD
├── Group      : Defines the API domain
├── Version(s) : Defines the API version
├── Scope      : Determines where resources live, namespace/cluster
├── Names      : Defines how users interact with the resource, plural, singular
├── Schema     : Defines validation rules
├── Subresources : Status, Spec
├── Additional Printer Columns
└── Conversion Strategy (optional)
```

---

### Q7. What is the difference between `spec` and `status` in a Custom Resource?

- **spec** — the desired state. Written by the user or automation. It describes what the user wants. Operators read spec to know what to create or change.
- **status** — the observed state. Written only by the Operator. It describes what actually exists in the cluster right now. Users read status to understand the current health of the resource.

They are updated separately:
- `spec` is updated with `client.Update()`
- `status` is updated with `client.Status().Update()` via the status subresource

This separation is important: if a user updates spec and the Operator updates status at the same time, they do not overwrite each other because they go to different API endpoints.

**The golden rule:** users write spec, Operators write status, nobody writes both in the same call.

---

### Q8. What is client-go and what does it provide?

client-go is the official Go client library for interacting with the Kubernetes API server. It is the foundational library that almost every Go-based Kubernetes tool uses — including kubectl, controller-runtime, and every major Operator framework.

It provides:
- **Typed clients** — strongly typed Go functions for every built-in Kubernetes resource (`clientset.CoreV1().Pods(ns).Get(...)`)
- **Dynamic client** — works with any resource without generated code, using `unstructured.Unstructured`
- **REST client** — low-level HTTP client for the Kubernetes API
- **Informers** — efficient list/watch mechanism with local caching
- **Work queues** — rate-limited queues for controller patterns
- **Discovery client** — discovers what API groups and resources are available

---

### Q9. How do you create a client-go clientset and what are the authentication options?

Create a `rest.Config` using either `clientcmd.BuildConfigFromFlags()` (outside the cluster), `rest.InClusterConfig()` (inside the cluster), or by constructing it manually.

Pass the `rest.Config` to `kubernetes.NewForConfig()` to obtain a typed Clientset.

A clientset is the typed Kubernetes client provided by client-go. It exposes strongly-typed clients for each Kubernetes API group:
```go
clientset.CoreV1().Pods(namespace)
clientset.AppsV1().Deployments(namespace)
clientset.BatchV1().Jobs(namespace)
```

**Option 1: In-cluster (running inside a Pod)**
```go
rest.InClusterConfig()
```
`InClusterConfig()` automatically reads:
- ServiceAccount token
- CA certificate
- API server address

**Option 2: Out-of-cluster (local dev, CI)**
```go
// Reads from kubeconfig file
config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
clientset, err := kubernetes.NewForConfig(config)
pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
```

**Authentication options:**
1. Client certificates — the kubeconfig contains them
2. Bearer Token — the client sends it directly
3. ServiceAccount Token — automatically mounted inside a Pod

---

### Q10. What is the difference between List and Watch in client-go?

- **List** — a one-time snapshot of all resources matching a selector. Returns a list of objects at a point in time with a `resourceVersion`. Expensive at scale — always fetches from etcd.
- **Watch** — a long-running HTTP connection that receives incremental events (ADDED, MODIFIED, DELETED) as resources change. Much more efficient than repeated polling.

**In practice:** never use List+Watch manually. Use Informers, which combine them correctly — List to get initial state, then Watch from the returned `resourceVersion` to receive only changes. Informers also handle reconnection automatically.

---

## Level 2 — Intermediate

---

### Q11. What are the phases of an Operator's lifecycle for a managed resource?

A well-designed Operator manages the full lifecycle of its resource:

```
Install     → create all necessary Kubernetes objects
Configure   → apply configuration based on CR spec
Upgrade     → roll out new application versions safely
Scale       → add/remove replicas with data awareness
Backup      → trigger and manage backup jobs
Recovery    → detect failure and restore automatically
Delete      → clean up all resources (finalizer-based)
```

Each phase maps to specific reconciliation logic. The Operator checks the current phase of the CR status and decides what action to take next — making it effectively a state machine.

---

### Q12. What is Owner Reference and why is it important for Operators?

Owner Reference is a field on a Kubernetes object that points to another object that owns it. When the owner is deleted, Kubernetes garbage collects all owned objects automatically — this is called cascading deletion.

In Operators, you set owner references on every object your Operator creates (Pods, Services, ConfigMaps, PVCs) to point to the parent Custom Resource.

---

### Q13. What is a finalizer and when should you use it?

A finalizer is a string added to a resource's `metadata.finalizers` list that prevents the resource from being deleted until your Operator explicitly removes it.

When a user runs `kubectl delete mycr`, Kubernetes sets `deletionTimestamp` on the resource but does NOT delete it as long as finalizers are present. Your Operator sees the non-zero `deletionTimestamp`, performs cleanup logic, then removes the finalizer — allowing Kubernetes to proceed with deletion.

---

### Q14. What is the difference between Cluster-scoped and Namespace-scoped Operators?

- A **Namespace-scoped Operator** watches and manages resources only within specific namespaces. It uses a `RoleBinding` for permissions. Safer — blast radius is limited to one namespace. Good for tenant-specific operators.
- A **Cluster-scoped Operator** watches resources across all namespaces. It uses a `ClusterRoleBinding` for permissions. Required when the managed resource is itself cluster-scoped (like StorageClass, ClusterRole, Node) or when the Operator needs to manage resources across multiple namespaces.

---

### Q15. What is a subresource in a CRD and why does it matter?

A subresource is a dedicated API endpoint for a specific part of a Kubernetes resource, such as status or scale. It enables independent updates, reduces conflicts, enforces separation of responsibilities, and allows controllers to update resource status without modifying the desired state stored in the spec.

- RBAC can grant write access to status separately from spec
- Users cannot accidentally overwrite Operator-managed status when updating spec

---

### Q16. What is controller-runtime and what problem does it solve?

controller-runtime is a highly opinionated, production-grade Go library maintained by the Kubernetes Special Interest Group (SIG) API Machinery. It provides the structural scaffolding, tools, and design patterns required to build custom Kubernetes Operators and Controllers.

If you are writing an application that manages custom or core Kubernetes resources (like an ImageArchive controller), you are almost certainly using controller-runtime under the hood — likely generated via tools like Kubebuilder or Operator SDK.

Writing a bare-metal Kubernetes controller using only raw client-go requires a massive amount of complex, repetitive boilerplate code. Before controller-runtime, developers had to manually orchestrate low-level multi-threaded streaming architectures just to safely watch resource events.

**Without controller-runtime, building a stable controller forces you to solve several difficult distributed systems problems by hand:**
- Set up watch connections to the API server for every resource type
- Implement a work queue with rate limiting and retry logic
- Write leader election code
- Set up health check and metrics endpoints
- Handle cache synchronisation and list/watch mechanics

**The three core concepts are:**
- **Manager** — the top-level object that runs everything (cache, controllers, webhooks, health checks)
- **Reconciler** — the interface you implement with your business logic
- **Builder** — fluent API to wire up which resources your controller watches

---

### Q17. What is the Manager in controller-runtime and what does it manage?

The Manager is the central orchestrator of a controller-runtime based Operator. It owns and coordinates:

In controller-runtime, the Manager is the central orchestrator and the single execution engine of your entire controller application. If a controller application were a ship, the Manager would be the engine room — it initializes, configures, and runs all the core dependencies required to keep your Operator alive.

- **Client** — a Kubernetes API client (with caching) shared across all controllers
- **Cache** — an informer-based cache that keeps local copies of watched resources in memory
- **Work queues** — per-controller rate-limited work queues
- **Leader election** — ensures only one active Operator when running with replicas
- **Health/readiness endpoints** — `/healthz` and `/readyz`
- **Metrics server** — Prometheus metrics at `/metrics`
- **Webhook server** — serves admission and conversion webhooks

---

### Q18. What is the Reconciler interface? What must it implement?

The Reconciler is the heart of every Kubernetes Operator built with controller-runtime. It contains the business logic that moves the cluster from its current state to the desired state.

A Reconciler is a Go type that implements the `reconcile.Reconciler` interface.

For every event (Create, Update, Delete, periodic resync), Kubernetes calls it.

- `Request` contains the namespace and name of the object that triggered reconciliation.
- `Result` controls what happens after reconciliation.
- Returning an error causes exponential backoff retry. Return `ctrl.Result{RequeueAfter: ...}` with nil error when you want controlled polling without triggering backoff.

---

### Q19. What is client-go's typed clientset vs the resources it covers, and how do informers fit alongside it?

(Established context for Level 3 onward — see List vs Watch in Q10, and Informer internals in Q23.)

---

## Level 3 — Advanced

---

### Q20. How does an Operator know when to reconcile? What are the trigger mechanisms?

When an Operator initializes, it registers explicit "Watches" on specific resource types. There are three common categories of resources an Operator watches to trigger a reconciliation:

1. **Watching the Primary Resource** (The Custom Resource itself)
2. **Watching Secondary (Owned) Resources**
3. **Watching Arbitrary External Resources**

**The Internal Trigger Mechanics (Under the Hood)**

```
┌──────────────┐      ┌───────────┐      ┌───────────┐      ┌───────────┐
│  Kubernetes  │      │           │      │           │      │           │
│  API Server  ├─────►│ Reflector ├─────►│ Informer  ├─────►│ Indexer   │
│  (HTTP Watch)│      │           │      │  (Delta)  │      │  (Cache)  │
└──────────────┘      └───────────┘      └─────┬─────┘      └───────────┘
                                               │
                                               ▼ (Event Handlers)
                                         ┌───────────┐
                                         │ WorkQueue │
                                         └─────┬─────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │ Reconcile() │
                                        └─────────────┘
```

- **The Reflector & Informer** — A Reflector establishes a long-lived HTTP ListWatch connection to the API server. When an event happens, the Reflector captures it and pushes it into an Informer delta FIFO queue.
- **The Local Cache (Indexer)** — To avoid hitting the API server repeatedly, the Informer updates a highly efficient, local in-memory cache (Indexer). The Operator queries this local cache during reconciliation rather than overloading the API server.
- **The Event Handlers** — The Informer passes the raw event (Add, Update, Delete) to your controller's event handlers.
- **The WorkQueue** — The event handler does not invoke your business logic directly. Instead, it extracts just the identifying coordinates of the object — its namespace and name (`namespace/name`) — and places that string key into a thread-safe WorkQueue.
- **The Reconciler Deduping** — The WorkQueue automatically deduplicates keys. If an object is changing rapidly and generates 10 update events within a millisecond, the key is only placed in the queue once. A worker routine pulls the key out of the WorkQueue and passes it to your `Reconcile(ctx, req)` method.

---

### Q21. How do you design an Operator that is safe to run with multiple replicas (leader election)?

Running multiple Operator replicas is important for high availability but creates the risk of multiple controllers reconciling the same resource simultaneously — causing conflicts, duplicate actions, and split-brain.

The solution is leader election: only one replica is the active leader and runs the reconciliation loop. Others are on standby and take over if the leader dies.

controller-runtime has built-in leader election using a Kubernetes Lease object.

**Design considerations:**
- Make your reconciler idempotent — if leader election fails over mid-reconcile, the new leader must safely re-run from scratch
- Avoid storing reconcile state in memory — use the CR status subresource as the source of truth
- Be careful with external side effects (cloud API calls) — the new leader re-running could trigger duplicate resource creation. Use idempotency keys.

---

### Q22. How do you handle the "thundering herd" problem when an Operator manages thousands of custom resources?

When managing thousands of custom resources, prevent thundering herd effects by combining event-driven reconciliation, rate-limited work queues, exponential backoff with jitter, informer caches, and controlled concurrency. Avoid periodic requeues for all resources, limit `MaxConcurrentReconciles`, update status only when changes occur, and rely on controller-runtime's rate-limiting queue. For very large deployments, shard controllers horizontally by namespace, tenant, or region. The goal is to smooth reconciliation load and protect both the API server and downstream dependencies from synchronized bursts.

- **Event-driven reconciliation** — Reconcile only when relevant resources change instead of periodically reconciling all objects.
- **Rate-limited work queue** — Prevents failed reconciliations from being requeued immediately and overwhelming the controller or downstream systems.
- **Exponential backoff with jitter** — Spreads retries over time using increasing delays and randomness to avoid synchronized retry storms.
- **Informer cache** — Serves reads from a local cache instead of repeatedly querying the Kubernetes API server, reducing API load.
- **Limit Concurrent Reconciles** — Limits the number of concurrent reconciliations to protect the API server and external dependencies from traffic spikes.
- **Horizontal sharding** — Distributes resources across multiple controller instances to spread reconciliation load.
- **Leader election** — Ensures only one controller instance actively reconciles a resource set, preventing duplicate work.

---

### Q23. How do you implement a status condition pattern correctly in an Operator?

Status conditions are the standard Kubernetes pattern for communicating the health and readiness of a resource. They follow the `metav1.Condition` structure.

A status condition pattern should follow Kubernetes API conventions by using `metav1.Condition` with fields such as `Type`, `Status`, `Reason`, `Message`, `ObservedGeneration`, and `LastTransitionTime`. Conditions should be updated using `meta.SetStatusCondition()` to avoid duplicates, and status updates should occur only when the condition actually changes. The `Ready` condition should represent the overall usability of the resource, while `ObservedGeneration` ensures users can determine whether the controller has processed the latest spec. This provides a machine-readable, extensible, and observable way to communicate resource state.

---

### Q24. How does CRD versioning work? How do you handle breaking changes?

CRDs support multiple versions simultaneously. Each version can be:
- `served: true` — the API accepts requests for this version
- `storage: true` — only ONE version can be the storage version (stored in etcd)

When you have v1alpha1 and v1beta1:
```yaml
versions:
  - name: v1beta1
    served: true
    storage: true    # new storage version
  - name: v1alpha1
    served: true     # still served for backward compatibility
    storage: false
```

- For **non-breaking changes** (adding optional fields): just add to the schema and release.
- For **breaking changes** (renaming/removing fields): you need a conversion webhook.

**Conversion webhook** — a webhook that Kubernetes calls to convert between versions when reading or writing objects stored in a different version than requested.

```yaml
conversion:
  strategy: Webhook
  webhook:
    clientConfig:
      service:
        name: my-operator-webhook
        namespace: default
        path: /convert
```

The webhook receives the object in the stored version and must return it in the requested version — applying whatever field mappings are needed. This allows both versions to coexist without migrating all existing objects at once.

---

### Q25. How do you set up watches for secondary resources in controller-runtime?

In controller-runtime, a Secondary Resource is any object that your controller doesn't directly own, but needs to watch because changes to it impact your primary resource.

For example, if your primary resource is an ImageArchive (which you reconcile), but it creates a secondary StorageVolume (a core or custom resource) underneath it, you want your ImageArchive reconciler to wake up immediately if someone modifies or deletes that StorageVolume.

To set this up, you use the `Watches()` or `Owns()` chains when building your controller manager.

---

### Q26. What are Predicates in controller-runtime and how do you use them to reduce unnecessary reconciliations?

Predicates are filters that sit between the watch event and the work queue. If a predicate returns false, the event is dropped and the reconciler is never called. This reduces unnecessary reconciliations and saves CPU.

**Built-in predicates:**
- `GenerationChangedPredicate` — only spec changes, not status updates
- `ResourceVersionChangedPredicate` — any change
- `LabelChangedPredicate` — label changes only
- `AnnotationChangedPredicate` — annotation changes only

Custom predicates implement the `Predicate` interface.

---

### Q27. What is the dynamic client in client-go and when do you use it over the typed clientset?

The typed clientset only works with built-in Kubernetes resource types (Pods, Deployments, etc.) because it uses generated Go types. For Custom Resources, you either:

- Use generated typed clients (via controller-gen or kubebuilder) — type-safe, but requires code generation per CRD
- Use the **dynamic client** — works with any resource using `unstructured.Unstructured`, no code generation needed

```go
import (
    "k8s.io/client-go/dynamic"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
)

dynClient, _ := dynamic.NewForConfig(config)

// Define the GVR (Group Version Resource)
gvr := schema.GroupVersionResource{
    Group:    "myapp.example.com",
    Version:  "v1",
    Resource: "databases",
}

// List custom resources
list, err := dynClient.Resource(gvr).Namespace("default").List(ctx, metav1.ListOptions{})

// Access fields via unstructured
for _, item := range list.Items {
    name, _, _ := unstructured.NestedString(item.Object, "metadata", "name")
    replicas, _, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
    fmt.Printf("%s: %d replicas\n", name, replicas)
}

// Create a custom resource
obj := &unstructured.Unstructured{
    Object: map[string]interface{}{
        "apiVersion": "myapp.example.com/v1",
        "kind":       "Database",
        "metadata":   map[string]interface{}{"name": "my-db"},
        "spec":       map[string]interface{}{"replicas": 3},
    },
}
dynClient.Resource(gvr).Namespace("default").Create(ctx, obj, metav1.CreateOptions{})
```

**Use dynamic client when:** writing generic tools (like Helm, ArgoCD), building multi-tenant platforms where CRDs are not known at compile time, or writing tests that need to work with arbitrary resources.

---

## Level 4 — Expert

---

### Q28. How does the controller-runtime cache work? What is its relationship to informers?

The controller-runtime cache is a wrapper around Kubernetes informers.

**An informer:**
- Starts a ListWatch against the API server for a specific resource type
- Stores all objects in an in-memory Store (thread-safe map)
- Sends Add, Update, Delete events to registered event handlers

**The controller-runtime cache:**
- Creates one informer per resource type that any controller watches
- Shares a single cache instance across all controllers in a Manager — if two controllers watch Pods, they share one Pod informer
- Wraps the informer store in a `client.Reader` interface, so controllers call `r.Get()` and `r.List()` against the cache, not the API server

**Why this matters:**
- All reads (Get, List) in a controller are served from cache — sub-millisecond, no API server load
- The cache is always eventually consistent with the cluster (informer receives watch events)

Each controller (or application manager) registers its own Event Handler (callback function) to that shared informer. However, they do not get their own physical local copies of the resources. Instead, they all share one single memory map in RAM, and the informer passes them memory pointers to that exact same data.

```
┌────────────────────────────────────────┐
                  │  Kubernetes API Server (Control Plane)  │
                  └───────────▲────────────────▲───────────
                              │                │
         HTTP List-Watch #1   │                │   HTTP List-Watch #2
         (Network Stream)     │                │   (Network Stream)
                              ▼                ▼
 ╔═══════════════════════════════════╗  ╔═══════════════════════════════════╗
 ║ APPLICATION 1 (Pod 1 Memory)      ║  ║ APPLICATION 2 (Pod 2 Memory)      ║
 ║                                   ║  ║                                   ║
 ║ ┌───────────────────────────────┐ ║  ║ ┌───────────────────────────────┐ ║
 ║ │    controller-runtime CACHE   │ ║  ║ │    controller-runtime CACHE   │ ║
 ║ │ ┌───────────────────────────┐ │ ║  ║ │ ┌───────────────────────────┐ │ ║
 ║ │ │    Shared Pod Informer    │ │ ║  ║ │ │    Shared Pod Informer    │ │ ║
 ║ │ │ (Global RAM Map 1)        │ │ ║  ║ │ │ (Global RAM Map 2)        │ │ ║
 ║ │ └─────────────┬─────────────┘ │ ║  ║ │ └─────────────┬─────────────┘ │ ║
 ║ └───────────────┼───────────────┘ ║  ║ └───────────────┼───────────────┘ ║
 ║                 │                 ║  ║                 │                 ║
 ║                 ▼                 ║  ║                 ▼                 ║
 ║      [ Predicate Filter A ]       ║  ║      [ Predicate Filter B ]       ║
 ║                 │                 ║  ║                 │                 ║
 ║                 ▼                 ║  ║                 ▼                 ║
 ║           Workqueue A             ║  ║           Workqueue B             ║
 ║                 │                 ║  ║                 │                 ║
 ║                 ▼                 ║  ║                 ▼                 ║
 ║       ┌───────────────────┐       ║  ║       ┌───────────────────┐       ║
 ║       │   Controller A    │       ║  ║       │   Controller B    │       ║
 ║       │  Reconcile Loop   │       ║  ║       │  Reconcile Loop   │       ║
 ║       └───────────────────┘       ║  ║       └───────────────────┘       ║
 ╚═══════════════════════════════════╝  ╚═══════════════════════════════════╝
```

**Note:** The filtering logic sits on each controller's registration block, but it is executed by the Shared Pod Informer thread right before it attempts to hand the event over to that specific controller.

In Go architecture terms, the gates belong conceptually to the controllers, but they physically run inside the context of the Shared Informer's event loop.

---

### Q29. How does controller-runtime implement rate limiting on the work queue? How do you tune it?

In controller-runtime, rate limiting is not handled directly by the controller loops themselves, but rather at the entry point of the Workqueue.

When a reconciliation fails and your Reconcile function returns an error or requests a retry via `reconcile.Result{Requeue: true}`, the item isn't thrown back into the queue immediately. Instead, it passes through a rate-limiting subsystem that calculates an exponential backoff penalty to prevent your controller from entering a high-CPU crash loop.

**How it is Implemented: The Token Bucket & Exponential Backoff**

Under the hood, controller-runtime leverages `client-go/util/workqueue`. By default, it constructs a `MaxOfRateLimiter`, which combines two separate rate-limiting algorithms to protect your cluster:

- **A: The Item Exponential Failure Rate Limiter**
- **B: The Global Bucket Rate Limiter (Token Bucket)** — This acts as a global speed limit for the entire queue, protecting the Kubernetes API server from a thundering herd. It uses a token bucket algorithm with a default configuration of 100 tokens and a refill rate of 10 tokens per second.

Each failing object consumes one token to get requeued. If your controller experiences a massive cascading failure across hundreds of resources simultaneously, the bucket quickly empties. Once empty, requeue requests are choked down to a maximum of 10 items per second regardless of how fast individual item exponential backoffs expire.

---

### Q30. What is an Informer in client-go? How does it work internally?

An Informer is a client-go component that efficiently keeps a local cache of Kubernetes objects in sync with the API server. It is the correct pattern for controllers and any component that needs to watch resources.

**Internally it works in two phases:**
- **Phase 1 — List:** call the API server with List to get all current objects and their `resourceVersion`.
- **Phase 2 — Watch:** start a watch from that `resourceVersion`. Any event with a later version arrives via the watch stream. Reconnect automatically if the watch breaks.

The ListWatcher feeds into a Delta FIFO queue — a queue that holds the sequence of changes (Add, Update, Delete) for each object. A Store (thread-safe in-memory map) is maintained and updated by processing the FIFO queue.

You register event handlers on the informer that are called for Add/Update/Delete.

---

### Q31. What is a SharedInformerFactory and why should you use it instead of creating individual informers?

A SharedInformerFactory creates and shares informers across multiple consumers. If three different parts of your code all need to watch Pods, the factory creates only ONE pod informer — one watch connection, one cache — and all three consumers share it.

Without it, each component would create its own informer — three separate watch connections to the API server, three separate caches consuming memory, triple the load on the API server.

**Full working example:**

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 1. Initialize your rest config and clientset (Assumes standard local kubeconfig setup)
	// For local development, change this path to your actual ~/.kube/config if needed.
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// 2. Set up the cancellation channel (stopCh) to manage internal background threads
	stopCh := make(chan struct{})
	defer close(stopCh)

	// 3. Creates shared informers with 30s resync period (THE FACTORY BOX)
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)

	// These all share one Pod informer and one Deployment informer internally
	podLister := factory.Core().V1().Pods().Lister()
	_ = factory.Apps().V1().Deployments().Lister() // Registered but omitted for example brevity

	// 4. Define your event handlers
	myHandler1 := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Handler 1] Pod Added: %s/%s\n", pod.Namespace, pod.Name)
		},
	}

	myHandler2 := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("[Handler 2] Double verification log for pod: %s\n", pod.Name)
		},
	}

	// Register handlers BEFORE starting (Both get events from the single Informer box!)
	factory.Core().V1().Pods().Informer().AddEventHandler(myHandler1)
	factory.Core().V1().Pods().Informer().AddEventHandler(myHandler2)

	// 5. Start informers (Fires up background HTTP List-Watch loops)
	fmt.Println("Starting Shared Informer Factory streams...")
	factory.Start(stopCh)

	// 6. Block until cache is perfectly synchronized with the cluster API server
	fmt.Println("Synchronizing local global RAM map cache...")
	syncedResults := factory.WaitForCacheSync(stopCh)
	for informerType, synced := range syncedResults {
		if !synced {
			log.Fatalf("Error: Cache sync failed for informer resource type: %v", informerType)
		}
	}
	fmt.Println("Cache sync complete! Mirror database populated.")

	// 7. Read directly from the application's RAM cache (not the API server) — O(1) lookup
	pod, err := podLister.Pods("default").Get("my-pod")
	if err != nil {
		// If the pod doesn't exist yet, it safely returns an error from RAM without making a network call
		fmt.Printf("Cache lookup finished: %v (Note: Create 'my-pod' in the default namespace to see it return data)\n", err)
	} else {
		fmt.Printf("Successfully read pod status from RAM! UID: %s\n", pod.UID)
	}

	// Keep the application process awake briefly to catch arriving cluster events
	fmt.Println("Listening for cluster updates for 10 seconds...")
	time.Sleep(10 * time.Second)
}
```

The `30*time.Second` resync period causes the informer to re-list all objects periodically and call UpdateFunc even if nothing changed — ensures controllers re-process all objects periodically and do not miss anything due to bugs or external changes.

---

### Q32. How do you handle resourceVersion and conflict errors when updating objects with client-go?

ResourceVersion is a string stored in every object's metadata that the API server increments on every write. It is used for optimistic concurrency control — the equivalent of a row version in a database.

When you Update an object, you must include the current resourceVersion. If another writer modified the object between when you read it and when you write it, the API server returns a 409 Conflict error. You must re-read and retry.

`retry.RetryOnConflict` retries automatically with exponential backoff on 409 errors. This is the standard pattern in client-go.

**Important:** after an Update succeeds, your local object has a new resourceVersion. Never reuse an old object reference without re-reading.

---

*End of guide. Read top to bottom the night before — Level 1 is recall/definitions, Level 2 builds on CRD and controller-runtime basics, Level 3 covers triggering mechanics and design tradeoffs, and Level 4 is internals/production tuning — the depth Staff/Principal rounds probe for.*



=========================================================================================================
                                     KUBERNETES CONTROL PLANE
=========================================================================================================
 ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │                                   Kubernetes API Server (etcd)                                      │
 └──────────────────────────────────────────────────┬──────────────────────────────────────────────────┘
                                                    │
                                                    │  1. ONE Persistent HTTP/2 Stream
                                                    ▼
=========================================================================================================
                                    YOUR APPLICATION PROCESS (RAM)
=========================================================================================================
 ╔═════════════════════════════════════════════════════════════════════════════════════════════════════╗
 ║  SharedInformerFactory (The Outer Lifecycle Manager)                                                ║
 ║                                                                                                     ║
 ║  ┌───────────────────────────────────────────────────────────────────────────────────────────────┐ ║
 ║  │  Shared Pod Informer (The Resource-Specific Container Box)                                   │ ║
 ║  │                                                                                               │ ║
 ║  │  [Reflector Component] ◄──────────────────────────────────────────────────────────────────┐   │ ║
 ║  │  - Performs the initial HTTP GET "List" of all objects.                                   │   │ ║
 ║  │  - Establishes long-running HTTP "Watch" chunks.                                          │   │ ║
 ║  │          │                                                                                │   │ ║
 ║  │          │ 2. Emits transactional deltas (e.g., "ADD", "UPDATE")                           │   │ ║
 ║  │          ▼                                                                                │   │ ║
 ║  │  [DeltaFIFO Queue]                                                                        │   │ ║
 ║  │  - A sequential transaction log buffer within the informer.                               │   │ ║
 ║  │          │                                                                                │   │ ║
 ║  │          │ 3. Pops state changes sequentially                                             │   │ ║
 ║  │          ▼                                                                                │   │ ║
 ║  │  [The Informer Controller Engine] ──────────────────────────────────────────────┐         │   │ ║
 ║  │          │                                                                      │         │   │ ║
 ║  │          │ 4a. Writes state change to cache                                     │ 4b.     │   │ ║
 ║  │          ▼                                                                      │ Broadcasts    ║
 ║  │  [Cache Subsystem: Indexer Store]                                               │ pointer │   │ ║
 ║  │  - The actual Local Global RAM Map.                                              │ down    │   │ ║
 ║  │  - Key: "namespace/name" ──► Value: [Literal Data Object Structure]             │         │   │ ║
 ║  └─────────────────────────────────────────────────────────────────────────────────┼─────────┼───┘ ║
 ║                                                                                    │         │     ║
 ║                                                  ┌─────────────────────────────────┘         │     ║
 ║                                                  ▼                                           │     ║
 ║                                     [Resource Event Handlers]                                │     ║
 ║                                     - Runs AddEventHandler custom functions                  │     ║
 ║                                                  │                                           │     ║
 ║                                                  │ 5. Enqueues 8-byte string key only        │     ║
 ║                                                  ▼                                           │     ║
 ║                                     [Workqueue / Rate Limiter]                               │     ║
 ║                                     - Keeps keys: e.g., []string{"default/my-pod"}           │     ║
 ║                                                  │                                           │     ║
 ║                                                  │ 6. Triggers processing loop               │     ║
 ║                                                  ▼                                           │     ║
 ║                                     [Your Custom Reconcile Loop]                             │     ║
 ║                                     - Recovers the raw string key.                           │     ║
 ║                                     - Directly calls the Indexer Store to fetch data ────────┘     ║
 ║                                       without crossing the network network.                  ║
 ╚═════════════════════════════════════════════════════════════════════════════════════════════════════╝


 # Kubernetes Operators — Part 2
### RBAC for Operators & Admission Webhooks (Beginner → Expert)

Companion file to `custom-controller-organized.md`. That file covers Operator/Controller concepts, CRDs, controller-runtime, and client-go. This file fills two specific gaps: **RBAC for Operators** and **Admission Webhooks**.

---

## Table of Contents

- [Part A — RBAC for Operators](#part-a--rbac-for-operators)
- [Part B — Admission Webhooks](#part-b--admission-webhooks)

---

## Part A — RBAC for Operators

---

### Q1. Why does an Operator need RBAC at all? (Beginner)

- The Operator runs as a Pod, and that Pod authenticates to the API server using a **ServiceAccount**.
- By default, a ServiceAccount has almost no permissions — it cannot list, watch, create, or update anything.
- An Operator's entire job is to watch and modify resources (the CR it manages, plus everything it creates — Pods, Services, ConfigMaps, etc.). Without RBAC granting those permissions, every API call the Operator makes fails with `403 Forbidden`.
- **In short:** RBAC is what turns "the Operator's code knows what to do" into "the Operator is actually allowed to do it."

---

### Q2. What are the four core RBAC objects, and how do they fit together? (Beginner)

```
ServiceAccount        — the Operator's identity (who is asking)
        │
        ▼
Role / ClusterRole     — the permissions (what actions are allowed, on what resources)
        │
        ▼
RoleBinding / ClusterRoleBinding  — links the identity to the permissions
```

- **Role** — namespace-scoped permission set (e.g., "can get/list/watch Pods in namespace `default`")
- **ClusterRole** — cluster-scoped permission set (works across all namespaces, or for cluster-scoped resources like Nodes, StorageClasses)
- **RoleBinding** — grants a Role to a subject (ServiceAccount/User/Group) within one namespace
- **ClusterRoleBinding** — grants a ClusterRole to a subject across the entire cluster

**Rule of thumb:** if your Operator only manages resources within namespaces it's told to watch, use Role + RoleBinding. If your Operator manages cluster-scoped resources (or needs to watch all namespaces), you need ClusterRole + ClusterRoleBinding.

---

### Q3. What does a minimal RBAC manifest for an Operator actually look like? (Intermediate)

```yaml
# 1. Identity
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-operator
  namespace: my-operator-system

---
# 2. Permissions (cluster-scoped because CRDs are cluster-scoped resources)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-operator-role
rules:
  # Permission on the Custom Resource itself
  - apiGroups: ["myapp.example.com"]
    resources: ["databases"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

  # Permission on the status subresource (separate from the main resource!)
  - apiGroups: ["myapp.example.com"]
    resources: ["databases/status"]
    verbs: ["get", "update", "patch"]

  # Permission on secondary resources the Operator creates
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

  # Permission to record Kubernetes Events (for visibility/debugging)
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]

---
# 3. Binding — connects the ServiceAccount to the ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-operator-rolebinding
subjects:
  - kind: ServiceAccount
    name: my-operator
    namespace: my-operator-system
roleRef:
  kind: ClusterRole
  name: my-operator-role
  apiGroup: rbac.authorization.k8s.io
```

- Notice `databases` and `databases/status` are listed as **separate resources** in the RBAC rules — this mirrors the subresource separation discussed in the main file (spec vs status). Granting access to `databases` does **not** automatically grant access to `databases/status`.

---

### Q4. What is `kubebuilder:rbac` and how does it generate these manifests automatically? (Intermediate)

- Writing RBAC YAML by hand for every resource type is tedious and error-prone (easy to forget a verb or a subresource).
- Kubebuilder/controller-runtime projects use special **marker comments** directly above the Reconciler code. A code-generation tool (`controller-gen`) scans these markers and generates the ClusterRole YAML automatically.

```go
// +kubebuilder:rbac:groups=myapp.example.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myapp.example.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods;services;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...
}
```

- Run `make manifests` (or `controller-gen rbac:roleName=manager-role paths="./..."`) and it produces the full `role.yaml` for you.
- **Why this matters in an interview:** it shows you understand RBAC isn't just static YAML — it's tightly coupled to what the reconciler code actually touches, and good Operator projects keep that coupling explicit and automated rather than manually maintained.

---

### Q5. What is the Principle of Least Privilege, and how do you apply it to an Operator's RBAC? (Advanced)

- Never grant `verbs: ["*"]` or `resources: ["*"]` "just to be safe" — this is the single most common RBAC mistake in real Operator projects, and it's exactly what an interviewer will probe for.
- Grant only the verbs the reconciler logic actually performs:
  - If the Operator only ever reads a resource (e.g., watching ConfigMaps for configuration, never writing them), grant only `get, list, watch` — not `create, update, delete`.
  - If the Operator never deletes the CR itself (Kubernetes garbage collection handles that via OwnerReferences), you may not need `delete` on owned resources at all.
- Use **namespace-scoped** Role/RoleBinding instead of ClusterRole/ClusterRoleBinding wherever possible — an Operator that only manages resources in namespaces it's explicitly configured to watch should not have permissions across the whole cluster.
- Split RBAC into multiple smaller roles when an Operator has genuinely different responsibilities (e.g., a "reader" role for a metrics-exporting sidecar vs a "writer" role for the main reconciler), rather than one giant role granting everything to every component.

---

### Q6. Your Operator's reconciler works fine in your local dev cluster (using your personal kubeconfig) but fails with `403 Forbidden` once deployed to the real cluster. Why, and how do you debug it? (Expert)

- **Why it happens:** locally, your code authenticates as *you* — typically a cluster-admin user via your kubeconfig, which has unrestricted permissions. Once deployed, the Pod authenticates as its **ServiceAccount**, which only has whatever RBAC was explicitly granted to it. It's extremely common for a permission to work "by accident" locally and then break in the real deployment because the RBAC manifest is missing a verb, a resource, or — most commonly — the `/status` subresource.

**Debugging steps:**
1. Read the exact error — it tells you precisely what was denied:
   ```
   pods "my-pod" is forbidden: User "system:serviceaccount:my-operator-system:my-operator"
   cannot update resource "pods" in API group "" in the namespace "default"
   ```
   This single line gives you the ServiceAccount, the verb, the resource, the API group, and the namespace — everything needed to fix the Role.

2. Use `kubectl auth can-i` impersonating the ServiceAccount directly, rather than guessing:
   ```bash
   kubectl auth can-i update pods \
     --as=system:serviceaccount:my-operator-system:my-operator \
     -n default
   ```

3. Check whether the missing permission is on the **subresource**, not the main resource — this is the single most common cause of "it works for spec but status updates fail":
   ```bash
   kubectl auth can-i update databases/status \
     --as=system:serviceaccount:my-operator-system:my-operator
   ```

4. If using kubebuilder markers, confirm `make manifests` was actually re-run after adding new logic — a very common mistake is adding code that touches a new resource type without regenerating (and redeploying) the RBAC YAML.

5. Check that the RoleBinding/ClusterRoleBinding's `subjects` block has the correct ServiceAccount **name and namespace** — a binding pointing to the wrong namespace silently grants permission to nothing, with no error until something tries to use it.

---

## Part B — Admission Webhooks

---

### Q7. What is an admission webhook, and where does it sit in the request lifecycle? (Beginner)

When you run `kubectl apply -f my-resource.yaml`, the request doesn't go straight from the API server into etcd. It passes through several stages:

```
kubectl apply
     │
     ▼
Authentication  →  Authorization (RBAC)  →  Admission Control  →  etcd (persisted)
                                                   │
                                    ┌──────────────┴──────────────┐
                                    ▼                             ▼
                          Mutating Webhooks              Validating Webhooks
                          (can MODIFY the object)         (can only ACCEPT/REJECT)
```

- **Admission webhooks** are HTTP callbacks the API server calls *after* RBAC has already approved the request, but *before* the object is written to etcd.
- They let you inject custom logic into the request pipeline — defaulting values, enforcing policies, or rejecting invalid objects — without modifying Kubernetes itself.

---

### Q8. What is the difference between a Mutating and a Validating admission webhook? (Beginner)

| | Mutating Webhook | Validating Webhook |
|---|---|---|
| **Can modify the object?** | Yes — returns a JSON patch | No — can only allow or deny |
| **Runs when?** | Always before validating webhooks | Always after mutating webhooks |
| **Typical use** | Setting defaults, injecting sidecars, normalizing fields | Enforcing business rules, rejecting invalid configurations |
| **Example** | "If `replicas` is unset, default it to 1" | "Reject if `replicas > 10`" |

- **Order matters:** mutating webhooks run first so that validating webhooks see the *final*, fully-defaulted object — not the raw, possibly-incomplete one the user submitted.

---

### Q9. In a kubebuilder/controller-runtime project, how do you actually implement these two webhook types in Go? (Intermediate)

```go
// Mutating webhook — implements webhook.Defaulter
func (d *Database) Default() {
    if d.Spec.Replicas == 0 {
        d.Spec.Replicas = 1
    }
    if d.Spec.StorageClass == "" {
        d.Spec.StorageClass = "standard"
    }
}

// Validating webhook — implements webhook.Validator
func (d *Database) ValidateCreate() (admission.Warnings, error) {
    return d.validate()
}

func (d *Database) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    oldDB := old.(*Database)
    // Example business rule: storage can only grow, never shrink
    if d.Spec.StorageSize < oldDB.Spec.StorageSize {
        return nil, fmt.Errorf("storage size cannot be decreased (was %s, got %s)",
            oldDB.Spec.StorageSize, d.Spec.StorageSize)
    }
    return d.validate()
}

func (d *Database) ValidateDelete() (admission.Warnings, error) {
    return nil, nil // usually no validation needed on delete
}

func (d *Database) validate() (admission.Warnings, error) {
    if d.Spec.Replicas > 10 {
        return nil, field.Invalid(
            field.NewPath("spec").Child("replicas"),
            d.Spec.Replicas,
            "replicas cannot exceed 10",
        )
    }
    return nil, nil
}
```

```go
// Register both with the manager
func (d *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).
        For(d).
        Complete()
}
```

---

### Q10. What infrastructure does an admission webhook actually require to run in a real cluster? (Advanced)

A webhook is not just Go code — it requires several pieces wired together correctly:

1. **TLS certificates** — the API server only calls webhooks over HTTPS, and it validates the webhook's certificate against a CA bundle. Most production setups use **cert-manager** to automatically issue and rotate this certificate rather than managing it by hand.

2. **A Service** fronting the webhook Pod (the same Pod running your Operator's manager, typically on a separate port like `:9443`).

3. **A `MutatingWebhookConfiguration` / `ValidatingWebhookConfiguration` object** registered in the cluster, telling the API server:
   - Which resource types/operations to intercept (`rules`)
   - Which Service+path to call (`clientConfig`)
   - The CA bundle to trust (`caBundle`)
   - The **failure policy** — see Q11

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: database-validating-webhook
webhooks:
  - name: vdatabase.myapp.example.com
    clientConfig:
      service:
        name: my-operator-webhook-service
        namespace: my-operator-system
        path: /validate-myapp-example-com-v1-database
      caBundle: <base64-encoded-ca-cert>   # usually injected by cert-manager
    rules:
      - apiGroups: ["myapp.example.com"]
        apiVersions: ["v1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["databases"]
    sideEffects: None
    admissionReviewVersions: ["v1"]
    failurePolicy: Fail
```

---

### Q11. What is `failurePolicy`, and why is choosing the wrong value a production-outage risk? (Expert)

- `failurePolicy` decides what the API server does if the webhook **cannot be reached at all** (Pod crashed, network partition, certificate expired, etc.). It has two values:

```
failurePolicy: Fail    →  if the webhook is unreachable, REJECT the request
failurePolicy: Ignore   →  if the webhook is unreachable, ALLOW the request through unchecked
```

- **The outage scenario this causes:** if your webhook Pod is the *same* Pod as your Operator's manager (very common setup) and that Pod crashes or is mid-rollout, then with `failurePolicy: Fail`, **every single create/update of that resource type across the entire cluster starts failing** — including unrelated resources that have nothing to do with your Operator, if your webhook's `rules` are scoped too broadly (e.g., accidentally intercepting all Pods instead of just your CRD).

- This is a real, well-known failure mode: a webhook with `failurePolicy: Fail` and broad `rules` scope, combined with a webhook Pod outage, can make an entire cluster unable to schedule new Pods.

**Mitigations an interviewer wants to hear:**
- Scope `rules` as narrowly as possible — only the exact `apiGroups`/`resources`/`operations` you actually need to intercept.
- Run the webhook with multiple replicas and a PodDisruptionBudget so a single Pod crash/rollout doesn't take down all webhook capacity.
- Set a reasonable `timeoutSeconds` (default 10s, often reduced) so a slow webhook fails fast rather than stalling every request in the cluster.
- For non-critical validation (nice-to-have checks, not safety-critical ones), prefer `failurePolicy: Ignore` so a webhook outage degrades gracefully instead of taking down unrelated cluster operations.
- For safety-critical validation (e.g., preventing data-loss-causing updates), `Fail` is correct — but only if you've invested in webhook high availability to match that requirement.

---

### Q12. A validating webhook deployed last week is now silently failing to reject objects it used to reject — `kubectl apply` succeeds when it shouldn't. The webhook Pod is healthy and logs show it's receiving no requests at all for this resource type. What's the most likely cause, and how do you confirm it? (Expert)

- **Most likely cause:** the `ValidatingWebhookConfiguration`'s certificate (in `caBundle`) has expired, or the underlying TLS cert served by the webhook Pod has rotated (e.g., cert-manager renewed it) but the `caBundle` field in the webhook configuration was never updated to match — so the API server's TLS handshake to the webhook silently fails, and because `failurePolicy: Ignore` was set, the API server just lets requests through without ever successfully calling your webhook.

**How to confirm:**
1. Check API server logs (or `kube-apiserver` audit logs if enabled) for TLS handshake errors when calling the webhook's Service — they typically show up as connection/certificate errors, not as your webhook's own application logs (which is exactly why your webhook Pod shows zero incoming requests — the call never successfully reaches it).
2. Compare the `caBundle` in the `ValidatingWebhookConfiguration` against the actual CA currently signing the webhook's serving certificate (`kubectl get secret <webhook-cert-secret> -o yaml` and decode it) — a mismatch confirms the theory.
3. If using cert-manager, check whether it has an `Injector` annotation (`cert-manager.io/inject-ca-from`) on the webhook configuration — this is the mechanism that's supposed to keep `caBundle` automatically in sync with certificate rotations, and if that annotation is missing or misconfigured, the `caBundle` goes stale every time the cert rotates.
4. Confirm `failurePolicy` — if it's `Ignore`, this exact silent-failure mode is expected behavior by design, which is precisely why `Ignore` is risky for anything safety-critical (ties back to Q11).

---

*End of Part 2. Pair with `custom-controller-organized.md` for full Operator interview coverage — concepts/internals there, permissions and admission control here.*