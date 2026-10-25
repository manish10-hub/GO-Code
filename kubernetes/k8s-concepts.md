Kubernetes
- Kubernetes is an open source container orchestration tool that allows us to manage multiple containers at the same time.
- It schedules, runs and manages isolated containers running on virtual/physical/cloud machines
- Written in GoLang, developed by Google & later donated to CNCF (Cloud Native Computing Foundation)
- Cloud Services
  - Google Kubernetes Services (GKE)
  - Azure Kubernetes Services (AKS)
  - Elastic Kubernetes Services (EKS)
- Features
  - Orchestration
  - Auto scaling (vertical & horizontal)
  - Auto healing
  - Reliability (24*7) and Load balancing
  - Platform independent (cloud/virtual/physical)
  - Fault tolerance (node/pod failure)
  - Easy updates and Rollback (going back to previous version)
  - Health monitoring of containers
  - Batch execution (one time, sequential, parallel)
  - Storage Distribution
  - Secret Handling
- K8S over Swarm
  - GUI available, support scaling, inbuilt logging and monitoring tool (swarm uses 3rd party tool like splunk)
  
Architecture
- Master (Control Plane)
  - kube apiserver
    - front end of control plane, responsible for communication betn all components on master as well as relaying msgs to node kubelet
    - interacts directly with user (apply yml or json manifest)
  - etcd
    - distributed reliable key-value store, consistent and high available
    - stores metadata and status of cluster
    - Features
      - fully replicated (entire state is available on every node in cluster)
      - secure (implements automatic TLS with optional client certificate authentication)
      - fast (10,000 writes per second)
  - kube scheduler
    - take action on user request for creation and mngmt of pods
    - responsible for scheduling newly created pods to available nodes
    - filters out best node for pod by evaluating cluster topology, workload, etc
  - controller manager
    - manages actual state of cluster to match with desired state
    - cloud-controller-manager (for cloud) and kube-controller-manager (for non-cloud)
    - Components
      - Node controller (checking if node has been detected after it stops responding)
      - Route controller (responsible for setting up network routes)
      - Service controller (responsible for load balancer against services)
      - Volume controller (for creating, attaching and mounting volumes)
      - Replication controller (ensures that requested number of pod replicas are running at any given time, support equality based selector)
- Node
  - kubelet
    - k8s agent that manages the node, responsible for setting up pods using podSpec provided through ApiServer
  - container engine
    - pull images, start/stop containers, expose ports
  - kube proxy
    - network proxy that runs on each node in the cluster.
    - maintains network traffic on nodes inside or outside of the cluster
    - responsible for watching services and endpoint associated with each service
    - When the client is going to connect to the service using the virtual IP the kubeproxy is responsible for sending traffic to actual pods.
    - Kubernetes Services are implemented using iptables rules. Every time a Service has been altered, kube-proxy updates the iptables rules.
- Pod
  - basic unit of k8s, group of one or more containers that are deployed together on same host
  - A cluster is a group of nodes, having at-least 1 worker node and master node

Kubernetes Objects
- Pod (manages container)
- ReplicaSet (manages pods, auto scaling and healing, support equality and set based selectors )
- DeploymentSet (versioning and rollback)
- StatefulSet
- DaemonSet
- Service (expose pod processes to outside world, provides virtual IP)
- Volume (non ephemeral storage)
- ConfigMaps, Secrets (configure pods) 
- Namespace (dns: <service-name>.<ns-name>.svc.cluster.local) 
- Jobs

Kubernetes Object Management
- Imperative (commands)
- Declarative (manifest files yml/json)

Labels
- Used to identify set of resources
- Like Service identifies set of Pods it has to fwd request to, Deployment figures out set of Pods it has to manage  

Selectors
Equality based (=, !=)
Set based (in, notin, exists)

Failed Deployment
- Insufficient Quota
- Readiness probe failures
- Image pull errors
- Insufficient permission
- Limit ranges
- Application runtime misconfiguration

Service
- Works on selector to forward traffic/request to pods
- Types
  - ClusterIP (default, within the cluster)
  - NodePort (outside the cluster, Ports Range 30000 - 32767)
  - LoadBalancer (route external traffic to every node on the NodePort, created by cloud provider)
  - Headless (creates several endpoints used to produce DNS records. Each DNS record is bound to a Pod)

