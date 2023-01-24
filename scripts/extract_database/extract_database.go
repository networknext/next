package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {

	pgsql, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=gaffer dbname=network_next sslmode=disable")
	if err != nil {
		fmt.Printf("error: could not connect to postgres: %v\n", err)
		os.Exit(1)
	}

	err = pgsql.Ping()
	if err != nil {
		fmt.Printf("error: could not ping postgres: %v\n", err)
		os.Exit(1)
	}

	rows, err := pgsql.Query("SELECT id, display_name FROM relays")
	if err != nil {
        fmt.Printf("error: could not extract relays: %v\n", err)
        os.Exit(1)
    }

	defer rows.Close()

	fmt.Printf("successfully connected to postgres\n")

	fmt.Printf("\nrelays:\n")

	for rows.Next() {
        var id uint64
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            fmt.Printf("error: failed to scan relay row: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("%d: %s\n", id, name)
    }
}
