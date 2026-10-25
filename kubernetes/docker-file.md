# Complete Dockerfile Guide

A Dockerfile is a text file containing instructions used by Docker to build a container image.

A typical production Dockerfile may look like this:

```dockerfile ̰
# ---------------------------------------------------------
# 1. Base image
# ---------------------------------------------------------
FROM golang:1.25.4 AS builder

# ---------------------------------------------------------
# 2. Metadata
# ---------------------------------------------------------
LABEL maintainer="platform-team@example.com"
LABEL description="Example Go application"

# ---------------------------------------------------------
# 3. Build arguments
# ---------------------------------------------------------
ARG APP_VERSION="1.0.0"

# ---------------------------------------------------------
# 4. Environment variables
# ---------------------------------------------------------
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# ---------------------------------------------------------
# 5. Working directory
# ---------------------------------------------------------
WORKDIR /app

# ---------------------------------------------------------
# 6. Copy dependency files first
# ---------------------------------------------------------
COPY go.mod go.sum ./

# ---------------------------------------------------------
# 7. Download dependencies
# ---------------------------------------------------------
RUN go mod download

# ---------------------------------------------------------
# 8. Copy application source
# ---------------------------------------------------------
COPY . .

# ---------------------------------------------------------
# 9. Build the application
# ---------------------------------------------------------
RUN go build \
    -ldflags="-X main.version=${APP_VERSION}" \
    -o /app/myapp .

# ---------------------------------------------------------
# 10. Runtime image
# ---------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

# ---------------------------------------------------------
# 11. Runtime metadata
# ---------------------------------------------------------
LABEL description="Runtime image for myapp"

# ---------------------------------------------------------
# 12. Working directory
# ---------------------------------------------------------
WORKDIR /app

# ---------------------------------------------------------
# 13. Copy binary from builder image
# ---------------------------------------------------------
COPY --from=builder /app/myapp /app/myapp

# ---------------------------------------------------------
# 14. Environment variable
# ---------------------------------------------------------
ENV APP_ENV=production

# ---------------------------------------------------------
# 15. Expose application port
# ---------------------------------------------------------
EXPOSE 8080

# ---------------------------------------------------------
# 16. Health check
# ---------------------------------------------------------
# Note: distroless images generally don't contain curl/wget,
# so health checks are often implemented by Kubernetes instead.
# HEALTHCHECK CMD ...

# ---------------------------------------------------------
# 17. Default command
# ---------------------------------------------------------
ENTRYPOINT ["/app/myapp"]

# ---------------------------------------------------------
# 18. Default arguments
# ---------------------------------------------------------
CMD ["--config=/etc/myapp/config.yaml"]
```

---

# 1. FROM

```dockerfile
FROM golang:1.25.4
```

`FROM` specifies the **base image**.

Every Dockerfile normally starts with a `FROM` instruction.

For example:

```dockerfile
FROM ubuntu:24.04
```

or:

```dockerfile
FROM golang:1.25.4
```

or:

```dockerfile
FROM alpine:3.22
```

The base image provides the initial filesystem and environment for the image.

Think of it as:

```text
FROM ubuntu
     |
     +-- Linux filesystem
     +-- system libraries
     +-- package manager
     +-- basic utilities
```

## Why use different base images?

### Full Ubuntu

```dockerfile
FROM ubuntu:24.04
```

Large but convenient.

### Alpine

```dockerfile
FROM alpine:3.22
```

Much smaller.

### Go image

```dockerfile
FROM golang:1.25.4
```

Contains the Go compiler and tooling.

### Distroless

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
```

Very small runtime image with minimal utilities.

---

# 2. AS

You can give a build stage a name:

```dockerfile
FROM golang:1.25.4 AS builder
```

Now the stage is called:

```text
builder
```

You can later copy files from it:

```dockerfile
COPY --from=builder /app/myapp /app/myapp
```

This is called a **multi-stage build**.

The benefit is that the Go compiler doesn't need to exist in the final image.

Instead:

```text
Go source
   |
   v
golang image
   |
   | compile
   v
myapp binary
   |
   v
distroless image
   |
   v
