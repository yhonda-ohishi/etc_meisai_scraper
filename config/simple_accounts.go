package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SimpleAccount represents a simple user:password pair
type SimpleAccount struct {
	UserID   string
	Password string
}

// LoadCorporateAccountsFromEnv loads corporate accounts from environment variable array
func LoadCorporateAccountsFromEnv() ([]SimpleAccount, error) {
	// Get the environment variable
	accountsEnv := os.Getenv("ETC_CORP_ACCOUNTS")
	if accountsEnv == "" {
		return nil, fmt.Errorf("ETC_CORP_ACCOUNTS not set")
	}

	// Parse as JSON array
	var accountStrings []string
	if err := json.Unmarshal([]byte(accountsEnv), &accountStrings); err != nil {
		// If JSON parsing fails, try simple comma-separated format
		// Format: user1:pass1,user2:pass2,user3:pass3
		accountStrings = strings.Split(accountsEnv, ",")
	}

	var accounts []SimpleAccount
	for _, accStr := range accountStrings {
		// Trim whitespace
		accStr = strings.TrimSpace(accStr)
		if accStr == "" {
			continue
		}

		// Split by colon
		parts := strings.Split(accStr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid account format: %s (expected user:password)", accStr)
		}

		account := SimpleAccount{
			UserID:   strings.TrimSpace(parts[0]),
			Password: strings.TrimSpace(parts[1]),
		}

		accounts = append(accounts, account)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no valid accounts found")
	}

	return accounts, nil
}

// LoadPersonalAccountsFromEnv loads personal accounts from environment variable array
func LoadPersonalAccountsFromEnv() ([]SimpleAccount, error) {
	// Get the environment variable
	accountsEnv := os.Getenv("ETC_PERSONAL_ACCOUNTS")
	if accountsEnv == "" {
		return nil, fmt.Errorf("ETC_PERSONAL_ACCOUNTS not set")
	}

	// Parse as JSON array
	var accountStrings []string
	if err := json.Unmarshal([]byte(accountsEnv), &accountStrings); err != nil {
		// If JSON parsing fails, try simple comma-separated format
		accountStrings = strings.Split(accountsEnv, ",")
	}

	var accounts []SimpleAccount
	for _, accStr := range accountStrings {
		accStr = strings.TrimSpace(accStr)
		if accStr == "" {
			continue
		}

		parts := strings.Split(accStr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid account format: %s", accStr)
		}

		account := SimpleAccount{
			UserID:   strings.TrimSpace(parts[0]),
			Password: strings.TrimSpace(parts[1]),
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}