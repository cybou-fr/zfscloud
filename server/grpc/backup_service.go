package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/ioustamora/zfscloud/proto"
	"github.com/ioustamora/zfscloud/server/db"
	"github.com/ioustamora/zfscloud/server/storage"
	"github.com/ioustamora/zfscloud/server/zfs"
	"github.com/ioustamora/zfscloud/shared/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BackupServiceServer implements the BackupService gRPC service
type BackupServiceServer struct {
	proto.UnimplementedBackupServiceServer
	database   *db.Database
	zfsManager *zfs.Manager
	config     *config.ServerConfig
}

// NewBackupServiceServer creates a new backup service server
func NewBackupServiceServer(database *db.Database, zfsManager *zfs.Manager, cfg *config.ServerConfig) *BackupServiceServer {
	return &BackupServiceServer{
		database:   database,
		zfsManager: zfsManager,
		config:     cfg,
	}
}

// Register handles client registration
func (s *BackupServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Check if client is already registered
	existingReg, err := s.database.GetRegistrationByClientID(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check existing registration")
	}

	if existingReg != nil {
		return &proto.RegisterResponse{
			Success:        false,
			Message:        "Client already registered",
			RegistrationId: existingReg.RegistrationID,
			Status:         getRegistrationStatus(existingReg.Status),
		}, nil
	}

	// Check if user already exists
	existingUser, err := s.database.GetUserByClientID(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check existing user")
	}

	if existingUser != nil {
		return &proto.RegisterResponse{
			Success: false,
			Message: "User already exists",
			Status:  proto.RegistrationStatus_REGISTRATION_STATUS_APPROVED,
		}, nil
	}

	// Create registration
	registration := &db.Registration{
		RegistrationID:      uuid.New().String(),
		ClientID:            req.ClientId,
		PublicKey:           req.PublicKey,
		EncryptedPassphrase: req.EncryptedPassphrase,
		Status:              "pending",
		Hostname:            sql.NullString{String: req.Metadata.Hostname, Valid: req.Metadata.Hostname != ""},
		Username:            sql.NullString{String: req.Metadata.Username, Valid: req.Metadata.Username != ""},
		OS:                  sql.NullString{String: req.Metadata.Os, Valid: req.Metadata.Os != ""},
		ClientVersion:       sql.NullString{String: req.Metadata.ClientVersion, Valid: req.Metadata.ClientVersion != ""},
	}

	if err := s.database.CreateRegistration(registration); err != nil {
		return nil, status.Error(codes.Internal, "failed to create registration")
	}

	// Log the action
	s.database.LogAction("", req.ClientId, "registration_request", fmt.Sprintf("Client %s requested registration", req.ClientId), "")

	return &proto.RegisterResponse{
		Success:        true,
		Message:        "Registration request submitted. Waiting for admin approval.",
		RegistrationId: registration.RegistrationID,
		Status:         proto.RegistrationStatus_REGISTRATION_STATUS_PENDING,
	}, nil
}

// GetRegistrationStatus returns the status of a registration
func (s *BackupServiceServer) GetRegistrationStatus(ctx context.Context, req *proto.RegistrationStatusRequest) (*proto.RegistrationStatusResponse, error) {
	registration, err := s.database.GetRegistrationByClientID(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get registration")
	}

	if registration == nil {
		return &proto.RegistrationStatusResponse{
			Status:  proto.RegistrationStatus_REGISTRATION_STATUS_UNKNOWN,
			Message: "Registration not found",
		}, nil
	}

	resp := &proto.RegistrationStatusResponse{
		Status:  getRegistrationStatus(registration.Status),
		Message: fmt.Sprintf("Registration status: %s", registration.Status),
	}

	if registration.ProcessedAt != nil {
		resp.ApprovedAt = timestamppb.New(*registration.ProcessedAt)
	}

	return resp, nil
}

// GetBackupPolicy returns the backup policy for a client
func (s *BackupServiceServer) GetBackupPolicy(ctx context.Context, req *proto.BackupPolicyRequest) (*proto.BackupPolicyResponse, error) {
	return &proto.BackupPolicyResponse{
		CheckIntervalSeconds: 300,              // 5 minutes
		MaxChunkSizeBytes:    1024 * 1024 * 16, // 16 MB
		CompressionEnabled:   true,
		CompressionAlgorithm: "zstd",
	}, nil
}

// GetHashIndex returns the hash index for a client
func (s *BackupServiceServer) GetHashIndex(ctx context.Context, req *proto.HashIndexRequest) (*proto.HashIndexResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// TODO: Implement actual hash index retrieval from ZFS dataset
	// For now, return empty list
	return &proto.HashIndexResponse{
		Hashes: []*proto.FileHash{},
	}, nil
}

