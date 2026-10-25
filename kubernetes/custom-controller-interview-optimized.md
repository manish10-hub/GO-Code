# Kubernetes Custom Controllers — Optimized Interview Revision Guide

> **Goal:** Quickly revise Kubernetes Controllers, Operators, CRDs, controller-runtime, client-go, RBAC, webhooks, and reconciliation before an interview.
>
> **How to use:** First read the **Quick Mental Model**, then the questions from top to bottom. The answers are intentionally short and interview-friendly. Deeper details are included only where they are useful for senior/tech-lead interviews.

---

# 1. Quick Mental Model

The easiest way to remember a Kubernetes custom controller is:

```text
                Kubernetes API Server
                       │
                List / Watch events
                       │
                       ▼
                 Informer / Cache
                       │
                       ▼
                   Work Queue
                       │
                       ▼
                  Reconcile()
                       │
             ┌─────────┴─────────┐
             ▼                   ▼
        Read current state    Compare with
        from cache/API        desired state
             │                   │
             └─────────┬─────────┘
                       ▼
                Create / Update
                / Delete resources
                       │
                       ▼
                 Update Status
                       │
                       ▼
                 Reconcile again
```

### One sentence to remember

**A controller continuously watches Kubernetes resources and makes the actual state match the desired state.**

---

# 2. Beginner Level

## Q1. What is a Kubernetes Controller?

A **Controller** is a program that continuously watches Kubernetes resources and tries to make the **actual state match the desired state**.

Simple example:

```text
Desired:  Deployment should have 3 replicas
Actual:   Deployment has 2 replicas

Controller:
    sees the difference
        ↓
    takes action
        ↓
    Kubernetes creates the missing Pod
```

### Interview answer

> A controller is a reconciliation loop that observes the current state, compares it with the desired state, and takes action to make them match.

---

## Q2. What is an Operator? Is it different from a Controller?

An **Operator is a specialized controller** that contains application-specific operational knowledge.

For example, a database Operator may know how to:

- create a database cluster
- scale it safely
- perform backups
- recover from failures
- perform upgrades in the correct order

### Remember

```text
Controller
   └── Operator
```

**All Operators are Controllers, but not all Controllers are Operators.**

The source explains this distinction explicitly: Operators manage applications using domain-specific knowledge, while ordinary controllers can manage Kubernetes resources already understood by Kubernetes. fileciteturn6file0L30-L42

---

## Q3. Why do we need Operators?

Kubernetes built-in resources such as Deployments and Services handle common application needs.

But applications such as databases and message queues often require special logic:

- ordered startup
- data-aware scaling
- backups
- recovery
- safe upgrades

An Operator puts this operational knowledge into code.

---

## Q4. What is a CRD?

**CRD = Custom Resource Definition.**

It extends the Kubernetes API with your own resource type.

For example:

```yaml
kind: WebApp
```

After installing the CRD, Kubernetes understands:

```bash
kubectl get webapps
kubectl describe webapp my-webapp
```

### Remember

```text
CRD = Definition of a new Kubernetes resource type
CR  = Actual instance of that type
```

---

## Q5. What is a Custom Resource (CR)?

A **Custom Resource** is an actual object created from a CRD.

Example:

```yaml
apiVersion: apps.example.com/v1
kind: WebApp
metadata:
  name: my-webapp
spec:
  replicas: 3
  image: nginx:1.25
```

Here:

```text
WebApp = resource type defined by CRD
my-webapp = Custom Resource instance
```

---

## Q6. What is the difference between `spec` and `status`?

| Field | Meaning | Usually written by |
|---|---|---|
| `spec` | What the user wants | User / automation |
| `status` | What actually happened | Controller |

Example:

```yaml
spec:
  replicas: 3

status:
  availableReplicas: 2
  phase: Degraded
```

### Easy memory trick

**Spec = Want**

**Status = Reality**

The source also emphasizes using the `/status` subresource so spec and status updates remain separate. fileciteturn6file0L184-L195

---

## Q7. What is the reconciliation loop?

The reconciliation loop is the controller's main logic:

```text
Observe → Compare → Act
```

Example:

```text
Observe:
  Deployment has 2 Pods

Desired:
  CR says 3 Pods

Compare:
  2 != 3

Act:
  Update Deployment to 3 replicas
```

