package api

import (
	"context"
	"sync"
)

type SpreadsheetsAPIMock struct{
	lock sync.Mutex
	Rows map[string][][]string
}

func (c *SpreadsheetsAPIMock) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Rows == nil {
		c.Rows = make(map[string][][]string)
	}

	c.Rows[spreadsheetName] = append(c.Rows[spreadsheetName], row)
	return nil
}