Volume
- Kubernetes volumes are used for storing temporary data.
- Types
  - emptyDir (shared between containers on the same pod)
  - hostPath (access content of pod from host, mounts directory from host node's filesystem into your pod)

Persistent Volume
- A PV is a cluster wide resource that you can use to store data in a way that it persists beyond the lifetime of a pod
- The PV is not backed by locally attached storage on a worker node but by n/w storage system such as EBS or NFS or distributed filesystem like Ceph
- The PersistentVolumeClaim request a PV with your desired specification (size, access modes, speed, etc) from k8s
- Portworx is a software-defined storage platform built for containers and micro-services.

Resource Quota
- CPU is specified in MiCpu and memory in MiB
- You can limit - compute, memory and storage
- Constraints
  - requests is amount of resources that the system will guarantee for container
  - limits is max amount of resources that k8s will allow container to use
- If request is not specified & limit is given, then request = limit
- Failure: failed quota: myquota: must specify limits.cpu,limits.memory,requests.cpu,requests.memory
minimum memory usage per Container is 500Mi, but request is 64Mi
- When available cpu or memory resources on the nodes reach their limit, Kubernetes will look for Pods that are using more resources than they requested. 
If some Pods containers have no resource requests/limits set (QoS: BestEffort), then by default those are considered to use more than requested.
- Kubernetes assigns Quality of Service classes (Burstable, BestEffort) to Pods based on the defined resources and limits.

Horizontal Pod Autoscaler (HPA)
- K8S has the possibility to automatically scale pods based on observed CPU utilisation
- HPA is implemented as a k8s API resource and a controller
- HPA is implemented as a controlled loop with a period controlled by --horizantal-pod-autoscaler-sync-period flag (default value of 30sec)
- Cool down period to wait before downscale is controlled by --horizantal-pod-autoscaler-downtime-stabilization flag (default value of 5min)
- During each period, the controller fetches the metrics from the resource metrics API for each pod targeted by HPA
- Metric server needs to be deployed in the cluster to provide metrics via resource metrics API (add this arg to skip certificate --kubelet-insecure-tls)

Init Containers
- Specialised containers that run before main container in a pod, always run to completion

HealthCheck / LivenessProbe
- Health checks or probes are carried out by the kubelet to determine when to recreate a container and used by services n deployments
- For e.g. LP could catch a deadlock, where an appln is running but enable to make progress
- For running health check, we would use specific cmds specific to the appln, if cmd succeeds, it returns 0, and kubelet considers container to be alive and healthy.
If cmd returns non zero value, kubelet kills the pod and recreate it.

ConfigMap & Secret
- ConfigMap are useful for storing and sharing non-sensitive, unencrypted configuration information.
- ConfigMap/Secret can be accessed via environment variables or volume from a container running in a pod.
- Secrets provide you mechanism to store information in a safe and reliable way. The Secret data on nodes is stored in tmpfs volumes. 

Pod lifecycle
- Pending 
  - the pod has been accepted by k8s system but its not running
  - one or more container image is still downloading
  - if pod cannot be scheduled coz of resource constraints
- Running
  - pod has been bound to a node
  - all of containers have been created
  - at least one container is still running
- Succeeded
  - All containers in pod have terminated in success and will not be restarted
- Failed
  - at least one container in pod has terminated in failure
- Unknown
  - state of pod could not be obtained due to error in n/w
- Completed
  - pod has run to completion

Pod Conditions
- PodScheduled (pod has been scheduled to a node)
- Ready (pod is able to serve request and will be added to load balancing pools of all matching services)
- Initialised (all init containers have started successfully)
- Unscheduled (the scheduler cannot schedule pod due to lack of resources or other constraints)
- ContainerReady (all containers in pod are ready)

HELM
- Packaging Manager, introduced first time in 2015. Helm 3 was released in Nov 2019.  
- Helm helps you manage k8s applications with helm charts which helps you define, install n upgrade k8s applications
- Writing and managing k8s yaml manifest for all required k8s objects can be time consuming n tedious task. Helm simplifies process n create single package that can be advertised to your cluster.
- HELM CHARTS are simply k8s yaml manifest combined into a single package that can be advertised to your k8s cluster.
- RELEASE is an instance of a chart running in a k8s cluster. Helm keeps track of all chart execution (install/upgrade/rollback)
- REPOSITORY - location where packaged charts can be stored n shared
- Helm 2 is a client server architecture, where Tiller is a server coz of absence of RBAC
- Helm 3 is a single client architecture
 
Minikube
- minikube is a tool that lets you run k8s locally
- minikube runs a single node k8s cluster on your personal computer
- Pre-requisites
  - 2 CPUs or more
  - 2GB of free memory
  - 20GB of free disk space
  - Container or virtual machine manager, such as: Docker, Hyper-V, VirtualBox or VMWare
- Reference: https://kubernetes.io/docs/tasks/tools/

Kind
kind is a tool for running local Kubernetes clusters using Docker container nodes

kubectl
- The Kubernetes command-line tool, kubectl, allows you to run commands against Kubernetes clusters.
- You can use kubectl to deploy applications, inspect and manage cluster resources, and view logs.

kubeadm 
- kubeadm tool is used to create and manage Kubernetes clusters
- performs the actions necessary to get a minimum viable, secure cluster up and running in a user friendly way

kube config
- contains details of kube apiserver endpoint, certificate details, user that will access cluster
- kubectl looks for kube config file to interact with cluster

Custom Resources
- Custom resources are extensions of the Kubernetes API.

-----------------------------------------

Lab
Login to AWS account -> Launch 3 instances -> Ubuntu 16.04 (2 vCPU and 4GB RAM) -> login as: ubuntu

Common commands
$ sudo su
$ apt-get update

// turn off swap space
$ swapoff -a
$ nano /etc/fstab		// comment line swapfile
// update hostname, hosts and set static ip
$ nano /etc/hostname	// kmaster
$ ifconfig
$ nano /etc/network/interfaces		
// Add below lines at end of file
auto enp0s8
iface enp0s8 inet static
address <ip-adress>
$ nano /etc/hosts		// <ip-adress> kmaster
$ reboot
// install openSSH server
$ sudo apt-get install openssh-server

$ apt-get install apt-transport-https curl
$ sudo apt install docker.io -y
$ docker --version
$ systemctl start docker
$ systemctl enable docker
// setup open gpg key required for intra cluster communication
$ curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add 
$ nano /etc/apt/sources.list.d/kubernetes.list
$ deb http://apt.kubernetes.io/ kubernetes-xenial main
(exit -> ctrl+x -> caps Y -> enter)
$ apt-get update
$ apt-get install -y kubelet kubeadm kubectl kubernetes-cni

//update k8s configuration
$ nano /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
// add below line
Environment="cgroup-driver=systemd/cgroup-driver=cgroupfs"

Master node (2 core CPUs and 4GB RAM)
// initialize k8s cluster
$ kubeadm init
$ mkdir -p $HOME/.kube
$ cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
$ chown $(id -u):$(id -g) $HOME/.kube/config
// deploy flannel node network for its repository path
$ kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
$ kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/k8s-manifests/kube-flannel-rbac.yml

// setup dashboard
Reference: https://www.youtube.com/watch?v=YXfLAWGTI38

Worker node (1 core CPU and 4GB RAM)
$ kubeadm join ...

$ kubectl get nodes
$ kubectl cluster-info			// kubectl cluster-info dump
$ kubectl api-resources

Launch instance -> ubuntu 18.04 -> t2.medium (2 vCPU)
// install docker
// install kubectl
$ curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
// install minikube
$ curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
$ apt install conntrack
$ ctrl+L		// clear
$ minikube start --vm-driver=none
$ minikube status
$ kubectl version
$ kubectl get nodes
$ kubectl describe node ip-...

$ kubectl apply -f pod1.yml					// kubectl create -f pod1.yml
$ kubectl get pods
$ kubectl get pods -o wide
$ kubectl describe pod testpod				// kubectl describe pod/testpod			
$ kubectl describe pods
$ kubectl logs -f testpod					// kubectl logs -f testpod -c c00
$ kubectl exec testpod -c c00 -- hostname -i
$ kubectl exec testpod -c c00 -- cat /etc/os-release
$ kubectl exec testpod -it -c c00 -- /bin/bash
$ curl <pod-ip>:80
$ kubectl delete pod testpod				// kubectl delete -f pod1.yml

$ kubectl label pods testpod myname=jesrani
$ kubectl get pods --show-labels
$ kubectl get pods -l myname=jesrani
$ kubectl get pods -l 'myname in (jesrani,gore)'
$ kubectl get pods -l myname!=jesrani
$ kubectl delete pod -l myname!=jesrani

$ kubectl label nodes <node-name> hardware=t2-medium

$ kubectl apply -f pod2.yml
$ kubectl get rc
$ kubectl describe rc sample-object
$ kubectl scale --replicas=3 rc -l myname=ajesrani
$ kubectl delete rc sample-object

$ kubectl apply -f pod2.yml
$ kubectl get rs
$ kubectl scale --replicas=3 rs/sample-object
$ kubectl delete rs/sample-object

$ kubectl apply -f pod2.yml
$ kubectl get deployments				// kubectl get deploy
$ kubectl describe deployment sample-object
$ kubectl scale deployment sample-object --replicas=3
$ kubectl apply -f pod2.yml			// redeploy with changes
$ kubectl rollout history deployment/sample-object
$ kubectl rollout undo deployment/sample-object --to-revision=1
$ kubectl rollout status deployment/sample-object

$ kubectl rollout pause deployment/sample-object
$ kubectl set image deploy/sample-object c00=nginx:1.9.1
$ kubectl set resources deployment sample-object -c=c00 --limits=cpu=200m,memory=512Mi
$ kubectl rollout resume deployment/sample-object

$ kubectl rollout restart deployment mockserver -n  mockserver

$ wget -O metricserver.yml https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
$ kubectl apply -f metricserver.yml
$ kubectl get pods -n kube-system
$ kubectl logs -f metrics-server-7795cfc6ff-gsxdt -n kube-system
$ kubectl get all
$ kubectl autoscale deployment sample-object --min=2 --max=5 --cpu-percent=80
$ kubectl get hpa
$ kubectl describe hpa sample-object
$ watch kubectl get all
$ kubectl exec sample-object-86cd447d9-l7sd9 -it -- /bin/bash
  # apt update
$ kubectl delete hpa sample-object
$ kubectl delete deployment sample-object

$ kubectl get svc

$ kubectl get namespaces
$ kubectl config set-context --current --namespace=genctl
$ kubectl config set-context $(kubectl config current-context) --namespace=genctl
$ kubectl config view | grep namespace

$ kubectl apply -f pv.yml
$ kubectl get pv
$ kubectl get pvc

$ kubectl create configmap myconfigmap --from-file=sample.conf
$ kubectl get cm
$ kubectl describe cm myconfigmap
$ kubectl create secret generic mysecret --from-file=sample.conf
$ kubectl get secret
$ kubectl describe secret mysecret

$ 

GCP -> Kubernetes Engine -> Clusters -> Create Cluster -> Create -> Connect

$ kubectl get pods -n kube-system
$ kubectl logs my-scheduler -n kube-system --since=5s


-- Credit: Vivek Singh
Ref: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/

$ brew install kubectl
$ kubectl version --client
$ brew install kind
$ kind create cluster
$ kind get clusters
$ alias k='kubectl'
$ cat $HOME/.kube/config
$ k get nodes
$ k get pods -n kube-system
$ k describe pod -n kube-system kube-apiserver-kind-control-plane
$ k get ns
$ k proxy
$ curl localhost:8001/api/v1/namespaces/kube-system/pods
$ k get pod -n kube-system kube-apiserver-kind-control-plane -v 6
$ curl localhost:8001/apis/apps/v1/namespaces/kube-system/deployments
$ k exec -it -n kube-system kube-controller-manager-docker-desktop -- sh
  # kube-controller-manager --help

$ kubectl create ns nginx-ns
$ kubectl run nginx --image nginx -n nginx-ns --generator=run-pod/v1
$ kubectl get pods -n nginx-ns -owide
$ kubectl create deployment nginx -n nginx-ns --image nginx --dry-run -oyaml
$ kubectl get all
$ k label deployment nginx tier=frontend
$ k get deploy nginx -oyaml
$ k scale deploy nginx --replicas 3
$ k get pods -l app=nginx
$ k expose deploy nginx --port 80
$ k get endpoints

------------------------------------------

## CKAD: Certified Kubernetes Application Developer - mumshad mannambeth

Core Concepts (Day 1)
  - Kubernetes Architecture
    - nodes / minions
    - cluster
    - master vs worker nodes
    - Components (API server, etcd service, kubelet, container runtime, controller, scheduler)
    - kubectl
  - Pods
  - ReplicaSet
  - Deployment
  - Namespace
  - Imperative commands

Configuration (Day 1)
  - Commands and Arguments  (CMD - cmd line parameter will be replaced entirely. ENTRYPOINT - cmd line parameter will be appended. ENTRYPOINT+CMD - CMD will serve as default value)
  - Environment variables
  - ConfigMap
  - Secrets (data is stored in encoded format)
  - Security Context (limit abilities of the container via linux capabilities such as MAC_ADMIN, KILL. See full list at /usr/include/linux/capability.h)
  - Service Account
  - Resource Requirements
  - Taints and Tolerations (tells node to only accept pod with certain tolerations, it does not tell pod to go to particular node)
  - Node Selectors and Affinity (restrict pod to certain node)

Multi-Container Pods (Day 2)
  - Ambassador, Adapter, Sidecar
  - Init Containers

Observability (Day 2)
  - Readiness and Liveness probes
  - Container Logging
  - Monitoring (Metrics server, Prometheus, Elastic Stack, DataDog, dynatrace)

Pod Design (Day 2)
  - Labels, Selectors (used to group and select objects) and Annotations (used to record details for info purpose)
  - Rolling Updates & Rollbacks in deployments
  - Deployment Strategy
    - Recreate
    - RollingUpdate (default)
    - Blue/Green (new version is deployed along side old version)
    - Canary (route less traffic to the new deployment)
  - Jobs
  - CronJobs

State Persistence (Day 2)
  - Volumes
  - Persistent Volumes
  - Persistent Volume Claims
  - Storage Class (Dynamic Provisioning)
  - Stateful Sets (Ordered graceful deployment, Stable unique network identifier)
  - Headless Service (create DNS entry for each pod)

Services & Networking (Day 3)
  - Services (enable connectivity)
  - Ingress (access application using single external URL, configure to route to different services, implement SSL security)
  - Network Policies (manage traffic for the app)
    - Ingress traffic (incoming traffic from the user)
    - Egress traffic (outgoing request to the app)

Security (Day 3)
  - Authentication (Service accounts)
  - Authorization (RBAC)
  - KubeConfig
  - Role and RoleBinding (Namespaced)
  - ClusterRole and ClusterRoleBinding (Cluster scoped)
  - Admission controllers (perform additional operations before pods get created)
    - Mutating (change request)
    - Validating (validate request)
  - Custom Resource Definition (extension of the Kubernetes API)
  - Custom Controllers (developing in go with k8s go client provides support for shared informers like caching and queuing mechanism)

Helm Fundamentals (Day 3)
  - Helm Chart (Kubernetes YAML template-files combined into a single package)
  - Helm Release (Installed instance of a Chart)
  - Helm Values (Customise the YAML template-files)


------------------------------------------

// Pod
$ k run redis --image=redis123 --dry-run=client -o yaml > redis.yaml
$ k run redis --image=redis:alpine --labels="app=hazelcast,env=prod" --env="DNS_DOMAIN=cluster"
$ k run nginx --image=nginx --port=8080
$ k run httpd --image=httpd:alpine --port=80 --expose=true // create pod + service
$ k replace --force -f /tmp/kubectl-edit-1138259109.yaml
$ k run busybox123 --image=busybox --dry-run=client -o yaml --command -- sh -c 'sleep 1d'         // --image=curlimages/curl
$ k delete pod ubuntu-sleeper --force
$ k get pods --selector type=front-end,env=dev
$ k delete pods -l type=runner                            // delete set of pods
$ k label pod -l type=runner protected=true               // set labels for specific set of pods
$ k get pod -o yaml | grep -A10 metadata                  // show annotations
$ k get pods -L env                                       // list pod with label env
$ k get pod pod1 -o jsonpath="{.status.phase}"            // show status
$ k get pod pod1 -o custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[].image --sort-by=.metadata.name
$ k get pods -o jsonpath="{range .items[*]} {.metadata.name} {.spec.containers[*].resources} {'\n'}"
$ k describe pods | egrep "^(Name:|    Requests:)" -A1
$ k describe pods | less -p Requests                      // highlight Requests
$ k get pod pods -o name
$ k get pods -n blue dark-blue-app --as dev-user
$ k exec secret-handler -- env | grep SECRET1             // check env
$ k exec secret-handler -- find /tmp/secret2              // check mounted volumes
$ k exec frontend -- wget google.com                      // check egress
$ k get pods --watch
$ k logs web -f --previous                                // view logs of previous pod

// ReplicaSet
$ k explain replicaset
$ k edit rs new-replica-set
$ k scale rs new-replica-set --replicas=5

// Deployment
$ k get all
$ k create deploy --help
$ k create deploy httpd-frontend --image=httpd:2.4-alpine --replicas=3
$ k scale deploy httpd-frontend --replicas=4
$ k rollout status deploy/myapp-deployment
$ k rollout history deploy/myapp-deployment
$ k rollout history deploy/myapp-deployment --revision=1                      // check the status of each revision
$ k rollout undo deploy/myapp-deployment
$ k rollout undo deploy/myapp-deployment --to-revision=1
$ k set image deploy/myapp-deployment nginx-container=nginx:1.9.1
$ k set image deploy/myapp-deployment nginx-container=nginx:1.9.1 --record    // save the command and shows it in change-cause field
$ k create -f deployment-definition.yml --record

// ConfigMap
$ k create cm app-config --from-literal APP_COLOR=blue
$ k create cm app-config --from-file app_config.properties        // [config.props=app_config.properties]

// Secret
$ k create secret generic app-secret --from-literal DB_HOST=mysql --from-literal 'password=A!B*d$zDsb='
$ k create secret generic app-secret --from-file=app_secret.properties
$ k create secret tls -n webhook-demo webhook-server-tls \
> --cert /root/keys/webhook-server-tls.crt \
> --key /root/keys/webhook-server-tls.key

// Security context
$ docker run --user=1001 ubuntu
$ docker run --cap-add MAC_ADMIN ubuntu
$ docker run --cap-drop KILL ubuntu
$ docker run --privileged ubuntu
$ k exec ubuntu-sleeper -- whoami

// Service Account
$ k create sa dashboard-sa                                                                // creates token, stored as secret object
$ k describe sa dashboard-sa                                                              // shows secret name
$ k get secret dashboard-sa -o yaml                                                       // has annotation kubernetes.io/service-account.name, shows base64 encoded token
$ k describe secret dashboard-sa                                                          // shows base64 decoded token
$ curl https://192.168.56.70:6443/api -insecure --header "Authorization: Bearer eyJh..."  // used for making rest call to k8s api
$ k create token dashboard-sa                                                             // create an authorization token
### Test
$ k exec -it pod1 -- sh
  $ TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
  $ curl -k https://kubernetes.default/api/v1/secrets -H "Authorization: Bearer ${TOKEN}"                       // insecure https connection
  $ CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  $ curl --cacert ${CACERT} https://kubernetes.default/api/v1/secrets -H "Authorization: Bearer ${TOKEN}"       // encrypted https connection


// Taints (on Node) and Tolerations (on Pod)
$ k taint nodes node1 app=blue:NoSchedule   // taint-effect: NoSchedule | PreferNoSchedule | NoExecute (existing pod will be evicted)
$ k taint node controlplane node-role.kubernetes.io/control-plane:NoSchedule-

// Node Selectors (single label and selector) and Affinity (advanced expressions to limit pod placement on specific nodes)
$ k label nodes node-1 size=Large

// Job
$ k create job busybox-job --image=busybox --dry-run=client -o yaml -- sh -c "sleep 2 && echo done"
$ k create cronjob my-job --image=busybox --schedule="*/1 * * * *" -- /bin/sh -c date
$ k create job dice-test --from=cronjob/dice    // Manually trigger cronjob

// Storage
$ gcloud beta compute disks create --size 1GB --region us-east pd-disk            // create disk on google cloud

// Service
$ k expose pod redis --port=6379 --name redis-service // This will automatically use the pod's labels as selectors
$ k expose deploy ingress-controller -n ingress-space --name ingress --port=80 --target-port=80 --type NodePort // you cannot specify the node port
$ k create service clusterip redis --tcp=6379:6379  // This will assume selectors as app=redis
$ k create service nodeport nginx --tcp=80:80 --node-port=30080
### check the connection using a temporary Pod
# k run tmp --restart=Never --rm -i --image=nginx:alpine -- curl -m 5 project-svc.pluto:3333    // using service
# k run tmp --restart=Never --rm -i --image=nginx:alpine -- curl 10.44.0.78                     // using pod cluster ip
# k run tmp --restart=Never --rm -i --image=busybox -- wget -O- 10.44.0.78                      // using pod cluster ip
$ k exec busybox -- nslookup nginx-service                                                      // test service within cluster
$ k exec busybox -- nslookup 10-50-192-4.default.pod.cluster.local                              // test pod for dns lookup
$ k run curl alpine --image=alpine/curl --rm -it -- sh                                          // exec into pod and curl service
### check connectivity using nodeport
$ k get nodes -o wide
$ curl 192.168.100.11:30100
### test connectivity
$ cat /root/curl-test.sh
for i in {1..35}; do
   kubectl exec --namespace=kube-public curl -- sh -c 'test=`wget -qO- -T 2  http://webapp-service.default.svc.cluster.local:8080/info 2>&1` && echo "$test OK" || echo "Failed"';
   echo ""
done

// Ingress
$ k create ingress ingress-test --rule="wear.my-online-store.com/wear*=wear-service:80"
$ k create ingress ingress-pay --rule="/pay=pay-service:80"
$ curl <ingress-ip>/hello

// Role
$ k create role developer-role --resource=pods,svc,pvc --verb="*" -n development
$ k create rolebinding developer-rolebinding --role=developer-role --user=martin -n development
$ k auth can-i delete nodes [--as dev-user]                                       // check access
$ k -n project auth can-i create pod --as system:serviceaccount:project:processor
$ k get pods --as dev-user -n blue                                                // check if the dev-user can list pods

// metrics-server (in-memory monitoring solution)
$ git clone https://github.com/kodekloudhub/kubernetes-metrics-server.git
$ cd kubernetes-metrics-server && kubectl create -f .
$ k top node
$ k top pod
$ k top pod --containers=true
$ k top pod -l name=cpu-loader -A --sort-by cpu

// Helm
$ sudo apt-get install helm
$ helm ls
$ helm uninstall internal-issue-report-apiv1
### list the charts
$ helm repo list
$ helm repo update
$ helm search repo nginx
### upgrade a release
$ helm upgrade internal-issue-report-apiv2 bitnami/nginx        // [helm rollback]
### install new release
$ helm show values bitnami/apache | yq e                        // show a long list of all possible value-settings
$ helm install internal-issue-report-apache bitnami/apache --set replicaCount=2 --set image.debug=true
### delete broken release
$ helm ls -a
$ helm uninstall internal-issue-report-daniel


-----------------------------------------

definition.yml
  apiVersion:
  kind:
  metadata:
  spec:

pod-definition.yml (1)
  apiVersion: v1
  kind: Pod
  metadata:
    name: myapp-pod
    labels:
      type: front-end
    annotations:
      version: 1.34.0
  spec:
    containers:
      - name: nginx-container
        image: nginx
        ports:
        - containerPort: 80

rc-definition.yml
  apiVersion: v1
  kind: ReplicationController
  metadata:
    name: myapp-rc
    labels:
      type: front-end
  spec:
    template:
      metadata:
        name: myapp-pod
        labels:
          type: front-end
      spec:
        containers:
          - name: nginx-container
            image: nginx
    replicas: 3

rs-definition.yml
  apiVersion: apps/v1
  kind: ReplicaSet
  metadata:
    name: myapp-rs
    labels:
      type: front-end
  spec:
    template:
      metadata:
        name: myapp-pod
        labels:
          type: front-end
      spec:
        containers:
          - name: nginx-container
            image: nginx
    replicas: 3
    selector:                 // rs can also manage pods that were not created as a part of rs creation
      matchLabels:
        type: front-end

deploy-definition.yml (9)
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: myapp-deploy
    labels:
      type: front-end
  spec:
    template:
      metadata:
        name: myapp-pod
        labels:
          type: front-end
      spec:
        containers:
          - name: nginx-container
            image: nginx
    replicas: 3
    selector: 
      matchLabels:
        type: front-end
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate

stateful-definition.yml
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: mysql
  spec:
    template:
      metadata:
        name: myapp-pod
        labels:
          app: mysql
      spec:
        containers:
          - name: mysql
            image: mysql
    replicas: 3
    selector: 
      matchLabels:
        app: mysql
    serviceName: mysql-h                  // Headless Service (must be the actual service also)
  //podManagementPolicy: Parallel         // [OrderedReady]
    volumeClaimTemplates:
    - metadata:
        name: pvc-vol
      spec:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 500Mi
      //storageClassName: google-storage

ns-definition.yml
  apiVersion: v1
  kind: Namespace
  metadata:
    name: dev

args-env-definition.yml (2)
  apiVersion: v1 
  kind: Pod 
  metadata:
    name: ubuntu-sleeper-3
  spec:
    containers:
    - name: ubuntu
      image: ubuntu
      command: ["sleep","1200"]     // overides entrypoint instruction in the dockerfile
  //  command: ["kube-proxy","--config=/var/lib/kube-proxy/config.conf"]
  //  command:
  //    - "sleep"
  //    - "1200"
      args: ["10"]                  // overrides cmd instruction
      env:
      - name: APP_COLOR
        value: pink
      - name: APP_COLOR
        valueFrom:
          configMapKeyRef:          // [secretKeyRef]
            name: app-config
            key: APP_COLOR
      envFrom:
      - configMapRef:               // [secretRef]
          name: app-config
      volumeMounts:
      - name: app-config-volume     // [app-secret-volume]
        readOnly: true              // default is false
        mountPath: /etc/config
      - name: kube-proxy
        mountPath: /var/lib/kube-proxy
    volumes:
    - name: app-config-volume
      configMap:
        name: app-config
        defaultMode: 0755
    - name: app-secret-volume       // each attribute in the secret is created as a file with value of secret as its content
      secret:
        secretName: app-secret
    - name: kube-proxy
      configMap:
        name: kube-proxy
        defaultMode: 420

cm-definition.yml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: app-config
  data:
    APP_COLOR: blue
    config.conf: |-
      apiVersion: kubeproxy.config.k8s.io/v1alpha1
    kubeconfig.conf: |-
      apiVersion: v1

secret-definition.yml
  apiVersion: v1
  kind: Secret
  metadata:
    name: app-secret
  data:
    DB_HOST: bX1zcWw==    // echo -n "mysql" | base64

sa-definition.yml (3)
  apiVersion: v1 
  kind: Pod 
  metadata:
    name: web-dashboard
  spec:
    serviceAccount: default                 [serviceAccountName]
  //serviceAccount: dashboard-sa
    containers:
    - name: nginx
      image: nginx
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-api-access-42kv6
    volumes:
    - name: kube-api-access-42kv6
      projected:
        sources:
        - serviceAccountToken:
            expirationSeconds: 3600
            ...

sec-definition.yml (3)
  apiVersion: v1 
  kind: Pod 
  metadata:
    name: ubuntu-sleeper-3
  spec:
    securityContext:          // support at pod level as well
      runAsUser: 1000
      fsGroup: 2000
    containers:
    - name: ubuntu
      image: ubuntu
      securityContext:
        allowPrivilegeEscalation: false
        capabilities:         // only supported at container level
          add: ["MAC_ADMIN"]

resource-definition.yml (4)
  apiVersion: v1 
  kind: Pod 
  metadata:
    name: ubuntu-sleeper-3
  spec:
    containers:
    - name: ubuntu
      image: ubuntu
      resources:
        requests:
          memory: "1Gi"
          cpu: 1
        limits:
          memory: "2Gi"
          cpu: 2

limit-cpu-definition.yml              
  apiVersion: v1 
  kind: LimitRange            // define default value for containers, applicable at namespace level 
  metadata:
    name: cpu-constraint
  spec:
    limits:
    - default:
        cpu: 500m
      defaultRequest:
        cpu: 500m
      max:
        cpu: "1"
      min:
        cpu: 100m
      type: Container

limit-memory-definition.yml
  apiVersion: v1 
  kind: LimitRange 
  metadata:
    name: memory-constraint
  spec:
    limits:
    - default:
        memory: 1Gi
      defaultRequest:
        memory: 1Gi
      max:
        memory: 1Gi
      min:
        memory: 500Mi
      type: Container

quota-definition.yml
  apiVersion: v1 
  kind: ResourceQuota         // restrict total amount of resources consumed, applicable at namespace level
  metadata:
    name: resource-quota
    namespace: dev
  spec:
    hard:
      pods: "10"
      requests.cpu: 4
      requests.memory: 4Gi
      limits.cpu: 10
      limits.memory: 10Gi

tol-sel-aff-definition.yml (5)
  apiVersion: v1
  kind: Pod
  metadata:
    name: myapp-pod
  spec:
    containers:
      - name: nginx-container
        image: nginx
    nodeName: node02                      // to schedule on specific node
    nodeSelector:
      size: Large
    //kubernetes.io/hostname: node02      // to schedule on specific node
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: size
              operator: In
              values:
              - Large
    affinity:                             // There should be only ever one Pod of that Deployment running on one worker node (alternate TopologySpreadConstraints)                           
      podAntiAffinity:                                    
        requiredDuringSchedulingIgnoredDuringExecution:   
        - labelSelector:                                  
            matchExpressions:                             
            - key: id                                     
              operator: In                                
              values:                                     
              - very-important                            
          topologyKey: kubernetes.io/hostname
    tolerations:
    - key: "app"
      operator: "Equal"
      value: "blue"
      effect: "NoSchedule"
    - key: node-role.kubernetes.io/control-plane
      effect: NoSchedule
      

multi-init-cont-definition.yml (6)
  apiVersion: v1
  kind: Pod
  metadata:
    name: app
  spec:
    initContainers:               // each init container is run one at a time in sequential order
    - name: init-myservice 
      image: busybox
      command: ['sh', '-c', 'git clone <some-repository-that-will-be-used-by-application> ;']
    - name: init-mydb
      image: busybox:1.28
      command: ['sh', '-c', 'until nslookup mydb; do echo waiting for mydb; sleep 2; done;']
    containers:
    - image: kodekloud/filebeat-configured
      name: sidecar
    - image: kodekloud/event-simulator
      name: app

read-live-definition.yml (7)
  apiVersion: v1
  kind: Pod
  metadata:
    name: myapp-pod
  spec:
    containers:
      - name: nginx-container
        image: nginx
        readinessProbe:             // [livenessProbe]
          httpGet:
            path: /api/ready
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 5
          failureThreshold: 8       // default is 3
    //    tcpSocket:
    //      port: 3306
    //    exec:
    //      command:
    //        - /bin/sh
    //        - -c
    //        - >
    //          while true; do
    //            echo "i luv cncf" >> /ckad/log/input.log; sleep 10;
    //          done

job-definition.yml
  apiVersion: batch/v1
  kind: Job
  metadata:
    name: math-add-job
  spec:
    completions: 3
    parallelism: 3
    template:
      spec:
        containers:
        - name: math-add
          image: ubuntu
          command: ['expr', '3', '+', '2']
        restartPolicy: Never      [OnFailure | Always]
    backoffLimit: 4               // attempts
    activeDeadlineSeconds: 20     // fails job if task not completed within 20 secs


cron-definition.yml
  apiVersion: batch/v1
  kind: CronJob
  metadata:
    name: dice
  spec:
    schedule: "*/1 * * * *"
    successfulJobsHistoryLimit: 2   // Keep 2 completed jobs
    failedJobsHistoryLimit: 4       // Keep 4 failed jobs
    jobTemplate:
      spec:
        template:
          spec:
            containers:
            - name: dice
              image: kodekloud/throw-dice

volume-definition.yml (8)
  apiVersion: v1
  kind: Pod
  metadata:
    name: myapp-pod
  spec:
    containers:
      - name: alpine
        image: alpine
        command: ["/bin/sh", "-c", "shuf -i 0-100 -n 1 >> /opt/number.out;"]
        volumeMounts:
        - mountPath: /opt
          name: data-volume
    volumes:
    - name: data-volume
      hostPath:
        path: /data
        type: Directory                     // [DirectoryOrCreate]
    //emptyDir: {}
    //awsElasticBlockStore:
    //  volumeID: <volume-id>
    //  fsType: ext4
    //persistentVolumeClaim:
    //  claimName: pvc-vol

pv-definition.yml (10)
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: pv-vol
  spec:
    accessModes:
      - ReadWriteOnce                         // [ReadOnlyMany | ReadWriteMany]
    capacity:
      storage: 1Gi
    persistentVolumeReclaimPolicy: Retain     // [Delete | Recycle]
    hostPath:
      path: /tmp/data
    storageClassName: manual
  //awsElasticBlockStore:
  //  volumeID: <volume-id>
  //  fsType: ext4
  //gcePersistentDisk:
  //  pdName: pd-disk
  //  fsType: ext4

pvc-definition.yml (11)
  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
    name: pvc-vol
  spec:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  //storageClassName: google-storage

sc-definition.yml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass                            // sc creates a pv automatically and binds pvc to that volume
  metadata:
    name: google-storage
  provisioner: kubernetes.io/gce-pd             [kubernetes.io/no-provisioner]
  volumeBindingMode: WaitForFirstConsumer       // This will delay the binding and provisioning of a PV until a Pod using the PVC is created.
  parameters:
    type: pd-standard                           // [pd-ssd]
    replication-type: none                      // [regional-pd]

service-definition.yml (12)
  apiVersion: v1
  kind: Service
  metadata:
    name: my-service
  spec:
  //type: NodePort
    ports:
    - name: http
      targetPort: 80                          // port on the pod
      port: 80                                // port on the service (mandatory)
    //nodePort: 30008
    selector:
      type: front-end
    clusterIP: None         // Headless service

nginx-ingress-definition.yml  (+ configmap + service + serviceaccount)
  apiVersion: extensions/v1beta1
  kind: Deployment
  metadata:
    name: nginx-ingress-controller
  spec:
    replicas: 1
    selector: 
      matchLabels:
        name: nginx-ingress
    template:
      metadata:
        labels:
          name: nginx-ingress
      spec:
        containers:
          - name: nginx-ingress-controller
            image: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.21.0
        args:
          - /nginx-ingress-controller
          - --configmap=$(POD_NAMESPACE)/nginx-configuration                  // (err-log-path, keep-alive, ssl-protocols, session timeout)
        env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
                apiVersion: v1
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        ports:
          - name: http
            containerPort: 80
          - name: https
            containerPort: 443

ingress-definition.yml
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: ingress-wear
  spec:
    backend:
      service
        name: wear-service
        port:
          number: 80

ingress-definition.yml (13)
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: ingress-wear-watch
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: /
      nginx.ingress.kubernetes.io/ssl-redirect: "false"
  spec:
    rules:                                // to route traffic based on URL
    - http:
        paths:
        - path: /wear
          backend:
            service:
              name: wear-service
              port:
                number: 80
        - path: /watch
          backend:
            service:
              name: watch-service
              port:
                number: 80

ingress-definition.yml
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: ingress-wear-watch
  spec:
    rules:                                // to route traffic based on domain name
    - host: wear.my-online-store.com
      http:
        paths:
        - path: /wear
          backend:
            service:
              name: wear-service
              port:
                number: 80
    - host: watch.my-online-store.com
      http:
        paths:
        - path: /watch
          backend:
            service:
              name: watch-service
              port:
                number: 80

netpol-definition.yml (14)
  apiVersion: networking.k8s.io/v1
  kind: NetworkPolicy
  metadata:
    name: db-policy
  spec:
    podSelector:                      // apply policy to db pod, protect db pod and only allow traffic from api-pod
      matchLabels:
        role: db
  //podSelector: {}                  // for all pods in default namespace
    policyTypes:
    - Ingress
    - Egress
    ingress:
    - from:
      - podSelector:
          matchLabels:
            name: api-pod
        namespaceSelector:
          matchLabels:
            name: prod
      ports:
      - protocol: TCP
        port: 3306
    egress:
    - to:
      - ipBlock:
          cidr: 192.168.5.10/32
      ports:
      - protocol: TCP
        port: 80

role-definition.yml
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: developer
  rules:
  - apiGroups: [""]                                         // leave empty for core group
    resources: ["pods"]
    verbs: ["list", "get", "create", "update", "delete"]
  //resourceNames: ["blue", "orange"]                       // allow user to some pods
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["create"]

role-bind-definition.yml
  apiVersion: rbac.authorization.k8s.io/v1
  kind: RoleBinding              // link user to the role
  metadata:
    name: developer-binding
  subjects:                     // specify user details
  - kind: User                                [Group]                 [ServiceAccount]
    name: dev-user                            [system:masters]        [default:dashsa]
    apiGroup: rbac.authorization.k8s.io
  roleRef:
    kind: Role                  // specify role details
    name: developer
    apiGroup: rbac.authorization.k8s.io

crd-definition.yml
  apiVersion: apiextensions.k8s.io/v1
  kind: CustomResourceDefinition
  metadata:
    name: flighttickets.flight.co
  spec:
    scope: Namespaced
    group: flights.com
    names:
      kind: FlightTicket
      singular: flightticket
      plural: flighttickets
      shortNames:
        - ft
    versions:
    - name: v1
      schema:
        openAPIV3Schema:
          ...

----------------------------------------

## CKA: Certified Kubernetes Administrator - mumshad mannambeth

Core Concepts
  - Cluster Architecture
    - etcd
    - kube-api server           // (1) Authenticate user (2) Validate request (3) Retrieve data from etcd cluster (4) Update Etcd (5) Scheduler (6) Kubelet 
    - kube controller manager
    - kube scheduler
    - kubelet
    - kube proxy
  - Pods
  - Replication controller / Replica sets
  - Deployments
  - Services
  - Namespaces
  - Imperative vs Declarative
Scheduling
  - Manual Scheduling
  - Labels & Selectors
  - Taints and Tolerations
  - Node Selectors and Affinity
  - Resource Requirements
  - Daemon Sets (ensures one copy of pod is always present in all nodes)
  - Static Pods (pods created by kubelet without intervention of kube api server, definition files under /etc/kubernetes/manifests, name of pod is appended with node name)
  - Multiple Schedulers
Logging Monitoring
  - Monitor Cluster Components
  - Application Logs
Application Lifecycle Management
  - Rolling updates & Rollbacks
  - Configure Application
    - Commands and Arguments
    - Environment variables
    - ConfigMap
    - Secrets
  - Multi Containers
  - Init Containers
Cluster Maintenance
  - OS Upgrades
  - Cluster Upgrade Process
  - Backup & Restore etcd
Security
  - TLS Basics
    - Symmetric Encryption (uses same key to encrypt and decrypt data, attacker can decrypt data with same key)
    - Asymmetric Encryption (uses pair of keys, data is encrypted with public key at client side and decrypted with private key at server side)
    - Cerificate Authority (trusted entity that issues digital certificates which are used to authenticate users and services)
    - Public Key Infrastructure (PKI is the process of generating, distributing and maintaining digital certificates)
  - TLS in Kubernetes
    - Server (kube-apiserver, etcd, kubelet)
    - Client (admin/kubectl, kube-scheduler, kube-controller-manager, kube-proxy)
  - Certificates API (csr-approving, csr-signing controllers)
  - KubeConfig
  - API Groups (metrics, healthz, version, api, apis, logs)
  - Authorization Mechanisms/Mode
    - Node (kubelet are part of system:nodes group)
    - ABAC (change in policy requires to restart kube-apiserver)
    - RBAC (provide more standard approach to manage access within kuberenetes cluster)
    - Webhook (third party tool)
    - AlwaysAllow (default)
    - AlwaysDeny
  - RBAC
    - Role and RoleBinding
    - ClusterRole and ClusterRoleBinding
  - Service Accounts (used by application to interact with kubernetes cluster)
  - Image Security
  - Security contexts
  - Network Policy
Storage
  - Storage in Docker             // docker uses storage drivers (AUFS, ZFS, Overlay,..) to enable layered architecture
  - Container Interface
    - Container Runtime Interface // standard that defines how k8s communicate with container runtime engine (docker, rkt, cri-o)
    - Container Network Interface // standard that help to extend support to different networking solution (weaveworks, flannel)
    - Container Storage Interface // support multiple storage solution (portworx, Amazon EBS, Dell EMC, GlusterFS)
  - Volumes
  - Persistent Volumes
  - Persistent Volume Claims
  - Storage Class
Networking (Configuring IP addresses, Gateways and Routes, DNS servers)
  - Switching, Routing, Gateway
  - DNS (manage name resolution)
  - CoreDNS (CoreDNS is a flexible, extensible DNS server that can serve as the Kubernetes cluster DNS)
  - Network Namespaces
  - Docker Networking
  - CNI
  - Networking Configuration on Kubernetes
  - POD Networking
  - CNI in Kuberenetes
  - Service Networking
  - Ingress (k8s cluster does not come up with ingress controller by default)
Troubleshooting
  - Application Failure (curl webservice, check web service for Endpoints n TargetPort n Selectors, check web deployment, check db service, check logs of db pod)
  - ControlPlane Failure (check kube-system pod status or services, check logs, correct deployment files)
  - Worker Node Failure (check node status, check kubelet service status and logs)
  - Network Failure (check network plugin, check kube-proxy ds)

----------------------------------------

// Cluster Maintenance
$ kube-controller-manager --pod-eviction-timeout=5m0s ...
$ k drain node-1          // pods are gracefully terminated from the node and recreated on another, node is also cordon (marked as unschedulable)
$ k uncordon node-1
$ k cordon node-2         // marks a node unschedulable, new pods are not scheduled on this node

// Upgrade the current version of kubernetes from 1.30.0 to 1.31.0 [https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/]

// On the controlplane node
$ kubeadm version
$ vim /etc/apt/sources.list.d/kubernetes.list
deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /

// install kubeadm
$ kubectl drain controlplane --ignore-daemonsets
$ apt update
$ apt-cache madison kubeadm               // apt show kubectl -a | grep 1.31
$ apt-get install kubeadm=1.31.0-1.1      [kubelet=1.31.0-1.1 kubectl=1.31.0-1.1]

// upgrade cluster
$ kubeadm upgrade plan v1.31.0
$ kubeadm upgrade apply v1.31.0

// upgrade and restart Kubelet
$ apt-get install kubelet=1.31.0-1.1
$ systemctl daemon-reload
$ systemctl restart kubelet

// schedule controlplane node
$ kubectl uncordon controlplane
$ kubectl describe node controlplane | grep -i taint
$ kubectl taint node controlplane node-role.kubernetes.io/control-plane:NoSchedule-

// now worker node
$ kubectl drain node01 --ignore-daemonsets

// SSH to the node01
$ vim /etc/apt/sources.list.d/kubernetes.list
$ apt update
$ apt-cache madison kubeadm
$ apt-get install kubeadm=1.31.0-1.1
$ kubeadm upgrade node

// upgrade and restart Kubelet
// back on controlplane node
$ kubectl uncordon node01

// Add new node to cluster
$ kubeadm token create --print-join-command             // from control plane node
$ kubeadm join ....                                     // from worker node (If you have troubles, you might need to run kubeadm reset before.)

// etcd
$ apt-get install etcd-client
$ ETCDCTL_API=3 etcdctl snapshot save snapshot.db --endpoints=https://127.0.0.1:2379 --cacert=/etc/etcd/ca.crt --cert=/etc/etcd/server.crt --key=/etc/etcd/server.key
$ ETCDCTL_API=3 etcdctl snapshot status snapshot.db
4d4e953, 7213, 1291, 2.7 MB                                                                   // Hash, Revision, Total Keys, Total Size
$ ETCDCTL_API=3 etcdctl snapshot restore snapshot.db --data-dir /var/lib/etcd-from-backup     // cert details not required, back up all manifest files, change etcd volume hostPath
// only for externally managed etcd
$ chown -R etcd:etcd etcd-from-backup
$ ps -ef | grep etcd
$ vi /etc/systemd/system/etcd.service                                   // change data-dir
$ systemctl daemon-reload && systemctl restart etcd
$ k delete pod kube-controller-manager-.. kube-scheduler-..
$ systemctl restart kubelet                                             // ssh to controlpane node

// Cerificate Authority (Symantec, Digicert, Globalsign, ..)
$ openssl genrsa -out ca.key 2048                                       // generate private key (ca.key)
$ openssl req -new -key ca.key -out ca.csr -subj "/CN=kubernetes-ca"    // generate Certificate Signing Request (ca.csr)
$ openssl x509 -req -in ca.csr -signkey ca.key -out ca.crt              // self-signed CA (ca.crt)

// Admin user
$ openssl genrsa -out admin.key 2048                                          // generate private key (admin.key)
$ openssl req -new -key admin.key -out admin.csr -subj "/CN=kube-admin"       // generate Certificate Signing Request (admin.csr)
$ openssl x509 -req -in admin.csr -CA ca.crt -CAkey ca.key -out admin.crt     // sign the certificate (admin.crt)

// kube-apiserver (communicates with etcd server, kubelet server)
$ openssl genrsa -out apiserver.key 2048
$ cat openssl.cnf
[alt_names]
DNS.1 = kubernetes
DNS.2 = kubernetes.default ....
$ openssl req -new -key apiserver.key -out apiserver.csr -subj "/CN=kube-apiserver" -config openssl.cnf
$ openssl x509 -req -in apiserver.csr -CA ca.crt -CAkey ca.key -out apiserver.crt
$ kubeadm certs check-expiration | grep apiserver             // list the expiration dates
$ kubeadm certs renew apiserver                               // renew the apiserver server certificate

// view cert details (to verify name, alt names, organization, account, issuer, expiry date)
$ openssl x509 -in /etc/kubernetes/pki/apiserver.crt -text -noout
Subject: CN=kube-apiserver                                        // name of certificate (CN)
Subject Alternative Name: DNS:master, DNS: kubernetes, ..         // alternate names
Validity
    Not Before: Feb 11 2024 05:39:00 2025 GMT                     // issued date
    Not After: Feb 11 2025 05:39:00 2025 GMT                      // expiry date
Issuer: CN=kubernetes                                             // CA who issued the certificate

$ openssl x509 -in /etc/kubernetes/pki/etcd/server.crt  -text -noout
Subject: CN=node01                                                // name of certificate (CN)

// view logs
$ journalctl -u etcd.service -l
$ kubectl logs etcd-master
$ docker ps -a && docker logs 87fc
$ crictl ps -a | grep kube-apiserver && crictl logs 87fc
$ crictl inspect 87fc
$ ssh cluster1-node2 'crictl logs b01edbe6f89ed' &> /opt/course/17/pod-container.log

// Kubelet
$ service kubelet status
$ journalctl -u kubelet
$ cat /etc/kubernetes/kubelet.conf                                // kubeconfig file used by the kubelet (details of api-server)
$ cat /var/lib/kubelet/config.yaml                                // kubelet service options
$ ps -ef | grep kubelet | grep container-runtime                  // identify the container runtime via kubelet service
$ find /usr/lib/systemd | grep kube                               // components controlled via systemd
$ whereis kubelet                                                 // kubelet binary path
$ systemctl daemon-reload && service kubelet restart

// kube proxy
$ crictl ps | grep kube-proxy
$ crictl stop 27b6a18c0f89c
$ crictl logs 27b6a18c0f89c
I0913 12:53:03.096620       1 server_others.go:212] Using iptables Proxier.
$ iptables-save | grep p2-service                                 // iptables rules belonging to p2-service

