# Kubernetes Dynamic Volume Provisioning
### Complete Guide + Interview Q&A (Beginner → Expert)

---

## Part 1 — The Problem: Why Dynamic Provisioning Exists

### The World Before Dynamic Provisioning (Static Provisioning)

Before dynamic provisioning, a cluster administrator had to manually
create storage before any application could use it. The workflow looked like this:

```
Step 1 — Admin creates storage in the cloud:
  aws ec2 create-volume --size 10 --availability-zone us-east-1a
  → Returns: vol-0abc123def456789

Step 2 — Admin creates a PersistentVolume (PV) manifest:
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: my-pv-001
  spec:
    capacity:
      storage: 10Gi
    accessModes: [ReadWriteOnce]
    awsElasticBlockStore:
      volumeID: vol-0abc123def456789
      fsType: ext4

Step 3 — Developer creates a PVC:
  Kubernetes matches PVC to the pre-created PV

Step 4 — Developer deploys application
```

**Problems with this approach:**

```
1. Operational bottleneck:
   Every storage request requires admin action.
   100 developers need storage → 100 manual admin tasks.

2. Resource waste:
   Admin pre-creates volumes in advance (guessing what devs need).
   Pre-created volumes sit idle if sizes don't match.

3. No self-service:
   Developer cannot deploy without admin creating storage first.
   CI/CD pipelines block on admin approvals.

4. Human error:
   Manual processes = mistakes (wrong size, wrong zone, wrong type).

5. Does not scale:
   Fine for 10 apps, impossible for 1000.
```

### The Solution: Dynamic Provisioning

```
Developer creates a PVC → Kubernetes automatically:
  1. Calls the cloud provider API
  2. Creates the actual storage (EBS volume, GCP disk, Azure disk)
  3. Creates the PV object in Kubernetes
  4. Binds the PVC to the PV

Developer gets storage without waiting for admin.
Admin defines POLICY (StorageClass) once → developers self-serve.
```

---

## Part 2 — The Four Core Components

### Component 1 — PersistentVolume (PV)

```
What it is:
  A piece of actual storage that exists in the cluster.
  It is a cluster-level resource (not namespaced).
  It represents the real storage: an EBS volume, a GCP disk, an NFS share.

Who creates it:
  Static provisioning  → Admin creates it manually
  Dynamic provisioning → Kubernetes creates it automatically

Key fields:
  capacity.storage    → how big is the storage (10Gi)
  accessModes         → how can pods use it (ReadWriteOnce, ReadWriteMany, ReadOnlyMany)
  reclaimPolicy       → what happens when PVC is deleted (Retain, Delete, Recycle)
  storageClassName    → which StorageClass created this PV
  volumeMode          → Block or Filesystem
  status              → Available, Bound, Released, Failed
```

```yaml
# Example PV (auto-created by dynamic provisioning)
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pvc-abc123-def456        # auto-generated name
  annotations:
    pv.kubernetes.io/provisioned-by: ebs.csi.aws.com
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  reclaimPolicy: Delete           # auto-set from StorageClass
  storageClassName: gp3-encrypted
  persistentVolumeReclaimPolicy: Delete
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: vol-0abc123def456789   # actual AWS volume ID
    fsType: ext4
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: topology.kubernetes.io/zone
          operator: In
          values:
          - us-east-1a             # EBS volumes are zone-specific
```

### Component 2 — PersistentVolumeClaim (PVC)

```
What it is:
  A REQUEST for storage made by a developer.
  It is namespaced (belongs to a namespace).
  Like a "ticket" saying: "I need 10Gi of fast SSD storage."

Who creates it:
  The developer (or a StatefulSet's volumeClaimTemplates).

What it does:
  Kubernetes finds or creates a PV that satisfies the request.
  The PVC binds to the PV — now the developer "owns" that storage.
  The Pod then mounts the PVC to use the actual storage.
```

```yaml
# Developer creates this PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-database-storage
  namespace: production
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: gp3-encrypted    # which StorageClass to use
  resources:
    requests:
      storage: 10Gi                  # how much storage I need
  # Optional: label selector to target specific PVs (static provisioning only)
  # selector:
  #   matchLabels:
  #     environment: production
```

### Component 3 — StorageClass

```
What it is:
  A template/policy that defines HOW to provision storage.
  Set by the admin ONCE. Used by developers MANY times.
  Like a "menu item" — defines what kind of storage you can order.

Who creates it:
  The cluster administrator.

What it defines:
  - provisioner: which plugin creates the storage (ebs.csi.aws.com)
  - parameters: what kind of storage (SSD vs HDD, encrypted, IOPS)
  - reclaimPolicy: what happens to storage when PVC is deleted
  - volumeBindingMode: when to provision (immediately or when pod is scheduled)
  - allowVolumeExpansion: can PVC size be increased later
  - allowedTopologies: which zones to provision in
```

```yaml
# Admin creates StorageClasses once
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3-encrypted
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"  # make this the default
provisioner: ebs.csi.aws.com          # which CSI driver to use
parameters:
  type: gp3                            # AWS EBS volume type
  encrypted: "true"                    # encrypt at rest
  iops: "3000"                         # IOPS for gp3
  throughput: "125"                    # MB/s throughput
reclaimPolicy: Delete                  # Delete AWS volume when PVC is deleted
volumeBindingMode: WaitForFirstConsumer # Don't provision until pod is scheduled
allowVolumeExpansion: true             # Allow PVC resize
allowedTopologies:
- matchLabelExpressions:
  - key: topology.kubernetes.io/zone
    operator: In
    values:
    - us-east-1a
    - us-east-1b
    - us-east-1c
```

