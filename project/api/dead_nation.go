package api

import (
	"context"
	"fmt"
	"net/http"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/dead_nation"
)

type DeadNationClient struct {
	clients *clients.Clients
}

func NewDeadNationClient(clients *clients.Clients) *DeadNationClient {
	if clients == nil {
		panic("NewDeadNationClient: clients is nil")
	}

	return &DeadNationClient{clients: clients}
}

func (d DeadNationClient) BookInDeadNation(ctx context.Context, request entities.DeadNationBooking) error {
	resp, err := d.clients.DeadNation.PostTicketBookingWithResponse(
		ctx,
		dead_nation.PostTicketBookingRequest{
			CustomerAddress: request.CustomerEmail,
			EventId:         request.DeadNationEventID,
			NumberOfTickets: request.NumberOfTickets,
			BookingId:       request.BookingID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to book ticket: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to book ticket: unexpected status code %d", resp.StatusCode())
	}

	return nil
}
