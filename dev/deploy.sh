#!/bin/bash
# Script to deploy a single component to the dev cluster

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/kubeconfig"
COMPONENT=$1

if [ -z "$COMPONENT" ]; then
    echo "Usage: $0 <component-name>"
    echo ""
    echo "Available components:"
    echo "  Namespaces:"
    echo "    - namespace"
    echo ""
    echo "  Cert Operators:"
    echo "    - cert-manager"
    echo "    - trust-manager"
    echo ""
    echo "  Operators:"
    echo "    - cnpg-operator"
    echo ""
    echo "  Infrastructure:"
    echo "    - certificates"
    echo "    - environment"
    echo "    - otel-collector"
    echo "    - postgres"
    echo ""
    echo "  Migrations:"
    echo "    - dbmate-config"
    echo "    - dbmate-auth"
    echo ""
    echo "  Applications:"
    echo "    - kratos"
    echo "    - grpcserver"
    echo "    - frontend"
    echo ""
    echo "Examples:"
    echo "  $0 grpcserver       # Deploy just grpcserver"
    echo "  $0 frontend         # Deploy just frontend"
    exit 1
fi

# Check if cluster is running
if ! docker ps --format '{{.Names}}' | grep -q "^k3s-dev$"; then
    echo "ERROR: Cluster 'k3s-dev' is not running"
    echo "Create it first with: ./reset-cluster.sh"
    exit 1
fi

# Check if kubeconfig exists
if [ ! -f "${KUBECONFIG_DIR}/kubeconfig.yaml" ]; then
    echo "ERROR: Kubeconfig not found at ${KUBECONFIG_DIR}/kubeconfig.yaml"
    exit 1
fi

# Set kubeconfig
export KUBECONFIG="${KUBECONFIG_DIR}/kubeconfig.yaml"

echo "========================================="
echo "Deploying: ${COMPONENT}"
echo "========================================="
echo ""

# Run the bazel target
bazel run //k8s/app:${COMPONENT}-dev-apply

echo ""
echo "========================================="
echo "✓ ${COMPONENT} deployed successfully!"
echo "========================================="
