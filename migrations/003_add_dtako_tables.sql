-- Migration: Add Dtako related tables
-- Date: 2025-09-18

-- Create mapping_batch_jobs table
CREATE TABLE IF NOT EXISTS mapping_batch_jobs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'ジョブステータス: pending, running, completed, failed, cancelled',
    start_date DATE NOT NULL COMMENT '処理開始日',
    end_date DATE NOT NULL COMMENT '処理終了日',
    total_records INT DEFAULT 0 COMMENT '総レコード数',
    processed_records INT DEFAULT 0 COMMENT '処理済みレコード数',
    matched_records INT DEFAULT 0 COMMENT 'マッチ成功数',
    error_count INT DEFAULT 0 COMMENT 'エラー数',
    error_details TEXT COMMENT 'エラー詳細',
    started_at TIMESTAMP NULL COMMENT '開始時刻',
    completed_at TIMESTAMP NULL COMMENT '完了時刻',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    created_by VARCHAR(100) COMMENT '作成者',
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='バッチ処理ジョブ管理テーブル';