small production image
```

This is extremely common in production.

---

# 3. LABEL

```dockerfile
LABEL maintainer="platform-team@example.com"
LABEL description="Example Go application"
```

`LABEL` adds metadata to the image.

You can inspect it using:

```bash
docker image inspect myapp
```

Labels are commonly used for:

* ownership
* versioning
* documentation
* automation
* CI/CD
* Kubernetes tooling

For example:

```dockerfile
LABEL org.opencontainers.image.version="1.2.3"
```

---

# 4. ARG

```dockerfile
ARG APP_VERSION="1.0.0"
```

`ARG` defines a **build-time variable**.

You can override it:

```bash
docker build \
  --build-arg APP_VERSION=2.0.0 \
  -t myapp:2.0.0 .
```

Inside the Dockerfile:

```dockerfile
RUN echo ${APP_VERSION}
```

Important distinction:

```text
ARG → available during image BUILD
ENV → available during container RUNTIME
```

For example:

```dockerfile
ARG VERSION=1.0.0
ENV APP_ENV=production
```

---

# 5. ENV

```dockerfile
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
```

`ENV` defines environment variables that are available in subsequent Dockerfile instructions and normally remain available when the container runs.

For example:

```dockerfile
ENV APP_ENV=production
```

Inside the container:

```bash
echo $APP_ENV
```

produces:

```text
production
```

You can also override an environment variable:

```bash
docker run -e APP_ENV=staging myapp
```

---

# 6. WORKDIR

```dockerfile
WORKDIR /app
```

Sets the current working directory.

After:

```dockerfile
WORKDIR /app
```

this:

```dockerfile
RUN pwd
```

produces:

```text
/app
```

It also affects:

```dockerfile
RUN
COPY
ADD
CMD
ENTRYPOINT
```

For example:

```dockerfile
WORKDIR /app

COPY . .

RUN go build .
```

is cleaner than repeatedly using absolute paths.

---

# 7. COPY

```dockerfile
COPY go.mod go.sum ./
```

Copies files from the Docker build context into the image.

For example:

```dockerfile
COPY . .
```

means:

```text
Host:
.
|
+-- main.go
+-- go.mod
+-- go.sum
|
v
Container:
/app
|
+-- main.go
+-- go.mod
+-- go.sum
```

You can also copy a specific file:

```dockerfile
COPY config.yaml /app/config.yaml
```

---

# 8. ADD

You may also see:

```dockerfile
ADD file.tar.gz /app/
```

`ADD` is similar to `COPY`, but has additional behavior such as handling local tar archives.

For most normal file copying, prefer:

```dockerfile
COPY
```

rather than:

```dockerfile
ADD
```

Interview answer:

> `COPY` is generally preferred for simple file copying because it has more predictable behavior. `ADD` provides additional features such as local tar extraction and remote URL support in certain contexts.

---

# 9. RUN

```dockerfile
RUN go mod download
```

`RUN` executes a command **during image building**.

For example:

```dockerfile
RUN apt-get update
RUN apt-get install -y curl
RUN go build .
```

Each `RUN` instruction creates a new image layer.

You can combine commands:

```dockerfile
RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*
```

This can help keep the resulting image smaller.

---

# 10. EXPOSE

```dockerfile
EXPOSE 8080
```

Documents that the application listens on port `8080`.

Important:

`EXPOSE` **does not actually publish the port**.

You still need:

```bash
docker run -p 8080:8080 myapp
```

Think of:

```text
EXPOSE
```

as documentation/metadata.

While:

```text
-p 8080:8080
```

actually publishes the port from host to container.

---

# 11. CMD

```dockerfile
CMD ["--config=/etc/myapp/config.yaml"]
```

`CMD` specifies the **default command or default arguments** for the container.

There are two common forms.

## Exec form

```dockerfile
CMD ["./myapp"]
```

Preferred.

## Shell form

```dockerfile
CMD ./myapp
```

Exec form is generally preferred because it provides better signal handling and doesn't unnecessarily start a shell.

---

# 12. ENTRYPOINT

```dockerfile
ENTRYPOINT ["/app/myapp"]
```

`ENTRYPOINT` specifies the executable that the container is intended to run.

If you have:

```dockerfile
ENTRYPOINT ["/app/myapp"]
CMD ["--config=/etc/myapp/config.yaml"]
```

Docker effectively runs:

```bash
/app/myapp --config=/etc/myapp/config.yaml
```

This is a very useful combination.

---

# 13. ENTRYPOINT vs CMD

This is a common interview question.

Suppose:

```dockerfile
ENTRYPOINT ["/app/myapp"]
CMD ["--port=8080"]
```

Running:

```bash
docker run myapp
```

runs:

```bash
/app/myapp --port=8080
```

But:

```bash
docker run myapp --port=9090
```

runs:

```bash
/app/myapp --port=9090
```

The `CMD` arguments were replaced.

Think:

```text
ENTRYPOINT = what the container IS
CMD        = default arguments
```

For example:

```text
ENTRYPOINT = myapp
CMD        = --port=8080
```

---

# 14. USER

You can specify which Linux user runs the application:

```dockerfile
USER 1000
```

or:

```dockerfile
USER appuser
```

Running as a non-root user is strongly recommended for production containers when possible.

For example:

```dockerfile
RUN useradd -r appuser