// UploadChunk handles file chunk uploads
func (s *BackupServiceServer) UploadChunk(stream proto.BackupService_UploadChunkServer) error {
	clientID, err := GetClientIDFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return status.Error(codes.NotFound, "user not found")
	}

	// Get dataset mount point
	datasetPath, err := s.zfsManager.GetMountpoint(user.ZFSDataset)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get dataset mountpoint: %v", err))
	}

	writer := storage.NewChunkWriter(datasetPath)
	defer writer.Close()

	var totalBytesReceived int64
	var currentMetadata *proto.ChunkMetadata

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client finished sending
			if currentMetadata != nil {
				if err := writer.FinishFile(); err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("failed to finalize file: %v", err))
				}
				totalBytesReceived += writer.BytesWritten()
			}
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "failed to receive chunk")
		}

		switch data := req.Data.(type) {
		case *proto.UploadChunkRequest_Metadata:
			// Finish previous file if exists
			if currentMetadata != nil {
				if err := writer.FinishFile(); err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("failed to finalize file: %v", err))
				}
				totalBytesReceived += writer.BytesWritten()
			}

			// Start new file
			currentMetadata = data.Metadata
			if err := writer.StartFile(data.Metadata); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("failed to start file: %v", err))
			}

		case *proto.UploadChunkRequest_ChunkData:
			if currentMetadata == nil {
				return status.Error(codes.InvalidArgument, "metadata must be sent before chunk data")
			}
			if err := writer.WriteChunk(data.ChunkData); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("failed to write chunk: %v", err))
			}
		}
	}

	// Calculate actual used space from filesystem
	usedSpace, err := storage.GetDatasetUsedSpace(datasetPath)
	if err != nil {
		// Fall back to estimated space if calculation fails
		usedSpace = user.UsedBytes + totalBytesReceived
	}

	// Update user's used space
	if err := s.database.UpdateUserUsedSpace(user.UserID, usedSpace); err != nil {
		return status.Error(codes.Internal, "failed to update used space")
	}

	return stream.SendAndClose(&proto.UploadChunkResponse{
		Success:       true,
		Message:       "Upload completed successfully",
		BytesReceived: totalBytesReceived,
	})
}

// CompleteBackup marks a backup as complete
func (s *BackupServiceServer) CompleteBackup(ctx context.Context, req *proto.CompleteBackupRequest) (*proto.CompleteBackupResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	// Log the action
	s.database.LogAction("", clientID, "backup_complete", fmt.Sprintf("Backup completed: %d files", len(req.FilePaths)), "")

	return &proto.CompleteBackupResponse{
		Success:         true,
		Message:         "Backup completed successfully",
		BackupTimestamp: timestamppb.Now(),
	}, nil
}

// ListFiles lists files in the dataset
func (s *BackupServiceServer) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Get dataset mount point
	datasetPath, err := s.zfsManager.GetMountpoint(user.ZFSDataset)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get dataset mountpoint: %v", err))
	}

	// List files
	reader := storage.NewFileReader(datasetPath)
	files, err := reader.ListFiles(req.Path)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list files: %v", err))
	}

	return &proto.ListFilesResponse{
		Files: files,
	}, nil
}

// DownloadFile streams a file to the client
func (s *BackupServiceServer) DownloadFile(req *proto.DownloadFileRequest, stream proto.BackupService_DownloadFileServer) error {
	clientID, err := GetClientIDFromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return status.Error(codes.NotFound, "user not found")
	}

	// Get dataset mount point
	datasetPath, err := s.zfsManager.GetMountpoint(user.ZFSDataset)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get dataset mountpoint: %v", err))
	}

	// Get file info
	reader := storage.NewFileReader(datasetPath)
	fileInfo, err := reader.GetFileInfo(req.FilePath)
	if err != nil {
		return status.Error(codes.NotFound, fmt.Sprintf("file not found: %v", err))
	}

	// Send metadata first
	if err := stream.Send(&proto.DownloadFileResponse{
		Data: &proto.DownloadFileResponse_Metadata{
			Metadata: &proto.FileMetadata{
				Path:         fileInfo.Path,
				Size:         fileInfo.Size,
				ModifiedTime: timestamppb.Now(), // TODO: Get actual modified time
				Hash:         fileInfo.Hash,
			},
		},
	}); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to send metadata: %v", err))
	}

	// Open and stream file
	file, err := reader.OpenFile(req.FilePath)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to open file: %v", err))
	}
	defer file.Close()

	// Stream file in chunks
	buffer := make([]byte, 64*1024) // 64KB chunks
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to read file: %v", err))
		}

		if err := stream.Send(&proto.DownloadFileResponse{
			Data: &proto.DownloadFileResponse_ChunkData{
				ChunkData: buffer[:n],
			},
		}); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to send chunk: %v", err))
		}
	}

	return nil
}

