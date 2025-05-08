package controller

import (
	"log"
	"net/http"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"
	"school_vaccination_portal/response"
	"school_vaccination_portal/usecase"

	"github.com/labstack/echo/v4"
)

type BulkFileJobsController interface{}

type BController struct {
	req requests.BulkFileJobRequestHandler
	uc  usecase.BulkFileJobUsecaseHandler
}

const (
	BULK_STUDENT_RECORD = "STUDENT_CREATION_REC"
	BULK_VACCINE_RECORD = "VACCINE_CREATION_REC"
)

func (v BController) CreateStudentRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.BulkFileJobRequest)
	model := new(models.BulkFileJobsModel)
	model.RequestType = BULK_STUDENT_RECORD
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	if err = v.uc.UploadBulkRequestFile(model); err != nil {
		log.Println("error in processing student record update Request")
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message_string": "Request Accepted! Please check after sometime",
		"data": map[string]string{
			"request_id": model.RequestId,
			"status":     model.Status,
		},
	})
}
func (v BController) CreateVaccinationRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.BulkFileJobRequest)
	model := new(models.BulkFileJobsModel)
	model.RequestType = BULK_VACCINE_RECORD
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	if err = v.uc.UploadBulkRequestFile(model); err != nil {
		log.Println("error in processing student record update Request")
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message_string": "Request Accepted! Please check after sometime",
		"data": map[string]string{
			"request_id": model.RequestId,
			"status":     model.Status,
		},
	})
}
func (v BController) GetBulkJobStatus(c echo.Context) error {
	var err error
	req := new(requests.GetBulkFileRequest)
	model := new(models.BulkFileJobsModel)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, resp, err := v.uc.GetBulkFileJobDetails(req.RequestId, req.Pagination)
	if err != nil {
		log.Println("error in getting bulkUpload details Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message_string": "bulk job details fetched successfully",
		"data":           resp,
		"limit":          req.Pagination.Limit,
		"offset":         req.Pagination.Offset,
		"total":          total,
	})
}

func NewBulkUploadController(e *echo.Echo, req requests.BulkFileJobRequestHandler, uc usecase.BulkFileJobUsecaseHandler) BulkFileJobsController {
	studentServiceController := BController{
		req: req,
		uc:  uc,
	}
	e.POST("school-vaccine-portal/bulk-upload/students", studentServiceController.CreateStudentRecordBulk)
	e.POST("school-vaccine-portal/bulk-upload/vaccine-records", studentServiceController.CreateVaccinationRecordBulk)
	e.GET("school-vaccine-portal/bulk-upload/:request_id", studentServiceController.GetBulkJobStatus)
	e.GET("school-vaccine-portal/bulk-upload", studentServiceController.GetBulkJobStatus)
	return e
}
