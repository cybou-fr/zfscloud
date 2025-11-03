package grpc

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/shared/crypto"
)

// Define a custom type for the context key to avoid collisions
type contextKey string

const clientIDKey contextKey = "client_id"

// GetClientIDFromContext retrieves the client ID from the context
func GetClientIDFromContext(ctx context.Context) (string, error) {
	clientID, ok := ctx.Value(clientIDKey).(string)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing client-id in context")
	}
	return clientID, nil
}

// AuthInterceptor is a server interceptor for authentication
func AuthInterceptor(database *db.Database) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// List of public methods that don't require authentication
		publicMethods := []string{
			"/zfscloud.v1.BackupService/GetServerPublicKey",
			"/zfscloud.v1.AdminService/ListRegistrations",
			"/zfscloud.v1.AdminService/ApproveRegistration",
			"/zfscloud.v1.AdminService/RejectRegistration",
		}

		for _, publicMethod := range publicMethods {
			if info.FullMethod == publicMethod {
				return handler(ctx, req)
			}
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		clientIDs := md.Get("client-id")
		if len(clientIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing client-id")
		}
		clientID := clientIDs[0]

		// Add clientID to context for use in RPC handlers
		newCtx := context.WithValue(ctx, clientIDKey, clientID)

		return handler(newCtx, req)
	}
}

// StreamAuthInterceptor is a server interceptor for stream authentication
func StreamAuthInterceptor(database *db.Database) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		clientIDs := md.Get("client-id")
		if len(clientIDs) == 0 {
			return status.Errorf(codes.Unauthenticated, "missing client-id")
		}
		clientID := clientIDs[0]

		// Add clientID to context for use in RPC handlers
		newCtx := context.WithValue(stream.Context(), clientIDKey, clientID)

		wrappedStream := newWrappedStream(stream, newCtx)
		return handler(srv, wrappedStream)
	}
}

func signatureAuth(ctx context.Context, database *db.Database) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Extract client ID, signature, and timestamp from metadata
	clientIDs := md.Get("client-id")
	signatures := md.Get("signature")
	timestamps := md.Get("timestamp")

	if len(clientIDs) == 0 || len(signatures) == 0 || len(timestamps) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authentication headers")
	}

	clientID := clientIDs[0]
	signatureB64 := signatures[0]
	timestamp := timestamps[0]

	// Decode signature from Base64
	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "invalid signature format")
	}

	// Get user from database
	user, err := database.GetUserByClientID(clientID)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "invalid client ID")
	}

	// Construct message to verify
	message := fmt.Sprintf("%s:%s", clientID, timestamp)

	// Verify signature
	if !crypto.VerifySignature(user.PublicKey, []byte(message), signature) {
		return "", status.Errorf(codes.Unauthenticated, "invalid signature")
	}

	return clientID, nil
}

// wrappedStream wraps the grpc.ServerStream to override the Context()
// See: https://github.com/grpc/grpc-go/issues/119
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func newWrappedStream(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &wrappedStream{stream, ctx}
}
