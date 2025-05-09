package http

import (
	"errors"
	"net/http"
	"tickets/db"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type bookingRequest struct {
	ShowID          uuid.UUID `json:"show_id"`
	NumberOfTickets int       `json:"number_of_tickets"`
	CustomerEmail   string    `json:"customer_email"`
}

type bookingsResponse struct {
	BookingID uuid.UUID `json:"booking_id"`
}

func (h Handler) PostBookTickets(c echo.Context) error {
	var request bookingRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	if request.NumberOfTickets < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "number of tickets must be greater than 0")
	}

	booking := entities.Booking{
		ID:              uuid.New(),
		ShowID:          request.ShowID,
		NumberOfTickets: request.NumberOfTickets,
		CustomerEmail:   request.CustomerEmail,
	}
	err = h.bookingsRepository.Add(c.Request().Context(), booking);
	if err != nil {
		if errors.Is(err, db.ErrExceedingTicketLimit) {
			return echo.NewHTTPError(http.StatusBadRequest, "not enough seats available")
		}

		return err
	}

	return c.JSON(http.StatusCreated, bookingsResponse{BookingID: booking.ID})
}
