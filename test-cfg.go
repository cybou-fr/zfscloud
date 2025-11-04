package main

import (
"fmt"
"log"
"github.com/ioustamora/zfscloud/shared/config"
)

func main() {
cfg, err := config.LoadClientConfig("test-client-config-50052.yaml")
if err != nil {
%v", err)
}
fmt.Printf("ServerAddress: %s\n", cfg.Client.ServerAddress)
}
