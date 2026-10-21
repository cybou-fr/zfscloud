//go:build ignore

// Standalone developer tool; run explicitly: go run final-test.go

package main
import ("crypto/ed25519"; "encoding/base64"; "fmt"; "log"; "time"; "github.com/ioustamora/zfscloud/client"; "github.com/ioustamora/zfscloud/proto"; "github.com/ioustamora/zfscloud/shared/config"; "github.com/ioustamora/zfscloud/shared/crypto")
func main() {
cfg, _ := config.LoadClientConfig("test-client-config-50052.yaml")
pub, priv, _ := ed25519.GenerateKey(nil)
clientID := "test" + time.Now().Format("150405")
c := client.NewClient(cfg.Client.ServerAddress, clientID, priv, pub)
c.Connect()
defer c.Close()
pubKey, _ := c.GetServerPublicKey()
fmt.Printf("Connected! Got server key (%d bytes)\n", len(pubKey))
passphrase, _ := crypto.GeneratePassphrase()
encPass, _ := crypto.EncryptPassphrase(passphrase, pubKey)
metadata := &proto.ClientMetadata{Hostname: "test", Username: "test", Os: "linux", ClientVersion: "test"}
if err := c.Register(metadata, encPass); err != nil {
%v", err)
}
fmt.Println("Registered! Passphrase:", base64.StdEncoding.EncodeToString(passphrase))
fmt.Println("Run: ./admin-helper.sh list-pending")
fmt.Println("Then: ./admin-helper.sh approve <id> 50")
}