The controller keeps doing this whenever relevant events occur.

---

# 3. CRD Design

## Q8. What are the important parts of a CRD?

Remember these:

```text
CRD
├── Group
├── Version
├── Names
├── Scope
├── Schema
├── Status subresource
├── Printer columns
└── Conversion strategy (if required)
```

### Most important for interviews

- **Group** — API domain
- **Version** — API version
- **Scope** — Namespaced or Cluster
- **Schema** — valid fields and validation rules
- **Status subresource** — separates status updates from spec updates

---

## Q9. What is the difference between Namespaced and Cluster-scoped resources?

### Namespaced

The object belongs to a namespace.

Examples:

```text
Pod
Service
Deployment
```

### Cluster-scoped

The object belongs to the whole cluster.

Examples:

```text
Node
StorageClass
ClusterRole
```

Most application-level CRDs are usually namespaced.

---

## Q10. What is the `/status` subresource and why is it important?

It creates a separate API endpoint for updating status.

Instead of:

```text
Update entire CR
```

the controller can do:

```text
Status().Update()
```

This gives:

- separate permissions
- fewer update conflicts
- clear separation between desired and observed state

The source specifically describes `/status` as a separate endpoint and RBAC target. fileciteturn7file0L738-L751

---

## Q11. What are Conditions and `ObservedGeneration`?

A good controller should expose health using standard Kubernetes conditions.

Example:

```yaml
conditions:
- type: Ready
  status: "True"
  reason: DeploymentAvailable
  message: "3/3 replicas available"
```

Useful fields include:

- `Type`
- `Status`
- `Reason`
- `Message`
- `ObservedGeneration`
- `LastTransitionTime`

### Why `ObservedGeneration`?

Suppose:

```text
metadata.generation = 5
status.observedGeneration = 4
```

It means:

> The user changed the spec, but the controller has not processed generation 5 yet.

The source recommends `metav1.Condition` and `ObservedGeneration` for machine-readable status. fileciteturn6file0L442-L446

---

# 4. client-go and controller-runtime

## Q12. What is client-go?

`client-go` is the official Go library for communicating with the Kubernetes API server.

It provides:

- typed clients
- dynamic client
- REST client
- informers
- work queues
- discovery client

### Simple mental model

```text
client-go
   ├── Client
   ├── Informer
   ├── Cache
   └── Work Queue
```

---

## Q13. What is controller-runtime?

`controller-runtime` provides higher-level building blocks for writing Kubernetes controllers and Operators.

It gives you things such as:

```text
Manager
Controller
Reconciler
Cache
Client
Leader Election
Webhooks
Health checks
Metrics
```

### Simple difference

```text
client-go
    ↓
Low-level building blocks

controller-runtime
    ↓
Higher-level controller framework
```

The source identifies Manager, Reconciler, and Builder as core controller-runtime concepts. fileciteturn6file0L315-L333

---

## Q14. What is the Manager?

The **Manager** is the central component of controller-runtime.

It manages:

- controllers
- cache
- Kubernetes client
- leader election
- health/readiness
- metrics
- webhook server

### Interview answer

> The Manager is the central orchestrator that starts and coordinates the controllers and their shared infrastructure such as cache, client, leader election, and webhooks.

The source lists these Manager responsibilities directly. fileciteturn5file1L88-L100

---

## Q15. What is the Reconciler?

The **Reconciler contains the business logic**.

It receives a request containing:

```text
namespace + name
```

and performs the reconciliation.

Typical flow:

```text
Get CR
  ↓
Handle deletion
  ↓
Check/create child resources
  ↓
Update child resources
  ↓
Update CR status
  ↓
Return
```

Returning an error normally causes a retry with rate limiting/backoff. `RequeueAfter` can be used when controlled polling is needed. fileciteturn5file1L104-L114

---

# 5. How Events Reach Reconcile()

## Q16. How does a controller know when to reconcile?

Usually through watches.

Three important sources are:

1. **Primary resource** — the CR itself
2. **Owned resources** — Deployment, Service, Job, etc.
3. **Other resources** — resources that are related but not owned

Example:

```text
WebApp changes
     ↓
Reconcile WebApp

Deployment changes
     ↓
Find owning WebApp
     ↓
Reconcile WebApp
```

