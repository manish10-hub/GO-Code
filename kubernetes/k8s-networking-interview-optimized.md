# Kubernetes Networking + StatefulSets
## Optimized Interview Revision Guide

> **Goal:** Read this document before an interview and quickly recall the complete picture.
>
> **Style:** Simple language, duplicate questions merged, similar questions combined, and answers reduced to the key points you should say in an interview.

---

# Part 1 — Kubernetes Networking

## 1. The Big Picture

Think about Kubernetes networking as **4 separate problems**:

```text
1. Pod Networking
   How does a Pod get an IP?
   → CNI + IPAM

2. Pod-to-Pod Networking
   How does one Pod reach another Pod?
   → CNI + routing / encapsulation / eBPF

3. Service Networking
   How does an application reach a stable backend?
   → Service + EndpointSlice + kube-proxy / eBPF

4. DNS
   How does an application discover the Service?
   → CoreDNS + Service DNS
```

Kubernetes networking covers communication between:

- Containers in the same Pod
- Pods on the same Node
- Pods on different Nodes
- Pods and Services
- External clients and Services
- Pods and external systems

---

# 2. Core Concepts You Must Know

## Q1. What is a Pod network namespace?

**Simple answer:**

A network namespace is an isolated networking environment for a Pod. It has its own IP address, network interfaces, routing table, sockets, and networking rules.

```text
Pod
┌─────────────────────────┐
│ Network Namespace       │
│                         │
│ eth0                    │
│ IP: 10.244.1.5          │
│ routing table           │
│ sockets                 │
│ lo                      │
└─────────────────────────┘
```

**Remember:**

> **Network namespace = isolated networking environment.**

---

## Q2. What is `eth0` inside a Pod?

`eth0` is the Pod's main network interface.

```text
Application
    |
    v
  socket
    |
    v
  eth0
    |
    v
  veth pair
    |
    v
  Node
```

The Pod's IP is normally assigned to this interface.

---

## Q3. What is a veth pair?

A **veth pair** is like a virtual network cable.

One end is inside the Pod's network namespace and the other end is connected to the Node's networking.

```text
Pod namespace                 Node namespace

   eth0  <================>  vethXXXX
             veth pair
```

Anything sent into one end comes out of the other end.

**Interview answer:**

> A veth pair connects the isolated Pod network namespace to the Node's network.

---

## Q4. Does every container get its own IP?

Normally, **no**.

The **Pod gets the IP**, and containers inside the same Pod share:

- Network namespace
- IP address
- Network interfaces
- Port space

Therefore:

```text
Container A → localhost:8080 → Container B
```

works when Container B listens on port 8080.

---

## Q5. Can two containers in the same Pod listen on the same port?

No.

They share the same network namespace and port space.

```text
Container A → :8080  ✅
Container B → :8080  ❌
```

The second container gets an "address already in use" type of error.

---

## Q6. What is CNI?

CNI stands for **Container Network Interface**.

It is the standard interface used to configure container networking.

Kubernetes relies on a CNI plugin for Pod networking.

Examples:

- Calico
- Cilium
- Flannel
- AWS VPC CNI
- Azure CNI
- OVN-Kubernetes

The CNI typically handles:

```text
Create/configure Pod network
        ↓
Create network interface
        ↓
Assign Pod IP
        ↓
Configure routes
        ↓
Connect Pod to cluster network
```

**Remember:**

> **CNI = Pod networking.**

---

# 3. How a Pod Gets an IP

## Q7. Explain how a Pod gets an IP.

High-level flow:

```text
kubectl
   ↓
API Server
   ↓
Scheduler
   ↓
Kubelet
   ↓
Container Runtime
   ↓
CNI
   ├── network namespace
   ├── network interface
   ├── IP allocation
   └── routes
   ↓
Pod
```

The exact implementation depends on the CNI.

---

# 4. Pod-to-Pod Communication

## Q8. How do two Pods on the same Node communicate?

Conceptually:

```text
Pod A
10.244.1.2
   |
  eth0
   |
 veth-a
   |
   v
 Bridge / Node networking
   |
 veth-b
   |
  eth0
   |
   v
Pod B
10.244.1.3
```

The communication stays inside the Node.

In a bridge-based setup, the bridge behaves like a virtual Layer-2 switch and forwards frames using MAC addresses.

**Interview answer:**

> Each Pod has its own network namespace and veth pair. On the same Node, traffic can go through the veth pairs and the Node's virtual bridge without leaving the Node.

---

## Q9. What is the role of a bridge such as `cni0`?

A bridge acts like a **virtual Layer-2 switch**.

It connects the Node-side ends of Pod veth pairs.

```text
Pod A ── veth-a ──┐
                  │
Pod B ── veth-b ──┤── cni0 bridge
                  │
Pod C ── veth-c ──┘
```

The bridge uses MAC addresses to decide where to forward Ethernet frames.

---

## Q10. How does ARP help with same-node Pod communication?

