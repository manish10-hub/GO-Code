# Helm — The Complete Guide
### What It Is, Problems It Solves, Chart Design + Interview Q&A (Beginner → Expert)

---

## Part 1 — What Problem Does Helm Solve?

### The World Without Helm

Imagine deploying a simple web application to Kubernetes.
You need these files:

```
deployment.yaml
service.yaml
configmap.yaml
secret.yaml
ingress.yaml
hpa.yaml
serviceaccount.yaml
networkpolicy.yaml
```

**Problem 1 — Repetition across environments:**

```yaml
# deployment-dev.yaml
image: myapp:1.0.0
replicas: 1
resources:
  limits:
    memory: 128Mi

# deployment-staging.yaml
image: myapp:1.0.0
replicas: 2
resources:
  limits:
    memory: 256Mi

# deployment-prod.yaml
image: myapp:1.0.0
replicas: 5
resources:
  limits:
    memory: 512Mi
```

You maintain 3 copies of almost identical files.
Change the image tag → update 3 files. Miss one → inconsistency.
With 10 services × 3 environments = 30 sets of YAML files.

**Problem 2 — No versioning or rollback:**

```bash
kubectl apply -f deployment.yaml   # deploy v1.0.0
# Something goes wrong...
# How do you rollback? You need to:
# 1. Remember what the previous YAML looked like
# 2. Find the old file in git
# 3. kubectl apply the old file
# 4. Hope nothing else changed
```

**Problem 3 — No install / uninstall as a unit:**

```bash
# To delete the app:
kubectl delete -f deployment.yaml
kubectl delete -f service.yaml
kubectl delete -f configmap.yaml
kubectl delete -f ingress.yaml
# Miss one → orphaned resources left behind
```

**Problem 4 — No dependency management:**

Your app needs PostgreSQL, Redis, and Kafka.
You copy their YAML files into your repo, manually manage versions,
and have no standard way to update them.

**Problem 5 — Sharing and reuse:**

Your team has 5 applications. Each needs nearly identical monitoring
setup (Prometheus ServiceMonitor, Grafana dashboard). You copy-paste
YAML across repos. One update needs to be applied in 5 places.

### What Helm Provides

```
Problem                          Helm Solution
─────────────────────────────────────────────────────────────
Repetition across environments → Templates + values files
No versioning                  → Chart versions + release history
No atomic install/uninstall    → Release management (install/uninstall)
No rollback                    → helm rollback
No dependency management       → Chart dependencies (Chart.yaml)
Hard to share                  → Chart repositories (like npm/pip for K8s)
```

---

## Part 2 — What Is Helm?

Helm is the **package manager for Kubernetes**. Think of it like:
- `apt` / `yum` for Linux packages
- `npm` for Node.js packages
- `pip` for Python packages

But for Kubernetes applications.

### Core Concepts

```
Chart
  A package of pre-configured Kubernetes resources.
  Contains templates + default values.
  Like a blueprint that can be customised.

Release
  A specific instance of a chart installed in a cluster.
  Installing the same chart twice creates two releases.
  Each release has its own name, history, and state.

Repository
  A place where charts are stored and shared.
  Like DockerHub for images, but for Helm charts.
  Examples: Artifact Hub, Bitnami, your own private repo.

Values
  Configuration parameters that customise a chart.
  Override defaults for your specific environment.
  Can be in values.yaml or passed via --set flag.

Revision
  Every install, upgrade, or rollback creates a new revision.
  Helm stores each revision's state in Kubernetes Secrets.
  Enables rollback to any previous revision.
```

### Helm Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Developer Machine                        │
│                                                                 │
│  ┌─────────┐    helm install myapp ./chart                     │
│  │  Helm   │    --values prod-values.yaml                      │
│  │  CLI    │                                                    │
│  └────┬────┘                                                    │
└───────┼─────────────────────────────────────────────────────────┘
        │
        │ 1. Load chart + values
        │ 2. Render templates → plain YAML
        │ 3. Apply to cluster via Kubernetes API
        │ 4. Store release state in Secret
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                           │
│                                                                 │
│  ┌──────────────────┐   ┌──────────────────────────────────┐   │
│  │  API Server      │   │  Release Storage (Secrets)       │   │
│  │  (applies YAML)  │   │  sh.helm.release.v1.myapp.v1     │   │
│  └──────────────────┘   │  sh.helm.release.v1.myapp.v2     │   │
│                          └──────────────────────────────────┘   │
│                                                                 │
│  Created resources:                                             │
│  Deployment, Service, ConfigMap, Ingress...                     │
└─────────────────────────────────────────────────────────────────┘
```

**Key architectural point:** Helm v3 has no server-side component (Tiller
was removed in v3). All processing happens in the Helm CLI. Release state
is stored in Kubernetes Secrets in the same namespace as the release.

---

## Part 3 — Helm Chart Structure

```
myapp/                          ← Chart directory (chart name)
├── Chart.yaml                  ← Chart metadata (name, version, description)
├── values.yaml                 ← Default configuration values
├── values-dev.yaml             ← Environment-specific overrides (optional, not in spec)
├── values-prod.yaml            ← Environment-specific overrides (optional, not in spec)
├── charts/                     ← Dependency charts (subcharts)
│   └── postgresql-12.1.2.tgz  ← Packaged dependency
├── templates/                  ← Kubernetes manifest templates
│   ├── _helpers.tpl            ← Named templates (reusable snippets)
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── hpa.yaml
│   ├── serviceaccount.yaml
│   ├── NOTES.txt               ← Post-install instructions shown to user
│   └── tests/
│       └── test-connection.yaml ← Helm test pods
└── .helmignore                 ← Files to ignore when packaging
```

---

## Part 4 — Designing a Helm Chart Step by Step

We will build a complete, production-quality Helm chart for a web application.

---

### Step 1 — Chart.yaml

```yaml
# myapp/Chart.yaml
apiVersion: v2                    # Helm 3 uses v2

