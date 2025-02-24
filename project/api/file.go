package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

type FileAPIClient struct {
	clients *clients.Clients
}

func NewFileAPIClient(clients *clients.Clients) *FileAPIClient {
	if clients == nil {
		panic("NewFileAPIClient: client is nil")
	}

	return &FileAPIClient{
		clients: clients,
	}
}

func (f FileAPIClient) UploadFile(ctx context.Context, fileID string, fileContent string) error {
	resp, err := f.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, fileContent)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	if resp.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).Infof("file %s already exists", fileID)
		return nil
	}

	if resp.StatusCode() == http.StatusCreated {
		return fmt.Errorf("unexpected status code while uploading file %s: %d", fileID, resp.StatusCode())
	}

	return nil
}