$ cat /etc/kubernetes/manifests/kube-apiserver.yaml
 - --tls-cert-file=/etc/kubernetes/pki/apiserver.crt                              // server certificate file
 - --client-ca-file=/etc/kubernetes/pki/ca.crt                                    // root CA certificate
 - --etcd-certfile=/etc/kubernetes/pki/apiserver-etcd-client.crt                  // etcd client certificate
 - --etcd-cafile=/etc/kubernetes/pki/etcd/server.crt                              // etcd ca file
 - --kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt  // kubelet client certificate
 - --etcd-servers=https://127.0.0.1:2379                                          // address of etcd server
 - --service-cluster-ip-range=10.96.0.0/12                                        // ip range (CIDR) for services

$ cat /etc/kubernetes/manifests/etcd.yaml
 - --cert-file=/etc/kubernetes/pki/etcd/server.crt                                // server certificate file
 - --key-file=/etc/kubernetes/pki/etcd/server.key                                 // server private key
 - --trusted-ca-file=/etc/kubernetes/pki/etcd/ca.crt                              // CA root certificate
 - --client-cert-auth=true                                                        // client certificate authentication

$ cat /etc/kubernetes/manifests/kube-controller-manager.yaml
 - --kubeconfig=/etc/kubernetes/controller-manager.conf                           // kubeconfig file used by the controller manager
 - --cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt                         // to sign certificates
 - --cluster-signing-key-file=/etc/kubernetes/pki/ca.key
 - --service-cluster-ip-range=10.96.0.0/12                                        // ip range (CIDR) for services

