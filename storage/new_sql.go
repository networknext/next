package storage

import (
	"database/sql"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/routing"
)

// NewSQLite3 returns an SQLite3 backed database pointer
func NewSQLite3(logger log.Logger) (*SQL, error) {
	db, err := sql.Open("sqlite3", "./network_next.sql")
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
		return nil, err
	}

	return &SQL{
		Client:             db,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		syncSequenceNumber: -1,
	}, nil

}

// NewPostgreSQL returns an PostgreSQL backed database pointer
func NewPostgreSQL(logger log.Logger) (*SQL, error) {

	// move sensitive stuff to env w/ GCP vars
	const (
		host     = ""
		port     = 5432
		user     = ""
		password = ""
		dbname   = ""
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error creating db connection: %w", err)
		return nil, err
	}

	// db.Ping actually establishes the connection and validates the parameters
	err = db.Ping()
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error pinging db: %w", err)
		return nil, err
	}

	return &SQL{
		Client:             db,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		syncSequenceNumber: -1,
	}, nil

}
