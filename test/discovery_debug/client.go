package main

import (
  "fmt"
  "flag"
  "github.com/gedilabs/services/factory"
  "google.golang.org/grpc"
  "context"
  "github.com/gedilabs/services/discovery"
  pb "github.com/gedilabs/services/common"
)

var(
  satelliteAddr = flag.String("srv", "", "Server listening address")
)

func main() {
  flag.Parse()

  if *satelliteAddr == "" {
    panic("provide satellite address '-srv'")
  }

  ctx := context.Background()
  var opts []grpc.DialOption
  opts = append(opts, grpc.WithInsecure())
  dcl, err := factory.NewClientInstance("discovery", *satelliteAddr, opts)
  if err != nil { panic(err)}

  q := &pb.Query{Id: "000000000000010000000000000000000000000000000000000000000000000000000000000000011"}
  pbn, err := dcl.Inf.(discovery.DiscoveryClient).Random(ctx, q)

  if err != nil { panic(err)}
  fmt.Printf("Received random node: %+v", pbn)
}
