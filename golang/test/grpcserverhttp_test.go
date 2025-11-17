package test_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/berendjan/golang-bazel-starter/golang/test"
)

// HTTP Tests using TestContext

func TestHTTPCreateAccount(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	reqBody := map[string]string{
		"name": "http-test-account",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		httpBaseURL+"/v1/accounts",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Failed to create account via HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	accountID, ok := result["accountId"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should contain accountId")
	}

	id, ok := accountID["id"].(string)
	if !ok || id == "" {
		t.Fatal("Account ID should not be empty")
	}

	t.Logf("Created account via HTTP with ID: %s", id)

	// Clean up - delete the account
	deleteReq, _ := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/accounts/%s", httpBaseURL, id),
		nil,
	)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Logf("Warning: Failed to clean up account: %v", err)
	} else {
		deleteResp.Body.Close()
	}
}

func TestHTTPListAccounts(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	// Create a test account first
	reqBody := map[string]string{
		"name": "http-list-test-account",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	createResp, err := http.Post(
		httpBaseURL+"/v1/accounts",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	accountID := createResult["accountId"].(map[string]interface{})["id"].(string)

	defer func() {
		deleteReq, _ := http.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("%s/v1/accounts/%s", httpBaseURL, accountID),
			nil,
		)
		deleteResp, _ := http.DefaultClient.Do(deleteReq)
		if deleteResp != nil {
			deleteResp.Body.Close()
		}
	}()

	// List accounts
	resp, err := http.Get(httpBaseURL + "/v1/accounts")
	if err != nil {
		t.Fatalf("Failed to list accounts via HTTP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	accounts, ok := result["accounts"].([]interface{})
	if !ok {
		t.Fatal("Response should contain accounts array")
	}

	if len(accounts) == 0 {
		t.Fatal("Expected at least one account")
	}

	t.Logf("Found %d accounts via HTTP", len(accounts))
}

func TestHTTPDeleteAccount(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	// Create account first
	reqBody := map[string]string{
		"name": "http-delete-test-account",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	createResp, err := http.Post(
		httpBaseURL+"/v1/accounts",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	accountID := createResult["accountId"].(map[string]interface{})["id"].(string)

	// Delete the account
	deleteReq, _ := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/accounts/%s", httpBaseURL, accountID),
		nil,
	)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(deleteResp.Body)
		t.Fatalf("Expected status 200, got %d: %s", deleteResp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(deleteResp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	code, ok := result["code"].(float64)
	if !ok || code != 200 {
		t.Fatalf("Expected status code 200, got %v", result["code"])
	}

	t.Logf("Delete status via HTTP: %s (code: %.0f)", result["message"], code)
}

func TestHTTPAccountLifecycle(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	// 1. List initial accounts
	initialResp, err := http.Get(httpBaseURL + "/v1/accounts")
	if err != nil {
		t.Fatalf("Failed to list initial accounts: %v", err)
	}
	defer initialResp.Body.Close()

	var initialResult map[string]interface{}
	json.NewDecoder(initialResp.Body).Decode(&initialResult)
	initialAccounts, _ := initialResult["accounts"].([]interface{})
	initialCount := len(initialAccounts)
	t.Logf("Initial account count via HTTP: %d", initialCount)

	// 2. Create an account
	reqBody := map[string]string{
		"name": "http-lifecycle-test-account",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	createResp, err := http.Post(
		httpBaseURL+"/v1/accounts",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	defer createResp.Body.Close()

	var createResult map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&createResult)
	accountID := createResult["accountId"].(map[string]interface{})["id"].(string)
	t.Logf("Created account via HTTP: %s", accountID)

	// 3. Verify account appears in list
	afterCreateResp, err := http.Get(httpBaseURL + "/v1/accounts")
	if err != nil {
		t.Fatalf("Failed to list accounts after create: %v", err)
	}
	defer afterCreateResp.Body.Close()

	var afterCreateResult map[string]interface{}
	json.NewDecoder(afterCreateResp.Body).Decode(&afterCreateResult)
	afterCreateAccounts, _ := afterCreateResult["accounts"].([]interface{})
	if len(afterCreateAccounts) != initialCount+1 {
		t.Fatalf("Expected %d accounts, got %d", initialCount+1, len(afterCreateAccounts))
	}

	// 4. Delete the account
	deleteReq, _ := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/accounts/%s", httpBaseURL, accountID),
		nil,
	)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}
	defer deleteResp.Body.Close()

	var deleteResult map[string]interface{}
	json.NewDecoder(deleteResp.Body).Decode(&deleteResult)
	if deleteResult["code"].(float64) != 200 {
		t.Fatalf("Expected delete status code 200, got %v", deleteResult["code"])
	}
	t.Logf("Deleted account via HTTP: %s", accountID)

	// 5. Verify account no longer in list
	afterDeleteResp, err := http.Get(httpBaseURL + "/v1/accounts")
	if err != nil {
		t.Fatalf("Failed to list accounts after delete: %v", err)
	}
	defer afterDeleteResp.Body.Close()

	var afterDeleteResult map[string]interface{}
	json.NewDecoder(afterDeleteResp.Body).Decode(&afterDeleteResult)
	afterDeleteAccounts, _ := afterDeleteResult["accounts"].([]interface{})
	if len(afterDeleteAccounts) != initialCount {
		t.Fatalf("Expected %d accounts after delete, got %d", initialCount, len(afterDeleteAccounts))
	}

	t.Log("HTTP account lifecycle test completed successfully")
}

func TestHTTPDeleteAccountNotFound(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	// Try to delete a non-existent account
	deleteReq, _ := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/accounts/%s", httpBaseURL, "non-existent-account"),
		nil,
	)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("Failed to send delete request: %v", err)
	}
	defer deleteResp.Body.Close()

	// Should get a non-200 status code
	if deleteResp.StatusCode == http.StatusOK {
		t.Fatal("Expected error status when deleting non-existent account")
	}

	t.Logf("Got expected error status: %d", deleteResp.StatusCode)
}

func TestHTTPCreateAccountValidation(t *testing.T) {
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

	httpBaseURL := tc.GetHttpClient(test.GrpcServer)

	// Try to create account with empty name
	reqBody := map[string]string{
		"name": "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		httpBaseURL+"/v1/accounts",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Should get a non-200 status code for validation error
	if resp.StatusCode == http.StatusOK {
		t.Fatal("Expected error status when creating account with empty name")
	}

	t.Logf("Got expected validation error status: %d", resp.StatusCode)
}