If Pod A wants to reach Pod B and both are on the same subnet:

```text
Pod A:
"Who has 10.244.1.3?"
       ↓
ARP
       ↓
MAC address of Pod B's interface
```

Pod A then sends the Ethernet frame to Pod B's MAC address.

**Remember:**

> ARP resolves **IPv4 address → MAC address** on the local Layer-2 network.

---

# 5. Most Important Question: Pod on Node 1 → Pod on Node 2

## Q11. How does a Pod on Node 1 communicate with a Pod on Node 2?

Example:

```text
Node 1                         Node 2

Pod A                          Pod B
10.244.1.2                     10.244.2.3
   |                               |
   v                               v
 veth                            veth
   |                               |
   v                               v
CNI / Node 1                  CNI / Node 2
   |                               |
   +-------- Physical Network -----+
```

The packet path is:

```text
Application
    ↓
Pod network namespace
    ↓
Pod eth0
    ↓
veth
    ↓
Node 1
    ↓
CNI routing / encapsulation
    ↓
Physical or overlay network
    ↓
Node 2
    ↓
CNI / Node networking
    ↓
veth
    ↓
Pod B
```

The exact implementation depends on the CNI.

Possible implementations include:

- Native Linux routing
- VXLAN
- Geneve
- IP-in-IP
- eBPF
- Cloud-native networking

### Two common approaches

### A. Overlay networking

The CNI encapsulates the Pod packet inside another packet.

```text
Outer packet:
Node 1 IP → Node 2 IP

Inner packet:
Pod A IP → Pod B IP
```

Example with VXLAN:

```text
Pod A
10.244.1.2
   ↓
Node 1
   ↓
VXLAN encapsulation
   ↓
Node 2
   ↓
VXLAN decapsulation
   ↓
Pod B
10.244.2.3
```

### B. Underlay routing / BGP

The network learns routes to Pod CIDRs.

```text
10.244.1.0/24 → Node 1
10.244.2.0/24 → Node 2
```

No tunnel is required.

### Best interview answer

> A Pod sends the packet through its `eth0` and veth pair to Node 1. The CNI provides the route to the destination Pod network. Depending on the CNI, the packet may be routed directly or encapsulated using something like VXLAN/Geneve. The packet reaches Node 2, where the CNI delivers it through the destination Pod's veth interface. Normal Pod-to-Pod traffic is designed to work without NAT.

---

# 6. Pod CIDR

## Q12. What is Pod CIDR?

Pod CIDR is the IP range used for Pod addresses.

Example:

```text
Cluster Pod CIDR:
10.244.0.0/16

Node 1:
10.244.1.0/24

Node 2:
10.244.2.0/24

Node 3:
10.244.3.0/24
```

Each Node receives a non-overlapping range so Pod IPs remain unique.

---

## Q13. What is IPAM?

IPAM = **IP Address Management**.

It is responsible for allocating Pod IP addresses from the available Pod address ranges.

**Remember:**

> CNI provides networking; IPAM handles IP allocation.

---

# 7. Routing vs MAC vs ARP

## Q14. What is the difference between IP, MAC and ARP?

| Concept | Meaning |
|---|---|
| IP | Layer-3 logical address |
| MAC | Layer-2 hardware/interface address |
| ARP | Resolves IPv4 IP → MAC |
| Routing | Decides where an IP packet should go |
| TCP/UDP | Layer-4 transport |

### Important correction

```text
Layer 1 → Physical
Layer 2 → Ethernet / MAC
Layer 3 → IP / Routing
Layer 4 → TCP / UDP
```

A traditional switch primarily works at Layer 2.

A router primarily works at Layer 3.

---

# 8. Kubernetes Services

## Q15. Why do we need a Service?

Pods are ephemeral.

A Pod can:

- Restart
- Be deleted
- Be rescheduled
- Receive a different IP

A Service provides a stable endpoint.

```text
Service
ClusterIP
   |
   +---- Pod A
   +---- Pod B
   +---- Pod C
```

**Interview answer:**

> A Service gives applications a stable IP and DNS name while the underlying Pod IPs can change.

---

## Q16. What is ClusterIP?

ClusterIP is the virtual IP assigned to a Service.

It is normally used for communication **inside the cluster**.

```text
Client Pod
    ↓
Service ClusterIP
    ↓
Backend Pod
```

---

## Q17. What is kube-proxy?

`kube-proxy` implements Service networking on Nodes in many Kubernetes configurations.

It watches:

- Services
- EndpointSlices

and configures networking rules.

Possible mechanisms include:

- iptables
- IPVS
- nftables in newer configurations

Some networking solutions can replace kube-proxy functionality, for example using eBPF.

**Remember:**

> **CNI → Pod networking**
>
> **kube-proxy → Service networking**

---

## Q18. What is EndpointSlice?

EndpointSlices contain the network endpoints backing a Service.

Example:

```text
Service: backend

EndpointSlice:
10.1.1.10:8080
10.1.1.11:8080
10.1.2.10:8080
```

