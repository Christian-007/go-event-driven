package event

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
)

type Handler struct {
	spreadsheetsService SpreadsheetsAPI
	receiptsService     ReceiptsService
	ticketsRepository   TicketsRepository
	fileService         FileAPI
	deadNationAPI       DeadNationAPI
	showRepository      ShowsRepository
	eventBus            *cqrs.EventBus
}

func NewHandler(
	spreadsheetsService SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketsRepository TicketsRepository,
	fileService FileAPI,
	deadNationAPI DeadNationAPI,
	showRepository ShowsRepository,
	eventBus *cqrs.EventBus,
) Handler {
	if spreadsheetsService == nil {
		panic("missing spreadsheetsService")
	}
	if receiptsService == nil {
		panic("missing receiptsService")
	}
	if ticketsRepository == nil {
		panic("missing ticketRepository")
	}
	if fileService == nil {
		panic("missing fileService")
	}
	if deadNationAPI == nil {
		panic("missing deadNationAPI")
	}
	if showRepository == nil {
		panic("missing showRepository")
	}
	if eventBus == nil {
		panic("missing eventBus")
	}

	return Handler{
		spreadsheetsService: spreadsheetsService,
		receiptsService:     receiptsService,
		ticketsRepository:   ticketsRepository,
		fileService:         fileService,
		deadNationAPI:       deadNationAPI,
		showRepository:      showRepository,
		eventBus:            eventBus,
	}
}

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type TicketsRepository interface {
	Add(ctx context.Context, ticket entities.Ticket) error
	Remove(ctx context.Context, ticketId string) error
}

type FileAPI interface {
	UploadFile(ctx context.Context, fileID string, fileContent string) error
}

type DeadNationAPI interface {
	BookInDeadNation(ctx context.Context, request entities.DeadNationBooking) error
}

type ShowsRepository interface {
	GetOne(ctx context.Context, showId uuid.UUID) (entities.Show, error)
}