$ cat /etc/kubernetes/manifests/kube-scheduler.yaml
    - --kubeconfig=/etc/kubernetes/scheduler.conf                                 // kubeconfig file used by the scheduler
    - --config=/etc/kubernetes/my-scheduler.yaml                                  // in case of custom scheduler

// Certificate API (new user creates a key, generate csr, sends a request to admin, admin creates a csr object)
$ k get csr
$ k certificate approve jane         // k8s signs the certificate using CA key pairs and generates certificate
$ k get csr jane -o yaml             // view the generated certificate (base64 encoded)
$ k certificate deny agent-smith
$ kubectl get csr jane -o jsonpath="{.status.certificate}" | base64 -d > jane.crt

// Assign Roles and Bindings, Generate and Configure a kubeconfig File
$ kubectl config view
$ kubectl config set-cluster kubernetes --server=https://controlplane:6443 --certificate-authority=/etc/kubernetes/pki/ca.crt --embed-certs=true --kubeconfig=jane-config
$ kubectl config set-credentials jane --client-certificate=jane.crt --client-key=jane.key --embed-certs=true --kubeconfig=jane-config
$ kubectl config set-context jane-context --cluster=kubernetes --namespace=default --user=jane --kubeconfig=jane-config
$ kubectl config use-context jane-context --kubeconfig=jane-config

