package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yhonda-ohishi/etc_meisai/src/handlers"
)

// TestNewAccountsHandler tests accounts handler creation
func TestNewAccountsHandler(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)

		handler := handlers.NewAccountsHandler(baseHandler)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.BaseHandler)
	})
}

// TestGetAccounts tests account retrieval functionality
func TestGetAccounts(t *testing.T) {
	// Store original environment variables
	originalCorporate := os.Getenv("ETC_CORPORATE_ACCOUNTS")
	originalPersonal := os.Getenv("ETC_PERSONAL_ACCOUNTS")

	// Clean up after tests
	defer func() {
		if originalCorporate != "" {
			os.Setenv("ETC_CORPORATE_ACCOUNTS", originalCorporate)
		} else {
			os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		}
		if originalPersonal != "" {
			os.Setenv("ETC_PERSONAL_ACCOUNTS", originalPersonal)
		} else {
			os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
		}
	}()

	t.Run("get accounts with both corporate and personal", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1:etc123,corp2:pass2:etc456")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "personal1:pass3:etc789,personal2:pass4:etc012")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Contains(t, response.Message, "Accounts retrieved successfully")

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 4, int(count)) // 2 corporate + 2 personal
		assert.Len(t, accounts, 4)

		// Check corporate accounts
		foundCorp1 := false
		foundCorp2 := false
		foundPersonal1 := false
		foundPersonal2 := false

		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			id := account["id"].(string)
			accType := account["type"].(string)

			switch id {
			case "corp1":
				foundCorp1 = true
				assert.Equal(t, "corporate", accType)
				assert.Equal(t, "corp1", account["name"])
				assert.Equal(t, "corp1", account["username"])
				assert.Equal(t, "etc123", account["etc_num"])
				assert.True(t, account["is_active"].(bool))
			case "corp2":
				foundCorp2 = true
				assert.Equal(t, "corporate", accType)
				assert.Equal(t, "etc456", account["etc_num"])
			case "personal1":
				foundPersonal1 = true
				assert.Equal(t, "personal", accType)
				assert.Equal(t, "etc789", account["etc_num"])
			case "personal2":
				foundPersonal2 = true
				assert.Equal(t, "personal", accType)
				assert.Equal(t, "etc012", account["etc_num"])
			}
		}

		assert.True(t, foundCorp1, "Should find corp1 account")
		assert.True(t, foundCorp2, "Should find corp2 account")
		assert.True(t, foundPersonal1, "Should find personal1 account")
		assert.True(t, foundPersonal2, "Should find personal2 account")
	})

	t.Run("get accounts with only corporate", func(t *testing.T) {
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1:etc123")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 1, int(count))
		assert.Len(t, accounts, 1)

		account := accounts[0].(map[string]interface{})
		assert.Equal(t, "corp1", account["id"])
		assert.Equal(t, "corporate", account["type"])
	})

	t.Run("get accounts with only personal", func(t *testing.T) {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "personal1:pass1:etc456")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 1, int(count))
		assert.Len(t, accounts, 1)

		account := accounts[0].(map[string]interface{})
		assert.Equal(t, "personal1", account["id"])
		assert.Equal(t, "personal", account["type"])
	})

	t.Run("get accounts with no environment variables", func(t *testing.T) {
		os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 0, int(count))
		assert.Len(t, accounts, 0)
	})

	t.Run("get accounts with malformed corporate accounts", func(t *testing.T) {
		// Test with incomplete account format (missing password or etc_num)
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1:etc123,incomplete_account,corp2:pass2")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		// Should only parse valid accounts (corp1 with 3 parts, corp2 with 2 parts)
		assert.Equal(t, 2, int(count))
		assert.Len(t, accounts, 2)

		// Verify corp1 has etc_num
		found_corp1_with_etc := false
		found_corp2_without_etc := false
		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			id := account["id"].(string)

			if id == "corp1" {
				assert.Equal(t, "etc123", account["etc_num"])
				found_corp1_with_etc = true
			} else if id == "corp2" {
				// corp2 should have empty etc_num since it only has 2 parts
				etcNum, exists := account["etc_num"]
				if exists {
					assert.Empty(t, etcNum)
				}
				found_corp2_without_etc = true
			}
		}

		assert.True(t, found_corp1_with_etc)
		assert.True(t, found_corp2_without_etc)
	})

	t.Run("get accounts with empty environment variables", func(t *testing.T) {
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 0, int(count))
		assert.Len(t, accounts, 0)
	})

	t.Run("get accounts with special characters", func(t *testing.T) {
		// Test with accounts containing special characters
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp-1:pass@123:etc_456,corp.2:pass#456:etc.789")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 2, int(count))
		assert.Len(t, accounts, 2)

		// Verify accounts with special characters are parsed correctly
		foundCorp1 := false
		foundCorp2 := false
		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			id := account["id"].(string)

			if id == "corp-1" {
				assert.Equal(t, "corp-1", account["username"])
				assert.Equal(t, "etc_456", account["etc_num"])
				foundCorp1 = true
			} else if id == "corp.2" {
				assert.Equal(t, "corp.2", account["username"])
				assert.Equal(t, "etc.789", account["etc_num"])
				foundCorp2 = true
			}
		}

		assert.True(t, foundCorp1)
		assert.True(t, foundCorp2)
	})

	t.Run("get accounts with single account per type", func(t *testing.T) {
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "single_corp:pass1:etc123")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "single_personal:pass2:etc456")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 2, int(count))
		assert.Len(t, accounts, 2)

		// Verify both account types exist
		typeFound := map[string]bool{"corporate": false, "personal": false}
		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			accType := account["type"].(string)
			typeFound[accType] = true
		}

		assert.True(t, typeFound["corporate"])
		assert.True(t, typeFound["personal"])
	})
}

