package interfaces

import (
	"context"

	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

// AccountRepository defines the interface for account operations
type AccountRepository interface {

	// CreateAccount creates a new account and returns the account configuration
	CreateAccount(ctx context.Context, req *configpb.AccountCreationRequestProto) (*configpb.AccountConfigurationProto, error)

	// DeleteAccount deletes an account by ID
	// Returns the status response
	DeleteAccount(ctx context.Context, req *configpb.AccountDeletionRequestProto) (*commonpb.StatusResponseProto, error)

	// ListAccounts retrieves all accounts ordered by creation time
	ListAccounts(ctx context.Context, req *configpb.ListAccountsRequestProto) (*configpb.ListAccountsResponseProto, error)
}
