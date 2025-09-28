# Data Model: Full gRPC Architecture Migration

**Feature**: 006-refactor-src-to
**Date**: 2025-09-26
**Source**: Extracted from Protocol Buffer definitions

## Overview
This document defines all data entities that will be migrated from manual GORM models to Protocol Buffer messages. All entities are defined in `.proto` files and generate Go code automatically.

## Core Entities

### 1. ETCMappingEntity
**Purpose**: Maps ETC records to external entities (e.g., dtako records)
**Proto File**: `src/proto/repository.proto`

```proto
message ETCMappingEntity {
  int64 id = 1;
  int64 etc_record_id = 2;
  string mapping_type = 3;
  int64 mapped_entity_id = 4;
  string mapped_entity_type = 5;
  float confidence = 6;
  MappingStatusEnum status = 7;
  google.protobuf.Struct metadata = 8;
  string created_by = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}
```

**Database Mapping**:
| Proto Field | Database Column | Type | Notes |
|------------|----------------|------|-------|
| id | id | BIGINT | Primary key, auto-increment |
| etc_record_id | etc_record_id | BIGINT | Foreign key to ETC records |
| mapping_type | mapping_type | VARCHAR(50) | Type of mapping (e.g., "dtako") |
| mapped_entity_id | mapped_entity_id | BIGINT | ID in the mapped system |
| mapped_entity_type | mapped_entity_type | VARCHAR(50) | Entity type in mapped system |
| confidence | confidence | FLOAT | Matching confidence score |
| status | status | VARCHAR(20) | Enum string value |
| metadata | metadata | JSON | Additional mapping data |
| created_by | created_by | VARCHAR(100) | User/system that created |
| created_at | created_at | TIMESTAMP | Creation timestamp |
| updated_at | updated_at | TIMESTAMP | Last update timestamp |

### 2. ETCMeisaiRecordEntity
**Purpose**: Represents individual ETC toll transaction records
**Proto File**: `src/proto/repository.proto`

```proto
message ETCMeisaiRecordEntity {
  int64 id = 1;
  string hash = 2;
  string date = 3;  // YYYY-MM-DD format
  string time = 4;  // HH:MM:SS format
  string entrance_ic = 5;
  string exit_ic = 6;
  int32 toll_amount = 7;
  string car_number = 8;
  string etc_card_number = 9;
  optional string etc_num = 10;
  optional int64 dtako_row_id = 11;
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}
```

**Database Mapping**:
| Proto Field | Database Column | Type | Notes |
|------------|----------------|------|-------|
| id | id | BIGINT | Primary key |
| hash | hash | VARCHAR(64) | Unique record hash |
| date | date | DATE | Transaction date |
| time | time | TIME | Transaction time |
| entrance_ic | entrance_ic | VARCHAR(100) | Entry interchange |
| exit_ic | exit_ic | VARCHAR(100) | Exit interchange |
| toll_amount | toll_amount | INT | Amount in yen |
| car_number | car_number | VARCHAR(20) | Vehicle registration |
| etc_card_number | etc_card_number | VARCHAR(20) | ETC card identifier |
| etc_num | etc_num | VARCHAR(50) | Optional ETC number |
| dtako_row_id | dtako_row_id | BIGINT | Optional dtako reference |
| created_at | created_at | TIMESTAMP | Import timestamp |
| updated_at | updated_at | TIMESTAMP | Last update timestamp |

### 3. ImportSessionEntity
**Purpose**: Tracks CSV import sessions and progress
**Proto File**: `src/proto/repository.proto`

```proto
message ImportSessionEntity {
  string id = 1;  // UUID
  string account_type = 2;
  string account_id = 3;
  string file_name = 4;
  int64 file_size = 5;
  ImportStatusEnum status = 6;
  int32 total_rows = 7;
  int32 processed_rows = 8;
  int32 success_rows = 9;
  int32 error_rows = 10;
  int32 duplicate_rows = 11;
  google.protobuf.Timestamp started_at = 12;
  optional google.protobuf.Timestamp completed_at = 13;
  string created_by = 14;
  google.protobuf.Timestamp created_at = 15;
}
```

**Database Mapping**:
| Proto Field | Database Column | Type | Notes |
|------------|----------------|------|-------|
| id | id | VARCHAR(36) | UUID primary key |
| account_type | account_type | VARCHAR(20) | corporate/personal |
| account_id | account_id | VARCHAR(50) | Account identifier |
| file_name | file_name | VARCHAR(255) | Original filename |
| file_size | file_size | BIGINT | File size in bytes |
| status | status | VARCHAR(20) | Enum string value |
| total_rows | total_rows | INT | Total CSV rows |
| processed_rows | processed_rows | INT | Rows processed |
| success_rows | success_rows | INT | Successfully imported |
| error_rows | error_rows | INT | Failed imports |
| duplicate_rows | duplicate_rows | INT | Duplicate records |
| started_at | started_at | TIMESTAMP | Import start time |
| completed_at | completed_at | TIMESTAMP | Import end time |
| created_by | created_by | VARCHAR(100) | User initiating import |
| created_at | created_at | TIMESTAMP | Session creation time |

### 4. ImportErrorEntity
**Purpose**: Records errors during import processing
**Proto File**: `src/proto/repository.proto`

