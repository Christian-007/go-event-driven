package main

import (
	"context"
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

type TicketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
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
	e.POST("/tickets-confirmation", func(c echo.Context) error {
		var request TicketsConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		for _, ticket := range request.Tickets {
			message := message.NewMessage(watermill.NewUUID(), []byte(ticket))
			err := publisher.Publish(IssueReceiptTopic, message)
			if err != nil {
				fmt.Println("Error publishing issueReceiptMsg:", err)
				return err
			}

			err = publisher.Publish(AppendToTrackerTopic, message)
			if err != nil {
				fmt.Println("Error publishing appendToTrackerMsg:", err)
				return err
			}
		}

		return c.NoContent(http.StatusOK)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
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
			ticketID := string(message.Payload)
			err := receiptsClient.IssueReceipt(message.Context(), ticketID)
			if err != nil {
				logrus.WithError(err).Error("Failed to issue receipt")
				return err
			} 
			
			return nil
		},
	)

	router.AddNoPublisherHandler(
		"print_ticket",
		AppendToTrackerTopic,
		appendToTrackerSub,
		func(message *message.Message) error {
			ticketID := string(message.Payload)
			err := spreadsheetsClient.AppendRow(message.Context(), "tickets-to-print", []string{ticketID})
			if err != nil {
				logrus.WithError(err).Error("Failed to append to tracker")
				return err
			}
			
			return nil
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

func (c ReceiptsClient) IssueReceipt(ctx context.Context, ticketID string) error {
	body := receipts.PutReceiptsJSONRequestBody{
		TicketId: ticketID,
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
