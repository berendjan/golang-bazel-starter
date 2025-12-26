#!/bin/bash
# Script to delete the k3s cluster

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/kubeconfig"

echo "========================================="
echo "Deleting k3s Cluster"
echo "========================================="
echo ""

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo "ERROR: docker is not installed"
    exit 1
fi

# Check if cluster exists
if ! docker ps -a --format '{{.Names}}' | grep -q "^k3s-dev$"; then
    echo "Cluster 'k3s-dev' does not exist"
    exit 0
fi

# Stop and remove containers
echo "Stopping containers..."
docker compose -f "${SCRIPT_DIR}/docker-compose.yaml" down -v

# Clean up kubeconfig
rm -rf "${KUBECONFIG_DIR}"

echo ""
echo "✓ Cluster deleted successfully!"
echo ""
