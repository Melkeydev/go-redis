package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

var port = 3000

type Server struct {
	port int
	db   *sql.DB
}

func NewServer() *http.Server {
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = initTestTable(db)
	if err != nil {
		log.Fatalf("Could not initialize test table: %v", err)
	}

	// Insert a test row
	err = insertTestRow(db)
	if err != nil {
		log.Fatalf("Could not insert test row: %v", err)
	}

	NewServer := &Server{
		port: port,
		db:   db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func connectToDB() (*sql.DB, error) {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "postgres"
		dbname   = "postgres"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to database")
	return db, nil
}

func initTestTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS test_table (
		id SERIAL PRIMARY KEY,
		data TEXT NOT NULL
	)`
	_, err := db.Exec(query)
	return err
}

func insertTestRow(db *sql.DB) error {
	query := `INSERT INTO test_table (data) VALUES ('Test data')`
	_, err := db.Exec(query)
	return err
}

func (s *Server) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
