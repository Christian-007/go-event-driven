package db

import (
	"fmt"
	"tickets/message/outbox"

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
		CREATE TABLE IF NOT EXISTS bookings (
			id UUID PRIMARY KEY,
			show_id UUID NOT NULL,
			number_of_tickets INTEGER NOT NULL,
			customer_email VARCHAR(255) NOT NULL,
			FOREIGN KEY (show_id) REFERENCES shows(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("could not initialize database schema: %w", err)
	}

	err = outbox.InitializeSchema(db.DB)
	if err != nil {
		return fmt.Errorf("could not initialize outbox schema: %w", err)
	}

	return nil
}
