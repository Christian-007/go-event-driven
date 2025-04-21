package db_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	ticketsDb "tickets/db"
	"tickets/entities"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingsRepository_AddBooking_seats_limit(t *testing.T) {
	ctx := context.Background()
	db := GetDb()

	err := ticketsDb.InitializeDatabaseSchema(db)
	require.NoError(t, err)

	bookingsRepo := ticketsDb.NewBookingsRepository(db)
	showsRepo := ticketsDb.NewShowsRepository(db)

	t.Run("overbooking", func(t *testing.T) {
		showID := uuid.New()

		err := showsRepo.Add(ctx, entities.Show{
			ID: showID,
			DeadNationID: uuid.New(),
			NumberOfTickets: 2,
			StartTime: time.Now().Add(time.Hour),
			Title: "Example title",
			Venue: "Example venue",
		})
		require.NoError(t, err)

		err = bookingsRepo.Add(ctx, entities.Booking{
			ID: uuid.New(),
			ShowID: showID,
			NumberOfTickets: 2,
			CustomerEmail: "foo@bar.com",
		})
		require.NoError(t, err)

		err = bookingsRepo.Add(ctx, entities.Booking{
			ID: uuid.New(),
			ShowID: showID,
			NumberOfTickets: 2,
			CustomerEmail: "foo@bar.com",
		})
		requireNotEnoughSeatsError(t, err)
	})

	t.Run("parallel_overbooking", func(t *testing.T) {
		showID := uuid.New()

		workersCount := 50
		workersErr := make(chan error, workersCount)

		unlock := make(chan struct{})

		err := showsRepo.Add(ctx, entities.Show{
			ID: showID,
			DeadNationID: uuid.New(),
			NumberOfTickets: 2,
			StartTime: time.Now().Add(time.Hour),
			Title: "Example title",
			Venue: "Example venue",
		})
		require.NoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(workersCount)

		for i:=0; i < workersCount; i++ {
			go func()  {
				defer wg.Done()

				// we are synchronizing goroutines to make sure that chance of overbooking is as high as possible
				<-unlock
				err = bookingsRepo.Add(ctx, entities.Booking{
					ID: uuid.New(),
					ShowID: showID,
					NumberOfTickets: 2,
					CustomerEmail: "foo@bar.com",
				})
				workersErr <- err
			}()
		}
		close(unlock)

		wg.Wait()
		close(workersErr)

		failedWorkers := 0
		succededWorkers := 0
		errors := []error{}

		for err := range workersErr {
			if err != nil {
				failedWorkers++
				errors = append(errors, err)
			} else {
				succededWorkers++
			}
		}

		assert.Equal(t, 1, succededWorkers)
		assert.Equal(t, workersCount-1, failedWorkers)

		if succededWorkers == 0 {
			// all workers failed, let's print error
			for _, err := range errors {
				t.Log("error:", err)
			}
		}
	})
}

func requireNotEnoughSeatsError(t *testing.T, err error) {
	var echoErr *echo.HTTPError
	require.ErrorAs(t, err, &echoErr)

	require.Equal(t, http.StatusBadRequest, echoErr.Code)
	require.Equal(t, "not enough seats available", echoErr.Message)
}