The Service selects Pods using labels.

```yaml
selector:
  app: backend
```

Matching Pods become Service endpoints.

---

# 9. How Service Traffic Works

## Q19. What happens when a Pod calls a Service?

Example:

```text
Pod A
  |
  | backend.default.svc.cluster.local
  ↓
CoreDNS
  |
  | returns ClusterIP
  ↓
Service ClusterIP
  |
  ↓
kube-proxy / eBPF
  |
  ↓
selected endpoint
  |
  ↓
Pod B
```

In a traditional iptables implementation:

```text
Service IP: 10.96.5.10:80
       ↓
DNAT
       ↓
Pod IP: 10.244.2.3:8080
```

**DNAT = Destination NAT.**

The destination changes from:

```text
Service IP → Pod IP
```

Conntrack tracks the connection so return traffic can be translated back correctly.

---

## Q20. What is the difference between CNI and kube-proxy?

**CNI:**

> Provides Pod networking: IPs, interfaces, routes, cross-node connectivity.

**kube-proxy:**

> Implements Service networking: Service IP → backend endpoint forwarding.

Modern eBPF-based CNIs can provide both.

---

# 10. Service Types

## Q21. ClusterIP vs NodePort vs LoadBalancer

```text
ClusterIP
  → Internal cluster access

NodePort
  → Node IP + fixed port

LoadBalancer
  → External/cloud load balancer
```

### ClusterIP

```text
Pod → Service → Pod
```

### NodePort

```text
External Client
    ↓
NodeIP:30080
    ↓
Service
    ↓
Pod
```

### LoadBalancer

```text
Internet
   ↓
Cloud Load Balancer
   ↓
Service
   ↓
Pod
```

---

# 11. Ingress

## Q22. What is the difference between Ingress and Ingress Controller?

**Ingress resource:**

Defines routing rules.

```text
/api → backend-service
/web → frontend-service
```

**Ingress Controller:**

Actually implements those rules.

Examples include:

- NGINX
- HAProxy
- Traefik
- Kong
- Cloud-provider controllers

**Remember:**

> Ingress = configuration.
>
> Ingress Controller = actual software handling the traffic.

Ingress is primarily Layer 7 HTTP/HTTPS routing.

---

# 12. DNS

## Q23. What is CoreDNS?

CoreDNS provides DNS resolution inside the Kubernetes cluster.

Typical flow:

```text
Application
   ↓
/etc/resolv.conf
   ↓
CoreDNS Service
   ↓
CoreDNS Pod
   ↓
Kubernetes DNS data
```

---

## Q24. What does `/etc/resolv.conf` contain?

Typical example:

```text
nameserver 10.96.0.10
search default.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
```

Important fields:

- `nameserver` → DNS server address
- `search` → search domains
- `ndots` → resolver search behaviour

---

## Q25. What is the Kubernetes Service DNS format?

```text
<service>.<namespace>.svc.cluster.local
```

Example:

```text
backend.production.svc.cluster.local
```

From the same namespace:

```text
backend
```

From another namespace:

```text
backend.production
```

---

## Q26. What happens when `curl http://backend` is executed?

```text
Application
   ↓
DNS resolver
   ↓
/etc/resolv.conf
   ↓
CoreDNS
   ↓
Service ClusterIP
   ↓
Service routing
   ↓
Backend Pod
```

---

# 13. Headless Service

## Q27. What is a Headless Service?

A Headless Service has:

```yaml
clusterIP: None
```

Unlike a normal Service, it does not provide a virtual ClusterIP.

DNS can return individual Pod IPs.

```text
database.default.svc.cluster.local
        ↓
  Pod A
  Pod B
  Pod C
```

It is especially useful for:

- Stateful applications
- Databases
- Kafka
- Distributed systems
- StatefulSets

---

# 14. NetworkPolicy

## Q28. What is NetworkPolicy?

NetworkPolicy defines which network traffic is allowed.

It can control:

- Ingress
- Egress
- Pod-to-Pod communication
- Namespace-based communication
- IP/CIDR-based traffic

Example mental model:

```text
Frontend ─────→ Backend    ALLOWED

Frontend ─────X→ Database  BLOCKED
```

Kubernetes defines the API, but the CNI/networking implementation enforces the policy.

---

## Q29. Does NetworkPolicy apply to Services or Pods?

Think in terms of **Pods/endpoints**.

If traffic goes:

```text
Client
  ↓
Service
  ↓
Destination Pod
```

the policy ultimately controls traffic involving the selected Pod.

A common mistake is to think:

> "I allowed the Service IP, so the Pod is allowed."

The policy should be designed around the destination Pod and its port.

---

# 15. NAT

## Q30. What is SNAT vs DNAT?

### SNAT

**Source NAT**

```text
Pod IP → Node/NAT Gateway IP
```

Common for outbound traffic.

### DNAT

**Destination NAT**

```text
Service ClusterIP → Pod IP
```

