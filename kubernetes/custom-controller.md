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

```
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: storagevolumes.storage.ibm.com

spec:
  group: storage.ibm.com

  names:
    kind: StorageVolume
    plural: storagevolumes
    singular: storagevolume
    shortNames:
      - sv

  scope: Namespaced

  versions:
    - name: v1alpha1
      served: true
      storage: true

      schema:
        openAPIV3Schema:
          type: object

          properties:
            spec:
              type: object
              properties:
                size:
                  type: integer
                pool:
                  type: string
                encrypted:
                  type: boolean

            status:
              type: object
              properties:
                phase:
                  type: string
                volumeID:
                  type: string

      subresources:
        status: {}

      additionalPrinterColumns:
        - name: Size
          type: integer
          jsonPath: .spec.size

        - name: Pool
          type: string
          jsonPath: .spec.pool

        - name: Phase
          type: string
          jsonPath: .status.phase

        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
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

Owner References are one of the most important concepts when writing Kubernetes Operators. They allow Kubernetes to automatically clean up child resources when the parent is deleted
Owner Reference is a field on a Kubernetes object that points to another object that owns it. When the owner is deleted, Kubernetes garbage collects all owned objects automatically — this is called cascading deletion.
The OwnerReference ensures Kubernetes-native child resources (PVCs, Jobs, Secrets, etc.) are automatically garbage-collected.

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

In controller-runtime, the primary resource is registered using For(), while secondary resources are registered using Owns() or Watches().

Owns() is used when the secondary resource has an OwnerReference pointing to the primary resource. When the secondary resource changes, controller-runtime follows the OwnerReference, determines the owning resource, and enqueues a reconcile request for the parent. This is the most common pattern for resources like Pods, Jobs, Deployments, or PVCs created by an Operator.

If the secondary resource isn't owned by the primary resource, Watches() is used with a mapping function (EnqueueRequestsFromMapFunc) to explicitly determine which primary resources should be reconciled when that secondary resource changes.

```go
// Own
func (r *ZonalImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&imagev1beta1.ZonalImage{}).
        Owns(&batchv1.Job{}).
        Complete(r)
}

// Watch
ctrl.NewControllerManagedBy(mgr).
    For(&StorageVolume{}).
    Watches(
        &source.Kind{Type: &corev1.ConfigMap{}},
        handler.EnqueueRequestsFromMapFunc(...),
    )

// Watch and Own
func (r *ZonalImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagev1beta1.ZonalImage{}).      // Primary resource
		Owns(&batchv1.Job{}).                 // Owned secondary resource
		Owns(&corev1.ConfigMap{}).            // Another owned resource
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapSecretToZonalImages),
		).                                    // Custom watch
		Complete(r)
}

```
---

### Q26. What are Predicates in controller-runtime and how do you use them to reduce unnecessary reconciliations?

Predicates are filters that sit between the watch event and the work queue. If a predicate returns false, the event is dropped and the reconciler is never called. This reduces unnecessary reconciliations and saves CPU.

Predicates are event filters in controller-runtime that determine whether an event should enqueue a reconcile request. They help reduce unnecessary reconciliations by filtering out events the controller doesn't care about, such as status-only updates or metadata changes.

**Built-in predicates:**
- `GenerationChangedPredicate` — only spec changes, not status updates
- `ResourceVersionChangedPredicate` — any change
- `LabelChangedPredicate` — label changes only
- `AnnotationChangedPredicate` — annotation changes only

Custom predicates implement the `Predicate` interface.

```go
type SizeChangedPredicate struct {
	predicate.Funcs
}

func (SizeChangedPredicate) Update(e event.UpdateEvent) bool {

	oldObj := e.ObjectOld.(*storagev1.StorageVolume)
	newObj := e.ObjectNew.(*storagev1.StorageVolume)

	return oldObj.Spec.Size != newObj.Spec.Size
}

func (r *StorageVolumeReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&storagev1.StorageVolume{},
			builder.WithPredicates(SizeChangedPredicate{}),
		).
		Complete(r)
}
```

---

### Q27. What is the dynamic client in client-go and when do you use it over the typed clientset?

The dynamic client in client-go is not limited to Custom Resources. It can interact with both built-in Kubernetes resources and CRDs. The key difference is that it operates on unstructured.Unstructured objects rather than generated Go types. This makes it ideal for generic tools, plugins, or controllers that need to work with arbitrary Kubernetes resources without having compile-time knowledge of their schemas. For strongly typed controllers that manage known resources, the typed client or controller-runtime client is generally preferred because it provides compile-time type safety and a better developer experience.

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

```
+--------------------------------------------------------------------------------+
|                      controller-runtime Manager                                |
|                                                                                |
|  +------------------------------------------------------------------------+    |
|  |                    SharedInformerFactory (client-go)                   |    |
|  |                                                                        |    |
|  |   +-------------------+      +--------------------+                    |    |
|  |   | SharedInformer    |      | SharedInformer     |                    |    |
|  |   | (StorageVolume)   |      | (ImageArchive)     |                    |    |
|  |   +-------------------+      +--------------------+                    |    |
|  |            |                           |                               |    |
|  +------------|---------------------------|-------------------------------+    |
|               |                           |                                    |
|               ▼                           ▼                                    |
|      +------------------+       +------------------+                           |
|      | Local Cache      |       | Local Cache      |                           |
|      | (Indexer/Store)  |       | (Indexer/Store)  |                           |
|      +------------------+       +------------------+                           |
|               |                           |                                    |
|               +------------+--------------+                                    |
|                            |                                                   |
|                            ▼                                                   |
|                  controller-runtime Cache                                      |
|                            |                                                   |
|      +---------------------+-------------------------+                         |
|      |                                               |                         |
|      ▼                                               ▼                         |
| Controller A                                 Controller B                      |
|      |                                               |                         |
|      ▼                                               ▼                         |
| EventHandler                                  EventHandler                     |
|      |                                               |                         |
|      ▼                                               ▼                         |
| WorkQueue                                     WorkQueue                        |
|      |                                               |                         |
|      ▼                                               ▼                         |
| Reconcile()                                   Reconcile()                      |
|      |                                               |                         |
|      +-------------------+-----------------------+                             |
|                          |                                                     |
|                          ▼                                                     |
|                  client.Get()/List()                                           |
|                          |                                                     |
|                          ▼                                                     |
|             Reads from controller-runtime Cache                                |
+--------------------------------------------------------------------------------+
```

```
// This is shared informer and cache 
                    Manager
                       │
             controller-runtime Cache
                       │
      +----------------+----------------+
      |                                 |
      ▼                                 ▼
