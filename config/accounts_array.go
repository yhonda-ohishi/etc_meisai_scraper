package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadAccountsFromArrayEnv loads accounts from array-style environment variables
func LoadAccountsFromArrayEnv() (*AccountsConfig, error) {
	config := &AccountsConfig{
		Accounts: []ETCAccount{},
	}

	// Load corporate accounts from JSON array
	// ETC_CORP_ACCOUNTS='[{"name":"Company A","user_id":"user1","password":"pass1","password_corp":"corp1","cards":["1234","5678"]}]'
	if corpJSON := os.Getenv("ETC_CORP_ACCOUNTS"); corpJSON != "" {
		var corpAccounts []ETCAccount
		if err := json.Unmarshal([]byte(corpJSON), &corpAccounts); err != nil {
			// Try parsing as comma-separated format for simpler input
			corpAccounts = parseSimpleFormat(corpJSON, AccountTypeCorporate)
		}

		// Set type for all corporate accounts
		for i := range corpAccounts {
			corpAccounts[i].Type = AccountTypeCorporate
			corpAccounts[i].Active = true
		}
		config.Accounts = append(config.Accounts, corpAccounts...)
	}

	// Load personal accounts from JSON array
	// ETC_PERSONAL_ACCOUNTS='[{"name":"Personal 1","user_id":"user1","password":"pass1","cards":["1234"]}]'
	if personalJSON := os.Getenv("ETC_PERSONAL_ACCOUNTS"); personalJSON != "" {
		var personalAccounts []ETCAccount
		if err := json.Unmarshal([]byte(personalJSON), &personalAccounts); err != nil {
			// Try parsing as comma-separated format
			personalAccounts = parseSimpleFormat(personalJSON, AccountTypePersonal)
		}

		// Set type for all personal accounts
		for i := range personalAccounts {
			personalAccounts[i].Type = AccountTypePersonal
			personalAccounts[i].Active = true
		}
		config.Accounts = append(config.Accounts, personalAccounts...)
	}

	// Alternative: Load from simple delimited format
	// ETC_ACCOUNTS="corp,Company A,user1,pass1,corp1,1234-5678|personal,Personal 1,user2,pass2,,9876-5432"
	if simpleAccounts := os.Getenv("ETC_ACCOUNTS"); simpleAccounts != "" {
		accounts := parseDelimitedAccounts(simpleAccounts)
		config.Accounts = append(config.Accounts, accounts...)
	}

	// Alternative: Load from separate arrays
	// ETC_ACCOUNT_TYPES="corporate,personal,corporate"
	// ETC_ACCOUNT_NAMES="Company A,Personal 1,Company B"
	// ETC_ACCOUNT_USERS="user1,user2,user3"
	// ETC_ACCOUNT_PASSWORDS="pass1,pass2,pass3"
	// ETC_ACCOUNT_CORP_PASSWORDS="corp1,,corp3"
	// ETC_ACCOUNT_CARDS="1234-5678;2345-6789,9876-5432,3456-7890"
	if types := os.Getenv("ETC_ACCOUNT_TYPES"); types != "" {
		accounts := parseArrayEnvVars()
		config.Accounts = append(config.Accounts, accounts...)
	}

	// Fallback to numbered env vars if no array format found
	if len(config.Accounts) == 0 {
		return LoadAccountsFromEnv()
	}

	return config, nil
}