Common in Service forwarding.

---

## Q31. Why is SNAT commonly needed when a Pod accesses the Internet?

Pod IPs are normally internal cluster addresses.

The external system needs a routable source address.

```text
Pod
10.244.1.5
   ↓
SNAT
   ↓
Node/NAT Gateway IP
   ↓
Internet
```

The response is reverse-translated using connection tracking.

---

# 16. eBPF

## Q32. What problem does eBPF solve in Kubernetes networking?

Traditional networking can involve:

```text
veth
→ bridge
→ routing
→ iptables
→ conntrack
→ other processing
```

eBPF allows programmable packet processing inside the Linux kernel.

Cilium can use eBPF for:

- Packet forwarding
- Service load balancing
- NetworkPolicy
- Observability
- Cross-node networking

**Interview answer:**

> eBPF lets a CNI move packet-processing logic into efficient kernel hooks, reducing dependence on large iptables rule sets and enabling fast networking, load balancing and policy enforcement.

---

# 17. VXLAN vs BGP

## Q33. VXLAN overlay vs BGP underlay

| VXLAN | BGP |
|---|---|
| Encapsulates packets | Routes Pod CIDRs directly |
| Works without special physical network support | Requires BGP-capable network/configuration |
| Adds encapsulation overhead | No tunnel overhead |
| Useful when physical network is not under your control | Useful when you control the network |
| Simpler network integration | Better native routing/performance |

Mental model:

```text
VXLAN:
Pod → Node → Tunnel → Node → Pod

BGP:
Pod → Node → Routed network → Node → Pod
```

---

# 18. Pod-to-Internet

## Q34. Explain Pod → Internet traffic.

```text
Pod
 ↓
CNI
 ↓
Node
 ↓
SNAT / NAT Gateway
 ↓
Cloud Network
 ↓
Internet
```

Typical flow:

```text
Pod:
10.244.1.2 → 1.2.3.4

SNAT:

192.168.1.10 → 1.2.3.4
```

The external system replies to the routable address and conntrack/NAT maps the response back to the Pod.

---

# 19. Troubleshooting Kubernetes Networking

## Q35. What is your general network-debugging approach?

Always debug **from source to destination**.

```text
1. Identify the scenario
   ↓
2. Check source
   ↓
3. Check destination
   ↓
4. Check DNS
   ↓
5. Check IP connectivity
   ↓
6. Check port
   ↓
7. Check Service
   ↓
8. Check NetworkPolicy
   ↓
9. Check CNI/routing
   ↓
10. Capture packets if necessary
```

---

## Q36. Pod A cannot reach Pod B on the same Node. What do you check?

### First

```bash
kubectl get pods -o wide
```

Confirm both Pods:

- Are Running
- Have IPs
- Are on the same Node

### Then

```bash
kubectl exec pod-a -- nc -zv <pod-b-ip> <port>
```

Check:

- Pod B is listening
- Correct port
- NetworkPolicy
- veth/bridge
- Node routes

Useful Node commands:

```bash
ip route
ip addr
bridge link show
tcpdump
```

---

## Q37. Pod A cannot reach Pod B on another Node. What do you check?

Use this order:

```text
1. Are both Pods running?
2. Can Node 1 reach Node 2?
3. Does Node 1 have a route to Node 2's Pod CIDR?
4. Is the CNI agent healthy?
5. Is VXLAN/BGP working?
6. Is firewall/security group blocking traffic?
7. Is MTU correct?
8. Is NetworkPolicy blocking traffic?
9. Use tcpdump to locate the drop.
```

Useful commands:

```bash
kubectl get pods -o wide
kubectl get nodes -o wide
ip route show
kubectl get pods -n kube-system
tcpdump
```

---

## Q38. How do you troubleshoot Kubernetes DNS failure?

Check:

```bash
kubectl get pods -n kube-system
kubectl get svc -n kube-system
kubectl logs -n kube-system <coredns-pod>
```

Inside the Pod:

```bash
cat /etc/resolv.conf
nslookup kubernetes.default
nslookup <service-name>
```

Also check NetworkPolicy.

If egress policy is enabled, DNS traffic to CoreDNS must be allowed.

---

## Q39. A Service is not reachable. What do you check?

Use this order:

```text
DNS
 ↓
Service exists
 ↓
Selector
 ↓
EndpointSlices
 ↓
Pod readiness
 ↓
targetPort
 ↓
Service ClusterIP connectivity
 ↓
kube-proxy / eBPF
 ↓
NetworkPolicy
 ↓
Pod connectivity
```

Commands:

```bash
kubectl get svc <service>
kubectl describe svc <service>
kubectl get endpointslices
kubectl get pods --show-labels
kubectl get networkpolicy
```

---

## Q40. Service has no endpoints. What are the likely causes?

Most common:

1. Service selector does not match Pod labels.
2. Matching Pods are not Ready.
3. Pods are not running.

Check:

