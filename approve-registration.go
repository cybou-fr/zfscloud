package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ioustamora/zfscloud/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run approve-registration.go <registration-id> [quota-gb]")
		fmt.Println("Example: go run approve-registration.go f416d393-9456-4fd8-956b-2d26a36e3d74 50")
		os.Exit(1)
	}

	registrationID := os.Args[1]
	quotaGB := int64(50)

	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &quotaGB)
	}

	quotaBytes := quotaGB * 1024 * 1024 * 1024

	fmt.Printf("Approving registration %s with %d GB quota...\n", registrationID, quotaGB)

	// Connect to server
	conn, err := grpc.NewClient("127.0.0.1:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	adminService := proto.NewAdminServiceClient(conn)

	// Create context with admin token
	md := metadata.Pairs("admin-token", "test-admin-token-12345")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Approve registration
	req := &proto.ApproveRegistrationRequest{
		RegistrationId: registrationID,
		Approved:       true,
		QuotaBytes:     quotaBytes,
		AdminToken:     "test-admin-token-12345",
	}

	resp, err := adminService.ApproveRegistration(ctx, req)
	if err != nil {
		log.Fatalf("Failed to approve: %v", err)
	}

	if !resp.Success {
		log.Fatalf("Approval rejected: %s", resp.Message)
	}

	fmt.Println("✓ Registration approved successfully!")
	fmt.Println(resp.Message)

	if resp.UserId != "" {
		fmt.Printf("User ID: %s\n", resp.UserId)
		fmt.Printf("Dataset: %s\n", resp.Dataset)
	}
}
