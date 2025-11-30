#!/bin/bash
# Script to build and push application images to the registry

set -e  # Exit on error

REGISTRY_URL="registry.localhost"

echo "========================================="
echo "Building and Pushing Images"
echo "Registry: ${REGISTRY_URL}"
echo "========================================="
echo ""

# Check if /etc/hosts has registry.localhost entry
if ! grep -q "registry.localhost" /etc/hosts; then
    echo "Adding registry.localhost to /etc/hosts..."
    echo "127.0.0.1 registry.localhost" | sudo tee -a /etc/hosts
fi

# Check if registry is accessible
echo "Checking registry availability at ${REGISTRY_URL}..."
if ! curl -s http://${REGISTRY_URL}/v2/ > /dev/null; then
    echo "ERROR: Registry not accessible at ${REGISTRY_URL}"
    echo ""
    echo "Troubleshooting steps:"
    echo "  1. Check nginx-ingress is running: kubectl get pods -n ingress-nginx"
    echo "  2. Check registry ingress: kubectl get ingress -n mgmt"
    echo "  3. Verify /etc/hosts has: 127.0.0.1 registry.localhost"
    echo "  4. Test manually: curl http://registry.localhost/v2/"
    exit 1
fi

echo "✓ Registry is accessible"
echo ""

# Build and push migrate-runner
echo "Building and pushing migrate-runner..."
echo "-----------------------------------"
bazel run //golang/migrate-runner:migrate-runner_push

echo ""

# Build and push grpcserver
echo "Building and pushing grpcserver..."
echo "-----------------------------------"
bazel run //golang/grpcserver:grpcserver_push

echo ""
echo "Verifying images in registry..."
echo "-----------------------------------"

# List all images in registry
echo "Images in registry:"
curl -s http://${REGISTRY_URL}/v2/_catalog | jq '.'

echo ""

# Check migrate-runner specifically
if curl -s http://${REGISTRY_URL}/v2/_catalog | grep -q migrate-runner; then
    echo "✓ migrate-runner image pushed successfully"
    echo ""
    echo "Tags for migrate-runner:"
    curl -s http://${REGISTRY_URL}/v2/migrate-runner/tags/list | jq '.'
else
    echo "✗ migrate-runner image not found in registry"
    exit 1
fi

echo ""

# Check grpcserver specifically
if curl -s http://${REGISTRY_URL}/v2/_catalog | grep -q grpcserver; then
    echo "✓ grpcserver image pushed successfully"
    echo ""
    echo "Tags for grpcserver:"
    curl -s http://${REGISTRY_URL}/v2/grpcserver/tags/list | jq '.'
else
    echo "✗ grpcserver image not found in registry"
    exit 1
fi

echo ""
echo "========================================="
echo "✓ All images pushed successfully!"
echo "========================================="
echo ""
