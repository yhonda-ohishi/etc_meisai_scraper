package config_test

import (
	"testing"

	"github.com/yhonda-ohishi/etc_meisai/src/config"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestParseAccounts_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single account",
			input:    "account1",
			expected: []string{"account1"},
		},
		{
			name:     "multiple accounts",
			input:    "account1,account2,account3",
			expected: []string{"account1", "account2", "account3"},
		},
		{
			name:     "accounts with spaces",
			input:    " account1 , account2 , account3 ",
			expected: []string{"account1", "account2", "account3"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "comma only",
			input:    ",",
			expected: []string{},
		},
		{
			name:     "multiple commas",
			input:    "account1,,,account2",
			expected: []string{"account1", "account2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.ParseAccounts(tt.input)
			helpers.AssertEqual(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				helpers.AssertEqual(t, expected, result[i])
			}
		})
	}
}

func TestAccountConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		account *config.AccountConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid corporate account",
			account: &config.AccountConfig{
				Username:    "corp_user",
				Password:    "corp_pass",
				AccountType: "corporate",
				Index:       0,
			},
			wantErr: false,
		},
		{
			name: "valid personal account",
			account: &config.AccountConfig{
				Username:    "personal_user",
				Password:    "personal_pass",
				AccountType: "personal",
				Index:       1,
			},
			wantErr: false,
		},
		{
			name: "empty username",
			account: &config.AccountConfig{
				Username:    "",
				Password:    "password",
				AccountType: "corporate",
				Index:       0,
			},
			wantErr: true,
			errMsg:  "Username is required",
		},
		{
			name: "empty password",
			account: &config.AccountConfig{
				Username:    "username",
				Password:    "",
				AccountType: "corporate",
				Index:       0,
			},
			wantErr: true,
			errMsg:  "Password is required",
		},
		{
			name: "invalid account type",
			account: &config.AccountConfig{
				Username:    "username",
				Password:    "password",
				AccountType: "invalid",
				Index:       0,
			},
			wantErr: true,
			errMsg:  "AccountType must be 'corporate' or 'personal'",
		},
		{
			name: "negative index",
			account: &config.AccountConfig{
				Username:    "username",
				Password:    "password",
				AccountType: "corporate",
				Index:       -1,
			},
			wantErr: true,
			errMsg:  "Index must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.Validate()

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestAccountConfig_String(t *testing.T) {
	account := &config.AccountConfig{
		Username:    "test_user",
		Password:    "secret_password",
		AccountType: "corporate",
		Index:       2,
	}

	str := account.String()

	// Should contain username and type but not password
	helpers.AssertContains(t, str, "test_user")
	helpers.AssertContains(t, str, "corporate")
	helpers.AssertContains(t, str, "2")
	helpers.AssertNotContains(t, str, "secret_password") // Password should be masked
}

func TestAccountConfig_GetIdentifier(t *testing.T) {
	account := &config.AccountConfig{
		Username:    "test_user",
		AccountType: "corporate",
		Index:       2,
	}

	identifier := account.GetIdentifier()
	helpers.AssertEqual(t, "corporate_2_test_user", identifier)
}

func TestAccountConfig_IsCorporate(t *testing.T) {
	tests := []struct {
		name        string
		accountType string
		expected    bool
	}{
		{
			name:        "corporate account",
			accountType: "corporate",
			expected:    true,
		},
		{
			name:        "personal account",
			accountType: "personal",
			expected:    false,
		},
		{
			name:        "invalid account type",
			accountType: "invalid",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &config.AccountConfig{
				AccountType: tt.accountType,
			}
			result := account.IsCorporate()
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestAccountConfig_IsPersonal(t *testing.T) {
	tests := []struct {
		name        string
		accountType string
		expected    bool
	}{
		{
			name:        "personal account",
			accountType: "personal",
			expected:    true,
		},
		{
			name:        "corporate account",
			accountType: "corporate",
			expected:    false,
		},
		{
			name:        "invalid account type",
			accountType: "invalid",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &config.AccountConfig{
				AccountType: tt.accountType,
			}
			result := account.IsPersonal()
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestBuildAccountConfigs_Corporate(t *testing.T) {
	accounts := []string{"corp1", "corp2", "corp3"}
	configs := config.BuildAccountConfigs(accounts, "corporate")

	helpers.AssertLen(t, configs, 3)

	for i, cfg := range configs {
		helpers.AssertEqual(t, accounts[i], cfg.Username)
		helpers.AssertEqual(t, "corporate", cfg.AccountType)
		helpers.AssertEqual(t, i, cfg.Index)
		helpers.AssertEmpty(t, cfg.Password) // Password should be empty by default
	}
}

func TestBuildAccountConfigs_Personal(t *testing.T) {
	accounts := []string{"personal1", "personal2"}
	configs := config.BuildAccountConfigs(accounts, "personal")

	helpers.AssertLen(t, configs, 2)

	for i, cfg := range configs {
		helpers.AssertEqual(t, accounts[i], cfg.Username)
		helpers.AssertEqual(t, "personal", cfg.AccountType)
		helpers.AssertEqual(t, i, cfg.Index)
		helpers.AssertEmpty(t, cfg.Password) // Password should be empty by default
	}
}

func TestBuildAccountConfigs_EmptyList(t *testing.T) {
	accounts := []string{}
	configs := config.BuildAccountConfigs(accounts, "corporate")

	helpers.AssertLen(t, configs, 0)
}

func TestAccountManager_AddAccount(t *testing.T) {
	manager := config.NewAccountManager()

	account := &config.AccountConfig{
		Username:    "test_user",
		Password:    "test_pass",
		AccountType: "corporate",
		Index:       0,
	}

	err := manager.AddAccount(account)
	helpers.AssertNoError(t, err)

	// Should be able to retrieve the account
	retrieved := manager.GetAccount("corporate", 0)
	helpers.AssertNotNil(t, retrieved)
	helpers.AssertEqual(t, "test_user", retrieved.Username)
}

func TestAccountManager_AddAccount_DuplicateIndex(t *testing.T) {
	manager := config.NewAccountManager()

	account1 := &config.AccountConfig{
		Username:    "user1",
		Password:    "pass1",
		AccountType: "corporate",
		Index:       0,
	}

	account2 := &config.AccountConfig{
		Username:    "user2",
		Password:    "pass2",
		AccountType: "corporate",
		Index:       0, // Same index
	}

	err := manager.AddAccount(account1)
	helpers.AssertNoError(t, err)

	err = manager.AddAccount(account2)
	helpers.AssertError(t, err)
	helpers.AssertContains(t, err.Error(), "already exists")
}

func TestAccountManager_GetAccount(t *testing.T) {
	manager := config.NewAccountManager()

	account := &config.AccountConfig{
		Username:    "test_user",
		Password:    "test_pass",
		AccountType: "personal",
		Index:       1,
	}

	manager.AddAccount(account)

	// Should find the account
	retrieved := manager.GetAccount("personal", 1)
	helpers.AssertNotNil(t, retrieved)
	helpers.AssertEqual(t, "test_user", retrieved.Username)

	// Should not find non-existent account
	notFound := manager.GetAccount("personal", 99)
	helpers.AssertNil(t, notFound)

	// Should not find with wrong type
	notFound = manager.GetAccount("corporate", 1)
	helpers.AssertNil(t, notFound)
}

func TestAccountManager_GetAllAccounts(t *testing.T) {
	manager := config.NewAccountManager()

	accounts := []*config.AccountConfig{
		{Username: "corp1", Password: "pass1", AccountType: "corporate", Index: 0},
		{Username: "corp2", Password: "pass2", AccountType: "corporate", Index: 1},
		{Username: "personal1", Password: "pass3", AccountType: "personal", Index: 0},
	}

	for _, account := range accounts {
		manager.AddAccount(account)
	}

	allAccounts := manager.GetAllAccounts()
	helpers.AssertLen(t, allAccounts, 3)
}

func TestAccountManager_GetAccountsByType(t *testing.T) {
	manager := config.NewAccountManager()

	accounts := []*config.AccountConfig{
		{Username: "corp1", Password: "pass1", AccountType: "corporate", Index: 0},
		{Username: "corp2", Password: "pass2", AccountType: "corporate", Index: 1},
		{Username: "personal1", Password: "pass3", AccountType: "personal", Index: 0},
	}

	for _, account := range accounts {
		manager.AddAccount(account)
	}

	corporateAccounts := manager.GetAccountsByType("corporate")
	helpers.AssertLen(t, corporateAccounts, 2)

	personalAccounts := manager.GetAccountsByType("personal")
	helpers.AssertLen(t, personalAccounts, 1)

	invalidAccounts := manager.GetAccountsByType("invalid")
	helpers.AssertLen(t, invalidAccounts, 0)
}

func TestAccountManager_RemoveAccount(t *testing.T) {
	manager := config.NewAccountManager()

	account := &config.AccountConfig{
		Username:    "test_user",
		Password:    "test_pass",
		AccountType: "corporate",
		Index:       0,
	}

	manager.AddAccount(account)

	// Should exist before removal
	retrieved := manager.GetAccount("corporate", 0)
	helpers.AssertNotNil(t, retrieved)

	// Remove the account
	removed := manager.RemoveAccount("corporate", 0)
	helpers.AssertTrue(t, removed)

	// Should not exist after removal
	retrieved = manager.GetAccount("corporate", 0)
	helpers.AssertNil(t, retrieved)

	// Removing non-existent account should return false
	removed = manager.RemoveAccount("corporate", 0)
	helpers.AssertFalse(t, removed)
}