SharedInformer                  SharedInformer
(StorageVolume)                 (ImageArchive)
      │                                 │
      ▼                                 ▼
 Reflector                         Reflector
      │                                 │
 LIST/WATCH                       LIST/WATCH
      │                                 │
      ▼                                 ▼
 DeltaFIFO                        DeltaFIFO
      │                                 │
      ▼                                 ▼
 Indexer                          Indexer
(Local Cache)                   (Local Cache)
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


```
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
```

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










Step 1 — Define the CRD
Step 2 — Create a Custom Resource
Step 3 — Define Go API types. //pkg/apis/storage/v1/types.go
Step 4 — Register the type with a Scheme
```
storage.example.com/v1
        +
VolumeBackup
        ↓
Go type: VolumeBackup
```
Step 5 — DeepCopy //They are generated
Step 6 — Create the client //Generated code would provide methods roughly like
```
VolumeBackups(namespace).Get(...)
VolumeBackups(namespace).List(...)
VolumeBackups(namespace).Create(...)
VolumeBackups(namespace).Update(...)
```
Step 7 — Create an Informer
Step 8 — Work Queue
Step 9 — Controller
Step 10. Worker
Step 10 — Reconcile
Step 11. Let's put actual business logic
Where does RBAC fit?



This is the part I'd emphasize in your interview.

                     YOU WRITE
                        │
        ┌───────────────┼─────────────────┐
        │               │                 │
        ▼               ▼                 ▼
      CRD YAML       Go API Types      Controller
        │               │                 │
        │               │                 │
        ▼               ▼                 ▼
    Schema          Spec/Status       Reconcile()
                                        │
                                        ▼
                                  Business Logic


                   GENERATED / PROVIDED
                        │
        ┌───────────────┼─────────────────┐
        │               │                 │
        ▼               ▼                 ▼
    DeepCopy       Typed Client       Informer
                                        │
                                        ▼
                                   LIST/WATCH
                                   Local Cache



                Controller
                    │
          ┌─────────┴─────────┐
          │                   │
      client-go          controller-runtime
          │                   │
          ▼                   ▼
   Informer/Queue/etc.   Manager/Controller
                              │
                              ▼
                         Reconcile()








# Introducing a New CRD + Controller in Kubernetes
### Complete Flow — From Definition to `kubectl` Management
### Interview-Oriented: Every Change, Every File, Every Reason

---

## The Big Picture — What You Are Building

```
┌─────────────────────────────────────────────────────────────────────┐
│                     What you need to create                         │
│                                                                     │
│  1. API Types (Go structs)   → defines what the CRD looks like     │
│  2. CRD manifest (YAML)      → registers the CRD in Kubernetes     │
│  3. Controller (Go)          → watches and reconciles the CRD      │
│  4. RBAC (YAML)              → grants controller permissions        │
│  5. Deployment (YAML)        → runs the controller as a pod        │
└─────────────────────────────────────────────────────────────────────┘

How they connect:

 User runs:                 Controller watches:          Controller manages:
 kubectl apply              API Server notifies          Secondary resources
 -f myapp.yaml              controller via informer      (Pods, Services,
      │                          │                        ConfigMaps...)
      ▼                          ▼                             │
 [MyApp CR] ──────────► [Reconcile Loop] ─────────────────────┘
  stored in etcd           reads CR spec
                           compares to current state
                           creates/updates/deletes resources
```

---

## The Example We Will Use Throughout

We are building a `WebApp` CRD. When a user creates a `WebApp` resource,
the controller automatically:
- Creates a `Deployment` with the specified number of replicas and image
- Creates a `Service` to expose it
- Updates `WebApp.Status` with the current state

```yaml
# What the user writes:
apiVersion: apps.example.com/v1
kind: WebApp
metadata:
  name: my-webapp
spec:
  replicas: 3
  image: nginx:1.25
  port: 8080

# What the controller does automatically:
# → Creates Deployment: my-webapp (3 replicas, nginx:1.25)
# → Creates Service: my-webapp (port 8080)
# → Updates status.availableReplicas, status.phase
```

---

## Step 1 — API Types (Go Structs)

### File: `api/v1/webapp_types.go`

### What it is

The Go structs that define the schema of your CRD. This is the
**source of truth** for what fields the CRD accepts, what validation
applies, and what the status looks like.

### The code

