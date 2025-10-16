#!/bin/bash

# Elang API Test Report Generator
# Generates detailed HTML report with htmlextra reporter

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo -e "${GREEN}📊 Elang API Test Report Generator${NC}"
echo "===================================="
echo ""

# Check if htmlextra is installed
if ! npm list -g newman-reporter-htmlextra &> /dev/null; then
    echo -e "${YELLOW}⚠ newman-reporter-htmlextra is not installed${NC}"
    echo "Installing..."
    npm install -g newman-reporter-htmlextra
fi

echo "🔍 Running tests with detailed reporting..."
echo ""

cd "$PROJECT_ROOT"

# Generate timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="test-report-${TIMESTAMP}.html"

newman run tests/Elang_API_Tests.postman_collection.json \
    -e tests/Elang_Environment.postman_environment.json \
    --reporters cli,htmlextra \
    --reporter-htmlextra-export "reports/${REPORT_FILE}" \
    --reporter-htmlextra-title "Elang API Test Report" \
    --reporter-htmlextra-showOnlyFails false \
    --reporter-htmlextra-darkTheme true \
    --color on \
    --delay-request 500

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Report generated successfully!${NC}"
    echo ""
    echo -e "📄 Report location: ${GREEN}reports/${REPORT_FILE}${NC}"
    echo ""
    echo "🌐 Open in browser:"
    echo "   open reports/${REPORT_FILE}"
    echo ""
else
    echo ""
    echo -e "${RED}❌ Test execution failed${NC}"
    echo ""
fi
