package db

import (
	"context"
	"database/sql"
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
	tx, err := b.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
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

	availableSeats := 0
	err = tx.GetContext(
		ctx,
		&availableSeats,
		`
			SELECT
				number_of_tickets AS available_seats
			FROM
				shows
			WHERE
				id = $1
		`,
		booking.ShowID,
	)
	if err != nil {
		return fmt.Errorf("could not get available seats: %w", err)
	}

	alreadyBookedSeats := 0
	err = tx.GetContext(
		ctx,
		&alreadyBookedSeats,
		// COALESCE(value, fallback) -> if `number_of_tickets` is empty, SELECT will return 0
		// SUM(number_of_tickets) -> accumulates the total of `number_of_tickets` from the same `show_id`
		`
			SELECT
				coalesce(SUM(number_of_tickets), 0) AS already_booked_seats
			FROM
				bookings
			WHERE
				show_id = $1
	`,
	booking.ShowID)
	if err != nil {
		return fmt.Errorf("could not get already booked seats: %w", err)
	}

	if availableSeats - alreadyBookedSeats < booking.NumberOfTickets {
		return ErrExceedingTicketLimit
	}
		
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