```go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebAppSpec — desired state declared by the user (the "spec")
// +kubebuilder:validation:Optional
type WebAppSpec struct {
    // Replicas is the number of pod replicas to run
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=100
    Replicas int32 `json:"replicas"`

    // Image is the container image to run
    // +kubebuilder:validation:Required
    Image string `json:"image"`

    // Port is the container port to expose
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=65535
    Port int32 `json:"port,omitempty"`
}

// WebAppStatus — current observed state written by the controller (the "status")
type WebAppStatus struct {
    // Phase is the current lifecycle phase of the WebApp
    // +kubebuilder:validation:Enum=Pending;Running;Degraded;Failed
    Phase string `json:"phase,omitempty"`

    // AvailableReplicas is the number of pods currently available
    AvailableReplicas int32 `json:"availableReplicas,omitempty"`

    // Conditions is a list of current observed conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // ObservedGeneration is the CR generation the controller last reconciled
    // This tells users whether the controller has processed the latest spec
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// WebApp is the Schema for the webapps API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wa;webapp
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type WebApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   WebAppSpec   `json:"spec,omitempty"`
    Status WebAppStatus `json:"status,omitempty"`
}

// WebAppList contains a list of WebApp
// +kubebuilder:object:root=true
type WebAppList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []WebApp `json:"items"`
}

func init() {
    SchemeBuilder.Register(&WebApp{}, &WebAppList{})
}
```

### Why each part matters

**`TypeMeta` and `ObjectMeta`:**
Every Kubernetes object needs TypeMeta (`apiVersion`, `kind`) and
ObjectMeta (`name`, `namespace`, `labels`, `annotations`, `resourceVersion`).
These are embedded via composition — your CRD inherits all standard
Kubernetes object behaviour for free.

**`json:"..."` tags:**
Tell the JSON serializer how to map Go fields to JSON keys. Kubernetes
stores all objects as JSON in etcd. The tag `json:"replicas"` means
the user writes `spec.replicas:` in their YAML.

**`+kubebuilder:` marker comments:**
These are not just comments — the `controller-gen` tool reads them
to generate the CRD YAML and RBAC manifests automatically.

| Marker | What it generates |
|---|---|
| `+kubebuilder:object:root=true` | Registers this as a root CRD object |
| `+kubebuilder:subresource:status` | Enables `/status` subresource (separate RBAC for status updates) |
| `+kubebuilder:resource:shortName=wa` | Lets users run `kubectl get wa` instead of `kubectl get webapps` |
| `+kubebuilder:printcolumn:...` | Custom columns in `kubectl get webapps` output |
| `+kubebuilder:validation:Minimum=1` | Adds OpenAPI validation to the generated CRD schema |

**`ObservedGeneration` in status:**
This is a Kubernetes best practice. The controller always sets
`status.observedGeneration = metadata.generation`. Users can check
whether the controller has processed their latest spec change by
comparing these two values. If `observedGeneration < generation`,
the controller has not reconciled the latest spec yet.

**`metav1.Condition` in status:**
The standard Kubernetes condition type for communicating resource health.
Has fields: `Type`, `Status` (True/False/Unknown), `Reason`, `Message`,
`ObservedGeneration`, `LastTransitionTime`. Use this instead of custom
boolean fields — it is the Kubernetes API convention.

---

### File: `api/v1/groupversion_info.go`

```go
package v1

import (
    "k8s.io/apimachinery/pkg/runtime/schema"
    "sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
    // GroupVersion is the API group and version for this API
    GroupVersion = schema.GroupVersion{
        Group:   "apps.example.com",
        Version: "v1",
    }

    // SchemeBuilder adds our types to the runtime scheme
    SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

    // AddToScheme adds the types from this group/version to the given scheme
    AddToScheme = SchemeBuilder.AddToScheme
)
```

**Why this file exists:**
Kubernetes uses a runtime `Scheme` — a registry that maps
`(Group, Version, Kind)` strings to Go types. Without registering your
types, the API server and controller cannot serialize/deserialize your
objects. `GroupVersion` declares: "our CRD belongs to API group
`apps.example.com`, version `v1`." The `AddToScheme` function is called
at startup to register `WebApp` and `WebAppList` so the controller can
work with them.

---

## Step 2 — Generate the CRD YAML

### What changes and why

```bash
# Run controller-gen to generate CRD manifest from your Go types
make generate    # generates DeepCopyObject methods
make manifests   # generates CRD YAML and RBAC from markers
```

### File: `config/crd/bases/apps.example.com_webapps.yaml`

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: webapps.apps.example.com    # must be: plural.group
  annotations:
    controller-gen.kubebuilder.io/version: v0.15.0
spec:
  group: apps.example.com           # API group
  names:
    kind: WebApp                    # CamelCase — what you write in YAML
    listKind: WebAppList            # used for list responses
    plural: webapps                 # URL path: /apis/apps.example.com/v1/webapps
    singular: webapp                # used in kubectl output
    shortNames:
    - wa                            # kubectl get wa (from +kubebuilder:resource)
  scope: Namespaced                 # or Cluster
  versions:
  - name: v1
    served: true                    # this version is active
    storage: true                   # this version is stored in etcd

    # Status subresource — separates spec and status update permissions
    subresources:
      status: {}

    # Additional kubectl get columns (from +kubebuilder:printcolumn markers)
    additionalPrinterColumns:
    - name: Phase
      type: string
      jsonPath: .status.phase
    - name: Replicas
      type: integer
      jsonPath: .spec.replicas
    - name: Available
      type: integer
      jsonPath: .status.availableReplicas
    - name: Age
      type: date
      jsonPath: .metadata.creationTimestamp

    # OpenAPI v3 schema — generated from your Go structs and markers
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            required:
            - replicas
            - image
            properties:
              replicas:
                type: integer
                format: int32
                minimum: 1
                maximum: 100
              image:
                type: string
              port:
                type: integer
                format: int32
                minimum: 1
                maximum: 65535
          status:
            type: object
            properties:
              phase:
                type: string
                enum: [Pending, Running, Degraded, Failed]
              availableReplicas:
                type: integer
                format: int32
              observedGeneration:
                type: integer
                format: int64
              conditions:
                type: array
                items:
                  type: object
                  required: [type, status]
                  properties:
                    type:   { type: string }
                    status: { type: string }
                    reason: { type: string }
                    message: { type: string }
                    lastTransitionTime: { type: string, format: date-time }
                    observedGeneration: { type: integer }
