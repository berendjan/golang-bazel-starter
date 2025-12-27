# Local Development Environment

This directory contains scripts and configuration for managing a local Kubernetes cluster using [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker).

## Prerequisites

### Required
- **Docker**: Must be running
- **Kind**: Install from https://kind.sigs.k8s.io/docs/user/quick-start/#installation

```bash
# macOS (Homebrew)
brew install kind

# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Recommended
- **kubectl**: Kubernetes CLI
  ```bash
  # macOS (Homebrew)
  brew install kubectl

  # Linux
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl
  sudo mv kubectl /usr/local/bin/
  ```

## Cluster Configuration

The cluster is configured in `kind-config.yaml`:

- **Default CNI** (kindnet) - Pre-installed for networking
- **kube-proxy** - Enabled for service routing
- **Single control-plane node** - Minimal resource usage
- **Network configuration**:
  - API Server: `6443`
  - Pod subnet: `10.244.0.0/16`
  - Service subnet: `10.96.0.0/12`
- **Port mappings**:
  - `80` → nginx ingress HTTP
  - `443` → nginx ingress HTTPS
  - `5001` → Container registry (port 5000)
  - `25000` → gRPC service (port 25000)
  - `26000` → HTTP gateway (port 26000)

## Scripts

### `deploy.sh`

**Quick deploy**: Deploy a single component to the cluster.

```bash
./deploy.sh <component-name>
```

Examples:
```bash
# Deploy just grpcserver after code changes
./deploy.sh grpcserver

# Update just the registry
./deploy.sh registry

# Apply certificate changes
./deploy.sh certificates
```

**Note**: Component must already exist in cluster. For first-time setup, use `init-cluster.sh`.

### `reset-cluster.sh`

**Recommended**: Tears down and recreates the cluster from scratch, then deploys all resources.

```bash
./reset-cluster.sh
```

This script:
1. Deletes existing cluster (if present)
2. Creates new cluster from config
3. Deploys all resources (namespaces, operators, infrastructure)
4. Builds and pushes images to in-cluster registry
5. Deploys applications
6. Verifies everything is running
7. Shows deployed resources

### `init-cluster.sh`

Deploys all Kubernetes resources to the cluster in the correct order.

```bash
./init-cluster.sh
```

This script:
1. Deploys namespaces (mgmt, app-namespace)
2. Deploys operators (cnpg-operator, cert-manager, trust-manager, nginx-ingress)
3. Deploys infrastructure (certificates, registry, otel-collector, postgres)
4. Deploys Ory Kratos for authentication
5. Builds and pushes application images to registry (grpcserver, frontend, dbmate)
6. Deploys applications (grpcserver, frontend)
7. Waits for each layer to be ready before proceeding
8. Shows deployed resources

**Note**: Run this after `create-cluster.sh` or `reset-cluster.sh`.

### `create-cluster.sh`

Creates a new cluster. Fails if cluster already exists.

```bash
./create-cluster.sh
```

### `delete-cluster.sh`

Deletes the existing cluster.

```bash
./delete-cluster.sh
```

## Quick Start

### 1. Create and initialize the cluster

```bash
cd dev
./reset-cluster.sh
```

This single command will:
- Delete any existing cluster
- Create a new Kind cluster
- Deploy all resources (namespaces, operators, infrastructure, applications)
- Wait for everything to be ready

### 2. (Optional) Set kubectl context

The script automatically sets the context, but you can manually set it:

```bash
kubectl config use-context kind-dev
```

### 3. (Optional) Replace CNI

The cluster comes with **kindnet CNI** pre-installed. Pods can communicate out of the box.

If you want to use a different CNI (optional):

**Option A: Cilium** (advanced networking with eBPF)
```bash
# Remove kindnet first
kubectl delete daemonset kindnet -n kube-system
# Install Cilium
cilium install
```

**Option B: Calico** (policy-based networking)
```bash
# Remove kindnet first
kubectl delete daemonset kindnet -n kube-system
# Install Calico
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/calico.yaml
```

**Note**: To disable the default CNI, uncomment `disableDefaultCNI: true` in `kind-config.yaml`

### 4. Verify cluster

```bash
# Check nodes
kubectl get nodes