name: myapp                       # chart name (lowercase, no spaces)
description: A production web application with database support
type: application                 # or "library" for shared templates

# Chart version — follows SemVer
# Increment when chart structure changes, not app changes
version: 1.2.0

# Application version — the version of the app being packaged
# Used as the default image tag if not overridden
appVersion: "2.5.1"

keywords:
  - web
  - api
  - backend

maintainers:
  - name: Manish Agarwal
    email: manish@example.com

# Dependencies — other charts this chart requires
dependencies:
  - name: postgresql
    version: "12.x.x"              # range — allows patch updates
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled  # only install if this values key is true
    tags:
      - database

  - name: redis
    version: "17.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
```

---

### Step 2 — values.yaml (the heart of the chart)

```yaml
# myapp/values.yaml
# This file defines ALL configurable parameters with sensible defaults.
# Every value here can be overridden by the user.

# ── Replica count ─────────────────────────────────────────────────────────
replicaCount: 1

# ── Image configuration ───────────────────────────────────────────────────
image:
  repository: registry.company.com/myapp
  pullPolicy: IfNotPresent
  # Overrides Chart.yaml appVersion when set
  tag: ""

imagePullSecrets: []
# - name: registry-credentials

# ── Naming ────────────────────────────────────────────────────────────────
nameOverride: ""          # overrides the chart name in resource names
fullnameOverride: ""      # completely overrides the generated full name

# ── Service Account ───────────────────────────────────────────────────────
serviceAccount:
  create: true
  annotations: {}
  name: ""                # auto-generated if empty

# ── Pod annotations and labels ────────────────────────────────────────────
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"

podLabels: {}

# ── Security contexts ─────────────────────────────────────────────────────
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop: ["ALL"]

# ── Service ───────────────────────────────────────────────────────────────
service:
  type: ClusterIP
  port: 80
  targetPort: 8080
  annotations: {}

# ── Ingress ───────────────────────────────────────────────────────────────
ingress:
  enabled: false
  className: "nginx"
  annotations: {}
    # kubernetes.io/tls-acme: "true"
    # nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: myapp.example.com
      paths:
        - path: /
          pathType: Prefix
  tls: []
  # - secretName: myapp-tls
  #   hosts:
  #     - myapp.example.com

# ── Resources ─────────────────────────────────────────────────────────────
resources:
  limits:
    cpu: 500m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

# ── Autoscaling ───────────────────────────────────────────────────────────
autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

# ── Probes ────────────────────────────────────────────────────────────────
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 15
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3

# ── Application configuration ─────────────────────────────────────────────
config:
  logLevel: info
  environment: production
  databaseUrl: ""            # set per-environment, or use external secret
  cacheUrl: ""

# ── Environment variables ─────────────────────────────────────────────────
env: []
# - name: MY_CUSTOM_VAR
#   value: "my-value"
# - name: SECRET_VALUE
#   valueFrom:
#     secretKeyRef:
#       name: my-secret
#       key: value

# ── Volumes ───────────────────────────────────────────────────────────────
volumes: []
volumeMounts: []

# ── Node placement ────────────────────────────────────────────────────────
nodeSelector: {}
tolerations: []
affinity: {}

# ── Pod Disruption Budget ─────────────────────────────────────────────────
podDisruptionBudget:
  enabled: false
  minAvailable: 1

# ── PostgreSQL (dependency) ───────────────────────────────────────────────
postgresql:
  enabled: false              # set to true to deploy PostgreSQL subchart
  auth:
    username: myapp
    password: ""              # generate or set per-environment
    database: myapp
  primary:
    persistence:
      size: 10Gi

# ── Redis (dependency) ────────────────────────────────────────────────────
redis:
  enabled: false
  auth:
    enabled: true
    password: ""
```

---

### Step 3 — _helpers.tpl (reusable template snippets)

```yaml
# myapp/templates/_helpers.tpl
# Files starting with _ are not rendered as Kubernetes manifests.
# They define named templates used by other templates.