```

### Why each field matters

**`name: webapps.apps.example.com`:**
The CRD's own Kubernetes name must follow the format `plural.group`.
This is how the API server looks up the CRD definition when it receives
a request for `apps.example.com/v1/webapps`.

**`scope: Namespaced`:**
Namespaced means `WebApp` resources belong to a namespace (like Pods,
Services). Cluster scope means they are global (like Nodes, StorageClasses).
Choose based on your use case — most application-level CRDs are Namespaced.

**`served: true` vs `storage: true`:**
`served` means the API server accepts requests for this version.
`storage` means objects are stored in etcd using this version's schema.
Only one version can be `storage: true`. Multiple versions can be `served: true`
(for backward compatibility during version migrations).

**`subresources: status: {}`:**
This creates a separate HTTP endpoint `/apis/apps.example.com/v1/namespaces/{ns}/webapps/{name}/status`.
It enables:
- Separate RBAC: you can grant a controller `update` on `webapps/status`
  without granting `update` on `webapps` (which could let it change spec)
- When a user does `kubectl apply`, only the spec is updated
- When the controller calls `Status().Update()`, only the status is updated
  — they do not overwrite each other

**`openAPIV3Schema`:**
The API server validates all create and update requests against this schema
before accepting them. Invalid objects are rejected with a clear error
message — before your controller ever sees them. This is much better than
discovering invalid input in your reconciler.

---

## Step 3 — The Controller

### File: `internal/controller/webapp_controller.go`

```go
package controller

import (
    "context"
    "fmt"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/util/intstr"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"

    myapiv1 "github.com/example/webapp-controller/api/v1"
)

const webappFinalizer = "apps.example.com/finalizer"

// WebAppReconciler reconciles WebApp objects
// RBAC markers — controller-gen generates ClusterRole from these
// +kubebuilder:rbac:groups=apps.example.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.example.com,resources=webapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.example.com,resources=webapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
type WebAppReconciler struct {
    client.Client   // embedded — gives us r.Get(), r.List(), r.Create(), etc.
    Scheme *runtime.Scheme
}

// Reconcile is the core function — called whenever a WebApp changes
// or a related resource (Deployment, Service) changes
func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    log.Info("Reconciling WebApp", "name", req.Name, "namespace", req.Namespace)

    // ── Step 1: Fetch the WebApp resource ─────────────────────────────────
    webapp := &myapiv1.WebApp{}
    if err := r.Get(ctx, req.NamespacedName, webapp); err != nil {
        if errors.IsNotFound(err) {
            // Object deleted before we could reconcile — nothing to do
            // Owned resources (Deployment, Service) cleaned up by GC via OwnerReference
            return ctrl.Result{}, nil
        }
        // Real error — requeue with exponential backoff
        return ctrl.Result{}, fmt.Errorf("failed to get WebApp: %w", err)
    }

    // ── Step 2: Handle deletion (finalizer pattern) ────────────────────────
    if !webapp.DeletionTimestamp.IsZero() {
        // Object is being deleted — run cleanup before allowing deletion
        if controllerutil.ContainsFinalizer(webapp, webappFinalizer) {
            if err := r.cleanupExternalResources(ctx, webapp); err != nil {
                return ctrl.Result{}, err
            }
            // Remove finalizer — Kubernetes will now complete deletion
            controllerutil.RemoveFinalizer(webapp, webappFinalizer)
            if err := r.Update(ctx, webapp); err != nil {
                return ctrl.Result{}, err
            }
        }
        return ctrl.Result{}, nil
    }

    // ── Step 3: Add finalizer if not present ──────────────────────────────
    if !controllerutil.ContainsFinalizer(webapp, webappFinalizer) {
        controllerutil.AddFinalizer(webapp, webappFinalizer)
        if err := r.Update(ctx, webapp); err != nil {
            return ctrl.Result{}, err
        }
        // Return here — Update triggers a new reconcile
        return ctrl.Result{}, nil
    }

    // ── Step 4: Reconcile the Deployment ──────────────────────────────────
    deployment, err := r.reconcileDeployment(ctx, webapp)
    if err != nil {
        r.setCondition(webapp, "DeploymentReady", metav1.ConditionFalse,
            "DeploymentFailed", err.Error())
        r.updateStatus(ctx, webapp, "Failed", 0)
        return ctrl.Result{}, err
    }

    // ── Step 5: Reconcile the Service ─────────────────────────────────────
    if err := r.reconcileService(ctx, webapp); err != nil {
        return ctrl.Result{}, err
    }

    // ── Step 6: Update status based on Deployment state ───────────────────
    available := deployment.Status.AvailableReplicas
    phase := "Running"
    if available == 0 {
        phase = "Pending"
    } else if available < webapp.Spec.Replicas {
        phase = "Degraded"
    }

    r.setCondition(webapp, "DeploymentReady", metav1.ConditionTrue,
        "DeploymentAvailable", fmt.Sprintf("%d/%d replicas available", available, webapp.Spec.Replicas))

    if err := r.updateStatus(ctx, webapp, phase, available); err != nil {
        return ctrl.Result{}, err
    }

    log.Info("Reconciliation complete", "phase", phase, "available", available)
    return ctrl.Result{}, nil
}

