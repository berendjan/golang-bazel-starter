#!/bin/bash
# Script to build and push application images to the registry

set -e  # Exit on error

REGISTRY_URL="registry.localhost:5001"

echo "========================================="
echo "Building and Pushing Images"
echo "Registry: ${REGISTRY_URL}"
echo "========================================="
echo ""

# Check if registry is accessible
echo "Checking registry availability at ${REGISTRY_URL}..."
if ! curl -s http://${REGISTRY_URL}/v2/ > /dev/null; then
    echo "ERROR: Registry not accessible at ${REGISTRY_URL}"
    echo ""
    echo "Troubleshooting steps:"
    echo "  1. Check registry container is running: docker ps | grep registry-dev"
    echo "  2. Verify /etc/hosts has: 127.0.0.1 registry.localhost"
    echo "  3. Test manually: curl http://registry.localhost:5001/v2/"
    exit 1
fi

echo "✓ Registry is accessible"
echo ""

# Build and push dbmate
echo "Building and pushing dbmate..."
echo "-----------------------------------"
bazel run //db/config:dbmate_config_push
bazel run //db/auth:dbmate_auth_push

echo ""

# Build and push grpcserver
echo "Building and pushing grpcserver..."
echo "-----------------------------------"
bazel run //golang/grpcserver:grpcserver_push

echo ""

# Build and push frontend
echo "Building and pushing frontend..."
echo "-----------------------------------"
bazel run //frontend:frontend_push

echo ""
echo "Verifying images in registry..."
echo "-----------------------------------"

# List all images in registry
echo "Images in registry:"
curl -s http://${REGISTRY_URL}/v2/_catalog | jq '.'

echo ""

# Check dbmate specifically
if curl -s http://${REGISTRY_URL}/v2/_catalog | grep -q dbmate; then
    echo "✓ dbmate image pushed successfully"
    echo ""
    echo "Tags for dbmate:"
    curl -s http://${REGISTRY_URL}/v2/dbmate/tags/list | jq '.'
else
    echo "✗ dbmate image not found in registry"
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

# Check frontend specifically
if curl -s http://${REGISTRY_URL}/v2/_catalog | grep -q frontend; then
    echo "✓ frontend image pushed successfully"
    echo ""
    echo "Tags for frontend:"
    curl -s http://${REGISTRY_URL}/v2/frontend/tags/list | jq '.'
else
    echo "✗ frontend image not found in registry"
    exit 1
fi

echo ""
echo "========================================="
echo "✓ All images pushed successfully!"
echo "========================================="
echo ""