### Component 4 — CSI Driver (Container Storage Interface)

```
What it is:
  The plugin that actually talks to the storage backend (AWS, GCP, etc.)
  CSI is a STANDARD interface — one spec, many implementations.
  Before CSI: storage code was baked into Kubernetes (hard to update).
  After CSI: storage providers ship their own plugin independently.

Who creates it:
  Storage providers (AWS, GCP, Azure, NetApp, Ceph, Portworx, etc.)
  Deployed as pods in the cluster (usually in kube-system namespace).

What it does:
  Kubernetes says: "Create a 10Gi volume in us-east-1a"
  CSI driver calls: aws ec2 create-volume --size 10 --az us-east-1a
  CSI driver says: "Done, here's the volume ID: vol-0abc123"

CSI architecture — three key sidecars:
  external-provisioner: watches PVCs, triggers CreateVolume
  external-attacher: watches for pods, triggers ControllerPublishVolume
  node-driver-registrar: registers the CSI driver on each node
```

```
CSI Driver Components:

┌──────────────────────────────────────────────────────────┐
│                kube-system namespace                      │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  CSI Controller StatefulSet                      │    │
│  │                                                  │    │
│  │  ┌──────────────────────┐  ┌─────────────────┐  │    │
│  │  │ external-provisioner │  │ external-attacher│  │    │
│  │  │ (watches PVCs,       │  │ (attaches volume │  │    │
│  │  │  calls CreateVolume) │  │  to node)        │  │    │
│  │  └──────────────────────┘  └─────────────────┘  │    │
│  │                                                  │    │
│  │  ┌────────────────────────────────────────────┐  │    │
│  │  │ ebs-plugin (ControllerServer)              │  │    │
│  │  │ Talks to AWS API: create/delete/attach EBS │  │    │
│  │  └────────────────────────────────────────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │  CSI Node DaemonSet (runs on EVERY node)         │    │
│  │                                                  │    │
│  │  ┌──────────────────────┐  ┌─────────────────┐  │    │
│  │  │ node-driver-registrar│  │ ebs-plugin      │  │    │
│  │  │ (registers with      │  │ (NodeServer)    │  │    │
│  │  │  kubelet on node)    │  │ (mounts volume  │  │    │
│  │  └──────────────────────┘  │  on the node)   │  │    │
│  │                             └─────────────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
```

---

## Part 3 — The Complete End-to-End Flow

This is the most important section to understand deeply.

```
Developer creates a PVC → Kubernetes provisions storage automatically

Here is EXACTLY what happens, step by step:
```

### Visual Overview

```
┌──────────┐   1.create PVC    ┌─────────────┐
│Developer │──────────────────►│ API Server   │
└──────────┘                   └──────┬───────┘
                                      │ stores PVC in etcd
                                      │ PVC status: Pending
                                      │
                                      │ 2. notifies
                                      ▼
                              ┌───────────────────┐
                              │external-provisioner│
                              │(CSI sidecar)       │
                              └────────┬───────────┘
                                       │ 3. calls
                                       ▼
                              ┌───────────────────┐
                              │ CSI ControllerServer│
                              │ (ebs-plugin)       │
                              └────────┬───────────┘
                                       │ 4. calls AWS API
                                       ▼
                              ┌───────────────────┐
                              │   AWS EBS API      │
                              │ CreateVolume()     │
                              └────────┬───────────┘
                                       │ 5. volume created
                                       │ vol-0abc123
                                       │
                              ┌────────▼───────────┐
                              │external-provisioner │
                              │ creates PV object   │
                              └────────┬────────────┘
                                       │ 6. PV created
                                       ▼
                              ┌────────────────────┐
                              │ PV/PVC Controller  │
                              │(kube-controller-mgr)│
                              └────────┬────────────┘
                                       │ 7. binds PVC to PV
                                       │ PVC status: Bound
                                       ▼
                              ┌────────────────────┐
                              │   Pod Scheduler    │
                              └────────┬───────────┘
                                       │ 8. schedules pod to node-1
                                       ▼
                              ┌────────────────────┐
                              │  external-attacher  │
                              └────────┬────────────┘
                                       │ 9. ControllerPublishVolume
                                       │    (attach EBS to node-1)
                                       ▼
                              ┌────────────────────┐
                              │    kubelet          │
                              │   (on node-1)       │
                              └────────┬────────────┘
                                       │ 10. NodeStageVolume
                                       │     (format disk, mount to staging path)
                                       │ 11. NodePublishVolume
                                       │     (bind mount into pod's volume path)
                                       ▼
                              ┌────────────────────┐
                              │  Pod running with  │
                              │  /data mounted      │
                              └────────────────────┘
```

### Step-by-Step Detailed Explanation

---

**Step 1 — Developer creates a PVC**

```yaml
kubectl apply -f pvc.yaml
```

The PVC is stored in etcd. Status is `Pending`.
Nothing happens yet — just a request sitting in the system.

---

**Step 2 — External Provisioner notices the PVC**

The `external-provisioner` sidecar container (running alongside the CSI plugin)
watches the Kubernetes API for new PVCs. It sees the new `Pending` PVC and checks:

```
- Is there a StorageClass specified? Yes: gp3-encrypted
- Does this provisioner handle gp3-encrypted? Yes: ebs.csi.aws.com
- Is volumeBindingMode WaitForFirstConsumer? If yes, wait for pod to be scheduled first
- Is there already a matching PV? No → need to create one
```

---

**Step 3 — External Provisioner calls CSI CreateVolume**

The external-provisioner calls the CSI plugin's `CreateVolume` RPC:

