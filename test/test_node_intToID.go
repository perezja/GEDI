package main

import(
  "fmt"
  "flag"
  "log"
  "github.com/gedilabs/abstract/node"
  mb "github.com/multiformats/go-multibase"
  mh "github.com/multiformats/go-multihash"
)

var(
  lisAddr = flag.String("ip", "0.0.0.0", "Server listening address")
  port = flag.Int("port", 12132, "Server port")
  keysDir = flag.String("ssh_dir", "", "SSH Keys Directory")
)

func main() {
  flag.Parse()

  if *keysDir == "" {
    log.Fatal("Provide the directory storing RSA keys '-ssh_dir'")
  }
  digest := []byte{7}
  //n, _ := node.New(&node.Opts{Port:*port, KeysDir:*keysDir, IdHash:idHash, IdRepr: "HEX"})
  n, _ := node.New(&node.Opts{Port:*port, KeysDir:*keysDir, IdHash: digest, IdRepr: "BINARY"})
  fmt.Printf("Int->ID: %d\n", id)
}
