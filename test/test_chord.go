package main

import (
  "log"
  "flag"
  "github.com/gedilabs/abstract/chord"
  "github.com/gedilabs/abstract/node"
  "context"
  //"github.com/gedilabs/services/dht"
)

var(
  lisAddr = flag.String("ip", "0.0.0.0", "Server listening address")
  port = flag.Int("port", 12132, "Server port")
  keysDir = flag.String("ssh_dir", "", "SSH keys directory")
  satelliteAddr = flag.String("srv", "", "Satellite Address")
)

func main() {
  flag.Parse()

  if *satelliteAddr == "" {
    log.Fatal("Provide the address of the Satellite to connect to '-srv'")
  }

  var id uint64 = 3
  opts := &chord.Opts{Opts: node.Opts{Port:*port, KeysDir:*keysDir, CustomId: node.IdKey{Key: &id}}, SatelliteAddr:*satelliteAddr}

  c, err := chord.New(opts)
  if err != nil {
    log.Fatal(err)
  }
  ctx := context.Background()

  go c.ListenAndServe()
  c.Join(ctx)
  go c.Stabilize(ctx)
  go c.FixFingers(ctx)
  select{}
}
