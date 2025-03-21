package db

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/google/uuid"
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

func (s ShowsRepository) GetOne(ctx context.Context, showId uuid.UUID) (entities.Show, error) {
	var result entities.Show

	err := s.db.GetContext(
		ctx,
		&result,
		`
			SELECT
				id,
				dead_nation_id,
				number_of_tickets,
				start_time,
				title,
				venue
			FROM
				shows
			WHERE
				id = $1
		`,
		showId,
	)

	if err != nil {
		return entities.Show{}, err
	}

	return result, nil

}
