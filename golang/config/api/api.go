package api

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"
	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
	gw "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1/gateway"
)

// ConfigurationApi implements the Configuration gRPC service
type ConfigurationApi struct {
	gw.UnimplementedConfigurationServer

	accountRepo geninterfaces.AccountApiSendable
}

// Build creates a new Configuration service Api
func NewConfigurationApi(accountRepo geninterfaces.AccountApiSendable) *ConfigurationApi {
	return &ConfigurationApi{
		accountRepo: accountRepo,
	}
}

// CreateAccount creates a new account
func (s *ConfigurationApi) CreateAccount(
	ctx context.Context,
	req *configpb.AccountCreationRequestProto,
) (*configpb.AccountConfigurationProto, error) {
	// Validate request
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Wrap request in MiddleOneRequestProto
	wrappedReq := &configpb.MiddleOneRequestProto{
		Request: req,
	}

	// Pass proto message directly to repository
	account, err := s.accountRepo.SendMiddleOneRequestFromAccountApi(ctx, wrappedReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	log.Printf("Created account: %s", req.GetName())
	return account, nil
}

// DeleteAccount deletes an account
func (s *ConfigurationApi) DeleteAccount(
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
	response, err := s.accountRepo.SendAccountDeletionRequestFromAccountApi(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}

	log.Printf("Deleted account: %s", accountKey)
	return response, nil
}

// ListAccounts lists all accounts
func (s *ConfigurationApi) ListAccounts(
	ctx context.Context,
	req *configpb.ListAccountsRequestProto,
) (*configpb.ListAccountsResponseProto, error) {
	// Pass proto message directly to repository
	response, err := s.accountRepo.SendListAccountsRequestFromAccountApi(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %v", err)
	}

	return response, nil
}

// RegisterGRPC implements server_builder.GRPCServiceRegistrar
func (s *ConfigurationApi) RegisterGRPC(Api grpc.ServiceRegistrar) {
	gw.RegisterConfigurationServer(Api, s)
}

// RegisterGateway implements server_builder.HTTPGatewayRegistrar
func (s *ConfigurationApi) RegisterGateway(ctx context.Context, mux *runtime.ServeMux) error {
	return gw.RegisterConfigurationHandlerServer(ctx, mux, s)
}
