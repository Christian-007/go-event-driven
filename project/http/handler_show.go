package http

import (
	"net/http"
	"tickets/entities"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type showsRequest struct {
	DeadNationID    uuid.UUID `json:"dead_nation_id" db:"dead_nation_id"`
	NumberOfTickets int       `json:"number_of_tickets" db:"number_of_tickets"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	Title           string    `json:"title" db:"title"`
	Venue           string    `json:"venue" db:"venue"`
}

type showsResponse struct {
	ShowID uuid.UUID `json:"show_id"`
}

func (h Handler) PostShows(c echo.Context) error {
	var request showsRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	show := entities.Show{
		ID:              uuid.New(),
		DeadNationID:    request.DeadNationID,
		NumberOfTickets: request.NumberOfTickets,
		StartTime:       request.StartTime,
		Title:           request.Title,
		Venue:           request.Venue,
	}
	if err = h.showsRepository.Add(c.Request().Context(), show); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, showsResponse{ShowID: show.ID})
}
