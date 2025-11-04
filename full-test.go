package main

import (
"crypto/ed25519"
"encoding/base64"
"fmt"
"log"
"time"

"github.com/ioustamora/zfscloud/client"
"github.com/ioustamora/zfscloud/proto"
"github.com/ioustamora/zfscloud/shared/config"
"github.com/ioustamora/zfscloud/shared/crypto"
)

func main() {
cfg, _ := config.LoadClientConfig("test-client-config-50052.yaml")
fmt.Printf("Server: %s\n\n", cfg.Client.ServerAddress)

pub, priv, _ := ed25519.GenerateKey(nil)
clientID := "test-" + time.Now().Format("150405")
fmt.Printf("Client ID: %s\n\n", clientID)

c := client.NewClient(cfg.Client.ServerAddress, clientID, priv, pub)

fmt.Println("Connecting...")
if err := c.Connect(); err != nil {
err)
}
defer c.Close()
fmt.Println("Connected!\n")

fmt.Println("Getting server public key...")
pubKey, err := c.GetServerPublicKey()
if err != nil {
err)
}
fmt.Printf("Got key (%d bytes)\n\n", len(pubKey))

fmt.Println("Registering...")
passphrase, _ := crypto.GeneratePassphrase()
encPass, _ := crypto.EncryptPassphrase(passphrase, pubKey)

metadata := &proto.ClientMetadata{
ame:      "test",
ame:      "test",
          "linux",
tVersion: "test",
}

if err := c.Register(metadata, encPass); err != nil {
 failed: %v", err)
}

fmt.Println("Registered!")
fmt.Printf("PASSPHRASE: %s\n\n", base64.StdEncoding.EncodeToString(passphrase))
fmt.Println("NOW: Run ./admin-helper.sh list-pending")
fmt.Println("THEN: Run ./admin-helper.sh approve <id> 50")
fmt.Println("Press ENTER when done...")
fmt.Scanln()

fmt.Println("\nChecking status...")
status, _ := c.GetRegistrationStatus()
fmt.Printf("Status: %v\n\n", status.Status)

if status.Status == proto.RegistrationStatus_REGISTRATION_STATUS_APPROVED {
tln("APPROVED! Testing heartbeat...")
:= c.Heartbeat(); err != nil {
%v", err)
tln("SUCCESS!")
}
}