// TestAccountStructure tests the account structure
func TestAccountStructure(t *testing.T) {
	t.Run("complete account structure", func(t *testing.T) {
		account := handlers.Account{
			ID:       "test_corp",
			Name:     "Test Corporate Account",
			Type:     "corporate",
			Username: "test_corp",
			ETCNum:   "ETC123456",
			IsActive: true,
		}

		jsonData, err := json.Marshal(account)
		assert.NoError(t, err)

		var unmarshaledAccount handlers.Account
		err = json.Unmarshal(jsonData, &unmarshaledAccount)
		assert.NoError(t, err)

		assert.Equal(t, "test_corp", unmarshaledAccount.ID)
		assert.Equal(t, "Test Corporate Account", unmarshaledAccount.Name)
		assert.Equal(t, "corporate", unmarshaledAccount.Type)
		assert.Equal(t, "test_corp", unmarshaledAccount.Username)
		assert.Equal(t, "ETC123456", unmarshaledAccount.ETCNum)
		assert.True(t, unmarshaledAccount.IsActive)
	})

	t.Run("minimal account structure", func(t *testing.T) {
		account := handlers.Account{
			ID:       "minimal_account",
			Type:     "personal",
			IsActive: false,
		}

		jsonData, err := json.Marshal(account)
		assert.NoError(t, err)

		var unmarshaledAccount handlers.Account
		err = json.Unmarshal(jsonData, &unmarshaledAccount)
		assert.NoError(t, err)

		assert.Equal(t, "minimal_account", unmarshaledAccount.ID)
		assert.Empty(t, unmarshaledAccount.Name)
		assert.Equal(t, "personal", unmarshaledAccount.Type)
		assert.Empty(t, unmarshaledAccount.Username)
		assert.Empty(t, unmarshaledAccount.ETCNum)
		assert.False(t, unmarshaledAccount.IsActive)
	})
}

