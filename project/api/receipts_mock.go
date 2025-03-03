package api

import (
	"context"
	"sync"
	"tickets/entities"
	"time"
)

type ReceiptsServiceMock struct {
	mock sync.Mutex

	IssuedReceipts map[string]entities.IssueReceiptRequest
}

func (r *ReceiptsServiceMock) IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error) {
	r.mock.Lock()
	defer r.mock.Unlock()

	r.IssuedReceipts[request.IdempotencyKey] = request

	return entities.IssueReceiptResponse{
		ReceiptNumber: "123",
		IssuedAt:      time.Now(),
	}, nil
}
