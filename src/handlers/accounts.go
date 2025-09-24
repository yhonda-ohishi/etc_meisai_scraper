package handlers

import (
	"net/http"
	"os"
	"strings"
)

// AccountsHandler はアカウント関連のハンドラー
type AccountsHandler struct {
	BaseHandler
}

// Account はETCアカウント情報
type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Username string `json:"username"`
	ETCNum   string `json:"etc_num,omitempty"`
	IsActive bool   `json:"is_active"`
}

// NewAccountsHandler creates a new accounts handler
func NewAccountsHandler(base BaseHandler) *AccountsHandler {
	return &AccountsHandler{BaseHandler: base}
}

// GetAccounts は登録されているアカウント一覧を返す
func (h *AccountsHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := []Account{}

	// 法人アカウントの取得
	corporateAccounts := os.Getenv("ETC_CORPORATE_ACCOUNTS")
	if corporateAccounts != "" {
		for _, accountStr := range strings.Split(corporateAccounts, ",") {
			parts := strings.Split(accountStr, ":")
			if len(parts) >= 2 {
				account := Account{
					ID:       parts[0],
					Name:     parts[0],
					Type:     "corporate",
					Username: parts[0],
					IsActive: true,
				}
				if len(parts) >= 3 {
					// If there are more than 3 parts, take the last one as ETC number
					// to handle cases like "id:password:extra:colons:etc_num"
					account.ETCNum = parts[len(parts)-1]
				}
				accounts = append(accounts, account)
			}
		}
	}

	// 個人アカウントの取得
	personalAccounts := os.Getenv("ETC_PERSONAL_ACCOUNTS")
	if personalAccounts != "" {
		for _, accountStr := range strings.Split(personalAccounts, ",") {
			parts := strings.Split(accountStr, ":")
			if len(parts) >= 2 {
				account := Account{
					ID:       parts[0],
					Name:     parts[0],
					Type:     "personal",
					Username: parts[0],
					IsActive: true,
				}
				if len(parts) >= 3 {
					// If there are more than 3 parts, take the last one as ETC number
					// to handle cases like "id:password:extra:colons:etc_num"
					account.ETCNum = parts[len(parts)-1]
				}
				accounts = append(accounts, account)
			}
		}
	}

	response := map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	}

	h.RespondSuccess(w, response, "Accounts retrieved successfully")
}