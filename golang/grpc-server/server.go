package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "github.com/berendjan/golang-bazel-starter/proto/common/v1"
	configpb "github.com/berendjan/golang-bazel-starter/proto/configuration/v1"
	configservicepb "github.com/berendjan/golang-bazel-starter/proto/configuration_service/v1"
)

// ConfigurationServer implements the Configuration gRPC service
type ConfigurationServer struct {
	configservicepb.UnimplementedConfigurationServer

	// In-memory storage (replace with actual database in production)
	mu                  sync.RWMutex
	accounts            map[string]*configpb.AccountConfigurationProto
	groups              map[string]*configpb.GroupConfigurationProto
	configurationEvents map[string][]*configpb.ConfigurationEventProto
}

// NewConfigurationServer creates a new Configuration service server
func NewConfigurationServer() *ConfigurationServer {
	return &ConfigurationServer{
		accounts:            make(map[string]*configpb.AccountConfigurationProto),
		groups:              make(map[string]*configpb.GroupConfigurationProto),
		configurationEvents: make(map[string][]*configpb.ConfigurationEventProto),
	}
}

// CreateAccount creates a new account
func (s *ConfigurationServer) CreateAccount(
	ctx context.Context,
	req *configpb.AccountCreationRequestProto,
) (*configpb.AccountConfigurationProto, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a simple account ID (in production, use proper UUID)
	accountID := &commonpb.ConfigurationIdProto{
		Id:   []byte(fmt.Sprintf("account-%s", req.GetName())),
		Type: 1, // Account type
	}

	account := &configpb.AccountConfigurationProto{
		AccountId: accountID,
	}

	s.accounts[string(accountID.GetId())] = account

	log.Printf("Created account: %s", req.GetName())
	return account, nil
}

// DeleteAccount deletes an account
func (s *ConfigurationServer) DeleteAccount(
	ctx context.Context,
	req *configpb.AccountDeletionRequestProto,
) (*commonpb.StatusResponseProto, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In production, validate the account exists and perform cleanup
	log.Printf("Deleted account")

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Account deleted successfully",
	}, nil
}

// RequestToJoinGroup handles a request to join a group
func (s *ConfigurationServer) RequestToJoinGroup(
	ctx context.Context,
	req *configpb.RequestToJoinGroupProto,
) (*commonpb.StatusResponseProto, error) {
	if req.GetAccountId() == nil || req.GetGroupId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id and group_id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a pending member event
	event := &configpb.ConfigurationEventProto{
		Event: &configpb.ConfigurationEventProto_PendingMemberEvent{
			PendingMemberEvent: &configpb.PendingMemberEventProto{
				AccountId:       req.GetAccountId(),
				GroupId:         req.GetGroupId(),
				InviterId:       req.GetInviterId(),
				InviteId:        req.GetInviteId(),
				X25519PublicKey: req.GetX25519PublicKey(),
			},
		},
	}

	// Store the event (keyed by group ID for simplicity)
	groupKey := string(req.GetGroupId().GetId())
	s.configurationEvents[groupKey] = append(s.configurationEvents[groupKey], event)

	log.Printf("Request to join group received for account: %x, group: %x",
		req.GetAccountId().GetId(), req.GetGroupId().GetId())

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Request to join group submitted successfully",
	}, nil
}

// ListConfigurationEvents lists all configuration events for a group
func (s *ConfigurationServer) ListConfigurationEvents(
	ctx context.Context,
	req *configpb.ListConfigurationEventsRequestProto,
) (*configpb.ListConfigurationEventsResponseProto, error) {
	if req.GetAccountId() == nil || req.GetGroupId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id and group_id are required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	groupKey := string(req.GetGroupId().GetId())
	events := s.configurationEvents[groupKey]

	log.Printf("Listing %d configuration events for group: %x", len(events), req.GetGroupId().GetId())

	return &configpb.ListConfigurationEventsResponseProto{
		ConfigurationEvents: events,
	}, nil
}