{{/*
Expand the name of the chart.
*/}}
{{- define "myapp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited.
If release name contains chart name it will be used as a full name.
*/}}
{{- define "myapp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart label — includes chart name and version.
*/}}
{{- define "myapp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels — applied to ALL resources for consistent selection.
*/}}
{{- define "myapp.labels" -}}
helm.sh/chart: {{ include "myapp.chart" . }}
{{ include "myapp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — used for Service selectors and Deployment selectors.
These must be IMMUTABLE after first deploy — changing them breaks the Service.
*/}}
{{- define "myapp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "myapp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name — either the overridden name or auto-generated.
*/}}
{{- define "myapp.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "myapp.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Image reference — combines repository and tag.
Uses .Chart.AppVersion as default tag if image.tag is empty.
*/}}
{{- define "myapp.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}
```

---

### Step 4 — deployment.yaml template

```yaml
# myapp/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
  annotations:
    # Trigger rolling restart when ConfigMap changes
    checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "myapp.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        {{- with .Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      labels:
        {{- include "myapp.labels" . | nindent 8 }}
        {{- with .Values.podLabels }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "myapp.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
      - name: {{ .Chart.Name }}
        securityContext:
          {{- toYaml .Values.securityContext | nindent 10 }}
        image: {{ include "myapp.image" . }}
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: http
          containerPort: {{ .Values.service.targetPort }}
          protocol: TCP
        env:
        - name: LOG_LEVEL
          value: {{ .Values.config.logLevel | quote }}
        - name: ENVIRONMENT
          value: {{ .Values.config.environment | quote }}
        {{- if .Values.config.databaseUrl }}
        - name: DATABASE_URL
          value: {{ .Values.config.databaseUrl | quote }}
        {{- end }}
        {{- with .Values.env }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
        envFrom:
        - configMapRef:
            name: {{ include "myapp.fullname" . }}-config
        livenessProbe:
          {{- toYaml .Values.livenessProbe | nindent 10 }}
        readinessProbe:
          {{- toYaml .Values.readinessProbe | nindent 10 }}
        resources:
          {{- toYaml .Values.resources | nindent 10 }}
        {{- with .Values.volumeMounts }}
        volumeMounts:
          {{- toYaml . | nindent 10 }}
        {{- end }}
      {{- with .Values.volumes }}
      volumes:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
```

---

### Step 5 — service.yaml template

```yaml
# myapp/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "myapp.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
  {{- with .Values.service.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  type: {{ .Values.service.type }}
  ports:
  - port: {{ .Values.service.port }}
    targetPort: http               # references the named port in deployment
    protocol: TCP
    name: http
  selector:
    {{- include "myapp.selectorLabels" . | nindent 4 }}
```

---

### Step 6 — ingress.yaml template (conditional)

```yaml
# myapp/templates/ingress.yaml
{{- if .Values.ingress.enabled -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "myapp.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
  {{- with .Values.ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if .Values.ingress.className }}
  ingressClassName: {{ .Values.ingress.className }}
  {{- end }}
  {{- if .Values.ingress.tls }}
  tls:
    {{- range .Values.ingress.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ . | quote }}
        {{- end }}
      secretName: {{ .secretName }}
    {{- end }}
  {{- end }}
  rules:
    {{- range .Values.ingress.hosts }}
    - host: {{ .host | quote }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            pathType: {{ .pathType }}
            backend:
              service:
                name: {{ include "myapp.fullname" $ }}
                port:
                  name: http
          {{- end }}
    {{- end }}
{{- end }}
```

---

### Step 7 — hpa.yaml template (conditional)

```yaml
# myapp/templates/hpa.yaml
{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "myapp.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "myapp.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
  {{- if .Values.autoscaling.targetCPUUtilizationPercentage }}
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
  {{- end }}
  {{- if .Values.autoscaling.targetMemoryUtilizationPercentage }}
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
  {{- end }}
{{- end }}
```

---

### Step 8 — NOTES.txt

```
# myapp/templates/NOTES.txt
# This file is printed to the terminal after every install/upgrade.
# Use it to give users next steps.

Thank you for installing {{ .Chart.Name }} version {{ .Chart.Version }}!

Application: {{ .Chart.AppVersion }}
Release:     {{ .Release.Name }}
Namespace:   {{ .Release.Namespace }}

{{- if .Values.ingress.enabled }}
Your application is accessible at:
{{- range .Values.ingress.hosts }}
  http{{ if $.Values.ingress.tls }}s{{ end }}://{{ .host }}
{{- end }}
{{- else if contains "NodePort" .Values.service.type }}
  export NODE_PORT=$(kubectl get svc {{ include "myapp.fullname" . }} -n {{ .Release.Namespace }} -o jsonpath="{.spec.ports[0].nodePort}")
  export NODE_IP=$(kubectl get nodes -o jsonpath="{.items[0].status.addresses[0].address}")
  echo http://$NODE_IP:$NODE_PORT
{{- else if contains "LoadBalancer" .Values.service.type }}
  NOTE: It may take a few minutes for the LoadBalancer IP to be assigned.
  export SERVICE_IP=$(kubectl get svc {{ include "myapp.fullname" . }} -n {{ .Release.Namespace }} -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo http://$SERVICE_IP:{{ .Values.service.port }}
{{- else if contains "ClusterIP" .Values.service.type }}
  kubectl port-forward svc/{{ include "myapp.fullname" . }} 8080:{{ .Values.service.port }} -n {{ .Release.Namespace }}
  Then visit: http://localhost:8080
{{- end }}

To check the deployment status:
  kubectl get pods -l "app.kubernetes.io/name={{ include "myapp.name" . }}" -n {{ .Release.Namespace }}
```

---

### Step 9 — Environment-specific values files

```yaml
# values-dev.yaml
replicaCount: 1
image:
  tag: "2.5.1-dev"
config:
  logLevel: debug
  environment: development
resources:
  limits:
    cpu: 200m
    memory: 128Mi
ingress:
  enabled: true
  hosts:
    - host: myapp.dev.company.com
      paths:
        - path: /
          pathType: Prefix
```

```yaml
# values-prod.yaml
replicaCount: 3
image:
  tag: "2.5.1"
config:
  logLevel: info
  environment: production
resources:
  limits:
    cpu: 1000m
    memory: 512Mi
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: myapp.company.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: myapp-prod-tls
      hosts:
        - myapp.company.com
podDisruptionBudget:
  enabled: true
  minAvailable: 2
```

---

### Step 10 — Common Helm Commands

```bash
# ── Development workflow ───────────────────────────────────────────────────

# Render templates locally (dry run, see the generated YAML)
helm template myapp ./myapp

# Render with specific values
helm template myapp ./myapp -f values-prod.yaml

# Render a specific template
helm template myapp ./myapp -f values-prod.yaml -s templates/deployment.yaml

# Validate chart structure
helm lint ./myapp
helm lint ./myapp -f values-prod.yaml    # lint with specific values

# Dry run against actual cluster (validates against cluster API)
helm install myapp ./myapp --dry-run --debug -f values-prod.yaml

# ── Installation ──────────────────────────────────────────────────────────

# Install (first time)
helm install myapp ./myapp -n production -f values-prod.yaml

# Install and wait until all pods are ready
helm install myapp ./myapp -n production -f values-prod.yaml --wait

# Install with timeout
helm install myapp ./myapp -n production -f values-prod.yaml --wait --timeout 5m

# Install from repository
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install my-postgres bitnami/postgresql --version 12.1.2

# ── Upgrade ───────────────────────────────────────────────────────────────

# Upgrade (subsequent deploys)
helm upgrade myapp ./myapp -n production -f values-prod.yaml

# Install if not exists, upgrade if exists
helm upgrade --install myapp ./myapp -n production -f values-prod.yaml

# Override specific values at command line
helm upgrade myapp ./myapp -f values-prod.yaml --set image.tag=2.5.2

# ── History and rollback ──────────────────────────────────────────────────

# View release history
helm history myapp -n production
# REVISION  STATUS     CHART          DESCRIPTION
# 1         superseded myapp-1.2.0    Install complete
# 2         superseded myapp-1.2.0    Upgrade complete
# 3         deployed   myapp-1.2.1    Upgrade complete

# Rollback to previous revision
helm rollback myapp -n production

# Rollback to specific revision
helm rollback myapp 1 -n production

# ── Inspection ────────────────────────────────────────────────────────────

# List all releases
helm list -n production
helm list -A                             # all namespaces

# Get values used in current release
helm get values myapp -n production

# Get all values (including defaults)
helm get values myapp -n production --all

# Get the manifest that was applied
helm get manifest myapp -n production

# Get release status
helm status myapp -n production

# ── Uninstall ─────────────────────────────────────────────────────────────

# Uninstall (deletes all resources AND release history)
helm uninstall myapp -n production

# Keep history after uninstall (for audit)
helm uninstall myapp -n production --keep-history

# ── Dependencies ──────────────────────────────────────────────────────────

# Download dependencies into charts/
helm dependency update ./myapp

# List dependencies
helm dependency list ./myapp

# ── Packaging and publishing ───────────────────────────────────────────────

# Package chart into .tgz
helm package ./myapp

# Push to OCI registry (Helm 3.8+)
helm push myapp-1.2.0.tgz oci://registry.company.com/helm-charts

# Install from OCI registry
helm install myapp oci://registry.company.com/helm-charts/myapp --version 1.2.0
```

---

## Part 5 — Template Language Reference

```yaml
# ── Variables ─────────────────────────────────────────────────────────────
{{ .Values.replicaCount }}          # from values.yaml
{{ .Chart.Name }}                    # from Chart.yaml: name field
{{ .Chart.Version }}                 # from Chart.yaml: version field
{{ .Chart.AppVersion }}              # from Chart.yaml: appVersion field
{{ .Release.Name }}                  # helm release name (given at install)
{{ .Release.Namespace }}             # namespace where installed
{{ .Release.IsInstall }}             # true on first install
{{ .Release.IsUpgrade }}             # true on upgrade
{{ .Release.Revision }}              # revision number (1, 2, 3...)

# ── Control flow ──────────────────────────────────────────────────────────
{{- if .Values.ingress.enabled }}
  # rendered only if ingress.enabled is true
{{- end }}

{{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
{{- end }}

{{- if eq .Values.service.type "ClusterIP" }}
  # only for ClusterIP services
{{- else if eq .Values.service.type "NodePort" }}
  # only for NodePort services
{{- end }}

# ── Loops ─────────────────────────────────────────────────────────────────
{{- range .Values.ingress.hosts }}
- host: {{ .host }}
{{- end }}

{{- range $key, $value := .Values.podAnnotations }}
  {{ $key }}: {{ $value | quote }}
{{- end }}

# ── String functions ──────────────────────────────────────────────────────
{{ .Values.name | upper }}           # MYAPP
{{ .Values.name | lower }}           # myapp
{{ .Values.name | title }}           # Myapp
{{ .Values.name | quote }}           # "myapp" (adds quotes)
{{ .Values.name | trunc 63 }}        # truncate to 63 chars
{{ .Values.name | trimSuffix "-" }}  # remove trailing dash
{{ printf "%s-%s" "foo" "bar" }}     # foo-bar (string formatting)

# ── Default values ────────────────────────────────────────────────────────
{{ .Values.image.tag | default .Chart.AppVersion }}    # use AppVersion if tag empty
{{ .Values.nameOverride | default .Chart.Name }}

# ── Type conversion ───────────────────────────────────────────────────────
{{ .Values.port | int }}             # convert to integer
{{ .Values.enabled | toString }}     # convert to string
{{ .Values.resources | toYaml }}     # convert map to YAML string

# ── Indentation ───────────────────────────────────────────────────────────
{{ .Values.resources | toYaml | nindent 10 }}  # indent 10 spaces, add leading newline
{{ .Values.resources | toYaml | indent 10 }}   # indent 10 spaces, no leading newline

# ── Whitespace control ────────────────────────────────────────────────────
{{- }}   # strip whitespace BEFORE the tag
{{- }}   # strip whitespace AFTER the tag
# Most template lines should use {{- and -}} to avoid extra blank lines

# ── Checksums (trigger pod restart on ConfigMap change) ───────────────────
annotations:
  checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
# When ConfigMap content changes → checksum changes → annotation changes → rolling restart
```

---

## Interview Questions — Helm

---

## Beginner Level

---

### Q1. What is Helm and what problems does it solve?

Helm is the package manager for Kubernetes — it packages, versions, and
deploys Kubernetes applications as reusable units called charts. It solves
five main problems:

First, **configuration management** — instead of maintaining separate YAML
files per environment, Helm uses templates with values files, so one chart
serves dev, staging, and prod with different values.

Second, **release management** — Helm tracks every install and upgrade as
a versioned release and stores history, enabling one-command rollback.

Third, **atomic operations** — `helm install` and `helm uninstall` treat
all related resources as one unit, so you never have partial installs or
orphaned resources.

Fourth, **dependency management** — a chart can declare dependencies on
other charts (like a Bitnami PostgreSQL chart), which are automatically
downloaded and installed.

Fifth, **sharing and reuse** — charts can be published to repositories
and shared across teams or publicly, similar to npm packages or pip packages.

---

### Q2. What are the three main Helm concepts?

**Chart:** the package itself — a directory of templates, a values file,
and metadata. Like an RPM or npm package.

**Release:** a specific running instance of a chart in a cluster. Installing
the same chart twice (with different release names) creates two independent
releases — two complete sets of resources with no shared state.

**Repository:** a storage location for charts, similar to DockerHub for images.
Charts are published to repositories (Artifact Hub, Bitnami, your own private
OCI registry) and installed from there.

---

### Q3. What is the difference between `helm install` and `helm upgrade`?

`helm install` creates a brand new release — it fails if a release with that
name already exists. It creates all the Kubernetes resources defined in the
chart and records the release as revision 1.

`helm upgrade` updates an existing release with a new version of the chart or
new values. It calculates the diff between the current and desired state, applies
changes, and records the new revision in the release history.

`helm upgrade --install` (most common in CI/CD) combines both — installs if the
release does not exist, upgrades if it does. This is idempotent and the standard
pattern for automated deployments.

---

### Q4. How do you override values in Helm?

Three ways, in increasing priority (later overrides earlier):

```bash
# 1. Default values (lowest priority)
#    Defined in chart's values.yaml

# 2. Values file
helm install myapp ./myapp -f values-prod.yaml

# 3. Multiple values files (merged left to right)
helm install myapp ./myapp -f values-common.yaml -f values-prod.yaml

# 4. --set flag (highest priority — overrides everything else)
helm install myapp ./myapp -f values-prod.yaml --set image.tag=2.5.2

# 5. --set-string (forces value to be treated as string)
helm install myapp ./myapp --set-string config.port=8080

# 6. --set-file (reads value from file)
helm install myapp ./myapp --set-file config.tlsCert=./cert.pem
```

Priority order: `--set` > `-f values-file.yaml` > `values.yaml` (chart defaults).

---

### Q5. How do you roll back a Helm release?

```bash
# View release history
helm history myapp -n production
# REVISION  STATUS
# 1         superseded   (original install)
# 2         deployed     (current upgrade — but it is broken)

# Rollback to previous revision (revision 1)
helm rollback myapp -n production

# Rollback to a specific revision
helm rollback myapp 1 -n production

# Helm applies the manifests from revision 1 back to the cluster
# and creates a new revision 3 that is identical to revision 1
# REVISION  STATUS
# 1         superseded
# 2         superseded
# 3         deployed   (rollback of 2 to state of 1)
```

Rollback does NOT delete revision 2 — it creates a new revision that
restores the old state. The full history is preserved.

---

## Intermediate Level

---

### Q6. What is _helpers.tpl and why is it important?

`_helpers.tpl` is a special file in the `templates/` directory where you
define **named templates** — reusable snippets of template code that can
be called from any other template using `{{ include "name" . }}`.

Files starting with `_` are not rendered as Kubernetes manifests — they
are only for defining named templates. This is important because without
it, Kubernetes would try to apply the raw template content as a manifest
and fail.

Common things defined in `_helpers.tpl`:
- `myapp.fullname` — generates a consistent resource name (release-name + chart-name)
- `myapp.labels` — standard Kubernetes labels applied to every resource
- `myapp.selectorLabels` — the stable labels used for Service-to-Pod selection
- `myapp.image` — combines image repository and tag

Using `_helpers.tpl` ensures that if you change the naming convention,
you change it in one place and all templates automatically pick up the change.

---

### Q7. Explain the `checksum/config` annotation pattern in Helm. Why is it needed?

```yaml
# In deployment.yaml:
annotations:
  checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
```

**The problem it solves:**

When you change a ConfigMap (e.g., update a configuration value), Kubernetes
does not automatically restart the pods that use it. The pods keep running
with the old config until they restart for another reason. This means a
`helm upgrade` that only changes a ConfigMap value would not trigger the
pods to reload the new config — a silent inconsistency.

**How the pattern works:**

The annotation computes a SHA256 hash of the ConfigMap template content.
When the ConfigMap values change, the hash changes. Changing a pod annotation
causes Kubernetes to trigger a rolling restart of the pods. So:
ConfigMap changes → hash changes → annotation changes → rolling restart → pods
load new config.

This is a Helm idiom — not a Kubernetes feature. It is widely used in
production charts (including the official Bitnami charts) because ConfigMap
changes without pod restarts are a common source of "why isn't my config
taking effect?" bugs.

---

### Q8. How do Helm dependencies work? How do you add PostgreSQL as a dependency?

Dependencies are declared in `Chart.yaml`:

```yaml
dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled   # only install if this value is true
```

**Workflow:**

```bash
# 1. Download dependencies into charts/ directory
helm dependency update ./myapp
# Downloads postgresql-12.x.x.tgz into charts/

# 2. Install — Helm installs the parent chart AND all dependencies
helm install myapp ./myapp --set postgresql.enabled=true

# 3. Configure the dependency via values
# In your values.yaml:
postgresql:
  enabled: true
  auth:
    username: myapp
    database: myapp
```

**The `condition` field:**
When `postgresql.enabled: false`, Helm skips installing the PostgreSQL
subchart entirely. This lets you use an external PostgreSQL in production
(where you would not want a chart-managed DB) while using the subchart
locally for development.

**`tags` field:**
An alternative to `condition` — you can group multiple dependencies
under a tag and enable/disable them together with `--set tags.database=false`.

---

### Q9. What is `helm template` and when do you use it?

`helm template` renders chart templates locally and prints the output to
stdout — without connecting to a Kubernetes cluster. No resources are created.

```bash
helm template myapp ./myapp -f values-prod.yaml
```

**When to use it:**

1. **Debugging templates** — see exactly what YAML Helm generates before applying
2. **CI/CD pipeline validation** — render and inspect the manifests in CI without a cluster
3. **GitOps with ArgoCD/Flux** — some GitOps tools use `helm template` output as plain
   YAML stored in git, rather than running Helm directly
4. **Piping to other tools** — render Helm output and pipe to `kubectl apply`:
   `helm template myapp ./chart | kubectl apply -f -`
5. **Checking a specific template** — `helm template myapp ./chart -s templates/deployment.yaml`

The difference from `--dry-run`: `helm template` works completely offline and
does not validate against the cluster API. `helm install --dry-run` connects
to the cluster and validates the manifests against the cluster's API server.

---

### Q10. What is the difference between `helm lint` and `helm template`?

**`helm lint`:**
Static analysis of the chart structure and values. Checks for:
- Required fields in Chart.yaml
- Template syntax errors
- Known bad practices (deprecated APIs, missing required values)
- YAML structure validity

```bash
helm lint ./myapp
helm lint ./myapp -f values-prod.yaml   # lint with specific values
# Output: [INFO] Chart.yaml: icon is recommended
# Output: Error: templates/deployment.yaml: rendering failed...
```

**`helm template`:**
Actually renders all templates to YAML output. Useful for:
- Seeing exactly what Kubernetes manifests will be created
- Debugging template logic (variables, conditionals, loops)
- Piping output to other tools

**Best practice in CI/CD:**
```bash
helm lint ./myapp -f values-prod.yaml    # 1. syntax check
helm template myapp ./myapp \
  -f values-prod.yaml | \
  kubectl --dry-run=client -f -          # 2. validate against cluster API
helm upgrade --install myapp ./myapp \
  -f values-prod.yaml --atomic           # 3. deploy with atomic flag
```

---

## Advanced Level

---

### Q11. What is `--atomic` in Helm upgrade and why should you use it in CI/CD?

```bash
helm upgrade --install myapp ./myapp -f values.yaml --atomic --timeout 5m
```

`--atomic` causes Helm to:
1. Apply the new manifests
2. **Wait** for all pods to become ready (within the timeout)
3. If anything fails within the timeout → **automatically rollback** to the previous revision
4. Exit with a non-zero status code (failing the CI/CD pipeline)

Without `--atomic`:
- Helm applies manifests and exits immediately
- If pods crash, the CI/CD pipeline shows "success" (Helm returned 0)
- The cluster is in a broken state but CI/CD does not know
- No automatic rollback — someone must manually intervene

**Why it matters:**
In a CI/CD pipeline, you want deploy failures to be loud and to automatically
restore a known-good state. `--atomic` gives you both — it fails the pipeline
AND rolls back the cluster, leaving the cluster running the previous working version.

**The `--timeout` flag:**
Set this to longer than your pod startup time (including init containers,
image pull time, readiness probe delays). If your pods take 2 minutes to
start, set `--timeout 5m` to give headroom.

---

### Q12. How does Helm store release state? What happens if the Secret is deleted?

Helm v3 stores release metadata in **Kubernetes Secrets** in the same namespace
as the release. Each revision creates one Secret:

```
Secret name: sh.helm.release.v1.<release-name>.v<revision>
Namespace:   same as release namespace

Contents (base64 + gzip encoded):
  - The rendered manifests
  - The values used
  - Release metadata (name, status, chart info, timestamps)
```

```bash
kubectl get secrets -n production | grep helm
# sh.helm.release.v1.myapp.v1   helm.sh/release.v1   1      5d
# sh.helm.release.v1.myapp.v2   helm.sh/release.v1   1      2d
# sh.helm.release.v1.myapp.v3   helm.sh/release.v1   1      1h
```

**If the Secrets are deleted:**
- `helm list` shows no releases
- `helm history` shows nothing
- `helm rollback` is impossible (no history)
- BUT the actual Kubernetes resources (Deployment, Service, etc.) still exist!
- Helm has "forgotten" about them — it does not know it manages them

**Recovery:** use `helm upgrade --install` which will create a new revision 1,
adopting the existing resources. Or delete the orphaned resources and reinstall.

**Why Secrets not ConfigMaps:**
Release manifests may contain Secrets (passwords, TLS certificates) baked
into values. Storing in Kubernetes Secrets means they are encrypted at rest
if etcd encryption is enabled. ConfigMaps are not encrypted.

---

### Q13. What is a Helm hook and how does it differ from an init container?

A Helm hook is a Kubernetes resource (usually a Job or Pod) that Helm runs
at specific points in the release lifecycle — before install, after install,
before upgrade, before delete, etc. — rather than as part of the regular
chart resources.

```yaml
# templates/db-migrate-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "myapp.fullname" . }}-db-migrate
  annotations:
    # This annotation makes it a hook — not a regular chart resource
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "-5"       # run before other pre-upgrade hooks
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migrate
        image: {{ include "myapp.image" . }}
        command: ["./migrate", "--up"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: {{ include "myapp.fullname" . }}-db-secret
              key: url
```

**Available hook annotations:**
- `pre-install` — runs before any chart resources are created
- `post-install` — runs after all chart resources are created
- `pre-upgrade` — runs before upgrading chart resources
- `post-upgrade` — runs after upgrading chart resources
- `pre-delete` — runs before deleting chart resources
- `post-delete` — runs after deleting chart resources
- `pre-rollback` / `post-rollback` — around rollback operations
- `test` — runs when `helm test` is called

**Comparison with init containers:**

| | Helm Hook (Job) | Init Container |
|---|---|---|
| **When it runs** | At Helm lifecycle events (upgrade, install) | Every time the pod starts |
| **Purpose** | One-time operations (DB migration, cert generation) | Per-pod setup (wait for dependencies) |
| **Runs even if** | Chart upgrade doesn't change that resource | Pod restarts for any reason |
| **Scope** | Can interact with external systems | Runs inside the pod's environment |
| **State** | Job history persists until deleted | Ephemeral — no record after completion |
| **Failure handling** | Helm waits, can rollback on hook failure | Pod restart policy applies |

**Hook delete policy:**
```yaml
"helm.sh/hook-delete-policy": before-hook-creation  # delete before running again
"helm.sh/hook-delete-policy": hook-succeeded         # delete after success
"helm.sh/hook-delete-policy": hook-failed            # delete after failure
```

---

### Q14. What is a Library Chart and how is it different from an Application Chart?

```yaml
# Chart.yaml
type: library   # or type: application (default)
```

**Application chart (`type: application`):**
- Can be installed directly: `helm install myapp ./myapp`
- Creates Kubernetes resources when installed
- Is what most people mean by "a Helm chart"

**Library chart (`type: library`):**
- **Cannot** be installed directly — `helm install` fails
- Contains only named templates (like `_helpers.tpl`) in the `templates/` directory
- No Kubernetes resources are created
- Must be used as a dependency by other charts
- Solves the problem of sharing template logic across multiple charts

**Example use case:**
Your company has 10 microservices, each with a Helm chart. Every chart needs
the same security context, standard labels, and naming conventions. Without
a library chart, you copy `_helpers.tpl` into each chart and maintain 10 copies.

With a library chart:
```yaml
# shared-templates (library chart) _helpers.tpl defines:
# company.labels, company.securityContext, company.fullname

# Each service's Chart.yaml:
dependencies:
  - name: shared-templates
    version: "1.0.0"
    repository: oci://registry.company.com/helm-charts

# Each service's deployment.yaml:
labels:
  {{- include "company.labels" . | nindent 4 }}   # from library chart
```

Change the standard labels in one library chart, all 10 charts update on
their next dependency update.

---

### Q15. How do you handle Secrets in Helm charts securely? What are the options?

Helm values and release state are stored in Kubernetes Secrets, but they
are only base64-encoded — not encrypted. Putting actual secrets (passwords,
API keys) in `values.yaml` or passing them via `--set password=abc123`
stores them in git or shell history.

**Option 1 — Helm Secrets plugin (helm-secrets):**
```bash
# Uses SOPS (Secrets OPerationS) to encrypt values files
# Supports AWS KMS, GCP KMS, Azure Key Vault, PGP

# Encrypt a values file
helm secrets enc secrets-prod.yaml

# Install using encrypted file (plugin decrypts on the fly)
helm secrets install myapp ./myapp -f secrets-prod.yaml
```

**Option 2 — External Secrets Operator:**
Do not put secrets in Helm at all. Reference secrets stored in a secret
manager (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager):
```yaml
# Helm creates an ExternalSecret CR (not the actual Secret)
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
spec:
  secretStoreRef:
    name: aws-secrets-manager
  target:
    name: myapp-db-password
  data:
  - secretKey: password
    remoteRef:
      key: prod/myapp/database
      property: password
# External Secrets Operator creates the actual Kubernetes Secret
```

**Option 3 — Reference pre-existing Secrets:**
```yaml
# values.yaml — reference an existing Secret by name, not value
existingSecret: myapp-db-credentials
existingSecretKey: password

# deployment.yaml template:
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.existingSecret }}
      key: {{ .Values.existingSecretKey }}
```
The Secret is created outside Helm (by a separate process or Vault agent).
Helm only references it by name — the actual password never touches Helm.

**Option 4 — Sealed Secrets (Bitnami):**
Encrypt secrets with the cluster's public key — only the cluster can decrypt.
The encrypted SealedSecret YAML can be safely committed to git.

**Best practice recommendation:**
Use External Secrets Operator (or Vault Agent Injector) for production —
secrets live in a dedicated secret manager with access control and audit logs,
and Helm never sees the actual secret values.

---

## Quick Reference

### Helm Built-in Objects

```
.Values         → chart values (from values.yaml and overrides)
.Chart          → Chart.yaml contents (.Chart.Name, .Chart.Version, .Chart.AppVersion)
.Release        → release info (.Release.Name, .Release.Namespace, .Release.Revision)
.Template       → current template info (.Template.Name, .Template.BasePath)
.Files          → access non-template files in chart (.Files.Get "config.json")
.Capabilities   → cluster capabilities (.Capabilities.KubeVersion.Major)
```

### Common Template Functions

```
default         → {{ .Values.tag | default "latest" }}
quote           → {{ .Values.name | quote }}    → "myapp"
upper/lower     → {{ .Values.name | upper }}    → MYAPP
trunc           → {{ .Values.name | trunc 63 }} → truncate string
trimSuffix      → {{ .Values.name | trimSuffix "-" }}
replace         → {{ .Values.name | replace "+" "_" }}
toYaml          → {{ .Values.resources | toYaml | nindent 6 }}
toJson          → {{ .Values.config | toJson }}
fromYaml        → parse YAML string into object
sha256sum       → {{ .Values.config | sha256sum }}  → checksum
b64enc/b64dec   → base64 encode/decode
include         → {{ include "myapp.fullname" . }}  → call named template
tpl             → {{ tpl .Values.configTemplate . }} → render string as template
```

### Helm Chart Best Practices

```
✓ Always use named templates in _helpers.tpl for names and labels
✓ Apply standard labels to ALL resources (app.kubernetes.io/*)
✓ Make resources conditional with {{- if .Values.feature.enabled }}
✓ Set sensible defaults in values.yaml — chart should work with zero overrides
✓ Add checksum annotation to trigger pod restart on ConfigMap change
✓ Always set resource requests and limits in default values
✓ Exclude kube-system from namespace selectors in webhooks
✓ Use --atomic in CI/CD for automatic rollback on failure
✓ Never hardcode secrets in values.yaml — use secret references
✓ Lint and template test in CI before deploying
✓ Document every value in values.yaml with comments
✓ Use helm test to verify the release works after deployment
✓ Set .helmignore to exclude README, CI files from packaged chart
```