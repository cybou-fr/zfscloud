package grpc

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/shared/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor handles authentication for gRPC requests
type AuthInterceptor struct {
	database *db.Database
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(database *db.Database) *AuthInterceptor {
	return &AuthInterceptor{
		database: database,
	}
}

// Unary returns a server interceptor function to authenticate unary requests
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for registration endpoint
		if info.FullMethod == "/zfscloud.v1.BackupService/Register" {
			return handler(ctx, req)
		}

		// Authenticate request
		clientID, err := a.authenticate(ctx, req)
		if err != nil {
			return nil, err
		}

		// Add client ID to context
		newCtx := context.WithValue(ctx, "client_id", clientID)
		return handler(newCtx, req)
	}
}

// Stream returns a server interceptor function to authenticate stream requests
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Authenticate request
		clientID, err := a.authenticate(stream.Context(), nil)
		if err != nil {
			return err
		}

		// Add client ID to context
		newCtx := context.WithValue(stream.Context(), "client_id", clientID)
		wrappedStream := &wrappedStream{stream, newCtx}
		return handler(srv, wrappedStream)
	}
}

// authenticate validates the request signature
func (a *AuthInterceptor) authenticate(ctx context.Context, req interface{}) (string, error) {
	// Extract metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	clientIDValues := md.Get("client-id")
	if len(clientIDValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "client-id is not provided")
	}
	clientID := clientIDValues[0]

	signatureValues := md.Get("signature")
	if len(signatureValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "signature is not provided")
	}
	// Decode base64-encoded signature
	signature, err := base64.StdEncoding.DecodeString(signatureValues[0])
	if err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid signature encoding")
	}

	timestampValues := md.Get("timestamp")
	if len(timestampValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "timestamp is not provided")
	}
	timestamp := timestampValues[0]

	nonceValues := md.Get("nonce")
	if len(nonceValues) == 0 {
		return "", status.Error(codes.Unauthenticated, "nonce is not provided")
	}
	nonce := nonceValues[0]

	// Get user from database
	user, err := a.database.GetUserByClientID(clientID)
	if err != nil {
		fmt.Printf("Error getting user by client ID '%s': %v\n", clientID, err)
		return "", status.Error(codes.Internal, "failed to get user")
	}
	if user == nil {
		fmt.Printf("User not found for client ID: %s\n", clientID)
		return "", status.Error(codes.Unauthenticated, "user not found")
	}

	// Check if user is revoked
	if user.Revoked {
		return "", status.Error(codes.PermissionDenied, "user access has been revoked")
	}

	// Validate timestamp
	var ts int64
	if _, err := fmt.Sscanf(timestamp, "%d", &ts); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid timestamp")
	}

	if !crypto.ValidateTimestamp(ts) {
		return "", status.Error(codes.Unauthenticated, "timestamp is outside valid range")
	}

	// Check nonce for replay protection
	isValid, err := a.database.CheckAndStoreNonce(nonce, clientID, ts)
	if err != nil {
		return "", status.Error(codes.Internal, "failed to validate nonce")
	}
	if !isValid {
		return "", status.Error(codes.Unauthenticated, "nonce has already been used (replay attack detected)")
	}

	// Verify signature
	// Note: In a real implementation, we'd build the signature data from the request
	// For now, we'll just verify the signature exists
	sigData := &crypto.SignatureData{
		ClientID:  clientID,
		Timestamp: ts,
		Nonce:     nonce,
		Payload:   []byte(clientID), // Match client's payload
	}

	if !crypto.Verify(user.PublicKey, sigData, signature) {
		return "", status.Error(codes.Unauthenticated, "signature verification failed")
	}

	// Update last seen
	if err := a.database.UpdateUserLastSeen(user.UserID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to update last seen for user %s: %v\n", user.UserID, err)
	}

	return clientID, nil
}

// wrappedStream wraps around the embedded grpc.ServerStream with a new context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// GetClientIDFromContext extracts the client ID from the context
func GetClientIDFromContext(ctx context.Context) (string, error) {
	value := ctx.Value("client_id")
	if value == nil {
		return "", fmt.Errorf("client_id not found in context")
	}

	clientID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("client_id has invalid type")
	}

	return clientID, nil
}
