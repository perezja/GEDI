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
  id_mhash, _ := mh.Encode(n.Id.Digest, n.Id.Code)
  fmt.Printf("ID mhash: %d\n", id_mhash)
  id_string, _ := mb.Encode(mb.Base2, n.Id.Digest)
  fmt.Printf("ID digest: %d\n", n.Id.Digest)
  fmt.Printf("ID string (multibase=base2): %v\n", id_string)
  fmt.Printf("protobuf: %+v\n", n.Marshall())
  fmt.Printf("multihash: %v, %v 0x%x %d\n", n.Id.Digest, n.Id.Name, n.Id.Code, n.Id.Length)
  idint := n.ParseIdToInt()
  fmt.Printf("Id->Int: %d\n", idint)
}
