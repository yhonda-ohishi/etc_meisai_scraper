# Rollback Procedure for gRPC Migration

## Overview
This document describes the step-by-step rollback procedure for the gRPC architecture migration.

## Prerequisites
- Git access to the repository
- Proper permissions to deploy changes
- Access to monitoring systems

## Rollback Steps

### 1. Immediate Rollback (Emergency)
```bash
# Stop current services
pkill -f etc_meisai

# Checkout pre-migration baseline
git checkout pre-migration-baseline

# Rebuild without proto dependencies
go build -o etc_meisai

# Restart services
./etc_meisai
```

### 2. Planned Rollback
```bash
# Tag current state for reference
git tag migration-attempt-$(date +%Y%m%d)

# Create rollback branch
git checkout -b rollback-from-grpc

# Revert to baseline
git reset --hard pre-migration-baseline

# Clean generated files
rm -rf src/pb/
rm -rf tests/mocks/mock_*.go

# Rebuild
go mod tidy
go build -o etc_meisai
```

### 3. Partial Rollback (Keep some changes)
```bash
# Keep performance improvements
git checkout pre-migration-baseline -- src/repositories/
git checkout HEAD -- src/services/

# Rebuild with partial changes
go build -o etc_meisai
```

## Verification Steps

1. **Service Health Check**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Database Connectivity**
   ```bash
   go test ./tests/integration/database_integration_test.go
   ```

3. **Performance Baseline**
   ```bash
   go run tests/performance/compare.go baseline.json current.json
   ```

## Recovery Time Objective (RTO)
- Emergency rollback: < 5 minutes
- Planned rollback: < 15 minutes
- Partial rollback: < 30 minutes

## Rollback Triggers
- Performance degradation > 20%
- Critical functionality broken
- Data corruption detected
- System instability

## Post-Rollback Actions
1. Document failure reasons
2. Update test cases
3. Plan remediation
4. Schedule retry

## Contact Information
- Team Lead: [Contact]
- DevOps: [Contact]
- On-call: [Contact]