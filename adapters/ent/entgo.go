package ent

import (
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func SetupAndConnectDatabase(baseConnectionString string) (*Client, error) {

	db, err := sql.Open("pgx", baseConnectionString)
	if err != nil {
		log.Fatalf("[sql.Open]: %v", err)
		return nil, err
	}

	// Create the "pills" table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (user_id TEXT PRIMARY KEY, categories json)`)
	if err != nil {
		log.Fatalf("[db.Exec]: error executing the init query: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := NewClient(Driver(drv))

	log.Printf("\n[Info]: Database connection established")
	return client, err
}
