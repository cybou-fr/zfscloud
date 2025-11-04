package tests
package main
import (
    "context"
    "crypto/ed25519"
    "crypto/rand"
    "fmt"
    "log"
    "os"
    "time"
    "github.com/ioustamora/zfscloud/proto"
    "github.com/ioustamora/zfscloud/shared/crypto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)
func main() {
    // Generate client keys
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil { log.Fatal(err) }
    
    clientID := fmt.Sprintf("test-client-%d", time.Now().Unix())
    
    // Connect
    conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { log.Fatal(err) }
    defer conn.Close()
    
    client := proto.NewBackupServiceClient(conn)
    
    // Get server public key
    ctx := context.Background()
    keyResp, err := client.GetServerPublicKey(ctx, &proto.ServerPublicKeyRequest{})
    if err != nil { log.Fatal(err) }
    
    // Generate and encrypt passphrase
    passphrase, err := crypto.GeneratePassphrase()
    if err != nil { log.Fatal(err) }
    
    encPass, err := crypto.EncryptPassphrase(passphrase, keyResp.PublicKey)
    if err != nil { log.Fatal(err) }
    
    // Create signature
    nonce, _ := crypto.GenerateNonce()
    timestamp := crypto.GetTimestamp()
    sigData := &crypto.SignatureData{
        ClientID: clientID,
        Timestamp: timestamp,
        Nonce: nonce,
        Payload: []byte(clientID),
    }
    signature := ed25519.Sign(priv, sigData.ToBytes())
    
    // Register
    req := &proto.RegisterRequest{
        ClientId: clientID,
        PublicKey: pub,
        EncryptedPassphrase: encPass,
        Metadata: &proto.ClientMetadata{
            Hostname: "test-host",
            Username: "test-user",
            Os: "linux/amd64",
            ClientVersion: "test-0.1",
        },
        Signature: signature,
        Timestamp: timestamp,
        Nonce: nonce,
    }
    
    resp, err := client.Register(ctx, req)
    if err != nil { log.Fatal(err) }
    
    fmt.Printf("SUCCESS: Registration submitted\n")
    fmt.Printf("Registration ID: %s\n", resp.RegistrationId)
    fmt.Printf("Status: %s\n", resp.Status)
    fmt.Printf("Message: %s\n", resp.Message)
    
    // Save client ID for later tests
    os.WriteFile("/tmp/test-client-id.txt", []byte(clientID), 0644)
}