The source describes these three watch categories and the event path into the work queue. fileciteturn6file0L377-L409

---

## Q17. What is `For()`, `Owns()`, and `Watches()`?

This is a very common interview question.

### `For()`

Defines the **primary resource**.

```go
For(&WebApp{})
```

Meaning:

> Reconcile when WebApp changes.

### `Owns()`

Used when the secondary resource has an OwnerReference to the primary resource.

```go
Owns(&appsv1.Deployment{})
```

Meaning:

> If an owned Deployment changes, reconcile the owning WebApp.

### `Watches()`

Used when the resource is not owned and you need custom mapping.

```text
ConfigMap changes
      ↓
Find affected WebApps
      ↓
Enqueue WebApp reconcile requests
```

The source distinguishes `Owns()` from `Watches()` based on OwnerReferences and explicit mapping. fileciteturn6file0L487-L499

---

## Q18. What is an Informer?

An **Informer keeps a local cache of Kubernetes resources synchronized with the API server.**

Basic flow:

```text
API Server
    │
    ├── LIST ──► initial objects
    │
    └── WATCH ─► future changes
                    │
                    ▼
                 Cache
```

Informer handles:

- initial List
- Watch
- Add/Update/Delete events
- local cache updates
- reconnecting the watch

The source describes List → Watch → DeltaFIFO → Store as the internal flow. fileciteturn5file3L240-L250

---

## Q19. What is a SharedInformerFactory?

It allows multiple consumers to share the same informer and cache.

Without sharing:

```text
Controller A → Pod informer → API server
Controller B → Pod informer → API server
Controller C → Pod informer → API server
```

With a shared factory:

```text
             Pod Informer
             /     |     \
            A      B      C
```

Benefits:

- fewer API-server watch connections
- less memory
- less duplicate work

The source explicitly recommends sharing informers for this reason. fileciteturn5file3L254-L258

---

## Q20. What is the Work Queue?

The work queue stores **reconciliation requests**, usually represented by:

```text
namespace/name
```

Important point:

> The event handler normally puts the key into the queue; it does not execute the business logic directly.

Then:

```text
Work Queue
    ↓
Worker
    ↓
Reconcile()
```

The queue also prevents duplicate keys from being processed unnecessarily when many events arrive quickly. fileciteturn5file6L485-L500

---

# 6. Cache and Client

## Q21. Does `r.Get()` always call the API server?

With controller-runtime's normal cached client, reads are generally served from the **local informer cache**.

```text
r.Get()
   ↓
Controller-runtime cache
   ↓
Local memory
```

The cache is:

- fast
- shared within the Manager
- eventually consistent

If you specifically need strongly current data, controller-runtime also provides ways to bypass the cache for particular reads.

### Interview point

> Controller reads are normally cache-based, while writes go to the API server.

The source describes the controller-runtime cache as an informer-backed local cache used by `Get()` and `List()`. fileciteturn5file5L459-L475

---

## Q22. What is the difference between typed and dynamic clients?

### Typed client

Uses known Go types.

```go
clientset.CoreV1().Pods(...)
```

Advantages:

- compile-time type safety
- easier development
- strongly typed fields

### Dynamic client

Uses:

```go
unstructured.Unstructured
```

and works with resources whose types are not known at compile time.

Use it for:

- generic tools
- platforms handling arbitrary resources
- controllers that do not know CRDs beforehand

The source recommends typed/generated clients for strongly typed controllers and dynamic clients for generic resource handling. fileciteturn5file5L406-L451

---

# 7. Reconciliation Design

## Q23. What does an idempotent controller mean?

**Running reconciliation multiple times should produce the same final state.**

Example:

```text
First reconcile:
  Deployment doesn't exist
  → create it

Second reconcile:
  Deployment already exists
  → verify/update if needed

Third reconcile:
  already correct
  → do nothing
```

This is critical because reconciliation can happen many times.

The source calls `CreateOrUpdate()` an idempotent workhorse and explains that repeated reconciliation should produce the same result. fileciteturn7file1L1058-L1064

---

## Q24. What is `CreateOrUpdate()`?

`controllerutil.CreateOrUpdate()` simplifies the common controller pattern:

```text
Get object
   ↓
Does it exist?
   ├── No  → Create
   └── Yes → Mutate to desired state → Update
```

It helps make reconciliation idempotent.

