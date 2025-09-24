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

// ParseAccounts parses a comma-separated string into a slice of account names
func ParseAccounts(input string) []string {
	if input == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	parts := strings.Split(input, ",")
	accounts := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			accounts = append(accounts, trimmed)
		}
	}

	return accounts
}

// AccountConfig represents a single ETC account configuration
type AccountConfig struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	AccountType string `json:"account_type"` // "corporate" or "personal"
	Index       int    `json:"index"`
}

// Validate validates the account configuration
func (a *AccountConfig) Validate() error {
	if a.Username == "" {
		return fmt.Errorf("Username is required")
	}
	if a.Password == "" {
		return fmt.Errorf("Password is required")
	}
	if a.AccountType != "corporate" && a.AccountType != "personal" {
		return fmt.Errorf("AccountType must be 'corporate' or 'personal'")
	}
	if a.Index < 0 {
		return fmt.Errorf("Index must be non-negative")
	}
	return nil
}

// String returns a string representation of the account (masks password)
func (a *AccountConfig) String() string {
	return fmt.Sprintf("AccountConfig{Username: %s, AccountType: %s, Index: %d, Password: ***}",
		a.Username, a.AccountType, a.Index)
}

// GetIdentifier returns a unique identifier for this account
func (a *AccountConfig) GetIdentifier() string {
	return fmt.Sprintf("%s_%d_%s", a.AccountType, a.Index, a.Username)
}

// IsCorporate returns true if this is a corporate account
func (a *AccountConfig) IsCorporate() bool {
	return a.AccountType == "corporate"
}

// IsPersonal returns true if this is a personal account
func (a *AccountConfig) IsPersonal() bool {
	return a.AccountType == "personal"
}

// BuildAccountConfigs builds account configs from a slice of usernames
func BuildAccountConfigs(usernames []string, accountType string) []*AccountConfig {
	configs := make([]*AccountConfig, len(usernames))
	for i, username := range usernames {
		configs[i] = &AccountConfig{
			Username:    username,
			AccountType: accountType,
			Index:       i,
		}
	}
	return configs
}

// AccountManager manages a collection of account configurations
type AccountManager struct {
	accounts map[string]*AccountConfig // key: accountType_index
}

// NewAccountManager creates a new account manager
func NewAccountManager() *AccountManager {
	return &AccountManager{
		accounts: make(map[string]*AccountConfig),
	}
}

// AddAccount adds an account to the manager
func (m *AccountManager) AddAccount(account *AccountConfig) error {
	if err := account.Validate(); err != nil {
		return err
	}

	key := fmt.Sprintf("%s_%d", account.AccountType, account.Index)
	if _, exists := m.accounts[key]; exists {
		return fmt.Errorf("account with type %s and index %d already exists", account.AccountType, account.Index)
	}

	m.accounts[key] = account
	return nil
}

// GetAccount retrieves an account by type and index
func (m *AccountManager) GetAccount(accountType string, index int) *AccountConfig {
	key := fmt.Sprintf("%s_%d", accountType, index)
	return m.accounts[key]
}

// GetAllAccounts returns all accounts
func (m *AccountManager) GetAllAccounts() []*AccountConfig {
	accounts := make([]*AccountConfig, 0, len(m.accounts))
	for _, account := range m.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

// GetAccountsByType returns all accounts of a specific type
func (m *AccountManager) GetAccountsByType(accountType string) []*AccountConfig {
	accounts := make([]*AccountConfig, 0)
	for _, account := range m.accounts {
		if account.AccountType == accountType {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

// RemoveAccount removes an account by type and index
func (m *AccountManager) RemoveAccount(accountType string, index int) bool {
	key := fmt.Sprintf("%s_%d", accountType, index)
	if _, exists := m.accounts[key]; exists {
		delete(m.accounts, key)
		return true
	}
	return false
}