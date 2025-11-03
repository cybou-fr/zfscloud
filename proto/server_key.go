// Temporary stub for ServerPublicKey types until proto regeneration
// TODO: Regenerate proto files properly with protoc

package proto

// ServerPublicKeyRequest - request for server's RSA public key
type ServerPublicKeyRequest struct {
	// No fields - public key is freely available
}

// ServerPublicKeyResponse - response with server's RSA public key
type ServerPublicKeyResponse struct {
	PublicKey []byte // RSA public key in PEM format
	KeyId     string // Key identifier for key rotation
}

// Reset implements protobuf message interface
func (x *ServerPublicKeyRequest) Reset() {
	*x = ServerPublicKeyRequest{}
}

// String implements protobuf message interface
func (x *ServerPublicKeyRequest) String() string {
	return "ServerPublicKeyRequest{}"
}

// ProtoMessage implements protobuf message interface
func (*ServerPublicKeyRequest) ProtoMessage() {}

// Reset implements protobuf message interface
func (x *ServerPublicKeyResponse) Reset() {
	*x = ServerPublicKeyResponse{}
}

// String implements protobuf message interface
func (x *ServerPublicKeyResponse) String() string {
	return "ServerPublicKeyResponse{...}"
}

// ProtoMessage implements protobuf message interface
func (*ServerPublicKeyResponse) ProtoMessage() {}

// GetPublicKey returns the public key
func (x *ServerPublicKeyResponse) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

// GetKeyId returns the key ID
func (x *ServerPublicKeyResponse) GetKeyId() string {
	if x != nil {
		return x.KeyId
	}
	return ""
}
