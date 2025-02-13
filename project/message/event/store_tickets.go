package event

import (
	"context"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h Handler) StoreTickets(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Saving tickets to DB")

	ticket := entities.Ticket{
		TicketID: event.TicketID,
		Price: event.Price,
		CustomerEmail: event.CustomerEmail,
	}

	return h.ticketsRepository.Add(ctx, ticket)
}
