-- Migration: Update etc_dtako_mapping table
-- Date: 2025-09-18

-- Update mapping_type enum values for better categorization
ALTER TABLE etc_dtako_mapping
MODIFY COLUMN mapping_type VARCHAR(20) NOT NULL DEFAULT 'manual'
COMMENT 'マッピングタイプ: manual, auto_exact, auto_partial, auto_candidate';

-- Add index for mapping_type if not exists
CREATE INDEX IF NOT EXISTS idx_mapping_type
ON etc_dtako_mapping(mapping_type);

-- Update existing 'auto' values to 'auto_partial' for consistency
UPDATE etc_dtako_mapping
SET mapping_type = 'auto_partial'
WHERE mapping_type = 'auto';