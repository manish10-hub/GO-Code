# Kubernetes Networking and DNS

## 1. Introduction

Kubernetes networking is one of the most important concepts to understand when working with Kubernetes in production.

Kubernetes networking is responsible for enabling communication between:

- Containers within the same Pod
- Pods on the same Node
- Pods on different Nodes
- Pods and Services
- External clients and Services
- Pods and external systems
- Kubernetes components such as API Server, CoreDNS, kubelet, etc.

A good understanding of Kubernetes networking requires understanding **Pod networking, CNI, Services, kube-proxy, DNS, Ingress, NetworkPolicy, and the underlying Linux networking concepts**.

---

# 2. Kubernetes Networking Model

Kubernetes follows a relatively simple networking model.

The key requirements are:

1. Every Pod gets its own IP address.
2. Pods on the same Node can communicate directly.
3. Pods on different Nodes can communicate directly without NAT in the normal Pod-to-Pod model.
4. Nodes can communicate with Pods.
5. Services provide a stable virtual IP and DNS name for a group of Pods.

For example:

```text
                    Kubernetes Cluster
                           |
          +----------------+----------------+
          |                                 |
       Node 1                             Node 2
          |                                 |
    +-----+------+                    +-----+------+
    |            |                    |            |
  Pod A        Pod B                 Pod C        Pod D
10.1.1.10    10.1.1.11            10.1.2.10    10.1.2.11
    |            |                    |            |
    +------------+--------------------+------------+
                 Pod Network
```

Pod A should be able to communicate directly with Pod C using:

```text
10.1.2.10
```

without the application needing to know which Node hosts Pod C.

---

# 3. Container Networking vs Pod Networking

A common interview question is:

> Does every container get its own IP address?

Normally, **no**.

The IP address is assigned to the **Pod**, not each individual container.

For example:

```text
Pod
10.1.1.20
 |
 +-- Container A
 |
 +-- Container B
```

Both containers share:

- Network namespace
- Pod IP
- Network interfaces
- Ports

Therefore, two containers in the same Pod can communicate using:

```text
localhost
```

For example:

```text
Container A → localhost:8080 → Container B
```

provided the application is listening on that port.

---

# 4. Linux Network Namespace

Kubernetes networking is heavily based on Linux networking concepts.

Each Pod normally gets its own network namespace.

A network namespace provides an isolated networking environment containing things such as:

- Network interfaces
- Routing table
- IP addresses
- iptables/nftables rules
- Network sockets

Example:

```text
Node
 |
 +---------------------------+
 |                           |
 | Network Namespace         |
 |                           |
 | Pod                       |
 |  eth0: 10.1.1.10          |
 |                           |
 +---------------------------+
```

Containers inside the same Pod share this network namespace.

---

# 5. What is CNI?

CNI stands for:

**Container Network Interface**

CNI is a standard interface used to configure networking for containers.

Kubernetes itself does not implement the complete Pod networking implementation.

Instead, Kubernetes relies on a CNI plugin.

Examples include:

- Calico
- Cilium
- Flannel
- AWS VPC CNI
- Azure CNI
- OVN-Kubernetes

When a Pod is created, the CNI plugin is responsible for tasks such as:

1. Creating/configuring the Pod network interface.
2. Assigning an IP address.
3. Configuring routes.
4. Connecting the Pod to the cluster network.

---

# 6. How a Pod Gets an IP Address

Suppose a Pod is created:

```text
kubectl create deployment nginx --image=nginx
```

The high-level flow is:

```text
kubectl
   |
   v
API Server
   |
   v
Scheduler
   |
   v
Kubelet
   |
   v
Container Runtime
   |
   v
CNI Plugin
   |
   +---- Create network namespace
   |
   +---- Create network interface
   |
   +---- Allocate Pod IP
   |
   +---- Configure routes
   |
   v
Pod
10.1.1.10
```

The exact implementation differs depending on the CNI.


Network namespace = a separate networking environment
Network interface = a network device inside that environment
Think of a network namespace as a room and a network interface as a door/window in that room.

A network namespace is an isolated networking environment containing its own interfaces, IP addresses, routing table, and sockets. A network interface is a virtual or physical network device inside that namespace through which packets are sent and received.

A network namespace (netns) provides an isolated networking environment.
A network interface is the thing through which packets enter or leave a network namespace.

```
Pod Network Namespace
┌─────────────────────────────┐
│                             │
│  eth0 <--- Network Interface│
│  IP: 10.244.1.5             │
│  MAC: aa:bb:cc:dd:ee:01     │
│                             │
└─────────────────────────────┘
```


---

# 7. Pod-to-Pod Communication

Consider:

```text
Node 1                         Node 2

Pod A                          Pod B
10.1.1.10                      10.1.2.10
   |                              |
   +------------ Network ---------+
```

Pod A sends traffic to:

```text
10.1.2.10
```

The networking layer determines how to route the packet to Node 2.

Depending on the CNI, this can be implemented using:

- Linux routing
- VXLAN
- Geneve
- IP-in-IP
- eBPF
- Cloud-native networking
- Other encapsulation/routing mechanisms

---

# 8. How Packets Leave a Pod

A simplified Linux networking setup may look like:

```text
Pod
10.1.1.10
   |
 eth0
   |
 veth pair
   |
Node network
   |
Routing / CNI
   |
External network
```

A common implementation uses a **veth pair**.

One end is inside the Pod's network namespace:

```text
Pod eth0
```

The other end exists on the Node.

Conceptually:

```text
Pod Namespace

eth0
 |
 | veth pair
 |
Node Namespace
 |
CNI / bridge / routing
```

---

# 9. Kubernetes Services

Pods are ephemeral.

A Pod can be:

- Deleted
- Recreated
- Rescheduled
- Given a different IP address

Therefore, applications should not normally depend directly on Pod IPs.

Kubernetes provides **Services** to provide a stable endpoint.

Example:

```text
             Service
          10.96.10.100
                |
        +-------+-------+
        |       |       |
      Pod A   Pod B   Pod C
      10.1.1.1 10.1.1.2 10.1.1.3
```

The Service provides a stable virtual IP called the **ClusterIP**.

---

# 10. Kubernetes Service Types

The common Service types are:

## ClusterIP

Default Service type.

Accessible inside the cluster.

```text
Pod
 |
 v
Service ClusterIP
 |
 +---- Pod
 +---- Pod
 +---- Pod
```

Example:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend
spec:
  type: ClusterIP
  selector:
    app: backend
  ports:
    - port: 80
      targetPort: 8080
```

---

## NodePort

Exposes the Service on a port on every Node.

Example:

```text
Node IP
192.168.1.10

NodePort
30080

Client
 |
 v
192.168.1.10:30080
 |
 v
Service
 |
 v
Pod
```

---

## LoadBalancer

Typically integrates with a cloud provider's load balancer.

Example:

```text
Internet
   |
   v
Cloud Load Balancer
   |
   v
Kubernetes Service
   |
   v
Pods
```

---

# 11. How Kubernetes Service Routing Works

Suppose:

```text
Service:

backend
ClusterIP = 10.96.20.10

Pods:

10.1.1.10
10.1.1.11
10.1.2.10
```

A client sends:

```text
10.96.20.10:80
```

The Service IP is virtual.

Kubernetes networking redirects the traffic to one of the backend Pods.

Historically, this has commonly been implemented using **kube-proxy** and iptables/IPVS.

Modern Kubernetes environments can also use eBPF-based implementations such as Cilium.

---

# 12. What is kube-proxy?

`kube-proxy` is a Kubernetes component responsible for implementing Service networking on Nodes in many Kubernetes configurations.

It watches:

- Services
- EndpointSlices

and configures networking rules accordingly.

For example:

```text
Service
10.96.20.10:80
       |
       v
kube-proxy rules
       |
 +-----+-----+-----+
 |           |     |
Pod A       Pod B Pod C
```

Depending on configuration, kube-proxy can use:

- iptables
- IPVS
- nftables in newer configurations

However, some networking solutions can replace kube-proxy functionality entirely.

---

# 13. Endpoints and EndpointSlices

A Service doesn't directly contain Pod IPs.

Kubernetes maintains backend information through:

**EndpointSlices**

Example:

```text
Service: backend

EndpointSlice:

10.1.1.10:8080
10.1.1.11:8080
10.1.2.10:8080
```

The Service selects Pods using labels.

Example:

```yaml
selector:
  app: backend
```

Pods matching:

```yaml
labels:
  app: backend
```

become Service endpoints.

---

# 14. Kubernetes DNS

DNS is extremely important in Kubernetes.

Kubernetes normally deploys a DNS service called:

**CoreDNS**

CoreDNS runs as Pods inside the cluster.

Typical architecture:

```text
Application Pod
      |
      | DNS query
      v
/etc/resolv.conf
      |
      v
Cluster DNS Service
      |
      v
CoreDNS
      |
      +---- Kubernetes API
      |
      +---- External DNS
```

---

# 15. /etc/resolv.conf Inside a Pod

Inside a typical Pod, you may see:

```bash
cat /etc/resolv.conf
```

Something similar to:

```text
nameserver 10.96.0.10
search default.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
```

The important fields are:

### nameserver

The IP address of the DNS Service.

### search

DNS search domains.

### ndots

Controls when the resolver attempts search-domain expansion.

---

# 16. Kubernetes DNS Naming

Suppose you have:

```text
Service:

backend

Namespace:

production
```

The fully qualified DNS name is:

```text
backend.production.svc.cluster.local
```

The structure is:

```text
<service>.<namespace>.svc.<cluster-domain>
```

Usually:

```text
<service>.<namespace>.svc.cluster.local
```

---

# 17. DNS Resolution Example

Suppose:

```text
Service:
backend

Namespace:
production
```

From another Pod:

```bash
curl http://backend.production.svc.cluster.local
```

DNS resolution:

```text
backend.production.svc.cluster.local
            |
            v
        CoreDNS
            |
            v
        Service IP
            |
            v
      Kubernetes Service
            |
            v
           Pod
```

---

# 18. Short DNS Names

You don't always need to use the full name.

From the same namespace:

```bash
curl http://backend
```

From another namespace:

```bash
curl http://backend.production
```

Fully qualified:

```bash
curl http://backend.production.svc.cluster.local
```

The search domains in `/etc/resolv.conf` allow shorter names to resolve.

---

# 19. Pod DNS

Kubernetes can also create DNS records for Pods.

Historically, Pod DNS records can look like:

```text
10-1-1-10.default.pod.cluster.local
```

However, applications should generally use **Services rather than Pod DNS/IP addresses**, unless there is a specific reason to address individual Pods.

---

# 20. Headless Services

A Headless Service is created using:

```yaml
clusterIP: None
```

Example:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: database
spec:
  clusterIP: None
  selector:
    app: database
```

Unlike a normal Service, there is no virtual ClusterIP.

DNS can return the individual Pod IPs.

```text
database.default.svc.cluster.local
             |
       +-----+-----+
       |     |     |
     Pod A Pod B Pod C
```

Headless Services are commonly useful for:

- Stateful applications
- Databases
- Kafka
- Distributed systems
- StatefulSets

---

# 21. CoreDNS Architecture

CoreDNS is deployed as Pods.

Example:

```text
                 Kubernetes API
                       |
                       |
                  +----v----+
                  | CoreDNS |
                  +----+----+
                       |
          +------------+------------+
          |                         |
   Kubernetes DNS             External DNS
          |                         |
          v                         v
Services/Pods                  Internet
```

CoreDNS uses the Kubernetes API to learn about:

- Services
- EndpointSlices
- Pods, depending on configuration

---

# 22. What Happens When a Pod Resolves a Service?

Suppose:

```bash
curl http://backend
```

The process is approximately:

```text
Application
    |
    v
glibc / resolver
    |
    v
/etc/resolv.conf
    |
    v
CoreDNS Service
    |
    v
CoreDNS Pod
    |
    v
Kubernetes DNS data
    |
    v
backend.default.svc.cluster.local
    |
    v
ClusterIP
```

Then the application connects to the Service IP.

---

# 23. DNS and External Domains

Suppose a Pod runs:

```bash
curl https://google.com
```

CoreDNS can forward external DNS queries to configured upstream DNS servers.

```text
Pod
 |
 v
CoreDNS
 |
 v
Upstream DNS
 |
 v
Internet DNS
```

This behavior depends on the CoreDNS configuration and cluster setup.

---

# 24. Ingress

Ingress provides HTTP/HTTPS routing into Kubernetes.

Example:

```text
Internet
    |
    v
Load Balancer
    |
    v
Ingress Controller
    |
    +--------+--------+
    |                 |
 /api              /web
    |                 |
    v                 v
Backend Service    Frontend Service
    |                 |
    v                 v
Backend Pods       Frontend Pods
```

Ingress is primarily concerned with **Layer 7 HTTP/HTTPS routing**.

Common Ingress Controllers include:

- NGINX Ingress Controller
- HAProxy
- Traefik
- Kong
- Cloud-provider controllers

---

# 25. NetworkPolicy

NetworkPolicy controls which network traffic is allowed between Pods and other endpoints.

For example:

```text
Frontend
   |
   | Allowed
   v
Backend

Frontend
   |
   X
Database
```

A NetworkPolicy can restrict:

- Ingress traffic
- Egress traffic
- Namespace-based communication
- Pod-to-Pod communication
- IP-based traffic

Example:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: frontend
```

This allows traffic from Pods labeled `app=frontend` to the backend.

---

# 26. NetworkPolicy Implementation

Kubernetes defines the NetworkPolicy API, but the actual enforcement is done by the networking implementation/CNI.

Examples:

```text
Kubernetes
     |
NetworkPolicy API
     |
     v
CNI
     |
     +---- iptables
     |
     +---- eBPF
     |
     +---- other mechanisms
```

A NetworkPolicy object by itself does not necessarily enforce traffic restrictions unless the cluster's networking implementation supports NetworkPolicy.

---

# 27. Node-to-Pod Networking

Consider:

```text
Node
192.168.1.10

Pod
10.1.1.10
```

The Node needs a route to the Pod network.

The CNI configures the necessary networking.

For example:

```text
Node
192.168.1.10
     |
     | route
     v
10.1.1.0/24
     |
     +---- Pod
     +---- Pod
```

The exact routing mechanism depends on the CNI.

---

# 28. Pod-to-Internet Traffic

Suppose a Pod accesses:

```text
https://example.com
```

A simplified flow could be:

```text
Pod
 |
 v
Pod network
 |
 v
Node
 |
 v
NAT
 |
 v
Node/Cloud network
 |
 v
Internet
```

Typically, traffic leaving the cluster may be SNATed depending on the networking implementation and environment.

---

# 29. Important Linux Networking Concepts

For Kubernetes networking interviews, understand these Linux concepts:

### Network namespace

Provides network isolation.

### veth pair

Virtual Ethernet pair connecting network namespaces.

### Bridge

Layer 2 forwarding between interfaces.

### Routing table

Determines where packets should go.

### iptables/nftables

Packet filtering and NAT.

### IPVS

Kernel-based load balancing mechanism.

### eBPF

Allows programmable packet processing in the Linux kernel.

---

# 30. Common Packet Flow

A useful mental model is:

```text
Application
    |
    v
Socket
    |
    v
Pod eth0
    |
    v
veth
    |
    v
Node networking
    |
    +---- Service rules
    |
    +---- Routing
    |
    +---- NetworkPolicy
    |
    v
CNI
    |
    v
Other Node / External Network
```

---

# 31. How to Troubleshoot Kubernetes Networking

A good troubleshooting process is:

### Step 1 — Check Pod IP

```bash
kubectl get pods -o wide
```

### Step 2 — Check Service

```bash
kubectl get svc
```

### Step 3 — Check EndpointSlices

```bash
kubectl get endpointslices
```

### Step 4 — Test DNS

```bash
nslookup backend
```

or:

```bash
dig backend.default.svc.cluster.local
```

### Step 5 — Check Pod's DNS configuration

```bash
kubectl exec -it <pod> -- cat /etc/resolv.conf
```

### Step 6 — Test connectivity

```bash
kubectl exec -it <pod> -- curl backend:8080
```

### Step 7 — Check NetworkPolicy

```bash
kubectl get networkpolicy
```

### Step 8 — Check CNI

Depending on your environment, inspect the CNI Pods:

```bash
kubectl get pods -n kube-system
```

### Step 9 — Check routes

On the Node:

```bash
ip route
```

### Step 10 — Inspect packet flow

For advanced debugging:

```bash
tcpdump
```

and inspect:

```bash
iptables
```

or:

```bash
nft
```

---

# 32. Important Interview Questions

## Beginner Level

### Q1. What is Kubernetes networking?

Kubernetes networking provides communication between Pods, Services, Nodes and external systems.

### Q2. Does every Pod get an IP?

Yes. Kubernetes networking is designed around each Pod having its own IP.

### Q3. Do individual containers get different IPs?

Normally no. Containers within a Pod share the Pod's network namespace and IP.

### Q4. Why do we need Services?

Pods are ephemeral and their IP addresses can change. Services provide a stable endpoint.

### Q5. What is ClusterIP?

ClusterIP is the virtual IP assigned to a Kubernetes Service for internal cluster communication.

### Q6. What is CNI?

CNI is the standard interface used to configure container networking. Kubernetes relies on CNI plugins for Pod networking.

### Q7. What is CoreDNS?

CoreDNS provides DNS resolution inside the Kubernetes cluster.

---

# 33. Intermediate Interview Questions

### Q8. Explain Pod-to-Pod communication.

Explain:

```text
Pod
 |
v
Pod network interface
 |
v
CNI
 |
v
Node routing/encapsulation
 |
v
Destination Node
 |
v
Destination Pod
```

Mention that the exact implementation depends on the CNI.

### Q9. How does a Service route traffic to Pods?

Service traffic is implemented through mechanisms such as kube-proxy rules or CNI/eBPF-based implementations.

The Service selects Pods using labels, and EndpointSlices maintain the backend endpoints.

### Q10. What is kube-proxy?

kube-proxy watches Services and EndpointSlices and implements Service traffic forwarding on Nodes in many Kubernetes configurations.

### Q11. What is an EndpointSlice?

EndpointSlices contain the network endpoints associated with Services.

### Q12. What is a Headless Service?

A Service with:

```yaml
clusterIP: None
```

It does not provide a virtual ClusterIP and can return individual Pod IPs through DNS.

### Q13. What is NetworkPolicy?

NetworkPolicy defines allowed network traffic between Pods and other network endpoints.

### Q14. How does DNS work in Kubernetes?

Explain:

```text
Pod
 |
