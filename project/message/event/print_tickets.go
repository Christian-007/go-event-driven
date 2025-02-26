package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h Handler) PrintTickets(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Info("Printing tickets")

	fileName := event.TicketID + "-ticket.html"
	ticketIDHtml := "<div>Ticket ID: " + event.TicketID + "</div>"
	priceHtml := "<div>Price: " + event.Price.Amount + event.Price.Currency + "</div>"
	htmlBody := ticketIDHtml + priceHtml
	
	err := h.fileService.UploadFile(ctx, fileName, htmlBody)
	if err != nil {
		return fmt.Errorf("failed to upload ticket file: %w", err)
	}

	err = h.eventBus.Publish(ctx, entities.TicketPrinted{
		Header: event.Header,
		TicketID: event.TicketID,
		FileName: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to publish TicketPrinted event: %w", err)
	}

	return nil
}
