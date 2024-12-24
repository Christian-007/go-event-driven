package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/spreadsheets"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketStatus struct {
	TicketID      string `json:"ticket_id"`
	Status        string `json:"status"`
	Price         Money  `json:"price"`
	CustomerEmail string `json:"customer_email"`
}

type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type IssueReceiptPayload struct {
	TicketID string `json:"ticket_id"`
	Price    Money  `json:"price"`
}

type AppendToTrackerPayload struct {
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type IssueReceiptRequest struct {
	TicketID string
	Price    Money
}

type Header struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
}

func NewHeader(eventName string) Header {
	return Header{
		ID:          uuid.NewString(),
		PublishedAt: time.Now().UTC(),
	}
}

type TicketBookingConfirmed struct {
	Header        Header `json:"header"`
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketBookingCanceled struct {
	Header        Header `json:"header"`
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

func main() {
	log.Init(logrus.InfoLevel)

	watermillLogger := watermill.NewStdLogger(false, false)

	clients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), 
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	receiptsClient := NewReceiptsClient(clients)
	spreadsheetsClient := NewSpreadsheetsClient(clients)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating publisher:", err)
		panic(err)
	}

	issueReceiptSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "issue-receipt",
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating receiptSub:", err)
		panic(err)
	}

	appendToTrackerSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "append-to-tracker",
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating spreadSheetSub:", err)
		panic(err)
	}

	cancelTicketSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "cancel-ticket",
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating cancelTicketSub:", err)
		panic(err)
	}

	e := commonHTTP.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/tickets-status", func(c echo.Context) error {
		var request TicketsStatusRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			if ticket.Status == "canceled" {
				canceledEvent := TicketBookingCanceled{
					Header:        NewHeader(TicketBookingConfirmedTopic),
					TicketID:      ticket.TicketID,
					CustomerEmail: ticket.CustomerEmail,
					Price:         ticket.Price,
				}

				payload, err := json.Marshal(canceledEvent)
				if err != nil {
					fmt.Println("Error marshaling canceledEvent:", err)
					return err
				}

				msg := message.NewMessage(watermill.NewUUID(), payload)
				msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))

				err = publisher.Publish(TicketBookingCanceledTopic, msg)
				if err != nil {
					fmt.Println("Error publishing canceledEventMsg:", err)
					return err
				}
			} else if ticket.Status == "confirmed" {
				event := TicketBookingConfirmed{
					Header:        NewHeader(TicketBookingConfirmedTopic),
					TicketID:      ticket.TicketID,
					CustomerEmail: ticket.CustomerEmail,
					Price:         ticket.Price,
				}

				payload, err := json.Marshal(event)
				if err != nil {
					fmt.Println("Error marshaling ticketBookingConfirmed:", err)
					return err
				}

				msg := message.NewMessage(watermill.NewUUID(), payload)
				msg.Metadata.Set("correlation_id", c.Request().Header.Get("Correlation-ID"))

				err = publisher.Publish(TicketBookingConfirmedTopic, msg)
				if err != nil {
					fmt.Println("Error publishing ticketBookingConfirmedMsg:", err)
					return err
				}
			} else {
				return fmt.Errorf("unknown ticket status: %s", ticket.Status)
			}
		}

		return c.NoContent(http.StatusOK)
	})

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddMiddleware(PropagateCorrelationIMiddleware)
	router.AddMiddleware(PubSubLoggingMiddleware)

	router.AddNoPublisherHandler(
		"issue_receipt",
		TicketBookingConfirmedTopic,
		issueReceiptSub,
		func(message *message.Message) error {
			var event TicketBookingConfirmed
			err := json.Unmarshal(message.Payload, &event)
			if err != nil {
				return err
			}

			return receiptsClient.IssueReceipt(message.Context(), IssueReceiptRequest{
				TicketID: event.TicketID,
				Price: Money{
					Amount:   event.Price.Amount,
					Currency: event.Price.Currency,
				},
			})
		},
	)

	router.AddNoPublisherHandler(
		"print_ticket",
		TicketBookingConfirmedTopic,
		appendToTrackerSub,
		func(message *message.Message) error {
			var event TicketBookingConfirmed
			err := json.Unmarshal(message.Payload, &event)
			if err != nil {
				return err
			}

			return spreadsheetsClient.AppendRow(
				message.Context(),
				"tickets-to-print",
				[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
			)
		},
	)

	router.AddNoPublisherHandler(
		"tickets_to_refund",
		TicketBookingCanceledTopic,
		cancelTicketSub,
		func(message *message.Message) error {
			var event TicketBookingCanceled
			err := json.Unmarshal(message.Payload, &event)
			if err != nil {
				return err
			}

			return spreadsheetsClient.AppendRow(
				message.Context(),
				"tickets-to-refund",
				[]string{event.TicketID, event.CustomerEmail, event.Price.Amount, event.Price.Currency},
			)
		},
	)

	logrus.Info("Server starting...")

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return router.Run(ctx)
	})

	errGroup.Go(func() error {
		// we don't want to start HTTP server before Watermill router (so service won't be healthy before it's ready)
		<-router.Running() // wait till router is running

		err = e.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	errGroup.Go(func() error {
		// Shut down the HTTP server
		<-ctx.Done()
		return e.Shutdown(context.Background())
	})

	// Will block until all goroutines finish
	if err = errGroup.Wait(); err != nil {
		panic(err)
	}
}

func PubSubLoggingMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		logger := log.FromContext(msg.Context())
		logger = logger.WithField("message_uuid", (*msg).UUID)

		logger.Info("Handling a message")
		
		return next(msg)
	}
}

func PropagateCorrelationIMiddleware(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		ctx := msg.Context()

		reqCorrelationID := msg.Metadata.Get("correlation_id")
		if reqCorrelationID == "" {
			reqCorrelationID = shortuuid.New()
		}

		ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": reqCorrelationID}))
		ctx = log.ContextWithCorrelationID(ctx, reqCorrelationID)

		msg.SetContext(ctx)

		return next(msg)
	}
}

type ReceiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) ReceiptsClient {
	return ReceiptsClient{
		clients: clients,
	}
}

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request IssueReceiptRequest) error {
	body := receipts.PutReceiptsJSONRequestBody{
		TicketId: request.TicketID,
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
	}

	receiptsResp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, body)
	if err != nil {
		return err
	}
	if receiptsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", receiptsResp.StatusCode())
	}

	return nil
}

type SpreadsheetsClient struct {
	clients *clients.Clients
}

func NewSpreadsheetsClient(clients *clients.Clients) SpreadsheetsClient {
	return SpreadsheetsClient{
		clients: clients,
	}
}

func (c SpreadsheetsClient) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	request := spreadsheets.PostSheetsSheetRowsJSONRequestBody{
		Columns: row,
	}

	sheetsResp, err := c.clients.Spreadsheets.PostSheetsSheetRowsWithResponse(ctx, spreadsheetName, request)
	if err != nil {
		return err
	}
	if sheetsResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", sheetsResp.StatusCode())
	}

	return nil
}
