package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
	gw "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1/gateway"
)

// ConfigurationServer implements the Configuration gRPC service
type ConfigurationServer[T repository.AccountRepository] struct {
	gw.UnimplementedConfigurationServer

	accountRepo T
}

// NewConfigurationServer creates a new Configuration service server
func NewConfigurationServer[T repository.AccountRepository](accountRepo T) *ConfigurationServer[T] {
	return &ConfigurationServer[T]{
		accountRepo: accountRepo,
	}
}

// CreateAccount creates a new account
func (s *ConfigurationServer[T]) CreateAccount(
	ctx context.Context,
	req *configpb.AccountCreationRequestProto,
) (*configpb.AccountConfigurationProto, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Generate account ID
	accountID := []byte(fmt.Sprintf("account-%s", req.GetName()))
	accountType := uint32(1) // Account type

	// Create account via repository
	account, err := s.accountRepo.CreateAccount(ctx, accountID, accountType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	log.Printf("Created account: %s, with id %s", req.GetName(), string(accountID))
	return account, nil
}

// DeleteAccount deletes an account
func (s *ConfigurationServer[T]) DeleteAccount(
	ctx context.Context,
	req *configpb.AccountDeletionRequestProto,
) (*commonpb.StatusResponseProto, error) {
	// The ID from HTTP gateway comes base64-encoded, decode it
	accountKey := req.GetId()

	// Try to decode from base64 (HTTP gateway sends it encoded)
	if decoded, err := base64.StdEncoding.DecodeString(accountKey); err == nil {
		accountKey = string(decoded)
	}

	// Delete account via repository
	rowsAffected, err := s.accountRepo.DeleteAccount(ctx, []byte(accountKey))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}

	// Check if account was found
	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "account not found: "+accountKey)
	}

	log.Printf("Deleted account: %s", accountKey)

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Account deleted successfully",
	}, nil
}

// ListAccounts lists all accounts
func (s *ConfigurationServer[T]) ListAccounts(
	ctx context.Context,
	req *configpb.ListAccountsRequestProto,
) (*configpb.ListAccountsResponseProto, error) {
	// List accounts via repository
	accounts, err := s.accountRepo.ListAccounts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	return &configpb.ListAccountsResponseProto{
		Accounts: accounts,
	}, nil
}

// RegisterGRPC implements server_builder.GRPCServiceRegistrar
func (s *ConfigurationServer[T]) RegisterGRPC(server grpc.ServiceRegistrar) {
	gw.RegisterConfigurationServer(server, s)
}

// RegisterGateway implements server_builder.HTTPGatewayRegistrar
func (s *ConfigurationServer[T]) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	return gw.RegisterConfigurationHandlerServer(ctx, mux, s)
}
