#!/bin/bash
# Script to delete the Kind cluster

set -e  # Exit on error

CLUSTER_NAME="dev"

echo "========================================="
echo "Deleting Kind Cluster"
echo "========================================="
echo ""

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "ERROR: kind is not installed"
    exit 1
fi

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster '${CLUSTER_NAME}' does not exist"
    exit 0
fi

# Delete cluster
kind delete cluster --name "${CLUSTER_NAME}"

echo ""
echo "✓ Cluster deleted successfully!"
echo ""
