// This package is an implementation of the Chord Algorithm
// Stoica I., et al. (2001) Chord: A Scalable Peer-to-peer Lookup Service for 
// Internet Applications. SIGCOMM’01,

package chord

import(
  "log"
  "fmt"
  "math"
  "time"
  "context"
  "math/rand"

  "google.golang.org/grpc"
  "github.com/gedilabs/abstract/node"
  "github.com/gedilabs/services/factory"
  "github.com/gedilabs/services/chord"
  //"github.com/gedilabs/services/dht"
  "github.com/gedilabs/services/discovery"
  pb "github.com/gedilabs/services/common"
)

const (
  stabilizePeriodSecs = 5
  fixFingersPeriodSecs = 5
)

type Opts struct {
  node.Opts
  SatelliteAddr string
}

type Chord struct {
  // Node struct embedded with protobuf representation 
  node.EmbeddedNode

  // May need a more considerable structure to store a Key.
  // For now, will be using integers. 
  Keys []int

  // Satellite Discovery API 
  dcl *factory.ClientInstance

  // private members
  successor *finger
  predecessor *finger

  // Stores client instance to other chord nodes 
  FingerTable []*finger

  // Indicates whether instantiated DHT node is first in a satellite network
  founder bool

 // number of bits generating an Id
  m uint64

  // Services store
  chord.UnimplementedChordServer
  serviceInstances []*factory.ServiceInstance
}

// Helper objects 

type finger struct {
  start uint64
  m uint64
  node *node.EmbeddedNode
  cl *factory.ClientInstance
}

// A convenience structure to abstract away the difference betwen rpc or local 
// function calls for chord types. See findPredecessor() for an example of the 
// usefulness of this structure, where it is not known at runtime whether
// a given node is local or remote. All we have to do is check whether 'cl'
// is a nil pointer.
type _chord struct {
  *Chord
  cl *factory.ClientInstance
}

func New(opts *Opts) (*Chord, error) {

  NodeOpts := &node.Opts{
              Port: opts.Port,
              KeysDir: opts.KeysDir,
              CustomId: opts.CustomId, }

  n, err := node.New(NodeOpts); if err != nil {
    return nil, err
  }
  pbn := n.Marshall()

  discoveryClient := NewClient("discovery", opts.SatelliteAddr)

  c := &Chord {
    EmbeddedNode: node.EmbeddedNode{*n, pbn},
		dcl: discoveryClient,
		}

  // 1. Initiate finger table
	// Assign the starting m-bit value for every finger
  c.FingerTableSkeleton()

  ctx := context.Background()

  // 2. Register with Satellite to be added to edge's iptable 
  log.Println("Registering with satellite")
  st, _ := c.dcl.Inf.(discovery.DiscoveryClient).Add(ctx, c.Marshall())
  log.Println(st.Msg)

  // 3. Start Chord server
  log.Println("Starting Chord server")
  c.NewServer("chord")

  return c, nil
}

///////////////////////////////////////////////////////////////////////////////
// Public Service Implementation 
///////////////////////////////////////////////////////////////////////////////

func (c *Chord) FindSuccessor(ctx context.Context, key *chord.Key) (*pb.Node, error) {

  log.Printf("FindSuccessor: Finding successor to id %d\n", key.Id)
  pbc := c.findPredecessor(ctx, key)

  succ := c.Unmarshall(pbc.Successor)
  log.Printf("FindSuccessor: Found successor to id %d (node: %d)\n", key.Id, succ.Int())

  return pbc.Successor, nil
}