```bash
kubectl describe svc <service>
kubectl get pods --show-labels
kubectl get endpointslices
kubectl get pods
```

---

## Q41. Service works by ClusterIP but DNS name does not work. What does that tell you?

It strongly suggests:

```text
Service routing = probably OK
DNS = likely problem
```

Focus on:

```text
CoreDNS
/etc/resolv.conf
NetworkPolicy
DNS configuration
```

---

## Q42. Direct Pod IP works but Service IP does not. What do you suspect?

Focus on Service networking:

- kube-proxy
- iptables
- IPVS
- nftables
- eBPF Service implementation
- Service selector/endpoints

---

## Q43. Small packets work but large packets fail across Nodes. What do you suspect?

Think **MTU**.

VXLAN adds packet overhead.

Example:

```text
Node MTU = 1500
VXLAN overhead ≈ 50 bytes
Pod MTU ≈ 1450
```

A useful test:

```bash
ping -M do -s 1400 <pod-ip>
```

---

# 20. Essential Commands

```bash
# Pods / IP / Node
kubectl get pods -o wide

# Nodes
kubectl get nodes -o wide

# Services
kubectl get svc

# EndpointSlices
kubectl get endpointslices

# NetworkPolicy
kubectl get networkpolicy

# DNS configuration
kubectl exec <pod> -- cat /etc/resolv.conf

# DNS test
kubectl exec <pod> -- nslookup kubernetes.default

# TCP test
kubectl exec <pod> -- nc -zv <ip> <port>

# HTTP test
kubectl exec <pod> -- curl -v http://<ip>:<port>

# Routes
ip route

# Interfaces
ip addr

# Bridge
bridge link show

# iptables
iptables -L -n

# nftables
nft list ruleset

# Packet capture
tcpdump
```

---

# 21. Kubernetes Networking — Senior Interview Questions

Be ready to answer these without memorizing long definitions:

1. Walk through Pod A → Pod B on different Nodes.
2. What does CNI do?
3. What is a network namespace?
4. What is a veth pair?
5. How do same-node Pods communicate?
6. How does ARP work in Pod networking?
7. What is Pod CIDR?
8. What is IPAM?
9. VXLAN vs BGP?
10. What does kube-proxy do?
11. How does Service → Pod routing work?
12. What is DNAT?
13. What is conntrack?
14. Endpoint vs EndpointSlice?
15. What is CoreDNS?
16. What is `ndots`?
17. ClusterIP vs NodePort vs LoadBalancer?
18. Ingress vs Ingress Controller?
19. What is NetworkPolicy?
20. SNAT vs DNAT?
21. Why is SNAT used for Internet traffic?
22. What does eBPF improve?
23. Why can DNS be slow in Kubernetes?
24. How do you debug cross-node packet drops?
25. How do you debug a Service with no endpoints?

---

# Part 22 — Kubernetes StatefulSets

# 22. What Problem Does StatefulSet Solve?

## Q26. StatefulSet vs Deployment — why do we need StatefulSet?

A Deployment is designed for **stateless, interchangeable Pods**.

A StatefulSet is for applications that need:

- Stable Pod identity
- Stable network identity
- Dedicated persistent storage
- Ordered startup/shutdown/update behaviour

Example:

```text
Deployment:

app-7f9d4-xk2p1
app-7f9d4-ab3q2

StatefulSet:

postgres-0
postgres-1
postgres-2
```

Typical stateful workloads:

- PostgreSQL
- MySQL
- MongoDB
- Kafka
- etcd
- ZooKeeper
- Elasticsearch

---

# 23. Four Things to Remember About StatefulSet

## Q27. What are the key guarantees of a StatefulSet?

Remember:

```text
I N S O
```

### I — Identity

```text
postgres-0
postgres-1
postgres-2
```

### N — Network identity

Stable per-Pod DNS when used with a Headless Service.

```text
postgres-0.postgres-svc
postgres-1.postgres-svc
```

### S — Storage

One PVC per Pod using `volumeClaimTemplates`.

```text
postgres-0 → data-postgres-0
postgres-1 → data-postgres-1
postgres-2 → data-postgres-2
```

### O — Ordering

Pods can be created/deleted/updated in a controlled order.

**Interview answer:**

> StatefulSet gives Pods stable identity, stable network identity, dedicated persistent storage, and ordered lifecycle operations.

---

# 24. StatefulSet Pod Naming

## Q28. Why is `postgres-0` important?

The ordinal is stable.

```text
postgres-0
postgres-1
postgres-2
```

If `postgres-1` is recreated:

```text
old postgres-1
      ↓
new postgres-1
```

The identity remains `postgres-1`.

This is useful when the application needs to know which instance it is talking to.

---

# 25. StatefulSet + Headless Service

## Q29. Why is a Headless Service used with StatefulSet?

A normal Service provides:

```text
One virtual IP
       ↓
multiple Pods
```

A Headless Service uses:

```yaml
clusterIP: None
```

