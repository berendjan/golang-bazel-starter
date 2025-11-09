package repository

import (
	"context"

	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

// AccountRepository defines the interface for account operations
type AccountRepository interface {
	// CreateAccount creates a new account and returns the account configuration
	CreateAccount(ctx context.Context, accountID []byte, accountType uint32) (*configpb.AccountConfigurationProto, error)

	// DeleteAccount deletes an account by ID
	// Returns the number of rows affected
	DeleteAccount(ctx context.Context, accountID []byte) (int64, error)

	// ListAccounts retrieves all accounts ordered by creation time
	ListAccounts(ctx context.Context) ([]*configpb.AccountConfigurationProto, error)
}