// TestAccountsHandlerConcurrency tests concurrent access
func TestAccountsHandlerConcurrency(t *testing.T) {
	// Store original environment variables
	originalCorporate := os.Getenv("ETC_CORPORATE_ACCOUNTS")
	originalPersonal := os.Getenv("ETC_PERSONAL_ACCOUNTS")

	// Clean up after test
	defer func() {
		if originalCorporate != "" {
			os.Setenv("ETC_CORPORATE_ACCOUNTS", originalCorporate)
		} else {
			os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		}
		if originalPersonal != "" {
			os.Setenv("ETC_PERSONAL_ACCOUNTS", originalPersonal)
		} else {
			os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
		}
	}()

	t.Run("concurrent account requests", func(t *testing.T) {
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass1:etc123,corp2:pass2:etc456")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "personal1:pass3:etc789")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		const numRequests = 10
		done := make(chan bool, numRequests)

		// Launch concurrent requests
		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/api/accounts", nil)
				w := httptest.NewRecorder()

				handler.GetAccounts(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response handlers.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				data := response.Data.(map[string]interface{})
				count := data["count"].(float64)
				assert.Equal(t, 3, int(count)) // 2 corporate + 1 personal

				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			<-done
		}
	})
}

// TestAccountsHandlerEdgeCases tests edge cases
func TestAccountsHandlerEdgeCases(t *testing.T) {
	// Store original environment variables
	originalCorporate := os.Getenv("ETC_CORPORATE_ACCOUNTS")
	originalPersonal := os.Getenv("ETC_PERSONAL_ACCOUNTS")

	// Clean up after tests
	defer func() {
		if originalCorporate != "" {
			os.Setenv("ETC_CORPORATE_ACCOUNTS", originalCorporate)
		} else {
			os.Unsetenv("ETC_CORPORATE_ACCOUNTS")
		}
		if originalPersonal != "" {
			os.Setenv("ETC_PERSONAL_ACCOUNTS", originalPersonal)
		} else {
			os.Unsetenv("ETC_PERSONAL_ACCOUNTS")
		}
	}()

	t.Run("accounts with colons in values", func(t *testing.T) {
		// Test with values that contain colons (edge case for splitting)
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "corp1:pass:with:colons:etc123")
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		// Should parse this as having many parts, but only use first 3
		assert.Equal(t, 1, int(count))
		assert.Len(t, accounts, 1)

		account := accounts[0].(map[string]interface{})
		assert.Equal(t, "corp1", account["id"])
		assert.Equal(t, "etc123", account["etc_num"]) // Should be the 4th part
	})

	t.Run("accounts with unicode characters", func(t *testing.T) {
		os.Setenv("ETC_CORPORATE_ACCOUNTS", "法人1:パスワード1:etc123,企業2:パスワード2:etc456")
		os.Setenv("ETC_PERSONAL_ACCOUNTS", "個人1:パスワード3:etc789")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		handler.GetAccounts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		accounts := data["accounts"].([]interface{})
		count := data["count"].(float64)

		assert.Equal(t, 3, int(count))
		assert.Len(t, accounts, 3)

		// Verify unicode account names are preserved
		foundUnicodeAccount := false
		for _, acc := range accounts {
			account := acc.(map[string]interface{})
			id := account["id"].(string)

			if id == "法人1" {
				assert.Equal(t, "法人1", account["name"])
				assert.Equal(t, "法人1", account["username"])
				assert.Equal(t, "corporate", account["type"])
				foundUnicodeAccount = true
			}
		}

		assert.True(t, foundUnicodeAccount, "Should find unicode account")
	})

	t.Run("very long account configurations", func(t *testing.T) {
		// Create a very long account configuration string
		var corporateAccounts []string
		for i := 0; i < 100; i++ {
			corporateAccounts = append(corporateAccounts,
				"corp"+string(rune(i+'0'))+":pass"+string(rune(i+'0'))+":etc"+string(rune(i+'0')))
		}

		os.Setenv("ETC_CORPORATE_ACCOUNTS", strings.Join(corporateAccounts, ","))
		os.Unsetenv("ETC_PERSONAL_ACCOUNTS")

		mockRegistry := createMockServiceRegistry()
		logger := createTestLogger()
		baseHandler := *handlers.NewBaseHandler(mockRegistry, logger)
		handler := handlers.NewAccountsHandler(baseHandler)

		req := httptest.NewRequest("GET", "/api/accounts", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.GetAccounts(w, req)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Less(t, duration, 100*time.Millisecond, "Large account list should be processed quickly")

		var response handlers.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data := response.Data.(map[string]interface{})
		count := data["count"].(float64)
		assert.Equal(t, 100, int(count))
	})
}