---

## Q25. What happens when the resource is NotFound?

This is an important controller pattern.

If:

```go
r.Get(...)
```

returns `NotFound`, normally return success:

```go
return ctrl.Result{}, nil
```

Why?

Because the object may have been deleted after the event was queued but before reconciliation started.

The source explicitly recommends treating NotFound as a normal deletion case rather than an error. fileciteturn7file1L1034-L1041

---

# 8. OwnerReferences and Finalizers

## Q26. What is an OwnerReference?

An OwnerReference tells Kubernetes:

> "This resource belongs to that resource."

Example:

```text
WebApp
  │
  ├── Deployment
  └── Service
```

If Deployment and Service have WebApp as their owner, Kubernetes garbage collection can remove them when the WebApp is deleted.

This avoids writing manual cleanup logic for normal Kubernetes child resources. fileciteturn5file0L24-L30

---

## Q27. What is a Finalizer?

A finalizer allows the controller to perform cleanup **before Kubernetes finishes deleting the resource**.

Flow:

```text
kubectl delete WebApp
        ↓
deletionTimestamp set
        ↓
Finalizer still present
        ↓
Controller performs cleanup
        ↓
Controller removes finalizer
        ↓
Kubernetes completes deletion
```

Use finalizers mainly for resources Kubernetes cannot automatically garbage-collect, such as:

- external load balancers
- DNS records
- external databases

The source describes this deletion flow directly. fileciteturn5file0L34-L38

---

## Q28. OwnerReference vs Finalizer — what is the difference?

This is worth memorizing.

| OwnerReference | Finalizer |
|---|---|
| Automatic child cleanup | Custom cleanup before deletion |
| Kubernetes garbage collection | Controller performs cleanup |
| Good for Kubernetes child resources | Good for external resources |
| Example: Deployment owned by CR | Example: cloud LB created by CR |

### Easy memory

**OwnerReference = "Who owns me?"**

**Finalizer = "Wait, I need to clean something first."**

---

# 9. RBAC

## Q29. Why does a Controller need RBAC?

A controller runs as a Pod and normally uses a **ServiceAccount**.

RBAC determines what that ServiceAccount is allowed to do.

Without the required permissions:

```text
Controller → API Server
              ↓
           403 Forbidden
```

The source explains that the controller's ServiceAccount needs permissions for the resources it watches and manages. fileciteturn7file0L1082-L1090

---

## Q30. Explain ServiceAccount, Role, RoleBinding, ClusterRole, ClusterRoleBinding.

Remember:

```text
ServiceAccount
      ↓
   Identity

Role / ClusterRole
      ↓
   Permissions

RoleBinding / ClusterRoleBinding
      ↓
   Connect identity → permissions
```

### Role

Permissions inside one namespace.

### ClusterRole

Cluster-wide permissions or permissions for cluster-scoped resources.

### RoleBinding

Connects a subject to a Role in a namespace.

### ClusterRoleBinding

Connects a subject to a ClusterRole across the cluster.

---

## Q31. What is `+kubebuilder:rbac`?

It is a marker placed in controller code.

Example:

```go
// +kubebuilder:rbac:groups=apps.example.com,resources=webapps,verbs=get;list;watch;update
```

`controller-gen` reads these markers and generates RBAC manifests.

Typical command:

```bash
make manifests
```

This keeps permissions close to the code that uses them. fileciteturn7file0L80-L97

---

## Q32. What is the Principle of Least Privilege?

Give the controller **only the permissions it actually needs**.

Avoid:

```yaml
verbs: ["*"]
resources: ["*"]
```

Prefer:

```text
Read-only resource:
  get, list, watch

Resource the controller manages:
  only required create/update/patch/delete permissions
```

Also prefer namespace-scoped permissions when cluster-wide access is unnecessary.

---

## Q33. How do you debug `403 Forbidden` from a controller?

Use this sequence:

```text
1. Read the exact 403 error
        ↓
2. Identify:
   ServiceAccount
   verb
   resource
   namespace
        ↓
3. Run kubectl auth can-i
        ↓
4. Check Role/ClusterRole
        ↓
5. Check RoleBinding/ClusterRoleBinding
        ↓
6. Check subresources:
   /status
   /finalizers
        ↓
7. Regenerate manifests if using Kubebuilder markers
```

