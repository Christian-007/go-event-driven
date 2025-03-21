package message

import (
	"fmt"
	"tickets/message/event"
	"tickets/message/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewWatermillRouter(postgresSubscriber message.Subscriber, publisher message.Publisher, eventProcessorConfig cqrs.EventProcessorConfig, eventHandler event.Handler, watermillLogger watermill.LoggerAdapter) *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	useMiddlewares(router, watermillLogger)

	outbox.AddForwarderHandler(postgresSubscriber, publisher, router, watermillLogger)

	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		eventProcessorConfig,
	)
	if err != nil {
		fmt.Println("Cannot create Event Processor:", err)
		panic(err)
	}

	ep.AddHandlers(
		cqrs.NewEventHandler(
			"AppendToTracker",
			eventHandler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"TicketRefundToSheet",
			eventHandler.TicketRefundToSheet,
		),
		cqrs.NewEventHandler(
			"IssueReceipt",
			eventHandler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"SaveTickets",
			eventHandler.StoreTickets,
		),
		cqrs.NewEventHandler(
			"CancelTickets",
			eventHandler.RemoveCanceledTicket,
		),
		cqrs.NewEventHandler(
			"PrintTicketHandler",
			eventHandler.PrintTickets,
		),
		cqrs.NewEventHandler(
			"BookPlaceInDeadNation",
			eventHandler.BookPlaceInDeadNation,
		),
	)

	return router
}
