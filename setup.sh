#!/bin/bash

# Elang Setup Script
# This script helps you set up the Elang platform

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                                                           ║"
echo "║           🦅 Elang Platform Setup Script 🦅              ║"
echo "║                                                           ║"
echo "║     Dependency Security Monitoring Platform Setup        ║"
echo "║                                                           ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"
echo ""

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✓${NC} Docker found: $(docker --version)"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose from https://docs.docker.com/compose/install/"
    exit 1
fi
echo -e "${GREEN}✓${NC} Docker Compose found: $(docker-compose --version)"

echo ""
echo -e "${YELLOW}Setting up environment files...${NC}"
echo ""

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✓${NC} Created .env from .env.example"
        echo -e "${YELLOW}⚠${NC}  Please edit .env file and update the configuration"
        echo ""
        echo "Important settings to configure:"
        echo "  - DB_PASSWORD: Change the default database password"
        echo "  - GITHUB_TOKEN: Add your GitHub token (optional)"
        echo "  - TELEGRAM_BOT_TOKEN: Add your Telegram bot token (optional)"
        echo ""
        read -p "Press Enter to continue or Ctrl+C to exit and edit .env first..."
    else
        echo -e "${RED}❌ .env.example not found${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓${NC} .env file already exists"
fi

echo ""
echo -e "${YELLOW}Starting services...${NC}"
echo ""

# Pull images first
echo "📥 Pulling Docker images..."
docker-compose pull

echo ""
echo "🏗️  Building backend application..."
docker-compose build elang-backend

echo ""
echo "🚀 Starting all services..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be ready..."
echo "   This may take 30-60 seconds..."

# Wait for health checks
MAX_RETRIES=30
RETRY_COUNT=0
SLEEP_TIME=2

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo ""
        echo -e "${GREEN}✅ All services are ready!${NC}"
        break
    fi
    
    echo -n "."
    sleep $SLEEP_TIME
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

echo ""
echo ""

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}❌ Services did not start in time${NC}"
    echo "Check logs with: docker-compose logs"
    exit 1
fi

# Test health endpoint
echo -e "${YELLOW}Testing API health endpoint...${NC}"
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                                                           ║${NC}"
echo -e "${GREEN}║              🎉 Setup Complete! 🎉                        ║${NC}"
echo -e "${GREEN}║                                                           ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Services running:${NC}"
echo ""
echo "  🌐 Elang API:        http://localhost:8080"
echo "  🗄️  MinIO Console:    http://localhost:9001"
echo "  🐘 PostgreSQL:       localhost:5432"
echo ""
echo -e "${BLUE}Quick commands:${NC}"
echo ""
echo "  📊 View logs:        docker-compose logs -f elang-backend"
echo "  🔍 Check status:     docker-compose ps"
echo "  🛑 Stop services:    docker-compose down"
echo "  🧪 Run tests:        npm test"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo ""
echo "  1. Test the API:     curl http://localhost:8080/health"
echo "  2. Read the docs:    cat README.md"
echo "  3. Run tests:        npm install && npm test"
echo "  4. Check MinIO:      Open http://localhost:9001 (admin/minioadmin)"
echo ""
echo -e "${YELLOW}For more information, see:${NC}"
echo "  - README.md for full documentation"
echo "  - QUICKSTART.md for quick reference"
echo "  - tests/README.md for testing guide"
echo ""
echo -e "${GREEN}Happy monitoring! 🚀${NC}"
echo ""