Example:

```bash
kubectl auth can-i update pods \
  --as=system:serviceaccount:my-operator-system:my-operator \
  -n default
```

The source specifically highlights `/status`, generated RBAC, and ServiceAccount binding as common causes. fileciteturn7file0L112-L139

---

# 10. Leader Election and Scaling

## Q34. Why do controllers use leader election?

You may run multiple controller replicas for HA:

```text
Controller Pod 1
Controller Pod 2
Controller Pod 3
```

But you generally want only one active leader performing the reconciliation for a given controller setup.

```text
Pod 1 → Leader → Reconcile
Pod 2 → Standby
Pod 3 → Standby
```

If the leader dies, another replica acquires the lease.

controller-runtime uses a Kubernetes `Lease` for leader election. fileciteturn5file6L489-L500

---

## Q35. Is leader election enough?

No.

The reconciler should still be:

- **idempotent**
- safe to retry
- safe after leader failover
- careful with external side effects

For external APIs, use idempotent operations or idempotency keys when possible.

---

## Q36. How do you handle thousands of CRs?

Avoid reconciling everything periodically.

Use:

```text
Event-driven reconciliation
        +
Informer cache
        +
Rate-limited queue
        +
Exponential backoff
        +
Controlled concurrency
        +
Horizontal sharding when necessary
```

Also:

- update status only when it actually changes
- avoid unnecessary requeues
- limit `MaxConcurrentReconciles`
- protect external APIs from bursts

The source recommends these techniques for preventing thundering-herd behavior. fileciteturn5file6L504-L514

---

# 11. Predicates

## Q37. What are Predicates?

Predicates are **event filters**.

They decide whether an event should reach the work queue.

```text
Event
  ↓
Predicate
  ├── false → ignore
  └── true  → enqueue → Reconcile
```

Example:

> Reconcile only when `spec.size` changes.

Common predicates include:

- `GenerationChangedPredicate`
- `ResourceVersionChangedPredicate`
- `LabelChangedPredicate`
- `AnnotationChangedPredicate`

The source describes predicates as filters that reduce unnecessary reconciliation. fileciteturn6file0L534-L568

---

# 12. ResourceVersion and Conflicts

## Q38. What is `resourceVersion`?

`resourceVersion` helps Kubernetes perform **optimistic concurrency control**.

Think of it like a database row version.

Example:

```text
Controller A reads version 10
Controller B updates object → version 11

Controller A tries to update using version 10
        ↓
API Server
        ↓
409 Conflict
```

The controller should re-read the latest object and retry.

The source recommends `RetryOnConflict` for this pattern. fileciteturn6file0L898-L906

---

# 13. CRD Versioning

## Q39. What are `served` and `storage` in CRD versions?

Example:

```yaml
versions:
- name: v1beta1
  served: true
  storage: true

- name: v1alpha1
  served: true
  storage: false
```

### `served`

The API server accepts requests using this version.

### `storage`

This is the version used to store objects in etcd.

Only **one version** can be the storage version.

Multiple versions can be served.

---

## Q40. When do you need a conversion webhook?

If you have breaking changes between API versions, such as:

```text
v1alpha1
  field: storageSize

v1beta1
  field: diskSize
```

a conversion webhook can translate between versions.

Simple idea:

```text
Client asks for v1alpha1
        ↓
Kubernetes converts
        ↓
Stored representation
```

The source describes conversion webhooks as the mechanism for handling breaking version changes while serving multiple API versions. fileciteturn5file7L554-L587

---

# 14. Admission Webhooks

## Q41. What is an Admission Webhook?

An admission webhook lets you add custom logic to the Kubernetes API request path.

Simplified flow:

```text
kubectl apply
      ↓
Authentication
      ↓
Authorization / RBAC
      ↓
Admission
      ↓
etcd
```

Admission webhooks can:

- modify objects
- validate objects
- reject invalid requests

The source places admission webhooks after authentication/authorization and before persistence. fileciteturn7file0L147-L164

---

## Q42. Mutating vs Validating webhook?

| Mutating | Validating |
|---|---|
| Can modify the object | Cannot modify |
| Can set defaults | Can accept/reject |
| Runs before validation | Runs after mutation |
| Example: default replicas to 1 | Example: reject replicas > 10 |

### Easy memory

