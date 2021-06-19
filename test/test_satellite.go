package main

import (
  "fmt"
  "log"
  "flag"
  "github.com/gedilabs/abstract/satellite"
)

var(
  lisAddr = flag.String("ip", "0.0.0.0", "Server listening address")
  port = flag.Int("port", 12132, "Server port")
  keysDir = flag.String("ssh_dir", "", "SSH keys directory")
  dbUrl = flag.String("db_url", "postgresql://localhost:5432/rodinia?user=rodinia&password=rodinia", "The database connection URL (e.g., postgresql://<ip>:<port>/<db>?user=<user>)")
)

func main() {
  flag.Parse()
  if *keysDir == "" {
    log.Fatal("Missing argument 'ssh_dir'")
  }

  opts := &satellite.Opts{*port, *keysDir, *dbUrl}

  s, err := satellite.New(opts)
  if err != nil {
    log.Fatal(err)
  }


  fmt.Printf("New Satellite: %+v\n", s)
  s.ListenAndServe()
  select{}
}