// KubeConfig
$ curl https://my-kube-playground:6443/api/v1/pods --key admin.key --cert admin.crt --cacert ca.crt
$ k get pods --server my-kube-playground:6443 --client-key admin.key --client-certificate admin.crt --certificate-authority ca.crt
$ cat $HOME/.kube/config
apiVersion: v1
kind: Config
current-context: my-kube-playground@my-kube-admin
clusters:
- name: my-kube-playground
  cluster:
    certificate-authority: /etc/kubernetes/pki/ca.crt
  //certificate-authority-data: LS0t.......bnJ
    server: https://my-kube-playground:6443
contexts:
- name: my-kube-playground@my-kube-admin
  context:
    cluster: my-kube-playground
    user: my-kube-admin
    namespace: genctl
users:
- name: my-kube-admin
  user:
   client-certificate: /etc/kubernetes/pki/users/admin.crt
   client-key: /etc/kubernetes/pki/users/admin.key

// API Groups
api (core)
  - v1
    - namespaces, pods, nodes, services, ...
apis (named)
  - apps                                                    // API Groups
    - v1
      - deployments, replicates, statefulsets               // Resources
        - list, get, create, delete, update, watch          // Verbs (actions)
  - extensions
  - networking.k8s.io
    - v1
      - networkpolicies
  - storage.k8s.io
  - authentication.k8s.io
  - certificates.k8s.io
    - v1
      - certificatesigningrequests

