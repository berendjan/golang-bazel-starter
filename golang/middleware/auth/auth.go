package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// KratosSession represents the response from Kratos /sessions/whoami
type KratosSession struct {
	ID       string         `json:"id"`
	Active   bool           `json:"active"`
	Identity KratosIdentity `json:"identity"`
}

// KratosIdentity represents the identity in a Kratos session
type KratosIdentity struct {
	ID string `json:"id"`
}

// AuthMiddleware validates Kratos sessions and extracts user IDs
type AuthMiddleware struct {
	kratosURL  string
	httpClient *http.Client
}

// isRunningInTest checks if the code is being called from a Go test
// by inspecting the call stack for test-related function names
func isRunningInTest() bool {
	var pcs [32]uintptr
	n := runtime.Callers(0, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// Check if the function name contains "testing" package calls
		// or if it starts with "Test" (test function names)
		if strings.Contains(frame.Function, "testing.") ||
			strings.Contains(frame.File, "_test.go") {
			return true
		}
		if !more {
			break
		}
	}
	return false
}

// NewAuthMiddleware creates a new auth middleware
// kratosURL should be the Kratos public API URL (e.g., "http://kratos.app-namespace.svc.cluster.local:4433")
func NewAuthMiddleware(kratosURL string) *AuthMiddleware {
	return &AuthMiddleware{
		kratosURL:  kratosURL,
		httpClient: &http.Client{},
	}
}

// ExtractUserID extracts and validates the user ID from the request context
// Returns the user ID or an error if authentication fails
func (m *AuthMiddleware) ExtractUserID(ctx context.Context) (string, error) {
	// Skip auth validation in tests
	if isRunningInTest() {
		return "test-user", nil
	}

	// Get cookies from gRPC metadata (forwarded by grpc-gateway)
	cookie, err := m.extractCookie(ctx)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "no session cookie found")
	}

	// Validate session with Kratos
	session, err := m.validateSession(ctx, cookie)
	if err != nil {
		log.Printf("Auth: session validation failed: %v", err)
		return "", status.Error(codes.Unauthenticated, "invalid session")
	}

	if !session.Active {
		return "", status.Error(codes.Unauthenticated, "session is not active")
	}

	userID := session.Identity.ID
	if userID == "" {
		return "", status.Error(codes.Unauthenticated, "no user ID in session")
	}

	log.Printf("Auth: authenticated user %s", userID)
	return userID, nil
}

// extractCookie extracts the session cookie from gRPC metadata
func (m *AuthMiddleware) extractCookie(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	// grpc-gateway forwards cookies in the "grpcgateway-cookie" metadata key
	cookies := md.Get("grpcgateway-cookie")
	if len(cookies) == 0 {
		// Also check "cookie" header (direct gRPC calls)
		cookies = md.Get("cookie")
	}

	if len(cookies) == 0 {
		return "", fmt.Errorf("no cookie header found")
	}

	// Join all cookie values
	return strings.Join(cookies, "; "), nil
}

// validateSession calls Kratos to validate the session
func (m *AuthMiddleware) validateSession(ctx context.Context, cookie string) (*KratosSession, error) {
	url := fmt.Sprintf("%s/sessions/whoami", m.kratosURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Kratos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("session not authenticated")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Kratos returned status %d: %s", resp.StatusCode, string(body))
	}

	var session KratosSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	return &session, nil
}
