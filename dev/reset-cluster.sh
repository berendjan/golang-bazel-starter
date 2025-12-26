#!/bin/bash
# Script to tear down and recreate the k3s cluster from scratch

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "k3s Cluster Reset Script"
echo "========================================="
echo ""

# Step 1: Delete existing cluster
echo "Step 1: Deleting existing cluster (if present)..."
"${SCRIPT_DIR}/delete-cluster.sh"

echo ""

# Step 2: Create new cluster
echo "Step 2: Creating new cluster..."
"${SCRIPT_DIR}/create-cluster.sh"

echo ""

# Step 3: Initialize cluster with all resources
echo "Step 3: Deploying all resources..."
"${SCRIPT_DIR}/init-cluster.sh"

echo ""
echo "========================================="
echo "✓ Cluster reset and initialized complete!"
echo "========================================="
echo ""