func (c *Chord) ClosestPrecedingFinger(ctx context.Context, key *chord.Key) (*chord.ChordNode, error) {

  for i := int(c.m-1); i >= 0; i-- {

    n := c.Int()
    if n == key.Id {
      continue
    }

    fn := finger{start: c.FingerTable[i].node.Int(), m: c.m}
    log.Printf("ClosestPrecedingFinger: Checking if finger %d (st: %d, node: %d) is between (%d, %d)", i, c.FingerTable[i].start,c.FingerTable[i].node.Int(), c.Int(), key.Id)

    if fn.between(n, key.Id) {

      log.Printf("ClosestPrecedingFinger: Returning my closest finger (node %d) to id %d", fn.start, key.Id)

      if c.FingerTable[i].cl == nil {
        cl := NewClient("chord", c.FingerTable[i].node.IPAddr())
        c.FingerTable[i].cl = cl
      }

      pbn, _ := c.FingerTable[i].cl.Inf.(chord.ChordClient).Successor(ctx, &chord.Empty{})
      pbc := &chord.ChordNode{Node: c.FingerTable[i].node.Pb, Successor: pbn}

      return pbc, nil
    }
  }

  log.Printf("ClosestPrecedingFinger: No fingers between myself (node: %d) and id %d...Returning myself", c.Int(), key.Id)

  pbc := &chord.ChordNode{Node: c.Pb, Successor: c.successor.node.Pb}

  return pbc, nil
}
/*
// rpc implementation
func (c *Chord) ClosestPrecedingFinger(ctx context.Context, key *chord.Key) (*chord.ChordNode, error) {

  for i := int(c.m-1); i >= 0; i-- {

    fn := finger{start: c.FingerTable[i].node.Int(), m: c.m}

    log.Printf("ClosestPrecedingFinger: Checking if finger %d (st: %d, node: %d) is between (%d, %d)", i, c.FingerTable[i].start,c.FingerTable[i].node.Int(), c.Int(), key.Id)
    n := c.Int()
    if n == key.Id {
      continue
    }

    if fn.between(n, key.Id) {

      log.Printf("ClosestPrecedingFinger: Returning my closest finger (node %d) to id %d", fn.start, key.Id)

      if c.Self(&c.FingerTable[i].node.Node) {
        pbc := &chord.ChordNode{Node: c.Pb, Successor: c.successor.node.Pb}
        return pbc, nil

      } else if c.FingerTable[i].cl == nil {
        cl := NewClient("chord", c.FingerTable[i].node.IPAddr())
        c.FingerTable[i].cl = cl
      }

      pbn, _ := c.FingerTable[i].cl.Inf.(chord.ChordClient).Successor(ctx, &chord.Empty{})
      pcb := &chord.ChordNode{Node: c.FingerTable[i].node.Pb, Successor: pbn}

      return pcb, nil
    }
  }

  pbc := &chord.ChordNode{Node: c.Pb, Successor: c.successor.node.Pb}
  log.Printf("ClosestPrecedingFinger: No fingers between myself (node: %d) and id %d...Returning myself", c.Int(), key.Id)

  return pbc, nil
}
*/
// local implementation
func (c *Chord) closestPrecedingFinger(ctx context.Context, key *chord.Key) (*_chord, error) {

  for i := int(c.m-1); i >= 0; i-- {

    n := c.Int()
    if n == key.Id {
      continue
    }

    fn := finger{start: c.FingerTable[i].node.Int(), m: c.m}
    log.Printf("ClosestPrecedingFinger: Checking if finger %d (st: %d, node: %d) is between (%d, %d)", i, c.FingerTable[i].start,c.FingerTable[i].node.Int(), c.Int(), key.Id)

    if fn.between(n, key.Id) {

      log.Printf("ClosestPrecedingFinger: Returning my closest finger (node %d) to id %d", fn.start, key.Id)

      if c.FingerTable[i].cl == nil {
        cl := NewClient("chord", c.FingerTable[i].node.IPAddr())
        c.FingerTable[i].cl = cl
      }

      pbn, _ := c.FingerTable[i].cl.Inf.(chord.ChordClient).Successor(ctx, &chord.Empty{})
      pbc := &chord.ChordNode{Node: c.FingerTable[i].node.Pb, Successor: pbn}

      np := c.newChord(pbc, c.FingerTable[i].cl)

      return np, nil
    }
  }

  log.Printf("ClosestPrecedingFinger: No fingers between myself (node: %d) and id %d...Returning myself", c.Int(), key.Id)

  pbc := &chord.ChordNode{Node: c.Pb, Successor: c.successor.node.Pb}
  np := c.newChord(pbc, nil)

  return np, nil
}

func (c *Chord) Predecessor(ctx context.Context, null *chord.Empty) (*pb.Node, error) {

  var predecessor *pb.Node
  if c.predecessor == nil {
    log.Printf("Predecessor: Predecessor not set, sending null Node.node\n")
    predecessor = &pb.Node{}
  } else {
    predecessor = c.predecessor.node.Pb
  }
  return predecessor, nil
}