**Mutating = Change**

**Validating = Check**

The source explicitly uses this distinction and ordering. fileciteturn7file0L168-L177

---

## Q43. What does an admission webhook need in a real cluster?

Typically:

```text
Webhook Go code
      +
HTTPS/TLS certificate
      +
Service
      +
MutatingWebhookConfiguration
or
ValidatingWebhookConfiguration
```

The webhook configuration tells the API server:

- which resources to intercept
- which operations to intercept
- which Service/path to call
- which CA to trust
- what failure policy to use

---

## Q44. What is `failurePolicy`?

It controls what happens if the API server cannot successfully call the webhook.

```text
Fail
  → reject the request

Ignore
  → allow the request
```

Use carefully:

- `Fail` for safety-critical validation
- `Ignore` when graceful degradation is more important

Always keep webhook rules narrow and make the webhook highly available.

The source highlights webhook availability, scope, timeout, and the outage risk of an unavailable webhook with `failurePolicy: Fail`. fileciteturn7file0L275-L293

---

# 15. Complete "Create a Custom Controller" Interview Question

## Q45. If an interviewer asks: "How would you create a new CRD and controller?"

Give this structured answer:

### Step 1 — Define API types

Create Go structs:

```text
WebAppSpec
WebAppStatus
WebApp
WebAppList
```

Define:

```text
spec
status
validation
conditions
```

### Step 2 — Register the type

Register the GroupVersion and types with the runtime Scheme.

```text
(Group, Version, Kind)
        ↓
    Go type
```

### Step 3 — Generate code

Typical Kubebuilder flow:

```bash
make generate
make manifests
```

This generates things such as:

- DeepCopy methods
- CRD YAML
- RBAC manifests

### Step 4 — Install the CRD

```bash
kubectl apply -f config/crd/
```

Now the API server understands the new resource.

### Step 5 — Implement the Reconciler

The controller:

```text
Get CR
  ↓
Check deletion
  ↓
Create/update child resources
  ↓
Set OwnerReferences
  ↓
Update status
```

### Step 6 — Configure watches

```go
For(&WebApp{}).
    Owns(&appsv1.Deployment{}).
    Owns(&corev1.Service{})
```

### Step 7 — Add RBAC

Give the controller permission to:

```text
get/list/watch
create/update/patch/delete
status
finalizers
```

only where required.

### Step 8 — Run with a Manager

The Manager starts:

```text
Cache
Controller
Leader election
Metrics
Health checks
Webhook server
```

### Step 9 — Deploy the controller

Run it inside Kubernetes using:

```text
Deployment
+
ServiceAccount
+
RBAC
```

### Step 10 — Test self-healing

Create:

```text
WebApp
```

Controller creates:

```text
Deployment
Service
```

Delete the Deployment manually:

```bash
kubectl delete deployment my-webapp
```

The `Owns()` watch causes reconciliation and the controller recreates it.

This complete flow is the central example in the source material. fileciteturn5file4L306-L336

---

# 16. End-to-End Example to Remember

Suppose the user creates:

```yaml
apiVersion: apps.example.com/v1
kind: WebApp
metadata:
  name: my-webapp
spec:
  replicas: 3
  image: nginx:1.25
  port: 8080
```

The complete flow is:

```text
User
 │
 │ kubectl apply
 ▼
API Server
 │
 ├── Validate CRD schema
 │
 ├── Store WebApp
 │
 └── Notify watchers
          │
          ▼
     Informer / Cache
          │
          ▼
      Work Queue
          │
          ▼
     Reconcile()
          │
          ├── Get WebApp
          │
          ├── Check finalizer
          │
          ├── Create/Update Deployment
          │
          ├── Create/Update Service
          │
          └── Update WebApp.Status
```

Then:

```text
Deployment changes
       ↓
Owns(Deployment)
       ↓
WebApp reconcile request
       ↓
Reconcile()
       ↓
Self-healing
```

---

# 17. What Happens During Deletion?

Remember this sequence:

```text
kubectl delete WebApp
        ↓
API Server sets deletionTimestamp
        ↓
Finalizer exists?
        │
        ├── Yes
        │     ↓
        │  Reconcile deletion
        │     ↓
        │  Clean external resources
        │     ↓
        │  Remove finalizer
        │
        └── No
              ↓
        Kubernetes deletes object
```

