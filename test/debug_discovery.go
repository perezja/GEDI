package main

import (
	"fmt"
  "log"
  "flag"
  "context"
	"encoding/hex"
  "strconv"
  "github.com/gedilabs/abstract/satellite"
  pb "github.com/gedilabs/services/common"
  "github.com/gedilabs/crypto"
)

var(
  lisAddr = flag.String("ip", "0.0.0.0", "Server listening address")
  port = flag.Int("port", 12132, "Server port")
  keysDir = flag.String("ssh_dir", "", "SSH keys directory")
  dbUrl = flag.String("db_url", "postgresql://localhost:5432/rodinia?user=rodinia&password=rodinia", "The database connection URL (e.g., postgresql://<ip>:<port>/<db>?user=<user>)")
)
func AppendByte(slice []byte, data ...byte) []byte {
    m := len(slice)
    n := m + len(data)
    if n > cap(slice) { // if necessary, reallocate
        // allocate double what's needed, for future growth.
        newSlice := make([]byte, (n+1)*2)
				//fmt.Printf("Allocated new slice: %v, len=%d, cap=%d\n", newSlice, len(newSlice), cap(newSlice))
        copy(newSlice, slice)
        slice = newSlice
    }
    slice = slice[0:n]
    copy(slice[m:n], data)
    return slice
}

// This can only be run on a clean (zero-entry) table. Open the database in 
// another terminal and clear the testing table if necessary  
// (i.e., rodinia=# DELETE FROM ACTORS)
func TestAdd(ctx context.Context, s *satellite.Satellite) {

  ht := crypto.HashType("sha1")
  b := []byte("Begin")

  numAdds := 30
  for i := 1; i <= numAdds; i++ {
    var digest []byte
    if i == 1 {
      digest, _ = ht.Sum(b)
    } else {
			// Create new hash digest by appending a dummy string
      newByteSlice := []byte(strconv.Itoa(i))
			appendedByteSlice := AppendByte(digest, newByteSlice...)
      digest, _ = ht.Sum(appendedByteSlice)
		}
    s.Add(ctx, pb.Node{Id:string(hex.EncodeToString(digest)), Addrs:fmt.Sprintf("/some/random/addr%d", i)})
	}

	fmt.Printf("Finished adding %d records\n", numAdds)
}

// This can only be executed on a populated table (i.e., after `TestAdd()`)
func TestFind(ctx context.Context, s *satellite.Satellite, id string) {
	q := pb.Query{Id:id}
	res, err := s.Find(ctx, q)
	if err != nil {
		log.Fatal(err)
	} else if res.Id == "" {
    fmt.Printf("ID not found: %s\n", id)
  } else {
    fmt.Printf("ID found: %+v\n", res)
	}
}

func TestRandom(ctx context.Context, s *satellite.Satellite) {
	res, err := s.Random(ctx, pb.Query{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Random node: %v\n", res)
}

func main() {
  flag.Parse()

  opts := &satellite.Opts{*port, *keysDir, *dbUrl}

  s, err := satellite.New(opts)
  if err != nil {
    log.Fatal(err)
  }

	ctx := context.Background()
	TestAdd(ctx, s)

	// In the table
	TestFind(ctx, s, "4d134bc072212ace2df385dae143139da74ec0ef")
	// Not in the table
	TestFind(ctx, s, "4d1343c072212ace2df385dae143139da74ec3ef")
	TestRandom(ctx, s)
	TestRandom(ctx, s)
}
