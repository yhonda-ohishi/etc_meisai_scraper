package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AccountType represents the type of ETC account
type AccountType string

const (
	AccountTypeCorporate AccountType = "corporate"
	AccountTypePersonal  AccountType = "personal"
)

// ETCAccount represents a single ETC account
type ETCAccount struct {
	Name         string      `json:"name"`          // Account name for identification
	UserID       string      `json:"user_id"`       // Login user ID
	Password     string      `json:"password"`      // Login password
	PasswordCorp string      `json:"password_corp"` // Corporate password (if applicable)
	Type         AccountType `json:"type"`          // Account type (corporate/personal)
	CardNumbers  []string    `json:"card_numbers"`  // Associated card numbers
	Active       bool        `json:"active"`        // Whether this account is active
}

// AccountsConfig holds all ETC accounts configuration
type AccountsConfig struct {
	Accounts []ETCAccount `json:"accounts"`
}

// LoadAccountsFromEnv loads accounts from environment variables
func LoadAccountsFromEnv() (*AccountsConfig, error) {
	config := &AccountsConfig{
		Accounts: []ETCAccount{},
	}

	// Load corporate accounts (ETC_CORP_USER_1, ETC_CORP_PASS_1, ETC_CORP_PASS2_1, etc.)
	for i := 1; i <= 10; i++ {
		userKey := fmt.Sprintf("ETC_CORP_USER_%d", i)
		passKey := fmt.Sprintf("ETC_CORP_PASS_%d", i)
		pass2Key := fmt.Sprintf("ETC_CORP_PASS2_%d", i)
		nameKey := fmt.Sprintf("ETC_CORP_NAME_%d", i)
		cardsKey := fmt.Sprintf("ETC_CORP_CARDS_%d", i)

		userID := os.Getenv(userKey)
		if userID == "" {
			continue
		}

		account := ETCAccount{
			Name:         os.Getenv(nameKey),
			UserID:       userID,
			Password:     os.Getenv(passKey),
			PasswordCorp: os.Getenv(pass2Key),
			Type:         AccountTypeCorporate,
			Active:       true,
		}

		if account.Name == "" {
			account.Name = fmt.Sprintf("Corporate Account %d", i)
		}

		// Parse card numbers
		cards := os.Getenv(cardsKey)
		if cards != "" {
			account.CardNumbers = strings.Split(cards, ",")
			for j := range account.CardNumbers {
				account.CardNumbers[j] = strings.TrimSpace(account.CardNumbers[j])
			}
		}

		config.Accounts = append(config.Accounts, account)
	}

	// Load personal accounts (ETC_PERSONAL_USER_1, ETC_PERSONAL_PASS_1, etc.)
	for i := 1; i <= 10; i++ {
		userKey := fmt.Sprintf("ETC_PERSONAL_USER_%d", i)
		passKey := fmt.Sprintf("ETC_PERSONAL_PASS_%d", i)
		nameKey := fmt.Sprintf("ETC_PERSONAL_NAME_%d", i)
		cardsKey := fmt.Sprintf("ETC_PERSONAL_CARDS_%d", i)

		userID := os.Getenv(userKey)
		if userID == "" {
			continue
		}

		account := ETCAccount{
			Name:     os.Getenv(nameKey),
			UserID:   userID,
			Password: os.Getenv(passKey),
			Type:     AccountTypePersonal,
			Active:   true,
		}

		if account.Name == "" {
			account.Name = fmt.Sprintf("Personal Account %d", i)
		}

		// Parse card numbers
		cards := os.Getenv(cardsKey)
		if cards != "" {
			account.CardNumbers = strings.Split(cards, ",")
			for j := range account.CardNumbers {
				account.CardNumbers[j] = strings.TrimSpace(account.CardNumbers[j])
			}
		}

		config.Accounts = append(config.Accounts, account)
	}

	// Also check for simple single account (backward compatibility)
	if len(config.Accounts) == 0 {
		if userID := os.Getenv("ETC_USER_ID"); userID != "" {
			account := ETCAccount{
				Name:     "Default Account",
				UserID:   userID,
				Password: os.Getenv("ETC_PASSWORD"),
				Type:     AccountTypePersonal,
				Active:   true,
			}
			config.Accounts = append(config.Accounts, account)
		}
	}

	return config, nil
}

// LoadAccountsFromFile loads accounts from a JSON file
func LoadAccountsFromFile(filepath string) (*AccountsConfig, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open accounts file: %w", err)
	}
	defer file.Close()

	var config AccountsConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode accounts file: %w", err)
	}

	return &config, nil
}

// GetActiveAccounts returns only active accounts
func (c *AccountsConfig) GetActiveAccounts() []ETCAccount {
	active := []ETCAccount{}
	for _, account := range c.Accounts {
		if account.Active {
			active = append(active, account)
		}
	}
	return active
}

// GetCorporateAccounts returns only corporate accounts
func (c *AccountsConfig) GetCorporateAccounts() []ETCAccount {
	corp := []ETCAccount{}
	for _, account := range c.Accounts {
		if account.Type == AccountTypeCorporate && account.Active {
			corp = append(corp, account)
		}
	}
	return corp
}

// GetPersonalAccounts returns only personal accounts
func (c *AccountsConfig) GetPersonalAccounts() []ETCAccount {
	personal := []ETCAccount{}
	for _, account := range c.Accounts {
		if account.Type == AccountTypePersonal && account.Active {
			personal = append(personal, account)
		}
	}
	return personal
}