package http

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type Handler struct {
	spreadsheetsAPIClient SpreadsheetsAPI
	eventBus              *cqrs.EventBus
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, spreadsheetName string, row []string) error
}
