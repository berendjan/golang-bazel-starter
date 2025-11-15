package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/berendjan/golang-bazel-starter/golang/config/interfaces"
	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

const (
	DbName string = "config"
)

// AccountDbRepository implements the AccountRepository interface
type AccountDbRepository struct {
	pool *db.DBPool
}

// Compile-time check that AccountDbRepository implements AccountRepository
var _ interfaces.AccountRepository = (*AccountDbRepository)(nil)

// dependency injection provider
type AccountRepositoryProvider[T interfaces.AccountRepository] interface {
	GetAccountRepository() T
}

// NewAccountRepository creates a new AccountRepository implementation
func NewAccountRepository(pool *db.DBPool) *AccountDbRepository {
	return &AccountDbRepository{
		pool: pool,
	}
}

// CreateAccount creates a new account and returns the account configuration
func (r *AccountDbRepository) CreateAccount(ctx context.Context, req *configpb.AccountCreationRequestProto) (*configpb.AccountConfigurationProto, error) {
	if req.GetName() == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Generate account ID from name
	accountID := []byte(req.GetName())
	accountType := uint32(1) // Default account type

	query := `
		INSERT INTO accounts (id, type)
		VALUES ($1, $2)
		RETURNING id, type
	`

	var id []byte
	var accType uint32
	err := r.pool.QueryRow(ctx, query, accountID, accountType).Scan(&id, &accType)
	if err != nil {
		log.Printf("Failed to create account in database: %v", err)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	account := &configpb.AccountConfigurationProto{
		AccountId: &commonpb.ConfigurationIdProto{
			Id:   id,
			Type: accType,
		},
	}

	log.Printf("Created account with id %s", string(accountID))
	return account, nil
}

// DeleteAccount deletes an account by ID and returns status response
func (r *AccountDbRepository) DeleteAccount(ctx context.Context, req *configpb.AccountDeletionRequestProto) (*commonpb.StatusResponseProto, error) {
	accountKey := req.GetId()

	// Try to decode from base64 (HTTP gateway sends it encoded)
	// Removed base64 decoding logic - moved from API layer

	query := `DELETE FROM accounts WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, []byte(accountKey))
	if err != nil {
		log.Printf("Failed to delete account from database: %v", err)
		return nil, fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return &commonpb.StatusResponseProto{
			Code:    404,
			Message: "Account not found: " + accountKey,
		}, fmt.Errorf("account not found: %s", accountKey)
	}

	log.Printf("Deleted account: %s", accountKey)

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Account deleted successfully",
	}, nil
}

// ListAccounts retrieves all accounts ordered by creation time
func (r *AccountDbRepository) ListAccounts(ctx context.Context, req *configpb.ListAccountsRequestProto) (*configpb.ListAccountsResponseProto, error) {
	query := `SELECT id, type, created_at, updated_at FROM accounts ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to list accounts from database: %v", err)
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*configpb.AccountConfigurationProto
	for rows.Next() {
		var id []byte
		var accountType uint32
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &accountType, &createdAt, &updatedAt); err != nil {
			log.Printf("Failed to scan account row: %v", err)
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}

		account := &configpb.AccountConfigurationProto{
			AccountId: &commonpb.ConfigurationIdProto{
				Id:   id,
				Type: accountType,
			},
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating account rows: %v", err)
		return nil, fmt.Errorf("failed to iterate accounts: %w", err)
	}

	log.Printf("Listed %d accounts", len(accounts))
	return &configpb.ListAccountsResponseProto{
		Accounts: accounts,
	}, nil
}
