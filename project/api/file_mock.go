package api

import "context"

type FileServiceMock struct {}

func (f FileServiceMock) UploadFile(ctx context.Context, fileID string, fileContent string) error {
	return nil
}