```proto
message ImportErrorEntity {
  int64 id = 1;
  string session_id = 2;
  int32 row_number = 3;
  string error_type = 4;
  string error_message = 5;
  string raw_data = 6;
  google.protobuf.Timestamp created_at = 7;
}
```

**Database Mapping**:
| Proto Field | Database Column | Type | Notes |
|------------|----------------|------|-------|
| id | id | BIGINT | Primary key |
| session_id | session_id | VARCHAR(36) | Foreign key to session |
| row_number | row_number | INT | CSV row number |
| error_type | error_type | VARCHAR(50) | Error classification |
| error_message | error_message | TEXT | Detailed error message |
| raw_data | raw_data | TEXT | Original CSV row |
| created_at | created_at | TIMESTAMP | Error timestamp |

## Enumerations

### MappingStatusEnum
```proto
enum MappingStatusEnum {
  MAPPING_STATUS_ENUM_UNSPECIFIED = 0;
  MAPPING_STATUS_ENUM_ACTIVE = 1;
  MAPPING_STATUS_ENUM_INACTIVE = 2;
  MAPPING_STATUS_ENUM_PENDING = 3;
  MAPPING_STATUS_ENUM_REJECTED = 4;
}
```

### ImportStatusEnum
```proto
enum ImportStatusEnum {
  IMPORT_STATUS_ENUM_UNSPECIFIED = 0;
  IMPORT_STATUS_ENUM_PENDING = 1;
  IMPORT_STATUS_ENUM_PROCESSING = 2;
  IMPORT_STATUS_ENUM_COMPLETED = 3;
  IMPORT_STATUS_ENUM_FAILED = 4;
  IMPORT_STATUS_ENUM_CANCELLED = 5;
}
```

### SortOrderEnum
```proto
enum SortOrderEnum {
  SORT_ORDER_ENUM_UNSPECIFIED = 0;
  SORT_ORDER_ENUM_ASC = 1;
  SORT_ORDER_ENUM_DESC = 2;
}
```

## Request/Response Messages

### Repository Operations
Each repository service uses standardized request/response patterns:

1. **CRUD Operations**
   - Create: `Create{Entity}RepoRequest/Response`
   - Read: `GetByIDRequest` → `{Entity}`
   - Update: `Update{Entity}RepoRequest/Response`
   - Delete: `DeleteRequest/Response`

2. **Query Operations**
   - List: `List{Entity}RepoRequest/Response` with pagination
   - Search: Specific query methods per entity

3. **Bulk Operations**
   - BulkCreate: `BulkCreate{Entity}Request/Response`
   - BulkUpdate: `BulkUpdate{Entity}Request/Response`

4. **Transaction Support**
   - Begin: `BeginTransactionRequest/Response`
   - Commit: `TransactionRequest` → `Empty`
   - Rollback: `TransactionRequest` → `Empty`

## State Transitions

### Import Session States
```
PENDING → PROCESSING → COMPLETED
         ↓           ↓
       FAILED    CANCELLED
```

### Mapping Status States
```
PENDING → ACTIVE → INACTIVE
   ↓
REJECTED
```

## Validation Rules

### ETCMeisaiRecord
- `hash`: Must be unique (SHA256 of record data)
- `date`: Format YYYY-MM-DD
- `time`: Format HH:MM:SS
- `toll_amount`: Non-negative integer
- `car_number`: Valid Japanese plate format
- `etc_card_number`: 16-digit format

### ETCMapping
- `confidence`: Range 0.0 to 1.0
- `etc_record_id`: Must exist in ETCMeisaiRecord
- `mapping_type`: Predefined values only

### ImportSession
- `id`: Valid UUID v4
- `file_size`: Positive integer
- `processed_rows` <= `total_rows`
- `success_rows` + `error_rows` + `duplicate_rows` <= `processed_rows`

## Migration Considerations

### From GORM Models
1. **Field Naming**: Snake_case in proto → CamelCase in generated Go
2. **Timestamps**: Use `google.protobuf.Timestamp` instead of `time.Time`
3. **Optionals**: Use `optional` for nullable fields
4. **Enums**: Replace string constants with proto enums
5. **JSON Fields**: Use `google.protobuf.Struct` for dynamic JSON

### Database Adapter Requirements
- Handle timestamp conversions
- Map enum values to strings for database
- Convert proto Struct to/from JSON
- Maintain backward compatibility with existing schema

## Performance Optimizations

### Indexing Strategy
```sql
-- ETCMeisaiRecord
CREATE INDEX idx_etc_meisai_hash ON etc_meisai_records(hash);
CREATE INDEX idx_etc_meisai_date ON etc_meisai_records(date);
CREATE INDEX idx_etc_meisai_car ON etc_meisai_records(car_number);

-- ETCMapping
CREATE INDEX idx_mapping_record ON etc_mappings(etc_record_id);
CREATE INDEX idx_mapping_entity ON etc_mappings(mapped_entity_type, mapped_entity_id);

-- ImportSession
CREATE INDEX idx_import_status ON import_sessions(status);
CREATE INDEX idx_import_account ON import_sessions(account_type, account_id);
```

### Caching Strategy
- Cache frequently accessed mappings
- Cache active import sessions
- Use proto binary format for cache serialization

## Next Steps

1. Generate Go code from proto definitions
2. Create database adapter implementations
3. Write contract tests for each entity
4. Implement repository gRPC servers
5. Migrate service layer to use proto messages

---
*Data model defined: 2025-09-26*