#!/bin/bash
# Script to create a k3s cluster using Docker

set -e  # Exit on error

CLUSTER_NAME="dev"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/kubeconfig"

echo "========================================="
echo "Creating k3s Cluster"
echo "========================================="
echo ""

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo "ERROR: docker is not installed"
    exit 1
fi

# Check if cluster already exists
if docker ps -a --format '{{.Names}}' | grep -q "^k3s-dev$"; then
    echo "ERROR: Cluster 'k3s-dev' already exists"
    echo "Delete it first with: ./delete-cluster.sh"
    echo "Or use ./reset-cluster.sh to recreate from scratch"
    exit 1
fi

# Create kubeconfig directory
mkdir -p "${KUBECONFIG_DIR}"

# Start k3s cluster
echo "Starting k3s cluster..."
docker compose -f "${SCRIPT_DIR}/docker-compose.yaml" up -d

# Wait for kubeconfig to be generated
echo "Waiting for kubeconfig..."
for i in {1..30}; do
    if [ -f "${KUBECONFIG_DIR}/kubeconfig.yaml" ]; then
        break
    fi
    sleep 1
done

if [ ! -f "${KUBECONFIG_DIR}/kubeconfig.yaml" ]; then
    echo "ERROR: Kubeconfig not generated"
    exit 1
fi

# Set up kubectl context
export KUBECONFIG="${KUBECONFIG_DIR}/kubeconfig.yaml"

# Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
for i in {1..60}; do
    # Wait until at least one node appears
    if kubectl get nodes --no-headers 2>/dev/null | grep -q .; then
        break
    fi
    sleep 2
done

# Merge kubeconfig into ~/.kube/config
echo "Setting ~/.kube/config..."
mkdir -p ~/.kube
cp "${KUBECONFIG_DIR}/kubeconfig.yaml" ~/.kube/config
kubectl wait --for=condition=Ready node --all --timeout=120s

# Wait for coredns deployment to exist
echo "Waiting for coredns..."
for i in {1..30}; do
    if kubectl get deploy/coredns -n kube-system &>/dev/null; then
        break
    fi
    sleep 2
done
kubectl wait --for=condition=available -n kube-system deploy/coredns --timeout=120s

echo ""
echo "✓ Cluster created successfully!"
echo ""

# Verify cluster
echo "Verifying cluster..."
kubectl cluster-info

echo ""
echo "========================================="
echo "Cluster ready!"
echo "========================================="
echo ""

echo ""
echo "Kubeconfig: merged into ~/.kube/config"
echo "Context:    default"
echo "Registry:   localhost:5001"
echo ""
