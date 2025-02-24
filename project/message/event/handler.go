package event

import (
	"context"
	"tickets/entities"
)

type Handler struct {
	spreadsheetsService SpreadsheetsAPI
	receiptsService     ReceiptsService
	ticketsRepository   TicketsRepository
	fileService FileAPI
}

func NewHandler(
	spreadsheetsService SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketsRepository TicketsRepository,
	fileService FileAPI,
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

	return Handler{
		spreadsheetsService: spreadsheetsService,
		receiptsService:     receiptsService,
		ticketsRepository:   ticketsRepository,
		fileService: fileService,
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
