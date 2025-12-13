package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
	gw "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1/gateway"
)

var (
	clientInstance *ConfigurationClient
	clientOnce     sync.Once
)

// ConfigurationClient is a client for the Configuration service
type ConfigurationClient struct {
	conn   *grpc.ClientConn
	client gw.ConfigurationClient
}

// Config holds client configuration
type Config struct {
	// ServerAddress is the gRPC server address (default: "localhost:25000")
	ServerAddress string

	// Insecure determines whether to use insecure connection (default: true)
	Insecure bool
}

// DefaultConfig returns default client configuration
func DefaultConfig() *Config {
	return &Config{
		ServerAddress: "localhost:25000",
		Insecure:      true,
	}
}

// NewClient creates a new Configuration service client
func NewClient(ctx context.Context, cfg *Config) (*ConfigurationClient, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use passthrough resolver for localhost to avoid slow DNS resolution
	target := cfg.ServerAddress
	if strings.HasPrefix(target, "localhost") || strings.HasPrefix(target, "127.0.0.1") {
		target = "passthrough:///" + target
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &ConfigurationClient{
		conn:   conn,
		client: gw.NewConfigurationClient(conn),
	}, nil
}

// MustNewClient creates a new client or panics on error
func MustNewClient(ctx context.Context, cfg *Config) *ConfigurationClient {
	client, err := NewClient(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create client: %v", err))
	}
	return client
}

// GetClient returns a singleton Configuration service client
// It automatically creates the client on first call with default configuration
// Panics if client creation fails
func GetClient() *ConfigurationClient {
	clientOnce.Do(func() {
		ctx := context.Background()
		cfg := DefaultConfig()

		clientInstance = MustNewClient(ctx, cfg)
	})
	return clientInstance
}

// Close closes the client connection
func (c *ConfigurationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateAccount creates a new account
func (c *ConfigurationClient) CreateAccount(ctx context.Context, name string) (*configpb.AccountConfigurationProto, error) {
	req := &configpb.AccountCreationRequestProto{
		Name: name,
	}

	resp, err := c.client.CreateAccount(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return resp, nil
}

// DeleteAccount deletes an account by ID
func (c *ConfigurationClient) DeleteAccount(ctx context.Context, accountID string) (*commonpb.StatusResponseProto, error) {
	req := &configpb.AccountDeletionRequestProto{
		Id: accountID,
	}

	resp, err := c.client.DeleteAccount(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete account: %w", err)
	}

	return resp, nil
}

// ListAccounts lists all accounts
func (c *ConfigurationClient) ListAccounts(ctx context.Context) ([]*configpb.AccountConfigurationProto, error) {
	req := &configpb.ListAccountsRequestProto{}

	resp, err := c.client.ListAccounts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	return resp.GetAccounts(), nil
}
