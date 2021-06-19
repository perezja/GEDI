package main

import(
  "fmt"
  "flag"
  "log"
  //"path"
  "github.com/gedilabs/abstract/node"
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

  // Create a new node with default options
  var id uint64 = 0
  myID := node.IdKey{Key: &id}
  n, _ := node.New(&node.Opts{Port:*port, KeysDir:*keysDir, CustomId: myID })
  fmt.Printf("Nodes IP Adress: %s", n.IPAddr())
  //n, _ := node.New(&node.Opts{Port:*port, KeysDir:*keysDir, IdHashType: mh.SHA1})

  //n, _ := node.New(&node.Opts{Port:*port, KeysDir:*keysDir})


  // Test if can marshall into protobuf node message
  pb := n.Marshall()
  fmt.Printf("protobuf: %+v\n", pb)
  pbtonode := n.Unmarshall(pb)
  newpb := pbtonode.Marshall()
  fmt.Printf("new protobuf: %+v\n", newpb)
  fmt.Printf("new protobuf equals old protobuf: %v\n", newpb.Id==pb.Id)

  return


  // View attributes of the node's ID (decoded multihash)
  fmt.Printf("multihash: (%v): %v 0x%x %d\n", n.Id.Digest, n.Id.Name, n.Id.Code, n.Id.Length)

}