// AcceptRequestToJoinGroup accepts a request to join a group
func (s *ConfigurationServer) AcceptRequestToJoinGroup(
	ctx context.Context,
	req *configpb.AcceptRequestToJoinGroupProto,
) (*commonpb.StatusResponseProto, error) {
	if req.GetAccountId() == nil || req.GetGroupId() == nil || req.GetInviteeId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id, group_id, and invitee_id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a pending member accepted event
	event := &configpb.ConfigurationEventProto{
		Event: &configpb.ConfigurationEventProto_PendingMemberAcceptedEvent{
			PendingMemberAcceptedEvent: &configpb.PendingMemberAcceptedEventProto{
				AccountId:         req.GetInviteeId(),
				GroupId:           req.GetGroupId(),
				EncryptedGroupKey: req.GetEncryptedGroupKey(),
			},
		},
	}

	// Store the event
	groupKey := string(req.GetGroupId().GetId())
	s.configurationEvents[groupKey] = append(s.configurationEvents[groupKey], event)

	// Add member to group
	if group, exists := s.groups[groupKey]; exists {
		group.Members = append(group.Members, &configpb.GroupConfigurationProto_MemberProto{
			AccountId: req.GetInviteeId(),
		})
	}

	log.Printf("Accepted request to join group for invitee: %x, group: %x",
		req.GetInviteeId().GetId(), req.GetGroupId().GetId())

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Request to join group accepted successfully",
	}, nil
}

// DenyRequestToJoinGroup denies a request to join a group
func (s *ConfigurationServer) DenyRequestToJoinGroup(
	ctx context.Context,
	req *configpb.DenyRequestToJoinGroupProto,
) (*commonpb.StatusResponseProto, error) {
	if req.GetAccountId() == nil || req.GetGroupId() == nil || req.GetInviteeId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id, group_id, and invitee_id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// In production, remove the pending member event
	log.Printf("Denied request to join group for invitee: %x, group: %x",
		req.GetInviteeId().GetId(), req.GetGroupId().GetId())

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Request to join group denied successfully",
	}, nil
}

// ListGroups lists all groups for an account
func (s *ConfigurationServer) ListGroups(
	ctx context.Context,
	req *configpb.ListGroupsRequestProto,
) (*configpb.ListGroupsResponseProto, error) {
	if req.GetAccountId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// In production, filter groups by account membership
	var groups []*configpb.GroupConfigurationProto
	for _, group := range s.groups {
		// Check if account is a member
		for _, member := range group.GetMembers() {
			if string(member.GetAccountId().GetId()) == string(req.GetAccountId().GetId()) {
				groups = append(groups, group)
				break
			}
		}
	}

	log.Printf("Listing %d groups for account: %x", len(groups), req.GetAccountId().GetId())

	return &configpb.ListGroupsResponseProto{
		Groups: groups,
	}, nil
}

// DeleteMember removes a member from a group
func (s *ConfigurationServer) DeleteMember(
	ctx context.Context,
	req *configpb.MemberDeletionRequestProto,
) (*commonpb.StatusResponseProto, error) {
	if req.GetAccountId() == nil || req.GetGroupId() == nil {
		return nil, status.Error(codes.InvalidArgument, "account_id and group_id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	groupKey := string(req.GetGroupId().GetId())
	if group, exists := s.groups[groupKey]; exists {
		// Remove the member from the group
		var updatedMembers []*configpb.GroupConfigurationProto_MemberProto
		for _, member := range group.GetMembers() {
			if string(member.GetAccountId().GetId()) != string(req.GetAccountId().GetId()) {
				updatedMembers = append(updatedMembers, member)
			}
		}
		group.Members = updatedMembers
	}

	log.Printf("Deleted member: %x from group: %x", req.GetAccountId().GetId(), req.GetGroupId().GetId())

	return &commonpb.StatusResponseProto{
		Code:    200,
		Message: "Member deleted successfully",
	}, nil
}
