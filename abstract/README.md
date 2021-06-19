A `Node` is the concrete type for all functionaries in the GEDI network

While performing its function, a `Node` may need to request
services from other functionaries to fulfill its role.
They too may be solicitated by other nodes to provide services
on their behalf for the end user or `Patron`.

```
type Patron interface {}

type Actor interface {}


type Server interface {
  NewServer()
}
```

 All worker nodes behave as `Actors` or `Servers` depending on their roles
 and the particular state of the network during the time of a request from a 
 `Patron`.

```
type ClientInstance struct {
  Conn grpc.ClientConn
  Inf interface{}
}
```
Every instance of an open connection from a client actively making requests
from a server is characterized by the `ClientInstance` struct. Notice the 
`Inf` field is left anonymous to support various types of gRPC interfaces, 
depending on the nature of the service requested.

```
type ServiceInstance struct {
  Server *grpc.Server
  Name string
  Port int
}
```

In a recipricol manner, every instance of a service in performance is 
characterized by the `ServiceInstance` struct, which stores the name of the
service; in addition to the accountable `Server` and client node. 

The `ServiceName` can be a key value mapping to a gRPC-related package.

Every instance of a service request . Eventually, transaction IDs will be used
to keep track of a global network state of `Patron` data permissions;
Paying to rent other data or analyze those currently in its possession. 

Functionaries

```
type Satellite struct {}

type Archive struct {}

type Processor struct {}

type Coordinator struct {}
```

All code for services are defined in: "github.com/gedilabs/services".






