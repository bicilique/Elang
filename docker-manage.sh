#!/bin/bash

# Elang Docker Management Script

set -e

BACKEND_ENV="backend/.env"

print_help() {
    echo "Elang Docker Management"
    echo ""
    echo "Usage: ./docker-manage.sh [command]"
    echo ""
    echo "Commands:"
    echo "  start       - Start all services using Docker Compose"
    echo "  stop        - Stop all services"
    echo "  restart     - Restart all services"
    echo "  logs        - View logs (all services)"
    echo "  logs-app    - View application logs only"
    echo "  status      - Check status of all services"
    echo "  pull        - Pull latest Docker image from Docker Hub"
    echo "  reset       - Stop and remove all containers and volumes (WARNING: deletes data!)"
    echo "  setup-env   - Set up environment for Docker (updates DB_HOST and STORAGE_ENDPOINT)"
    echo "  local-env   - Set up environment for local development"
    echo "  help        - Show this help message"
    echo ""
}

setup_docker_env() {
    echo "Setting up environment for Docker..."
    
    if [ ! -f "$BACKEND_ENV" ]; then
        echo "Error: $BACKEND_ENV not found!"
        exit 1
    fi
    
    # Backup current .env
    cp "$BACKEND_ENV" "${BACKEND_ENV}.backup"
    
    # Update DB_HOST to postgres
    sed -i.tmp 's/^DB_HOST=.*/DB_HOST=postgres/' "$BACKEND_ENV"
    
    # Update STORAGE_ENDPOINT to nginx:9000
    sed -i.tmp 's/^STORAGE_ENDPOINT=.*/STORAGE_ENDPOINT=nginx:9000/' "$BACKEND_ENV"
    
    # Remove temporary files
    rm -f "${BACKEND_ENV}.tmp"
    
    echo "✓ Environment configured for Docker"
    echo "  DB_HOST=postgres"
    echo "  STORAGE_ENDPOINT=nginx:9000"
    echo "  Backup saved to ${BACKEND_ENV}.backup"
}

setup_local_env() {
    echo "Setting up environment for local development..."
    
    if [ ! -f "$BACKEND_ENV" ]; then
        echo "Error: $BACKEND_ENV not found!"
        exit 1
    fi
    
    # Update DB_HOST to localhost
    sed -i.tmp 's/^DB_HOST=.*/DB_HOST=localhost/' "$BACKEND_ENV"
    
    # Update STORAGE_ENDPOINT to localhost:9000
    sed -i.tmp 's/^STORAGE_ENDPOINT=.*/STORAGE_ENDPOINT=localhost:9000/' "$BACKEND_ENV"
    
    # Remove temporary files
    rm -f "${BACKEND_ENV}.tmp"
    
    echo "✓ Environment configured for local development"
    echo "  DB_HOST=localhost"
    echo "  STORAGE_ENDPOINT=localhost:9000"
}

case "${1:-help}" in
    start)
        echo "Starting Elang services..."
        docker-compose up -d
        echo ""
        echo "✓ Services started!"
        echo "  Backend: http://localhost:8080"
        echo "  MinIO Console: http://localhost:9001"
        echo ""
        echo "Run './docker-manage.sh logs' to view logs"
        ;;
    
    stop)
        echo "Stopping Elang services..."
        docker-compose stop
        echo "✓ Services stopped"
        ;;
    
    restart)
        echo "Restarting Elang services..."
        docker-compose restart
        echo "✓ Services restarted"
        ;;
    
    logs)
        docker-compose logs -f
        ;;
    
    logs-app)
        docker-compose logs -f elang-backend
        ;;
    
    status)
        docker-compose ps
        ;;
    
    pull)
        echo "Pulling latest Docker image..."
        docker pull afiffaizianur/elang:latest
        echo "✓ Image pulled"
        echo "Run './docker-manage.sh restart' to use the new image"
        ;;
    
    reset)
        echo "⚠️  WARNING: This will delete all data!"
        read -p "Are you sure? (type 'yes' to confirm): " confirm
        if [ "$confirm" = "yes" ]; then
            echo "Resetting services..."
            docker-compose down -v
            echo "✓ All containers and volumes removed"
        else
            echo "Cancelled"
        fi
        ;;
    
    setup-env)
        setup_docker_env
        ;;
    
    local-env)
        setup_local_env
        ;;
    
    help|*)
        print_help
        ;;
esac