func (c *Chord) UpdatePredecessor(ctx context.Context, pbn *pb.Node) (*pb.Status, error) {

  log.Println("UpdatePredecessor: Called")

  un := c.Unmarshall(pbn)
  c.predecessor.update(un, pbn, c.m)
  log.Printf("UpdatePredecessor: New predecessor\n%+v", c.predecessor)
  log.Printf("UpdatePredecessor: New predecessor is node %d (ip: %s)", c.predecessor.node.Int(), c.predecessor.node.IPAddr())

  predecessor := c.newChord(&chord.ChordNode{Node: pbn}, nil)
  c.predecessor.cl = predecessor.cl

  c.DisplayFingerTable()

  return &pb.Status{Code: 0, Msg:""}, nil
}

func (c *Chord) Successor(ctx context.Context, null *chord.Empty) (*pb.Node, error) {

  if c.successor == nil {
    panic(fmt.Sprintf("Successor: Successor not set for node %d", c.Int()))
  }

  log.Printf("Successor: Responding with my successor (node %d)\n", c.successor.node.Int())
  return c.successor.node.Pb, nil
}

func (c *Chord) UpdateFingerTable(ctx context.Context, f *chord.Finger) (*pb.Status, error) {

  un := c.Unmarshall(f.Node)

  log.Printf("UpdateFingerTable: Checking if new node %d fits at index=%d ", un.Int(), f.Idx)

  np := c.newChord(&chord.ChordNode{Node: f.Node}, nil)
  s := &finger{start: np.Int(), m: c.m}

  if s.halfClosedLeft(c.FingerTable[f.Idx].start, c.FingerTable[f.Idx].node.Int()){
    log.Printf("UpdateFingerTable: Replacing finger %d (st: %d)\n", f.Idx, c.FingerTable[f.Idx].start)

    c.FingerTable[f.Idx].update(un, f.Node, c.m)
    c.DisplayFingerTable()

    if c.Self(&c.predecessor.node.Node) {
      return &pb.Status{}, nil

    } else if c.predecessor.cl == nil {
      predecessor := c.newChord(&chord.ChordNode{Node: c.predecessor.node.Pb}, nil)
      c.predecessor.cl = predecessor.cl
    }

    c.predecessor.cl.Inf.(chord.ChordClient).UpdateFingerTable(ctx, f)
  }

  return &pb.Status{}, nil
}

func (c *Chord) Join(ctx context.Context) {

  self := c.Marshall()
  q := &pb.Query{Id: self.Id}
  pbn, err := c.dcl.Inf.(discovery.DiscoveryClient).Random(ctx, q)

  if err != nil {
    log.Fatal(err)
  }
  peer := c.newChord(&chord.ChordNode{Node: pbn}, nil)

  if c.Self(&peer.Node){
    log.Print("Chord.Join: No other node exists on network")
    c.founder = true
  }

  c.initFingerTable(ctx, peer)
  c.updateOthers(ctx)
  log.Println("Join: Finished joining network")
  c.DisplayFingerTable()
  return
}

// pbn thinks it might be our predecessor
func (c *Chord) Notify(ctx context.Context, pbn *pb.Node) (*pb.Status, error) {

  n := c.Unmarshall(pbn)
  x := &finger{start: n.Int(), m: c.m}
  if c.predecessor == nil || x.between(c.predecessor.node.Int(), c.Int()) {
    log.Println("Notify: New predecessor discovered")

    c.UpdatePredecessor(ctx, pbn)
    c.DisplayFingerTable()
  }

  return &pb.Status{Code: 0}, nil
}

///////////////////////////////////////////////////////////////////////////////
// Private Service Implementation 
///////////////////////////////////////////////////////////////////////////////

