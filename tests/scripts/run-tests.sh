#!/bin/bash

# Elang API Test Runner
# This script runs the complete API test suite using Newman

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo -e "${GREEN}🚀 Elang API Test Suite${NC}"
echo "================================"
echo ""

# Check if Newman is installed
if ! command -v newman &> /dev/null; then
    echo -e "${RED}❌ Newman is not installed${NC}"
    echo "Install it with: npm install -g newman"
    exit 1
fi

echo -e "${GREEN}✓${NC} Newman found: $(newman --version)"
echo ""

# Check if API is running
echo "🔍 Checking if API is running..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} API is running"
else
    echo -e "${YELLOW}⚠${NC} API is not responding at http://localhost:8080"
    echo "Starting services with docker-compose..."
    cd "$PROJECT_ROOT"
    docker-compose up -d
    echo "Waiting 30 seconds for services to start..."
    sleep 30
fi

echo ""
echo "================================"
echo "🧪 Running Tests..."
echo "================================"
echo ""

# Run Newman tests
cd "$PROJECT_ROOT"

newman run tests/Elang_API_Tests.postman_collection.json \
    -e tests/Elang_Environment.postman_environment.json \
    --reporters cli,html \
    --reporter-html-export test-results.html \
    --color on \
    --delay-request 500 \
    --timeout-request 10000

# Check exit code
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo "📊 HTML report generated: test-results.html"
    echo ""
    exit 0
else
    echo ""
    echo -e "${RED}================================${NC}"
    echo -e "${RED}❌ Some tests failed${NC}"
    echo -e "${RED}================================${NC}"
    echo ""
    echo "📊 Check test-results.html for details"
    echo ""
    exit 1
fi
