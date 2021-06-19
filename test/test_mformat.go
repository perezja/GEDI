package main

import(
  "fmt"
  mb "github.com/multiformats/go-multibase"
  mh "github.com/multiformats/go-multihash"
  "github.com/gedilabs/abstract/node"

  //"encoding/binary"
)

func nibToInt(b []byte){
  bitpos := [4]byte{2,4,8,16}
  for _, s := range bitpos {
    //bit := b[i] & s
    //_ = bit
    fmt.Printf("%08b\n", s)
    //fmt.Println(s)
  }
  //return(binary.BigEndian.Uint32(b))
}

func main() {

  // int: 3 
  //a := []byte{0,0,1,1}

  //nibToInt([]byte{13})
  a := []byte{4}
  fmt.Printf("8-bit: %08b, len=%d, cap=%d\n", a, len(a), cap(a))
  id := fmt.Sprintf("4-bit: %04b", a)
  fmt.Println(id)
  digest := a
  mbase, _ := mb.Encode(mb.Base2, digest)
  mhash, _ := mh.Encode(digest, mh.IDENTITY)
  enc, mbase_bytes, _ := mb.Decode(mbase)
  fmt.Printf("mhash: %v, type: %T\n", mhash, mhash)
  fmt.Printf("mbase: %v, type: %T\n", mbase, mbase)
  fmt.Printf("mbase(encoding): %d, digest: %d\n", enc, mbase_bytes)

}