// We are trying to find the closest preceding node to the id. In this way,
// calling successor on this node will return that node for which the key
// corresponding to the id is stored
func (c *Chord) findPredecessor(ctx context.Context, key *chord.Key) *chord.ChordNode {

    // Return predecessor if this node should be holding key 
    /*
    n := c.Int()
    if n == key.Id {
      pbc := &chord.ChordNode{Node: c.predecessor.node.Pb, Successor: c.Pb}
      return pbc
    }
    */

    // Else search for successor of key
    pbc := &chord.ChordNode{Node: c.Pb, Successor: c.successor.node.Pb}
    np := c.newChord(pbc, nil)

    log.Printf("findPredecessor: Finding key: %d\n", key.Id)

    // While the id is not between a given node 'np' and their successor 
    id := &finger{start: key.Id, m: c.m}
    for !id.halfClosedRight(np.Int(), np.successor.node.Int()) {

      // 'np' is a locally running instance
      // unexported closestPrecedingFinger() takes advantage of recycling client
      // connections to np's finger nodes
      if np.cl == nil {

        np, _ = np.closestPrecedingFinger(ctx, key)

     // else 'np' is a remote instance 
      } else {

        pbc, _ := np.cl.Inf.(chord.ChordClient).ClosestPrecedingFinger(ctx, key)
        np = c.newChord(pbc, nil)
      }
    }

  pbc = np.Marshall()
  return pbc
}

func (c *Chord) updateOthers(ctx context.Context) {

  var i uint64
  for i = 0; i < c.m; i++ {

    // find last node p whose ith finger might be n

    // one is added since findPredecessor looks over an open interval, and 
    // tberefore a node directly preceding the joining node will be overlooked  
    // (i.e., node 1 joins and node 0 exists, if one is not added, only the 
    // predecessor of node 0 will have their finger table updated and not
    // node 0 itself)
    _id := int(c.Int()) - int(math.Pow(2, float64(i)))
    // The id set is modulo 'c.m'
    if _id < 0 {
      _id = _id + int(math.Pow(2, float64(c.m)))
    }

    id := uint64(_id)
    log.Printf("updateOthers: Finding last node preceding Id (%d) to update their finger (idx=%d)", id, i)

    pbc := c.findPredecessor(ctx, &chord.Key{Id: id})
    np := c.newChord(pbc, nil)
    log.Printf("updateOthers: Found predecessor (node: %d)", np.Int())

    // 'np' is a self-reference
    if np.cl == nil {
      np.UpdateFingerTable(ctx, &chord.Finger{Node: c.Pb, Idx: i})

    } else {
      np.cl.Inf.(chord.ChordClient).UpdateFingerTable(ctx, &chord.Finger{Node: c.Pb, Idx: i})

    }
  }
}

// (Table 1. Definition of variables for node n, using m-bit identifiers) 
// Initializes the node's routing or "finger" table
// 'pbn' is an arbitrary node in the network
func (c *Chord) initFingerTable(ctx context.Context, peer *_chord) {

	// Only node on the network
	if c.founder {
		for _, f := range c.FingerTable {
			f.node = &c.EmbeddedNode
		}
    c.predecessor = &finger{node: &c.EmbeddedNode, m: c.m}
    return
	}

	// Else fill out finger table
  log.Printf("initFingerTable: Joining with peer node %d\n", peer.Int())

  st := c.FingerTable[0].start

  // First find successor node
  pbn, _ := peer.cl.Inf.(chord.ChordClient).FindSuccessor(ctx, &chord.Key{Id: st})

  successor := c.newChord(&chord.ChordNode{Node: pbn}, nil)
  /*
  c.FingerTable[0].node = &successor.EmbeddedNode
  c.FingerTable[0].cl = successor.cl
  */
  c.successor.node = &successor.EmbeddedNode
  c.successor.cl = successor.cl

  log.Printf("initFingerTable: Found my successor node %d\n", successor.Int())

  // Joining node now inherits their successor's predecessor 
  pbn, _ = successor.cl.Inf.(chord.ChordClient).Predecessor(ctx, &chord.Empty{})
  un := c.Unmarshall(pbn)

  c.predecessor = &finger{node: &node.EmbeddedNode{*un, pbn}, m: c.m}

  // Update successor that you are now their predecessor
  log.Printf("initFingerTable: calling UpdatePredecessor() on node %d", c.successor.node.Int())
  c.successor.cl.Inf.(chord.ChordClient).UpdatePredecessor(ctx, c.Pb)
  log.Printf("initFingerTable: called UpdatePredecessor() on node %d", c.successor.node.Int())

  var i uint64
  for i = 0; i < c.m-1; i++ {

		if c.FingerTable[i+1].halfClosedLeft(c.Int(), c.FingerTable[i].node.Int())  {
			c.FingerTable[i+1].node = c.FingerTable[i].node

		} else {

      pbn, _ = peer.cl.Inf.(chord.ChordClient).FindSuccessor(ctx, &chord.Key{Id: c.FingerTable[i+1].start})

      un := c.Unmarshall(pbn)
      c.FingerTable[i+1].node = &node.EmbeddedNode{*un, pbn}
		}

    log.Printf("initFingerTable: Initiated finger %d (st: %d) with node %d", i+1, c.FingerTable[i+1].start, c.FingerTable[i+1].node.Int())
	}

  return
}

