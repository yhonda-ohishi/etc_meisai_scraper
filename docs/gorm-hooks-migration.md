# GORM Hooks Migration Documentation

## Identified GORM Hooks

### ETCImportBatch Model
- **BeforeCreate**: Validates data before creation
- **BeforeUpdate**: Validates data before update

### ETCMeisaiMapping Model
- **BeforeCreate**: Validates data before creation
- **BeforeUpdate**: Validates data before update

### ETCMeisaiRecord Model
- **BeforeCreate**: Validates and generates hash before creation
- **BeforeSave**: Validates data before saving
- **BeforeUpdate**: Regenerates hash before update

### ETCMeisai Model
- **BeforeCreate**: Sets timestamps and validates
- **BeforeUpdate**: Updates timestamp

### ETCMapping Model
- **BeforeCreate**: Sets timestamps and validates
- **BeforeSave**: Updates timestamp and validates
- **BeforeUpdate**: Updates timestamp

### ImportSession Model
- **BeforeCreate**: Generates UUID and validates
- **BeforeSave**: Validates data
- **BeforeUpdate**: Updates timestamp

### ImportBatch Model
- **BeforeCreate**: Sets timestamps

## Migration Strategy

The hooks will be migrated to the following services:

1. **ValidationService** (src/services/validation_service.go)
   - All validation logic from hooks
   - Centralized validation methods

2. **AuditService** (src/services/audit_service.go)
   - Timestamp management
   - Audit trail logging

3. **HooksMigrator** (src/services/hooks_migrator.go)
   - Coordinates hook logic execution
   - UUID generation, hash calculation
   - Calls validation and audit services

## Implementation Points

Each adapter method will call the appropriate migrated hook logic:
- Before Create operations → Call HooksMigrator.BeforeCreate()
- Before Update operations → Call HooksMigrator.BeforeUpdate()
- Before Save operations → Call HooksMigrator.BeforeSave()