USER appuser
```

This reduces the impact of a container compromise.

---

# 15. VOLUME

```dockerfile
VOLUME ["/data"]
```

Declares a mount point for persistent data.

For example:

```dockerfile
VOLUME ["/var/lib/myapp"]
```

However, in modern Kubernetes environments, persistent storage is normally managed using Kubernetes volumes/PVCs rather than relying on Dockerfile `VOLUME`.

---

# 16. HEALTHCHECK

You can specify a container health check:

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s \
    CMD curl -f http://localhost:8080/health || exit 1
```

Docker can then report:

```text
healthy
unhealthy
starting
```

However, in Kubernetes, you will commonly use:

```yaml
livenessProbe:
readinessProbe:
startupProbe:
```

instead.

---

# 17. SHELL

You can change the default shell used by shell-form commands.

For example:

```dockerfile
SHELL ["/bin/bash", "-c"]
```

Then:

```dockerfile
RUN echo "hello"
```

uses Bash instead of the default shell.

This is less common in typical Go/Linux Dockerfiles.

---

# 18. STOPSIGNAL

You can specify which signal Docker sends when stopping the container:

```dockerfile
STOPSIGNAL SIGTERM
```

For most applications, `SIGTERM` is appropriate.

Your application should gracefully shut down when it receives the signal.

This is particularly important for Go applications because you often want to:

```text
SIGTERM
   |
   v
Stop accepting new requests
   |
   v
Finish existing requests
   |
   v
Close connections
   |
   v
Flush telemetry
   |
   v
Exit
```

---

# 19. ONBUILD

`ONBUILD` defines instructions that execute when another Dockerfile uses your image as its base.

Example:

```dockerfile
ONBUILD COPY . /app
```

This is relatively uncommon in modern application Dockerfiles.

---

# 20. Multi-stage builds

This is one of the most important concepts for production Docker images.

Instead of:

```dockerfile
FROM golang:1.25.4

COPY . .

RUN go build -o myapp .

CMD ["./myapp"]
```

you can use:

```dockerfile
FROM golang:1.25.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myapp .
```

Then:

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/myapp .

USER nonroot:nonroot

ENTRYPOINT ["/app/myapp"]
```

The final image contains only what is needed to run the application.

---

# 21. .dockerignore

A Docker project should normally have a `.dockerignore` file.

Example:

```text
.git
.gitignore
Dockerfile
README.md
*.log
tmp/
bin/
coverage/
.env
```

This prevents unnecessary files from being sent to the Docker build context.

For example, without `.dockerignore`:

```text
Project
 ├── .git/          ← unnecessary
 ├── node_modules/  ← unnecessary
 ├── logs/          ← unnecessary
 ├── source code
 └── Dockerfile
