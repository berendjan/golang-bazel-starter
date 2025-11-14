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

func TestBuilderDeleteAccount(t *testing.T) {
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
