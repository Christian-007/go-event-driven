package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type BookingsRepository struct {
	db *sqlx.DB
}

func NewBookingsRepository(db *sqlx.DB) BookingsRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingsRepository{db: db}
}

func (b BookingsRepository) Add(ctx context.Context, booking entities.Booking) error {
	_, err := b.db.NamedExecContext(
		ctx,
		`
		INSERT INTO 
			bookings (id, show_id, number_of_tickets, customer_email)
		VALUES
			(:id, :show_id, :number_of_tickets, :customer_email)
		ON CONFLICT DO NOTHING`,
		booking,
	)
	if err != nil {
		return fmt.Errorf("could not save booking: %w", err)
	}

	return nil
}
