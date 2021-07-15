package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// NewSQLite3 returns an SQLite3 backed database pointer. This package is
// only used for unit testing.
func NewSQLite3(ctx context.Context, logger log.Logger) (*SQL, error) {

	var sqlite3 *sql.DB

	// pwd, _ := os.Getwd()
	// fmt.Printf("NewSQLite3() pwd: %s\n", pwd)

	// remove the old db file if it exists (SQLite3 saves one by default when exiting)
	// fmt.Println("--> Attempting to remove db file")
	if _, err := os.Stat("testdata/sqlite3-empty.sql"); err == nil || os.IsExist(err) { // happy path
		// fmt.Println("--> Removing testdata/network_next.db")
		err = os.Remove("testdata/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			// return nil, err
		}
		// fmt.Println("--> Removed testdata/network_next.db")
		sqlite3, err = sql.Open("sqlite3", "file:testdata/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
			return nil, err
		}
		// fmt.Println("--> opened testdata/network_next.db")
	} else if _, err := os.Stat("../../testdata/sqlite3-empty.sql"); err == nil || os.IsExist(err) { // unit test
		// fmt.Println("--> Removing ../../testdata/network_next.db")
		err = os.Remove("../../testdata/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			// return nil, err
		}
		// fmt.Println("--> Removed ../../testdata/network_next.db")
		sqlite3, err = sql.Open("sqlite3", "file:../../testdata/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
			return nil, err
		}
		// fmt.Println("--> opened ../../testdata/network_next.db")
	} else {
		fmt.Println("--> did not find db file?")
		os.Exit(0)
	}

	// db.Ping actually establishes the connection and validates the parameters
	err := sqlite3.Ping()
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error pinging db: %w", err)
		return nil, err
	}

	db := &SQL{
		Client: sqlite3,
		Logger: logger,
	}

	// populate the db with some data from dev
	file, err := ioutil.ReadFile("testdata/sqlite3-empty.sql") // happy path
	if err != nil {
		file, err = ioutil.ReadFile("../../testdata/sqlite3-empty.sql") // unit test
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error opening seed file: %w", err)
			return nil, err
		}
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err := db.Client.Exec(request)
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error executing seed file sql line: %v", err)
			return nil, err
		}
	}

	return db, nil
}

func NewSQLite3Staging(ctx context.Context, logger log.Logger) (*SQL, error) {
	var sqlite3 *sql.DB

	if _, err := os.Stat("/app/sqlite3-empty.sql"); err == nil || os.IsExist(err) {

		err = os.Remove("/app/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %v", err)
		}

		// Boiler plate SQL file exists, load it in
		sqlite3, err = sql.Open("sqlite3", "file:/app/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			return nil, fmt.Errorf("NewSQLite3Staging() error creating db connection: %v", err)
		}
	} else {
		return nil, fmt.Errorf("NewSQLite3Staging() could not find /app/sqlite3-empty.sql")
	}

	// db.Ping actually establishes the connection and validates the parameters
	err := sqlite3.Ping()
	if err != nil {
		err = fmt.Errorf("NewSQLite3Staging() error pinging db: %v", err)
		return nil, err
	}

	db := &SQL{
		Client: sqlite3,
		Logger: logger,
	}

	// populate the db with basic tables
	file, err := ioutil.ReadFile("/app/sqlite3-empty.sql") // happy path
	if err != nil {
		return nil, fmt.Errorf("NewSQLite3Staging() error reading from ")
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err := db.Client.Exec(request)
		if err != nil {
			err = fmt.Errorf("NewSQLite3Staging() error executing seed file sql line: %v", err)
			return nil, err
		}
	}

	return db, nil

}

// NewPostgreSQL returns an PostgreSQL backed database pointer
func NewPostgreSQL(
	ctx context.Context,
	logger log.Logger,
	pgHostIP string,
	pgUserName string,
	pgPassword string,
) (*SQL, error) {

	fmt.Println("Creating PostgreSQL Storer.")

	// -- port and db name are the same regardless of the environment
	// -- sslmode is a driver req, connection is internal/IP and encrypted
	pgsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		pgHostIP, 5432, pgUserName, pgPassword, "network_next")

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
		Client: pgsql,
		Logger: logger,
	}

	return db, nil
}