// reconcileDeployment creates or updates the Deployment owned by this WebApp
func (r *WebAppReconciler) reconcileDeployment(ctx context.Context, webapp *myapiv1.WebApp) (*appsv1.Deployment, error) {
    labels := map[string]string{
        "app":                          webapp.Name,
        "app.kubernetes.io/managed-by": "webapp-controller",
    }

    replicas := webapp.Spec.Replicas
    desired := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      webapp.Name,
            Namespace: webapp.Namespace,
            Labels:    labels,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{MatchLabels: labels},
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{Labels: labels},
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Name:  "app",
                        Image: webapp.Spec.Image,
                        Ports: []corev1.ContainerPort{{
                            ContainerPort: webapp.Spec.Port,
                        }},
                    }},
                },
            },
        },
    }

    // Set WebApp as the owner of the Deployment
    // When WebApp is deleted, Kubernetes automatically deletes the Deployment
    if err := controllerutil.SetControllerReference(webapp, desired, r.Scheme); err != nil {
        return nil, err
    }

    // CreateOrUpdate: if Deployment doesn't exist, create it
    // If it exists, update it to match desired state
    existing := &appsv1.Deployment{}
    result, err := controllerutil.CreateOrUpdate(ctx, r.Client, existing, func() error {
        // Copy name/namespace for lookup
        existing.Name = desired.Name
        existing.Namespace = desired.Namespace
        // Mutate to desired state
        existing.Spec = desired.Spec
        existing.Labels = desired.Labels
        return nil
    })
    if err != nil {
        return nil, fmt.Errorf("failed to reconcile Deployment: %w", err)
    }

    if result != controllerutil.OperationResultNone {
        log.FromContext(ctx).Info("Deployment reconciled", "result", result)
    }

    // Re-fetch to get current status
    if err := r.Get(ctx, client.ObjectKeyFromObject(desired), existing); err != nil {
        return nil, err
    }
    return existing, nil
}

// reconcileService creates or updates the Service for this WebApp
func (r *WebAppReconciler) reconcileService(ctx context.Context, webapp *myapiv1.WebApp) error {
    labels := map[string]string{"app": webapp.Name}

    desired := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      webapp.Name,
            Namespace: webapp.Namespace,
            Labels:    labels,
        },
        Spec: corev1.ServiceSpec{
            Selector: labels,
            Ports: []corev1.ServicePort{{
                Port:       webapp.Spec.Port,
                TargetPort: intstr.FromInt32(webapp.Spec.Port),
                Protocol:   corev1.ProtocolTCP,
            }},
            Type: corev1.ServiceTypeClusterIP,
        },
    }

    if err := controllerutil.SetControllerReference(webapp, desired, r.Scheme); err != nil {
        return err
    }

    existing := &corev1.Service{}
    _, err := controllerutil.CreateOrUpdate(ctx, r.Client, existing, func() error {
        existing.Name = desired.Name
        existing.Namespace = desired.Namespace
        existing.Spec.Selector = desired.Spec.Selector
        existing.Spec.Ports = desired.Spec.Ports
        // Note: do NOT overwrite existing.Spec.ClusterIP — it is assigned by k8s
        return nil
    })
    return err
}

// updateStatus updates the WebApp's status subresource
func (r *WebAppReconciler) updateStatus(ctx context.Context, webapp *myapiv1.WebApp, phase string, available int32) error {
    webapp.Status.Phase = phase
    webapp.Status.AvailableReplicas = available
    // ObservedGeneration tells users the controller processed this generation of spec
    webapp.Status.ObservedGeneration = webapp.Generation
    // Status().Update() only modifies the status subresource
    // It does NOT touch spec — separate API endpoint
    return r.Status().Update(ctx, webapp)
}

// setCondition updates a condition in the status
func (r *WebAppReconciler) setCondition(webapp *myapiv1.WebApp, condType string,
    status metav1.ConditionStatus, reason, message string) {
    // Uses controller-runtime's meta helper — handles LastTransitionTime and deduplication
    // meta.SetStatusCondition(&webapp.Status.Conditions, metav1.Condition{...})
}

// cleanupExternalResources runs before finalizer removal
// Use this for resources outside Kubernetes (cloud LBs, DNS records, databases)
func (r *WebAppReconciler) cleanupExternalResources(ctx context.Context, webapp *myapiv1.WebApp) error {
    // Kubernetes resources (Deployment, Service) are cleaned up automatically
    // via OwnerReference garbage collection — no manual cleanup needed here
    // Only use this for resources Kubernetes cannot garbage collect:
    //   - AWS Route53 DNS records
    //   - External load balancers created via cloud API
    //   - Database users or schemas
    log.FromContext(ctx).Info("Running external cleanup for WebApp", "name", webapp.Name)
    return nil
}

