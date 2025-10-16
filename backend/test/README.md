# Testing Guide

## Overview

This document provides comprehensive information about testing the Elang Backend application. Our testing strategy includes unit tests, integration tests, and CI/CD automation.

## Test Structure

```
backend/test/
├── main_test.go                          # Entry point and basic tests
├── .env.test                             # Test environment configuration
├── repository/                           # Repository layer tests
│   ├── application_repository_test.go
│   ├── dependency_repository_test.go
│   ├── runtime_repository_test.go
│   ├── framework_repository_test.go
│   └── app_dependency_repository_test.go
├── services/                             # Service layer tests
│   ├── application_service_test.go
│   └── dependencies_service_test.go
└── usecase/                              # Usecase layer tests
    ├── github_api_usecase_test.go
    └── minio_usecase_test.go
```

## Test Coverage Goals

- **Repository Layer**: 90%+ coverage
- **Service Layer**: 85%+ coverage
- **Usecase Layer**: 80%+ coverage
- **Overall**: 85%+ coverage

## Running Tests

### Quick Start

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detector
make test-race
```

### Detailed Commands

```bash
# Run all tests with verbose output
go test -v ./test/...

# Run specific test file
go test -v ./test/repository/application_repository_test.go

# Run specific test function
go test -v ./test/repository -run TestApplicationRepository_Create

# Run tests with coverage
go test ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
go test -race ./test/...

# Run tests in parallel
go test -parallel 4 ./test/...
```

### Layer-Specific Tests

```bash
# Repository tests only
make test-repo
# or
go test -v ./test/repository/...

# Service tests only
make test-service
# or
go test -v ./test/services/...

# Usecase tests only
make test-usecase
# or
go test -v ./test/usecase/...
```

## Test Types

### 1. Unit Tests

Unit tests verify individual components in isolation using mocks for dependencies.

**Example:**
```go
func TestApplicationRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := repository.NewAppRepository(db)
    ctx := context.Background()

    app := &entity.App{
        ID:     uuid.New(),
        Name:   "test-app",
        Status: "active",
    }

    err := repo.Create(ctx, app)
    assert.NoError(t, err)
}
```

### 2. Integration Tests

Integration tests verify interactions between multiple components using an in-memory SQLite database.

**Example:**
```go
func TestAppDependencyRepository_GetByAppID(t *testing.T) {
    db := setupTestDB(t)
    appRepo := repository.NewAppRepository(db)
    depRepo := repository.NewDependencyRepository(db)
    appDepRepo := repository.NewAppDependencyRepository(db)
    // ... test implementation
}
```

### 3. Mock Tests

Mock tests use testify/mock to simulate external dependencies.

**Example:**
```go
func TestApplicationService_ListApplications(t *testing.T) {
    mockAppRepo := new(MockApplicationRepository)
    mockAppRepo.On("GetAll", ctx).Return(expectedApps, nil)
    // ... test implementation
}
```

## Test Database Setup

We use SQLite in-memory database for testing to ensure:
- Fast test execution
- No external dependencies
- Clean state for each test

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    err = db.AutoMigrate(
        &entity.Runtime{},
        &entity.Framework{},
        &entity.App{},
        // ... other entities
    )
    require.NoError(t, err)
    
    return db
}
```

## Testing Best Practices