```
CreateVolume request:
  name: "pvc-abc123-def456"
  capacity_range: { required_bytes: 10Gi }
  volume_capabilities: [{ access_mode: SINGLE_NODE_WRITER }]
  parameters: { type: gp3, encrypted: true, iops: 3000 }
  accessibility_requirements:
    preferred: [{ topology: {zone: us-east-1a} }]
```

---

**Step 4 — CSI Plugin calls Cloud API**

The EBS CSI plugin translates the CreateVolume request into an AWS API call:

```
aws ec2 create-volume \
  --size 10 \
  --volume-type gp3 \
  --encrypted \
  --iops 3000 \
  --availability-zone us-east-1a
→ Returns: { VolumeId: "vol-0abc123def456789", State: "creating" }

Waits for state = "available"
→ Volume is ready
```

---

**Step 5 — CSI Plugin responds with volume info**

```
CreateVolume response:
  volume:
    volume_id: "vol-0abc123def456789"
    capacity_bytes: 10737418240  (10Gi)
    accessible_topology: [{ zone: us-east-1a }]
```

---

**Step 6 — External Provisioner creates the PV object**

The external-provisioner creates a PersistentVolume in Kubernetes:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pvc-abc123-def456
spec:
  capacity:
    storage: 10Gi
  accessModes: [ReadWriteOnce]
  reclaimPolicy: Delete
  storageClassName: gp3-encrypted
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: vol-0abc123def456789
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: topology.kubernetes.io/zone
          operator: In
          values: [us-east-1a]
```

---

**Step 7 — PV/PVC Controller binds them**

The PersistentVolumeController (running in kube-controller-manager) sees
the new PV and matches it to the pending PVC. It binds them together:

```
PVC status: Pending → Bound
PV  status: Available → Bound

PVC.spec.volumeName = "pvc-abc123-def456"
PV.spec.claimRef = {namespace: production, name: my-database-storage}
```

---

**Step 8 — Pod is scheduled to a node**

The Kubernetes scheduler assigns the pod to a node (e.g., node-1).

**Important for `WaitForFirstConsumer`:**
With this binding mode, step 2-6 are delayed until NOW (after the pod
is scheduled). This ensures the volume is created in the SAME zone as the
node where the pod will run. Without this, you could have a pod on a node
in us-east-1a trying to use a volume created in us-east-1b — EBS is zone-specific.

---

**Step 9 — External Attacher attaches the volume to the node**

The `external-attacher` sees the pod is scheduled and calls
`ControllerPublishVolume` on the CSI plugin:

```
ControllerPublishVolume request:
  volume_id: "vol-0abc123def456789"
  node_id: "i-0node1instanceid"    ← the EC2 instance ID of node-1
  volume_capability: { access_mode: SINGLE_NODE_WRITER }

AWS API call:
  aws ec2 attach-volume \
    --volume-id vol-0abc123def456789 \
    --instance-id i-0node1instanceid \
    --device /dev/xvdba

→ EBS volume is now attached to the node
→ Node can see it as /dev/xvdba (or similar)
```

Kubernetes records this in a `VolumeAttachment` object.

---

**Step 10 — Kubelet stages the volume (NodeStageVolume)**

The kubelet on node-1 calls `NodeStageVolume` on the CSI node plugin:

```
NodeStageVolume:
  volume_id: "vol-0abc123def456789"
  staging_target_path: /var/lib/kubelet/plugins/ebs/pvc-abc123/globalmount

CSI node plugin:
  1. Finds the device: /dev/xvdba
  2. Formats it: mkfs.ext4 /dev/xvdba  (only on first use)
  3. Mounts it to staging path: mount /dev/xvdba /var/lib/kubelet/plugins/ebs/pvc-abc123/globalmount
```

This "global mount" (staging) is shared if multiple containers on the
same node use the same volume — format happens only once.

---

**Step 11 — Kubelet publishes the volume into the pod (NodePublishVolume)**

The kubelet calls `NodePublishVolume`:

```
NodePublishVolume:
  volume_id: "vol-0abc123def456789"
  staging_target_path: /var/lib/kubelet/plugins/ebs/pvc-abc123/globalmount
  target_path: /var/lib/kubelet/pods/<pod-uid>/volumes/pvc-abc123/mount

CSI node plugin:
  Creates a bind mount from staging path to pod's volume path:
  mount --bind /var/lib/kubelet/plugins/.../globalmount
               /var/lib/kubelet/pods/<pod-uid>/volumes/.../mount
```

---

**Step 12 — Pod starts, sees /data**

The container's `/data` directory is now backed by the EBS volume.
The pod is running with persistent storage.

```
Inside the pod:
  /data → bind mount → global mount → /dev/xvdba → EBS vol-0abc123
```

---

## Part 4 — Storage Class Details

### volumeBindingMode — One of the Most Important Fields

```
Immediate (default):
  Volume is provisioned AS SOON AS the PVC is created.
  Problem: Volume created in zone us-east-1a.
           Pod scheduled to node in us-east-1b.
           EBS volumes are zone-specific → pod cannot start!
  Use when: Storage is zone-independent (NFS, Ceph, EFS)

WaitForFirstConsumer:
  Volume is NOT provisioned when PVC is created.
  Waits until a pod using the PVC is scheduled to a node.
  Provisions volume in the SAME zone as the scheduled node.
  Use when: Storage is zone-specific (EBS, GCP PD, Azure Disk)
```

### reclaimPolicy — What Happens to Storage When PVC is Deleted

```
Delete (default for dynamic provisioning):
  When PVC is deleted:
  → PVC object deleted
  → PV object deleted
  → Actual storage (EBS volume) DELETED
  Use for: temporary data, dev environments
  Risk: data is GONE if PVC accidentally deleted

