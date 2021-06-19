package node

import(
  "log"
  //"fmt"
  "strconv"
  //"bytes"
  "encoding/binary"
  "path"
  "github.com/gedilabs/crypto"
  "github.com/gedilabs/network"
  ma "github.com/multiformats/go-multiaddr"
  mh "github.com/multiformats/go-multihash"
  mb "github.com/multiformats/go-multibase"
  pb "github.com/gedilabs/services/common"
)

type Node struct {
  Id *mh.DecodedMultihash
  Addr ma.Multiaddr // Interface
  baseEncoding mb.Encoding // Used for ID string to integer conversions
}

// EmbeddedNode is merely a composite struct of Node with its protobuf
// message representation to counsel against the costly use of Marshall()
// and Unmarshall() calls
type EmbeddedNode struct {
  Node
  Pb *pb.Node
}

// IdKey allows to user to set a custom ID for the node as an integer
type IdKey struct {
  Key *uint64
}

type Opts struct {
  Port int  // listening port
  KeysDir string  // RSA keys directory
  CustomId IdKey // Use the key provided as Id

  // Uint code of the hash type that is supported by multiformats/multihash,
  // https://github.com/multiformats/go-multihash/blob/master/multihash.go 
  // HashType is a simple interface wrapping multihash codes, implementing
  // checksums for the given hash function

  IdHashType uint64

  // A base encoding constant supported by multiformats/multibase 
  // https://github.com/multiformats/go-multibase/blob/master/multibase.go
  IdBaseEncoding mb.Encoding
}

func New (opts *Opts) (*Node, error){

  if err := checkOpts(opts); err != nil {
    log.Fatal(err)
  }

  // Set default hash type and base encoding
  n := &Node {
    Id: generateKey(opts),
    Addr: getAddr(opts),
    baseEncoding: opts.IdBaseEncoding,
    }

  return n, nil

}

///////////////////////////////////////////////////////////////////////////////
// Public Memeber Methods
///////////////////////////////////////////////////////////////////////////////

// Convert a node's self reference into a protobuf message specifying a multibase encoded 
// string as an ID and a multiaddress encoded string as the network location
func (n *Node) Marshall() *pb.Node {

  var mbase string

  //fmt.Printf("node.Marshall: ID bytes: %2b\n", n.Id.Digest)
  mhash, _ := mh.Encode(n.Id.Digest, n.Id.Code)

  // Encode resulting byte buffer as a multibase string
  //fmt.Printf("node.Marshall: mh.Encode: %2b\n", mhash)
  mbase, _ = mb.Encode(n.baseEncoding, mhash)

  return &pb.Node{Id: mbase, Addrs: n.Addr.String()}
}

// Convert a protobuf node message into a struct 
func (n *Node) Unmarshall(pbn *pb.Node) (*Node){
  if pbn == nil {
    return nil
  }

  // Decode multibase string into encoded multihash digest (byte slice)
  _, mhash, err := mb.Decode(pbn.Id); if err != nil {
    panic(err)
  }
  // Decode multihash digest into decoded multihash
  dmh, err := mh.Decode(mhash); if err != nil {
    panic(err)
  }
  maddr, err := ma.NewMultiaddr(pbn.Addrs); if err != nil {
    panic(err)
  }
  return &Node{Id: dmh,
               Addr: maddr,
               baseEncoding: n.baseEncoding}
}

// Converts ID to Key representation 
func (n *Node) Int() uint64 {

  bytes := n.Id.Digest

  // Truncate digest to max number of bytes in uint64
  if n.Id.Length > 8 {
    bytes = n.Id.Digest[:8]
  }

  num := binary.BigEndian.Uint64(bytes)
  return num
}

// Convert an integer representing a node in modular space to an ID string
// The string will have the same base encoding as the calling node  
func (n *Node) ParseIntToId(num uint64) (string, error) {

  bytes := make([]byte, 8)
  binary.BigEndian.PutUint64(bytes, num)

  //fmt.Printf("ParseIntToId: ID bytes: %2b\n", bytes)
  // mhash is a byte slice
  mhash, err := mh.Encode(bytes, n.Id.Code)
  if err != nil {
    return "", err
  }
  //fmt.Printf("ParseIntToId: mh.Encode: %2b\n", mhash)

	mbase, err := mb.Encode(n.baseEncoding, mhash)
  if err != nil {
    return "", err
  }

  return mbase, err
}

// Check if a protobuf node refers to self 
// May be good to allow IDs to represent different base encodings
func (n *Node) Self(oth *Node) (self bool) {

  if oth.Marshall().Id == n.Marshall().Id {
    self = true
  } else {
    self = false
  }
  return
}

func (n *Node) IPAddr() (ip string) {

  for _, component := range ma.Split(n.Addr) {
    for _, pr := range component.Protocols() {
      if pr.Name == "ip4" {
        ip = path.Base(component.String())
				return
			}
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////////////
// Helper Functions 
///////////////////////////////////////////////////////////////////////////////

// This function is a wrapper for ID generation to ease transition into 
// supporting multiple public key types. Also, allows quick changes to 
// the `Node` ID generation scheme for debugging
func generateKey (opts *Opts) (*mh.DecodedMultihash) {

	// If IdKey provided, use the key (integer) as the ID for the given node. 
  // This allows myself to debug Chord with terse IDs (i.e., 4-bit)
  if opts.CustomId.Key != nil {

    bytes := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes, *opts.CustomId.Key)

    mhash, err := mh.Encode(bytes, opts.IdHashType)
		if err != nil {
			log.Fatal(err)
		}
		dmh, _ := mh.Decode(mhash)
		return dmh
	}

  // Else generate ID as hash of the user's public key 
  rsaPubKeyPath := path.Join(opts.KeysDir, "id_rsa.pub")

  ht := crypto.HashType(opts.IdHashType)
  mhash, err := crypto.IDFromRSAPubKey(ht, rsaPubKeyPath)
	if err != nil {
	  log.Fatal(err)
  }

  // To allow for fluency between node ID and integer key translations, 
  // the hash will be truncated to 8 bytes so a key representation of the 
  // node can be translated back to the correct node ID.

	dmh, _ := mh.Decode(mhash)
  trmh , _ := mh.Encode(dmh.Digest[:8], dmh.Code)
	trdmh, _ := mh.Decode(trmh)

	return trdmh
}

// Currently only supports node connections using ip4 and TCP 
func getAddr(opts *Opts) (ma.Multiaddr){

  ip, err := network.GetLocalIPAddr(); if err != nil {
    panic(err)
  }
  addr := path.Join("/ip4", ip, "tcp", strconv.Itoa(opts.Port))

  maddr, err := ma.NewMultiaddr(addr); if err != nil {
    panic(err)
  }
  return maddr
}

func checkOpts(opts *Opts) error {

  // If setting integer as ID, provide clear ID encoding
  // for legibility
  if opts.CustomId.Key != nil {
    opts.IdHashType = mh.IDENTITY
    opts.IdBaseEncoding = mb.Base2

  // Else for pubkey hashes as ID, use efficient encoding

  } else {

    opts.IdHashType = mh.SHA1
    opts.IdBaseEncoding = mb.Base58BTC

  }

  // Check if base encoding string is valid
  if _, err := mb.Encode(opts.IdBaseEncoding, []byte{1}); err != nil {
    return err
  }

  return nil
}
