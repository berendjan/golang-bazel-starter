package client

import (
	"context"
	"testing"
)

// TestCreateAccount tests creating an account via gRPC
func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	c := GetClient()

	// Create account
	account, err := c.CreateAccount(ctx, "test-account")
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	if account == nil {
		t.Fatal("Account should not be nil")
	}

	if account.GetAccountId() == nil {
		t.Fatal("Account ID should not be nil")
	}

	if len(account.GetAccountId().GetId()) == 0 {
		t.Fatal("Account ID should not be empty")
	}

	t.Logf("Created account with ID: %x", account.GetAccountId().GetId())

	// Clean up - delete the account
	accountIDStr := string(account.GetAccountId().GetId())
	_, err = c.DeleteAccount(ctx, accountIDStr)
	if err != nil {
		t.Logf("Warning: Failed to clean up account: %v", err)
	}
}

// TestListAccounts tests listing accounts via gRPC
func TestListAccounts(t *testing.T) {
	ctx := context.Background()
	c := GetClient()

	// Create a test account first
	account, err := c.CreateAccount(ctx, "test-list-account")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	defer func() {
		accountIDStr := string(account.GetAccountId().GetId())
		c.DeleteAccount(ctx, accountIDStr)
	}()

	// List accounts
	accounts, err := c.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts) == 0 {
		t.Fatal("Expected at least one account")
	}

	// Verify our account is in the list
	found := false
	accountID := string(account.GetAccountId().GetId())
	for _, acc := range accounts {
		if string(acc.GetAccountId().GetId()) == accountID {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Created account not found in list")
	}

	t.Logf("Found %d accounts", len(accounts))
}

// TestDeleteAccount tests deleting an account via gRPC
func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()
	c := GetClient()

	// Create account first
	account, err := c.CreateAccount(ctx, "test-delete-account")
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	accountIDStr := string(account.GetAccountId().GetId())

	// Delete the account
	status, err := c.DeleteAccount(ctx, accountIDStr)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	if status == nil {
		t.Fatal("Status should not be nil")
	}

	if status.GetCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", status.GetCode())
	}

	t.Logf("Delete status: %s (code: %d)", status.GetMessage(), status.GetCode())
}

// TestAccountLifecycle tests the complete lifecycle of an account
func TestAccountLifecycle(t *testing.T) {
	ctx := context.Background()
	c := GetClient()

	// 1. List initial accounts
	initialAccounts, err := c.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list initial accounts: %v", err)
	}
	initialCount := len(initialAccounts)
	t.Logf("Initial account count: %d", initialCount)

	// 2. Create an account
	account, err := c.CreateAccount(ctx, "lifecycle-test-account")
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	accountID := string(account.GetAccountId().GetId())
	t.Logf("Created account: %s", accountID)

	// 3. Verify account appears in list
	afterCreateAccounts, err := c.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts after create: %v", err)
	}
	if len(afterCreateAccounts) != initialCount+1 {
		t.Fatalf("Expected %d accounts, got %d", initialCount+1, len(afterCreateAccounts))
	}

	// 4. Delete the account
	status, err := c.DeleteAccount(ctx, accountID)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}
	if status.GetCode() != 200 {
		t.Fatalf("Expected status code 200, got %d", status.GetCode())
	}
	t.Logf("Deleted account: %s", accountID)

	// 5. Verify account no longer in list
	afterDeleteAccounts, err := c.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts after delete: %v", err)
	}
	if len(afterDeleteAccounts) != initialCount {
		t.Fatalf("Expected %d accounts after delete, got %d", initialCount, len(afterDeleteAccounts))
	}

	t.Log("Account lifecycle test completed successfully")
}
