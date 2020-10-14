package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	"github.com/networknext/backend/routing"
)

// NewSQLite3 returns an SQLite3 backed database pointer
func NewSQLite3(ctx context.Context, logger log.Logger) (*SQL, error) {
	sqlite3, err := sql.Open("sqlite3", "./network_next.sql")
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
		return nil, err
	}

	db := &SQL{
		Client:             sqlite3,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		SyncSequenceNumber: -1,
	}

	syncIntervalStr := os.Getenv("DB_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "DB_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		os.Exit(1)
	}
	// Start a goroutine to sync from Firestore
	go func() {

		ticker := time.NewTicker(syncInterval)
		db.SyncLoop(ctx, ticker.C)
	}()

	return db, nil
}

// NewPostgreSQL returns an PostgreSQL backed database pointer
func NewPostgreSQL(ctx context.Context, logger log.Logger) (*SQL, error) {

	// move sensitive stuff to env w/ GCP vars
	const (
		host     = "localhost"
		port     = 5432
		user     = "engineering"
		password = "0xdeadbeef"
		dbname   = "nn"
	)

	pgsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	pgsql, err := sql.Open("postgres", pgsqlInfo)
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error creating db connection: %w", err)
		return nil, err
	}

	// db.Ping actually establishes the connection and validates the parameters
	err = pgsql.Ping()
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error pinging db: %w", err)
		return nil, err
	}

	db := &SQL{
		Client:             pgsql,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		SyncSequenceNumber: -1,
	}

	syncIntervalStr := os.Getenv("DB_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "DB_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		os.Exit(1)
	}
	// Start a goroutine to sync from Firestore
	go func() {

		ticker := time.NewTicker(syncInterval)
		db.SyncLoop(ctx, ticker.C)
	}()

	return db, nil

}