Retain:
  When PVC is deleted:
  → PVC object deleted
  → PV status changes to "Released"
  → Actual storage (EBS volume) KEPT
  → Admin must manually clean up PV and storage
  Use for: production databases, important data
  Benefit: accidental deletion cannot destroy data

Recycle (deprecated):
  Wipes the data and makes volume available again.
  Not recommended — use Delete instead.
```

### Access Modes — How Pods Can Use the Volume

```
ReadWriteOnce (RWO):
  One node can mount the volume read-write.
  Multiple pods ON THE SAME NODE can read-write.
  Different nodes cannot mount simultaneously.
  Most block storage: EBS, GCP PD, Azure Disk.
  Use for: databases, single-instance apps.

ReadWriteMany (RWX):
  Multiple nodes can mount read-write simultaneously.
  Requires shared/distributed storage.
  NFS, CephFS, AWS EFS, Azure Files.
  Use for: web app uploads, shared config, ML training data.

ReadOnlyMany (ROX):
  Multiple nodes can mount read-only simultaneously.
  Useful for shared reference data.

ReadWriteOncePod (RWOP) — Kubernetes 1.22+:
  Only ONE POD (not just one node) can mount read-write.
  Stricter than RWO.
  Prevents two pods on same node from both writing.
```

---

## Part 5 — Volume Lifecycle (Delete and Reclaim)

### When a Pod is Deleted (Volume Stays)

```
Pod deleted → volume is unmounted from pod → volume stays attached to node
            → next pod that uses same PVC gets same data
```

### When a PVC is Deleted (Depends on reclaimPolicy)

```
PVC deleted

If reclaimPolicy: Delete:
  Kubernetes calls CSI DeleteVolume
  AWS: aws ec2 delete-volume --volume-id vol-0abc123
  PV object deleted from Kubernetes
  EBS volume gone forever

If reclaimPolicy: Retain:
  PV status → Released
  EBS volume still exists in AWS
  Admin sees: kubectl get pv → pvc-abc123 Released
  Admin must manually clean up:
    1. kubectl delete pv pvc-abc123
    2. aws ec2 delete-volume --volume-id vol-0abc123
  Or admin can manually bind a new PVC to the retained PV
```

### Volume Expansion (Resize)

```
Requires: allowVolumeExpansion: true in StorageClass

To resize:
  kubectl patch pvc my-pvc -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'

What happens:
  1. Kubernetes sees PVC storage request increased
  2. Calls CSI ControllerExpandVolume → expand EBS volume to 20Gi
  3. Calls CSI NodeExpandVolume → resize filesystem on the node
  4. Pod now sees 20Gi storage

Important:
  You can ONLY increase, never decrease PVC storage.
  Some storage types require pod restart for filesystem resize.
  EBS with ext4/xfs supports online resize (no restart needed in recent versions).
```

---

## Part 6 — Common StorageClass Examples

### AWS EBS (Most Common)

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: aws-ebs-gp3
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  encrypted: "true"
  kmsKeyId: "arn:aws:kms:us-east-1:123456789:key/abc-123"
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

### GCP Persistent Disk

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gcp-ssd
provisioner: pd.csi.storage.gke.io
parameters:
  type: pd-ssd
  replication-type: none
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

### Ceph RBD (On-premises)

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ceph-rbd
provisioner: rbd.csi.ceph.com
parameters:
  clusterID: abc-ceph-cluster-id
  pool: kubernetes
  imageFormat: "2"
  imageFeatures: layering
  csi.storage.k8s.io/provisioner-secret-name: csi-rbd-secret
  csi.storage.k8s.io/controller-expand-secret-name: csi-rbd-secret
reclaimPolicy: Delete
allowVolumeExpansion: true
```

### Multiple StorageClasses for Different Use Cases

```yaml
# For databases (high IOPS, expensive)
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: database-storage
provisioner: ebs.csi.aws.com
parameters:
  type: io2
  iops: "10000"
  encrypted: "true"
reclaimPolicy: Retain        # never auto-delete database storage
volumeBindingMode: WaitForFirstConsumer

# For logs/cache (cheap, ok to lose)
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: cheap-storage
provisioner: ebs.csi.aws.com
parameters:
  type: st1                  # HDD, cheap, high throughput
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer

# For shared access (multiple pods read/write)
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: shared-storage
provisioner: efs.csi.aws.com  # AWS EFS supports ReadWriteMany
reclaimPolicy: Delete
volumeBindingMode: Immediate   # EFS is region-wide, not zone-specific
```

---

## Part 7 — StatefulSet and Dynamic Provisioning

StatefulSets use `volumeClaimTemplates` to automatically create
a PVC per pod using dynamic provisioning:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres-headless
  replicas: 3
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16
        volumeMounts:
        - name: data          # matches volumeClaimTemplates name
          mountPath: /var/lib/postgresql/data

  # Dynamic provisioning: creates one PVC per pod automatically
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ReadWriteOnce]
      storageClassName: database-storage   # uses Retain policy
      resources:
        requests:
          storage: 50Gi

