#!/bin/bash
# Script to create a Kind cluster

set -e  # Exit on error

CLUSTER_NAME="dev"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/kind-config.yaml"

echo "========================================="
echo "Creating Kind Cluster"
echo "========================================="
echo ""

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "ERROR: kind is not installed"
    echo "Install it from: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi

# Check if cluster already exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "ERROR: Cluster '${CLUSTER_NAME}' already exists"
    echo "Delete it first with: kind delete cluster --name ${CLUSTER_NAME}"
    echo "Or use ./reset-cluster.sh to recreate from scratch"
    exit 1
fi

# Create cluster
echo "Creating cluster '${CLUSTER_NAME}'..."
echo "Using config: ${CONFIG_FILE}"
kind create cluster --config="${CONFIG_FILE}" --name="${CLUSTER_NAME}"

# Wait for coredns to be ready
echo "Waiting for coredns to be ready..."
kubectl wait --for condition=available -n kube-system deploy/coredns

echo ""
echo "✓ Cluster created successfully!"
echo ""

# Verify cluster
echo "Verifying cluster..."
kubectl cluster-info --context "kind-${CLUSTER_NAME}"

echo ""
echo "========================================="
echo "Cluster ready!"
echo "========================================="
echo ""
echo "Context: kind-${CLUSTER_NAME}"
echo ""
echo "Set kubectl context:"
echo "  kubectl config use-context kind-${CLUSTER_NAME}"
echo ""
