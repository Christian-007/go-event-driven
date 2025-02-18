package http

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	spreadsheetsAPIClient SpreadsheetsAPI
	eventBus              *cqrs.EventBus
	ticketsRepository     TicketsRepository
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}

type TicketsRepository interface {
	GetAll(ctx context.Context) ([]entities.Ticket, error)
}