For normal Kubernetes children:

```text
WebApp
  ├── Deployment
  └── Service
```

OwnerReferences allow Kubernetes garbage collection to clean them up.

---

# 18. Most Important Interview Comparisons

## Controller vs Operator

```text
Controller = reconciliation logic

Operator = controller + application-specific knowledge
```

## CRD vs CR

```text
CRD = defines the resource type
CR  = instance of the resource
```

## Spec vs Status

```text
Spec   = desired state
Status = observed state
```

## OwnerReference vs Finalizer

```text
OwnerReference = automatic child cleanup

Finalizer = controller cleanup before deletion
```

## `For()` vs `Owns()` vs `Watches()`

```text
For()    = primary resource

Owns()   = owned secondary resource

Watches() = arbitrary/related resource + custom mapping
```

## Informer vs Work Queue

```text
Informer   = watches + maintains cache

Work Queue = stores reconcile requests
```

## Typed vs Dynamic client

```text
Typed   = known Go types, type-safe

Dynamic = unstructured, works with arbitrary resources
```

## Mutating vs Validating webhook

```text
Mutating   = modify

Validating = accept/reject
```

## Role vs ClusterRole

```text
Role        = namespace permissions

ClusterRole = cluster-wide / cluster-scoped permissions
```

---

# 19. Senior / Tech-Lead Questions to Prioritize

If interview time is limited, focus on these first:

### Tier 1 — Must know

1. What is a Controller?
2. What is an Operator?
3. What is reconciliation?
4. CRD vs CR
5. Spec vs Status
6. How does Reconcile work?
7. `For()` vs `Owns()` vs `Watches()`
8. Informer and cache
9. Work queue
10. OwnerReference vs Finalizer
11. Idempotency
12. RBAC

### Tier 2 — Very important

13. controller-runtime Manager
14. Typed vs Dynamic client
15. Predicates
16. ResourceVersion / 409 conflicts
17. Leader election
18. Conditions / ObservedGeneration
19. `status` and `finalizers` subresources
20. CRD versioning

### Tier 3 — Senior/Expert

21. Thundering herd
22. Rate limiting / exponential backoff
23. Admission webhooks
24. `failurePolicy`
25. Webhook TLS/availability
26. Designing a controller for thousands of CRs
27. Complete CRD + controller implementation flow

---

# 20. 30-Second Interview Answer

If the interviewer says:

> "Explain how a Kubernetes custom controller works."

Use this:

> A custom controller watches a Kubernetes resource, usually a CRD, through informer/cache mechanisms. When the resource or a related resource changes, a reconcile request is added to a work queue and `Reconcile()` compares the desired state from the CR spec with the actual state. The controller then creates, updates, or deletes resources to make them match and updates the CR status. The reconciler should be idempotent, use OwnerReferences for Kubernetes child resources, finalizers for external cleanup, and appropriate RBAC and leader election for production use.

---

# 21. Final Mental Model

Before the interview, remember only this picture:

```text
                  USER
                   │
             kubectl apply
                   │
                   ▼
            ┌──────────────┐
            │ API SERVER   │
            └──────┬───────┘
                   │
             CRD validates
                   │
                   ▼
            ┌──────────────┐
            │ Custom       │
            │ Resource     │
            └──────┬───────┘
                   │
                Watch
                   │
                   ▼
            ┌──────────────┐
            │ Informer +   │
            │ Cache        │
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Work Queue   │
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Reconcile()  │
            └──────┬───────┘
                   │
            Observe / Compare
                   │
                   ▼
            ┌──────────────┐
            │ Create /     │
            │ Update /     │
            │ Delete       │
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │ Deployment   │
            │ Service      │
            │ Job / PVC... │
            └──────┬───────┘
                   │
                   ▼
            Update Status
                   │
                   ▼
              Reconcile
                again
```

### The 8 words to remember

**Watch → Queue → Reconcile → Compare → Act → Status → Retry → Repeat**

---

## Source Coverage

This guide consolidates the uploaded material's repeated sections into one interview-oriented flow. The original file contains repeated coverage of CRDs, controller-runtime, informers/cache, RBAC, webhooks, and the complete CRD/controller implementation flow; those repeated sections have been merged here rather than repeated verbatim. fileciteturn7file1L1515-L1531