$ kubectl proxy
$ curl localhost:8001 -k                                    // list api groups

// Authorization Mode
$ cat /etc/kubernetes/manifests/kube-apiserver.yaml
 - --authorization-mode=Node,RBAC
$ ps -ef | grep authorization-mode

// RBAC
$ k api-resources --namespaced=true                                   // Namespaced
pods, rs, jobs, deploy, svc, secret, cm, role, rolebinding, pvc
$ k api-resources --namespaced=false                                  // Cluster scoped
nodes, pv, clusterrole, clusterrolebinding, csr, namespace

// Image Security
$ docker login private-registry.io
$ docker run private-registry.io/apps/internal-app
$ k create secret docker-registry regcred --docker-server=private-registry.io --docker-username=user --docker-password=pswd --docker-email=pr@org.com

// Storage
$ ls /var/lib/docker                                // docker filesysteem
aufs containers image volumes
$ docker volume create data_volume                  // create folder under /var/lib/docker/volumes  
$ docker run -v data_volume:/var/lib/mysql mysql
$ docker run --mount type=bind, source=/data/mysql, target=/var/lib/mysql mysql   [--volume-driver rexray/ebs]

// Networking
$ ip link                                             // list and modify interfaces (eth0)
$ ip addr                                             // ip addr assigned to interfaces
$ ip address show type bridge                         // list interface/bridge created on host
$ ip route                                            // list routes on host
$ arp                                                 // arp table
$ k exec busybox -- ip route                          // default gateway configured on POD scheduled on node
$ netstat -npl | grep -i scheduler                    // listening port
$ netstat -npa | grep -i etcd                         // client connections
$ ls /opt/cni/bin                                     // binaries of CNI supported plugins
$ ls /etc/cni/net.d                                   // plugin configured with cluster (flannel/weave)

// CoreDNS
kubectl -n kube-system get deployment coredns -o yaml
  allowPrivilegeEscalation: true

// Networking Configuration on Kubernetes (Ports)
Master
 - kube-apiserver 6443
 - kubelet 10250
 - kube-scheduler 10259
 - kube-controller-manager 10257
 - etcd-server 2379
 - etcd-client 2380               // etcd peer to peer connectivity in case of multiple master nodes
Worker
  - kubelet 10250
  - Services 30000-32767

// Networking (Misc)
// Switch creates a network
$ ip addr                                             // to connect to switch, we need interface (eth0)
$ ip addr add 192.168.1.1/24 dev eth0                 // assign IP addr to host
// Routing (router) helps to connect one network to other, Gateway is a door to outside world (internet)
$ ip route                                            // list routing configuration
$ ip route add 192.168.2.0/24 via 192.168.1.1         // configure internal private network (192.168.2.0)
$ ip route add default via 192.168.1.1                // configure default (0.0.0.0) gateway for internet
// forward packets between interfaces
$ echo 1 > /proc/sys/net/ipv4/ip_forward
$ cat /etc/sysctl.conf                                // persist change after reboot
  net.ipv4.ip_forward = 1
// DNS
$ cat /etc/hosts                                      // local
$ cat /etc/resolv.conf
nameserver    192.168.1.100                           // address of DNS server
nameserver    8.8.8.8                                 // public nameserver
search        ibm.com                                 // appends the domain (ping web -> ping web.ibm.com)
$ cat /etc/nsswitch.conf                              // order for name resolution
hosts: files dns

// Docker Networking
1. Create Bridge Network/Interface
$ ip link add docker0 type bridge
$ docker network ls                 // docker creates internal private network called bridge by default
$ ip link                           // on host, network is created by name docker0
$ ip addr                           // 172.17.0.1/24

2. Create Network Namespace
$ ip netns add b3165c               // docker creates network ns for container
$ ip netns

3. Create VETH Pairs (virtual cable)
$ ip link add veth0ns type veth peer name vethbb1cbr

