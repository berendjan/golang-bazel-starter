package middleone

import (
	"context"
	"log"

	"github.com/berendjan/golang-bazel-starter/golang/middleware/auth"

	geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

type MiddleOne struct {
	auth *auth.AuthMiddleware
}

// Compile-time check that MiddleOne implements MiddlewareOneInterface
var _ geninterfaces.MiddlewareOneInterface = (*MiddleOne)(nil)

// NewMiddleOne creates a new MiddleOne middleware
func NewMiddleOne(authMiddleware *auth.AuthMiddleware) *MiddleOne {
	return &MiddleOne{
		auth: authMiddleware,
	}
}

// HandleMiddleOneRequest authenticates the user and forwards to the next handler
func (m *MiddleOne) HandleMiddleOneRequest(ctx context.Context, req *configpb.MiddleOneRequestProto, next geninterfaces.MiddlewareOneSendable) (*configpb.AccountConfigurationProto, error) {
	// Extract and validate user ID from cookie
	userID, err := m.auth.ExtractUserID(ctx)
	if err != nil {
		log.Printf("MiddleOne: Authentication failed: %v", err)
		return nil, err
	}

	// Add user ID to context for downstream handlers
	ctx = auth.WithUserID(ctx, userID)

	log.Printf("MiddleOne: Processing request for user %s: %+v", userID, req)

	// Forward to next handler with authenticated context
	result, err := next.SendMiddleOneRequestFromMiddlewareOne(ctx, req)
	if err != nil {
		log.Printf("MiddleOne: Request failed for user %s: %v", userID, err)
		return nil, err
	}

	log.Printf("MiddleOne: Request successful for user %s: %+v", userID, result)
	return result, nil
}