# Check all pods in mgmt namespace
kubectl get pods -n mgmt

# Check all pods in app namespace
kubectl get pods -n app-namespace

# Check PostgreSQL cluster status
kubectl get cluster -n mgmt

# Check cluster info
kubectl cluster-info
```

## Deploying to the Cluster

### Deploy from local registry

If you've pushed images to `localhost:5001`:

```bash
# Push image to registry
bazel run //golang/grpcserver:grpcserver_push

# Create deployment using local image
kubectl apply -f k8s/app/grpcserver/base/deployment.yaml
```

### Load image directly into Kind

Alternatively, load images directly without a registry:

```bash
# Build the image
bazel build //golang/grpcserver:grpcserver_image

# Load into Kind
kind load docker-image localhost:5001/grpcserver:latest --name dev
```

## Accessing Services

Services running in the cluster can be accessed via:

### Ingress (Recommended)

The cluster includes nginx-ingress with TLS. Access services via hostname:

```bash
# Frontend (React app with Tailwind CSS)
https://frontend.localhost

# Kratos (authentication)
https://kratos.localhost

# API (via frontend proxy)
https://frontend.localhost/api/v1/accounts
```

> **Note**: Add entries to `/etc/hosts` if needed, though `.localhost` domains typically resolve automatically.

### NodePort Services

Map to the ports configured in `kind-config.yaml`:
- Port `25000` → gRPC service
- Port `26000` → HTTP gateway

```bash
# Access via localhost
curl http://localhost:26000/v1/accounts
```

### Port Forwarding

For other ports:

```bash
# Forward pod port to localhost
kubectl port-forward pod/grpcserver-xxx 8080:26000

# Access via forwarded port
curl http://localhost:8080/v1/accounts
```

## Troubleshooting

### Cluster won't start

```bash
# Check Docker is running
docker ps

# Clean up and retry
./reset-cluster.sh
```

### Pods stuck in Pending

Likely due to missing CNI. Either:
1. Install a CNI (see step 3 above)
2. Check pod events: `kubectl describe pod <pod-name>`

### Cannot access services

1. Verify port mappings in `kind-config.yaml`
2. Check service is running: `kubectl get pods`
3. Check service logs: `kubectl logs <pod-name>`

### Image pull errors

If using local registry:

```bash
# Verify registry is running
docker ps | grep registry

# Start registry if needed
docker run -d -p 5001:5000 --name registry registry:3

# Push image again
bazel run //golang/grpcserver:grpcserver_push
```

## Cluster Management

### List clusters

```bash
kind get clusters
```

### Get cluster info

```bash
kubectl cluster-info --context kind-dev
```

### Delete and recreate

```bash
./reset-cluster.sh
```

### Export kubeconfig

```bash
kind export kubeconfig --name dev
```

## Advanced Usage

### Add worker nodes

Edit `kind-config.yaml` and uncomment the worker node sections:

```yaml
nodes:
  - role: control-plane
    # ... config ...
  - role: worker  # Uncomment this
  - role: worker  # And this
```

Then recreate:
```bash
./reset-cluster.sh
```

### Custom registry

To use the cluster with the local registry at `localhost:5001`, ensure the registry is running:

```bash
# Start registry
docker run -d -p 5001:5000 --name registry registry:3

# Verify
curl http://localhost:5001/v2/_catalog
```

The Kind cluster is configured to access `localhost:5001` via the port mapping.

### Persisting data

Add volume mounts in `kind-config.yaml`:

```yaml
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: ./data
        containerPath: /data
```

## Additional Resources

- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
