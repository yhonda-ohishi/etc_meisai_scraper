package services

import (
	"fmt"
	"sync"
	"time"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// SessionService manages import sessions
type SessionService struct {
	sessions   map[string]*models.ImportSession
	mu         sync.RWMutex
	nextID     int64
	idCounter  sync.Mutex
}

// NewSessionService creates a new session service
func NewSessionService() *SessionService {
	return &SessionService{
		sessions: make(map[string]*models.ImportSession),
		nextID:   1,
	}
}

// StartSession starts a new import session
func (s *SessionService) StartSession(accountType string, startDate, endDate time.Time) *models.ImportSession {
	// Get unique ID
	s.idCounter.Lock()
	sessionID := s.nextID
	s.nextID++
	s.idCounter.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	session := &models.ImportSession{
		ID:          sessionID,
		AccountType: accountType,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      "pending",
		ExecutedAt:  time.Now(),
	}

	sessionKey := fmt.Sprintf("%d", session.ID)
	s.sessions[sessionKey] = session

	return session
}

// UpdateSession updates an existing session
func (s *SessionService) UpdateSession(sessionID int64, recordCount int, status string, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionKey := fmt.Sprintf("%d", sessionID)
	session, exists := s.sessions[sessionKey]
	if !exists {
		return fmt.Errorf("session not found: %d", sessionID)
	}

	session.RecordCount = recordCount
	session.Status = status
	if errorMsg != "" {
		session.ErrorMessage = errorMsg
	}

	return nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(sessionID int64) (*models.ImportSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionKey := fmt.Sprintf("%d", sessionID)
	session, exists := s.sessions[sessionKey]
	if !exists {
		return nil, fmt.Errorf("session not found: %d", sessionID)
	}

	return session, nil
}

// GetAllSessions retrieves all sessions
func (s *SessionService) GetAllSessions() []*models.ImportSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.ImportSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetSessionsByStatus retrieves sessions by status
func (s *SessionService) GetSessionsByStatus(status string) []*models.ImportSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*models.ImportSession, 0)
	for _, session := range s.sessions {
		if session.Status == status {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// CleanupOldSessions removes sessions older than the specified duration
func (s *SessionService) CleanupOldSessions(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for key, session := range s.sessions {
		if session.ExecutedAt.Before(cutoff) {
			delete(s.sessions, key)
			removed++
		}
	}

	return removed
}

// GetSessionStats returns statistics about sessions
func (s *SessionService) GetSessionStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total": len(s.sessions),
		"by_status": map[string]int{
			"pending": 0,
			"success": 0,
			"failed":  0,
		},
		"by_account_type": map[string]int{
			"corporate": 0,
			"personal":  0,
		},
	}

	statusCount := stats["by_status"].(map[string]int)
	accountCount := stats["by_account_type"].(map[string]int)

	for _, session := range s.sessions {
		if session.Status == "pending" || session.Status == "success" || session.Status == "failed" {
			statusCount[session.Status]++
		}
		if session.AccountType == "corporate" || session.AccountType == "personal" {
			accountCount[session.AccountType]++
		}
	}

	return stats
}