// SetupWithManager wires the controller to watch WebApp and owned resources
func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        // Primary resource — reconcile when WebApp changes
        For(&myapiv1.WebApp{}).
        // Secondary resources — reconcile the WebApp when owned Deployment changes
        // The controller maps Deployment → owning WebApp → enqueues WebApp for reconcile
        Owns(&appsv1.Deployment{}).
        // Also watch Services
        Owns(&corev1.Service{}).
        // Limit concurrent reconciles (default 1 — safe for most cases)
        WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
        Complete(r)
}
```

### Why each section matters

**`r.Get()` returning `IsNotFound`:**
The most common beginner mistake is panicking or returning an error when the
object is not found. Not found simply means the object was deleted between
the event being enqueued and the reconciler running — return `nil` (success)
and do nothing. The owned resources (Deployment, Service) will be cleaned up
by Kubernetes garbage collection via OwnerReferences.

**`DeletionTimestamp` check (finalizer pattern):**
When a user runs `kubectl delete webapp my-webapp`, Kubernetes sets
`DeletionTimestamp` on the object but does NOT delete it immediately if there
are finalizers. Your controller sees the non-zero timestamp, runs cleanup
(external resources that Kubernetes cannot garbage collect), removes the
finalizer, and then Kubernetes completes the deletion. Without finalizers,
Kubernetes would delete the object immediately — you would never get a chance
to clean up external resources.

**`controllerutil.SetControllerReference()`:**
Sets the `ownerReferences` field on the Deployment and Service to point to
the WebApp. When the WebApp is deleted, Kubernetes automatically garbage-collects
all objects that have it as an owner. This is how you get automatic cleanup of
secondary resources without writing any deletion logic.

**`controllerutil.CreateOrUpdate()`:**
The idempotent workhorse of every controller. It:
1. Tries to `Get` the existing object
2. If not found → `Create` it
3. If found → calls your mutate function → `Update` it
Running `Reconcile` 100 times on the same WebApp produces the same result as
running it once — this idempotency is the core contract of the reconciler.

**`r.Status().Update()` vs `r.Update()`:**
`r.Update()` updates the full object including spec — a controller should
almost never call this on the CR it manages (that would overwrite user intent).
`r.Status().Update()` hits the `/status` subresource endpoint — only the
status field is changed. This separation is critical and requires the
`subresources: status: {}` in the CRD manifest.

**`Owns(&appsv1.Deployment{})`:**
Without this, if someone manually deletes the Deployment that the controller
created, the controller would not know — the WebApp has not changed, so no
event is generated. `Owns()` sets up a watch on Deployments and when a
change occurs to an owned Deployment, the controller maps it back to the
owning WebApp and enqueues a reconcile. This is how self-healing works.

---

## Step 4 — RBAC

### File: `config/rbac/role.yaml`

### What it is

RBAC grants the controller's ServiceAccount permission to watch and
manage Kubernetes resources. Without correct RBAC, every API call
from the controller returns `403 Forbidden`.

### How RBAC is generated

The `+kubebuilder:rbac:` markers in the controller file are the source
of truth. Running `make manifests` generates this YAML automatically:

```yaml
# config/rbac/role.yaml — generated from +kubebuilder:rbac markers
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: webapp-controller-role
rules:
# Permission on WebApp CRs — must be able to get/list/watch to reconcile
- apiGroups: ["apps.example.com"]
  resources: ["webapps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Permission on status subresource — SEPARATE from the main resource
# Allows controller to update status without being able to change spec via update
- apiGroups: ["apps.example.com"]
  resources: ["webapps/status"]
  verbs: ["get", "update", "patch"]

# Permission on finalizers subresource — needed to add/remove finalizers
- apiGroups: ["apps.example.com"]
  resources: ["webapps/finalizers"]
  verbs: ["update"]

# Permission on secondary resources the controller creates/manages
- apiGroups: ["apps"]                     # apps/v1 API group
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

- apiGroups: [""]                         # core API group (v1)
  resources: ["services"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

# Events — for recording human-readable controller actions
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
```

```yaml
# config/rbac/service_account.yaml — the identity the controller runs as
apiVersion: v1
kind: ServiceAccount
metadata:
  name: webapp-controller
  namespace: webapp-system
```

```yaml
# config/rbac/role_binding.yaml — connects ServiceAccount to the ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: webapp-controller-rolebinding
subjects:
- kind: ServiceAccount
  name: webapp-controller           # must match ServiceAccount name
  namespace: webapp-system          # must match ServiceAccount namespace
roleRef:
  kind: ClusterRole
  name: webapp-controller-role      # must match ClusterRole name
  apiGroup: rbac.authorization.k8s.io
```

### Why each RBAC rule matters

**`webapps` vs `webapps/status` vs `webapps/finalizers`:**
These are three separate sub-resources with independent permissions:
- `webapps` — read/write the spec. Grant carefully.
- `webapps/status` — read/write only the status. Controller uses `r.Status().Update()`
- `webapps/finalizers` — needed to call `r.Update()` when adding/removing finalizers

**Why `list` and `watch` are needed (not just `get`):**
The controller uses an informer that first `List`s all existing WebApps
at startup, then `Watch`es for changes via a long-lived HTTP connection.
Without `list`, the informer cannot build its initial cache.
Without `watch`, it cannot receive change events.

**Why `events` permission is needed:**
Controllers record Events to give users visibility into what they are doing:
```bash
kubectl describe webapp my-webapp
# Events:
#   Normal   DeploymentCreated  5m    webapp-controller  Created deployment my-webapp
#   Normal   ServiceCreated     5m    webapp-controller  Created service my-webapp
#   Warning  DeploymentFailed   2m    webapp-controller  Image pull failed: nginx:invalid
```
Without events permission, these are silently dropped.

---

## Step 5 — Main Entry Point

### File: `cmd/main.go`

```go
package main

import (
    "flag"
    "os"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "sigs.k8s.io/controller-runtime/pkg/healthz"

    myapiv1 "github.com/example/webapp-controller/api/v1"
    "github.com/example/webapp-controller/internal/controller"
)

var scheme = runtime.NewScheme()

func init() {
    // Register ALL types that the controller will read or write
    // Without registration, the client cannot serialize/deserialize the objects
    _ = myapiv1.AddToScheme(scheme)   // our WebApp types
    _ = appsv1.AddToScheme(scheme)    // Deployment
    _ = corev1.AddToScheme(scheme)    // Service, Event
}

func main() {
    var metricsAddr string
    var probeAddr string
    var enableLeaderElection bool

    flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "Metrics address")
    flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "Health probe address")
    flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election")
    flag.Parse()

    ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

    // Manager is the central orchestrator — owns cache, controllers, webhook server
    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:                 scheme,
        MetricsBindAddress:     metricsAddr,
        HealthProbeBindAddress: probeAddr,
        LeaderElection:         enableLeaderElection,
        LeaderElectionID:       "webapp-controller.apps.example.com",
        // Namespace: "default", // uncomment to restrict to one namespace
    })
    if err != nil {
        ctrl.Log.Error(err, "Unable to create manager")
        os.Exit(1)
    }

    // Register our reconciler with the manager
    if err := (&controller.WebAppReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        ctrl.Log.Error(err, "Unable to create controller")
        os.Exit(1)
    }

    // Health checks — used by Kubernetes liveness and readiness probes
    mgr.AddHealthzCheck("healthz", healthz.Ping)
    mgr.AddReadyzCheck("readyz", mgr.GetCache().WaitForCacheSync)

    // Start all controllers — blocks until SIGTERM/SIGINT
    ctrl.Log.Info("Starting manager")
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        ctrl.Log.Error(err, "Problem running manager")
        os.Exit(1)
    }
}
```

**Why `init()` registers schemes:**
The `runtime.Scheme` is a lookup table: `(Group, Version, Kind) → Go type`.
When the controller receives `GET /apis/apps.example.com/v1/webapps/my-webapp`,
the client library uses the scheme to know which Go struct to deserialize
the JSON into. If `WebApp` is not registered → deserialization fails with
a cryptic "no kind registered" error.

**Why `LeaderElection` matters:**
If you run multiple replicas of the controller for high availability (HA),
only one should actively reconcile at a time — otherwise you get multiple
controllers simultaneously creating/updating the same resources, causing
race conditions. Leader election uses a Kubernetes `Lease` object — only
the current leader holds the lease and runs the reconcile loop. Others
stand by and take over if the leader fails.

---

## Step 6 — Controller Deployment

### File: `config/manager/manager.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-controller
  namespace: webapp-system
  labels:
    app: webapp-controller
