package satellite

import(
  "fmt"
	"log"
  "context"
  "errors"
  "github.com/jackc/pgx/v4"
	"github.com/gedilabs/abstract/node"
	//"github.com/gedilabs/abstract/common"

  "google.golang.org/grpc"
	"github.com/gedilabs/services/factory"
	"github.com/gedilabs/services/discovery"
  pb "github.com/gedilabs/services/common"
)

// Find a way to dynamically access the routing table from different table names
const (
  // Postgres routing table name
  routingTable = "actors"
)

type Opts struct {
  Port int        // listening port
  KeysDir string  // RSA keys directory
  DbConnString string
}

type Satellite struct {
  node.Node
  Db *pgx.Conn

  // Services store
  discovery.UnimplementedDiscoveryServer
  serviceInstances []*factory.ServiceInstance
}

func New(opts *Opts) (*Satellite, error) {

  NodeOpts := &node.Opts{
              Port: opts.Port,
              KeysDir: opts.KeysDir,
            }

	n, err := node.New(NodeOpts)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.Connect(context.Background(), opts.DbConnString)
	if err != nil {
		return nil, err
	}

	s := &Satellite {
		Node: *n,
		Db: conn,
		}

  s.NewServer("discovery")

	return s, nil
}

///////////////////////////////////////////////////////////////////////////////
// Public Member Functions
///////////////////////////////////////////////////////////////////////////////

// see `github.com/gedilabs/services/factory.go` for service names
func (s *Satellite) NewServer(serviceName string) {

  // Optionally add grpc server options in the future
  var opts []grpc.ServerOption
  si, err := factory.NewServerInstance(s, serviceName, opts); if err != nil {
    panic(err)
  }
  s.serviceInstances = append(s.serviceInstances, si)
}

func (s *Satellite) ListenAndServe() {
	for _, si := range s.serviceInstances {
			lis := si.Listener
			srv := si.Server
			go func() {
			  log.Printf("Listening on: %s:%d\n", si.Addr, si.Port)
				srv.Serve(lis)
		}()
	}
}

///////////////////////////////////////////////////////////////////////////////
// Service Declarations
///////////////////////////////////////////////////////////////////////////////

func (s *Satellite) Add(ctx context.Context, n *pb.Node) (*pb.Status, error){
  sql := "INSERT INTO actors (id, addr) VALUES ($1, $2)"
  _, err := s.Db.Exec(ctx, sql, n.Id, n.Addrs)
  if err != nil {
    panic(err)
    return &pb.Status{Code: 1, Msg: fmt.Sprintf("Satellite (id: %v) failed to add: %+v", n, s.Id)}, err
  }
  node := s.Unmarshall(n)
  return &pb.Status{Code: 0, Msg: fmt.Sprintf("Satellite (id: %v) registered node %+v", s.Marshall().Id, node.Int())}, err
}

func (s *Satellite) Random(ctx context.Context, q *pb.Query) (*pb.Node, error){
  // Check number of nodes in DHT
  dhtSize := s.checkDHTSize(ctx)
  if dhtSize == 0 {
    err := errors.New("Satellite.Random: No nodes registered")
    return &pb.Node{}, err
  } else if dhtSize > 1 && q.Id == "" {
    err := errors.New("Satellite.Random: Query.Id must be populated with client node Id")
    return &pb.Node{}, err
  }

  sql := "SELECT id, addr FROM actors ORDER BY RANDOM() LIMIT 1";

  var n *pb.Node
  var id, addrs string

  for n == nil {
    draw := func() *pb.Node {
      rows, err := s.Db.Query(ctx, sql)
      if err != nil {
        panic(err)
      }
      defer rows.Close()
      for rows.Next() {
        if err = rows.Scan(&id, &addrs); err != nil {
          panic(err)
        }
        if dhtSize > 1 && id == q.Id {
          log.Printf("Satellite.Random: Drawing again...\n")
          return nil
        }
      }
      return &pb.Node{Id: id, Addrs: addrs}
    }
    n = draw()
  }

  return n, nil
}

func (s *Satellite) Find(ctx context.Context, q *pb.Query) (*pb.Node, error){
  sql := "SELECT id, addr FROM actors WHERE id=$1"
  row := s.Db.QueryRow(ctx, sql, q.Id)

  var id, addrs string
  switch err := row.Scan(&id, &addrs); err {
    case nil:
      return &pb.Node{Id: id, Addrs: addrs}, nil

    case pgx.ErrNoRows:
      return &pb.Node{}, nil

    default:
      return &pb.Node{}, err
   }
}

func (s *Satellite) checkDHTSize(ctx context.Context) (count int) {
  sql := "SELECT COUNT(*) as count from actors"

  rows, err := s.Db.Query(ctx, sql); if err != nil {
    panic(err)
  }

  defer rows.Close()
	for rows.Next() {
    if err := rows.Scan(&count);  err != nil {
      panic(err)
    }
  }
   return count
}
