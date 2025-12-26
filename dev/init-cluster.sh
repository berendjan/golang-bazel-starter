#!/bin/bash
# Script to initialize the k3s cluster with all resources in order

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECONFIG_DIR="${SCRIPT_DIR}/kubeconfig"

echo "========================================="
echo "Initializing k3s Cluster"
echo "========================================="
echo ""

# Check if cluster is running
if ! docker ps --format '{{.Names}}' | grep -q "^k3s-dev$"; then
    echo "ERROR: Cluster 'k3s-dev' is not running"
    echo "Create it first with: ./create-cluster.sh"
    exit 1
fi

# Check if kubeconfig exists
if [ ! -f "${KUBECONFIG_DIR}/kubeconfig.yaml" ]; then
    echo "ERROR: Kubeconfig not found at ${KUBECONFIG_DIR}/kubeconfig.yaml"
    exit 1
fi

# Set kubeconfig
export KUBECONFIG="${KUBECONFIG_DIR}/kubeconfig.yaml"

echo "Step 1: Deploying namespaces..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-namespaces

echo ""
echo "Waiting for namespaces to be ready..."
kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/mgmt --timeout=30s
kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/app-namespace --timeout=30s

echo ""
echo "Step 2: Deploying cert-manager..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-cert-operators

echo ""
echo "Waiting for cert-manager to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager -n cert-manager
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager-webhook -n cert-manager
kubectl wait --for=condition=available --timeout=120s deployment/cert-manager-cainjector -n cert-manager
kubectl wait --for=condition=ready --timeout=120s pod -l app.kubernetes.io/name=webhook -n cert-manager

echo ""
echo "Step 3: Deploying operators (trust-manager, cnpg-operator)..."
echo "-----------------------------------"
# Note: Traefik ingress is provided by k3s, not deployed separately
bazel run //k8s/app:apply-dev-operators

echo ""
echo "Waiting for operators to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/trust-manager -n cert-manager
kubectl wait --for=condition=available --timeout=120s deployment/cnpg-controller-manager -n cnpg-system
kubectl wait --for=condition=available --timeout=120s deployment/traefik -n kube-system

echo ""
echo "Step 4: Deploying infrastructure (certificates, otel-collector, postgres)..."
echo "-----------------------------------"
# Note: Registry is provided by docker-compose, not deployed separately
bazel run //k8s/app:apply-dev-infra

echo ""
echo "Waiting for infrastructure to be ready..."
# Registry is managed by docker-compose - skip waiting for it
kubectl wait --for=condition=available --timeout=120s deployment/otel-collector -n mgmt || true
kubectl wait --for=condition=Ready --timeout=300s cluster/app-postgres -n app-namespace

echo ""
echo "Step 5: Building and pushing application images..."
echo "-----------------------------------"
"${SCRIPT_DIR}/push-images.sh"

echo ""
echo "Step 6: Running database migrations..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-migrations

echo ""
echo "Waiting for migrations to complete..."
kubectl wait --for=condition=complete --timeout=180s job/dbmate-config -n app-namespace
kubectl wait --for=condition=complete --timeout=180s job/dbmate-auth -n app-namespace

echo ""
echo "Step 7: Deploying applications (kratos, grpcserver, frontend)..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-apps

echo ""
echo "Waiting for applications to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/kratos -n app-namespace
kubectl wait --for=condition=available --timeout=120s deployment/grpcserver -n app-namespace
kubectl wait --for=condition=available --timeout=120s deployment/frontend -n app-namespace

echo ""
echo "========================================="
echo "Cluster initialized successfully!"
echo "========================================="
echo ""
echo "Deployed resources:"
kubectl get all -n mgmt
echo ""
kubectl get all -n app-namespace
echo ""
echo "Access:"
echo "  Frontend: https://frontend.localhost"
echo "  API:      https://api.localhost"
echo "  Auth:     https://auth.localhost"
echo ""
echo "To use kubectl:"
echo "  export KUBECONFIG=${KUBECONFIG_DIR}/kubeconfig.yaml"
echo ""
