package api

import (
	"context"
)

type SpreadsheetsAPIMock struct{}

func (c SpreadsheetsAPIMock) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	return nil
}