spec:
  replicas: 2                             # 2 replicas for HA with leader election
  selector:
    matchLabels:
      app: webapp-controller
  template:
    metadata:
      labels:
        app: webapp-controller
    spec:
      # The identity the controller uses for RBAC
      serviceAccountName: webapp-controller

      # Security best practice: do not run as root
      securityContext:
        runAsNonRoot: true

      containers:
      - name: manager
        image: my-registry/webapp-controller:v1.0.0
        args:
        - --leader-elect                  # enable leader election for HA
        - --metrics-bind-address=:8080
        - --health-probe-bind-address=:8081

        # Always set both requests and limits
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 500m
            memory: 128Mi

        # Liveness: restart if controller is broken
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20

        # Readiness: only receive traffic when cache is synced
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10

        # Seccomp profile — restrict syscalls
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
          readOnlyRootFilesystem: true
          runAsNonRoot: true

      terminationGracePeriodSeconds: 10   # allow in-flight reconciles to finish
```

**Why `serviceAccountName: webapp-controller`:**
This connects the pod to the ServiceAccount that has the ClusterRoleBinding
granting it RBAC permissions. Without this, the pod uses the `default`
ServiceAccount which has no permissions — every API call fails with 403.

**Why `replicas: 2` with `--leader-elect`:**
Two controller replicas provide high availability. If the leader pod crashes,
the standby acquires the lease within seconds and resumes reconciliation.
The `--leader-elect` flag activates the leader election mechanism.

---

## Step 7 — Complete Install Flow

```bash
# 1. Install the CRD into the cluster
#    This registers the new API type with the API server
kubectl apply -f config/crd/bases/apps.example.com_webapps.yaml

# Verify CRD is registered:
kubectl get crd webapps.apps.example.com
kubectl api-resources | grep webapp
# webapp   wa   apps.example.com/v1   true   WebApp

# 2. Create the controller's namespace
kubectl create namespace webapp-system

# 3. Apply RBAC
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml

# 4. Deploy the controller
kubectl apply -f config/manager/manager.yaml

# Verify controller is running:
kubectl get pods -n webapp-system
# NAME                                READY   STATUS    RESTARTS
# webapp-controller-7d9f8b-xk2p1     1/1     Running   0
# webapp-controller-7d9f8b-ab3q2     1/1     Running   0

# 5. Create a WebApp resource
cat <<EOF | kubectl apply -f -
apiVersion: apps.example.com/v1
kind: WebApp
metadata:
  name: my-webapp
  namespace: default
spec:
  replicas: 3
  image: nginx:1.25
  port: 8080
EOF

