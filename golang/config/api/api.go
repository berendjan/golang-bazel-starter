package api

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/berendjan/golang-bazel-starter/golang/config/interfaces"
	"github.com/berendjan/golang-bazel-starter/golang/config/repository"
	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
	gw "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1/gateway"
)

// ConfigurationApi implements the Configuration gRPC service
type ConfigurationApi[T interfaces.AccountRepository] struct {
	gw.UnimplementedConfigurationServer

	accountRepo T
}

type AccountApiProvider[T interfaces.AccountRepository] interface {
	GetAccountApi() *ConfigurationApi[T]
}

// Build creates a new Configuration service Api
func NewConfigurationApi[T interfaces.AccountRepository](accountRepoProvider repository.AccountRepositoryProvider[T]) *ConfigurationApi[T] {
	return &ConfigurationApi[T]{
		accountRepo: accountRepoProvider.GetAccountRepository(),
	}
}

// CreateAccount creates a new account
func (s *ConfigurationApi[T]) CreateAccount(
	ctx context.Context,
	req *configpb.AccountCreationRequestProto,
) (*configpb.AccountConfigurationProto, error) {
	// Validate request
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Pass proto message directly to repository
	account, err := s.accountRepo.CreateAccount(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	log.Printf("Created account: %s", req.GetName())
	return account, nil
}

// DeleteAccount deletes an account
func (s *ConfigurationApi[T]) DeleteAccount(
	ctx context.Context,
	req *configpb.AccountDeletionRequestProto,
) (*commonpb.StatusResponseProto, error) {
	// The ID from HTTP gateway comes base64-encoded, decode it
	accountKey := req.GetId()

	// Try to decode from base64 (HTTP gateway sends it encoded)
	if decoded, err := base64.StdEncoding.DecodeString(accountKey); err == nil {
		accountKey = string(decoded)
		// Update the request with decoded ID
		req.Id = accountKey
	}

	// Pass proto message directly to repository
	response, err := s.accountRepo.DeleteAccount(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}

	log.Printf("Deleted account: %s", accountKey)
	return response, nil
}

// ListAccounts lists all accounts
func (s *ConfigurationApi[T]) ListAccounts(
	ctx context.Context,
	req *configpb.ListAccountsRequestProto,
) (*configpb.ListAccountsResponseProto, error) {
	// Pass proto message directly to repository
	response, err := s.accountRepo.ListAccounts(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	return response, nil
}

// RegisterGRPC implements server_builder.GRPCServiceRegistrar
func (s *ConfigurationApi[T]) RegisterGRPC(Api grpc.ServiceRegistrar) {
	gw.RegisterConfigurationServer(Api, s)
}

// RegisterGateway implements server_builder.HTTPGatewayRegistrar
func (s *ConfigurationApi[T]) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	return gw.RegisterConfigurationHandlerServer(ctx, mux, s)
}
