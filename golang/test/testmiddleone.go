package test

import (
	"context"
	"log"

	geninterfaces "github.com/berendjan/golang-bazel-starter/golang/generated/interfaces"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
)

type TestMiddleOne struct{}

// Compile-time check that TestMiddleOne implements MiddlewareOneInterface
var _ geninterfaces.MiddlewareOneInterface = (*TestMiddleOne)(nil)

// NewTestMiddleOne creates a new TestMiddleOne middleware
func NewTestMiddleOne() *TestMiddleOne {
	return &TestMiddleOne{}
}

// HandleTestMiddleOneRequest logs the message and forwards to the repository
func (m *TestMiddleOne) HandleMiddleOneRequest(ctx context.Context, req *configpb.MiddleOneRequestProto, next geninterfaces.MiddlewareOneSendable) (*configpb.AccountConfigurationProto, error) {
	log.Printf("TestMiddleOne: Processing account creation request: %+v", req)

	// Forward to next handler
	result, err := next.SendMiddleOneRequestFromMiddlewareOne(ctx, req)

	if err != nil {
		log.Printf("TestMiddleOne: Account creation failed: %v", err)
		return nil, err
	}

	log.Printf("TestMiddleOne: Account creation successful: %+v", result)
	return result, nil
}