// ListSnapshots lists available snapshots
func (s *BackupServiceServer) ListSnapshots(ctx context.Context, req *proto.ListSnapshotsRequest) (*proto.ListSnapshotsResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	snapshots, err := s.database.ListSnapshotsByUserID(user.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list snapshots")
	}

	protoSnapshots := make([]*proto.SnapshotInfo, len(snapshots))
	for i, snap := range snapshots {
		protoSnapshots[i] = &proto.SnapshotInfo{
			SnapshotId:      snap.SnapshotID,
			SnapshotName:    snap.ZFSSnapshotName,
			CreatedAt:       timestamppb.New(snap.CreatedAt),
			SizeBytes:       snap.SizeBytes,
			ZfsSnapshotName: snap.ZFSSnapshotName,
		}
	}

	return &proto.ListSnapshotsResponse{
		Snapshots: protoSnapshots,
	}, nil
}

// RequestSnapshotRestore requests a snapshot restore
func (s *BackupServiceServer) RequestSnapshotRestore(ctx context.Context, req *proto.SnapshotRestoreRequest) (*proto.SnapshotRestoreResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Create restore request
	restoreReq := &db.RestoreRequest{
		RequestID:   uuid.New().String(),
		UserID:      user.UserID,
		ClientID:    clientID,
		SnapshotID:  req.SnapshotId,
		RestorePath: req.RestorePath,
		Reason:      req.Reason,
		Status:      "pending",
	}

	if err := s.database.CreateRestoreRequest(restoreReq); err != nil {
		return nil, status.Error(codes.Internal, "failed to create restore request")
	}

	// Log the action
	s.database.LogAction(user.UserID, clientID, "restore_request", fmt.Sprintf("Requested restore of snapshot %s", req.SnapshotId), "")

	return &proto.SnapshotRestoreResponse{
		Success:   true,
		Message:   "Restore request submitted. Waiting for admin approval.",
		RequestId: restoreReq.RequestID,
		Status:    proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_PENDING,
	}, nil
}

// GetRestoreRequestStatus gets the status of a restore request
func (s *BackupServiceServer) GetRestoreRequestStatus(ctx context.Context, req *proto.RestoreRequestStatusRequest) (*proto.RestoreRequestStatusResponse, error) {
	restoreReq, err := s.database.GetRestoreRequestByID(req.RequestId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get restore request")
	}

	if restoreReq == nil {
		return nil, status.Error(codes.NotFound, "restore request not found")
	}

	resp := &proto.RestoreRequestStatusResponse{
		Status:  getRestoreRequestStatus(restoreReq.Status),
		Message: fmt.Sprintf("Restore request status: %s", restoreReq.Status),
	}

	if restoreReq.ProcessedAt != nil {
		resp.UpdatedAt = timestamppb.New(*restoreReq.ProcessedAt)
	}

	return resp, nil
}

// LoadDatasetKey loads the encryption key for a dataset
func (s *BackupServiceServer) LoadDatasetKey(ctx context.Context, req *proto.LoadDatasetKeyRequest) (*proto.LoadDatasetKeyResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// TODO: Decrypt passphrase and load into ZFS
	// For now, just acknowledge
	return &proto.LoadDatasetKeyResponse{
		Success: true,
		Message: "Dataset key loaded successfully",
	}, nil
}

// UnloadDatasetKey unloads the encryption key for a dataset
func (s *BackupServiceServer) UnloadDatasetKey(ctx context.Context, req *proto.UnloadDatasetKeyRequest) (*proto.UnloadDatasetKeyResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Unload key from ZFS
	if err := s.zfsManager.UnloadKey(user.ZFSDataset); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unload key: %v", err))
	}

	return &proto.UnloadDatasetKeyResponse{
		Success: true,
		Message: "Dataset key unloaded successfully",
	}, nil
}

// Heartbeat handles client heartbeat
func (s *BackupServiceServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	clientID, err := GetClientIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "client not authenticated")
	}

	// Update last seen
	user, err := s.database.GetUserByClientID(clientID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := s.database.UpdateUserLastSeen(user.UserID); err != nil {
		return nil, status.Error(codes.Internal, "failed to update last seen")
	}

	return &proto.HeartbeatResponse{
		Success:         true,
		ServerTimestamp: time.Now().Unix(),
	}, nil
}

// Helper functions

func getRegistrationStatus(status string) proto.RegistrationStatus {
	switch status {
	case "pending":
		return proto.RegistrationStatus_REGISTRATION_STATUS_PENDING
	case "approved":
		return proto.RegistrationStatus_REGISTRATION_STATUS_APPROVED
	case "rejected":
		return proto.RegistrationStatus_REGISTRATION_STATUS_REJECTED
	default:
		return proto.RegistrationStatus_REGISTRATION_STATUS_UNKNOWN
	}
}

func getRestoreRequestStatus(status string) proto.RestoreRequestStatus {
	switch status {
	case "pending":
		return proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_PENDING
	case "approved":
		return proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_APPROVED
	case "rejected":
		return proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_REJECTED
	case "completed":
		return proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_COMPLETED
	default:
		return proto.RestoreRequestStatus_RESTORE_REQUEST_STATUS_UNKNOWN
	}
}