4. Attach vETH to Namespace and Bridge
$ ip link set veth0ns netns b3165c
$ ip link set vethbb1cbr master docker0
$ ip link                           // vethbb1cbr interface attach to docker0
$ ip -n b3165c link                 // veth0ns interface within container ns

5. Assign IP addresses and Add route to Gateway
$ ip -n b3165c addr add 172.17.0.3/16 dev veth0ns
$ ip -n b3165c addr
$ ip -n b3165c route add 172.17.0.1/24 via 172.17.0.3

6. Bring interfaces up
$ ip -n b3165c link set veth0ns up

7. Enable NAT - IP Masquerade
$ iptables -t nat -A POSTROUTING -s 172.17.0.1/24 -j MASQUERADE                       // forward packets via host IP address
$ iptables -t nat -A PREROUTING --dport 80 --to-destination 172.17.0.3:80 -j DNAT     // port mapping
$ iptables -nvL -t nat

// Events
$ k get events -A --sort-by=.metadata.creationTimestamp

// Cheatsheet
$ source <(kubectl completion bash)
$ alias k=kubectl
$ complete -o default -F __start_kubectl k

$ export do="--dry-run=client -o yaml"

$ sed -i 's/kube-contro1ler-manager/kube-controller-manager/g' /etc/kubernetes/manifests/kube-controller-manager.yaml
$ sed -i 's/6433/6443/' .kube/config
$ cat ~/.kube/config | grep current | sed -e "s/current-context: //"

$ cat ~/.vimrc
set ai

-----------------------------------------

pod-bind-definition.yml         // POST request to pod binding API to schedule on node
  apiVersion: v1
  kind: Binding
  metadata:
    name: nginx
  target:
    apiVersion: v1
    kind: Node
    name: node02

ds-definition.yml
  apiVersion: apps/v1
  kind: DaemonSet
  metadata:
    name: myapp-ds
  spec:
    template:
      metadata:
        name: myapp-pod
        labels:
          type: front-end
      spec:
        containers:
          - name: nginx-container
            image: nginx
    selector:
      matchLabels:
        type: front-end

csr-definition.yml
  apiVersion: certificates.k8s.io/v1
  kind: CertificateSigningRequest
  metadata:
    name: jane
  spec:
    request: LS0t.......bnJ                               // cat jane.csr | base64 -w 0  (tr -d "\n")
    signerName: kubernetes.io/kube-apiserver-client
    expirationSeconds: 86400                              // one day
    usages:
    - client auth
    groups:
    - system:masters

img-sec-definition.yml
  apiVersion: v1
  kind: Pod
  metadata:
    name: nginx-pod
  spec:
  //schedulerName: my-scheduler
    containers:
    - name: nginx
      image: private-registry.io/apps/internal-app
    imagePullSecrets:
    - name: regcred

my-scheduler.yml
  apiVersion: kubescheduler.config.k8s.io/v1
  kind: KubeSchedulerConfiguration
  profiles:
    schedulerName: my-scheduler
  leaderElection:                             // in case of multiple master nodes
    leaderElect: true
    resourceNamespace: kube-system
    resourceName: lock-object-my-scheduler



# Kubernetes — Interview Answers
### Configuration → Multi-Container → Observability → Pod Design → State Persistence → Services → Security

---

## Day 1 — Configuration

### Commands and Arguments
`CMD` replaces the entire command when you pass arguments at runtime. `ENTRYPOINT` appends runtime arguments to the defined command, making it the fixed executable. Using both together lets `CMD` act as a default argument that can be overridden — `ENTRYPOINT` stays fixed.

### Environment Variables
Environment variables inject configuration values directly into a container at runtime using `env` or `envFrom` in the pod spec. They can be hardcoded as plain values, sourced from a ConfigMap key, or sourced from a Secret key. This decouples configuration from the container image, following the 12-factor app principle.

### ConfigMap
A ConfigMap stores non-sensitive configuration data as key-value pairs, decoupling config from the container image. It can be consumed as environment variables, command-line arguments, or mounted as files inside a pod. It is not encrypted — never use it for passwords or tokens.

### Secrets
Secrets store sensitive data such as passwords, tokens, and keys in base64-encoded format — note that encoding is not encryption. They can be mounted as volumes or exposed as environment variables, and access is controlled via RBAC. For true security, integrate with external secret managers like HashiCorp Vault or enable etcd encryption at rest.

### Security Context
A Security Context defines privilege and access control settings at the pod or container level — such as running as a non-root user (`runAsUser`), making the filesystem read-only, or granting/dropping Linux capabilities like `NET_ADMIN` or `KILL`. Container-level settings override pod-level settings when both are specified. The full list of Linux capabilities is available at `/usr/include/linux/capability.h`.

### Service Account
A Service Account provides an identity for processes running inside a pod to authenticate with the Kubernetes API server. Each namespace has a `default` service account that is automatically mounted into pods unless explicitly disabled. You should create dedicated service accounts per workload and grant them only the minimum RBAC permissions required.

### Resource Requirements
Resource requests tell the scheduler how much CPU and memory a pod needs to find a suitable node — the pod is guaranteed at least that amount. Resource limits cap the maximum CPU and memory a container can consume — exceeding the memory limit causes the container to be OOMKilled. Setting both requests and limits is critical for cluster stability and preventing noisy-neighbour issues.

### Taints and Tolerations
A taint is applied to a node to repel pods that do not explicitly tolerate it — using `kubectl taint nodes node1 key=value:NoSchedule`. A toleration on a pod allows it to be scheduled onto a tainted node, but it does not guarantee it — it only removes the repulsion. To guarantee a pod lands on a specific node, combine tolerations with node affinity.

Taints are applied to Kubernetes nodes to prevent Pods from being scheduled on them unless the Pod has a matching toleration.
Tolerations are defined on Pods and allow them to run on nodes with matching taints.
Tolerations provide permission, not a guarantee that the Pod will be scheduled on that node.
Taints have three effects: NoSchedule, PreferNoSchedule, and NoExecute.
If specific placement is required, we can combine tolerations with nodeSelector or nodeAffinity.

nodeSelector is a simple mechanism to schedule Pods only on nodes having specific labels. nodeAffinity is a more powerful and flexible version that supports required and preferred rules and more complex label expressions.
Node affinity has two main types: requiredDuringSchedulingIgnoredDuringExecution, which is a hard requirement, and preferredDuringSchedulingIgnoredDuringExecution, which is a soft preference. Required affinity prevents scheduling if no matching node exists, while preferred affinity allows Kubernetes to choose another node if the preferred nodes aren't available.


### Node Selectors and Affinity
`nodeSelector` is the simplest way to constrain a pod to nodes with specific labels — for example, `disktype: ssd`. Node Affinity is the more expressive successor, supporting `requiredDuringScheduling` (hard rule) and `preferredDuringScheduling` (soft rule) with richer operators like `In`, `NotIn`, and `Exists`. Unlike taints which push pods away from nodes, affinity pulls pods toward specific nodes.

---

## Day 2 — Multi-Container Pods

### Ambassador, Adapter, Sidecar
The **Sidecar** pattern extends the main container's functionality — for example, a log shipper running alongside the app. The **Ambassador** pattern proxies communication from the main container to the outside world — for example, routing database connections to the correct environment. The **Adapter** pattern transforms the main container's output into a standardised format — for example, converting proprietary metrics into Prometheus format.

SIDEcar      → Support the application
AMBassador   → Communicate on behalf of application
ADAPter      → Adapt/convert application output

### Init Containers
Init containers run to completion before any app containers start in a pod, making them ideal for setup tasks like pulling secrets, running database migrations, or waiting for a dependency to be ready. They run sequentially — each must succeed before the next begins. If an init container fails, Kubernetes restarts the pod according to the pod's restart policy.

---

## Day 2 — Observability

### Readiness and Liveness Probes
A **Liveness probe** determines if a container is still running — if it fails, Kubernetes restarts the container. A **Readiness probe** determines if a container is ready to serve traffic — if it fails, the pod is removed from the Service's endpoints so no traffic is sent to it. Both can use HTTP GET, TCP socket, or exec command checks, and failing to configure them correctly is a leading cause of rolling update issues.

### Container Logging
Kubernetes captures stdout and stderr from each container, accessible via `kubectl logs <pod-name> -c <container-name>`. Logs are stored on the node and lost when the pod is deleted, so a persistent log aggregation solution like Loki, Fluentd, or Elastic Stack is required for production. For multi-container pods, always specify the `-c` flag to target the correct container.

### Monitoring
The Metrics Server is a lightweight in-cluster aggregator that provides CPU and memory metrics for `kubectl top` and the Horizontal Pod Autoscaler — it does not persist data. Prometheus scrapes and stores time-series metrics from pods and nodes with rich querying via PromQL, and is the industry standard for Kubernetes monitoring. Commercial solutions like Datadog, Dynatrace, and the Elastic Stack provide similar capabilities with additional APM, tracing, and log correlation features.

---

## Day 2 — Pod Design

### Labels, Selectors and Annotations
Labels are key-value pairs attached to Kubernetes objects used to group and select resources — for example, Services use label selectors to find the pods they route traffic to. Selectors filter objects by their labels and are used by controllers, Services, and `kubectl get` commands with `-l`. Annotations also store key-value metadata but are not used for selection — they record informational data like build versions, contact information, or tool-specific configuration.

### Rolling Updates and Rollbacks
A rolling update gradually replaces old pod instances with new ones without downtime, controlled by `maxSurge` (extra pods during update) and `maxUnavailable` (pods allowed to be down during update). You can trigger a rollout with `kubectl set image` and monitor it with `kubectl rollout status`. Rolling back with `kubectl rollout undo` reverts to the previous ReplicaSet — Kubernetes stores rollout history, configurable via `revisionHistoryLimit`.

### Deployment Strategies

| Strategy | Description |
|---|---|
| **Recreate** | Terminates all existing pods before creating new ones, causing downtime. Simple but only appropriate for dev environments or when two versions cannot run simultaneously. |
| **RollingUpdate** (default) | Replaces pods incrementally, ensuring zero downtime. Controlled by `maxSurge` and `maxUnavailable` parameters in the deployment spec. |
| **Blue/Green** | Runs the new version (green) alongside the old (blue) simultaneously, then switches traffic all at once by updating the Service selector. Requires double the resources but enables instant rollback. |
| **Canary** | Routes a small percentage of traffic to the new version by running a small number of new pods alongside the full old deployment. Risk is minimised by testing with real traffic before full rollout — typically implemented with weighted routing via a Service mesh or Ingress controller. |

