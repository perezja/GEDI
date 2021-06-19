
package main

import (
    "fmt"
    "context"
    "github.com/jackc/pgx/v4"
    pb "github.com/gedilabs/services/common"
)

const (
  DB_STRING     = "postgresql://localhost:5432/rodinia?user=rodinia&password=rodinia"
)

func main() {
  db, err := pgx.Connect(context.Background(), DB_STRING)
    if err != nil {
      panic(err)
  }
  fmt.Printf("Total count: %d", checkDHTSize(db, context.Background()))

}

func checkDHTSize(db *pgx.Conn, ctx context.Context) (count int) {
  sql := "SELECT COUNT(*) as count from actors"

  rows, err := db.Query(ctx, sql); if err != nil {
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

func random(ctx context.Context) (*pb.Node, error){
  sql := "SELECT id, addr FROM actors ORDER BY RANDOM() LIMIT 1";
  rows, err := s.Db.Query(ctx, sql)
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  for rows.Next() {
    var id, addrs string
		switch err = rows.Scan(&id, &addrs); err {
		case nil:
			return &pb.Node{Id: id, Addrs: addrs}, nil
		default:
			panic(err)
		}
	}
	return &pb.Node{}, errors.New("Returned more than one row")
}

