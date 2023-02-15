package ent

import (
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func SetupAndConnectDatabase(baseConnectionString string, database string) (*Client, error) {

	db, err := sql.Open("pgx", baseConnectionString+"/"+database)
	if err != nil {
		log.Fatalf("[sql.Open]: %v", err)
		return nil, err
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := NewClient(Driver(drv))

	return client, err
}
