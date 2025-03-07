package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/jmoiron/sqlx"
)

type ShowsRepository struct {
	db *sqlx.DB
}

func NewShowsRepository(db *sqlx.DB) ShowsRepository {
	if db == nil {
		panic("db is nil")
	}

	return ShowsRepository{db: db}
}

func (s ShowsRepository) Add(ctx context.Context, show entities.Show) error {
	_, err := s.db.NamedExecContext(
		ctx,
		`
		INSERT INTO
			shows (id, dead_nation_id, number_of_tickets, start_time, title, venue)
		VALUES
			(:id, :dead_nation_id, :number_of_tickets, :start_time, :title, :venue)
		ON CONFLICT DO NOTHING`,
		show,
	)
	if err != nil {
		return fmt.Errorf("could not save show: %w", err)
	}

	return nil
}
