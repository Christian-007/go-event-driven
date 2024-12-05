package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/spreadsheets"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)


type TicketsStatusRequest struct {
	Tickets []TicketStatus `json:"tickets"`
}

type TicketStatus struct {
	TicketID		string	`json:"ticket_id"`
	Status 			string	`json:"status"`
	Price			Money	`json:"price"`
	CustomerEmail 	string	`json:"customer_email"`
}

type Money struct {
	Amount 		string `json:"amount"`
	Currency 	string `json:"currency"`
}

type IssueReceiptPayload struct {
	TicketID	string	`json:"ticket_id"`
	Price		Money	`json:"price"`
}

type AppendToTrackerPayload struct {
	TicketID		string	`json:"ticket_id"`
	CustomerEmail 	string	`json:"customer_email"`
	Price			Money	`json:"price"`
}

type IssueReceiptRequest struct {
	TicketID string
	Price    Money
}

func main() {
	log.Init(logrus.InfoLevel)
	
	watermillLogger := watermill.NewStdLogger(false, false)

	clients, err := clients.NewClients(os.Getenv("GATEWAY_ADDR"), nil)
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
		Client: rdb,
		ConsumerGroup: IssueReceiptTopic,
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating receiptSub:", err)
		panic(err)
	}

	appendToTrackerSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
		ConsumerGroup: AppendToTrackerTopic,
	}, watermillLogger)
	if err != nil {
		fmt.Println("Error creating spreadSheetSub:", err)
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
			issueReceiptPayload := IssueReceiptPayload{
				TicketID: ticket.TicketID,
				Price: ticket.Price,
			}
			issueReceiptJson, err := json.Marshal(issueReceiptPayload)
			if err != nil {
				fmt.Println("Error marshaling issueReceipt:", err)
				return err
			}

			issueReceiptMsg := message.NewMessage(watermill.NewUUID(), issueReceiptJson)
			err = publisher.Publish(IssueReceiptTopic, issueReceiptMsg)
			if err != nil {
				fmt.Println("Error publishing issueReceiptMsg:", err)
				return err
			}

			appendToTrackerPayload := AppendToTrackerPayload{
				TicketID: ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price: ticket.Price,
			}
			appendToTrackerJson, err := json.Marshal(appendToTrackerPayload)
			if err != nil {
				fmt.Println("Error marshaling appendToTracker:", err)
				return err
			}

			appendToTrackerMsg := message.NewMessage(watermill.NewUUID(), appendToTrackerJson)
			err = publisher.Publish(AppendToTrackerTopic, appendToTrackerMsg)
			if err != nil {
				fmt.Println("Error publishing appendToTrackerMsg:", err)
				return err
			}
		}

		return c.NoContent(http.StatusOK)
	})

	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddNoPublisherHandler(
		"issue_receipt",
		IssueReceiptTopic,
		issueReceiptSub,
		func(message *message.Message) error {
			var payload IssueReceiptPayload
			err := json.Unmarshal(message.Payload, &payload)
			if err != nil {
				return err
			}

			return receiptsClient.IssueReceipt(message.Context(), IssueReceiptRequest{
				TicketID: payload.TicketID,
				Price: Money{
					Amount: payload.Price.Amount,
					Currency: payload.Price.Currency,
				},
			})
		},
	)

	router.AddNoPublisherHandler(
		"print_ticket",
		AppendToTrackerTopic,
		appendToTrackerSub,
		func(message *message.Message) error {
			var payload AppendToTrackerPayload
			err := json.Unmarshal(message.Payload, &payload)
			if err != nil {
				return err
			}

			return spreadsheetsClient.AppendRow(
				message.Context(),
				"tickets-to-print",
				[]string{payload.TicketID, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency},
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
			MoneyAmount: request.Price.Amount,
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
