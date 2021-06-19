package factory

import (
  "fmt"
	"net"
  "log"
  "reflect"
  "errors"
  "google.golang.org/grpc"
  "github.com/gedilabs/services/discovery"
  "github.com/gedilabs/services/chord"
)

type ServiceInstance struct {
  Server *grpc.Server
  Name string
	Addr string
  Port int
	Listener net.Listener // this is an interface
}

type ClientInstance struct {
  Conn *grpc.ClientConn
  Inf interface{}
}

func NewServerInstance(server interface{}, serviceName string, opts []grpc.ServerOption) (si *ServiceInstance, err error) {
  grpcServer := grpc.NewServer(opts...)
  _, err = call(registrar, serviceName, grpcServer, server)
  if err != nil {
    return
  }

  serverPort := ports[serviceName]
  if serverPort == 0 {
    err = errors.New(fmt.Sprintf("factory.NewServerInstance: There is no key '%s' in factory.ports", serviceName))
    return
  }

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", ports[serviceName]))
	if err != nil {
		return
	}
  si = &ServiceInstance{Server: grpcServer,
                         Name: serviceName,
                         Port: ports[serviceName],
												 Addr: "0.0.0.0",
												 Listener: lis,
												 }
  return
}

func NewClientInstance(serviceName string, serverAddr string, opts []grpc.DialOption) (*ClientInstance, error) {
  opts = append(opts, grpc.WithBlock())

  serverAddr = fmt.Sprintf("%s:%d", serverAddr, ports[serviceName])
  log.Printf("Dialing '%s' server: %s", serviceName, serverAddr)
  conn, err := grpc.Dial(serverAddr, opts...)
  if err != nil {
    return nil, err
  }
	res, err := call(api, serviceName, conn)
	if err != nil {
		panic(err)
	}
  // Get the interface of the first returned value (i.e., serviceName.[serviceName]Client)
  inf := res[0].Interface()
  cl := &ClientInstance {Conn: conn,
                         Inf: inf}
  log.Printf("Successfully connected to '%s' server", serviceName)
	return cl, nil
}

var registrar = map[string]interface{} {
  "discovery": discovery.RegisterDiscoveryServer,
  "chord": chord.RegisterChordServer,
}

var api = map[string]interface{} {
  "discovery": discovery.NewDiscoveryClient,
  "chord": chord.NewChordClient,
}

var ports = map[string]int {
  "discovery": 12130,
  "chord": 12140,
}

func call(m map[string]interface{},
                    name string,
                    params ...interface{})(result []reflect.Value, err error) {

  if m[name] == nil {
    err = errors.New(fmt.Sprintf("factory.call: There is no key '%s' in the server registrar", name))
    return
  }

  f := reflect.ValueOf(m[name])
  if len(params) != f.Type().NumIn() {
    err = errors.New("factory.call: the number of params is not adapted.")
    return
  }
  in := make([]reflect.Value, len(params))
  for k, param := range params {
    in[k] = reflect.ValueOf(param)
  }
  result = f.Call(in)
  return
}