/etc/resolv.conf
 |
CoreDNS Service
 |
CoreDNS
 |
Kubernetes API
 |
Service/Pod DNS records
```

---

# 34. Advanced Interview Questions

### Q15. Explain what happens when Pod A calls Service B.

A strong answer should cover:

```text
Application
    |
DNS lookup
    |
CoreDNS
    |
Service ClusterIP
    |
kube-proxy / eBPF
    |
Endpoint
    |
Pod B
```

Then explain that the exact packet transformation depends on the CNI and Service implementation.

---

### Q16. What happens when a Pod calls another Pod on a different Node?

Discuss:

```text
Pod A
 |
CNI
 |
Node A
 |
Routing / encapsulation
 |
Node B
 |
CNI
 |
Pod B
```

Explain possible implementations:

- Native routing
- VXLAN
- Geneve
- IP-in-IP
- eBPF

---

### Q17. What happens when DNS resolution fails inside a Pod?

Troubleshoot:

```text
Pod
 |
/etc/resolv.conf
 |
DNS Service
 |
CoreDNS
 |
Upstream DNS
```

Check:

```bash
kubectl get pods -n kube-system
kubectl get svc -n kube-system
kubectl logs -n kube-system <coredns-pod>
```

Also check:

```bash
kubectl exec <pod> -- cat /etc/resolv.conf
```

---

### Q18. What is the difference between Service DNS and Pod DNS?

Service DNS points to the Service's virtual endpoint or, for headless Services, can resolve to individual endpoints.

Pod DNS can identify individual Pods, but applications generally prefer Services for stable communication.

---

### Q19. What is the difference between kube-proxy and CNI?

**CNI** primarily handles Pod networking.

**kube-proxy** traditionally implements Kubernetes Service networking.

Modern CNI implementations such as Cilium can provide both networking and Service load-balancing functionality using eBPF.

---

### Q20. Explain ClusterIP vs NodePort vs LoadBalancer.

```text
ClusterIP
    Internal cluster access

NodePort
    Node IP + port

LoadBalancer
    External/cloud load balancer
```

---

# 35. Very Advanced Questions

For a senior/Tech Lead Kubernetes interview, expect questions such as:

1. Explain the complete packet path from Pod A to Pod B across Nodes.
2. How does kube-proxy implement Service load balancing?
3. Compare iptables vs IPVS vs eBPF.
4. What happens when a Service has no endpoints?
5. What happens when a Pod is not ready?
6. How does EndpointSlice improve upon Endpoints?
7. How does Kubernetes DNS discover Services?
8. What is `ndots` and why can it affect DNS performance?
9. Why might DNS resolution be slow inside Kubernetes?
10. How would you debug intermittent DNS failures?
11. How does NetworkPolicy enforcement work?
12. Does NetworkPolicy apply to Services or Pods?
13. How does NAT work in Kubernetes?
14. What is SNAT vs DNAT?
15. How does NodePort work internally?
16. What is `externalTrafficPolicy`?
17. Difference between `externalTrafficPolicy: Cluster` and `Local`.
18. What is `internalTrafficPolicy`?
19. What happens when a Pod communicates with a Service on the same Node?
20. What happens when a Pod communicates with a Service whose endpoint is on another Node?
21. How does a CNI assign Pod IPs?
22. What is IPAM?
23. What happens when the Pod CIDR is exhausted?
24. How does Kubernetes networking work in a cloud environment?
25. How would you troubleshoot packet drops between two Pods?

---

# 36. Questions You Should Be Able to Draw on a Whiteboard

For a senior Kubernetes interview, practice drawing these five diagrams.

### Diagram 1 — Pod networking

```text
Container
   |
Network Namespace
   |
eth0
   |
veth
   |
Node
```

### Diagram 2 — Pod-to-Pod across Nodes

```text
Pod A
 |
Node A
 |
CNI
 |
Network
 |
Node B
 |
CNI
 |
Pod B
```

### Diagram 3 — Service networking

```text
Client Pod
 |
Service ClusterIP
 |
kube-proxy / eBPF
 |
EndpointSlice
 |
Backend Pod
```

### Diagram 4 — DNS

```text
Application
 |
Resolver
 |
/etc/resolv.conf
 |
CoreDNS
 |
Kubernetes API
 |
Service
```

### Diagram 5 — Internet access

```text
Pod
 |
CNI
 |
Node
 |
SNAT
 |
Cloud Network
 |
Internet
```

---

# 37. Key Commands to Remember

```bash
# Pods and IPs
kubectl get pods -o wide

# Services
kubectl get svc

# Endpoints
kubectl get endpointslices

# Network policies
kubectl get networkpolicy

# DNS configuration
kubectl exec <pod> -- cat /etc/resolv.conf

# DNS testing
kubectl exec <pod> -- nslookup kubernetes.default

# Connectivity testing
kubectl exec <pod> -- curl <service>:<port>

# Routes
ip route

# Interfaces
ip addr

# Network namespaces
ip netns

# iptables
iptables -L -n

# nftables
nft list ruleset

# Packet capture
tcpdump
```

---

# 38. The Most Important Mental Model

If you remember only one thing, remember this:

```text
                    Kubernetes Networking

                         DNS
                          |
                       CoreDNS
                          |
                          v
Application ---> Service ---> EndpointSlice ---> Pod
                    |
                    |
             kube-proxy / eBPF
                    |
                    v
                   CNI
                    |
             +------+------+
             |             |
          Node A         Node B
             |             |
           Pod A         Pod B
```

Think about Kubernetes networking as **four separate problems**:

### 1. Pod Networking

How does a Pod get an IP?

→ **CNI + IPAM**

### 2. Pod-to-Pod Networking

How does one Pod reach another Pod?

→ **CNI + routing/encapsulation/eBPF**

### 3. Service Networking

How does an application reach a stable backend?

→ **Service + EndpointSlice + kube-proxy/eBPF**

### 4. DNS

How does the application discover the Service?

→ **CoreDNS + Service DNS**

Once these four concepts are clear, most Kubernetes networking interview questions become much easier.

---

# 39. Recommended Interview Preparation Order

For a senior/Tech Lead position, study in this order:

```text
1. Linux networking basics
       ↓
2. Network namespaces
       ↓
3. veth pairs
       ↓
4. CNI
       ↓
5. Pod networking
       ↓
6. Services
       ↓
7. kube-proxy
       ↓
8. EndpointSlices
       ↓
9. CoreDNS
       ↓
10. NetworkPolicy
       ↓
11. Ingress
       ↓
12. iptables / IPVS
       ↓
13. eBPF
       ↓
14. Advanced packet-flow troubleshooting
```

For a **senior Kubernetes interview**, I would particularly focus on being able to answer **"What happens to the packet?"** rather than memorizing definitions. Interviewers often start with a simple question such as:

> "A Pod sends a request to another Pod on a different Node. Walk me through the packet."

If you can explain the packet path from **application → network namespace → veth → CNI → Node → routing/encapsulation → destination Node → Pod**, and then separately explain **Service → EndpointSlice → kube-proxy/eBPF** and **DNS → CoreDNS**, you will have a strong foundation for Kubernetes networking interviews.




##############. Claude


# Kubernetes Networking — Communication Scenarios
### Detailed Explanations + Interview Q&A (Beginner → Expert)

---

## Table of Contents

1. [Communication between containers within the same Pod](#1-communication-between-containers-within-the-same-pod)
2. [Communication between Pods on the same Node](#2-communication-between-pods-on-the-same-node)
3. [Communication between Pods on different Nodes](#3-communication-between-pods-on-different-nodes)
4. [Communication between Pods and Services](#4-communication-between-pods-and-services)
5. [Communication between external clients and Services](#5-communication-between-external-clients-and-services)
6. [Communication between Pods and external systems](#6-communication-between-pods-and-external-systems)

---

## 1. Communication between containers within the same Pod

### Detailed Explanation

All containers within the same pod share the same **network namespace**. This is the most fundamental concept in Kubernetes networking and it means:

- All containers in a pod share **the same IP address**
- All containers in a pod share **the same set of ports**
- All containers can communicate with each other via `localhost`

#### How it works — the Pause container

When a pod is created, Kubernetes first starts a special hidden container called the **Pause container** (also called the **infra container** or **sandbox container**). This container does almost nothing — it just holds the network namespace open. All other containers in the pod then join this pre-existing network namespace rather than creating their own.

```
┌─────────────────────────────────────────────────────────────┐
│                          POD                                 │
│   Shared Network Namespace (one IP: 10.244.1.5)             │
│                                                             │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────┐   │
│  │   Pause     │   │    App      │   │    Sidecar      │   │
│  │  container  │   │  container  │   │   container     │   │
│  │ (holds ns)  │   │  :8080      │   │   :9090         │   │
│  └─────────────┘   └─────────────┘   └─────────────────┘   │
│                          │                    │             │
│              localhost:8080           localhost:9090         │
│                                                             │
│  Shared:  /proc/net, iptables, eth0, lo, IP address        │
└─────────────────────────────────────────────────────────────┘
```

#### Communication methods within a pod

**1. localhost (loopback) — most common**
Container A can call Container B using `localhost:<port>` or `127.0.0.1:<port>` because they share the same network namespace. A sidecar proxy (like Envoy) listening on port 15001 can intercept traffic from the app container on port 8080 — both use `localhost`.

**2. Shared volumes — for file-based IPC**
Containers in the same pod can share a volume mounted at the same path. One container writes a file; another reads it. `emptyDir` is the most common volume type for this purpose.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: shared-volume-pod
spec:
  containers:
  - name: writer
    image: busybox
    command: ["sh", "-c", "echo hello > /shared/data.txt && sleep 3600"]
    volumeMounts:
    - name: shared-data
      mountPath: /shared
  - name: reader
    image: busybox
    command: ["sh", "-c", "cat /shared/data.txt && sleep 3600"]
    volumeMounts:
    - name: shared-data
      mountPath: /shared
  volumes:
  - name: shared-data
    emptyDir: {}
```

**3. Process signals via shared PID namespace (optional)**
If `shareProcessNamespace: true` is set in the pod spec, containers share the same PID namespace and can send signals to each other's processes.

#### Port conflict rule
Since containers share one network namespace, two containers in the same pod cannot both listen on the same port. If container A listens on `:8080`, container B cannot also listen on `:8080` — this is a pod-level conflict, not a host-level conflict.

---

### Interview Questions — Scenario 1

#### Beginner

**Q1. How do containers within the same pod communicate with each other?**

Containers within the same pod share the same network namespace, which means they share the same IP address and port space. They communicate with each other using `localhost` — for example, if one container runs a web server on port 8080, another container in the same pod can reach it at `localhost:8080`. No service or DNS resolution is needed because they are effectively on the same virtual network interface.

---

**Q2. What is the Pause container and what role does it play?**

The Pause container (also called the infra or sandbox container) is a special container that Kubernetes creates first when starting a pod. Its only job is to hold the network namespace open — it creates the network namespace that all other containers in the pod will share. Without the Pause container, if the first application container crashed, its network namespace would be destroyed and all other containers would lose their network identity. The Pause container keeps the namespace alive regardless of what happens to the application containers.

---

**Q3. Can two containers in the same pod both listen on port 8080?**

No. Since all containers in a pod share the same network namespace and the same IP address, they also share the same port space. If container A is already listening on port 8080, container B cannot also bind to port 8080 — it would get an "address already in use" error. Each container must listen on a unique port within the pod.

---

#### Intermediate

**Q4. You have an application container and a sidecar logging container in the same pod. The sidecar needs to read logs that the app writes to a file. How do you implement this?**

Use a shared `emptyDir` volume mounted in both containers at the same path. The application container writes log files to `/var/log/app/`, and the sidecar container mounts the same `emptyDir` volume at `/var/log/app/` and reads from it. The `emptyDir` volume is created when the pod starts and deleted when it stops — its lifecycle matches the pod. This is the standard Sidecar pattern for log collection used by tools like Fluentd and Filebeat.

```yaml
volumes:
- name: log-volume
  emptyDir: {}
containers:
- name: app
  volumeMounts:
  - name: log-volume
    mountPath: /var/log/app
- name: log-shipper
  volumeMounts:
  - name: log-volume
    mountPath: /var/log/app
```

---

**Q5. What happens to inter-container communication if one container in the pod crashes and restarts?**

The network namespace is not destroyed when an individual container crashes — it is owned by the Pause container, which continues running. When the crashed container restarts, it rejoins the same network namespace with the same IP address and port bindings. Other containers in the pod can continue communicating normally during the restart window, though they will get connection refused errors if they try to reach the crashed container's port during the brief period it is restarting.

---

**Q6. How does an Istio sidecar proxy intercept traffic between a pod and external services without the application knowing?**

Istio injects an Envoy sidecar proxy container into the pod. It also injects an init container that configures iptables rules inside the pod's shared network namespace. These iptables rules redirect all outbound TCP traffic (except from Envoy itself and a few excluded ports) to Envoy's port 15001, and all inbound traffic to port 15006. Since both the app container and Envoy share the same network namespace, these iptables rules apply to all containers. The app thinks it is connecting directly to external services, but all traffic goes through Envoy, which applies mTLS, retries, circuit breaking, and telemetry transparently.

---

#### Expert

**Q7. Explain the Linux kernel primitives that make shared network namespaces possible. What exactly is a network namespace?**

A Linux network namespace (`netns`) is a kernel-level isolation primitive that provides an independent copy of the network stack — its own set of network interfaces, routing tables, iptables rules, sockets, and the loopback interface. Processes in different network namespaces are completely isolated at the network level.

When Kubernetes creates a pod, the container runtime (containerd, CRI-O) calls the Linux `clone()` or `unshare()` syscall with the `CLONE_NEWNET` flag to create a new network namespace for the Pause container. Each subsequent container in the pod joins this existing namespace using the `setns()` syscall, passing the file descriptor of the Pause container's `/proc/<pid>/ns/net`. This is why all containers see the same `eth0` interface and the same IP — they are all operating in the same kernel network namespace object.

                    Linux Network Namespace
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  1. Socket                                             │
│     "Which application/port is using the network?"     │
│                                                         │
│  2. Routing Table                                      │
│     "Where should this packet go?"                     │
│                                                         │
│  3. iptables Rules                                     │
│     "Should I allow/block/modify this packet?"         │
│                                                         │
│  4. Loopback Interface (lo)                            │
│     "How does a process talk to itself?"               │
│                                                         │
└─────────────────────────────────────────────────────────┘

Application
     │
     │ Socket
     ▼
┌──────────────────────────────────────────┐
│ Network Namespace                        │
│                                          │
│  Socket                                  │
│    │                                     │
│    ▼                                     │
│  Linux Network Stack                     │
│    │                                     │
│    ├── iptables rules                    │
│    │       │                             │
│    │       └── Allow / Drop / NAT        │
│    │                                     │
│    └── Routing table                     │
│            │                             │
│            └── Which interface/gateway?  │
│                    │                     │
│             ┌──────┴──────┐              │
│             ▼             ▼              │
│            eth0           lo             │
│             │             │              │
└─────────────┼─────────────┼──────────────┘
              │             │
              ▼             ▼
           Network       Local machine

---

**Q8. In a high-throughput system, should you prefer localhost communication between containers in the same pod over service-to-service calls to a separate pod? What are the trade-offs?**

**Advantages of localhost (same pod):**
- Zero network overhead — traffic goes through the loopback interface (no physical NIC, no CNI plugin overhead, no iptables traversal for service routing)
- No additional latency from DNS resolution or service IP translation
- No extra hops — sub-microsecond latency vs 1-3ms for cross-pod calls

**Disadvantages:**
- Tight coupling — both containers scale together; you cannot independently scale the sidecar vs the app
- Failure coupling — if the pod is evicted or killed, both containers go down together
- Harder to update independently — updating one container restarts the whole pod

**When to use same-pod:** when the sidecar is strictly a helper to the main container (Envoy proxy, log shipper, secrets injector) and its lifecycle is inherently tied to the app. When to use separate pods: when the "sidecar" needs independent scaling, independent deployment, or serves multiple consumers — at that point it should be its own service with its own pod.

---

## 2. Communication between Pods on the same Node

### Detailed Explanation

When two pods are on the same node, they have different IP addresses (each pod has its own network namespace) but they communicate through the node's **virtual ethernet (veth) pairs** and a **network bridge**.

#### The network bridge and veth pairs

When a pod is created on a node, the CNI (Container Network Interface) plugin creates a **veth pair** — a pair of connected virtual network interfaces. One end (`veth0`) goes inside the pod's network namespace (visible as `eth0` inside the pod), and the other end (`vethXXXXXX`) stays in the node's root network namespace and is attached to a **bridge interface** (usually named `cni0` or `cbr0` depending on the CNI plugin).

```
NODE (192.168.1.10)
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  Pod A (10.244.1.2)          Pod B (10.244.1.3)                 │
│  ┌──────────────┐            ┌──────────────┐                   │
│  │   eth0       │            │   eth0       │                   │
│  │ 10.244.1.2   │            │ 10.244.1.3   │                   │
│  └──────┬───────┘            └──────┬───────┘                   │
│         │ veth pair                 │ veth pair                 │
│         │                           │                           │
│  ┌──────┴───────────────────────────┴───────┐                   │
│  │              Bridge (cni0)               │                   │
│  │           10.244.1.1/24                  │                   │
│  └──────────────────────────────────────────┘                   │
│                          │                                      │
│                    (node's eth0)                                 │
│                   192.168.1.10                                   │
└──────────────────────────────────────────────────────────────────┘

Traffic: Pod A → Pod B
  10.244.1.2 → veth-a → bridge → veth-b → 10.244.1.3
  All at Layer 2 (ARP + MAC addresses) — no routing required
```

#### The ARP and MAC resolution