### 1. Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestRuntimeRepository_GetByNameCI(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected bool
    }{
        {"Lowercase", "node.js", true},
        {"Uppercase", "NODE.JS", true},
        {"NotFound", "python", false},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Subtests

Use subtests for related scenarios:

```go
func TestApplicationRepository_GetByID(t *testing.T) {
    t.Run("Found", func(t *testing.T) {
        // Test when record exists
    })
    
    t.Run("NotFound", func(t *testing.T) {
        // Test when record doesn't exist
    })
}
```

### 3. Test Isolation

Each test should be independent and not affect others:

```go
func TestExample(t *testing.T) {
    // Setup - create fresh database
    db := setupTestDB(t)
    
    // Test logic
    // ...
    
    // Teardown is automatic with in-memory DB
}
```

### 4. Mock Verification

Always verify mock expectations:

```go
func TestWithMock(t *testing.T) {
    mock := new(MockRepository)
    mock.On("Method", arg).Return(result, nil)
    
    // Call code under test
    
    mock.AssertExpectations(t) // Verify mock was called correctly
}
```

## Continuous Integration

### GitHub Actions

Our CI pipeline runs on every push and pull request:

```yaml
- Run tests with coverage
- Upload coverage reports
- Run race detector
- Build for multiple platforms
```

### GitLab CI

Similar pipeline for GitLab:

```yaml
stages:
  - test
  - build
  - deploy
```

### Jenkins

Jenkinsfile included for Jenkins users.

## Coverage Reports

### Generate Coverage

```bash
# Generate coverage report
go test ./test/... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Coverage by Package

```bash
go test ./test/... -cover
```

Example output:
```
ok      elang-backend/test              0.123s  coverage: 87.5% of statements
ok      elang-backend/test/repository   0.456s  coverage: 92.3% of statements
ok      elang-backend/test/services     0.234s  coverage: 85.1% of statements
ok      elang-backend/test/usecase      0.189s  coverage: 81.7% of statements
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./test/...
```

### Run Single Test

```bash
go test -v -run TestApplicationRepository_Create ./test/repository
```

### Print Statements

Use `t.Logf()` for debug output:

```go
func TestExample(t *testing.T) {
    result := someFunction()
    t.Logf("Result: %+v", result)
    assert.NotNil(t, result)
}
```

### Debug with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug specific test
dlv test ./test/repository -- -test.run TestApplicationRepository_Create
```

## Common Issues and Solutions

### Issue 1: Database Connection Errors

**Problem:** Tests fail with database connection errors

**Solution:** Ensure using in-memory SQLite for tests:
```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

### Issue 2: Flaky Tests

**Problem:** Tests pass/fail randomly

**Solution:**
- Ensure test isolation
- Avoid time-dependent tests
- Use mocks for external dependencies
- Run with race detector: `go test -race`

### Issue 3: Slow Tests

**Problem:** Tests take too long to run

**Solution:**
- Use in-memory database
- Run tests in parallel: `go test -parallel 4`
- Mock external services
- Use table-driven tests efficiently

### Issue 4: Import Cycles

**Problem:** Import cycle errors

**Solution:**
- Use separate test packages: `package repository_test`
- Import only what's needed
- Consider dependency injection

## Test Metrics

### Current Coverage

| Package | Coverage |
|---------|----------|
| Repository | 90%+ |
| Services | 85%+ |
| Usecase | 80%+ |
| Overall | 85%+ |

### Test Count

| Type | Count |
|------|-------|
| Repository Tests | 50+ |
| Service Tests | 30+ |
| Usecase Tests | 20+ |
| Total | 100+ |

## Contributing Tests

When adding new features:

1. **Write tests first** (TDD approach)
2. **Ensure coverage** stays above 85%
3. **Follow naming conventions**: `Test<Type>_<Method>_<Scenario>`
4. **Add table-driven tests** for multiple scenarios
5. **Mock external dependencies**
6. **Document complex test logic**

### Test Naming Convention

```
Test<Component>_<Method>_<Scenario>

Examples:
- TestApplicationRepository_Create
- TestApplicationService_GetByID_Success
- TestApplicationService_GetByID_NotFound
- TestDependencyRepository_GetByNameCI
```

## Performance Testing

### Benchmark Tests

```go
func BenchmarkApplicationRepository_Create(b *testing.B) {
    db := setupTestDB(nil)
    repo := repository.NewAppRepository(db)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        app := &entity.App{
            ID:     uuid.New(),
            Name:   "test-app",
            Status: "active",
        }
        repo.Create(ctx, app)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./test/...
go test -bench=BenchmarkApplicationRepository_Create ./test/repository
```

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [GORM Testing Guide](https://gorm.io/docs/testing.html)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

## Questions?

For questions about testing, please:
1. Check this documentation
2. Review existing tests for examples
3. Ask in team chat
4. Open an issue on GitHub
