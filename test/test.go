package main
import (
	"fmt"
	"io/ioutil"
)

func main() {
	f, err := ioutil.ReadFile("/root/test.txt")
	if err != nil { panic(err) } else{
	fmt.Printf("File: %s", f) }
}
