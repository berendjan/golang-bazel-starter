package config

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sync"

	"github.com/berendjan/golang-bazel-starter/golang/framework/db"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/berendjan/golang-bazel-starter/golang/config/repository"

	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	pool *pgxpool.Pool
	once sync.Once
)

// GetPool returns a singleton database connection pool
// It automatically runs migrations and creates the pool on first call
// Panics if pool creation fails
func GetPool() *pgxpool.Pool {
	once.Do(func() {
		ctx := context.Background()

		// Setup database configuration
		dbConfig := db.DefaultConfig()
		// Override defaults if needed:
		// dbConfig.Host = "localhost"
		// dbConfig.Database = "myapp"

		// Run migrations
		db.MustRunMigrations(dbConfig.ConnectionString(), migrationsFS)

		// Create database connection pool
		pool = db.MustNewPool(ctx, dbConfig)
	})
	return pool
}

// AccountDbRepository implements the AccountRepository interface
type AccountDbRepository struct {
	pool *pgxpool.Pool
}

// Compile-time check that AccountDbRepository implements AccountRepository
var _ repository.AccountRepository = (*AccountDbRepository)(nil)

// NewAccountRepository creates a new AccountRepository implementation
func NewAccountRepository(pool *pgxpool.Pool) *AccountDbRepository {
	return &AccountDbRepository{
		pool: pool,
	}
}

// CreateAccount creates a new account and returns the account configuration
func (r *AccountDbRepository) CreateAccount(ctx context.Context, accountID []byte, accountType uint32) (*configpb.AccountConfigurationProto, error) {
	query := `
		INSERT INTO accounts (id, type, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt string
	err := r.pool.QueryRow(ctx, query, accountID, accountType).Scan(&createdAt, &updatedAt)
	if err != nil {
		log.Printf("Failed to create account in database: %v", err)
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	account := &configpb.AccountConfigurationProto{
		AccountId: &commonpb.ConfigurationIdProto{
			Id:   accountID,
			Type: accountType,
		},
	}

	log.Printf("Created account with id %s", string(accountID))
	return account, nil
}

// DeleteAccount deletes an account by ID and returns the number of rows affected
func (r *AccountDbRepository) DeleteAccount(ctx context.Context, accountID []byte) (int64, error) {
	query := `DELETE FROM accounts WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, accountID)
	if err != nil {
		log.Printf("Failed to delete account from database: %v", err)
		return 0, fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Deleted account: %s", string(accountID))
	}

	return rowsAffected, nil
}

// ListAccounts retrieves all accounts ordered by creation time
func (r *AccountDbRepository) ListAccounts(ctx context.Context) ([]*configpb.AccountConfigurationProto, error) {
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
		var createdAt, updatedAt string

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
	return accounts, nil
}