and DNS can return individual Pod IPs.

Example:

```text
postgres-0.postgres-svc.default.svc.cluster.local
        ↓
current IP of postgres-0
```

If the Pod gets a new IP after restart, the DNS name remains stable while DNS points to the new IP.

**Remember:**

> StatefulSet gives the stable **name**; Headless Service + DNS gives stable **network discovery**.

---

# 26. StatefulSet Storage

## Q30. What is `volumeClaimTemplates`?

It creates a separate PVC for each StatefulSet Pod.

```text
postgres-0 → data-postgres-0
postgres-1 → data-postgres-1
postgres-2 → data-postgres-2
```

This is different from referencing one existing PVC from a normal `volumes` section.

### Why?

A database usually needs:

```text
Pod 0 → its own data
Pod 1 → its own data
Pod 2 → its own data
```

not:

```text
Pod 0 ─┐
Pod 1 ─┼→ same database storage
Pod 2 ─┘
```

---

## Q31. What happens to the PVC when a StatefulSet Pod is deleted?

The PVC normally remains.

When the Pod is recreated:

```text
postgres-1
    ↓
same PVC
data-postgres-1
```

The data therefore survives Pod recreation.

PVCs also remain when the StatefulSet itself is deleted by default; Kubernetes provides `persistentVolumeClaimRetentionPolicy` to control retention behaviour.

---

# 27. StatefulSet Pod Management

## Q32. `OrderedReady` vs `Parallel`

### OrderedReady

Default behaviour:

```text
postgres-0
   ↓ Ready
postgres-1
   ↓ Ready
postgres-2
```

Use when startup/shutdown order matters.

### Parallel

Pods can start/delete together.

Identity and storage are still preserved.

Use when Pods can start independently.

**Easy memory:**

```text
OrderedReady = order matters
Parallel     = order doesn't matter
```

---

# 28. StatefulSet Update Strategy

## Q33. RollingUpdate vs OnDelete

### RollingUpdate

Kubernetes automatically updates Pods.

For a StatefulSet:

```text
2 → 1 → 0
```

Highest ordinal first.

### OnDelete

Kubernetes does not automatically replace existing Pods just because the StatefulSet spec changed.

You manually delete Pods one by one.

```yaml
updateStrategy:
  type: OnDelete
```

**When useful:**

- Manual database upgrades
- Need to verify each node
- Need custom operational steps

---

# 29. StatefulSet Partition

## Q34. What is `partition`?

`partition: N` means Pods with ordinal:

```text
>= N
```

are updated.

Pods with ordinal:

```text
< N
```

remain on the old version.

Example:

```text
3 Pods:
postgres-0
postgres-1
postgres-2

partition: 2
```

Result:

```text
postgres-0 → OLD
postgres-1 → OLD
postgres-2 → NEW
```

Then:

```text
partition: 1
```

gives:

```text
postgres-0 → OLD
postgres-1 → NEW
postgres-2 → NEW
```

Finally:

```text
partition: 0
```

updates all Pods.

**Memory:**

> **Partition = StatefulSet canary update control.**

---

# 30. Scaling StatefulSet

## Q35. What happens when you scale a StatefulSet up?

Example:

```bash
kubectl scale statefulset postgres --replicas=5
```

New Pods are created with the next ordinals:

```text
postgres-3
postgres-4
```

and their PVCs are created automatically:

```text
data-postgres-3
data-postgres-4
```

---

## Q36. What happens when you scale down?

Example:

```text
5 → 2
```

Pods are removed from the highest ordinal first:

```text
postgres-4
postgres-3
postgres-2
```

The lower ordinals remain.

PVCs normally remain too.

This means scaling back up can reuse the existing PVCs.

---

# 31. StatefulSet Failure

## Q37. What happens when a StatefulSet Pod fails?

The StatefulSet controller recreates the Pod with the same identity.

Example:

```text
postgres-1
   ↓ crash
recreated postgres-1
   ↓
data-postgres-1
```

The identity and storage are preserved.

---

# 32. Stuck Pod / Dead Node

## Q38. What if a StatefulSet Pod is stuck in `Terminating` because its Node is unreachable?

This needs care.

Do **not** immediately force-delete the Pod.

First verify that the Node is genuinely dead.

Why?

If the old Pod is still running on an isolated Node and you force-delete it, Kubernetes may create a replacement while the old process still exists.

That can create a dangerous duplicate identity situation for stateful applications.

Safe high-level approach:

```text
1. Verify Node is truly dead
        ↓
2. Force-delete only if confirmed
        ↓
3. StatefulSet creates replacement
        ↓
4. Replacement reattaches to same PVC
```

Example:

```bash
kubectl delete pod postgres-1 --force --grace-period=0
```

Use force deletion carefully.

---

# 33. StatefulSet and Split-Brain

## Q39. Does StatefulSet completely prevent database split-brain?

No.

StatefulSet provides identity, storage and conservative Pod replacement behaviour, but it does **not** understand database leadership.

