# Elang - Dependency Security Monitoring Platform

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![CI/CD](https://img.shields.io/github/actions/workflow/status/your-org/elang/ci.yml?branch=main&label=CI/CD)
![Docker Hub](https://img.shields.io/badge/Docker%20Hub-Available-2496ED?style=flat&logo=docker)

**A powerful dependency security monitoring platform that tracks, analyzes, and alerts on security vulnerabilities in your application dependencies.**

[Quick Start](DOCKER_QUICK_START.md) | [Documentation](README.md) | [Docker Hub Setup](DOCKER_HUB_SETUP.md)

</div>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Architecture](#-architecture)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [Configuration](#-configuration)
- [API Documentation](#-api-documentation)
- [Testing](#-testing)
- [Development](#-development)
- [Deployment](#-deployment)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)

---

## 🔍 Overview

Elang is a comprehensive dependency security monitoring platform designed to help development teams maintain secure applications by:

- **Tracking dependencies** across multiple applications and runtimes (Node.js, Python, Go, Java, PHP, Ruby, Rust, .NET)
- **Scanning for vulnerabilities** using OSV (Open Source Vulnerabilities) database
- **Monitoring changes** in upstream dependencies with automated polling
- **Analyzing SBOM** (Software Bill of Materials) files
- **Alerting teams** about security issues through integrated messaging

### Supported Languages & Package Managers

| Language | Package Managers |
|----------|------------------|
| Node.js  | npm, yarn, pnpm |
| Python   | pip (requirements.txt, Pipfile) |
| Go       | go.mod |
| Java     | Maven (pom.xml), Gradle |
| PHP      | Composer |
| Ruby     | Bundler (Gemfile) |
| Rust     | Cargo |
| .NET     | NuGet |

---

## ✨ Features

### Core Features

- 🔐 **Security Scanning** - Automated OSV vulnerability scanning
- 📊 **SBOM Analysis** - Parse and analyze Software Bill of Materials
- 🔄 **Continuous Monitoring** - Track dependency changes and new releases
- 📦 **Multi-Runtime Support** - Support for 8+ programming languages
- 🎯 **Smart Detection** - Enhanced security detection with progressive monitoring
- 📝 **Audit Trail** - Complete history of scans and detections
- 💾 **Object Storage** - MinIO integration for SBOM file storage
- 🔔 **Notifications** - Telegram integration for security alerts

### Advanced Features

- **Tag Monitoring** - Track new version tags in dependencies
- **Commit Monitoring** - Monitor specific commits for security updates
- **Progressive Monitoring** - Gradually increase monitoring frequency for high-risk dependencies
- **Batch Operations** - Add/update/remove multiple dependencies at once
- **Flexible Scanning** - Manual and scheduled scanning options

---

## 🏗️ Architecture

```
┌─────────────────┐
│   Client/UI     │
└────────┬────────┘
         │ REST API
┌────────▼────────────────────────────────────┐
│         Elang Backend (Go)                  │
│  ┌──────────────────────────────────────┐  │
│  │  HTTP Handlers (Gin Framework)       │  │
│  └──────────────┬───────────────────────┘  │
│                 │                           │
│  ┌──────────────▼───────────────────────┐  │
│  │  Services Layer                      │  │
│  │  - Application Service               │  │
│  │  - Dependencies Service              │  │
│  └──────────────┬───────────────────────┘  │
│                 │                           │
│  ┌──────────────▼───────────────────────┐  │
│  │  Use Cases                           │  │
│  │  - GitHub API Integration            │  │
│  │  - MinIO Storage                     │  │
│  │  - Messaging (Telegram)              │  │
│  └──────────────┬───────────────────────┘  │
│                 │                           │
│  ┌──────────────▼───────────────────────┐  │
│  │  Repository Layer (GORM)             │  │
│  └──────────────────────────────────────┘  │
└────────┬─────────────────┬─────────────────┘
         │                 │
┌────────▼────────┐  ┌────▼──────────┐
│   PostgreSQL    │  │  MinIO Cluster│
│   Database      │  │  (S3-compat)  │
└─────────────────┘  └───────────────┘
```

---

## 📦 Prerequisites

Before you begin, ensure you have the following installed:

- **Docker** >= 20.10
- **Docker Compose** >= 2.0
- **Go** >= 1.23 (for local development)
- **Newman** (for API testing) - `npm install -g newman`

---

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/elang.git
cd elang
```

### 2. Set Up Environment Variables

Create a `.env` file in the root directory:

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=elang_user
DB_PASSWORD=secure_password_here
DB_NAME=elang_db
DB_SSLMODE=disable

# MinIO Configuration
MINIO_ENDPOINT=nginx:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=elang-sbom

# Application Configuration
APP_PORT=8080
GIN_MODE=release

# GitHub API (Optional - for enhanced features)
GITHUB_TOKEN=your_github_token_here

# Telegram Notifications (Optional)
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
```

### 3. Start the Services

```bash
# Start all services (PostgreSQL, MinIO, Elang Backend)
docker-compose up -d

# Check logs
docker-compose logs -f elang-backend

# Check service health
curl http://localhost:8080/health
```

### 4. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {
#   "status": "healthy",
#   "service": "elang-v1",
#   "version": "1.0",
#   "features": {
#     "enhanced_security_detection": true,
#     "progressive_monitoring": true,
#     "tag_monitoring": true,
#     "commit_monitoring": true
#   }
# }
```

---

## ⚙️ Configuration

### Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | PostgreSQL host | `postgres` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | Yes |
| `DB_USER` | Database user | - | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `DB_NAME` | Database name | - | Yes |
| `DB_SSLMODE` | SSL mode | `disable` | Yes |
| `MINIO_ENDPOINT` | MinIO endpoint | `nginx:9000` | Yes |
| `MINIO_ACCESS_KEY` | MinIO access key | - | Yes |
| `MINIO_SECRET_KEY` | MinIO secret key | - | Yes |
| `MINIO_BUCKET` | Bucket name | `elang-sbom` | Yes |
| `APP_PORT` | Application port | `8080` | Yes |
| `GITHUB_TOKEN` | GitHub API token | - | No |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token | - | No |
| `TELEGRAM_CHAT_ID` | Telegram chat ID | - | No |

---

## 📚 API Documentation

### Base URL

```
http://localhost:8080/api
```

### Authentication

Currently, the API does not require authentication. Add your authentication middleware as needed.

### Endpoints

#### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "elang-v1",
  "version": "1.0"
}
```

#### Application Management

##### Add Application

```http
POST /api/applications/add
Content-Type: multipart/form-data
```

**Parameters:**
- `app_name` (string): Application name
- `runtime_type` (string): Runtime (nodejs, python, go, java, php, ruby, rust, dotnet)
- `framework` (string): Framework name (optional)
- `description` (string): Description (optional)
- `file` (file): SBOM or dependency file (package.json, requirements.txt, go.mod, etc.)

**Response:**
```json
{
  "status": "success",
  "message": "application added successfully",
  "data": {
    "app_id": "uuid",
    "app_name": "my-app",
    "dependencies_count": 42
  }
}
```

##### List Applications

```http
GET /api/applications/list
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "uuid",
      "name": "my-app",
      "runtime": "nodejs",
      "framework": "express",
      "created_at": "2025-10-16T10:00:00Z"
    }
  ]
}
```

##### Get Application Dependencies

```http
GET /api/applications/:app_id/list
```

##### Remove Application

```http
DELETE /api/applications/:app_id/remove
```

##### Recover Application

```http
PATCH /api/applications/:app_id/recover
```

#### Dependency Management

##### Add Dependencies

```http
POST /api/applications/add/dependencies
Content-Type: application/json
```

**Body:**
```json
{
  "app_id": "uuid",
  "dependencies": [
    {
      "name": "express",
      "version": "4.18.0",
      "github_url": "https://github.com/expressjs/express"
    }
  ]
}
```

##### Update Dependencies

```http
PATCH /api/applications/update/dependencies
Content-Type: application/json
```

**Body:**
```json
{
  "app_id": "uuid",
  "dependencies": [
    {
      "dependency_id": "uuid",
      "version": "4.19.0"
    }
  ]
}
```

##### Remove Dependencies

```http
PATCH /api/applications/remove/dependencies
Content-Type: application/json
```

#### Security Scanning

##### Scan Application

```http
GET /api/applications/:app_id/scan
```

##### Scan Dependencies (Manual)

```http
POST /api/scan/dependencies
Content-Type: multipart/form-data
```

**Parameters:**
- `file` (file): Dependency file
- `runtime` (string): Runtime type
- `framework` (string): Framework (optional)

##### Get SBOM

```http
GET /api/scan/dependencies/:app_name/:sbom_id
```

#### Monitoring

##### Start Monitoring

```http
POST /api/scan/:app_id/start
```

##### Stop Monitoring

```http
POST /api/scan/:app_id/stop
```

##### Get Monitoring Status

```http
GET /api/scan/:app_id/status
```

---

## 🧪 Testing

### Run Newman Tests

We provide comprehensive Postman/Newman collections for API testing.

```bash
# Install Newman (if not already installed)
npm install -g newman

# Run all tests
npm test

# Or run directly with Newman
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --reporters cli,html \
  --reporter-html-export newman-report.html
```

### Run Go Tests

```bash
cd backend
make test

# Or manually
go test ./test/... -v
```

---

## 💻 Development

### Local Development Setup

```bash
# Navigate to backend
cd backend

# Install dependencies
go mod download

# Copy example env file
cp ../.env.example ../.env

# Edit .env with your local settings
# Set DB_HOST=localhost for local PostgreSQL

# Run the application
go run cmd/main.go

# Or use Make
make run
```

### Build from Source

```bash
cd backend
make build

# Binary will be in ./bin/elang-app
./bin/elang-app
```

### Code Structure

```
backend/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   ├── delivery/http/          # HTTP handlers & routing
│   ├── entity/                 # Database entities
│   ├── helper/                 # Helper functions & parsers
│   ├── model/                  # Request/Response models
│   ├── repository/             # Data access layer
│   ├── services/               # Business logic
│   └── usecase/                # Use case implementations
└── test/                       # Unit tests
```

---

## 🐳 Deployment

### Docker Deployment

```bash
# Build the image
docker build -t elang-backend:latest .

# Run with docker-compose
docker-compose up -d
```

### Production Considerations

1. **Security**
   - Use strong database passwords
   - Enable SSL for PostgreSQL (`DB_SSLMODE=require`)
   - Use HTTPS/TLS for the API
   - Implement authentication middleware
   - Restrict CORS origins

2. **Performance**
   - Adjust `polling_interval_minutes` based on your needs
   - Configure MinIO with appropriate storage
   - Use connection pooling for database
   - Monitor resource usage

3. **Monitoring**
   - Set up health check monitoring
   - Configure log aggregation
   - Monitor MinIO storage capacity
   - Track API response times

---

## 🔧 Troubleshooting

### Common Issues

**Issue: Cannot connect to database**
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Verify connection
docker-compose exec postgres psql -U elang_user -d elang_db
```

**Issue: MinIO connection failed**
```bash
# Check MinIO cluster health
curl http://localhost:9000/minio/health/live

# Check nginx proxy
docker-compose logs nginx
```

**Issue: Application won't start**
```bash
# Check application logs
docker-compose logs elang-backend

# Verify environment variables
docker-compose exec elang-backend env | grep DB_
```

### Debug Mode

Set `GIN_MODE=debug` in your `.env` file for verbose logging.

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Write unit tests for new features
- Update documentation as needed
- Use meaningful commit messages

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [OSV](https://osv.dev/) - Open Source Vulnerabilities database
- [Gin](https://gin-gonic.com/) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [MinIO](https://min.io/) - Object storage

---

## 📞 Support

For support, please open an issue in the GitHub repository or contact the development team.

---

<div align="center">
Made with ❤️ by the Elang Team
</div>