func (c *Chord) Stabilize(ctx context.Context) {

  stabilizeTicker := time.NewTicker(stabilizePeriodSecs * time.Second)

  for {
    select {
      case <-stabilizeTicker.C:
        c.stabilize(ctx)
    }
  }
}

// periodically verify n's immediate successor
// and tell the successor about n
func (c *Chord) stabilize(ctx context.Context) {

  if c.successor.cl == nil {
    return
  }

  pbn, _ := c.successor.cl.Inf.(chord.ChordClient).Predecessor(ctx, &chord.Empty{})
  un := c.Unmarshall(pbn)

  x := &finger{start: un.Int(), m: c.m}

  // new successor found 
  if x.between(c.Int(), c.successor.node.Int()) {

    c.successor.update(un, pbn, c.m)

    successor := c.newChord(&chord.ChordNode{Node: pbn}, nil)
    c.successor.cl = successor.cl

    log.Println("Stabilize: Updated successor")
    c.DisplayFingerTable()
  }

  // successor is self 
  if c.successor.cl == nil {
    c.Notify(ctx, c.Pb)
  } else {
    c.successor.cl.Inf.(chord.ChordClient).Notify(ctx, c.Pb)
  }
}

func (c *Chord) fixFingers(ctx context.Context) {

  rand.Seed(time.Now().UnixNano())

  // rand.Intn(n) returns random int [0, n)
  i := rand.Intn(int(c.m - 1)) + 1

  log.Printf("FixFingers: Checking finger %d\n", i)

  pbn, _ := c.FindSuccessor(ctx, &chord.Key{Id: c.FingerTable[i].start})
  un := c.Unmarshall(pbn)

  // no update to finger needed
  if c.FingerTable[i].node.Self(un) {
    return
  }

  // else update finger in table and close former client connection if necessary
  log.Printf("FixFingers: Replacing finger %d (st: %d)\n", i, c.FingerTable[i].start)
  c.FingerTable[i].update(un, pbn, c.m)

  c.DisplayFingerTable()

  return

}

func (c *Chord) FixFingers(ctx context.Context) {

  fixFingersTicker := time.NewTicker(fixFingersPeriodSecs * time.Second)

  for {
    select {
    case <- fixFingersTicker.C:
      c.fixFingers(ctx)
    }
  }
}
///////////////////////////////////////////////////////////////////////////////
// Helper Object Implementation 
///////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////
// chord 
///////////////////////////////////////////////////////////////////////////////

func (c *Chord) newChord(_c *chord.ChordNode, _cl *factory.ClientInstance) *_chord {

  n := c.Unmarshall(_c.Node)
  if c.Self(n){
    log.Println("newChord: Returning self")
    return &_chord{c, nil}
  }

  // create new client instance if not available
  cl := _cl
  if cl == nil {
    log.Println("newChord: Creating client instance")
    cl = NewClient("chord", n.IPAddr())
  }

  succ := &finger{}

  if _c.Successor != nil {
    sn := c.Unmarshall(_c.Successor)
    succ = &finger{node: &node.EmbeddedNode{*sn, _c.Successor}}
  }

  np := &_chord{&Chord{
          EmbeddedNode: node.EmbeddedNode{*n, _c.Node},
          successor: succ,
          },
        cl,
      }

  return np
}

func (_c *_chord) Marshall() *chord.ChordNode {
  if _c.successor != nil {
    return &chord.ChordNode{Node: _c.Pb, Successor: _c.successor.node.Pb}
  }
  return &chord.ChordNode{Node: _c.Pb}
}

///////////////////////////////////////////////////////////////////////////////
// finger 
///////////////////////////////////////////////////////////////////////////////

// update node pointed to by finger
func (f *finger) update(n *node.Node, pbn *pb.Node, m uint64) {

  former := f.node

  f.erase()
  f.node = &node.EmbeddedNode{*n, pbn}

  log.Printf("finger.update: Replaced node (id: %d, ip: %s) with (id: %d, ip: %s)",
  former.Int(), former.IPAddr(), n.Int(), n.IPAddr())

  return
}

