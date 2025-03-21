package http

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
)

type Handler struct {
	spreadsheetsAPIClient SpreadsheetsAPI
	eventBus              *cqrs.EventBus
	ticketsRepository     TicketsRepository
	showsRepository       ShowsRepository
	bookingsRepository    BookingsRepository
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type TicketsRepository interface {
	GetAll(ctx context.Context) ([]entities.Ticket, error)
}

type ShowsRepository interface {
	Add(ctx context.Context, show entities.Show) error
	GetOne(ctx context.Context, showId uuid.UUID) (entities.Show, error)
}

type BookingsRepository interface {
	Add(ctx context.Context, booking entities.Booking) error
}
