package middletwo

import (
	"context"
	"log"

	geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"
	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

type MiddleTwo struct{}

// Compile-time check that MiddleTwo implements MiddlewareTwoInterface
var _ geninterfaces.MiddlewareTwoInterface = (*MiddleTwo)(nil)

// NewMiddleTwo creates a new MiddleTwo middleware
func NewMiddleTwo() *MiddleTwo {
	return &MiddleTwo{}
}

// HandleAccountDeletionRequest logs the message and forwards to the repository
func (m *MiddleTwo) HandleAccountDeletionRequest(ctx context.Context, req *configpb.AccountDeletionRequestProto, next geninterfaces.MiddlewareTwoSendable) (*commonpb.StatusResponseProto, error) {
	log.Printf("MiddleTwo: Processing account deletion request: %+v", req)

	// Forward to next handler
	result, err := next.SendAccountDeletionRequestFromMiddlewareTwo(ctx, req)

	if err != nil {
		log.Printf("MiddleTwo: Account deletion failed: %v", err)
		return nil, err
	}

	log.Printf("MiddleTwo: Account deletion successful: %+v", result)
	return result, nil
}

// HandleListAccountsRequest logs the message and forwards to the repository
func (m *MiddleTwo) HandleListAccountsRequest(ctx context.Context, req *configpb.ListAccountsRequestProto, next geninterfaces.MiddlewareTwoSendable) (*configpb.ListAccountsResponseProto, error) {
	log.Printf("MiddleTwo: Processing list accounts request: %+v", req)

	// Forward to next handler
	result, err := next.SendListAccountsRequestFromMiddlewareTwo(ctx, req)

	if err != nil {
		log.Printf("MiddleTwo: List accounts failed: %v", err)
		return nil, err
	}

	log.Printf("MiddleTwo: List accounts successful: %d accounts found", len(result.GetAccounts()))
	return result, nil
}