Kubernetes knows things such as:

```text
Pod = Ready
Pod = NotReady
```

It does not automatically understand:

```text
"This database is the legitimate primary."
```

Production databases may need application-level mechanisms such as:

- Consensus
- Leader election
- Fencing
- STONITH
- Database operators

For example, PostgreSQL deployments commonly use an operator/HA solution such as Patroni rather than relying on StatefulSet alone.

**Strong interview answer:**

> StatefulSet provides the infrastructure primitives for a stateful workload, but it is not a database consensus or leader-election system.

---

# 34. StatefulSet + PostgreSQL

## Q40. How can StatefulSet identity be used for PostgreSQL primary/replica setup?

Example:

```text
postgres-0 → primary
postgres-1 → replica
postgres-2 → replica
```

The application can use the stable DNS name:

```text
postgres-0.postgres-svc
```

for the primary.

A common setup uses:

- Pod ordinal
- Init container
- Stable DNS
- Dedicated PVC
- Readiness probe

Conceptually:

```text
ordinal 0
   ↓
primary

ordinal > 0
   ↓
replica
   ↓
connect to postgres-0
```

---

# 35. StatefulSet + Kafka

## Q41. Why is StatefulSet useful for Kafka?

The mapping is natural:

```text
StatefulSet Pod ordinal
        ↓
Kafka broker identity

kafka-0 → broker 0
kafka-1 → broker 1
kafka-2 → broker 2
```

Dedicated storage:

```text
data-kafka-0
data-kafka-1
data-kafka-2
```

Stable DNS:

```text
kafka-0.kafka-svc
kafka-1.kafka-svc
kafka-2.kafka-svc
```

Operationally, Kafka still needs Kafka-aware logic for:

- Partition reassignment
- Broker decommissioning
- Safe upgrades
- External client access

An operator such as Strimzi can handle these application-specific operations.

---

# 36. PDB

## Q42. What is a PodDisruptionBudget?

A PDB controls how many Pods can be unavailable during **voluntary disruptions**.

For a 3-Pod database:

```yaml
minAvailable: 2
```

means at least 2 Pods should remain available.

Or:

```yaml
maxUnavailable: 1
```

means at most 1 Pod can be unavailable.

**Memory:**

```text
3 Pods
minAvailable = 2
→ keep majority available
```

A PDB is useful for planned operations such as node drains, but it is not a replacement for application-level HA.

---

# 37. StatefulSet Retention Policy

## Q43. What is `persistentVolumeClaimRetentionPolicy`?

It controls what happens to PVCs when:

- StatefulSet is deleted
- StatefulSet is scaled down

Example:

```yaml
persistentVolumeClaimRetentionPolicy:
  whenDeleted: Retain
  whenScaled: Retain
```

### Retain

Keep PVC.

### Delete

Delete PVC according to the configured lifecycle event.

For production database storage, retaining PVCs is generally the safer policy.

---

# 38. StatefulSet vs Deployment — Final Table

| Feature | Deployment | StatefulSet |
|---|---|---|
| Pod identity | Interchangeable | Stable identity |
| Pod name | Random suffix | `app-0`, `app-1` |
| Network identity | Usually changes | Stable per-Pod DNS with Headless Service |
| Storage | No dedicated PVC by default | Dedicated PVC via `volumeClaimTemplates` |
| Startup | Usually independent | Ordered by default |
| Scale-down | General | Highest ordinal first |
| Update | Rolling | Highest ordinal first |
| Best for | Stateless apps | Stateful apps |

---

# 39. StatefulSet Interview Questions — Senior Level

Be ready for these:

1. What problem does StatefulSet solve?
2. StatefulSet vs Deployment?
3. What are StatefulSet's key guarantees?
4. Why is stable identity important?
5. How does a Headless Service work?
6. How does StatefulSet Pod DNS work?
7. What is `volumeClaimTemplates`?
8. What happens to PVCs after Pod deletion?
9. What happens to PVCs after scale-down?
10. `OrderedReady` vs `Parallel`?
11. `RollingUpdate` vs `OnDelete`?
12. What is `partition`?
13. How do you perform a canary StatefulSet update?
14. What happens if a StatefulSet Pod fails?
15. What happens if the Node is dead?
16. Why is force deletion dangerous?
17. Does StatefulSet prevent split-brain?
18. How would you run PostgreSQL using StatefulSet?
19. How would you run Kafka using StatefulSet?
20. How does PDB help a StatefulSet?
21. What is PVC retention policy?
22. How would you safely upgrade a stateful database?

---

# 40. Most Important Whiteboard Diagrams

## Diagram 1 — Same Pod

```text
Pod
┌─────────────────────────────┐
│ Shared Network Namespace    │
│                             │
│ Container A   Container B   │
│     |             |         │
│     +--- localhost ---------+
└─────────────────────────────┘
```

## Diagram 2 — Pod Networking

