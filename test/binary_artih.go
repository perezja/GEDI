package main

import(
  "fmt"
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
  a := []byte{15}
  fmt.Println(len(a))
  fmt.Printf("%08b, len=%d\n", a, len(a))
  id := fmt.Sprintf("%04b",[]byte{13})
  fmt.Println(id)
  for _, i := range id {
    fmt.Println(i)
  }
}
