#!/bin/bash
# Script to initialize the Kind cluster with all resources in order

set -e  # Exit on error

CLUSTER_NAME="dev"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "Initializing Cluster: ${CLUSTER_NAME}"
echo "========================================="
echo ""

# Check if cluster exists
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "ERROR: Cluster '${CLUSTER_NAME}' does not exist"
    echo "Create it first with: ./create-cluster.sh"
    exit 1
fi

# Check if kubectl context is correct
kubectl config use-context "kind-${CLUSTER_NAME}"

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
echo "Step 3: Deploying operators (trust-manager, cnpg-operator, nginx-ingress)..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-operators

echo ""
echo "Waiting for operators to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/trust-manager -n cert-manager
kubectl wait --for=condition=available --timeout=120s deployment/cnpg-controller-manager -n cnpg-system
kubectl wait --for=condition=ready --timeout=120s pod -l app.kubernetes.io/component=controller -n ingress-nginx

echo ""
echo "Step 4: Deploying infrastructure (certificates, registry, otel-collector, postgres)..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-infra

echo ""
echo "Waiting for infrastructure to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/registry -n mgmt
kubectl wait --for=condition=available --timeout=120s deployment/otel-collector -n mgmt
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
kubectl wait --for=condition=complete --timeout=180s job/migrate-runner -n app-namespace

echo ""
echo "Step 7: Deploying applications (grpcserver)..."
echo "-----------------------------------"
bazel run //k8s/app:apply-dev-apps

echo ""
echo "Waiting for applications to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/grpcserver -n app-namespace

echo ""
echo "========================================="
echo "Cluster initialized successfully!"
echo "========================================="
echo ""
echo "Deployed resources:"
kubectl get all -n mgmt --context "kind-${CLUSTER_NAME}"
echo ""
kubectl get all -n app-namespace --context "kind-${CLUSTER_NAME}"
echo ""
