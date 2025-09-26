# GORM Hooks Analysis and Migration Plan

**Purpose**: Document all existing GORM hooks and their business logic for migration to service layer.

**Date**: 2025-09-26
**Task**: T036 - Identify and document all GORM hooks in existing models

## Summary of GORM Hooks

Found **16 GORM hooks** across **6 model files**:

### 1. ETCMeisaiRecord (`src/models/etc_meisai_record.go`)

**BeforeCreate(tx *gorm.DB) error** (Line 38-48):
- **Business Logic**:
  - Data validation via `validate()` method
  - Hash generation via `generateHash()` if empty
- **Hook Type**: Pre-creation data processing
- **Complexity**: High - comprehensive validation + hash generation

**BeforeSave(tx *gorm.DB) error** (Line 51-53):
- **Business Logic**: Data validation via `validate()` method
- **Hook Type**: Pre-save data validation
- **Complexity**: Medium - validation only

**BeforeUpdate() error** (Line 238):
- **Business Logic**: (Need to examine further)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

**Validation Logic in `validate()` method**:
- Date validation (not zero, not future)
- Time format validation (HH:MM:SS)
- IC name validation (not empty, max 100 chars)
- Toll amount validation (non-negative, max 999999)
- Car number validation (Japanese vehicle formats)
- ETC card number validation (16-19 digits)
- ETC number validation (optional, alphanumeric)

**Hash Generation in `generateHash()` method**:
- SHA256 hash of: Date|Time|EntranceIC|ExitIC|TollAmount|CarNumber|ETCCardNumber

### 2. ETCMapping (`src/models/etc_mapping.go`)

**BeforeCreate(tx *gorm.DB) error** (Line 63-81):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

**BeforeSave(tx *gorm.DB) error** (Line 82):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-save processing
- **Complexity**: Unknown

**BeforeUpdate() error** (Line 327):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

### 3. ImportSession (`src/models/import_session.go`)

**BeforeCreate(tx *gorm.DB) error** (Line 73-92):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

**BeforeSave(tx *gorm.DB) error** (Line 93):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-save processing
- **Complexity**: Unknown

**BeforeUpdate() error** (Line 465):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

### 4. ETCImportBatch (`src/models/etc_import_batch.go`)

**BeforeCreate() error** (Line 39-46):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

**BeforeUpdate() error** (Line 47):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

### 5. ETCMeisai (`src/models/etc_meisai.go`)

**BeforeCreate() error** (Line 43-61):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

**BeforeUpdate() error** (Line 62):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

### 6. ETCMeisaiMapping (`src/models/etc_meisai_mapping.go`)

**BeforeCreate() error** (Line 25-32):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

**BeforeUpdate() error** (Line 33):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-update processing
- **Complexity**: Unknown

### 7. ImportBatch (`src/models/import_session.go`)

**BeforeCreate() error** (Line 552):
- **Business Logic**: (Need to examine)
- **Hook Type**: Pre-creation processing
- **Complexity**: Unknown

## Hook Categories Identified

### Data Validation Hooks
- Comprehensive field validation
- Business rule enforcement
- Format validation (dates, times, numbers)
- Cross-field validation logic

### Data Processing Hooks
- Hash generation for uniqueness
- Default value assignment
- Data transformation/normalization
- Calculated field population

### Audit/Tracking Hooks
- Timestamp management
- User tracking
- Status updates
- Change logging

## Migration Strategy

1. **Extract to Validation Service**: All validation logic
2. **Extract to Business Service**: Hash generation, calculations
3. **Extract to Audit Service**: Tracking and logging
4. **Adapter Layer Integration**: Call extracted services at appropriate lifecycle points

## Next Steps

1. **T037**: Create `hooks_migrator.go` to centralize business logic
2. **T038**: Extract validation logic to `validation_service.go`
3. **T039**: Extract audit logic to `audit_service.go`
4. **T040**: Write comprehensive tests for extracted logic
5. **T041**: Update adapter layer to call migrated services

---
*Analysis complete - Ready for hook migration implementation*