### Jobs
A Job creates one or more pods and ensures a specified number of them complete successfully — unlike a Deployment, it is designed for finite tasks not long-running services. `completions` sets how many pods must succeed; `parallelism` controls how many run simultaneously. If a pod fails, the Job restarts it until the desired completions are reached or the `backoffLimit` is exceeded.

### CronJobs
A CronJob creates Jobs on a schedule defined using standard cron syntax — for example, `"0 2 * * *"` to run at 2am daily. It manages the lifecycle of Jobs it creates and retains a configurable history of successful and failed jobs via `successfulJobsHistoryLimit` and `failedJobsHistoryLimit`. Be aware of the `concurrencyPolicy` setting — `Forbid` prevents overlapping runs, while `Replace` cancels the running job and starts a new one.

---

## Day 2 — State Persistence

### Volumes
A Volume provides storage that exists for the lifetime of a pod — unlike a container's filesystem which is ephemeral and lost on restart. Volumes are defined in the pod spec and mounted into one or more containers, with types ranging from `emptyDir` (temporary shared storage between containers in a pod) to `hostPath` (mounts a node directory). The volume lifecycle is tied to the pod, not the container — data survives container restarts within the same pod.

### Persistent Volumes
A Persistent Volume (PV) is a cluster-wide storage resource provisioned by an administrator, independent of any individual pod's lifecycle. It abstracts the underlying storage implementation — whether NFS, cloud block storage, or Ceph — behind a common Kubernetes API. PVs are not namespaced and persist beyond pod or even namespace deletion until explicitly reclaimed.

### Persistent Volume Claims
A Persistent Volume Claim (PVC) is a namespaced request by a pod for storage — specifying size, access mode (`ReadWriteOnce`, `ReadWriteMany`, `ReadOnlyMany`), and optionally a storage class. Kubernetes binds the PVC to a suitable PV that satisfies the request, and the pod mounts the PVC like any other volume. If no matching PV exists and dynamic provisioning is configured, a new PV is automatically created.

### Storage Class (Dynamic Provisioning)
A StorageClass defines the provisioner, parameters, and reclaim policy for dynamically creating Persistent Volumes on demand — eliminating the need for administrators to pre-provision PVs manually. When a PVC references a StorageClass, Kubernetes automatically provisions a PV using the class's provisioner (e.g., `kubernetes.io/aws-ebs`, `disk.csi.azure.com`). The `reclaimPolicy` (`Delete` or `Retain`) determines what happens to the underlying storage when the PVC is deleted.

### Stateful Sets
StatefulSets are designed for stateful applications like databases that require stable, unique network identities and ordered deployment. Each pod gets a predictable, persistent name like `mysql-0`, `mysql-1` — not a random hash — and pods are created, scaled, and deleted in order. This guarantees that `mysql-0` is always the primary and is fully running before `mysql-1` (the replica) starts.

### Headless Service
A Headless Service is created by setting `clusterIP: None` — it does not get a cluster IP and does not load-balance traffic. Instead, DNS returns the individual pod IP addresses directly, creating a DNS `A` record for each pod — for example `mysql-0.mysql-service.default.svc.cluster.local`. StatefulSets use headless services so pods can discover and address each other individually, which is essential for leader election and replication in distributed databases.

---

## Day 3 — Services & Networking

### Services
A Service provides a stable network endpoint (IP + DNS name) to access a dynamic set of pods selected by label selectors — abstracting away the fact that pod IPs change on restart. `ClusterIP` (default) exposes the service internally; `NodePort` exposes it on each node's IP at a static port; `LoadBalancer` provisions an external cloud load balancer. Kubernetes uses `kube-proxy` and iptables/IPVS rules on each node to forward traffic from the Service's ClusterIP to the backing pods.

### Ingress
An Ingress is a Kubernetes API object that manages external HTTP/HTTPS access to services using a single external URL, with path-based or host-based routing rules. It requires an Ingress Controller (like nginx, Traefik, or AWS ALB) to actually implement the routing — the Ingress object itself is just configuration. It also centralises TLS termination, so SSL certificates are managed in one place rather than per service.

### Network Policies
A NetworkPolicy acts as a pod-level firewall, defining which pods can communicate with each other and with external endpoints using label selectors and CIDR blocks. `Ingress` rules control incoming traffic to a pod; `Egress` rules control outgoing traffic from a pod. By default all pod-to-pod traffic is allowed — applying a NetworkPolicy makes traffic implicitly denied unless explicitly permitted, so you must define both ingress and egress rules carefully.

---

## Day 3 — Security

### Authentication (Service Accounts)
Kubernetes does not manage human user accounts natively — external identity providers (OIDC, certificates, LDAP) handle human authentication. Service Accounts are the Kubernetes-native identity for pod workloads, with a JWT token automatically mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`. Every API request — from a human or a pod — must authenticate; unauthenticated requests are rejected by the API server.

### Authorization (RBAC)
RBAC (Role-Based Access Control) is the standard authorization mechanism in Kubernetes, controlling who can perform what actions on which resources. It works by binding a subject (user, group, or service account) to a Role or ClusterRole that defines permitted verbs (`get`, `list`, `create`, `delete`) on specific resources. Always follow the principle of least privilege — grant only the minimum permissions required, and prefer namespace-scoped Roles over cluster-scoped ClusterRoles.

### KubeConfig
The kubeconfig file (`~/.kube/config`) stores cluster connection details, credentials, and context definitions — a context is a named combination of cluster, user, and namespace. `kubectl` uses the current context to determine which cluster and identity to use for every command. Multiple clusters and users can coexist in one kubeconfig file, switchable with `kubectl config use-context`.

### Role and RoleBinding (Namespaced)
A Role defines a set of permissions within a specific namespace — for example, allowing `get` and `list` on pods in the `dev` namespace only. A RoleBinding grants those permissions to a subject (user, group, or service account) within that same namespace. Neither a Role nor a RoleBinding has any effect outside the namespace they are created in.

### ClusterRole and ClusterRoleBinding (Cluster Scoped)
A ClusterRole defines permissions that apply across all namespaces or for cluster-scoped resources like Nodes, PersistentVolumes, and Namespaces themselves. A ClusterRoleBinding grants those permissions cluster-wide to a subject — meaning the subject can perform those actions in every namespace. A ClusterRole can also be bound with a namespaced RoleBinding to grant its permissions within a single namespace, which is useful for reusing common permission sets.

### Admission Controllers
Admission controllers are plugins that intercept API server requests after authentication and authorisation but before the object is persisted to etcd — allowing additional validation or mutation. **Mutating** admission controllers can modify the incoming request object — for example, injecting a sidecar container or setting default resource limits. **Validating** admission controllers can only inspect the request and accept or reject it — for example, enforcing that all pods must have resource limits set.

### Custom Resource Definition (CRD)
A CRD extends the Kubernetes API by letting you define your own resource types — for example, `Database`, `ImageArchive`, or `AcadiaPool` — that are stored in etcd and managed via `kubectl` like built-in resources. Once a CRD is registered, users can create Custom Resources (CRs) of that type, and a custom controller watches and acts on them. CRDs are the foundation of the Kubernetes Operator pattern, enabling complex stateful application lifecycle management through declarative configuration.

### Custom Controllers
A custom controller watches Custom Resources (or built-in resources) via the Kubernetes API and continuously reconciles the actual cluster state with the desired state declared in those resources. Built in Go using `client-go`, controllers use shared informers that provide an efficient list-watch mechanism with local caching — avoiding repeated API server calls — and a rate-limited work queue that handles retry with exponential backoff. The core contract is the `Reconcile(ctx, Request) (Result, error)` function — it must be idempotent, meaning running it multiple times on the same state always produces the same outcome.



## 7. KUBERNETES DEBUGGING PLAYBOOK

### Pod Issues Quick Reference
| Symptom | Likely Cause | Key Commands |
|---|---|---|
| `CrashLoopBackOff` | App crash, missing config, OOMKilled | `kubectl logs <pod> --previous` |
| `ContainerCreating` | Image pull fail, PVC issue, CNI init | `kubectl describe pod <pod>` |
| `Pending` | Insufficient resources, taint mismatch, PVC not bound | `kubectl describe pod`, `kubectl describe node` |
| `OOMKilled` | Memory limit too low | `kubectl top pod`, increase limits |
| `ImagePullBackOff` | Wrong image, rate limit, missing pull secret | `kubectl describe pod`, check `imagePullSecrets` |

---

### Service/Network Issues
| Symptom | Likely Cause | Fix |
|---|---|---|
| Service has 0 endpoints | Selector labels don't match pod labels | Fix labels, ensure pods are Ready |
| Ingress 404 | Wrong backend port, IngressClass mismatch | `kubectl describe ingress`, check controller logs |
| DNS fails in pod | CoreDNS issue, wrong dnsPolicy | `kubectl exec -- nslookup`, restart CoreDNS |
| Cross-namespace DNS | Wrong FQDN format | Use `<svc>.<ns>.svc.cluster.local` |
| NetworkPolicy blocking | Overly restrictive rules | `kubectl get networkpolicy`, add allow rules |
| NodePort not accessible | Firewall blocking, kube-proxy issue | Open port in security group, check kube-proxy |

---

### Control Plane / Node Issues
| Symptom | Likely Cause | Fix |
|---|---|---|
| Node `NotReady` | Kubelet down, disk/memory pressure | `systemctl status kubelet`, free disk space |
| API server slow | etcd high latency, API server overloaded | Scale control-plane, optimize etcd IOPS |
| etcd data drift | Corruption, out-of-sync replicas | `etcdctl endpoint status`, restore from snapshot |
| Deployment rollout stuck | Readiness probe failing | `kubectl rollout status`, fix probe config |
| HPA not scaling | Metrics server down, thresholds too high | `kubectl get apiservices \| grep metrics`, check metrics-server |