// parseSimpleFormat parses a simple comma-separated format
// Format: "name:user:pass[:corp_pass]:card1,card2;name2:user2:pass2..."
func parseSimpleFormat(input string, accountType AccountType) []ETCAccount {
	var accounts []ETCAccount

	// Split by semicolon for multiple accounts
	accountStrings := strings.Split(input, ";")

	for _, accStr := range accountStrings {
		parts := strings.Split(accStr, ":")
		if len(parts) < 3 {
			continue
		}

		account := ETCAccount{
			Name:     strings.TrimSpace(parts[0]),
			UserID:   strings.TrimSpace(parts[1]),
			Password: strings.TrimSpace(parts[2]),
			Type:     accountType,
			Active:   true,
		}

		// Corporate password (optional)
		if len(parts) > 3 && accountType == AccountTypeCorporate {
			account.PasswordCorp = strings.TrimSpace(parts[3])
		}

		// Cards (optional)
		startIdx := 3
		if accountType == AccountTypeCorporate {
			startIdx = 4
		}

		if len(parts) > startIdx {
			cards := strings.Split(parts[startIdx], ",")
			for _, card := range cards {
				card = strings.TrimSpace(card)
				if card != "" {
					account.CardNumbers = append(account.CardNumbers, card)
				}
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

// parseDelimitedAccounts parses pipe-delimited account format
// Format: "type,name,user,pass,corp_pass,cards|type,name,user,pass,corp_pass,cards"
func parseDelimitedAccounts(input string) []ETCAccount {
	var accounts []ETCAccount

	// Split by pipe for multiple accounts
	accountStrings := strings.Split(input, "|")

	for _, accStr := range accountStrings {
		parts := strings.Split(accStr, ",")
		if len(parts) < 4 {
			continue
		}

		accountType := AccountTypePersonal
		if strings.ToLower(strings.TrimSpace(parts[0])) == "corporate" ||
		   strings.ToLower(strings.TrimSpace(parts[0])) == "corp" {
			accountType = AccountTypeCorporate
		}

		account := ETCAccount{
			Name:     strings.TrimSpace(parts[1]),
			UserID:   strings.TrimSpace(parts[2]),
			Password: strings.TrimSpace(parts[3]),
			Type:     accountType,
			Active:   true,
		}

		// Corporate password
		if len(parts) > 4 && parts[4] != "" {
			account.PasswordCorp = strings.TrimSpace(parts[4])
		}

		// Cards (semicolon separated)
		if len(parts) > 5 && parts[5] != "" {
			cards := strings.Split(parts[5], ";")
			for _, card := range cards {
				card = strings.TrimSpace(card)
				if card != "" {
					account.CardNumbers = append(account.CardNumbers, card)
				}
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

// parseArrayEnvVars parses parallel array environment variables
func parseArrayEnvVars() []ETCAccount {
	var accounts []ETCAccount

	types := strings.Split(os.Getenv("ETC_ACCOUNT_TYPES"), ",")
	names := strings.Split(os.Getenv("ETC_ACCOUNT_NAMES"), ",")
	users := strings.Split(os.Getenv("ETC_ACCOUNT_USERS"), ",")
	passwords := strings.Split(os.Getenv("ETC_ACCOUNT_PASSWORDS"), ",")
	corpPasswords := strings.Split(os.Getenv("ETC_ACCOUNT_CORP_PASSWORDS"), ",")
	cardsArray := strings.Split(os.Getenv("ETC_ACCOUNT_CARDS"), ",")

	// Find the maximum length
	maxLen := len(types)
	if len(users) > maxLen {
		maxLen = len(users)
	}

	for i := 0; i < maxLen; i++ {
		account := ETCAccount{
			Active: true,
		}

		// Type
		if i < len(types) {
			typeStr := strings.ToLower(strings.TrimSpace(types[i]))
			if typeStr == "corporate" || typeStr == "corp" {
				account.Type = AccountTypeCorporate
			} else {
				account.Type = AccountTypePersonal
			}
		}

		// Name
		if i < len(names) {
			account.Name = strings.TrimSpace(names[i])
		} else {
			account.Name = fmt.Sprintf("Account %d", i+1)
		}

		// User ID
		if i < len(users) {
			account.UserID = strings.TrimSpace(users[i])
		} else {
			continue // Skip if no user ID
		}

		// Password
		if i < len(passwords) {
			account.Password = strings.TrimSpace(passwords[i])
		}

		// Corporate password
		if i < len(corpPasswords) && corpPasswords[i] != "" {
			account.PasswordCorp = strings.TrimSpace(corpPasswords[i])
		}

		// Cards (semicolon separated within each account)
		if i < len(cardsArray) && cardsArray[i] != "" {
			cards := strings.Split(cardsArray[i], ";")
			for _, card := range cards {
				card = strings.TrimSpace(card)
				if card != "" {
					account.CardNumbers = append(account.CardNumbers, card)
				}
			}
		}

		accounts = append(accounts, account)
	}

	return accounts
}

// SaveAccountsToJSON saves accounts configuration to a JSON file
func SaveAccountsToJSON(accounts []ETCAccount, filepath string) error {
	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

// GenerateEnvExample generates example environment variable settings
func GenerateEnvExample() string {
	var sb strings.Builder

	sb.WriteString("# ===== Multiple Account Configuration =====\n\n")

	sb.WriteString("# Option 1: JSON Array Format\n")
	sb.WriteString(`ETC_CORP_ACCOUNTS='[{"name":"Company A","user_id":"corp1","password":"pass1","password_corp":"corp_pass1","cards":["1234-5678","2345-6789"]},{"name":"Company B","user_id":"corp2","password":"pass2","password_corp":"corp_pass2","cards":["3456-7890"]}]'`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_PERSONAL_ACCOUNTS='[{"name":"Personal 1","user_id":"user1","password":"pass1","cards":["9876-5432"]},{"name":"Personal 2","user_id":"user2","password":"pass2"}]'`)
	sb.WriteString("\n\n")

	sb.WriteString("# Option 2: Simple Delimited Format\n")
	sb.WriteString(`# Format: type,name,user,pass,corp_pass,cards|type,name,user,pass,corp_pass,cards`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNTS="corp,Company A,corp1,pass1,corp_pass1,1234-5678;2345-6789|personal,Personal 1,user1,pass1,,9876-5432|corp,Company B,corp2,pass2,corp_pass2,3456-7890"`)
	sb.WriteString("\n\n")

	sb.WriteString("# Option 3: Parallel Arrays Format\n")
	sb.WriteString(`ETC_ACCOUNT_TYPES="corporate,personal,corporate"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNT_NAMES="Company A,Personal 1,Company B"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNT_USERS="corp1,user1,corp2"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNT_PASSWORDS="pass1,pass1,pass2"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNT_CORP_PASSWORDS="corp_pass1,,corp_pass2"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_ACCOUNT_CARDS="1234-5678;2345-6789,9876-5432,3456-7890"`)
	sb.WriteString("\n\n")

	sb.WriteString("# Option 4: Simple Colon Format (for single type)\n")
	sb.WriteString(`# Format: name:user:pass[:corp_pass]:card1,card2;name2:user2:pass2...`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_CORP_ACCOUNTS="Company A:corp1:pass1:corp_pass1:1234-5678,2345-6789;Company B:corp2:pass2:corp_pass2:3456-7890"`)
	sb.WriteString("\n")
	sb.WriteString(`ETC_PERSONAL_ACCOUNTS="Personal 1:user1:pass1:9876-5432;Personal 2:user2:pass2"`)
	sb.WriteString("\n")

	return sb.String()
}