package requests

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"school_vaccination_portal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BulkFileJobRequestHandler interface {
	Bind(c echo.Context, request interface{}, model *models.BulkFileJobsModel) error
}
type BulkFileJobRequest struct {
	FilePath string
}

type GetBulkFileRequest struct {
	RequestId  string `param:"request_id"`
	Pagination Pagination
}

func (b BulkFileJobRequest) Bind(c echo.Context, request interface{}, model *models.BulkFileJobsModel) error {
	switch request.(type) {
	case *BulkFileJobRequest:
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return errors.New("file not received")
		}
		src, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("invalid file %s", err.Error())
		}
		defer src.Close()
		tmpPath := filepath.Join(os.TempDir(), fileHeader.Filename)
		dst, err := os.Create(tmpPath)
		if err != nil {
			return fmt.Errorf("unable to create temp file %s", err.Error())
		}
		defer dst.Close()
		io.Copy(dst, src)
		request.(*BulkFileJobRequest).FilePath = dst.Name()
		model.RequestId = uuid.NewString()
		model.FileName = fileHeader.Filename
		model.FilePath = dst.Name()
		model.Status = "PENDING"
	case *GetBulkFileRequest:
		err := c.Bind(request)
		if err != nil {
			log.Printf("error in binding Get Bulk Upload Request")
			return err
		}
		request.(*GetBulkFileRequest).Pagination = GetPagination(request.(*GetBulkFileRequest).Pagination)
	}

	return nil
}
func NewBulkUploadRequestHandler() BulkFileJobRequestHandler {
	return BulkFileJobRequest{}
}
