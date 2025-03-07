package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func InitializeDatabaseSchema(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tickets (
			ticket_id UUID PRIMARY KEY,
			price_amount NUMERIC(10, 2) NOT NULL,
			price_currency CHAR(3) NOT NULL,
			customer_email VARCHAR(255) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS shows (
			id UUID PRIMARY KEY,
			dead_nation_id UUID NOT NULL,
			number_of_tickets INTEGER NOT NULL,
			start_time timestamptz NOT NULL,
			title VARCHAR(255) NOT NULL,
			venue VARCHAR(255) NOT NULL,

			UNIQUE (dead_nation_id)
		);
	`)
	if err != nil {
		return fmt.Errorf("could not initialize database schema: %w", err)
	}

	return nil
}
