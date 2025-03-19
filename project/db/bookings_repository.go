package db

import (
	"context"
	"errors"
	"fmt"
	"tickets/entities"
	"tickets/message/event"
	"tickets/message/outbox"

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

func (b BookingsRepository) Add(ctx context.Context, booking entities.Booking) (err error) {
	tx, err := b.db.Beginx()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			err = errors.Join(err, rollbackErr)
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.NamedExecContext(
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
		return fmt.Errorf("could not add booking: %w", err)
	}

	outboxPublisher, err := outbox.NewPublisherForDb(ctx, tx)
	if err != nil {
		return fmt.Errorf("could not create SQL publisher: %w", err)
	}

	bus, err := event.NewEventBus(outboxPublisher)
	if err != nil {
		return fmt.Errorf("could not create event bus: %w", err)
	}
	err = bus.Publish(ctx, entities.BookingMade{
		BookingID: booking.ID,
		NumberOfTickets: booking.NumberOfTickets,
		CustomerEmail: booking.CustomerEmail,
		ShowId: booking.ShowID,
	})
	if err != nil {
		return fmt.Errorf("could not publish event: %w", err)
	}

	return nil
}