```text
Pod
 ↓
Network Namespace
 ↓
eth0
 ↓
veth
 ↓
Node
 ↓
CNI
```

## Diagram 3 — Cross-Node Pod Networking

```text
Pod A
 ↓
veth
 ↓
Node 1
 ↓
CNI
 ↓
Routing / VXLAN / Geneve / BGP
 ↓
Node 2
 ↓
CNI
 ↓
veth
 ↓
Pod B
```

## Diagram 4 — Service

```text
Application
 ↓
DNS
 ↓
CoreDNS
 ↓
Service ClusterIP
 ↓
kube-proxy / eBPF
 ↓
EndpointSlice
 ↓
Pod
```

## Diagram 5 — StatefulSet

```text
StatefulSet
     |
     +-------------------------+
     |            |            |
 postgres-0   postgres-1   postgres-2
     |            |            |
 data-0       data-1       data-2
     |
 Headless Service
     |
 Stable Pod DNS
```

---

# 41. Final 5-Minute Revision

Before the interview, remember these statements:

### Networking

> **Every Pod gets an IP.**

> **Containers in the same Pod share one network namespace and communicate through localhost.**

> **veth connects the Pod network namespace to the Node.**

> **CNI provides Pod networking.**

> **Same-node Pod traffic can stay inside the Node.**

> **Cross-node Pod traffic uses CNI routing, overlay, underlay or eBPF depending on the implementation.**

> **Service provides a stable endpoint for changing Pods.**

> **EndpointSlice contains Service backend endpoints.**

> **kube-proxy traditionally implements Service forwarding; some CNIs can replace it with eBPF.**

> **CoreDNS provides Kubernetes DNS.**

> **Headless Service provides per-Pod discovery through DNS.**

> **NetworkPolicy controls allowed network traffic; the CNI enforces it.**

> **SNAT changes the source address; DNAT changes the destination address.**

### StatefulSet

> **StatefulSet = stable identity + stable network identity + dedicated storage + ordered lifecycle.**

> **`postgres-0` remains `postgres-0` after recreation.**

> **`volumeClaimTemplates` creates one PVC per Pod.**

> **Headless Service + DNS provides stable per-Pod discovery.**

> **`OrderedReady` = order matters.**

> **`Parallel` = Pods can start independently.**

> **`RollingUpdate` updates automatically; `OnDelete` requires manual Pod deletion.**

> **`partition` gives controlled/canary updates.**

> **PVCs normally survive Pod deletion and scale-down.**

> **StatefulSet does not provide database leader election or complete split-brain protection.**

---

# 42. The Two Questions You Must Be Able to Answer

## Question 1

> **"A Pod on Node 1 sends a request to a Pod on Node 2. Walk me through the packet."**

Your answer:

```text
Application
 → Pod socket
 → Pod network namespace
 → eth0
 → veth
 → Node 1
 → CNI routing / encapsulation
 → physical/overlay network
 → Node 2
 → destination Pod networking
 → veth
 → eth0
 → Pod B
```

Then say:

> "The exact implementation depends on the CNI — it may use native routing, VXLAN, Geneve, IP-in-IP or eBPF."

---

## Question 2

> **"Why would you use StatefulSet instead of Deployment?"**

Your answer:

> "I use StatefulSet when Pods are not interchangeable and need stable identity, stable network discovery and dedicated persistent storage. StatefulSet gives predictable names like `app-0`, `app-1`, dedicated PVCs through `volumeClaimTemplates`, and controlled startup, scaling and update behaviour."

---

# Final Mental Model

```text
                    KUBERNETES

                       DNS
                        |
                     CoreDNS
                        |
                        v
Application ───────→ Service
                        |
                        v
                 EndpointSlice
                        |
                        v
                kube-proxy/eBPF
                        |
                        v
                       CNI
                        |
              +---------+---------+
              |                   |
           Node 1              Node 2
              |                   |
            Pod A               Pod B


For Stateful workloads:

              StatefulSet
                   |
       +-----------+-----------+
       |           |           |
     pod-0       pod-1       pod-2
       |           |           |
     PVC-0       PVC-1       PVC-2
       |
Stable identity + DNS + storage
```

## What to prioritize for a Senior/Tech Lead interview

```text
1. Pod network namespace
2. veth pair
3. Same-node Pod communication
4. Cross-node Pod communication
5. CNI
6. Routing / VXLAN / BGP
7. Service + EndpointSlice
8. kube-proxy + DNAT + conntrack
9. CoreDNS
10. NetworkPolicy
11. SNAT / DNAT
12. eBPF
13. StatefulSet identity
14. Headless Service + StatefulSet DNS
15. volumeClaimTemplates
16. OrderedReady / Parallel
17. RollingUpdate / OnDelete / partition
18. StatefulSet failure handling
19. PDB
20. Split-brain limitations
```

**Interview rule:** Don't memorize long definitions. For networking, always think **"Where does the packet go?"** For StatefulSet, always think **"What identity, storage, and ordering does this Pod need?"**