```

Docker may send all of this as build context.

With `.dockerignore`, only the required files are sent.

---

# 22. Build the image

Given:

```text
Dockerfile
go.mod
go.sum
main.go
```

run:

```bash
docker build -t myapp:1.0 .
```

The `.` means:

> Use the current directory as the Docker build context.

---

# 23. Run the container

```bash
docker run -p 8080:8080 myapp:1.0
```

The flow is:

```text
Dockerfile
    |
    | docker build
    v
Docker Image
    |
    | docker run
    v
Container
    |
    v
Go Application
```

---

# 24. Important Dockerfile instructions for interviews

You don't need to memorize every Dockerfile instruction equally.

Focus heavily on:

```text
FROM
RUN
COPY
ADD
WORKDIR
ENV
ARG
EXPOSE
CMD
ENTRYPOINT
USER
VOLUME
HEALTHCHECK
```

And especially understand:

```text
CMD vs ENTRYPOINT
ARG vs ENV
COPY vs ADD
RUN vs CMD
EXPOSE vs -p
multi-stage builds
Docker layers
.dockerignore
non-root containers
```

---

# 25. Dockerfile instructions vs runtime behavior

A useful mental model is:

```text
                 BUILD TIME
                     │
                     ▼
FROM ──► WORKDIR ──► COPY ──► RUN
                     │
                     ▼
                  IMAGE
                     │
                     │ docker run
                     ▼
                 RUNTIME
                     │
             ┌───────┴────────┐
             ▼                ▼
        ENTRYPOINT           CMD
             │
             ▼
         Application
```

`ARG` is primarily build-time.

`ENV` can be used during build and is also available in the resulting container environment.

`CMD` and `ENTRYPOINT` are primarily about what happens when the container starts.

---

# 26. A production-style Go Dockerfile

For your Go/Kubernetes work, this is probably the pattern I'd recommend knowing well:

```dockerfile
# Build stage
FROM golang:1.25.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/myapp .

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/myapp /app/myapp

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/myapp"]
```

This gives you:

* a compiler in the **builder image**
* a small runtime image
* no Go compiler in production
* a non-root application
* a statically compiled Go binary
* a clean separation between build and runtime

For a Kubernetes interview, this is a very good Dockerfile pattern to be able to explain from memory.






| Instruction   | One-line meaning                                                                             |
| ------------- | -------------------------------------------------------------------------------------------- |
| `FROM`        | Specifies the **base image** from which your image is built.                                 |
| `RUN`         | Executes a **command during image build time** and creates a new image layer.                |
| `COPY`        | Copies files/directories from the **build context into the image**.                          |
| `ADD`         | Copies files like `COPY`, but also supports features such as **local archive extraction**.   |
| `WORKDIR`     | Sets the **working directory** for subsequent Dockerfile commands and the container.         |
| `ENV`         | Defines an **environment variable** that is available during build and/or container runtime. |
| `ARG`         | Defines a **build-time variable** that can be supplied using `docker build --build-arg`.     |
| `EXPOSE`      | Documents the **port the application listens on**; it does not publish the port.             |
| `CMD`         | Defines the **default command or default arguments** executed when the container starts.     |
| `ENTRYPOINT`  | Defines the **main executable** that the container is intended to run.                       |
| `USER`        | Specifies the **Linux user** used to run subsequent commands and/or the application.         |
| `VOLUME`      | Declares a **mount point for persistent/external data**.                                     |
| `HEALTHCHECK` | Defines a command Docker can use to determine whether the **container is healthy**.          |



1. CMD vs ENTRYPOINT
   ENTRYPOINT defines the executable that should always run, while CMD provides default arguments or a default command that can be overridden at runtime.
   CMD ["./app", "--port=8080"]
   ENTRYPOINT ["./app"]
2. ARG vs ENV
   ARG is primarily used for build-time configuration, whereas ENV defines environment variables that are available in the resulting container.
   ARG VERSION=1.0
   ENV APP_ENV=production
3. COPY vs ADD
   COPY should generally be preferred for normal file copying because its behavior is simpler and more predictable. ADD has additional features such as local archive extraction.
   COPY . /app
   ADD app.tar.gz /app
4. RUN vs CMD
   RUN builds the image. CMD runs the application.
   RUN go build -o app .
   CMD ["./app"]