// expunge memory of node pointed to by finger
func (f *finger) erase() {

  if f.cl != nil {
    f.cl.Conn.Close()
    f.cl = nil
  }
  f.node = nil
}

// Returns true if the finger's start position is between node 'n' and
// finger 'i' in the circular m-bit Chord space   
func (f *finger) halfClosedLeft(k1 uint64, k2 uint64) bool {

  if k1 > k2 || k1 == k2 {
    log.Printf("halfClosedLeft: k1=%d and k2=%d special case", k1, k2)

    lastId := uint64(math.Pow(2, float64(f.m)))
    if ((f.start >= k1) && (f.start <= (lastId - 1))) || ((f.start >= 0) && (f.start < k2)) {
      return true
    }

    return false
  }

	if (f.start >= k1) && (f.start < k2) {
		return true
	}
	return false
}
func (f *finger) halfClosedRight(k1 uint64, k2 uint64) bool {

  if k1 > k2 || k1 == k2 {

    lastId := uint64(math.Pow(2, float64(f.m)))
    log.Printf("halfClosedRight: lastId=%d, k1=%d, k2=%d special case", lastId, k1, k2)

    if ((f.start > k1) && (f.start <= (lastId - 1))) || ((f.start >= 0) && (f.start <= k2)) {
      return true
    }

    return false
  }

	if (f.start > k1) && (f.start <= k2) {
		return true
	}
	return false
}
func (f *finger) between(k1 uint64, k2 uint64) bool {

  if k1 > k2 {

    log.Printf("between: finger=%+v\n", f)
    lastId := uint64(math.Pow(2, float64(f.m)))
    log.Printf("between: lastId=%d, k1=%d, k2=%d", lastId, k1, k2)

    if ((f.start > k1) && (f.start <= (lastId - 1))) || ((f.start >= 0) && (f.start < k2)) {
      return true
    }

    return false
  }

	if (f.start > k1) && (f.start < k2) {
		return true
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
// Helper Functions
///////////////////////////////////////////////////////////////////////////////

// Initialize the finger table with minimum distances for each finger (i.e., start)
func (c *Chord) FingerTableSkeleton() {

  c.m = uint64(c.Id.Length)
  n := c.Int()

  // debug: I'm lazy and assume all nodes employ a 3-bit ID if this one can 
  if (n < 8) {
    c.m = 3
  }

  c.FingerTable = make([]*finger, c.m)

  var k uint64
  for k = 0; k < c.m; k++ {
    root := float64(n) + math.Pow(2, float64(k))
    mod := math.Pow(2, float64(c.m))

    st := uint64(root) % uint64(mod)

    c.FingerTable[k] = &finger {
      start: st,
      m: c.m,
    }
  }
  c.successor = c.FingerTable[0]

}
func (c *Chord) DisplayFingerTable() {
  for k, v := range c.FingerTable {
    fmt.Printf("finger %d (st: %d): %d\n", k, v.start, v.node.Int())
  }
  fmt.Printf("predecessor: %d\n", c.predecessor.node.Int())
}

///////////////////////////////////////////////////////////////////////////////
// Factory Functions
///////////////////////////////////////////////////////////////////////////////

// see `github.com/gedilabs/services/factory.go` for service names 
func (c *Chord) NewServer(serviceName string) {

  // Optionally add grpc server options in the future
  var opts []grpc.ServerOption
  si, err := factory.NewServerInstance(c, serviceName, opts); if err != nil {
    panic(err)
  }
  c.serviceInstances = append(c.serviceInstances, si)
}

func (c *Chord) ListenAndServe() {
  for _, si := range c.serviceInstances {
      lis := si.Listener
      srv := si.Server
      go func() {
        log.Printf("Listening on: %s:%d\n", si.Addr, si.Port)
        srv.Serve(lis)
    }()
  }
}

func NewClient(serviceName string, serverAddr string)(*factory.ClientInstance){
  var opts []grpc.DialOption
  opts = append(opts, grpc.WithInsecure())
	cl, err := factory.NewClientInstance(serviceName, serverAddr, opts)
	if err != nil {
		panic(err)
	}
	return cl
}


