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

	rows, err := pgsql.Query("SELECT id, display_name, datacenter, public_ip, public_port, internal_ip, internal_port, ssh_ip, ssh_port, ssh_user, public_key_base64, private_key_base64, mrc, port_speed, max_sessions FROM relays")
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
        var datacenter uint64
        var public_ip string
        var public_port int
        var internal_ip string
        var internal_port int
        var ssh_ip string
        var ssh_port int
        var ssh_user string
        var public_key_base64 string
        var private_key_base64 string
        var mrc int
        var port_speed int
        var max_sessions int

        if err := rows.Scan(&id, &name, &datacenter, &public_ip, &public_port, &internal_ip, &internal_port, &ssh_ip, &ssh_port, &ssh_user, &public_key_base64, &private_key_base64, &mrc, &port_speed, &max_sessions); err != nil {
            fmt.Printf("error: failed to scan relay row: %v\n", err)
            os.Exit(1)
        }

        fmt.Printf("%d: %s, %d, %s, %d, %s, %d, %s, %d, %s, %s, %s, %d, %d, %d\n", id, name, datacenter, public_ip, public_port, internal_ip, internal_port, ssh_ip, ssh_port, ssh_user, public_key_base64, private_key_base64, mrc, port_speed, max_sessions)
    }
}
