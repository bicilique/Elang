.PHONY: help setup up down restart logs test test-verbose clean build rebuild ps health

# Default target
help:
	@echo ""
	@echo "🦅 Elang - Dependency Security Monitoring Platform"
	@echo "=================================================="
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  make setup          - Initial setup (creates .env, starts services)"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo "  make logs           - View all logs (follow mode)"
	@echo "  make logs-backend   - View backend logs only"
	@echo "  make ps             - Show service status"
	@echo "  make health         - Check API health"
	@echo "  make build          - Build backend image"
	@echo "  make rebuild        - Rebuild and restart services"
	@echo "  make test           - Run API tests with Newman"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-report    - Generate detailed HTML test report"
	@echo "  make clean          - Remove containers, volumes, and images"
	@echo "  make clean-all      - Remove everything including test reports"
	@echo "  make db-shell       - Connect to PostgreSQL shell"
	@echo "  make db-backup      - Backup database to backup.sql"
	@echo "  make minio-console  - Open MinIO console in browser"
	@echo ""

# Setup environment
setup:
	@echo "🚀 Running setup script..."
	@bash setup.sh

# Start services
up:
	@echo "🚀 Starting services..."
	@docker-compose up -d
	@echo "✓ Services started"
	@make health

# Stop services
down:
	@echo "🛑 Stopping services..."
	@docker-compose down
	@echo "✓ Services stopped"

# Restart services
restart:
	@echo "🔄 Restarting services..."
	@docker-compose restart
	@echo "✓ Services restarted"

# View logs
logs:
	@docker-compose logs -f

# View backend logs only
logs-backend:
	@docker-compose logs -f elang-backend

# Show service status
ps:
	@docker-compose ps

# Check API health
health:
	@echo "🏥 Checking API health..."
	@sleep 2
	@curl -s http://localhost:8080/health | jq '.' || echo "❌ API not responding"

# Build backend
build:
	@echo "🏗️  Building backend..."
	@docker-compose build elang-backend
	@echo "✓ Backend built"

# Rebuild and restart
rebuild: build
	@echo "🔄 Restarting backend..."
	@docker-compose up -d elang-backend
	@echo "✓ Backend restarted"

# Run tests
test:
	@echo "🧪 Running API tests..."
	@npm test

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running API tests (verbose)..."
	@npm run test:verbose

# Generate detailed test report
test-report:
	@echo "📊 Generating test report..."
	@bash tests/scripts/test-report.sh

# Clean up
clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@docker rmi elang-backend:latest 2>/dev/null || true
	@echo "✓ Cleanup complete"

# Clean everything including test reports
clean-all: clean
	@echo "🧹 Removing test reports..."
	@rm -f test-results.html test-results.json
	@rm -rf reports/
	@echo "✓ All cleaned"

# Database shell
db-shell:
	@echo "🐘 Connecting to PostgreSQL..."
	@docker-compose exec postgres psql -U elang_user -d elang_db

# Backup database
db-backup:
	@echo "💾 Backing up database..."
	@docker-compose exec postgres pg_dump -U elang_user elang_db > backup.sql
	@echo "✓ Database backed up to backup.sql"

# Open MinIO console
minio-console:
	@echo "🗄️  Opening MinIO console..."
	@open http://localhost:9001 || xdg-open http://localhost:9001 || echo "Please open http://localhost:9001"

# Install dependencies
install:
	@echo "📦 Installing test dependencies..."
	@npm install
	@echo "✓ Dependencies installed"

# Development mode (hot reload)
dev:
	@echo "🔧 Starting in development mode..."
	@docker-compose up

# Show environment info
env-info:
	@echo "🔍 Environment Information:"
	@echo ""
	@docker-compose exec elang-backend env | grep -E "^(DB_|MINIO_|APP_|GIN_)" || true
