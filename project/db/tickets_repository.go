package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type TicketsRepository struct {
	db *sqlx.DB
}

func NewTicketsRepository(db *sqlx.DB) TicketsRepository {
	if db == nil {
		panic("db is nil")
	}

	return TicketsRepository{db: db}
}

func (t TicketsRepository) Add(ctx context.Context, ticket entities.Ticket) error {
	_, err := t.db.NamedExecContext(
		ctx,
		`
		INSERT INTO 
    		tickets (ticket_id, price_amount, price_currency, customer_email) 
		VALUES 
		    (:ticket_id, :price.amount, :price.currency, :customer_email)
		ON CONFLICT DO NOTHING`,
		ticket,
	)
	if err != nil {
		return fmt.Errorf("could not save ticket: %w", err)
	}

	return nil
}

func (t TicketsRepository) Remove(ctx context.Context, ticketId string) error {
	_, err := t.db.ExecContext(
		ctx,
		`DELETE FROM tickets WHERE ticket_id = $1`,
		ticketId,
	)
	if err != nil {
		return fmt.Errorf("could not remove ticket: %w", err)
	}

	return nil
}

func (t TicketsRepository) GetAll(ctx context.Context) ([]entities.Ticket, error) {
	var tickets []entities.Ticket

	err := t.db.SelectContext(
		ctx,
		&tickets,
		`
			SELECT
				ticket_id,
				price_amount as "price.amount",
				price_currency as "price.currency",
				customer_email
			FROM
				tickets
		`,
	)

	if err != nil {
		return nil, err
	}

	return tickets, nil
}
