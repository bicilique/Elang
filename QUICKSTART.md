# Elang - Quick Reference Guide

## 🚀 Quick Start Commands

```bash
# 1. Clone and setup
git clone <repository-url>
cd elang
cp .env.example .env

# 2. Edit .env with your configuration
nano .env

# 3. Start all services
docker-compose up -d

# 4. Check health
curl http://localhost:8080/health

# 5. View logs
docker-compose logs -f elang-backend
```

## 📝 Common Commands

### Docker Management
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# Rebuild and start
docker-compose up -d --build

# View logs
docker-compose logs -f [service-name]

# Check service status
docker-compose ps

# Restart a service
docker-compose restart elang-backend
```

### Database Management
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U elang_user -d elang_db

# View tables
docker-compose exec postgres psql -U elang_user -d elang_db -c "\dt"

# Backup database
docker-compose exec postgres pg_dump -U elang_user elang_db > backup.sql

# Restore database
docker-compose exec -T postgres psql -U elang_user elang_db < backup.sql
```

### MinIO Management
```bash
# Access MinIO console
# Open browser: http://localhost:9001
# Username: minioadmin
# Password: minioadmin

# List buckets (from container)
docker-compose exec minio1 mc ls myminio
```

### Testing
```bash
# Install test dependencies
npm install

# Run all tests
npm test

# Run with verbose output
npm run test:verbose

# Generate detailed report
npm run test:report

# Run specific test folder
newman run tests/Elang_API_Tests.postman_collection.json \
  --folder "Health Check"
```

### Development
```bash
# Run backend locally
cd backend
go run cmd/main.go

# Run tests
cd backend
go test ./test/... -v

# Build binary
cd backend
make build
```

## 🔧 Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs elang-backend

# Check if ports are in use
lsof -i :8080
lsof -i :5432
lsof -i :9000

# Restart services
docker-compose restart
```

### Database connection issues
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Test connection
docker-compose exec postgres pg_isready -U elang_user

# Verify environment variables
docker-compose exec elang-backend env | grep DB_
```

### API returns errors
```bash
# Check backend logs
docker-compose logs -f elang-backend

# Verify health endpoint
curl -v http://localhost:8080/health

# Check database schema
docker-compose exec postgres psql -U elang_user -d elang_db -c "\dt"
```

## 📊 API Quick Reference

### Health Check
```bash
curl http://localhost:8080/health
```

### Add Application
```bash
curl -X POST http://localhost:8080/api/applications/add \
  -F "app_name=my-app" \
  -F "runtime_type=nodejs" \
  -F "framework=express" \
  -F "description=My application" \
  -F "file=@package.json"
```

### List Applications
```bash
curl http://localhost:8080/api/applications/list
```

### Scan Application
```bash
curl http://localhost:8080/api/applications/{app_id}/scan
```

### Start Monitoring
```bash
curl -X POST http://localhost:8080/api/scan/{app_id}/start
```

## 🔐 Security Checklist

- [ ] Change default database password in `.env`
- [ ] Change MinIO credentials in `.env`
- [ ] Set up GitHub token for dependency tracking
- [ ] Configure Telegram bot for notifications
- [ ] Enable SSL/TLS for production
- [ ] Set `GIN_MODE=release` for production
- [ ] Implement authentication middleware
- [ ] Configure CORS properly
- [ ] Regular security updates
- [ ] Monitor logs for suspicious activity

## 📁 Important Files

| File | Purpose |
|------|---------|
| `.env` | Environment configuration |
| `docker-compose.yaml` | Service orchestration |
| `Dockerfile` | Backend container build |
| `schema.sql` | Database schema |
| `package.json` | Test dependencies |
| `tests/` | API test suite |

## 🔗 Useful URLs

- **API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001
- **PostgreSQL**: localhost:5432
- **API Docs**: See README.md

## 💡 Tips

1. Always check health endpoint first
2. Use docker-compose logs for debugging
3. Keep .env file secure and never commit it
4. Run tests after code changes
5. Monitor MinIO storage usage
6. Backup database regularly
7. Use meaningful application names
8. Check monitoring status periodically

---

For detailed documentation, see [README.md](README.md)
For testing guide, see [tests/README.md](tests/README.md)
