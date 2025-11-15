package test_test

import (
	"context"
	"testing"

	configClient "github.com/berendjan/golang-bazel-starter/golang/config/client"
	"github.com/berendjan/golang-bazel-starter/golang/test"
)

// TestBuilderWithServers demonstrates using the builder to create servers
func TestCreateAccount(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Send request to server with client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	testName := "test account"

	acc, err := client.CreateAccount(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	name := string(acc.AccountId.Id)

	if name != testName {
		t.Fatalf("Returning name does not match")
	}
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	testName := "account-to-delete"

	// First, create an account
	acc, err := client.CreateAccount(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	accountID := string(acc.AccountId.Id)
	if accountID != testName {
		t.Fatalf("Created account ID does not match: got %s, want %s", accountID, testName)
	}

	// Delete the account
	deleteResp, err := client.DeleteAccount(ctx, accountID)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	if deleteResp.Code != 200 {
		t.Fatalf("Expected status code 200, got %d: %s", deleteResp.Code, deleteResp.Message)
	}

	if deleteResp.Message != "Account deleted successfully" {
		t.Fatalf("Unexpected delete message: %s", deleteResp.Message)
	}
}

func TestDeleteAccountNotFound(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	// Try to delete a non-existent account
	_, err = client.DeleteAccount(ctx, "non-existent-account")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent account, got nil")
	}

	// The error should indicate the account was not found
	if err.Error() == "" {
		t.Fatal("Error message should not be empty")
	}
	t.Logf("Got expected error: %v", err)
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	// Initially, list should be empty
	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	initialCount := len(accounts)
	t.Logf("Initial account count: %d", initialCount)

	// Create multiple accounts
	testAccounts := []string{"account-1", "account-2", "account-3"}
	for _, name := range testAccounts {
		_, err := client.CreateAccount(ctx, name)
		if err != nil {
			t.Fatalf("Failed to create account %s: %v", name, err)
		}
	}

	// List accounts again
	accounts, err = client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts after creation: %v", err)
	}

	expectedCount := initialCount + len(testAccounts)
	if len(accounts) != expectedCount {
		t.Fatalf("Expected %d accounts, got %d", expectedCount, len(accounts))
	}

	// Verify all test accounts are in the list
	accountMap := make(map[string]bool)
	for _, acc := range accounts {
		accountMap[string(acc.AccountId.Id)] = true
	}

	for _, name := range testAccounts {
		if !accountMap[name] {
			t.Errorf("Account %s not found in list", name)
		}
	}
}

func TestListAccountsEmpty(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	// List accounts on a fresh database (should be empty or return without error)
	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts on empty database: %v", err)
	}

	t.Logf("Empty database has %d accounts", len(accounts))
}

func TestAccountLifecycle(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	testName := "lifecycle-account"

	// 1. Create account
	acc, err := client.CreateAccount(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	t.Logf("Created account: %s", string(acc.AccountId.Id))

	// 2. Verify it appears in list
	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	found := false
	for _, a := range accounts {
		if string(a.AccountId.Id) == testName {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Created account not found in list")
	}

	// 3. Delete the account
	deleteResp, err := client.DeleteAccount(ctx, testName)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}
	if deleteResp.Code != 200 {
		t.Fatalf("Delete failed with code %d: %s", deleteResp.Code, deleteResp.Message)
	}
	t.Logf("Deleted account successfully")

	// 4. Verify it no longer appears in list
	accounts, err = client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts after deletion: %v", err)
	}

	for _, a := range accounts {
		if string(a.AccountId.Id) == testName {
			t.Fatal("Deleted account still appears in list")
		}
	}
	t.Logf("Verified account is no longer in list")
}

func TestCreateAccountValidation(t *testing.T) {
	ctx := context.Background()

	tc, err := test.NewTestContextBuilder().
		WithDatabase(test.ConfigDb).
		WithServer(test.GrpcServer).
		Build(ctx)
	if err != nil {
		t.Fatalf("Failed to create test context: %v", err)
	}
	defer func() {
		if err := tc.CleanUp(ctx); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Create a client
	client := configClient.MustNewClient(ctx, &configClient.Config{ServerAddress: tc.GetGrpcClient(test.GrpcServer), Insecure: true})

	// Try to create account with empty name
	_, err = client.CreateAccount(ctx, "")
	if err == nil {
		t.Fatal("Expected error when creating account with empty name, got nil")
	}
	t.Logf("Got expected validation error: %v", err)
}
