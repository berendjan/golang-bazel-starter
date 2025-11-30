#!/bin/bash
# Script to deploy a single component to the dev cluster

set -e

CLUSTER_NAME="dev"
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
    echo "    - nginx-ingress"
    echo ""
    echo "  Infrastructure:"
    echo "    - certificates"
    echo "    - environment"
    echo "    - registry"
    echo "    - otel-collector"
    echo "    - postgres"
    echo ""
    echo "  Applications:"
    echo "    - grpcserver"
    echo ""
    echo "Examples:"
    echo "  $0 grpcserver       # Deploy just grpcserver"
    echo "  $0 registry         # Deploy just registry"
    exit 1
fi

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "ERROR: Cluster '${CLUSTER_NAME}' does not exist"
    echo "Create it first with: ./reset-cluster.sh"
    exit 1
fi

# Set kubectl context
kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null 2>&1

echo "========================================="
echo "Deploying: ${COMPONENT}"
echo "Cluster: ${CLUSTER_NAME}"
echo "========================================="
echo ""

# Run the bazel target
bazel run //k8s/app:${COMPONENT}-dev-apply

echo ""
echo "========================================="
echo "✓ ${COMPONENT} deployed successfully!"
echo "========================================="
