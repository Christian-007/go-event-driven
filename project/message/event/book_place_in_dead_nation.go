package event

import (
	"context"
	"fmt"
	"tickets/entities"

	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func (h Handler) BookPlaceInDeadNation(ctx context.Context, event *entities.BookingMade) error {
	log.FromContext(ctx).Info("Booking tickets in Dead Nation")

	show, err := h.showRepository.GetOne(ctx, event.ShowId)
	if err != nil {
		return fmt.Errorf("failed to get show: %w", err)
	}

	return h.deadNationAPI.BookInDeadNation(ctx, entities.DeadNationBooking{
		BookingID:         event.BookingID,
		DeadNationEventID: show.DeadNationID,
		NumberOfTickets:   event.NumberOfTickets,
		CustomerEmail:     event.CustomerEmail,
	})
}
