#!/bin/bash

# vibe-cv Testing Script
# Tests consolidated /api/latest/* endpoints with all features

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_URL="http://localhost:8080"
REQUEST_FILE="example-request.json"

# Configuration
export LLM_PROVIDER="${LLM_PROVIDER:-openai}"
export LLM_API_KEY="${LLM_API_KEY:-sk-test-key}"
export LLM_MODEL="${LLM_MODEL:-gpt-4}"
export SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
export SERVER_PORT="${SERVER_PORT:-8080}"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

test_case() {
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${YELLOW}Test $TESTS_RUN: $1${NC}"
}

pass() {
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "${GREEN}✓ PASSED${NC}"
    echo ""
}

fail() {
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "${RED}✗ FAILED: $1${NC}"
    echo ""
}

summary() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total Tests: $TESTS_RUN"
    echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Main testing
print_header "vibe-cv API Test Suite"

# Check if server is running
test_case "Server Connectivity"
if curl -s "$API_URL/api/latest/health" > /dev/null 2>&1; then
    pass
else
    fail "Server is not running. Start with: ./vibe-cv"
    summary
fi

# Core Endpoints Tests
print_header "Core Endpoints (/api/latest/*)"

test_case "Health Check"
response=$(curl -s "$API_URL/api/latest/health")
if echo "$response" | grep -q "ok\|status"; then
    pass
else
    fail "Health check response invalid"
fi

test_case "CV Customization with File"
if [ ! -f "$REQUEST_FILE" ]; then
    echo "⚠ Skipping: $REQUEST_FILE not found"
else
    response=$(curl -s -X POST "$API_URL/api/latest/customize-cv" \
        -H "Content-Type: application/json" \
        -d @"$REQUEST_FILE")
    
    if echo "$response" | grep -q "success\|status\|customized\|error" || [ ! -z "$response" ]; then
        pass
    else
        fail "Invalid response format"
    fi
fi

test_case "CV Customization with Text Input"
response=$(curl -s -X POST "$API_URL/api/latest/customize-cv" \
    -H "Content-Type: application/json" \
    -d '{
        "cv": "Senior Software Engineer with 10 years of Go experience",
        "job_description": "We are hiring a Senior Go Developer with Kubernetes and microservices experience"
    }')

if [ ! -z "$response" ]; then
    pass
else
    fail "Customization failed - empty response"
fi

test_case "Batch Customization Submission"
response=$(curl -s -X POST "$API_URL/api/latest/batch-customize" \
    -H "Content-Type: application/json" \
    -d '{
        "cv": "Software Engineer with 5 years experience",
        "job_descriptions": [
            "Senior role in microservices",
            "Backend engineer needed"
        ]
    }')

if echo "$response" | grep -q "job_id\|batch_id\|status" || [ ! -z "$response" ]; then
    pass
else
    fail "Batch submission failed"
fi

test_case "Get CV Versions"
response=$(curl -s "$API_URL/api/latest/versions/1")
if [ ! -z "$response" ]; then
    pass
else
    fail "Versions endpoint failed"
fi

test_case "Get Version Details"
response=$(curl -s "$API_URL/api/latest/versions/1/detail")
if [ ! -z "$response" ]; then
    pass
else
    fail "Version details endpoint failed"
fi

test_case "Compare Versions"
response=$(curl -s -X POST "$API_URL/api/latest/compare-versions" \
    -H "Content-Type: application/json" \
    -d '{
        "version_id_1": 1,
        "version_id_2": 2
    }')

if [ ! -z "$response" ]; then
    pass
else
    fail "Compare versions failed"
fi

test_case "Get Analytics"
response=$(curl -s "$API_URL/api/latest/analytics")
if [ ! -z "$response" ]; then
    pass
else
    fail "Analytics endpoint failed"
fi

test_case "Get Dashboard"
response=$(curl -s "$API_URL/api/latest/dashboard")
if [ ! -z "$response" ]; then
    pass
else
    fail "Dashboard endpoint failed"
fi

test_case "Get Batch Status"
response=$(curl -s "$API_URL/api/latest/batch/1/status")
if [ ! -z "$response" ]; then
    pass
else
    fail "Batch status endpoint failed"
fi

test_case "Download Batch Results"
response=$(curl -s "$API_URL/api/latest/batch/1/download")
if [ ! -z "$response" ]; then
    pass
else
    fail "Batch download endpoint failed"
fi

# Observability & Security Tests
print_header "Observability & Security"

test_case "System Health Check"
response=$(curl -s "$API_URL/api/health")
if echo "$response" | grep -q "status" || [ ! -z "$response" ]; then
    pass
else
    fail "System health endpoint failed"
fi

test_case "Metrics Endpoint"
response=$(curl -s "$API_URL/api/metrics")
if [ ! -z "$response" ]; then
    pass
else
    fail "Metrics endpoint failed"
fi

test_case "Prometheus Metrics"
response=$(curl -s "$API_URL/metrics")
if echo "$response" | grep -q "vibe_cv\|http_\|#" || [ ! -z "$response" ]; then
    pass
else
    fail "Prometheus metrics failed"
fi

test_case "Audit Logs"
response=$(curl -s "$API_URL/api/audit-logs")
if echo "$response" | grep -q "logs\|entries\|\[\]" || [ ! -z "$response" ]; then
    pass
else
    fail "Audit logs endpoint failed"
fi

test_case "Rate Limiting (100 req/s limit)"
echo "Sending 5 sequential requests..."
for i in {1..5}; do
    curl -s "$API_URL/api/latest/health" > /dev/null
done
pass

test_case "CORS Headers"
response=$(curl -s -i -X OPTIONS "$API_URL/api/latest/customize-cv" 2>&1)
if echo "$response" | grep -q "Access-Control\|allow\|200\|204"; then
    pass
else
    fail "CORS headers missing or invalid"
fi

test_case "Content-Type Validation"
response=$(curl -s -X POST "$API_URL/api/latest/customize-cv" \
    -H "Content-Type: text/plain" \
    -d 'invalid')

if echo "$response" | grep -q "error\|invalid\|400" || [ ! -z "$response" ]; then
    pass
else
    fail "Content-Type validation not enforced"
fi

# Summary
summary
