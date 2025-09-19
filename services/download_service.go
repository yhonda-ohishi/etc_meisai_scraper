package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// DownloadService はダウンロード処理を管理
type DownloadService struct {
	db       *sql.DB
	logger   *log.Logger
	jobs     map[string]*DownloadJob
	jobMutex sync.RWMutex
}

// DownloadJob はダウンロードジョブの状態
type DownloadJob struct {
	ID           string
	Status       string
	Progress     int
	TotalRecords int
	ErrorMessage string
	StartedAt    time.Time
	CompletedAt  *time.Time
}

// NewDownloadService creates a new download service
func NewDownloadService(db *sql.DB, logger *log.Logger) *DownloadService {
	return &DownloadService{
		db:     db,
		logger: logger,
		jobs:   make(map[string]*DownloadJob),
	}
}

// GetAllAccountIDs は設定されているすべてのアカウントIDを取得
func (s *DownloadService) GetAllAccountIDs() []string {
	var accountIDs []string

	// 法人アカウント
	corporateAccounts := os.Getenv("ETC_CORPORATE_ACCOUNTS")
	if corporateAccounts != "" {
		for _, accountStr := range strings.Split(corporateAccounts, ",") {
			parts := strings.Split(accountStr, ":")
			if len(parts) >= 1 {
				accountIDs = append(accountIDs, parts[0])
			}
		}
	}

	// 個人アカウント
	personalAccounts := os.Getenv("ETC_PERSONAL_ACCOUNTS")
	if personalAccounts != "" {
		for _, accountStr := range strings.Split(personalAccounts, ",") {
			parts := strings.Split(accountStr, ":")
			if len(parts) >= 1 {
				accountIDs = append(accountIDs, parts[0])
			}
		}
	}

	return accountIDs
}

// ProcessAsync は非同期でダウンロードを実行
func (s *DownloadService) ProcessAsync(jobID string, accounts []string, fromDate, toDate string) {
	s.jobMutex.Lock()
	job := &DownloadJob{
		ID:        jobID,
		Status:    "processing",
		Progress:  0,
		StartedAt: time.Now(),
	}
	s.jobs[jobID] = job
	s.jobMutex.Unlock()

	// ダウンロード処理をシミュレート
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Printf("Panic in download job %s: %v", jobID, r)
				s.updateJobStatus(jobID, "failed", 0, fmt.Sprintf("Internal error: %v", r))
			}
		}()

		s.logger.Printf("Starting download job %s for %d accounts from %s to %s",
			jobID, len(accounts), fromDate, toDate)

		// 各アカウントを処理
		totalAccounts := len(accounts)
		for i, account := range accounts {
			// 進捗更新
			progress := int(float64(i+1) / float64(totalAccounts) * 100)
			s.updateJobProgress(jobID, progress)

			// 実際のダウンロード処理
			if err := s.downloadAccountData(account, fromDate, toDate); err != nil {
				s.logger.Printf("Error downloading data for account %s: %v", account, err)
				// エラーがあってもほかのアカウントの処理は続ける
			}

			// レート制限のため少し待機
			time.Sleep(time.Second)
		}

		// 完了
		now := time.Now()
		s.jobMutex.Lock()
		if job, exists := s.jobs[jobID]; exists {
			job.Status = "completed"
			job.Progress = 100
			job.CompletedAt = &now
		}
		s.jobMutex.Unlock()

		s.logger.Printf("Completed download job %s", jobID)
	}()
}

// downloadAccountData は単一アカウントのデータをダウンロード
func (s *DownloadService) downloadAccountData(accountID, fromDate, toDate string) error {
	// TODO: 実際のスクレイピング処理を実装
	s.logger.Printf("Downloading data for account %s from %s to %s", accountID, fromDate, toDate)

	// シミュレーション用の待機
	time.Sleep(2 * time.Second)

	return nil
}

// updateJobProgress はジョブの進捗を更新
func (s *DownloadService) updateJobProgress(jobID string, progress int) {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Progress = progress
	}
}

// updateJobStatus はジョブのステータスを更新
func (s *DownloadService) updateJobStatus(jobID string, status string, progress int, errorMsg string) {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	if job, exists := s.jobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
		if errorMsg != "" {
			job.ErrorMessage = errorMsg
		}
		if status == "completed" || status == "failed" {
			now := time.Now()
			job.CompletedAt = &now
		}
	}
}

// GetJobStatus はジョブのステータスを取得
func (s *DownloadService) GetJobStatus(jobID string) (*DownloadJob, bool) {
	s.jobMutex.RLock()
	defer s.jobMutex.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, false
	}

	// コピーを返す
	jobCopy := *job
	return &jobCopy, true
}