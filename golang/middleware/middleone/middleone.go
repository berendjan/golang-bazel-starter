package middleone

import (
	"context"
	"log"

	geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

type MiddleOne struct{}

// Compile-time check that MiddleOne implements MiddlewareOneInterface
var _ geninterfaces.MiddlewareOneInterface = (*MiddleOne)(nil)

// NewMiddleOne creates a new MiddleOne middleware
func NewMiddleOne() *MiddleOne {
	return &MiddleOne{}
}

// HandleMiddleOneRequest logs the message and forwards to the repository
func (m *MiddleOne) HandleMiddleOneRequest(ctx context.Context, req *configpb.MiddleOneRequestProto, next geninterfaces.MiddlewareOneSendable) (*configpb.AccountConfigurationProto, error) {
	log.Printf("MiddleOne: Processing account creation request: %+v", req)

	// Forward to next handler
	result, err := next.SendMiddleOneRequestFromMiddlewareOne(ctx, req)

	if err != nil {
		log.Printf("MiddleOne: Account creation failed: %v", err)
		return nil, err
	}

	log.Printf("MiddleOne: Account creation successful: %+v", result)
	return result, nil
}
