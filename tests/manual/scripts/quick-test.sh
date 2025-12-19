#!/bin/bash
# Quick test script for k8t with minikube

set -e

echo "=== k8t Quick Test Script ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo "Checking prerequisites..."
command -v minikube >/dev/null 2>&1 || { echo -e "${RED}Error: minikube not found${NC}"; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo -e "${RED}Error: kubectl not found${NC}"; exit 1; }

# Check if k8t binary exists
K8T_BIN="./bin/k8t"
if [ ! -f "$K8T_BIN" ]; then
    echo -e "${YELLOW}k8t binary not found, building...${NC}"
    make build
fi

# Start minikube if not running
echo ""
echo "Checking minikube status..."
if ! minikube status >/dev/null 2>&1; then
    echo -e "${YELLOW}Starting minikube...${NC}"
    minikube start
else
    echo -e "${GREEN}Minikube already running${NC}"
fi

# Deploy test pods
echo ""
echo "Deploying test pods..."
kubectl apply -f tests/manual/manifests/

# Wait a bit for pods to start pulling images
echo ""
echo "Waiting 30 seconds for ImagePullBackOff to occur..."
sleep 30

# Show pod status
echo ""
echo "Pod status:"
kubectl get pods -l app=k8t-test

# Test each scenario
echo ""
echo "=== Running k8t Tests ==="
echo ""

test_pod() {
    local pod_name=$1
    local expected_cause=$2

    echo -e "${YELLOW}Testing: $pod_name${NC}"
    echo "Expected root cause: $expected_cause"

    # Run k8t
    if $K8T_BIN analyze imagepullbackoff "$pod_name" 2>&1; then
        echo -e "${GREEN}✓ Test passed${NC}"
    else
        echo -e "${RED}✗ Test failed${NC}"
    fi

    echo ""
    echo "---"
    echo ""
}

# Test image not found
test_pod "test-image-not-found" "IMAGE_NOT_FOUND"

# Test auth failure
test_pod "test-auth-failure" "AUTHENTICATION_FAILURE"

# Test network issue
test_pod "test-network-issue" "NETWORK_ISSUE"

# Test success case (should show no issues)
echo -e "${YELLOW}Testing: test-success (should have no ImagePullBackOff)${NC}"
$K8T_BIN analyze imagepullbackoff test-success 2>&1 || true
echo ""

echo "=== Tests Complete ==="
echo ""
echo "To clean up: kubectl delete -f tests/manual/manifests/"
echo "To stop minikube: minikube stop"