# 6. Watch what happens
kubectl get webapps                      # custom columns from printcolumn
# NAME         PHASE     REPLICAS   AVAILABLE   AGE
# my-webapp    Running   3          3           30s

kubectl describe webapp my-webapp        # full details including conditions and events
kubectl get deployment my-webapp         # controller created this
kubectl get service my-webapp            # controller created this

# 7. Test self-healing
kubectl delete deployment my-webapp      # manually delete the Deployment
kubectl get deployment my-webapp         # controller recreates it within seconds

# 8. Test spec update
kubectl patch webapp my-webapp --type=merge -p '{"spec":{"replicas":5}}'
kubectl get webapp my-webapp             # status updates to reflect new replicas

# 9. Test deletion
kubectl delete webapp my-webapp
# Controller runs finalizer cleanup, then Deployment and Service are
# garbage-collected automatically via OwnerReferences
```

---

## Step 8 — Complete File Structure

```
webapp-controller/
├── api/
│   └── v1/
│       ├── webapp_types.go          ← API struct definitions (your schema)
│       ├── groupversion_info.go     ← group/version registration
│       └── zz_generated.deepcopy.go ← auto-generated by controller-gen
│
├── internal/
│   └── controller/
│       ├── webapp_controller.go     ← reconcile loop, business logic
│       └── webapp_controller_test.go ← unit + integration tests (envtest)
│
├── cmd/
│   └── main.go                      ← manager setup, scheme registration
│
├── config/
│   ├── crd/
│   │   └── bases/
│   │       └── apps.example.com_webapps.yaml  ← auto-generated CRD manifest
│   ├── rbac/
│   │   ├── service_account.yaml     ← controller's identity
│   │   ├── role.yaml                ← auto-generated permissions
│   │   └── role_binding.yaml        ← connects identity to permissions
│   └── manager/
│       └── manager.yaml             ← Deployment for the controller
│
├── Dockerfile                       ← builds the controller image
├── Makefile                         ← generate, build, deploy commands
└── go.mod                           ← Go module dependencies
```

---

## Everything Connected — The Full Flow

```
USER                   KUBERNETES API SERVER           CONTROLLER
 │                              │                          │
 │  kubectl apply -f webapp.yaml│                          │
 │─────────────────────────────►│                          │
 │                              │ 1. Validate against      │
 │                              │    openAPIV3Schema       │
 │                              │ 2. Store in etcd         │
 │                              │ 3. Notify watchers       │
 │                              │──────────────────────────►
 │                              │                          │ 4. Informer receives
 │                              │                          │    ADD event
 │                              │                          │ 5. Enqueue Request{
 │                              │                          │    namespace/name}
 │                              │                          │
 │                              │                          │ 6. Reconcile() called
 │                              │                          │    r.Get(webapp)
 │                              │◄─────────────────────────│
 │                              │ 7. Returns WebApp spec   │
 │                              │──────────────────────────►
 │                              │                          │ 8. r.Create(deployment)
 │                              │◄─────────────────────────│
 │                              │    stores in etcd        │
 │                              │──────────────────────────►
 │                              │                          │ 9. r.Create(service)
 │                              │◄─────────────────────────│
 │                              │    stores in etcd        │
 │                              │──────────────────────────►
 │                              │                          │ 10. r.Status().Update()
 │                              │◄─────────────────────────│     phase=Running
 │                              │    updates etcd status   │
 │                              │                          │
 │  kubectl get webapp          │                          │
 │─────────────────────────────►│                          │
 │◄─────────────────────────────│                          │
 │  NAME      PHASE    REPLICAS │                          │
 │  my-webapp Running  3        │                          │
```

---

## Interview Summary — Why Each Component Exists

| Component | File | Why it exists |
|---|---|---|
| **API Types** | `api/v1/webapp_types.go` | Defines the CRD schema as Go structs. Source of truth for validation, serialization, and generated CRD YAML |
| **GroupVersion** | `api/v1/groupversion_info.go` | Registers the CRD in the runtime Scheme so the client can serialize/deserialize objects |
| **CRD YAML** | `config/crd/bases/*.yaml` | Registers the new API type with the Kubernetes API server. Without it, `kubectl apply -f webapp.yaml` fails with "no matches for kind WebApp" |
| **Spec/Status split** | CRD `subresources: status: {}` | Separates user intent (spec) from controller observation (status). Different RBAC, different API endpoints, no overwrites |
| **Controller** | `internal/controller/webapp_controller.go` | Watches for WebApp events, reconciles desired vs actual state, creates/updates Deployment and Service |
| **Reconcile function** | Controller | The core control loop — idempotent by design, must handle not-found, conflicts, and transient errors |
| **OwnerReference** | `SetControllerReference()` | Enables automatic garbage collection of Deployment and Service when WebApp is deleted |
| **Finalizer** | Controller | Allows controller to run cleanup before Kubernetes completes deletion |
| **RBAC markers** | `+kubebuilder:rbac:...` | Source of truth for permissions — generates ClusterRole YAML automatically |
| **ServiceAccount** | `config/rbac/` | Identity for the controller pod — RBAC grants permissions to this identity |
| **Scheme registration** | `cmd/main.go init()` | Maps (Group, Version, Kind) strings to Go types for serialization |
| **Leader election** | `--leader-elect` flag | Ensures only one controller replica is active — prevents split-brain with multiple replicas |
| **Owns()** | `SetupWithManager()` | Triggers reconcile when owned Deployment/Service changes — enables self-healing |
| **Controller Deployment** | `config/manager/manager.yaml` | Runs the controller as a pod in the cluster with the correct ServiceAccount |