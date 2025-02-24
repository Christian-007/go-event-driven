package db_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	ticketsDb "tickets/db"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

var db *sqlx.DB
var getDbOnce sync.Once

func GetDb() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		db, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			fmt.Println("Error:", err)
			panic(err)
		}

	})
	return db
}

func TestTicketRepository_Add_Idempotency(t *testing.T) {
	ctx := context.Background()
	sqlxDb := GetDb()

	err := ticketsDb.InitializeDatabaseSchema(sqlxDb)
	require.NoError(t, err)

	repo := ticketsDb.NewTicketsRepository(sqlxDb)

	ticketToAdd := entities.Ticket{
		TicketID: uuid.NewString(),
		Price: entities.Money{
			Amount:   "50.30",
			Currency: "GBP",
		},
		CustomerEmail: "customer@gm.com",
	}

	for i := 0; i < 2; i++ {
		err := repo.Add(ctx, ticketToAdd)
		require.NoError(t, err)

		tickets, err := repo.GetAll(ctx)
		require.NoError(t, err)

		foundTickets := lo.Filter(tickets, func(t entities.Ticket, _ int) bool {
			return t.TicketID == ticketToAdd.TicketID
		})

		require.Len(t, foundTickets, 1)
	}

}