When Pod A wants to talk to Pod B:
1. Pod A checks its routing table — `10.244.1.0/24` is on the same subnet, reachable via `eth0`
2. Pod A sends an ARP request: "Who has 10.244.1.3?"
3. The bridge responds with the MAC address of veth-b
4. Pod A sends the Ethernet frame directly to veth-b's MAC address
5. The bridge forwards the frame to veth-b (Pod B's interface)
6. Pod B receives the packet on its `eth0`

This is Layer 2 communication — no IP routing is needed, just MAC address resolution through the bridge.

#### No NAT — a Kubernetes requirement

The Kubernetes networking model requires that **pods can communicate with all other pods without NAT**. On the same node, this is naturally satisfied because the bridge routes packets directly between pod namespaces.

---

### Interview Questions — Scenario 2

#### Beginner

**Q1. Two pods are running on the same node. How do they communicate?**

Each pod has its own network namespace and its own IP address. They communicate through the node's virtual network bridge. Each pod is connected to the bridge through a veth pair — a virtual ethernet cable where one end is inside the pod and the other is attached to the bridge. When Pod A sends a packet to Pod B's IP, it goes through Pod A's `eth0`, across its veth pair to the bridge, which then forwards it across Pod B's veth pair to Pod B's `eth0`. This is all local to the node — no physical network is involved.

---

**Q2. What is a veth pair and why does Kubernetes use them?**

A veth pair is a pair of connected virtual network interfaces — anything sent into one end comes out the other end, like a virtual network cable. Kubernetes uses veth pairs to connect a pod's isolated network namespace to the node's network bridge. One end of the veth pair is placed inside the pod's network namespace (visible as `eth0` inside the pod), and the other end is placed in the node's root namespace and attached to the bridge. This gives pods network connectivity without giving them access to the node's actual physical network interface directly.

---

#### Intermediate

**Q3. What is the role of the network bridge (cni0) on a Kubernetes node?**

The bridge (`cni0`, `cbr0`, or similar depending on CNI plugin) acts as a virtual Layer 2 switch on the node. All pod veth pair ends are attached to this bridge. When a packet arrives at the bridge from one veth pair, the bridge uses MAC address tables (just like a physical switch) to determine which veth pair to forward it to. The bridge also has an IP address (typically the `.1` address of the pod CIDR subnet) which acts as the default gateway for pods — when a pod needs to send traffic outside the node, it sends it to the bridge's IP, and the bridge forwards it to the node's physical interface.

---

**Q4. Why does same-node pod-to-pod communication not require NAT?**

The Kubernetes networking model explicitly requires that pod-to-pod communication happens without Network Address Translation (NAT). On the same node, this is naturally satisfied because the bridge directly connects pod namespaces using their real pod IP addresses — there is no address translation at any layer. Pod A's packets arrive at Pod B with Pod A's real source IP intact, and Pod B's responses return to Pod A's real IP. This is essential for application-level logic that depends on knowing the source IP of the caller (access control, logging, circuit breakers).

---

**Q5. How does a CNI plugin like Flannel or Calico set up same-node pod communication?**

When a new pod is created, the kubelet calls the CNI plugin via a standardized interface. The CNI plugin:
1. Creates a new network namespace for the pod
2. Creates a veth pair
3. Moves one end of the veth pair into the pod's network namespace and renames it `eth0`
4. Assigns the pod's IP address to `eth0` inside the pod namespace
5. Attaches the other end of the veth pair to the node's bridge interface (`cni0`)
6. Adds routes inside the pod namespace: default route via the bridge IP
7. Adds a route on the node: `<pod-IP>/32 via <veth-on-node>`

For same-node traffic, step 5 is sufficient — the bridge handles forwarding. For cross-node traffic, additional configuration (tunnels, BGP routes) is needed.

---

#### Expert

**Q6. Describe the exact packet journey from Pod A (10.244.1.2) to Pod B (10.244.1.3) on the same node at the kernel level, including which network namespaces and kernel subsystems are traversed.**

1. **Application layer (Pod A's namespace):** the application calls `connect()` / `sendto()` → kernel creates a socket, determines the destination IP is on-link (same subnet as `eth0`'s network `10.244.1.0/24`)

2. **ARP resolution (Pod A's namespace):** kernel checks ARP cache for `10.244.1.3`. On cache miss, sends ARP request out of `eth0` (the veth inside the pod namespace). The ARP request travels through the veth pair to `veth-a` on the node bridge. The bridge, knowing the MAC of `veth-b` (Pod B's veth end), responds.

3. **Packet construction (Pod A's namespace):** kernel builds an Ethernet frame: src MAC = Pod A's `eth0` MAC, dst MAC = `veth-b`'s MAC, src IP = 10.244.1.2, dst IP = 10.244.1.3

4. **veth transmit (kernel → node namespace):** the packet exits Pod A's `eth0` into the veth pair kernel code, which delivers it to `veth-a` in the root network namespace

5. **Bridge forwarding (root namespace):** the bridge receives the frame on `veth-a`, looks up `veth-b`'s MAC in its forwarding database, and forwards the frame out of `veth-b`

6. **veth receive (node namespace → Pod B namespace):** the frame enters `veth-b` in the root namespace, travels through the veth pair, and emerges at `eth0` inside Pod B's network namespace

7. **TCP stack (Pod B's namespace):** Pod B's kernel receives the Ethernet frame, strips the Ethernet header, processes the IP packet, strips the IP header, delivers the TCP segment to the listening socket

Total kernel namespaces traversed: 2 (Pod A's netns → root netns → Pod B's netns). No iptables PREROUTING/POSTROUTING for NAT (no NAT needed). iptables rules may still be traversed for NetworkPolicy enforcement depending on CNI plugin.

---

## 3. Communication between Pods on different Nodes

### Detailed Explanation

Cross-node pod communication is the hardest networking problem in Kubernetes. Each node has its own pod CIDR subnet (e.g., Node 1: `10.244.1.0/24`, Node 2: `10.244.2.0/24`), and pods on different nodes are on different subnets. The traffic must cross the physical network infrastructure.

The Kubernetes requirement: **pods on different nodes must be able to communicate without NAT**. There are two main approaches CNI plugins use to satisfy this:

#### Approach 1 — Overlay networks (tunneling)

The CNI plugin encapsulates the original pod-to-pod packet inside another packet for transit across the physical network, then de-encapsulates it at the destination node.

**VXLAN (used by Flannel in VXLAN mode, Calico in VXLAN mode):**
```
Node 1 (192.168.1.10)              Node 2 (192.168.1.20)
┌─────────────────────┐            ┌─────────────────────┐
│ Pod A: 10.244.1.2   │            │ Pod B: 10.244.2.3   │
│      ↓              │            │      ↑              │
│   veth/bridge       │            │   veth/bridge       │
│      ↓              │            │      ↑              │
│  flannel.1 (VTEP)   │            │  flannel.1 (VTEP)   │
│  Encapsulate:       │            │  De-encapsulate:    │
│  [UDP][VXLAN hdr]   │            │  strip outer headers│
│  [IP: 10.244.2.3]   │            │  deliver inner pkt  │
│      ↓              │            │      ↑              │
│    eth0             │────────────│    eth0             │
│ 192.168.1.10        │  physical  │ 192.168.1.20        │
└─────────────────────┘  network   └─────────────────────┘

Outer packet: src=192.168.1.10, dst=192.168.1.20, UDP port 8472
Inner packet: src=10.244.1.2,   dst=10.244.2.3
```

**Overhead:** VXLAN adds ~50 bytes of headers per packet (UDP + VXLAN header), reducing effective MTU. Requires VXLAN kernel module support. The VTEP (VXLAN Tunnel Endpoint) is typically a network interface (`flannel.1`) on each node.

#### Approach 2 — Underlay routing (BGP)

The CNI plugin programs the physical network to route pod CIDRs directly — no encapsulation needed.

**Calico with BGP:**
- Each node runs a BGP speaker (BIRD or GoBGP)
- Each node advertises its pod CIDR (`10.244.1.0/24` from Node 1) to the BGP peers (other nodes or a BGP route reflector)
- The physical routers/switches learn that `10.244.1.0/24` is reachable via Node 1's IP
- Packets between pods are routed at Layer 3 by the physical network with no encapsulation

```
Node 1 routes:            Physical network router routes:
10.244.1.0/24 → local    10.244.1.0/24 → 192.168.1.10
10.244.2.0/24 → 192.168.1.20    10.244.2.0/24 → 192.168.1.20
```

**Advantages:** no encapsulation overhead, full MTU, better performance. **Disadvantages:** requires BGP support in the physical network; not all environments support this.

---

### Interview Questions — Scenario 3

#### Beginner

**Q1. How does a pod on Node 1 communicate with a pod on Node 2?**

Each node has its own subnet of pod IPs. When Pod A on Node 1 wants to reach Pod B on Node 2, the packet travels from Pod A through the node's bridge to Node 1's physical network interface. At this point, the CNI plugin's networking takes over — either encapsulating the packet in a tunnel (like VXLAN) that the physical network carries to Node 2, or using network routing (BGP) so the physical network knows how to route pod-subnet traffic directly to Node 2. At Node 2, the packet is delivered to Pod B through Node 2's bridge and veth pair. The key Kubernetes requirement is that this happens without NAT — Pod B sees Pod A's real IP as the source.

---

**Q2. What is a CNI plugin? Why does Kubernetes need one?**

CNI (Container Network Interface) is a standard specification for network plugins in container orchestration systems. Kubernetes handles everything except the actual pod network setup — it delegates that responsibility to a CNI plugin. When a pod is created, Kubernetes calls the CNI plugin to set up the pod's network namespace, assign an IP address, create veth pairs, and configure routing. Different CNI plugins (Flannel, Calico, Weave, Cilium) implement the same CNI interface but use different approaches for pod networking. Without a CNI plugin, pods would be created but would have no network connectivity.

---

#### Intermediate

**Q3. Compare VXLAN overlay networking and BGP-based underlay routing. When would you choose each?**

**VXLAN overlay (Flannel VXLAN mode, Calico VXLAN mode):**
- Encapsulates pod packets inside UDP packets for transit
- Works on any network — cloud VPCs, bare metal, on-premises
- No special network hardware configuration needed
- Overhead: ~50 bytes per packet, slightly higher CPU usage for encap/decap
- Choose when you do not control the physical network or when the network does not support BGP

**BGP underlay (Calico BGP mode, Cilium with BGP):**
- Programs physical network routing tables with pod CIDR routes
- No encapsulation — full native packet throughput
- Requires BGP-capable network equipment and peering configuration
- Lower latency and higher throughput — important for high-performance workloads
- Choose when you control the physical network and need maximum network performance

**In cloud environments (AWS, GCP, Azure):** both approaches work, but cloud providers also offer native VPC routing integration (AWS VPC CNI assigns actual VPC IPs to pods — no encapsulation at all).

---

**Q4. What is the pod CIDR and how does Kubernetes ensure unique IP addresses across all nodes?**

The pod CIDR is the range of IP addresses assigned to pods. Kubernetes allocates a non-overlapping subnet of the cluster pod CIDR to each node when it joins. For example, cluster CIDR `10.244.0.0/16` might be divided as:
- Node 1: `10.244.1.0/24` (256 pod IPs)
- Node 2: `10.244.2.0/24`
- Node 3: `10.244.3.0/24`

The controller manager manages this allocation using the `--cluster-cidr` flag and records each node's allocation in the `spec.podCIDR` field of the Node object. The CNI plugin reads this field and assigns IPs from the node's subnet when creating pods. Since each node's subnet is non-overlapping, pod IPs are globally unique across the cluster.

---

#### Expert

**Q5. Explain how Cilium uses eBPF to implement cross-node pod networking differently from traditional CNI plugins.**

Traditional CNI plugins rely on Linux kernel networking primitives: veth pairs, bridges, iptables/ipvs for service routing, and VXLAN tunnels for cross-node traffic. Each of these involves multiple kernel subsystem traversals per packet.

Cilium replaces most of this with eBPF programs attached at strategic kernel hook points:

**XDP (eXpress Data Path):** eBPF programs at the NIC driver level — processes packets before they enter the kernel networking stack. Used for early-stage packet dropping (NetworkPolicy) and VXLAN decapsulation at wire speed.

**TC (Traffic Control) hooks:** eBPF programs at the network interface level — replaces the bridge and veth approach entirely. Cilium attaches eBPF programs directly to the physical NIC's `tc ingress/egress` hooks and to the pod's veth pair, creating a direct forwarding path that bypasses the bridge.

**Cross-node in Cilium:**
1. Packet exits pod → enters `tc egress` eBPF on veth pair
2. eBPF program looks up destination pod's node in a BPF map (populated by Cilium's control plane from Kubernetes endpoints)
3. If on a different node: encapsulate in GENEVE/VXLAN and send to destination node's IP directly — no bridge traversal, no iptables, no routing table lookup
4. At destination node: `XDP` eBPF program decapsulates and redirects directly to destination pod's veth — again bypassing the bridge

**Result:** Cilium achieves near-native performance, eliminates conntrack table bottlenecks, and enables per-packet NetworkPolicy enforcement without iptables — which is critical at Kubernetes cluster scale where iptables rules count in the tens of thousands.

---

## 4. Communication between Pods and Services

### Detailed Explanation

A Service is a stable virtual IP (ClusterIP) that provides load-balanced access to a set of pods. Pods are ephemeral — their IPs change when they restart. Services solve this by providing a permanent address that always routes to healthy pods.

#### The Service ClusterIP is virtual

The Service ClusterIP does **not** correspond to any real network interface. No pod or node has `10.96.0.1` assigned to its interface. The ClusterIP exists only in iptables/IPVS rules managed by kube-proxy on every node.

#### How kube-proxy works (iptables mode)

kube-proxy watches the Kubernetes API for Service and Endpoints (or EndpointSlice) objects. When a pod calls a Service ClusterIP, iptables rules intercept the packet and DNAT (Destination NAT) it to one of the backing pod IPs before routing.

```
Pod A calls Service at 10.96.5.10:80
        ↓
Packet enters iptables PREROUTING chain
        ↓
Rule: "-d 10.96.5.10 -p tcp --dport 80 → jump KUBE-SVC-XXXXX"
        ↓
KUBE-SVC-XXXXX chain (load balancing):
  33% chance → DNAT to 10.244.1.5:8080 (Pod 1)
  33% chance → DNAT to 10.244.2.3:8080 (Pod 2)
  33% chance → DNAT to 10.244.3.7:8080 (Pod 3)
        ↓
Packet is routed to the selected pod IP
(same-node or cross-node path from Scenario 2 or 3)
```

**Return traffic:** the connection tracking (`conntrack`) kernel module records the DNAT translation. When the pod responds, conntrack automatically un-DNATs the response (changes src IP from pod IP back to Service ClusterIP) before delivering it to the caller. The caller never sees the pod's IP — it only ever sees the Service ClusterIP.

#### kube-proxy IPVS mode

IPVS (IP Virtual Server) is a more scalable alternative to iptables:
- IPVS uses hash tables instead of linear iptables rule chains — O(1) lookup vs O(n)
- Supports more load balancing algorithms: round-robin, least-connection, source-hash, etc.
- At large scale (10,000+ services), iptables performance degrades significantly; IPVS handles this efficiently
- kube-proxy in IPVS mode creates virtual servers in the kernel's IPVS module

#### DNS resolution for Services

```
Pod A queries: my-service.my-namespace.svc.cluster.local
        ↓
CoreDNS resolves → 10.96.5.10 (ClusterIP)
        ↓
Pod A connects to 10.96.5.10:80
        ↓
iptables/IPVS intercepts and DNAT's to a pod IP
```

CoreDNS (running as a deployment in `kube-system`) watches Service objects and automatically creates DNS A records for every Service. The kubelet configures each pod's `/etc/resolv.conf` to point to CoreDNS's ClusterIP.

---

### Interview Questions — Scenario 4

#### Beginner

**Q1. What is a Kubernetes Service and why is it needed?**

A Service provides a stable, permanent IP address and DNS name for a set of pods. Pods are ephemeral — when they crash and restart, they get a new IP address. Without Services, other pods would need to track the constantly changing pod IPs, which is impractical. A Service solves this by providing a single ClusterIP that never changes and always routes traffic to healthy pods. kube-proxy on each node maintains iptables/IPVS rules that intercept traffic destined for the Service IP and forward it to an actual pod IP.

---

**Q2. How does a pod discover the IP of a Service it needs to call?**

Kubernetes automatically injects environment variables into every pod for Services that existed before the pod started (e.g., `MY_SERVICE_SERVICE_HOST`, `MY_SERVICE_SERVICE_PORT`). More commonly, pods use DNS — Kubernetes runs CoreDNS as a cluster DNS service, and every pod's `/etc/resolv.conf` is configured to use CoreDNS. A pod can resolve `my-service.my-namespace.svc.cluster.local` to the Service's ClusterIP. Within the same namespace, even the short form `my-service` resolves correctly via the search domains in `/etc/resolv.conf`.

---

#### Intermediate

**Q3. Explain how iptables implements Service load balancing. What is DNAT and conntrack?**

When kube-proxy creates iptables rules for a Service with three pods, it creates a chain with three rules — each with a probability that causes roughly equal distribution (33% chance to DNAT to pod 1, 50% of remaining to pod 2, 100% of remaining to pod 3 — which works out to 33/33/33 distribution).

**DNAT (Destination NAT)** rewrites the destination IP of the packet from the Service ClusterIP to a specific pod IP before the packet is routed. The packet is then routed normally to that pod's IP.

**conntrack (connection tracking)** is the Linux kernel module that tracks stateful connections. When a DNAT is applied to the first packet of a connection, conntrack records `(Service IP:port → Pod IP:port)` in its table. For all subsequent packets in that connection (responses, follow-up requests), conntrack applies the reverse translation automatically — ensuring that response packets from the pod are un-DNATed back to the Service IP before delivery to the caller, maintaining the illusion that the caller was always talking to the Service IP.

---

**Q4. What is the difference between Endpoints and EndpointSlices?**

**Endpoints** is the original Kubernetes object that stores the list of pod IPs backing a Service. For a Service selecting 100 pods, there is one Endpoints object containing all 100 IP:port pairs. Every change to any pod (restart, new pod, pod failure) triggers an update to the entire Endpoints object and a notification to all kube-proxy instances on all nodes — even if only one pod changed.

**EndpointSlices** (introduced in Kubernetes 1.17, GA in 1.21) split the endpoint information into smaller slices (default 100 endpoints per slice). When one pod changes, only the slice containing that pod is updated and distributed — not all endpoints. At scale (thousands of pods behind a service), this reduces network traffic to the API server and kube-proxy update latency significantly. EndpointSlices are now the default in modern Kubernetes clusters.

---

#### Expert

**Q5. Explain the performance difference between kube-proxy iptables mode and IPVS mode at scale. When does IPVS become necessary?**

**iptables mode:**
- Service rules are implemented as a linear chain of `PREROUTING` iptables rules
- For N services, kube-proxy creates O(N) rules in the `KUBE-SERVICES` chain
- Every packet destined for a service must traverse the chain until a matching rule is found — O(N) per packet in the worst case
- For 10,000 services with 5 pods each, that is 50,000+ iptables rules; packet processing latency increases measurably
- Updating rules requires flushing and rewriting all rules atomically — at 10,000 services, a single pod restart causes a full iptables rewrite which can take seconds and cause packet drops

**IPVS mode:**
- Uses the kernel's IPVS module (L4 load balancer built into Linux)
- IPVS uses hash tables — O(1) lookup regardless of number of services
- Supports native load balancing algorithms: rr (round-robin), lc (least connection), dh (destination hash), sh (source hash), lblc (locality-based least connection)
- Updates are incremental — adding/removing a single endpoint does not require a full rewrite
- At 10,000 services, packet processing latency is constant (O(1)) vs O(N) for iptables

**When IPVS is necessary:** generally recommended for clusters with >1,000 services or when service update latency during deployments is impacting traffic. Companies like WeChat and large Alibaba deployments reported iptables becoming unusable beyond 10,000+ services — IPVS solved this.

---

## 5. Communication between external clients and Services

### Detailed Explanation

By default, Services with `ClusterIP` type are only reachable from within the cluster. To expose services to external clients, Kubernetes provides several mechanisms.

#### NodePort

Opens a static port (default range: 30000-32767) on every node's IP. External traffic to `<NodeIP>:<NodePort>` is forwarded to the Service, which load-balances to pod IPs.

```
External Client → NodeIP:30080
                      ↓
             Node's iptables (kube-proxy)
             DNAT to Service ClusterIP:80
                      ↓
             DNAT to selected Pod IP:8080
```

**Limitation:** requires knowing node IPs; no built-in HA if the targeted node fails; port range limited to 30000-32767.

#### LoadBalancer

Provisions an external load balancer from the cloud provider (AWS ELB, GCP LB, Azure LB). The load balancer gets a public IP and forwards traffic to NodePorts on the cluster nodes.

```
External Client → Public IP:80 (Cloud LB)
                       ↓
              Any Node's eth0:NodePort
                       ↓
               kube-proxy → Pod IP
```

#### Ingress

An Ingress is a Layer 7 (HTTP/HTTPS) gateway that routes external requests based on URL path and hostname to different Services — all through a single external IP/LoadBalancer.

```
External: https://api.roku.com/v1/stream
                      ↓
              Ingress Controller (nginx/traefik/ALB)
              Rules:
               /v1/stream → streaming-service:80
               /v1/auth   → auth-service:80
               /v1/user   → user-service:80
                      ↓
              Service → Pod
```

**TLS termination:** the Ingress controller handles TLS decryption at the edge — backend pods communicate in plain HTTP within the cluster.

#### ExternalName

Maps a Service name to an external DNS name — the DNS returns a CNAME. Used to bring external services into the Kubernetes service discovery model.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: external-db
spec:
  type: ExternalName
  externalName: my-database.example.com
```

---

### Interview Questions — Scenario 5

#### Beginner

**Q1. How do you expose a Kubernetes pod to traffic from outside the cluster?**

There are three main ways. **NodePort** assigns a port (30000-32767) on every node's IP — external clients connect to any node's IP at that port. **LoadBalancer** provisions a cloud load balancer with a public IP that forwards to the cluster — this is the standard way to expose production services in cloud environments. **Ingress** is a Layer 7 HTTP/HTTPS gateway that routes multiple services through a single external IP based on URL paths and hostnames — it is the most flexible and cost-effective for exposing many HTTP services because one Ingress controller (and one load balancer) can serve all of them.

---

**Q2. What is the difference between NodePort and LoadBalancer service types?**

**NodePort** opens a static port on every node and routes external traffic hitting `<any-node-IP>:<node-port>` to the Service. It requires no cloud infrastructure but clients must know a node's IP, and if that node fails, the client must try another node — there is no automatic failover or HA.

**LoadBalancer** builds on NodePort by additionally provisioning a cloud load balancer (AWS ELB, GCP LB) with a stable, externally accessible IP. The cloud load balancer health-checks the nodes and only sends traffic to healthy ones — providing HA and automatic failover. It costs money (cloud LB pricing) but is the production-grade solution for externally accessible services.

---

#### Intermediate

**Q3. Explain how an Ingress controller works. What is the difference between an Ingress resource and an Ingress controller?**

An **Ingress resource** is a Kubernetes API object that defines routing rules — "route traffic for `api.roku.com/stream` to the streaming-service on port 80." It is just configuration — it does nothing by itself.

An **Ingress controller** is the actual software that reads Ingress resources and implements the routing. It is a pod (typically running nginx, Traefik, HAProxy, or a cloud-native solution like AWS ALB Ingress Controller) that watches Ingress objects and configures itself accordingly. The Ingress controller is exposed externally via a LoadBalancer service, giving it a public IP. All external HTTP/HTTPS traffic enters through this single IP, and the controller routes requests to the appropriate backend services based on the Ingress rules.

**Why Ingress over multiple LoadBalancers:** each LoadBalancer service provisions a separate cloud load balancer and charges accordingly. One Ingress controller with one LoadBalancer can route traffic to hundreds of internal services — significantly cheaper and simpler to manage.

---

**Q4. What happens to the source IP of a request when it reaches a pod via a NodePort or LoadBalancer service?**

By default, when a request comes in via NodePort or LoadBalancer, kube-proxy performs **SNAT (Source NAT)** on the packet — it replaces the client's source IP with the node's IP. This is necessary so that the response packet returns to the node that received the request (which can then un-NAT it back to the client). The pod sees the node's IP as the source, not the client's original IP — this breaks applications that use the client IP for logging, access control, or rate limiting.

**Solution:** set `externalTrafficPolicy: Local` on the Service. This tells kube-proxy to only route to pods on the same node where the traffic arrived — no cross-node hop needed — which means no SNAT is needed and the pod sees the real client IP. The trade-off: if no pods are running on the node that receives the traffic, the request fails — so this requires ensuring pods are spread across all nodes (or using a DaemonSet).

---

#### Expert

**Q5. How does the AWS Load Balancer Controller differ from the classic in-tree AWS cloud provider for Ingress? When would you use ALB vs NLB backed Ingress?**

**Classic in-tree cloud provider (legacy):**
- Creates a Classic ELB (Layer 4/7) for each LoadBalancer service
- Ingress used to be handled by third-party controllers (nginx-ingress with a Classic ELB in front)
- Limited AWS integration — no WAF, no native target group health checks, no pod-level routing

**AWS Load Balancer Controller (modern, out-of-tree):**
- Manages AWS ALBs (Application LB, Layer 7) for Ingress resources
- Manages AWS NLBs (Network LB, Layer 4) for LoadBalancer services
- Implements **IP target mode**: registers pod IPs directly as ALB targets instead of NodePort — traffic goes straight to the pod without a NodePort hop, preserving client IPs natively, reducing latency, and removing the kube-proxy iptables hop for external traffic
- Integrates with AWS WAF, Shield, ACM (certificate management), and Cognito

**ALB (Ingress) vs NLB (Service) choice:**
- **ALB:** HTTP/HTTPS routing, host/path-based rules, WebSocket, gRPC (via HTTP/2), WAF integration — use for user-facing APIs and web traffic
- **NLB:** ultra-low latency Layer 4 TCP/UDP, static IPs, TLS passthrough — use for non-HTTP protocols, gaming, real-time systems, or when preserving TCP-level connection handling is critical

---

## 6. Communication between Pods and external systems

### Detailed Explanation

Pods often need to reach services outside the Kubernetes cluster — external databases, third-party APIs, SaaS services. This outbound traffic path is distinct from inbound traffic.

#### Outbound traffic from a pod to the internet

```
Pod A (10.244.1.2) → external API (1.2.3.4:443)

1. Pod sends packet: src=10.244.1.2, dst=1.2.3.4
2. Packet routed via default gateway → node's bridge → node's eth0
3. At node's eth0: SNAT applied (iptables MASQUERADE rule)
   src IP changed from 10.244.1.2 → 192.168.1.10 (node's IP)
4. Packet leaves node to physical network/internet
5. External API responds to 192.168.1.10
6. Node receives response, conntrack un-SNATs: dst changed back to 10.244.1.2
7. Packet delivered to Pod A
```

**Why SNAT is needed outbound:** pod IPs are private, cluster-internal addresses not routable on the internet. The node's IP (or NAT gateway IP in cloud environments) is the only routable address.

#### ExternalName Services — bringing external services into cluster DNS

```yaml
apiVersion: v1
kind: Service
metadata:
  name: external-postgres
  namespace: production
spec:
  type: ExternalName
  externalName: prod-db.us-east-1.rds.amazonaws.com
```

A pod can now connect to `external-postgres.production.svc.cluster.local` and Kubernetes DNS resolves it to `prod-db.us-east-1.rds.amazonaws.com` via a CNAME record. This allows changing the external endpoint without updating application configuration.

#### Network Policies for egress control

By default, pods can reach any external IP. NetworkPolicy egress rules can restrict which external destinations a pod can reach:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: restrict-egress
spec:
  podSelector:
    matchLabels:
      app: payment-service
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 10.0.0.0/8      # allow internal
    ports:
    - protocol: TCP
      port: 5432              # only to internal PostgreSQL
  - to:
    - ipBlock:
        cidr: 52.94.0.0/16   # AWS payment API
    ports:
    - protocol: TCP
      port: 443
```

#### Cloud NAT Gateway

In cloud environments, pod outbound internet traffic typically goes through a NAT Gateway (AWS NAT GW, GCP Cloud NAT) rather than being SNATed at the node level. This provides a stable outbound IP for firewall whitelisting and scales better than per-node SNAT.

```
Pod → Node → NAT Gateway (static public IP) → Internet
```

---

### Interview Questions — Scenario 6

#### Beginner

**Q1. How does a pod connect to a database running outside the Kubernetes cluster?**

A pod can connect to an external database just like any regular application — by using the database's hostname or IP and port. The pod's traffic exits through the node's network interface (after SNAT to replace the pod's internal IP with the node's IP) and reaches the external database. To make this cleaner within Kubernetes, you can create an `ExternalName` Service that maps a Kubernetes DNS name to the external database's hostname — for example, `db.myapp.svc.cluster.local` resolves to `prod.db.example.com`. This allows you to change the external database endpoint by updating the Service without changing application configuration.

---

**Q2. Why does outbound traffic from a pod need SNAT?**

Pod IP addresses are private, cluster-internal addresses — they are not routable on the public internet or even on the physical corporate network outside the cluster. When a pod sends a packet to an external server, that packet's source IP is the pod's internal IP (e.g., `10.244.1.5`). The external server cannot send a response back to this IP because it has no route to it. SNAT (Source NAT) replaces the pod's source IP with the node's real IP (or a NAT gateway IP) so that the external server can respond to a routable address. The node's conntrack table records the translation so it can reverse it when the response arrives, delivering it correctly to the pod.

---

#### Intermediate

**Q3. How do you restrict which external services a pod can communicate with using NetworkPolicy?**

NetworkPolicy `egress` rules define what outbound traffic is allowed. By default, if no NetworkPolicy selects a pod, all egress is allowed. Once any NetworkPolicy selects a pod and includes `policyTypes: [Egress]`, all egress is denied except what is explicitly permitted.

To allow a payment service pod to only reach an internal PostgreSQL database and an external payment processor API:

```yaml
egress:
- to:
  - podSelector:
      matchLabels:
        app: postgres
  ports:
  - protocol: TCP
    port: 5432
- to:
  - ipBlock:
      cidr: 52.94.100.0/24   # payment processor IP range
  ports:
  - protocol: TCP
    port: 443
```

Note that you also need an egress rule allowing DNS (UDP port 53 to CoreDNS) — a common mistake is forgetting this, which causes DNS resolution to fail even though the NetworkPolicy intends to allow connections.

---

**Q4. A pod needs to call an external HTTPS API. The company firewall requires all traffic to come from a specific static IP. How do you implement this in Kubernetes on AWS?**

Use an **AWS NAT Gateway** with an Elastic IP. Route all outbound pod traffic from the cluster's VPC through the NAT Gateway:
1. In the VPC routing table for the subnet containing the cluster nodes, add a route: `0.0.0.0/0 → NAT Gateway`
2. The NAT Gateway has a fixed Elastic IP
3. All outbound traffic from any pod on any node will SNAT to the NAT Gateway's Elastic IP before leaving the VPC
4. Whitelist this Elastic IP in the company firewall

This works because: Pod IP → SNAT to Node IP (by kube-proxy masquerade) → SNAT to NAT Gateway Elastic IP (by AWS VPC routing). The external API always sees the NAT Gateway's IP. If you need different groups of pods to use different outbound IPs (for different external APIs with different IP whitelists), you can use **AWS VPC CNI with security groups per pod** and multiple NAT Gateways in different subnets with routing based on pod security groups.

---

#### Expert

**Q5. Explain how a service mesh (Istio) handles egress traffic control, and how it differs from NetworkPolicy for external communication control.**

**NetworkPolicy (kernel-level, L3/L4):**
- Enforced by the CNI plugin using iptables/eBPF rules in the kernel
- Controls traffic based on IP addresses, CIDR blocks, ports, and pod label selectors
- Cannot inspect HTTP headers, hostnames, or TLS SNI — only raw IP and port
- A NetworkPolicy allowing `52.94.0.0/16:443` allows ALL HTTPS to that IP range — you cannot further restrict to specific hostnames or API paths
- Zero overhead for allowed traffic — kernel enforces policy, no user-space proxy

**Istio Egress Control (application-level, L7):**
- Istio's Envoy sidecar intercepts all outbound traffic from the pod (via iptables rules in the pod's network namespace)
- `ServiceEntry` objects define external services that are allowed:
  ```yaml
  apiVersion: networking.istio.io/v1alpha3
  kind: ServiceEntry
  metadata:
    name: payment-api
  spec:
    hosts:
    - api.stripe.com
    ports:
    - number: 443
      protocol: HTTPS
    resolution: DNS
    location: MESH_EXTERNAL
  ```
- Can enforce: mTLS to external services, circuit breaking, retry policies, timeouts, and header-level routing rules for external HTTP APIs
- `EgressGateway`: routes all outbound traffic through a dedicated egress gateway pod — centralises egress, enables auditing, and allows a single static IP for all egress (same result as NAT Gateway, but at L7 with policy enforcement)

**When to use which:**
- NetworkPolicy for coarse-grained IP-level egress restrictions (compliance requirement: this pod must not reach the internet at all)
- Istio/service mesh for fine-grained hostname and protocol-aware egress control with observability (which pods are calling which external APIs, with what latency and error rates)
- In practice, use both — NetworkPolicy as the outer security boundary, service mesh for observability and L7 policy

---

## Quick Reference Summary

| Scenario | Mechanism | NAT | Key Components |
|---|---|---|---|
| Same-pod containers | Shared network namespace (localhost) | None | Pause container, shared namespace |
| Same-node pods | veth pairs + Linux bridge | None | CNI plugin, cni0 bridge, ARP |
| Cross-node pods | Overlay (VXLAN) or Underlay (BGP) | None | CNI plugin, tunnels or BGP routes |
| Pod → Service | iptables DNAT / IPVS | DNAT (Service→Pod) | kube-proxy, conntrack, CoreDNS |
| External → Service | NodePort / LoadBalancer / Ingress | SNAT (optional, external traffic) | Cloud LB, Ingress controller, kube-proxy |
| Pod → External | SNAT at node / NAT Gateway | SNAT (pod IP→node IP) | conntrack, NAT Gateway, NetworkPolicy |







what is network namespace
what is /proc/net, iptables, eth0, lo, IP address   
explain this , If `shareProcessNamespace: true` is set in the pod spec, containers share the same PID namespace and can send signals to each other's processes.
what is `eth0` interface


















# Kubernetes StatefulSets
### Everything You Need to Know + Interview Q&A (Beginner → Expert)

---

## Prerequisites — What You Need to Understand First

Before StatefulSets make sense, you need to be comfortable with these concepts:

| Concept | Why it matters for StatefulSets |
|---|---|
| **Pods** | StatefulSets manage pods with identity |
| **Deployments** | StatefulSets are the stateful alternative |
| **Persistent Volumes (PV)** | StatefulSets create PVs per pod |
| **Persistent Volume Claims (PVC)** | Each pod gets its own PVC |
| **Storage Classes** | Used for dynamic PVC provisioning |
| **Services** | StatefulSets require a Headless Service |
| **DNS in Kubernetes** | StatefulSets give each pod a stable DNS name |

---

## Part 1 — The Problem StatefulSets Solve

### Why Deployments Are Not Enough for Stateful Apps

A Deployment creates pods that are **identical and interchangeable**. Any pod can be killed and replaced with another — the new pod gets a new random name, a new IP, and starts fresh with no memory of the old pod. This is perfect for stateless apps (web servers, APIs) but catastrophic for stateful apps.

Consider running a **PostgreSQL cluster** with one primary and two replicas:

```
PRIMARY (postgres-primary) → writes go here
REPLICA 1 (postgres-replica-1) → replicates from primary
REPLICA 2 (postgres-replica-2) → replicates from primary
```

**Problems with a Deployment for this:**

```
Problem 1 — No stable identity:
  Deployment pod names: postgres-7f9d4-xk2p1, postgres-7f9d4-ab3q2
  After restart:        postgres-7f9d4-zzz99, postgres-7f9d4-yyy88
  → Replicas cannot find the primary by name — the name keeps changing

Problem 2 — No stable storage per pod:
  If pod postgres-7f9d4-xk2p1 (the primary) gets killed
  and a new pod postgres-7f9d4-zzz99 starts → it gets no storage
  → All primary data is lost

Problem 3 — No ordered startup:
  Deployment starts all pods simultaneously
  → Replica 1 starts before Primary is ready → replication fails

Problem 4 — No stable network identity:
  Replica cannot hardcode "connect to primary at postgres-0" because
  the primary's pod name/IP changes on every restart
```

**StatefulSets solve ALL of these problems.**

---

## Part 2 — What StatefulSets Guarantee

StatefulSets provide four guarantees that Deployments do not:

### Guarantee 1 — Stable, Unique Pod Identity

```
StatefulSet name: postgres, replicas: 3

Pods created:
  postgres-0  (always the 0th pod)
  postgres-1  (always the 1st pod)
  postgres-2  (always the 2nd pod)

After postgres-1 crashes and restarts:
  → New pod is STILL named postgres-1
  → Never gets a random suffix
  → Always postgres-0, postgres-1, postgres-2
```

The ordinal index (0, 1, 2) is permanent. The pod always comes back with the same name.

### Guarantee 2 — Stable Network Identity (DNS)

When paired with a **Headless Service**, each pod gets a stable, predictable DNS entry:

```
Format: <pod-name>.<service-name>.<namespace>.svc.cluster.local

postgres-0.postgres-svc.default.svc.cluster.local → 10.244.1.2
postgres-1.postgres-svc.default.svc.cluster.local → 10.244.2.5
postgres-2.postgres-svc.default.svc.cluster.local → 10.244.3.8

After postgres-1 restarts (new IP: 10.244.2.99):
postgres-1.postgres-svc.default.svc.cluster.local → 10.244.2.99
```

The DNS name stays the same even though the IP changes. Other pods can hardcode
`postgres-0.postgres-svc` as the primary address — it always resolves to the current primary pod.

### Guarantee 3 — Stable, Persistent Storage (One PVC Per Pod)

StatefulSets use `volumeClaimTemplates` to create a dedicated PVC for each pod:

```
postgres-0 → PVC: data-postgres-0 → PV: 10GB SSD
postgres-1 → PVC: data-postgres-1 → PV: 10GB SSD
postgres-2 → PVC: data-postgres-2 → PV: 10GB SSD
```

**Critical behaviour:** when `postgres-1` is deleted and recreated, it automatically
reattaches to `data-postgres-1` — the same PVC, the same data. Data survives
pod restarts, rescheduling to different nodes, and even StatefulSet deletion
(PVCs are NOT deleted when a StatefulSet is deleted by default).

### Guarantee 4 — Ordered, Graceful Deployment and Scaling

```
Scale up (0 → 3 replicas):
  Start postgres-0 → wait until Running and Ready
  Start postgres-1 → wait until Running and Ready
  Start postgres-2 → wait until Running and Ready

Scale down (3 → 1 replica):
  Delete postgres-2 → wait until terminated
  Delete postgres-1 → wait until terminated
  (postgres-0 remains)

Rolling update:
  Update postgres-2 → wait until ready
  Update postgres-1 → wait until ready
  Update postgres-0 → wait until ready
  (reverse order — highest ordinal first)
```

This ordering allows the primary (always `postgres-0`) to be the last pod updated/deleted,
maintaining cluster availability throughout the operation.

---

## Part 3 — The Headless Service

A StatefulSet **requires** a Headless Service. This is one of the most commonly misunderstood aspects.

### What is a Headless Service?

A normal Service has a ClusterIP — a virtual IP that load-balances across pods.
A Headless Service has `clusterIP: None` — no virtual IP, no load balancing.

Instead, DNS for a Headless Service returns the **individual IP of each pod**,
and gives each pod a stable DNS A record.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-svc
spec:
  clusterIP: None         # ← this makes it headless
  selector:
    app: postgres
  ports:
  - port: 5432
```

### Why StatefulSets Need a Headless Service

```
Regular Service DNS:
  postgres-svc → 10.96.5.10 (ClusterIP, load-balanced)
  You cannot target a specific pod — any pod may respond

Headless Service DNS:
  postgres-svc → [10.244.1.2, 10.244.2.5, 10.244.3.8] (all pod IPs)
  postgres-0.postgres-svc → 10.244.1.2 (specific pod)
  postgres-1.postgres-svc → 10.244.2.5 (specific pod)
  postgres-2.postgres-svc → 10.244.3.8 (specific pod)

For a database cluster:
  Primary: always postgres-0.postgres-svc
  Replica setup script: "replicate from postgres-0.postgres-svc"
  → This works even after postgres-0 restarts with a new IP
```

---

## Part 4 — Complete StatefulSet YAML

```yaml
# Step 1: Headless Service (required — must exist before StatefulSet)
apiVersion: v1
kind: Service
metadata:
  name: postgres-svc
  namespace: default
spec:
  clusterIP: None           # headless
  selector:
    app: postgres
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432

---
# Step 2: StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: default
spec:
  serviceName: "postgres-svc"    # must match the Headless Service name
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
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: data          # must match volumeClaimTemplates name
          mountPath: /var/lib/postgresql/data
        readinessProbe:
          exec:
            command: ["pg_isready", "-U", "postgres"]
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"

  # One PVC per pod, named: data-postgres-0, data-postgres-1, data-postgres-2
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: "fast-ssd"
      resources:
        requests:
          storage: 10Gi

  # Controls pod management behaviour
  podManagementPolicy: OrderedReady    # or Parallel

  # Controls update strategy
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      partition: 0    # update all pods (partition=N means only update pods >= N)
```

---

## Part 5 — Key Configuration Options Explained

### podManagementPolicy

```
OrderedReady (default):
  - Pods created/deleted one at a time in order
  - Each must be Ready before the next starts
  - Safe for databases where startup order matters

Parallel:
  - All pods created/deleted simultaneously
  - Still maintains stable identity (names, storage)
  - Use when pods are independent (Kafka, Elasticsearch nodes that self-discover)
```

### updateStrategy

```
RollingUpdate (default):
  - Updates pods from highest ordinal to lowest (2 → 1 → 0)
  - partition: N → only update pods with ordinal >= N
  - partition: 2 means only postgres-2 is updated (canary deployment)
  - partition: 0 means all pods are updated (default)

OnDelete:
  - Pods are NOT automatically updated
  - You must manually delete a pod to trigger update
  - New pod starts with the new spec
  - Use when you need manual control over each update step
```

### Partition — Canary Updates for StatefulSets

```
StatefulSet with 3 pods, partition: 2
  postgres-0 → OLD version (not updated)
  postgres-1 → OLD version (not updated)
  postgres-2 → NEW version (updated, can be tested)

After testing postgres-2, set partition: 1
  postgres-0 → OLD version (not updated)
  postgres-1 → NEW version (updated)
  postgres-2 → NEW version (already updated)

After validation, set partition: 0
  postgres-0 → NEW version (updated)
  postgres-1 → NEW version (already updated)
  postgres-2 → NEW version (already updated)
```

---

## Part 6 — StatefulSet vs Deployment

| Feature | Deployment | StatefulSet |
|---|---|---|
| **Pod names** | Random suffix (`app-7f9d4-xk2p1`) | Ordered index (`app-0`, `app-1`) |
| **Pod identity** | Interchangeable | Unique, stable |
| **Startup order** | All at once (default) | Ordered (0 → 1 → 2) |
| **Shutdown order** | Any order | Reverse order (2 → 1 → 0) |
| **Storage** | Shared PVC or no PVC | One dedicated PVC per pod |
| **Network identity** | Random pod DNS | Stable per-pod DNS |
| **Headless Service** | Not required | Required |
| **Rolling update order** | Random | Highest ordinal first |
| **Scaling** | Immediate | Ordered |
| **Use case** | Stateless apps (APIs, web servers) | Stateful apps (databases, queues) |

---

## Interview Questions — StatefulSets

---

## Beginner Level

---

### Q1. What is a StatefulSet and when would you use it instead of a Deployment?

A StatefulSet is a Kubernetes workload controller designed for applications that require stable, persistent identity across pod restarts. Unlike a Deployment — where pods are identical and interchangeable — a StatefulSet assigns each pod a persistent ordinal index (0, 1, 2), a stable DNS name, and a dedicated PersistentVolumeClaim that survives pod restarts.

Use a StatefulSet when your application:
- Requires stable network identity (each instance needs a unique, predictable hostname)
- Requires persistent storage that is private to each instance (not shared)
- Requires ordered startup (the primary must be ready before replicas start)
- Requires ordered shutdown (reverse of startup order)

**Common use cases:** databases (PostgreSQL, MySQL, MongoDB), message brokers (Kafka, RabbitMQ), distributed key-value stores (etcd, ZooKeeper), search engines (Elasticsearch).

---

### Q2. What are the four key guarantees that a StatefulSet provides?

1. **Stable, unique pod identity** — pods are named with a fixed ordinal index (`app-0`, `app-1`, `app-2`). After a pod restarts, it comes back with the same name — never a random suffix.

2. **Stable network identity** — when paired with a Headless Service, each pod gets a permanent DNS entry (`app-0.svc-name.namespace.svc.cluster.local`) that resolves to the current pod IP even after restarts.

3. **Stable persistent storage** — each pod gets its own dedicated PersistentVolumeClaim created from the `volumeClaimTemplates` section. The PVC is bound to the pod by name (`data-app-0`, `data-app-1`) and reattaches to the same pod after restarts — data is never lost.

4. **Ordered graceful deployment and scaling** — pods are created in order (0 first, then 1, then 2) and each must be Running and Ready before the next starts. Scaling down removes pods in reverse order (2 first, then 1).

---

### Q3. What is a Headless Service and why does a StatefulSet require one?

A Headless Service is a Service with `clusterIP: None`. Unlike a regular Service, it does not get a virtual ClusterIP and does not load-balance traffic. Instead, DNS queries for a Headless Service return the individual IP addresses of each pod, and each pod in a StatefulSet gets its own stable DNS A record.

StatefulSets require a Headless Service because stateful applications need to address individual pods by name — not just reach any healthy pod. For a database cluster, other components need to reach `postgres-0` specifically (the primary), not a random pod. The Headless Service enables this by creating `postgres-0.postgres-svc.default.svc.cluster.local` as a permanent DNS entry that always resolves to whichever IP `postgres-0` currently has, even after restarts.

Without a Headless Service, there is no stable per-pod DNS — the StatefulSet's identity feature would be incomplete because pods could not reliably find each other by name.

---

### Q4. What happens to the PersistentVolumeClaim when a StatefulSet pod is deleted?

The PVC is **not deleted** when a pod is deleted. This is intentional and one of the most important behaviours of StatefulSets. When the pod is recreated (by the StatefulSet controller after a deletion), it automatically reattaches to the same PVC by name. For example, `postgres-1` always binds to PVC `data-postgres-1` — the controller enforces this name-based binding.

The PVCs are also not deleted when the StatefulSet itself is deleted — you must delete the PVCs manually. This protects against accidental data loss. If you want automatic PVC deletion when the StatefulSet is deleted, Kubernetes 1.27+ introduced the `persistentVolumeClaimRetentionPolicy` field which can be set to `Delete` for the `whenDeleted` or `whenScaled` scenarios.

---

### Q5. What is the difference between `OrderedReady` and `Parallel` pod management policies?

**OrderedReady (default):**
- Pods are created one at a time, in ascending ordinal order (0, then 1, then 2)
- Each pod must be Running and Ready before the next one is created
- Pods are deleted in reverse order (2, then 1, then 0) with each waiting to be fully terminated before the next is deleted
- Used when startup and shutdown order matters — databases where the primary must be healthy before replicas start

**Parallel:**
- All pods are created or deleted simultaneously
- Does not wait for each pod to be Ready before starting the next
- Still maintains stable identity — pods still get ordered names and dedicated PVCs
- Used when pods are independent and startup order does not matter — Elasticsearch nodes, Kafka brokers that self-discover the cluster

---

## Intermediate Level

---

### Q6. Explain how DNS works for StatefulSet pods. What DNS records are created?

When a StatefulSet is created with a Headless Service named `postgres-svc` in namespace `default`, CoreDNS automatically creates the following DNS records for each pod:

**A records (per pod):**
```
postgres-0.postgres-svc.default.svc.cluster.local → 10.244.1.2
postgres-1.postgres-svc.default.svc.cluster.local → 10.244.2.5
postgres-2.postgres-svc.default.svc.cluster.local → 10.244.3.8
```

**SRV records (for service discovery):**
```
_postgres._tcp.postgres-svc.default.svc.cluster.local
  → postgres-0.postgres-svc.default.svc.cluster.local:5432
  → postgres-1.postgres-svc.default.svc.cluster.local:5432
  → postgres-2.postgres-svc.default.svc.cluster.local:5432
```

**Service A record (all pod IPs):**
```
postgres-svc.default.svc.cluster.local → [10.244.1.2, 10.244.2.5, 10.244.3.8]
```

When `postgres-1` restarts and gets a new IP (`10.244.2.99`), CoreDNS automatically updates its A record. Other pods that had cached the old IP will get the new IP on the next DNS TTL expiry. The DNS name itself never changes — only the IP it resolves to changes on restart.

---

### Q7. How do you perform a rolling update on a StatefulSet, and what is the `partition` field used for?

A StatefulSet rolling update updates pods from the highest ordinal to the lowest (2 → 1 → 0). Each pod is updated, becomes Ready, and then the next is updated. This reverse order ensures the primary (ordinal 0) is updated last, keeping the cluster serving traffic throughout the update.

**The partition field enables canary deployments:**

`partition: N` means only pods with ordinal >= N are updated. Pods with ordinal < N keep the old version.

```bash
# Start: all pods on v1. Set partition=2 to test on postgres-2 only:
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":2}}}}'
kubectl set image statefulset/postgres postgres=postgres:17

# postgres-2 → updates to v17 (ordinal 2 >= partition 2)
# postgres-1 → stays on v16 (ordinal 1 < partition 2)
# postgres-0 → stays on v16 (ordinal 0 < partition 2)

# After testing postgres-2, proceed to include postgres-1:
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":1}}}}'
# postgres-1 → updates to v17

# After validating, complete the rollout:
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":0}}}}'
# postgres-0 → updates to v17
```

---

### Q8. What is `volumeClaimTemplates` and how is it different from using a regular `volumes` section?

**Regular `volumes` section (in pod spec):**
- References an existing PVC by name: `claimName: my-pvc`
- All pods in the StatefulSet share the same PVC
- Only valid for ReadWriteMany access mode (NFS, shared storage)
- Not recommended for databases — all pods read/write the same storage

**`volumeClaimTemplates` (StatefulSet-specific):**
- A template that Kubernetes uses to create one PVC per pod automatically
- PVC names are generated as `<template-name>-<pod-name>`: `data-postgres-0`, `data-postgres-1`
- Each pod mounts its own isolated PVC — no sharing
- The PVCs persist independently from the pods
- Access mode is typically `ReadWriteOnce` (one pod at a time) which is correct for databases

```yaml
# Regular volumes — all pods share one PVC (wrong for databases):
volumes:
- name: data
  persistentVolumeClaim:
    claimName: shared-pvc         # same PVC for all pods!

# volumeClaimTemplates — one PVC per pod (correct for databases):
volumeClaimTemplates:
- metadata:
    name: data
  spec:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 10Gi
# Creates: data-postgres-0, data-postgres-1, data-postgres-2
```

---

### Q9. How do you scale a StatefulSet? What happens to PVCs when you scale down?

**Scaling up:**
```bash
kubectl scale statefulset postgres --replicas=5
```
Pods are created in order: `postgres-3` starts and becomes Ready, then `postgres-4` starts. New PVCs `data-postgres-3` and `data-postgres-4` are created automatically.

**Scaling down:**
```bash
kubectl scale statefulset postgres --replicas=2
```
Pods are deleted in reverse order: `postgres-4` is deleted first (and terminated), then `postgres-3`. The pods wait to be fully terminated before the next is deleted.

**What happens to PVCs when scaling down:**
The PVCs `data-postgres-3` and `data-postgres-4` are **NOT deleted automatically**. They remain in the cluster in `Released` or `Bound` state. This is a safety feature — scaling back up to 5 replicas will reattach the same PVCs with the same data.

**To delete PVCs after scaling down, you must do it manually:**
```bash
kubectl delete pvc data-postgres-3 data-postgres-4
```

In Kubernetes 1.27+, you can configure `persistentVolumeClaimRetentionPolicy`:
```yaml
persistentVolumeClaimRetentionPolicy:
  whenScaled: Delete    # delete PVCs when scaling down
  whenDeleted: Retain   # keep PVCs when StatefulSet is deleted
```

---

### Q10. How does a StatefulSet handle pod failures? What is the "at most one" guarantee?

When a StatefulSet pod fails, the controller creates a replacement pod with the same name and reattaches it to the same PVC. However, the StatefulSet controller enforces an **"at most one" guarantee** — it ensures that at most one pod with a given identity is running at any time.

**Why "at most one" matters:**
If `postgres-0` appears to fail (node becomes unreachable), and the controller immediately creates a new `postgres-0` on another node, there could briefly be two `postgres-0` pods trying to write to the same PVC — causing data corruption. The `ReadWriteOnce` PVC access mode would actually prevent both from mounting simultaneously, but the risk of split-brain behaviour is real.

**How Kubernetes handles this:**
- Before creating a replacement pod, Kubernetes waits for the old pod to be definitively terminated
- If the node is unreachable (not explicitly deleted), the pod stays in `Terminating` state indefinitely — the controller does NOT create a replacement until the node is confirmed dead or the pod is force-deleted
- This means you must sometimes manually force-delete stuck pods:
  ```bash
  kubectl delete pod postgres-0 --force --grace-period=0
  ```
- Only after this force-delete will the controller create the replacement

This conservative behaviour prevents data corruption at the cost of requiring manual intervention when nodes fail.

---

## Advanced Level

---

### Q11. Explain how you would run a PostgreSQL primary-replica cluster using StatefulSets. What init containers and scripting would you need?

Running a proper PostgreSQL cluster on Kubernetes requires solving the chicken-and-egg problem: the replicas need to know they are replicas and must replicate from the primary, but all pods start with the same container image.

**The solution — use init containers and pod ordinal detection:**

```yaml
initContainers:
- name: init-postgres
  image: postgres:16
  command:
  - bash
  - -c
  - |
    # Detect if this pod is primary (ordinal 0) or replica (ordinal > 0)
    POD_ORDINAL=${POD_NAME##*-}  # extract number from postgres-0, postgres-1, etc.

    if [ "$POD_ORDINAL" = "0" ]; then
      echo "I am the primary (postgres-0)"
      # Initialize a fresh database if data dir is empty
      if [ -z "$(ls -A /var/lib/postgresql/data/pgdata)" ]; then
        initdb -D /var/lib/postgresql/data/pgdata
        # Configure for replication
        echo "host replication replicator all md5" >> /var/lib/postgresql/data/pgdata/pg_hba.conf
      fi
    else
      echo "I am a replica (ordinal $POD_ORDINAL), cloning from primary"
      # Wait for primary to be ready
      until pg_isready -h postgres-0.postgres-svc; do
        echo "Waiting for primary..."; sleep 2
      done
      # Clone primary's data
      pg_basebackup -h postgres-0.postgres-svc -D /var/lib/postgresql/data/pgdata \
        -U replicator --wal-method=stream --checkpoint=fast
      # Configure as replica
      cat > /var/lib/postgresql/data/pgdata/recovery.conf << EOF
      standby_mode = on
      primary_conninfo = 'host=postgres-0.postgres-svc port=5432 user=replicator'
      EOF
    fi
  env:
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  volumeMounts:
  - name: data
    mountPath: /var/lib/postgresql/data
```

**Key points:**
- `POD_NAME` environment variable gives the pod its own name (`postgres-0`, `postgres-1`)
- The init container extracts the ordinal from the pod name
- Ordinal 0 → initialise as primary
- Ordinal > 0 → clone from `postgres-0.postgres-svc` (stable DNS name of primary)
- `OrderedReady` policy ensures postgres-0 is fully Ready before postgres-1 starts its init container

---

### Q12. What is the `OnDelete` update strategy and when is it the right choice over `RollingUpdate`?

`OnDelete` means Kubernetes will NOT automatically update pods when the StatefulSet spec changes. Pods are only updated when you manually delete them — the new pod that starts will use the new spec.

**When to use OnDelete:**

1. **Manual step-by-step database upgrades:** for major version upgrades of databases (PostgreSQL 15 → 16), you want to upgrade and verify each node individually, running manual checks between each step. RollingUpdate would automate this and move too fast.

2. **When readiness probes cannot reliably detect all failure modes:** some database states (replication lag, long-running transactions) look "Ready" to Kubernetes but indicate the node should not yet be considered safe for cutover. Manual deletion gives the operator full control.

3. **Operations that require custom pre/post steps:** backing up a specific replica before upgrading it, reconfiguring replication topology between updates.

```yaml
updateStrategy:
  type: OnDelete

# Workflow:
# 1. Update the StatefulSet image:
kubectl set image statefulset/postgres postgres=postgres:17

# 2. Manually delete postgres-2 to trigger its update:
kubectl delete pod postgres-2
# Wait for postgres-2 to come back on v17, verify replication...

# 3. Manually delete postgres-1:
kubectl delete pod postgres-1
# Verify...

# 4. Manually delete postgres-0:
kubectl delete pod postgres-0
```

---

### Q13. How do you handle a situation where a StatefulSet pod is stuck in `Terminating` state and the node it was on is unreachable?

This is a common production scenario with StatefulSets that requires careful handling to avoid data corruption.

**Why it happens:**
The node running `postgres-1` loses network connectivity. Kubernetes marks the node as `NotReady`. The pod status shows `Terminating` but the kubelet on the unreachable node cannot execute the termination — so the pod never actually terminates. The StatefulSet controller refuses to create a replacement because it would violate the "at most one" guarantee.

**What NOT to do first:**
Do not immediately force-delete the pod without verifying the node is truly offline. If the node is actually running but just temporarily disconnected from the control plane, force-deleting the pod while the old pod is still running would briefly create two pods with the same identity writing to storage — potential data corruption.

**Safe procedure:**

```bash
# Step 1: Verify the node is genuinely dead (not just partitioned)
# Check cloud provider console, hypervisor, or physical access
# Confirm the VM/node is terminated/powered off

# Step 2: If confirmed dead, force-delete the stuck pod
kubectl delete pod postgres-1 --force --grace-period=0

# Step 3: StatefulSet controller now creates replacement postgres-1
# on a healthy node, reattaches to data-postgres-1 PVC

# Step 4: If the node comes back online later, kubelet will see
# the force-deleted pod is gone and clean up — no duplicate will run
```

**Using Pod Disruption Budgets (PDB) to prevent this at cluster level:**
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: postgres-pdb
spec:
  minAvailable: 2   # at least 2 postgres pods must be available
  selector:
    matchLabels:
      app: postgres
```
A PDB prevents voluntary disruptions (node drain, cluster upgrade) from taking down too many pods at once, reducing the frequency of this scenario.

---

### Q14. How do you implement a zero-downtime migration from a Deployment to a StatefulSet for a running application?

Migrating from Deployment to StatefulSet for a database that has been running is a delicate operation requiring careful planning.

**The challenge:**
- The Deployment's pods have no stable names or dedicated PVCs
- Existing data may be in shared storage or local node storage
- The application must remain available during migration

**Migration strategy:**

**Phase 1 — Preparation:**
```bash
# 1. Snapshot all existing PVCs/data for backup
# 2. Create the new StatefulSet manifests but do NOT apply yet
# 3. Create the Headless Service
kubectl apply -f headless-service.yaml

# 4. Scale the Deployment to 0 (maintenance window starts)
kubectl scale deployment my-postgres --replicas=0
```

**Phase 2 — Data migration:**
```bash
# 5. Create the StatefulSet with replicas=0 to create PVCs
kubectl apply -f statefulset.yaml
kubectl scale statefulset my-postgres --replicas=0

# 6. Copy data from old PVC to new StatefulSet PVCs using a migration pod
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: data-migrator
spec:
  volumes:
  - name: old-data
    persistentVolumeClaim:
      claimName: old-postgres-pvc
  - name: new-data
    persistentVolumeClaim:
      claimName: data-my-postgres-0
  containers:
  - name: migrator
    image: busybox
    command: ["sh", "-c", "cp -av /old/. /new/ && echo Done"]
    volumeMounts:
    - name: old-data
      mountPath: /old
    - name: new-data
      mountPath: /new
EOF
```

**Phase 3 — Cutover:**
```bash
# 7. Scale StatefulSet to 1 (primary only initially)
kubectl scale statefulset my-postgres --replicas=1

# 8. Verify data integrity
kubectl exec my-postgres-0 -- psql -U postgres -c "SELECT count(*) FROM important_table"

# 9. Update application Services to point to StatefulSet pods
kubectl patch service app-service -p '{"spec":{"selector":{"app":"my-postgres"}}}'

# 10. Scale to full replica count
kubectl scale statefulset my-postgres --replicas=3

# 11. Delete old Deployment and old PVC after validation period
kubectl delete deployment my-postgres
```

---

### Q15. Explain the `persistentVolumeClaimRetentionPolicy` introduced in Kubernetes 1.27 and why it matters operationally.

Before Kubernetes 1.27, PVCs created by StatefulSet `volumeClaimTemplates` were never automatically deleted — they required manual cleanup after scaling down or deleting the StatefulSet. This was a safety feature but created operational burden and storage cost accumulation.

**The new field:**
```yaml
spec:
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain    # or Delete
    whenScaled:  Retain    # or Delete
```

**`whenDeleted`:** controls what happens to PVCs when the entire StatefulSet is deleted:
- `Retain` (default) — PVCs remain; you must delete them manually
- `Delete` — PVCs are automatically deleted when the StatefulSet is deleted

**`whenScaled`:** controls what happens to PVCs for pods removed during scale-down:
- `Retain` (default) — PVCs remain; reused if you scale back up
- `Delete` — PVCs are deleted when their pod is removed during scale-down

**Operational guidance:**

```yaml
# For a test/development StatefulSet — clean up everything automatically:
persistentVolumeClaimRetentionPolicy:
  whenDeleted: Delete
  whenScaled:  Delete

# For production database — protect data at all costs:
persistentVolumeClaimRetentionPolicy:
  whenDeleted: Retain    # never auto-delete production data
  whenScaled:  Retain    # keep data if we need to scale back up

# For a log storage StatefulSet — clean up scaled-down pods, keep on delete:
persistentVolumeClaimRetentionPolicy:
  whenDeleted: Retain
  whenScaled:  Delete    # freed logs storage can be reclaimed
```

This feature significantly reduces the operational burden of managing StatefulSet storage, especially for teams that frequently create and destroy test environments.

---

## Expert Level

---

### Q16. How does Kubernetes prevent split-brain in StatefulSets, and what are its limitations compared to what databases need?

**What Kubernetes does to prevent split-brain:**

Kubernetes uses the "at most one" guarantee — enforced through the StatefulSet controller's conservative pod replacement policy. The controller will not create `postgres-1` on a new node while the old `postgres-1` exists in the API server, even if the node is unreachable. Combined with `ReadWriteOnce` PVC access mode (which prevents two pods from mounting the same storage simultaneously), this provides a basic level of split-brain prevention.

**Limitations — what Kubernetes cannot prevent:**

1. **Force-delete bypasses the guarantee:** `kubectl delete pod --force --grace-period=0` overrides the "at most one" guarantee. After force-deletion, the old pod may still be running on an isolated node for some time before it terminates. During this window, two pods with the same identity can both be "running" — one on the isolated node (old) and one on a healthy node (new).

2. **Storage-level protection is incomplete:** `ReadWriteOnce` prevents two pods from mounting the same PVC at the OS level, but does not prevent network-level split-brain (two nodes in a database cluster both believing they are primary and accepting writes).

3. **Kubernetes has no knowledge of database state:** it knows a pod is "Ready" (healthcheck passes) but cannot determine if it is in a state where it is safe to promote it as primary.

**What databases need beyond Kubernetes:**

Real split-brain prevention requires application-level solutions:
- **Fencing tokens:** each primary gets a lease with a monotonically increasing token; writes to storage are rejected if the token is outdated (etcd leases, Patroni's DCS-based fencing)
- **STONITH (Shoot The Other Node In The Head):** automated power-cycling of the old primary before the new primary takes over (used in on-premise HA clusters)
- **Patroni/Zalando Postgres Operator:** uses etcd or ZooKeeper as a distributed consensus system; only one node holds the leader key at a time; a node without the leader key cannot accept writes

**The correct answer for production PostgreSQL on Kubernetes:** use StatefulSets for pod identity and storage management, but deploy Patroni (or use the Zalando Postgres Operator / Crunchy Data PGO) on top for actual primary election and split-brain prevention. Kubernetes StatefulSets are necessary but not sufficient for a truly HA database.

---

### Q17. Design a Kafka cluster on Kubernetes using StatefulSets. What specific Kafka characteristics map to StatefulSet features, and what are the operational challenges?

**Kafka characteristics that map to StatefulSet features:**

```
Kafka broker ID ↔ StatefulSet pod ordinal
  kafka-0 → broker.id=0 (set via POD_NAME env var parsing)
  kafka-1 → broker.id=1
  kafka-2 → broker.id=2

Kafka log directories ↔ volumeClaimTemplates
  Each broker needs dedicated storage for its partition data
  data-kafka-0 → /var/kafka/logs for broker 0
  data-kafka-1 → /var/kafka/logs for broker 1

Kafka inter-broker communication ↔ Headless Service DNS
  kafka-0.kafka-svc:9092 (stable address for controller election)
  kafka-1.kafka-svc:9092
  kafka-2.kafka-svc:9092

ZooKeeper-less Kafka (KRaft) ↔ StatefulSet ordered startup
  Controller nodes must start before broker-only nodes
  Use partition field to control update order

Kafka startup order ↔ OrderedReady or Parallel
  Parallel is appropriate because Kafka brokers self-discover the cluster
  via ZooKeeper/KRaft — startup order is not strict
```

**StatefulSet configuration for Kafka:**

```yaml
spec:
  podManagementPolicy: Parallel    # brokers can start simultaneously
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      partition: 2    # start with the last broker for safe rolling upgrades

  # Multiple PVCs per pod — log dirs on fast storage, data dirs on large storage
  volumeClaimTemplates:
  - metadata:
      name: kafka-logs
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 100Gi
```

**Operational challenges:**

1. **Partition reassignment during pod deletion:** when kafka-1 is deleted for an update, its partition leaders must be migrated to other brokers first. Kubernetes does not know about Kafka leadership — you need a pre-stop hook or operator logic to trigger preferred replica election before termination.

2. **Broker decommissioning vs failure:** scaling down Kafka (removing a broker) requires first reassigning all partitions away from the broker being removed. Kubernetes just deletes the pod without this coordination. The Strimzi Kafka Operator handles this by wrapping StatefulSet operations with Kafka-aware pre-deletion steps.

3. **Storage reclaim:** deleted broker PVCs contain partition data that may be needed for recovery. The `whenScaled: Retain` policy is essential — do not use `Delete` for Kafka storage.

4. **External client access:** Kafka clients need to connect directly to specific brokers (the partition leader). A single LoadBalancer Service is insufficient — you need per-broker Services (one LoadBalancer or NodePort per broker pod) or an Ingress with per-broker routing. Strimzi handles this via per-pod listener Services.

**The recommendation:** use the Strimzi Kafka Operator rather than managing Kafka StatefulSets directly. Strimzi wraps StatefulSets with Kafka-specific operational logic for upgrades, scaling, topic management, and client authentication.

---

### Q18. Explain how StatefulSet rolling updates interact with Kubernetes PodDisruptionBudgets (PDB). How do you design PDBs for a three-node database cluster?

**The interaction:**

During a StatefulSet rolling update, each pod is deleted and recreated. A PodDisruptionBudget limits how many pods can be simultaneously unavailable. The StatefulSet controller respects PDBs — if deleting a pod would violate the PDB, the update pauses until the constraint is satisfied.

**Designing a PDB for a three-node PostgreSQL cluster:**

```yaml
# Option 1: minAvailable — at least N pods must be available
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: postgres-pdb
spec:
  minAvailable: 2     # at least 2 out of 3 must be available
  selector:
    matchLabels:
      app: postgres
```

With this PDB and a 3-node cluster:
- During rolling update, one pod is deleted at a time (2 available → satisfies minAvailable: 2)
- If a pod is unhealthy before the update starts, the PDB blocks the update (only 2 available, deleting one would leave 1 < minAvailable: 2)
- This ensures a quorum (majority) is always available for PostgreSQL's synchronous replication

```yaml
# Option 2: maxUnavailable — at most N pods can be unavailable
spec:
  maxUnavailable: 1   # at most 1 out of 3 can be unavailable at once
```

**Practical consideration — PDB blocks StatefulSet updates if a pod is already down:**

```
Cluster state: postgres-0=Running, postgres-1=Running, postgres-2=Pending
PDB: minAvailable: 2

StatefulSet update tries to delete postgres-2:
  Current available = 2 (postgres-0, postgres-1)
  Deleting postgres-2 would make available = 2 (postgres-2 is Pending, not counted)
  2 >= minAvailable:2 → OK, deletion is allowed

StatefulSet update tries to delete postgres-1:
  Current available = 2
  Deleting postgres-1 would make available = 1
  1 < minAvailable:2 → BLOCKED. Update pauses.
  Alert fires, operator investigates why postgres-2 is still Pending.
```

This is actually the desired behaviour — the PDB catches degraded cluster state and prevents updates from making it worse, forcing an operator to fix the unhealthy pod before continuing.

---

## Quick Reference — StatefulSet Cheat Sheet

### Common kubectl Commands

```bash
# Create/apply
kubectl apply -f statefulset.yaml

# Check status
kubectl get statefulset postgres
kubectl rollout status statefulset/postgres

# Scale
kubectl scale statefulset postgres --replicas=5

# Update image (triggers rolling update)
kubectl set image statefulset/postgres postgres=postgres:17

# Pause rolling update (set partition to current replicas)
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":3}}}}'

# Resume rolling update
kubectl patch statefulset postgres -p '{"spec":{"updateStrategy":{"rollingUpdate":{"partition":0}}}}'

# Check rollout history (limited for StatefulSets vs Deployments)
kubectl rollout history statefulset/postgres

# Rollback (goes back to previous spec)
kubectl rollout undo statefulset/postgres

# Force delete a stuck pod
kubectl delete pod postgres-1 --force --grace-period=0

# List PVCs created by StatefulSet
kubectl get pvc -l app=postgres

# Check individual pod DNS from within cluster
kubectl exec -it postgres-0 -- nslookup postgres-1.postgres-svc.default.svc.cluster.local

# Delete StatefulSet but keep PVCs (safe default)
kubectl delete statefulset postgres --cascade=orphan
```

### Common Mistakes to Avoid

| Mistake | Impact | Fix |
|---|---|---|
| Not creating Headless Service before StatefulSet | Pods have no stable DNS | Create Service first |
| Using a regular Service instead of Headless | No per-pod DNS entries | Set `clusterIP: None` |
| Forgetting DNS egress NetworkPolicy rule | Pods cannot resolve pod DNS | Allow UDP 53 to CoreDNS |
| Force-deleting pods without verifying node is dead | Potential split-brain and data corruption | Verify node is offline first |
| Manually deleting PVCs during scale-down | Data loss when scaling back up | Let PVCs remain (Retain policy) |
| Using `podManagementPolicy: Parallel` for databases | Primary not ready when replicas start | Use `OrderedReady` for databases |
| Setting `maxUnavailable > 1` in PDB for a 3-node cluster | Update can take majority offline | `minAvailable: 2` for 3-node cluster |
| Not setting resource requests/limits | Pod gets evicted or OOMKilled during peak | Always set both |



















# Kubernetes Networking — Debugging Guide
### How to Debug When Packets Are Not Reaching the Destination

---

## How to Think About Kubernetes Network Debugging

Before jumping into commands, always follow this mental model:

```
Step 1 — Identify WHICH scenario you are in
  → Same pod? Same node? Different nodes? Via a Service?

Step 2 — Identify WHERE the packet is being dropped
  → Is it a DNS issue? A routing issue? A firewall (NetworkPolicy) issue?
    An application issue? A CNI issue?

Step 3 — Start from the SOURCE, work toward the DESTINATION
  → Can the source pod send? Can the destination receive?
    Is something in the middle dropping packets?

Step 4 — Eliminate layers one by one
  → IP connectivity first, then DNS, then ports, then application
```

---

## Debugging Toolkit — Install These First

```bash
# Run a debug pod with all networking tools (netshoot is the best for K8s debugging)
kubectl run debug --image=nicolaka/netshoot -it --rm -- bash

# Or attach a debug container to an existing pod (Kubernetes 1.23+)
kubectl debug -it <pod-name> --image=nicolaka/netshoot --target=<container-name>

# Or run a temporary pod in a specific namespace
kubectl run debug -n <namespace> --image=nicolaka/netshoot -it --rm -- bash

# Tools available in netshoot:
# ping, curl, wget, nc (netcat), nslookup, dig, traceroute,
# tcpdump, ss, ip, iptables, nmap, iperf3, wireshark (tshark)
```

---

## Scenario 1 — Containers Within the Same Pod

### Architecture Recap

```
┌────────────────────────────────────────────────────┐
│                      POD                            │
│  Shared Network Namespace (one IP)                  │
│                                                     │
│  ┌───────────────┐        ┌──────────────────┐     │
│  │  Container A  │        │   Container B    │     │
│  │   (app:8080)  │        │  (sidecar:9090)  │     │
│  └───────────────┘        └──────────────────┘     │
│         │                          │               │
│   localhost:8080             localhost:9090         │
└────────────────────────────────────────────────────┘
```

### Common Problems

```
1. Container B cannot reach Container A on localhost
2. Port conflict — two containers binding to the same port
3. Container A not listening on 0.0.0.0 (listening only on 127.0.0.1)
4. Wrong port number in the call
5. Container A not started yet when Container B tries to connect
```

### Debugging Steps

---

#### Step 1 — Confirm both containers are running

```bash
# Check all containers in the pod are Running (not CrashLoopBackOff)
kubectl get pod <pod-name> -o wide

# Check each container's status individually
kubectl describe pod <pod-name>
# Look for:
#   State: Running  ← good
#   State: Waiting  Reason: CrashLoopBackOff  ← bad, check logs
#   Last State: Terminated  Reason: Error  ← check why it crashed

# Check logs of each container
kubectl logs <pod-name> -c <container-a>
kubectl logs <pod-name> -c <container-b>
kubectl logs <pod-name> -c <container-b> --previous   # logs from previous crash
```

---

#### Step 2 — Verify the destination container is actually listening

```bash
# Exec into ANY container in the pod (they share network namespace)
kubectl exec -it <pod-name> -c <container-a> -- sh

# Inside the pod — list all listening ports
ss -tlnp
# or
netstat -tlnp

# Expected output if container-b listens on 9090:
# LISTEN  0  128  0.0.0.0:9090  0.0.0.0:*  users:(("app",pid=1,fd=3))
#              ↑
#         0.0.0.0 means it listens on ALL interfaces including localhost
#         If you see 127.0.0.1:9090 → only loopback, that is correct for same-pod
#         If you see nothing → the app is not listening yet or crashed

# Check if specific port is in use
ss -tlnp | grep 9090
```

**Common mistake:** if the app inside container-b is listening only on `127.0.0.1:9090`
(loopback-only) — that is actually fine for same-pod communication because both
containers share the same loopback interface. However, if the app listens on a
specific container IP like `10.244.1.5:9090` instead of `0.0.0.0:9090`, it will
NOT be reachable from container-a via localhost.

---

#### Step 3 — Test connectivity with localhost

```bash
# From inside container-a, test TCP connectivity to container-b
kubectl exec -it <pod-name> -c <container-a> -- sh

# Test TCP connect (does not send data, just checks port is open)
nc -zv localhost 9090
# Expected: Connection to localhost 9090 port [tcp/*] succeeded!
# Failure:  nc: connect to localhost port 9090 (tcp) failed: Connection refused

# Test with curl if HTTP
curl -v http://localhost:9090/healthz
curl -v http://127.0.0.1:9090/healthz

# Test with wget
wget -qO- http://localhost:9090/

# Check if it works from the same container (sanity check)
kubectl exec -it <pod-name> -c <container-b> -- curl localhost:9090
```

---

#### Step 4 — Check for port conflicts

```bash
# If container-a also tries to bind to port 9090, it will fail
# Check if two containers are trying to use the same port
kubectl exec -it <pod-name> -c <container-a> -- ss -tlnp | grep 9090
kubectl exec -it <pod-name> -c <container-b> -- ss -tlnp | grep 9090

# If both show the same port → port conflict
# One container will have failed to bind — check its logs
kubectl logs <pod-name> -c <container-a> | grep -i "address already in use"
kubectl logs <pod-name> -c <container-b> | grep -i "address already in use"
```

---

#### Step 5 — Check if it is a timing/startup issue

```bash
# Container-a may try to connect to container-b before it is ready
# Check if container-b has a readiness probe configured
kubectl describe pod <pod-name> | grep -A 10 "Readiness"

# If no readiness probe, container-b might not be ready when container-a starts
# Solution: add an init container or retry logic in container-a

# Check startup order — init containers run before all app containers
kubectl describe pod <pod-name> | grep -A 5 "Init Containers"
```

---

#### Step 6 — Verify shared volume communication (if using files)

```bash
# If containers communicate via shared volume (emptyDir)
# Check the volume is mounted in both containers
kubectl describe pod <pod-name> | grep -A 20 "Volumes"
kubectl describe pod <pod-name> | grep -A 20 "Mounts"

# Check if container-a is writing the file
kubectl exec -it <pod-name> -c <container-a> -- ls -la /shared/
kubectl exec -it <pod-name> -c <container-a> -- cat /shared/data.txt

# Check if container-b can read the file
kubectl exec -it <pod-name> -c <container-b> -- ls -la /shared/
kubectl exec -it <pod-name> -c <container-b> -- cat /shared/data.txt

# Check permissions — container-b might not have read permission
kubectl exec -it <pod-name> -c <container-b> -- stat /shared/data.txt
```

---

#### Scenario 1 — Decision Tree

```
Container A cannot reach Container B at localhost:PORT
            │
            ▼
  Is Container B running?
    NO → check logs: kubectl logs <pod> -c <container-b>
    YES ↓
            ▼
  Is Container B listening on the port?
    NO  → app is not started or wrong port configured
    YES ↓
            ▼
  Is it listening on 0.0.0.0 or 127.0.0.1?
    0.0.0.0 or 127.0.0.1 → both work for same-pod localhost
    specific IP (10.x.x.x) → app misconfigured, must bind 0.0.0.0
            ▼
  Can you nc -zv localhost PORT from container-a?
    NO  → port conflict? firewall in the container?
    YES → application layer issue (wrong path, auth, etc.)
```

---

## Scenario 2 — Pods on the Same Node

### Architecture Recap

```
NODE
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  Pod A (10.244.1.2)            Pod B (10.244.1.3)               │
│  ┌──────────────┐              ┌──────────────┐                 │
│  │    eth0      │              │    eth0      │                 │
│  └──────┬───────┘              └──────┬───────┘                 │
│         │ veth-a                      │ veth-b                  │
│         │                             │                         │
│  ┌──────┴─────────────────────────────┴──────┐                  │
│  │                Bridge (cni0)               │                 │
│  │             10.244.1.1/24                  │                 │
│  └────────────────────────────────────────────┘                 │
└──────────────────────────────────────────────────────────────────┘
```

### Common Problems

```
1. CNI plugin not installed or misconfigured
2. Bridge (cni0) not created or wrong IP range
3. veth pair for one pod missing or not connected to bridge
4. iptables rules dropping traffic (NetworkPolicy)
5. Pod has wrong IP or no IP assigned
6. Application in destination pod not listening
```

### Debugging Steps

---

#### Step 1 — Confirm both pods are Running and have IPs

```bash
# Check pod status and IPs
kubectl get pods -o wide
# Expected:
# NAME      READY  STATUS   IP           NODE
# pod-a     1/1    Running  10.244.1.2   node-1
# pod-b     1/1    Running  10.244.1.3   node-1  ← same node
# If IP is <none> → CNI plugin failed to assign IP

# Check pod events for CNI errors
kubectl describe pod pod-a | grep -A 20 "Events"
# Look for: "Failed to create pod sandbox" or "network plugin is not ready"
```

---

#### Step 2 — Test basic IP connectivity (ICMP ping)

```bash
# From pod-a, ping pod-b's IP directly
kubectl exec -it pod-a -- ping -c 3 10.244.1.3

# If ping fails:
#   → Check if ICMP is blocked by NetworkPolicy (ping uses ICMP, not TCP/UDP)
#   → NetworkPolicy might not block ping but app might
#   → Use nc or curl instead of ping for TCP connectivity test

# Test TCP connectivity (more reliable than ping)
kubectl exec -it pod-a -- nc -zv 10.244.1.3 8080
# Success: Connection to 10.244.1.3 8080 port [tcp/http-alt] succeeded!
# Failure: nc: connect to 10.244.1.3 port 8080 (tcp) failed: Connection refused
#          Connection refused = pod-b is reachable but nothing listening on 8080
#          Connection timed out = pod-b is unreachable (NetworkPolicy or routing)
```

---

#### Step 3 — Check the bridge and veth pairs on the node

```bash
# SSH into the node or use a privileged pod
kubectl node-shell node-1     # if kubectl-node-shell is installed
# OR
kubectl debug node/node-1 -it --image=ubuntu -- bash

# On the node:
# Check if bridge exists and has an IP
ip addr show cni0
# Expected: inet 10.244.1.1/24

# List all veth pairs attached to the bridge
bridge link show
# Expected: veth-a state UP, veth-b state UP

# If a veth is DOWN:
ip link set <veth-name> up

# Check if pod's veth pair is in the right network namespace
# Get pod's network namespace
POD_ID=$(crictl pods --name pod-b -q)
crictl inspectp $POD_ID | grep netns
# Returns: /var/run/netns/xxxx

# Check network interfaces in the pod's namespace
ip netns exec /var/run/netns/xxxx ip addr
# Expected: eth0 with 10.244.1.3/24
```

---

#### Step 4 — Check routing on the node

```bash
# On the node, check routes
ip route show
# Expected:
# 10.244.1.0/24 dev cni0 proto kernel scope link src 10.244.1.1
# (route for the pod subnet via the bridge)

# Trace the route from node to pod-b's IP
ip route get 10.244.1.3
# Expected: 10.244.1.3 dev cni0 src 10.244.1.1
# If it shows going out through eth0 → routing table is wrong

# Check if packets reach veth-b using tcpdump on the node
tcpdump -i cni0 host 10.244.1.3 -n
# Then from pod-a: kubectl exec pod-a -- ping 10.244.1.3
# You should see ICMP packets on cni0
# If no packets → packets are not reaching the bridge
# If packets seen but no reply → pod-b is not responding
```

---

#### Step 5 — Check NetworkPolicy

```bash
# List NetworkPolicies in the namespace
kubectl get networkpolicies -n <namespace>
kubectl describe networkpolicy <policy-name> -n <namespace>

# A NetworkPolicy that selects pod-b and has an Ingress section
# will block ALL traffic not explicitly allowed

# Check if pod-b is selected by any NetworkPolicy
kubectl get networkpolicies -n <namespace> -o yaml | grep -A 5 "podSelector"

# Test with a policy that allows all traffic temporarily (for testing only)
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-test
  namespace: <namespace>
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - {}
  egress:
  - {}
EOF

# Test connectivity again
kubectl exec -it pod-a -- nc -zv 10.244.1.3 8080

# If it works now → NetworkPolicy was blocking it
# Delete the test policy after debugging:
kubectl delete networkpolicy allow-all-test -n <namespace>
```

---

#### Step 6 — Check iptables rules on the node

```bash
# On the node, check iptables FORWARD chain
# (packets between pods on the same node go through FORWARD)
iptables -L FORWARD -n -v | head -30

# Check if FORWARD is set to DROP
iptables -L FORWARD -n | grep policy
# Expected: Chain FORWARD (policy ACCEPT)
# If DROP → add a rule: iptables -P FORWARD ACCEPT (temporary)

# Check NetworkPolicy enforcement rules
iptables -L -n | grep KUBE-NET  # for calico
iptables -L -n | grep CILIUM    # for cilium

# Check if any rule is dropping traffic to pod-b's IP
iptables -L -n -v | grep 10.244.1.3

# Use tcpdump to see if iptables is dropping before it reaches the pod
tcpdump -i any host 10.244.1.3 -n
```

---

#### Scenario 2 — Decision Tree

```
Pod A cannot reach Pod B (same node, different IPs)
            │
            ▼
  Do both pods have IPs?
    NO → CNI plugin issue, check: kubectl describe pod
    YES ↓
            ▼
  Can you ping pod-b from pod-a?
    NO ↓                     YES ↓
    Check:                   Check port specifically:
    - cni0 bridge exists     nc -zv 10.244.1.3 8080
    - veth pairs UP          if fails: app not listening
    - iptables FORWARD       if works: app-level issue
    - NetworkPolicy
            ▼
  Does tcpdump on cni0 show packets?
    NO  → packet not reaching bridge (veth issue)
    YES → bridge is forwarding but pod-b not responding
            ↓
  Check pod-b is actually listening:
    kubectl exec pod-b -- ss -tlnp
```

---

## Scenario 3 — Pods on Different Nodes

### Architecture Recap

```
NODE 1 (192.168.1.10)                NODE 2 (192.168.1.20)
┌──────────────────────┐             ┌──────────────────────┐
│ Pod A: 10.244.1.2    │             │ Pod B: 10.244.2.3    │
│   ↓ veth             │             │   ↑ veth             │
│ Bridge: 10.244.1.1   │             │ Bridge: 10.244.2.1   │
│   ↓ eth0             │             │   ↑ eth0             │
└──────────────────────┘             └──────────────────────┘
          │                                    │
          └──────── Physical Network ──────────┘
                (VXLAN tunnel or BGP route)
```

### Common Problems

```
1. No route on node-1 to pod-b's subnet (10.244.2.0/24)
2. VXLAN tunnel endpoint (VTEP) not reachable
3. Firewall on the physical network blocking VXLAN (UDP 8472) or BGP (TCP 179)
4. MTU mismatch — VXLAN adds 50 bytes overhead, causing fragmentation
5. Node-to-node connectivity broken (physical network issue)
6. CNI agent not running on one of the nodes
7. NetworkPolicy blocking cross-node traffic
```

### Debugging Steps

---

#### Step 1 — Confirm pods are on different nodes and have IPs

```bash
kubectl get pods -o wide
# NAME    READY  STATUS   IP           NODE
# pod-a   1/1    Running  10.244.1.2   node-1
# pod-b   1/1    Running  10.244.2.3   node-2   ← different node

# Confirm pod CIDRs per node
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.podCIDR}{"\n"}{end}'
# node-1  10.244.1.0/24
# node-2  10.244.2.0/24
```

---

#### Step 2 — Test node-to-node connectivity first

```bash
# Before debugging pod networking, verify the nodes can reach each other
# Get node IPs
kubectl get nodes -o wide

# From a pod on node-1, ping node-2's IP
kubectl exec -it pod-a -- ping -c 3 192.168.1.20

# If node-to-node ping fails → physical network or cloud VPC issue
# All pod networking depends on working node-to-node connectivity

# Test from the node itself (ssh or node-shell)
ping 192.168.1.20
# If this fails: cloud security group, firewall, or network misconfiguration
```

---

#### Step 3 — Test cross-node pod connectivity

```bash
# From pod-a on node-1, try to reach pod-b on node-2
kubectl exec -it pod-a -- ping -c 3 10.244.2.3

# If ping fails but node-to-node works → routing or CNI issue
kubectl exec -it pod-a -- nc -zv 10.244.2.3 8080

# Check traceroute — see where the packet stops
kubectl exec -it pod-a -- traceroute 10.244.2.3
# Expected for VXLAN:
#   10.244.1.1 (cni0 bridge on node-1)    ← hop 1
#   192.168.1.20 (node-2 IP)              ← hop 2
#   10.244.2.3 (pod-b)                    ← hop 3
# If traceroute stops at hop 1 → routing issue on node-1
# If traceroute stops at hop 2 → VXLAN not working or node-2 not forwarding
```

---

#### Step 4 — Check routes on both nodes

```bash
# On node-1 — does it know how to reach 10.244.2.0/24?
ip route show | grep 10.244.2
# For Flannel VXLAN — expected:
#   10.244.2.0/24 via 192.168.1.20 dev flannel.1 onlink
# For Calico BGP — expected:
#   10.244.2.0/24 via 192.168.1.20 dev eth0 proto bird
# For AWS VPC CNI — expected:
#   10.244.2.3 dev eniXXX  (direct ENI route per pod)

# If NO route for 10.244.2.0/24 on node-1:
#   → CNI plugin failed to add route
#   → Check CNI agent logs (flannel, calico-node, cilium-agent)

# Check CNI agent pod logs
kubectl logs -n kube-system -l app=flannel --tail=50
kubectl logs -n kube-system -l k8s-app=calico-node --tail=50
kubectl logs -n kube-system -l k8s-app=cilium --tail=50
```

---

#### Step 5 — Check VXLAN tunnel (if using Flannel/Calico VXLAN)

```bash
# On node-1: check VTEP (VXLAN Tunnel Endpoint) interface
ip addr show flannel.1       # Flannel
ip addr show vxlan.calico    # Calico VXLAN

# Check VXLAN FDB (Forwarding Database) — maps remote pod MACs to node IPs
bridge fdb show dev flannel.1
# Expected: <mac-of-remote-pod> dst 192.168.1.20 ...
# If empty → flannel daemon not populating FDB → check flannel logs

# Test VXLAN tunnel connectivity (UDP port 8472 for Flannel, 4789 for Calico VXLAN)
nc -zu 192.168.1.20 8472    # test UDP reachability
# Note: UDP test may not be conclusive; use tcpdump instead

# Capture VXLAN traffic on node-1's eth0
tcpdump -i eth0 port 8472 -n
# Then from pod-a: ping pod-b's IP
# You should see UDP 8472 packets going from 192.168.1.10 to 192.168.1.20
# If no packets → packet not reaching node-1's eth0 → check routes/bridge
# If packets sent but no reply → node-2 not receiving or not responding
```

---

#### Step 6 — Check MTU issues

```bash
# VXLAN adds ~50 bytes of header overhead
# If node's eth0 MTU is 1500, then effective pod MTU should be ~1450
# MTU mismatch causes large packets to be silently dropped

# Check MTU on node's eth0
ip link show eth0 | grep mtu
# Expected: mtu 1500

# Check MTU inside pod
kubectl exec -it pod-a -- ip link show eth0 | grep mtu
# For VXLAN: should be ~1450 (1500 - 50 overhead)
# If it is 1500 inside the pod but the node uses VXLAN → MTU mismatch!

# Test with a large ping to detect MTU issues
kubectl exec -it pod-a -- ping -c 3 -M do -s 1400 10.244.2.3
# -M do = don't fragment  -s 1400 = packet size
# If large ping fails but small ping (default 56 bytes) succeeds → MTU issue

# Fix: set correct MTU in CNI config or adjust pod MTU
# For flannel: edit the flannel ConfigMap
kubectl edit configmap kube-flannel-cfg -n kube-system
# Set: "MTU": 1450
```

---

#### Step 7 — Check firewall between nodes

```bash
# In cloud environments, check security groups / firewall rules
# Allow these ports between nodes:
#   UDP 8472  — Flannel VXLAN
#   UDP 4789  — Calico VXLAN, standard VXLAN
#   TCP 179   — Calico BGP
#   IP proto 4 — IP-in-IP (Calico default mode)
#   TCP 5473  — Calico Typha
#   UDP/TCP 8285, 8472 — Flannel

# Test UDP port from node-1 to node-2
# UDP is hard to test; use nmap for UDP probe
nmap -sU -p 8472 192.168.1.20

# Better: capture on node-2 while sending from node-1
# On node-2:
tcpdump -i eth0 udp port 8472 -n
# On node-1:
kubectl exec -it pod-a -- ping 10.244.2.3
# If capture on node-2 shows packets arriving → firewall is open
# If nothing arrives on node-2 → firewall is blocking
```

---

#### Step 8 — Check CNI agent health on both nodes

```bash
# CNI agent (flannel/calico/cilium) must be Running on EVERY node
kubectl get pods -n kube-system -o wide | grep -E "flannel|calico|cilium"
# Every node should have one CNI agent pod Running

# If a node has no CNI pod or it is CrashLoopBackOff:
kubectl describe pod <cni-pod-name> -n kube-system
kubectl logs <cni-pod-name> -n kube-system --previous

# Check CNI configuration on the node
cat /etc/cni/net.d/*.conf    # on the node
ls /opt/cni/bin/             # CNI binaries must exist here
```

---

#### Scenario 3 — Decision Tree

```
Pod A (node-1) cannot reach Pod B (node-2)
            │
            ▼
  Can node-1 ping node-2?
    NO  → Physical network / cloud firewall issue
          Check: security groups, VPC routing, cloud console
    YES ↓
            ▼
  Does traceroute from pod-a stop at node-1 or beyond?
    Stops at node-1 (cni0):
      → No route to 10.244.2.0/24 on node-1
      → Check: ip route show | grep 10.244.2
      → Fix: CNI agent not running or not synced
    Reaches node-2 but stops:
      → Firewall blocking VXLAN/BGP between nodes
      → MTU mismatch causing large packets to drop
      → Check: tcpdump -i eth0 port 8472 on node-2
    Reaches pod-b subnet:
      → Routing on node-2 wrong, or pod-b not listening
            ▼
  Does tcpdump on node-2 eth0 show VXLAN packets arriving?
    NO  → Firewall between nodes
    YES → Check if packets reach cni0/bridge on node-2
            → Check VXLAN decapsulation (flannel.1 or calico VTEP)
```

---

## Scenario 4 — Pods and Services

### Architecture Recap

```
Pod A calls Service ClusterIP 10.96.5.10:80
        ↓
  iptables/IPVS on the node intercepts
        ↓
  DNAT: 10.96.5.10:80 → 10.244.2.3:8080 (selected pod)
        ↓
  Packet routed to pod-b (same or different node path)
        ↓
  conntrack records the translation for return traffic
```

### Common Problems

```
1. DNS resolution failing (pod cannot resolve Service name)
2. Service has no Endpoints (no pods matching selector)
3. kube-proxy not running or not syncing rules
4. iptables/IPVS rules not created for the Service
5. Readiness probe failing → pod removed from Endpoints
6. Port mismatch between Service spec and pod containerPort
7. NetworkPolicy blocking traffic to the Service
8. Service selector does not match pod labels
```

### Debugging Steps

---

#### Step 1 — Check DNS resolution of the Service name

```bash
# From pod-a, try to resolve the service
kubectl exec -it pod-a -- nslookup my-service
kubectl exec -it pod-a -- nslookup my-service.default.svc.cluster.local
kubectl exec -it pod-a -- dig my-service.default.svc.cluster.local

# Expected:
# Server: 10.96.0.10 (CoreDNS ClusterIP)
# Name: my-service.default.svc.cluster.local
# Address: 10.96.5.10 (Service ClusterIP)

# If DNS fails:
# 1. Check if CoreDNS pods are running
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns

# 2. Check pod's /etc/resolv.conf
kubectl exec -it pod-a -- cat /etc/resolv.conf
# Expected:
# nameserver 10.96.0.10   ← CoreDNS ClusterIP
# search default.svc.cluster.local svc.cluster.local cluster.local
# options ndots:5

# 3. Test connectivity to CoreDNS directly
kubectl exec -it pod-a -- nc -zv 10.96.0.10 53

# 4. Check NetworkPolicy blocking DNS (UDP 53)
# A NetworkPolicy with egress rules must explicitly allow DNS
```

---

#### Step 2 — Verify the Service exists and has correct configuration

```bash
# Check service details
kubectl get svc my-service -n <namespace>
# NAME        TYPE       CLUSTER-IP    EXTERNAL-IP  PORT(S)
# my-service  ClusterIP  10.96.5.10    <none>       80/TCP

kubectl describe svc my-service -n <namespace>
# Look for:
#   Selector:  app=my-app      ← must match pod labels
#   Port:      80/TCP
#   TargetPort: 8080/TCP       ← must match containerPort in pod
#   Endpoints: 10.244.2.3:8080, 10.244.3.5:8080   ← must not be empty!

# If Endpoints is <none> or empty → selector does not match any pod
# Check pod labels:
kubectl get pods -n <namespace> --show-labels | grep my-app
# If labels don't match → fix Service selector or pod labels
```

---

#### Step 3 — Check Endpoints (or EndpointSlices) — the most common issue

```bash
# Check if Endpoints object has any pods
kubectl get endpoints my-service -n <namespace>
# Expected:
# NAME        ENDPOINTS                       AGE
# my-service  10.244.2.3:8080,10.244.3.5:8080  5m

# If ENDPOINTS shows <none>:
# → No pods match the Service selector
# → Or all matching pods are not Ready (readiness probe failing)

# Check pod readiness
kubectl get pods -n <namespace> -l app=my-app
# READY column should show 1/1, not 0/1

# If 0/1: readiness probe is failing
kubectl describe pod <pod-name> -n <namespace> | grep -A 20 "Readiness"
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 "Conditions"

# Check the actual readiness probe manually
kubectl exec -it <pod-name> -- curl localhost:8080/healthz
# If this fails → app is not ready → not added to Endpoints

# Use EndpointSlices (newer API, default in modern clusters)
kubectl get endpointslices -n <namespace> | grep my-service
kubectl describe endpointslice -n <namespace> | grep -A 10 "Endpoints"
```

---

#### Step 4 — Test connectivity to the Service ClusterIP directly

```bash
# Get Service ClusterIP
SVC_IP=$(kubectl get svc my-service -n <namespace> -o jsonpath='{.spec.clusterIP}')
SVC_PORT=$(kubectl get svc my-service -n <namespace> -o jsonpath='{.spec.ports[0].port}')

echo "Service IP: $SVC_IP Port: $SVC_PORT"

# From pod-a, test TCP to Service ClusterIP
kubectl exec -it pod-a -- nc -zv $SVC_IP $SVC_PORT
# If this works → Service routing is OK, issue is DNS
# If this fails → check iptables/kube-proxy

# Test with curl
kubectl exec -it pod-a -- curl -v http://$SVC_IP:$SVC_PORT/

# Also test to the pod directly (bypassing Service)
POD_IP=$(kubectl get pod <pod-b> -o jsonpath='{.status.podIP}')
kubectl exec -it pod-a -- nc -zv $POD_IP 8080
# If direct pod works but Service IP fails → iptables/kube-proxy issue
# If direct pod also fails → underlying pod networking issue (see Scenario 2/3)
```

---

#### Step 5 — Check kube-proxy and iptables rules

```bash
# Check kube-proxy is running on the node
kubectl get pods -n kube-system -l k8s-app=kube-proxy -o wide
# Must be Running on EVERY node

# Check kube-proxy logs for errors
kubectl logs -n kube-system -l k8s-app=kube-proxy --tail=30

# On the node, verify iptables rules exist for the Service
# Get the iptables chain for the service
iptables -t nat -L KUBE-SERVICES -n | grep <service-cluster-ip>
# Expected output like:
# KUBE-SVC-XXXXX  tcp  --  0.0.0.0/0  10.96.5.10  tcp dpt:80

# If no rule for the Service IP → kube-proxy has not synced yet
# Check kube-proxy mode:
kubectl get configmap kube-proxy -n kube-system -o yaml | grep mode
# If mode: ipvs:
ipvsadm -Ln | grep 10.96.5.10
# Expected: TCP  10.96.5.10:80 rr  with real servers listed

# Check if iptables FORWARD is set to ACCEPT on the node
iptables -L FORWARD -n | grep policy
# Must be: Chain FORWARD (policy ACCEPT)
```

---

#### Step 6 — Check for port and selector mismatches

```bash
# Verify the Service targetPort matches the pod's containerPort
kubectl get svc my-service -n <namespace> -o yaml | grep -A 5 "ports"
# ports:
# - port: 80           ← what clients call
#   targetPort: 8080   ← what the pod listens on
#   protocol: TCP

# Verify the pod ACTUALLY listens on the targetPort
kubectl exec -it <pod-b> -- ss -tlnp | grep 8080
# If nothing → pod listens on different port → fix Deployment containerPort

# Check selector matches pod labels exactly (case-sensitive)
kubectl get svc my-service -n <namespace> -o jsonpath='{.spec.selector}'
# {"app":"my-app"}

kubectl get pods -n <namespace> --show-labels
# NAME      LABELS
# pod-b-0   app=my-app,version=v2   ← "app=my-app" present → matches
# pod-c-0   App=my-app              ← "App" (capital A) → does NOT match
```

---

#### Step 7 — Check NetworkPolicy affecting Service traffic

```bash
# NetworkPolicy applies to pod-to-pod traffic even when going via a Service
# The DNAT happens before NetworkPolicy enforcement
# So: "allow traffic to Service IP 10.96.5.10" is NOT enough
# You must allow traffic to the DESTINATION POD'S IP on the POD'S PORT

# Check if pod-b has a NetworkPolicy with ingress rules
kubectl get networkpolicies -n <namespace>
kubectl describe networkpolicy <name> -n <namespace>

# Correct NetworkPolicy allowing traffic to pod-b on port 8080:
cat <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-my-service
spec:
  podSelector:
    matchLabels:
      app: my-app        # select destination pods
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: pod-a     # allow from pod-a
    ports:
    - protocol: TCP
      port: 8080         # use POD port, not Service port
EOF
```

---

#### Step 8 — Debug DNS issues in depth

```bash
# Check CoreDNS ConfigMap for custom settings
kubectl get configmap coredns -n kube-system -o yaml

# Check if CoreDNS has errors
kubectl logs -n kube-system -l k8s-app=kube-dns | grep -i error

# Test DNS from inside a debug pod in the same namespace
kubectl run dns-test -n <namespace> --image=nicolaka/netshoot -it --rm -- bash
# Inside:
nslookup my-service
nslookup my-service.<namespace>.svc.cluster.local
nslookup kubernetes.default.svc.cluster.local  # baseline test

# Check search domains — common mistake is calling service from wrong namespace
# From a pod in namespace "staging", calling "my-service" resolves to:
# my-service.staging.svc.cluster.local (NOT my-service.default.svc.cluster.local)
# Use FQDN to avoid namespace confusion:
nslookup my-service.default.svc.cluster.local  # always works regardless of namespace
```

---

#### Scenario 4 — Decision Tree

```
Pod A cannot reach Service my-service
            │
            ▼
  Can pod-a resolve the Service DNS name?
    NO ↓                         YES ↓
    Check:                    Service IP resolved?
    - CoreDNS running           NO → DNS resolves to wrong IP
    - /etc/resolv.conf          YES ↓
    - NetworkPolicy             ▼
      blocking UDP 53      Can pod-a reach Service ClusterIP?
                             NO ↓              YES ↓
                           Check:           Service works!
                           - kube-proxy     Issue was DNS
                           - iptables rules
                           - IPVS config
                                ↓
                        Does Endpoints have pods listed?
                          NO ↓               YES ↓
                        Check:           Check iptables:
                        - Selector        KUBE-SVC chain
                        - Pod readiness   KUBE-SEP chain
                        - Pod labels      DNAT rules exist?
                                ↓
                        Can pod-a reach pod-b DIRECTLY?
                          NO → Scenario 2 or 3 problem
                          YES → kube-proxy/iptables broken
```

---

## Master Debug Commands Reference

### Essential One-Liners

```bash
# ── Pod and IP Status ──────────────────────────────────────────────────────
kubectl get pods -o wide -A                          # all pods with IPs and nodes
kubectl get pods -o wide --field-selector spec.nodeName=node-1  # pods on node-1
kubectl get endpoints -A                             # all service endpoints
kubectl get endpointslices -A                        # modern endpoint slices

# ── Service Debugging ──────────────────────────────────────────────────────
kubectl describe svc <svc-name>                      # endpoints + selector
kubectl get svc -o yaml | grep -A5 selector          # check selectors
kubectl get pods --show-labels | grep <label>        # verify pod labels

# ── DNS Debugging ──────────────────────────────────────────────────────────
kubectl exec -it <pod> -- nslookup <service-name>
kubectl exec -it <pod> -- cat /etc/resolv.conf
kubectl logs -n kube-system -l k8s-app=kube-dns

# ── Connectivity Testing ───────────────────────────────────────────────────
kubectl exec -it <pod> -- nc -zv <ip> <port>        # TCP connect test
kubectl exec -it <pod> -- curl -v http://<ip>:<port>/  # HTTP test
kubectl exec -it <pod> -- ping -c 3 <ip>             # ICMP test
kubectl exec -it <pod> -- traceroute <ip>             # route tracing

# ── NetworkPolicy ──────────────────────────────────────────────────────────
kubectl get networkpolicies -A
kubectl describe networkpolicy <name>

# ── iptables (on node) ────────────────────────────────────────────────────
iptables -t nat -L KUBE-SERVICES -n | grep <service-ip>
iptables -L FORWARD -n | grep policy
ipvsadm -Ln | grep <service-ip>                     # if IPVS mode

# ── CNI Debugging (on node) ───────────────────────────────────────────────
ip route show                                        # routing table
ip addr show cni0                                    # bridge interface
bridge link show                                     # veth pairs on bridge
ip link show                                         # all interfaces

# ── Packet Capture ────────────────────────────────────────────────────────
tcpdump -i cni0 host <pod-ip> -n                    # on bridge
tcpdump -i eth0 port 8472 -n                        # VXLAN traffic on node
tcpdump -i eth0 host <node-ip> -n                   # node-to-node traffic
kubectl exec -it <pod> -- tcpdump -i eth0 -n        # inside pod

# ── CNI Agent Logs ────────────────────────────────────────────────────────
kubectl logs -n kube-system -l app=flannel --tail=50
kubectl logs -n kube-system -l k8s-app=calico-node --tail=50
kubectl logs -n kube-system -l k8s-app=cilium --tail=50
kubectl logs -n kube-system -l k8s-app=kube-proxy --tail=50
```

---

## Summary Table

| Scenario | First thing to check | Most common cause | Key command |
|---|---|---|---|
| Same pod | Is destination container listening? | App not started or wrong port | `kubectl exec -c <c> -- ss -tlnp` |
| Same node | Do both pods have IPs? | CNI not assigning IPs | `kubectl get pods -o wide` |
| Same node | Can pod-a ping pod-b? | NetworkPolicy blocking | `kubectl get networkpolicies` |
| Diff nodes | Can nodes ping each other? | Cloud firewall / security group | `ping <node-ip>` from node |
| Diff nodes | Is there a route to remote pod CIDR? | CNI agent not syncing routes | `ip route show` on node |
| Diff nodes | MTU large packet test | VXLAN overhead, MTU mismatch | `ping -M do -s 1400 <pod-ip>` |
| Via Service | Does DNS resolve? | CoreDNS down or NP blocks UDP 53 | `nslookup <svc-name>` |
| Via Service | Are there Endpoints? | Selector mismatch or pod not Ready | `kubectl get endpoints <svc>` |
| Via Service | Do iptables rules exist? | kube-proxy not running | `iptables -t nat -L KUBE-SERVICES` |