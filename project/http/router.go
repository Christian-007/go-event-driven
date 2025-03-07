package http

import (
	"net/http"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(eventBus *cqrs.EventBus, spreadsheetsAPIClient SpreadsheetsAPI, ticketsRepository TicketsRepository, showsRepository ShowsRepository) *echo.Echo {
	e := libHttp.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	handler := Handler{
		spreadsheetsAPIClient: spreadsheetsAPIClient,
		eventBus:              eventBus,
		ticketsRepository:     ticketsRepository,
		showsRepository:       showsRepository,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetAllTickets)
	e.POST("/shows", handler.PostShows)

	return e
}
