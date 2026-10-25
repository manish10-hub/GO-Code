# Kubernetes Admission Webhooks
### Mutating & Validating Webhooks — Complete Guide + Interview Q&A

---

## Part 1 — The Big Picture

### What Problem Do Webhooks Solve?

Kubernetes has built-in validation for its own resources — you cannot
create a Pod without a container name, or a Service with an invalid port.
But Kubernetes knows nothing about **your** business rules:

```
Built-in Kubernetes cannot enforce:
  ✗ "Every pod must have a cost-centre label"
  ✗ "Container images must come from our private registry only"
  ✗ "Memory limits must always be set — no unlimited containers"
  ✗ "Inject our logging sidecar into every pod automatically"
  ✗ "Production namespace cannot have replicas < 2"
  ✗ "Container must not run as root"
  ✗ "All deployments must have resource requests set"
```

Before webhooks, enforcing these rules required:
- Hoping developers remembered (they don't)
- Catching violations in CI/CD (too late — cluster drift happens)
- Writing a custom controller to patch things after the fact (reactive, not preventive)

**Admission webhooks solve this by intercepting every API request
before it is stored in etcd — giving you a chance to reject it or
modify it at the point of entry.**



                    Pod
                     |
          ┌──────────┴──────────┐
          ↓                     ↓
      MUTATING              VALIDATING
          |                     |
          |                     |
   Add sidecar             Check image
   Add labels              Check replicas
   Add env vars            Check resources
   Set defaults            Check security
   Set resources           Check policies
          |                     |
          └──────────┬──────────┘
                     ↓
                   etcd

---

## Part 2 — The Kubernetes API Request Lifecycle

Every `kubectl apply`, `kubectl create`, or API call goes through this pipeline:

```
kubectl apply -f pod.yaml
        │
        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Kubernetes API Server                            │
│                                                                     │
│  Step 1: Authentication                                             │
│    "Who are you?" → verify certificate / token / OIDC              │
│           │                                                         │
│           ▼                                                         │
│  Step 2: Authorization (RBAC)                                       │
│    "Are you allowed to create pods in this namespace?"              │
│           │                                                         │
│           ▼                                                         │
│  Step 3: Admission Control  ◄── WEBHOOKS HAPPEN HERE               │
│    ┌──────────────────────────────────────────────┐                │
│    │  3a. Mutating Admission Webhooks             │                │
│    │      (can MODIFY the object)                 │                │
│    │            │                                 │                │
│    │            ▼                                 │                │
│    │  3b. Object Schema Validation                │                │
│    │      (openAPIV3Schema from CRD)              │                │
│    │            │                                 │                │
│    │            ▼                                 │                │
│    │  3c. Validating Admission Webhooks           │                │
│    │      (can only ACCEPT or REJECT)             │                │
│    └──────────────────────────────────────────────┘                │
│           │                                                         │
│           ▼                                                         │
│  Step 4: Persist to etcd                                           │
│    Object is stored → controllers/schedulers notified              │
└─────────────────────────────────────────────────────────────────────┘
```

**Critical ordering — Mutating runs BEFORE Validating:**

Mutating webhooks run first so that by the time validating webhooks
inspect the object, they see the **final, fully-defaulted version**.
If validating ran first, it would reject objects that are missing fields
which mutating webhooks would have added.

---

## Part 3 — Mutating Admission Webhooks

### What they are

A mutating webhook can **modify** the incoming object before it is
stored in etcd. The API server sends the object to your webhook,
your webhook modifies it, and returns a JSON Patch describing the changes.
The API server applies the patch and continues.

### What they are used for

```
1. Injecting sidecars
   → Istio injects Envoy proxy into every pod automatically
   → Your logging agent injected without developers knowing

2. Setting defaults
   → "If resources.limits is not set, default to 500m CPU / 256Mi memory"
   → "If imagePullPolicy is not set, default to Always"

3. Adding labels and annotations
   → Auto-label every pod with team name from namespace annotation
   → Inject cost-centre labels for billing

4. Normalising values
   → Convert image tags: "nginx:latest" → "nginx:1.25.3" (pinned)
   → Add registry prefix: "nginx" → "registry.company.com/nginx"

5. Injecting environment variables
   → Add standard env vars (LOG_LEVEL, ENVIRONMENT) to all pods
```

### How it works — step by step

```
User runs: kubectl apply -f pod.yaml
                    │
                    ▼
         API Server receives request
                    │
                    ▼
         API Server calls your webhook:
         POST https://my-webhook.svc:443/mutate
         {
           "apiVersion": "admission.k8s.io/v1",
           "kind": "AdmissionReview",
           "request": {
             "uid": "abc-123",
             "kind": {"group":"","version":"v1","kind":"Pod"},
             "operation": "CREATE",
             "object": { <the original pod spec> }
           }
         }
                    │
                    ▼
         Your webhook logic runs:
         - Check if sidecar already present
         - If not, prepare a JSON Patch

         Your webhook responds:
         {
           "apiVersion": "admission.k8s.io/v1",
           "kind": "AdmissionReview",
           "response": {
             "uid": "abc-123",      ← must match request uid
             "allowed": true,
             "patchType": "JSONPatch",
             "patch": "<base64 encoded JSON patch>"
           }
         }
                    │
                    ▼
         API Server applies the patch to the object
         → Pod now has the sidecar container added
                    │
                    ▼
         Object stored in etcd with modifications
```

### JSON Patch — how modifications are expressed

```json
[
  {
    "op": "add",
    "path": "/spec/containers/-",
    "value": {
      "name": "log-shipper",
      "image": "fluent/fluent-bit:2.1",
      "resources": {
        "limits": {"memory": "64Mi", "cpu": "50m"}
      }
    }
  },
  {
    "op": "add",
    "path": "/metadata/labels/injected-by",
    "value": "webhook"
  },
  {
    "op": "replace",
    "path": "/spec/containers/0/imagePullPolicy",
    "value": "Always"
  }
]
```

JSON Patch operations:
- `add` — add a field or array element
- `remove` — delete a field
- `replace` — change an existing field's value
- `copy` — copy a value from one path to another
- `move` — move a value from one path to another

---

### Go implementation — Mutating Webhook

```go
package webhook

import (
    "context"
    "encoding/json"
    "net/http"

    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// PodMutator injects a logging sidecar into every pod
type PodMutator struct {
    Client  client.Client
    Decoder *admission.Decoder
}

func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
    pod := &corev1.Pod{}

    // Decode the incoming object
    if err := m.Decoder.DecodeRaw(req.Object, pod); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    // Business logic: inject sidecar if not already present
    if !hasSidecar(pod, "log-shipper") {
        sidecar := corev1.Container{
            Name:  "log-shipper",
            Image: "fluent/fluent-bit:2.1",
            Resources: corev1.ResourceRequirements{
                Limits: corev1.ResourceList{
                    corev1.ResourceMemory: resource.MustParse("64Mi"),
                    corev1.ResourceCPU:    resource.MustParse("50m"),
                },
            },
            VolumeMounts: []corev1.VolumeMount{{
                Name:      "varlog",
                MountPath: "/var/log",
            }},
        }
        pod.Spec.Containers = append(pod.Spec.Containers, sidecar)
    }

    // Set default resource limits if not set
    for i, container := range pod.Spec.Containers {
        if container.Resources.Limits == nil {
            pod.Spec.Containers[i].Resources.Limits = corev1.ResourceList{
                corev1.ResourceMemory: resource.MustParse("256Mi"),
                corev1.ResourceCPU:    resource.MustParse("500m"),
            }
        }
    }

    // Add standard labels
    if pod.Labels == nil {
        pod.Labels = map[string]string{}
    }
    pod.Labels["injected-by"] = "webhook"
    pod.Labels["log-collection"] = "enabled"

    // Encode the modified pod and return as JSON Patch
    marshaledPod, err := json.Marshal(pod)
    if err != nil {
        return admission.Errored(http.StatusInternalServerError, err)
    }

    // PatchResponseFromRaw computes the JSON diff between original and modified
    return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func hasSidecar(pod *corev1.Pod, name string) bool {
    for _, c := range pod.Spec.Containers {
        if c.Name == name {
            return true
        }
    }
    return false
}
```

---

## Part 4 — Validating Admission Webhooks

### What they are

A validating webhook can **only accept or reject** the request — it
cannot modify the object. It runs after mutating webhooks, so it sees
the final object with all mutations already applied. If it rejects,
the request fails with a clear error message returned to the user.

### What they are used for

```
1. Policy enforcement
   → "Images must come from registry.company.com — reject others"
   → "Pods must not run as root (runAsNonRoot: true required)"
   → "Resource limits must be set on all containers"

2. Business rules
   → "Production namespace pods must have replicas >= 2"
   → "Deployment name must follow naming convention: <team>-<app>"
   → "Specific labels must be present (team, cost-centre, environment)"

3. Security policies
   → "No privileged containers allowed"
   → "No hostPath volume mounts"
   → "No hostNetwork: true"

4. Quota and governance
   → "Memory limit must not exceed 4Gi per container"
   → "Cannot deploy to prod namespace without approval annotation"

5. Configuration validation
   → "Database URL must match allowed endpoints"
   → "TLS certificate secret must exist before ingress is created"
```

### How it works — step by step

```
User runs: kubectl apply -f deployment.yaml
                    │
                    ▼
         (Mutating webhooks have already run)
         Object is now in its final mutated form
                    │
                    ▼
         API Server calls your validating webhook:
         POST https://my-webhook.svc:443/validate
         {
           "request": {
             "uid": "abc-123",
             "operation": "CREATE",
             "object": { <the fully mutated object> }
           }
         }
                    │
                    ▼
         Your webhook runs validation logic:
         - Check all containers have resource limits
         - Check image comes from approved registry
         - Check required labels are present

         CASE 1 — Passes all checks:
         Response: { "allowed": true }
         → API Server stores the object in etcd

         CASE 2 — Fails a check:
         Response: {
           "allowed": false,
           "status": {
             "code": 403,
             "message": "Container 'app' is missing resource limits.
                         All containers must specify CPU and memory limits.
                         See https://wiki.company.com/k8s-policies"
           }
         }
         → kubectl output:
         Error from server: admission webhook "validate.pods.company.com"
         denied the request: Container 'app' is missing resource limits.
```

### Go implementation — Validating Webhook

```go
package webhook

import (
    "context"
    "fmt"
    "net/http"
    "strings"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/util/validation/field"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// PodValidator enforces company policies on pods
type PodValidator struct {
    Decoder       *admission.Decoder
    AllowedRegistry string  // e.g. "registry.company.com"
}

func (v *PodValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    pod := &corev1.Pod{}
    if err := v.Decoder.DecodeRaw(req.Object, pod); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }

    // Collect ALL validation errors — don't stop at first failure
    // This gives users a complete list of what needs fixing
    var allErrors field.ErrorList

    // Rule 1: Required labels
    allErrors = append(allErrors, v.validateLabels(pod)...)

    // Rule 2: Resource limits must be set
    allErrors = append(allErrors, v.validateResourceLimits(pod)...)

    // Rule 3: Images from approved registry only
    allErrors = append(allErrors, v.validateImageRegistry(pod)...)

    // Rule 4: No privileged containers
    allErrors = append(allErrors, v.validateSecurity(pod)...)

    if len(allErrors) > 0 {
        return admission.Denied(allErrors.ToAggregate().Error())
    }

    return admission.Allowed("all validation checks passed")
}

func (v *PodValidator) validateLabels(pod *corev1.Pod) field.ErrorList {
    var errors field.ErrorList
    requiredLabels := []string{"team", "cost-centre", "environment"}

    for _, label := range requiredLabels {
        if _, ok := pod.Labels[label]; !ok {
            errors = append(errors, field.Required(
                field.NewPath("metadata", "labels", label),
                fmt.Sprintf("label '%s' is required on all pods", label),
            ))
        }
    }
    return errors
}

func (v *PodValidator) validateResourceLimits(pod *corev1.Pod) field.ErrorList {
    var errors field.ErrorList
    containersPath := field.NewPath("spec", "containers")

    for i, container := range pod.Spec.Containers {
        containerPath := containersPath.Index(i)

        if container.Resources.Limits == nil {
            errors = append(errors, field.Required(
                containerPath.Child("resources", "limits"),
                "resource limits must be set on all containers",
            ))
            continue
        }

        if _, ok := container.Resources.Limits[corev1.ResourceMemory]; !ok {
            errors = append(errors, field.Required(
                containerPath.Child("resources", "limits", "memory"),
                "memory limit is required",
            ))
        }

        if _, ok := container.Resources.Limits[corev1.ResourceCPU]; !ok {
            errors = append(errors, field.Required(
                containerPath.Child("resources", "limits", "cpu"),
                "CPU limit is required",
            ))
        }
    }
    return errors
}

func (v *PodValidator) validateImageRegistry(pod *corev1.Pod) field.ErrorList {
    var errors field.ErrorList
    containersPath := field.NewPath("spec", "containers")

    for i, container := range pod.Spec.Containers {
        if !strings.HasPrefix(container.Image, v.AllowedRegistry) {
            errors = append(errors, field.Invalid(
                containersPath.Index(i).Child("image"),
                container.Image,
                fmt.Sprintf("image must be from registry '%s'", v.AllowedRegistry),
            ))
        }
    }
    return errors
}

func (v *PodValidator) validateSecurity(pod *corev1.Pod) field.ErrorList {
    var errors field.ErrorList
    containersPath := field.NewPath("spec", "containers")

    for i, container := range pod.Spec.Containers {
        if container.SecurityContext != nil &&
            container.SecurityContext.Privileged != nil &&
            *container.SecurityContext.Privileged {
            errors = append(errors, field.Forbidden(
                containersPath.Index(i).Child("securityContext", "privileged"),
                "privileged containers are not allowed",
            ))
        }
    }
    return errors
}
```

---

## Part 5 — The Webhook Configuration YAML

### MutatingWebhookConfiguration

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: pod-mutating-webhook
  annotations:
    # cert-manager injects the CA bundle automatically when this annotation is set
    cert-manager.io/inject-ca-from: webhook-system/webhook-serving-cert
webhooks:
- name: mutate.pods.company.com           # must be a DNS subdomain

  # Which requests to intercept
  rules:
  - apiGroups:   [""]                     # core API group
    apiVersions: ["v1"]
    operations:  ["CREATE", "UPDATE"]     # only on create and update
    resources:   ["pods"]
    scope:       "Namespaced"

  # Where to send the request
  clientConfig:
    service:
      name:      webhook-service
      namespace: webhook-system
      port:      443
      path:      /mutate-pods           # your handler path
    # caBundle is injected by cert-manager annotation above

  # Which namespaces to intercept (label selector)
  namespaceSelector:
    matchExpressions:
    - key: webhook-injection
      operator: In
      values: ["enabled"]              # only namespaces with this label

  # Which objects to intercept (label selector on the object itself)
  objectSelector:
    matchExpressions:
    - key: skip-injection
      operator: DoesNotExist           # skip pods with skip-injection label

  # CRITICAL: what to do if webhook is unreachable
  failurePolicy: Fail                  # Fail = reject if unreachable
                                       # Ignore = allow if unreachable

  # Timeout for webhook response (default 10s, max 30s)
  timeoutSeconds: 5

  # Prevents calling webhook multiple times for same object
  # (when other webhooks or admission controllers also modify the object)
  reinvocationPolicy: IfNeeded         # call again if another webhook modified

  # Webhook does not have side effects outside the cluster
  sideEffects: None

  # Which AdmissionReview API version to use
  admissionReviewVersions: ["v1", "v1beta1"]
```

### ValidatingWebhookConfiguration

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pod-validating-webhook
  annotations:
    cert-manager.io/inject-ca-from: webhook-system/webhook-serving-cert
webhooks:
- name: validate.pods.company.com
  rules:
  - apiGroups:   [""]
    apiVersions: ["v1"]
    operations:  ["CREATE", "UPDATE"]
    resources:   ["pods"]
    scope:       "Namespaced"

  clientConfig:
    service:
      name:      webhook-service
      namespace: webhook-system
      port:      443
      path:      /validate-pods

  # Validating webhooks can run in parallel (all called simultaneously)
  # Unlike mutating webhooks which run sequentially
  failurePolicy: Fail
  timeoutSeconds: 5
  sideEffects: None
  admissionReviewVersions: ["v1"]

  # Exclude system namespaces — never block kube-system operations
  namespaceSelector:
    matchExpressions:
    - key: kubernetes.io/metadata.name
      operator: NotIn
      values: ["kube-system", "kube-public", "kube-node-lease"]
```

---

## Part 6 — TLS and Certificate Management

### Why TLS is mandatory

The API server only calls webhooks over **HTTPS**. This is a hard requirement —
no HTTP webhook endpoints. The API server must trust the webhook server's
TLS certificate, verified by the `caBundle` field in the webhook configuration.

### Three approaches to TLS

**Approach 1 — cert-manager (recommended):**
```yaml
# cert-manager issues and rotates the certificate automatically
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: webhook-serving-cert
  namespace: webhook-system
spec:
  dnsNames:
  - webhook-service.webhook-system.svc
  - webhook-service.webhook-system.svc.cluster.local
  issuerRef:
    kind: Issuer
    name: selfsigned-issuer
  secretName: webhook-serving-cert-tls

---
# cert-manager injects caBundle into webhook config via annotation:
# cert-manager.io/inject-ca-from: webhook-system/webhook-serving-cert
```

**Approach 2 — Manual (not recommended for production):**
```bash
# Generate CA and server cert manually
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=webhook-ca" -out ca.crt
openssl genrsa -out server.key 2048
openssl req -new -key server.key -subj "/CN=webhook-service.webhook-system.svc" -out server.csr
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -out server.crt

# Base64-encode the CA cert and put in caBundle
CA_BUNDLE=$(cat ca.crt | base64 | tr -d '\n')
kubectl patch mutatingwebhookconfiguration pod-mutating-webhook \
  --type='json' -p="[{\"op\":\"replace\",\"path\":\"/webhooks/0/clientConfig/caBundle\",\"value\":\"$CA_BUNDLE\"}]"
```

**Approach 3 — Kubernetes CA injection (for clusters with this feature):**
```yaml
# Set caBundle to empty bytes — some admission controller managers
# inject the cluster CA automatically
clientConfig:
  caBundle: ""
```

---

## Part 7 — failurePolicy — The Most Important Operational Decision

```
failurePolicy: Fail
  If the webhook is UNREACHABLE (timeout, crash, TLS error):
  → The API request is REJECTED
  → kubectl apply fails with webhook error
  → Nothing can be created/updated while webhook is down

  Use for:
  → Security-critical policies (must enforce, cannot allow bypass)
  → Sidecar injection when sidecars are required for functionality

  Risk:
  → Webhook outage = cluster-wide impact on resource creation
  → Cluster upgrade, pod restart, TLS cert expiry → everything breaks
  → Must have high availability (multiple replicas + PDB)

---

failurePolicy: Ignore
  If the webhook is UNREACHABLE:
  → The API request is ALLOWED through unchecked
  → Policy is not enforced during webhook downtime
  → Operations continue normally

  Use for:
  → Non-critical policies (nice-to-have, not security-critical)
  → Development environments where availability matters more than enforcement
  → When you cannot guarantee webhook uptime

  Risk:
  → Policy bypass is silent — you may not know enforcement failed
  → Must monitor webhook availability and alert on failures
```

**The production outage this causes:**

```
Scenario: Mutating webhook with failurePolicy: Fail
          Webhook namespace selector: {} (matches ALL namespaces)
          Webhook resource selector: pods

  1. Webhook certificate expires at 3am
  2. Kubernetes cannot establish TLS with webhook
  3. EVERY pod creation across the ENTIRE cluster fails
  4. Deployments cannot scale
  5. DaemonSet pods on new nodes cannot start
  6. kube-system pods cannot restart
  7. Cluster is effectively broken

Prevention:
  - Exclude kube-system and other critical namespaces via namespaceSelector
  - Use cert-manager for automatic certificate rotation
  - Monitor certificate expiry (alert 30 days before)
  - Run webhook with multiple replicas + PodDisruptionBudget
  - Set appropriate timeoutSeconds (5s, not 30s)
```

---

## Part 8 — Webhook Registration with controller-runtime

```go
// In cmd/main.go or setup code

func main() {
    mgr, _ := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        // Webhook server configuration
        WebhookServer: webhook.NewServer(webhook.Options{
            Port:    9443,
            TLSOpts: []func(*tls.Config){}, // TLS configured via cert-manager
        }),
    })

    // Register mutating webhook
    mgr.GetWebhookServer().Register(
        "/mutate-pods",
        &webhook.Admission{Handler: &PodMutator{
            Client:  mgr.GetClient(),
            Decoder: admission.NewDecoder(mgr.GetScheme()),
        }},
    )

    // Register validating webhook
    mgr.GetWebhookServer().Register(
        "/validate-pods",
        &webhook.Admission{Handler: &PodValidator{
            Decoder:         admission.NewDecoder(mgr.GetScheme()),
            AllowedRegistry: "registry.company.com",
        }},
    )

    // Or use kubebuilder markers on your types for auto-registration:
    // +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,...
    // +kubebuilder:webhook:path=/validate-v1-pod,mutating=false,failurePolicy=fail,...
}
```

---

## Part 9 — Mutating vs Validating — Side by Side

```
┌─────────────────────────────────┬─────────────────────────────────┐
│    MUTATING WEBHOOK             │    VALIDATING WEBHOOK           │
├─────────────────────────────────┼─────────────────────────────────┤
│ Can modify the object?   YES    │ Can modify the object?   NO     │
│ Can reject the request?  YES    │ Can reject the request?  YES    │
├─────────────────────────────────┼─────────────────────────────────┤
│ Runs FIRST (before validating)  │ Runs SECOND (after mutating)    │
├─────────────────────────────────┼─────────────────────────────────┤
│ Runs SEQUENTIALLY               │ Runs in PARALLEL                │
│ (each webhook's output is the   │ (all validating webhooks called │
│  next webhook's input)          │  simultaneously for speed)      │
├─────────────────────────────────┼─────────────────────────────────┤
│ Response: JSON Patch            │ Response: allowed true/false    │
├─────────────────────────────────┼─────────────────────────────────┤
│ Use for:                        │ Use for:                        │
│  - Sidecar injection            │  - Policy enforcement           │
│  - Setting defaults             │  - Security validation          │
│  - Adding labels/annotations    │  - Business rule enforcement    │
│  - Normalizing image tags       │  - Required field validation    │
│  - Injecting env vars           │  - Cross-field validation       │
└─────────────────────────────────┴─────────────────────────────────┘
```

---

## Part 10 — Real-World Examples

### Istio Sidecar Injection (Mutating)
Istio's `istio-sidecar-injector` webhook intercepts every pod creation
in namespaces labeled `istio-injection: enabled`. It injects:
- The Envoy proxy sidecar container
- An init container that configures iptables rules
- Volume mounts for certificates and configuration

The developer creates a plain pod YAML — Istio's webhook transparently
transforms it into a service mesh-enabled pod without any developer action.

### OPA/Gatekeeper Policy Enforcement (Validating)
Open Policy Agent Gatekeeper installs validating webhooks that enforce
policies written in Rego. Organizations define policies like:
- "All namespaces must have a cost-centre label"
- "Images must not use the :latest tag"
- "PodSecurityPolicy: no privileged containers"

When a resource violates a policy, the webhook returns a detailed
rejection message telling the developer exactly what to fix.

### cert-manager (Both)
cert-manager uses a mutating webhook to inject default values into
Certificate resources and a validating webhook to ensure Certificate
specs are valid before storing them.

---

## Interview Questions — Admission Webhooks

---

### Beginner

**Q1. What is an admission webhook in Kubernetes?**

An admission webhook is an HTTP callback that the Kubernetes API server
calls during the admission phase of request processing — after authentication
and authorization, but before the object is stored in etcd. Webhooks let
you plug custom logic into the API server to either modify incoming objects
(mutating) or validate and accept/reject them (validating). They solve the
problem of enforcing custom business rules, security policies, and operational
standards that Kubernetes cannot enforce natively.

---

**Q2. What is the difference between a mutating and validating webhook?**

A mutating webhook can modify the incoming object — it receives the object,
transforms it (adds containers, sets defaults, injects labels), and returns
a JSON Patch describing the changes. A validating webhook can only accept or
reject — it cannot change the object, only inspect it and decide whether it
complies with policy. Mutating webhooks always run before validating webhooks
because validators must see the final, mutated version of the object. Mutating
webhooks run sequentially (each sees the previous webhook's output), while
validating webhooks run in parallel.

---

**Q3. Why must admission webhooks use HTTPS? How is TLS managed?**

The Kubernetes API server only trusts HTTPS endpoints — it validates the
webhook server's TLS certificate against the `caBundle` field in the
webhook configuration. This is mandatory because the API server is sending
potentially sensitive resource data to the webhook; unencrypted HTTP would
expose this data and allow man-in-the-middle attacks. In production, TLS
is typically managed with cert-manager, which automatically issues and
rotates the certificate and injects the CA bundle into the webhook
configuration via an annotation.

---

### Intermediate

**Q4. Explain `failurePolicy: Fail` vs `failurePolicy: Ignore`. When would you use each and what are the risks?**

`failurePolicy: Fail` means if the webhook is unreachable (timeout, crash,
TLS error), the API server rejects the request. This guarantees policy is
always enforced but means a webhook outage blocks all matching resource
operations cluster-wide — a serious availability risk. Use it for
security-critical policies where bypassing enforcement is unacceptable.

`failurePolicy: Ignore` means if the webhook is unreachable, the request
is allowed through without being checked. Operations continue normally during
webhook downtime but policy enforcement silently fails. Use it for non-critical
policies where availability matters more than strict enforcement.

The most important operational practice with `Fail`: always exclude
`kube-system` and other critical namespaces via `namespaceSelector` to
prevent webhook failures from breaking cluster operations. A webhook with
`failurePolicy: Fail` and no namespace exclusions that goes down can make
it impossible to create any pods — including the webhook pod itself, creating
a deadlock.

---

**Q5. Why do mutating webhooks run sequentially but validating webhooks run in parallel?**

Mutating webhooks must run sequentially because each webhook may see and
modify the object — the output of one webhook becomes the input of the next.
If they ran in parallel, two webhooks could produce conflicting patches
(both trying to set the same field to different values) with no defined winner.
Sequential execution provides a consistent, deterministic pipeline.

Validating webhooks only read the object — they never modify it, so there
is no conflict between parallel executions. Running them in parallel reduces
total latency: if you have five validating webhooks each taking 50ms,
sequential execution would take 250ms while parallel execution takes only
50ms (the slowest webhook's duration).

---

**Q6. What is `reinvocationPolicy: IfNeeded` and when is it necessary?**

When multiple mutating webhooks run sequentially, a later webhook's
modifications might affect fields that an earlier webhook already processed.
With `reinvocationPolicy: IfNeeded`, if any webhook modifies the object,
all previous webhooks are called again with the modified object to ensure
their logic is still consistent.

Example: Webhook A injects a sidecar. Webhook B adds a volume needed by
all containers including sidecars. After Webhook B runs, the sidecar
added by Webhook A does not have the volume mount. With `reinvocationPolicy:
IfNeeded`, Webhook A is called again with the modified pod (now including
the volume from Webhook B), and can add the volume mount to the sidecar.
Without reinvocation, the sidecar would be missing the required volume mount.

---

**Q7. How do you prevent a webhook from accidentally blocking kube-system operations and causing a cluster deadlock?**

Always configure `namespaceSelector` to exclude system namespaces:

```yaml
namespaceSelector:
  matchExpressions:
  - key: kubernetes.io/metadata.name
    operator: NotIn
    values:
    - kube-system
    - kube-public
    - kube-node-lease
    - your-webhook-namespace   # exclude the webhook's own namespace!
```

The last exclusion is critical — if your webhook's own namespace is not
excluded, and the webhook pod crashes, creating the replacement webhook
pod would itself be blocked by the failing webhook — a deadlock where the
cluster cannot create the pod needed to fix the webhook because the broken
webhook is blocking pod creation.

Additionally: set `failurePolicy: Ignore` for non-critical webhooks, use
`timeoutSeconds: 5` (not 30) to fail fast, run multiple webhook replicas
with a PodDisruptionBudget, and monitor certificate expiry.

---

### Expert

**Q8. How does Istio implement transparent sidecar injection using a mutating webhook? What exactly gets injected and why?**

Istio's `istio-sidecar-injector` MutatingWebhookConfiguration intercepts
pod creations in namespaces with the `istio-injection: enabled` label.

The webhook injects three components:

**1. An init container (`istio-init`):**
Runs before the application container and configures iptables rules inside
the pod's shared network namespace. These rules redirect all inbound TCP
traffic to port 15006 (Envoy's inbound listener) and all outbound TCP
traffic to port 15001 (Envoy's outbound listener). The application container
cannot bypass this — iptables rules apply to the entire shared network namespace.

**2. The Envoy sidecar container (`istio-proxy`):**
Runs alongside the application, intercepting all traffic via the iptables
rules. Handles mTLS, retries, circuit breaking, telemetry, and traffic shaping.
Reads its configuration from Istiod via xDS APIs.

**3. Volume mounts:**
Mounts the pod's service account token (for mTLS certificate issuance),
the Istio root CA certificate, and configuration volumes.

The application developer writes a plain pod spec. The webhook transparently
transforms it. The application believes it is making direct network connections
— it has no knowledge of Envoy.

---

**Q9. Design a webhook that enforces a policy: "Any deployment to the production namespace requires an annotation `approved-by` with a valid approver name from an approved-approvers ConfigMap." What challenges do you face and how do you solve them?**

**The implementation:**

```go
func (v *ApprovalValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    // Only enforce on production namespace
    if req.Namespace != "production" {
        return admission.Allowed("not production namespace")
    }

    // Only on CREATE and UPDATE operations
    if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
        return admission.Allowed("not a create or update operation")
    }

    deploy := &appsv1.Deployment{}
    v.Decoder.DecodeRaw(req.Object, deploy)

    approver, ok := deploy.Annotations["approved-by"]
    if !ok || approver == "" {
        return admission.Denied("deployments to production must have 'approved-by' annotation")
    }

    // Fetch approved approvers from ConfigMap
    cm := &corev1.ConfigMap{}
    v.Client.Get(ctx, types.NamespacedName{
        Name:      "approved-approvers",
        Namespace: "webhook-system",
    }, cm)

    approvers := strings.Split(cm.Data["approvers"], ",")
    for _, a := range approvers {
        if strings.TrimSpace(a) == approver {
            return admission.Allowed(fmt.Sprintf("approved by %s", approver))
        }
    }

    return admission.Denied(fmt.Sprintf("'%s' is not an approved approver", approver))
}
```

**Challenges:**

**1. The webhook needs RBAC to read the ConfigMap:**
```yaml
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get"]
```

**2. ConfigMap changes must not require webhook restart:**
Use the controller-runtime client which reads from the informer cache —
ConfigMap changes are reflected automatically without restarting.

**3. UPDATE operations — prevent removing the annotation:**
Check `req.OldObject` (the previous version) — if it had the annotation
and the new version removes it, reject. This prevents post-approval
annotation stripping.

**4. The bootstrap problem:**
The `approved-approvers` ConfigMap itself lives in `webhook-system`.
If the webhook validates Deployments in `webhook-system`, you have a
chicken-and-egg problem — deploying the webhook requires approval from
the webhook. Solution: exclude `webhook-system` from the webhook's scope.

**5. Webhook availability:**
If this webhook uses `failurePolicy: Fail`, a webhook outage blocks all
production deployments — potentially more damaging than the problem it solves.
Use `failurePolicy: Ignore` with strong monitoring, or accept the trade-off
of `Fail` with an on-call runbook for webhook recovery.

---

## Quick Reference — Key Fields

| Field | Values | Meaning |
|---|---|---|
| `failurePolicy` | `Fail` / `Ignore` | What to do if webhook is unreachable |
| `timeoutSeconds` | 1-30 (default 10) | How long to wait for webhook response |
| `sideEffects` | `None` / `NoneOnDryRun` | Does webhook modify things outside the cluster |
| `reinvocationPolicy` | `Never` / `IfNeeded` | Should webhook be called again if another webhook modified the object |
| `operations` | `CREATE` `UPDATE` `DELETE` `CONNECT` | Which operations to intercept |
| `scope` | `Namespaced` / `Cluster` / `*` | Which resource scope to intercept |
| `namespaceSelector` | Label selector | Which namespaces to intercept |
| `objectSelector` | Label selector | Which objects (by label) to intercept |
| `admissionReviewVersions` | `["v1", "v1beta1"]` | API version for AdmissionReview objects |