# Results in:
# postgres-0 → PVC: data-postgres-0 → PV → 50Gi EBS volume
# postgres-1 → PVC: data-postgres-1 → PV → 50Gi EBS volume
# postgres-2 → PVC: data-postgres-2 → PV → 50Gi EBS volume
```

---

## Part 8 — Interview Questions and Answers

---

## Beginner Level

---

### Q1. What is dynamic volume provisioning in Kubernetes and why do we need it?

Dynamic volume provisioning is the ability of Kubernetes to automatically
create storage (like an AWS EBS volume or a GCP disk) when an application
needs it, without requiring a cluster administrator to manually create
storage beforehand.

Before dynamic provisioning, an admin had to manually create physical storage,
write a PersistentVolume YAML file, and only then could a developer use it.
This was slow, did not scale, and blocked developer self-service.

With dynamic provisioning, a developer just creates a PVC saying "I need
10Gi of fast SSD storage." Kubernetes automatically calls the cloud provider
API, creates the actual storage, and makes it available to the pod — all
without any admin intervention. The admin only needs to define StorageClasses
once, like a menu of storage options, and developers self-serve.

---

### Q2. What are PV, PVC, and StorageClass? Explain the relationship between them.

Think of it like ordering food at a restaurant:

**StorageClass** is the menu item — it defines the type and quality of
storage available. The admin creates these once. Examples: "fast-ssd"
(NVMe, expensive), "standard-hdd" (cheap, slow), "shared-nfs" (multiple
pods can access).

**PVC (PersistentVolumeClaim)** is the order — a developer says "I want
10Gi of fast-ssd storage." It is namespaced and belongs to one team or app.

**PV (PersistentVolume)** is the actual storage that fulfills the order.
With dynamic provisioning, it is created automatically when the PVC is
created. It is cluster-level (not namespaced) and represents the real
physical storage.

The relationship: a PVC requests storage from a StorageClass → dynamic
provisioning creates a PV → the PVC binds to the PV → the pod mounts
the PVC to get access to the actual storage.

---

### Q3. What is a StorageClass and what are its key fields?

A StorageClass is a Kubernetes object that defines a "class" of storage
and how it should be provisioned. Think of it as a policy template that
the admin creates once and developers use many times.

Key fields:

**provisioner:** which CSI driver or plugin creates the storage.
For AWS: `ebs.csi.aws.com`. For GCP: `pd.csi.storage.gke.io`.

**parameters:** options passed to the provisioner — like the disk type
(gp3, io2, ssd), whether to encrypt, IOPS settings.

**reclaimPolicy:** what happens to the actual storage when the PVC is deleted.
`Delete` = destroy the storage (safe for dev). `Retain` = keep the storage
(safe for production databases).

**volumeBindingMode:** `Immediate` = create storage immediately when PVC
is created. `WaitForFirstConsumer` = wait until a pod is scheduled, then
create storage in the same zone as the pod.

**allowVolumeExpansion:** whether you can increase the PVC size after creation.

---

### Q4. What is the difference between static and dynamic provisioning?

**Static provisioning:**
An admin manually creates the physical storage (e.g., an EBS volume in AWS
console), then writes a PersistentVolume YAML file pointing to that storage,
and applies it to the cluster. Only then can a developer's PVC be fulfilled.
This is slow, manual, and does not scale.

**Dynamic provisioning:**
A developer creates a PVC with a StorageClass name. Kubernetes automatically
calls the cloud API to create the storage, creates the PV object, and binds
them. No admin action is needed per request. The admin only sets up
StorageClasses once.

---

### Q5. What are the different access modes and when would you use each?

**ReadWriteOnce (RWO):** only one node can mount the volume for read and
write at the same time. Use for: databases (PostgreSQL, MySQL), single-instance
apps, StatefulSet pods. This is the most common mode for block storage (EBS, GCP PD).

**ReadWriteMany (RWX):** multiple nodes can mount the volume for read and
write simultaneously. Use for: web applications where multiple pods need to
share uploaded files, ML training jobs running on multiple nodes, shared
configuration files. Requires shared filesystem storage like NFS, CephFS, or AWS EFS.

**ReadOnlyMany (ROX):** multiple nodes can mount read-only simultaneously.
Use for: distributing read-only reference data (ML models, reference datasets)
to many pods.

**ReadWriteOncePod (RWOP):** only one single pod (not just one node) can
mount read-write. Stricter than RWO — even if two pods are on the same node,
only one can mount. Use for: safety-critical applications where you must
guarantee exclusive access at the pod level.

---

## Intermediate Level

---

### Q6. Explain the role of a CSI driver. Why was CSI introduced?

Before CSI, storage plugins were compiled directly into the Kubernetes
source code (called "in-tree plugins"). Every storage vendor had to get
their code merged into the Kubernetes codebase. This caused several problems:
storage bugs required a Kubernetes release to fix, adding new storage types
was slow, and storage vendors could not move at their own pace.

CSI (Container Storage Interface) is a standardized API specification that
defines how container orchestrators (like Kubernetes) communicate with storage
plugins. Storage vendors implement the CSI spec and ship their plugin independently
from Kubernetes.

A CSI driver has three main components:

The **ControllerServer** runs centrally (as a StatefulSet) and handles
cluster-level operations: CreateVolume (call cloud API to create storage),
DeleteVolume, ControllerPublishVolume (attach storage to a node),
ControllerExpandVolume (resize storage).

The **NodeServer** runs on every node (as a DaemonSet) and handles
node-local operations: NodeStageVolume (format and mount to a staging path),
NodePublishVolume (bind-mount into the pod's directory), NodeUnpublishVolume,
NodeUnstageVolume.

The **Identity** component provides driver identity and capability information.

Between Kubernetes and the CSI driver run **sidecar containers** that translate
Kubernetes events into CSI API calls: `external-provisioner` watches PVCs and
calls CreateVolume, `external-attacher` watches VolumeAttachments and calls
ControllerPublishVolume, `node-driver-registrar` registers the node plugin
with the kubelet.

---

### Q7. What is `volumeBindingMode: WaitForFirstConsumer` and why is it important?

`WaitForFirstConsumer` tells Kubernetes: do not create the actual storage
when the PVC is created. Wait until a pod that uses this PVC is scheduled
to a specific node, then create the storage in the same zone as that node.

**Why it is critical for zone-specific storage:**

Block storage like AWS EBS is zone-specific — an EBS volume in us-east-1a
can only be attached to EC2 instances also in us-east-1a. If you use
`Immediate` binding mode, Kubernetes creates the EBS volume immediately
when the PVC is created (possibly in us-east-1a), but later the scheduler
places the pod on a node in us-east-1b — and the pod can never start because
the volume is in the wrong zone. This is a very common production problem.

With `WaitForFirstConsumer`:
1. PVC is created → status `Pending` (no volume created yet)
2. Pod is created → scheduler finds a suitable node (say node in us-east-1b)
3. NOW Kubernetes creates the EBS volume, specifically in us-east-1b
4. Pod starts successfully with its storage in the same zone

**When to use Immediate:** for zone-independent storage like NFS, CephFS,
or AWS EFS (which is region-wide, not zone-specific).

---

### Q8. Walk through what happens step by step when a developer deletes a PVC that was dynamically provisioned.

What happens depends on the `reclaimPolicy` of the StorageClass.

**If reclaimPolicy is Delete (common default):**

1. Developer runs `kubectl delete pvc my-pvc`
2. Kubernetes first checks if any pod is currently using the PVC
   — if yes, deletion is blocked until the pod is deleted (PVC protection finalizer)
3. After pod is gone, the PVC object is deleted
4. The PV controller sees the PVC is gone and sets the PV status to `Released`
5. The `external-provisioner` sidecar sees the Released PV was dynamically provisioned
6. It calls CSI `DeleteVolume` on the plugin
7. The CSI plugin calls the cloud API: `aws ec2 delete-volume --volume-id vol-0abc123`
8. The physical EBS volume is permanently deleted
9. The PV object is deleted from Kubernetes
10. All data is gone permanently

**If reclaimPolicy is Retain:**

Steps 1-4 are the same, but:
5. The PV status becomes `Released` and stays there
6. No automatic deletion happens — neither of the PV object nor the EBS volume
7. The EBS volume continues to exist and accrue charges in AWS
8. The admin must manually inspect the data, then decide to:
   - Delete: `kubectl delete pv pvc-abc123` + `aws ec2 delete-volume ...`
   - Reuse: manually remove the `claimRef` from the PV, then a new PVC can bind to it

**Important PVC protection:**
Kubernetes adds a finalizer `kubernetes.io/pvc-protection` to every PVC.
This prevents PVC deletion while a pod is using it. Even if you run
`kubectl delete pvc` while a pod is running, the PVC will be stuck in
`Terminating` state until the pod is deleted.

---

### Q9. How does volume expansion (resizing) work in Kubernetes?

Volume expansion allows you to increase the size of a PVC after it has
been created and is in use.

**Prerequisites:**
- StorageClass must have `allowVolumeExpansion: true`
- The underlying storage type must support online expansion (EBS, GCP PD, CephFS do)

**How to expand:**
```bash
kubectl patch pvc my-database-pvc \
  -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'
