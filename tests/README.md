# Elang API Testing Guide

This directory contains comprehensive API tests for the Elang platform using Newman (Postman CLI).

## 📁 Directory Structure

```
tests/
├── README.md                                    # This file
├── Elang_API_Tests.postman_collection.json      # Main test collection
├── Elang_Environment.postman_environment.json   # Environment variables
├── sample-files/                                # Sample dependency files
│   ├── package.json                             # Node.js sample
│   ├── requirements.txt                         # Python sample
│   └── go.mod                                   # Go sample
└── scripts/
    ├── run-tests.sh                             # Bash script to run all tests
    └── test-report.sh                           # Generate detailed HTML report
```

## 🚀 Quick Start

### Prerequisites

Install Newman globally:

```bash
npm install -g newman
npm install -g newman-reporter-htmlextra  # Optional: for better HTML reports
```

### Running Tests

#### Option 1: Using npm (Recommended)

```bash
# From project root
npm test
```

#### Option 2: Using Newman directly

```bash
# Run all tests with CLI output
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json

# Run with HTML report
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --reporters cli,html \
  --reporter-html-export test-results.html

# Run with detailed HTML report (requires newman-reporter-htmlextra)
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --reporters cli,htmlextra \
  --reporter-htmlextra-export test-results-detailed.html
```

#### Option 3: Using the provided script

```bash
chmod +x tests/scripts/run-tests.sh
./tests/scripts/run-tests.sh
```

## 📋 Test Collection Overview

The test collection includes the following test suites:

### 1. Health Check
- ✅ Verify API is running
- ✅ Check service status
- ✅ Validate feature flags

### 2. Application Management
- ✅ Add new application with SBOM file
- ✅ List all applications
- ✅ Get application dependencies
- ✅ Get application status
- ✅ Remove application
- ✅ Recover deleted application

### 3. Dependency Management
- ✅ Add dependencies to application
- ✅ Update dependency versions
- ✅ Remove dependencies from application
- ✅ Batch operations

### 4. Security Scanning
- ✅ Scan application for vulnerabilities
- ✅ Manual scan with file upload
- ✅ Get SBOM by ID
- ✅ View vulnerability details

### 5. Monitoring
- ✅ Start monitoring application
- ✅ Get monitoring status
- ✅ Stop monitoring application
- ✅ Check monitoring job status

## 🔧 Environment Configuration

The environment file (`Elang_Environment.postman_environment.json`) contains:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `base_url` | API base URL | `http://localhost:8080` |
| `test_app_id` | Test application ID (auto-set) | - |
| `test_dependency_id` | Test dependency ID (auto-set) | - |
| `sbom_id` | SBOM ID (auto-set) | - |

### Changing the Base URL

For different environments:

```bash
# Local development
newman run tests/Elang_API_Tests.postman_collection.json \
  --env-var "base_url=http://localhost:8080"

# Staging
newman run tests/Elang_API_Tests.postman_collection.json \
  --env-var "base_url=https://staging.example.com"

# Production
newman run tests/Elang_API_Tests.postman_collection.json \
  --env-var "base_url=https://api.example.com"
```

## 📊 Understanding Test Results

### Success Output
```
┌─────────────────────────┬──────────┬──────────┐
│                         │ executed │   failed │
├─────────────────────────┼──────────┼──────────┤
│              iterations │        1 │        0 │
├─────────────────────────┼──────────┼──────────┤
│                requests │       15 │        0 │
├─────────────────────────┼──────────┼──────────┤
│            test-scripts │       30 │        0 │
├─────────────────────────┼──────────┼──────────┤
│      prerequest-scripts │        0 │        0 │
├─────────────────────────┼──────────┼──────────┤
│              assertions │       45 │        0 │
└─────────────────────────┴──────────┴──────────┘
```

### Failure Output
When tests fail, you'll see detailed error messages:
```
1. Health Check / Check API Health
  ✓ Status code is 200
  ✗ Response has correct structure
    AssertionError: expected undefined to have property 'status'
```

## 🧪 Adding New Tests

### Step 1: Import Collection in Postman
1. Open Postman
2. Click "Import" → Select `Elang_API_Tests.postman_collection.json`
3. Add new requests to appropriate folders

### Step 2: Add Test Scripts
In Postman, add test scripts using JavaScript:

```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has data", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('data');
});

// Save values for later use
pm.environment.set("variable_name", jsonData.data.id);
```

### Step 3: Export and Update
1. Export the collection
2. Replace `Elang_API_Tests.postman_collection.json`
3. Run tests to verify

## 🔍 Debugging Tests

### Verbose Output
```bash
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --verbose
```

### Run Specific Folder
```bash
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --folder "Health Check"
```

### Run Single Request
```bash
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --folder "Health Check" \
  --delay-request 1000
```

## 📈 CI/CD Integration

### GitHub Actions Example

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Wait for services
        run: sleep 30
      
      - name: Install Newman
        run: npm install -g newman
      
      - name: Run API tests
        run: |
          newman run tests/Elang_API_Tests.postman_collection.json \
            -e tests/Elang_Environment.postman_environment.json \
            --reporters cli,json \
            --reporter-json-export test-results.json
      
      - name: Upload results
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: test-results.json
```

## 🎯 Best Practices

1. **Always run health check first** - Ensure API is available
2. **Use environment variables** - Don't hardcode values
3. **Clean up after tests** - Remove test data
4. **Test error cases** - Not just happy paths
5. **Keep tests independent** - Each test should work standalone
6. **Use descriptive names** - Make failures easy to understand
7. **Version your collections** - Track changes over time

## 🐛 Troubleshooting

### Problem: Connection refused
**Solution**: Ensure the API is running
```bash
docker-compose ps
curl http://localhost:8080/health
```

### Problem: Tests timing out
**Solution**: Increase timeout
```bash
newman run tests/Elang_API_Tests.postman_collection.json \
  --timeout-request 30000
```

### Problem: File upload failing
**Solution**: Check file paths are correct relative to where Newman runs

### Problem: Environment variables not set
**Solution**: Verify environment file is loaded
```bash
newman run tests/Elang_API_Tests.postman_collection.json \
  -e tests/Elang_Environment.postman_environment.json \
  --verbose
```

## 📚 Additional Resources

- [Newman Documentation](https://learning.postman.com/docs/running-collections/using-newman-cli/command-line-integration-with-newman/)
- [Postman Learning Center](https://learning.postman.com/)
- [Writing Tests in Postman](https://learning.postman.com/docs/writing-scripts/test-scripts/)

## 🤝 Contributing

When adding new API endpoints:

1. Add corresponding tests to the collection
2. Update this README if needed
3. Ensure all tests pass before committing
4. Update sample files if new formats are supported

---

**Happy Testing! 🚀**