```

**What happens internally:**

1. Kubernetes sees the PVC storage request increased from 10Gi to 20Gi
2. The PVC controller sets a condition: `FileSystemResizePending`
3. The external-resizer sidecar (part of CSI driver) detects this
4. It calls `ControllerExpandVolume` on the CSI ControllerServer
5. The CSI ControllerServer calls the cloud API:
   `aws ec2 modify-volume --volume-id vol-0abc123 --size 20`
6. The physical EBS volume is now 20Gi, but the filesystem is still formatted for 10Gi
7. The kubelet on the node calls `NodeExpandVolume`
8. The CSI NodeServer runs: `resize2fs /dev/xvdba` (for ext4) or `xfs_growfs` (for xfs)
9. The filesystem is expanded to use the full 20Gi
10. PVC condition `FileSystemResizePending` is removed

**Important rules:**
- You can ONLY increase, never decrease size (no shrinking)
- Most cloud providers require volume expansion to happen while the pod is running
  (the volume must be mounted for the filesystem resize to happen)
- Some older storage types require the pod to be restarted — check provider docs

---

### Q10. What is the difference between `volumeMode: Filesystem` and `volumeMode: Block`?

**Filesystem mode (default):**
Kubernetes formats the raw block device with a filesystem (ext4, xfs) and
mounts it as a directory inside the pod. The pod sees a regular directory
(e.g., `/data`) and uses normal file system operations (open, read, write, close).

This is what most applications use — databases like PostgreSQL, file servers,
log storage.

**Block mode:**
The raw block device is passed directly to the pod without any filesystem
on top. The pod sees `/dev/xvda` (or similar) — a raw block device.

The application is responsible for managing the raw blocks directly.

Use block mode when:
- The application has its own I/O management (some databases like Oracle or SAP HANA
  can use raw block for maximum performance)
- You need to pass a raw device to a virtual machine running inside a pod (KubeVirt)
- The application uses its own replication and does not want filesystem overhead

Most applications should use Filesystem mode. Block mode is only for
specialized, high-performance use cases.

---

## Advanced / Expert Level

---

### Q11. Explain the complete CSI flow from PVC creation to pod running, covering every CSI RPC call.

The CSI specification defines Remote Procedure Calls (RPCs) that Kubernetes
calls on the CSI driver at each stage. Here is the complete sequence:

**Phase 1 — Provision (external-provisioner sidecar)**

`CreateVolume` RPC is called when a new PVC needs storage:
```
Request:  name, capacity, parameters, topology requirements
Response: volume_id (the actual cloud resource ID), volume_context
```
The CSI plugin calls the cloud API to create physical storage and returns
the volume ID. The external-provisioner uses this to create the PV object
in Kubernetes.

**Phase 2 — Attach (external-attacher sidecar)**

After a pod is scheduled to a node, a `VolumeAttachment` object is created.
The external-attacher sees this and calls:

`ControllerPublishVolume` RPC:
```
Request:  volume_id, node_id (the node's identifier in the cloud), volume_capability
Response: publish_context (device path or other attachment info)
```
The CSI plugin attaches the cloud volume to the node (e.g., `aws ec2 attach-volume`).

**Phase 3 — Stage (kubelet on the node)**

The kubelet calls `NodeStageVolume` RPC on the CSI node plugin:
```
Request:  volume_id, staging_target_path, volume_capability, volume_context
Response: success/error
```
The CSI node plugin:
1. Finds the device (e.g., `/dev/nvme1n1`)
2. Formats it with a filesystem if empty (`mkfs.ext4`)
3. Mounts it to the staging (global) path

The staging path is shared — if multiple pods on the same node use the same
volume, formatting and global mount happen only once.

**Phase 4 — Publish (kubelet on the node)**

The kubelet calls `NodePublishVolume` RPC:
```
Request:  volume_id, staging_target_path, target_path (pod-specific), volume_capability
Response: success/error
```
The CSI node plugin creates a bind mount from the staging path to the
pod's specific volume directory. The pod can now access its storage.

**Teardown (reverse order):**

`NodeUnpublishVolume` → `NodeUnstageVolume` → `ControllerUnpublishVolume` → `DeleteVolume`

Each CSI RPC is idempotent by spec — calling it multiple times must be safe.
This handles retries from Kubernetes without side effects.

---

### Q12. A pod is stuck in `Pending` state with "waiting for a volume to be bound". How do you debug this?

This is one of the most common production storage issues. Systematic approach:

**Step 1 — Check PVC status**
```bash
kubectl get pvc -n <namespace>
# Look for: STATUS column showing "Pending" instead of "Bound"

kubectl describe pvc <pvc-name> -n <namespace>
# Look for Events section — usually tells you exactly what is wrong
```

**Step 2 — Check for StorageClass issues**
```bash
# Does the StorageClass exist?
kubectl get storageclass
# Check the name matches exactly what's in the PVC spec

# Is there a default StorageClass if PVC has no storageClassName?
kubectl get storageclass -o yaml | grep "is-default-class"
```

**Step 3 — Check the CSI driver is running**
```bash
kubectl get pods -n kube-system | grep csi
# All CSI pods should be Running

kubectl logs -n kube-system <csi-controller-pod> -c external-provisioner --tail=50
# Look for: errors trying to provision the PVC
```

**Step 4 — Common causes and fixes**

*StorageClass name mismatch:*
PVC specifies `storageClassName: gp3` but StorageClass is named `aws-ebs-gp3`.
Fix: correct the storageClassName in the PVC.

*No CSI driver for the StorageClass:*
The provisioner field in StorageClass points to a driver that is not installed.
Fix: install the correct CSI driver.

*Zone topology mismatch with WaitForFirstConsumer:*
PVC is Pending because no pod is scheduled yet (correct behaviour with
WaitForFirstConsumer). Fix: create a pod that uses the PVC; volume will
be provisioned once pod is scheduled.

*Cloud API permissions:*
CSI driver does not have IAM permissions to create EBS volumes.
Fix: check the CSI driver's ServiceAccount IRSA (IAM Role for Service Accounts)
permissions.
```bash
kubectl logs -n kube-system <csi-controller-pod> -c ebs-plugin | grep -i "permission\|auth\|denied"
```

*Storage quota exceeded:*
AWS/GCP account has hit volume limit or quota.
Check cloud provider console for quota errors.

*Node affinity requirements cannot be satisfied:*
StorageClass has allowedTopologies that no node matches.
```bash
kubectl get node --show-labels | grep topology.kubernetes.io/zone
```

**Step 5 — Check VolumeAttachment if PVC is Bound but pod still stuck**
```bash
kubectl get volumeattachment
kubectl describe volumeattachment <name>
# If attachment is failing, the volume is provisioned but cannot attach to the node
```

---

### Q13. What is the `WaitForFirstConsumer` binding mode and what problems can it cause in practice?

While `WaitForFirstConsumer` solves zone affinity problems, it introduces
its own complexities.

**Problem 1 — Volume Scheduling Deadlock with StatefulSets**

If you scale a StatefulSet and the scheduler cannot find a node that satisfies
both pod affinity rules AND the storage topology, pods remain Pending indefinitely.

Example:
- StatefulSet with podAntiAffinity (spread pods across zones)
- StorageClass with allowedTopologies (only us-east-1a and us-east-1b)
- 3rd replica: no node available in us-east-1a or us-east-1b that does not
  already have a pod → scheduling deadlock

**Problem 2 — Multi-PVC Pods**

If a pod requests two PVCs with WaitForFirstConsumer, Kubernetes must find
a node where BOTH volumes can be provisioned. If the two StorageClasses have
different topology constraints, no node may satisfy both — pod never schedules.

**Problem 3 — Debugging is harder**

With Immediate mode, you can see PV provisioning errors immediately after
PVC creation. With WaitForFirstConsumer, the PVC stays Pending until a pod
is scheduled — you might think the PVC is waiting for a pod, when actually
it will fail to provision once a pod is scheduled.

**Solution for complex scenarios:**
Use volume topology scheduling with careful StorageClass design:
```yaml
allowedTopologies:
- matchLabelExpressions:
  - key: topology.kubernetes.io/zone
    operator: In
    values:
    - us-east-1a
    - us-east-1b
    - us-east-1c    # all zones → fewer scheduling conflicts
```

---

### Q14. How does Kubernetes prevent a PVC from being deleted while a pod is using it?

Kubernetes uses a **finalizer** called `kubernetes.io/pvc-protection` for this.

When a PVC is created, Kubernetes automatically adds this finalizer to the
PVC metadata. A finalizer prevents an object from being deleted — Kubernetes
sets `deletionTimestamp` but does not remove the object until all finalizers
are removed.

The PVC Protection Controller (running in kube-controller-manager) watches
for PVCs that have been marked for deletion (have a `deletionTimestamp`).
If it finds a pod actively using the PVC, it leaves the finalizer in place.
The PVC stays in `Terminating` state until the pod is deleted.

```bash
# PVC stuck in Terminating:
kubectl get pvc my-pvc
# NAME    STATUS       VOLUME   CAPACITY   ACCESS MODES
# my-pvc  Terminating  pv-abc   10Gi       RWO

kubectl describe pvc my-pvc | grep Finalizers
# Finalizers: [kubernetes.io/pvc-protection]

# Find which pod is using it:
kubectl get pod -o json | jq '.items[] | select(.spec.volumes[]?.persistentVolumeClaim.claimName=="my-pvc") | .metadata.name'

# Delete the pod → PVC will then be deleted automatically
kubectl delete pod <pod-name>
```

Similarly, PVs have the `kubernetes.io/pv-protection` finalizer — a PV
cannot be deleted while it is bound to a PVC.

---

### Q15. How would you design StorageClasses for a production multi-tenant Kubernetes cluster?

For a production multi-tenant cluster, I would design StorageClasses around
four dimensions: performance tier, data criticality, access pattern, and cost.

**Tier 1 — High Performance (Databases, Stateful Apps)**
```yaml
name: premium-storage
provisioner: ebs.csi.aws.com
parameters:
  type: io2               # Highest IOPS SSD
  iops: "10000"
  encrypted: "true"
  kmsKeyId: <production-kms-key>
reclaimPolicy: Retain     # Never auto-delete — data is critical
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

**Tier 2 — Standard (General Applications)**
```yaml
name: standard-storage   # or set as default
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  encrypted: "true"
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

**Tier 3 — Shared Access (Web Apps, ML)**
```yaml
name: shared-storage
provisioner: efs.csi.aws.com   # supports ReadWriteMany
reclaimPolicy: Delete
volumeBindingMode: Immediate   # EFS is regional, not zone-specific
allowVolumeExpansion: true
```

**Tier 4 — Cheap Bulk (Logs, Archives)**
```yaml
name: bulk-storage
provisioner: ebs.csi.aws.com
parameters:
  type: st1    # Throughput-optimized HDD, cheapest
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

**Additional considerations for multi-tenancy:**

Use ResourceQuotas to limit storage per namespace:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: storage-quota
  namespace: team-a
spec:
  hard:
    requests.storage: "500Gi"
    premium-storage.storageclass.storage.k8s.io/requests.storage: "100Gi"
    persistentvolumeclaims: "20"
```

Use LimitRange for default storage requests:
```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: storage-limits
  namespace: team-a
spec:
  limits:
  - type: PersistentVolumeClaim
    max:
      storage: 100Gi
    min:
      storage: 1Gi
```

Use RBAC to restrict which StorageClasses namespaces can use — only allow
premium-storage in production namespaces, not dev/staging.

---

## Quick Reference

### PVC Status Values

```
Pending  → PVC created, waiting for PV to be bound (storage not yet provisioned)
Bound    → PVC successfully bound to a PV (storage is ready)
Lost     → PVC was bound to a PV that no longer exists (data may be lost)
Terminating → PVC is being deleted but finalizer is preventing it
```

### PV Status Values

```
Available  → PV exists but not bound to any PVC (ready to claim)
Bound      → PV is bound to a PVC (in use)
Released   → PVC was deleted but PV still exists (reclaimPolicy: Retain)
Failed     → Automatic reclamation failed
```

### Key Commands for Debugging

```bash
# Overview
kubectl get pvc,pv -A
kubectl get storageclass

# Debug PVC not binding
kubectl describe pvc <name> -n <namespace>    # check Events

# Debug volume not attaching
kubectl get volumeattachment
kubectl describe volumeattachment <name>

# Debug CSI driver
kubectl logs -n kube-system <csi-controller> -c external-provisioner
kubectl logs -n kube-system <csi-node-pod>   -c ebs-plugin

# Check node capacity for storage
kubectl describe node <node-name> | grep -A 10 "Allocated resources"

# Find which pod is using a PVC
kubectl get pod -A -o json | jq '.items[] | select(.spec.volumes[]?.persistentVolumeClaim.claimName=="<pvc-name>") | .metadata.name'
```

### Storage Decision Cheat Sheet

```
Need storage for ONE pod?              → ReadWriteOnce (RWO) + any block storage
Need storage shared between pods?      → ReadWriteMany (RWX) + NFS/EFS/CephFS
Database in production?                → reclaimPolicy: Retain
Dev/test storage?                      → reclaimPolicy: Delete
Zone-specific storage (EBS, GCP PD)?   → WaitForFirstConsumer
Zone-independent storage (EFS, NFS)?   → Immediate
Need resize later?                     → allowVolumeExpansion: true
Multiple pods on SAME node write?      → ReadWriteOnce is OK
Multiple pods on DIFFERENT nodes write?→ ReadWriteMany is REQUIRED
```