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

type StudentManagementController interface{}
type SController struct {
	req  requests.StudentManagementRequestHandler
	uc   usecase.StudentManagementUsecaseHandler
	resp response.StudentManagementResponseHandler
}

func (v SController) CreateStudentRecord(c echo.Context) error {
	var err error
	req := new(requests.StudentManagementCreateRequest)
	model := new([]models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp := v.uc.CreateStudentRecords(model)
	if resp[0].Status {
		return c.JSON(http.StatusCreated, v.resp.ProcessResponse(req, resp))
	}
	return c.JSON(http.StatusBadRequest, v.resp.ProcessResponse(req, resp))
}
func (v SController) CreateStudentRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.StudentManagementUpdateRequest)
	model := new(models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp, err := v.uc.UpdateStudentRecord(*model)
	if err != nil {
		log.Println("student record update failed", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}
func (v SController) EditStudentRecord(c echo.Context) error {
	var err error
	req := new(requests.StudentManagementUpdateRequest)
	model := new(models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp, err := v.uc.UpdateStudentRecord(*model)
	if err != nil {
		log.Println("student record update failed", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}
func (v SController) CreateVaccineRecord(c echo.Context) error {
	var err error
	req := new(requests.StudentVaccinationRecordCreateRequest)
	model := new([]models.StudentVaccineRecord)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding vaccine record is", model)
	resp := v.uc.CreateVaccinationRecords(model)
	if resp[0].Status {
		return c.JSON(http.StatusCreated, v.resp.ProcessResponse(req, resp))
	}
	return c.JSON(http.StatusBadRequest, v.resp.ProcessResponse(req, resp))
}
func (v SController) GetStudentVaccinationRecord(c echo.Context) error {
	var err error
	req := new(requests.GetStudentVaccinationRecordRequest)
	model := new(models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, resp, err := v.uc.GetStudentVaccinationRecords(req)
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
	}
	log.Printf("resp generated received is %+v and total records are %d", resp, total)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	finalResp := v.resp.ProcessResponse(req, resp)
	finalResp.Total = total
	return c.JSON(http.StatusOK, finalResp)
}

func (v SController) GenerateReport(c echo.Context) error {
	var err error
	req := new(requests.GenerateReportRequest)
	model := new(models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding generate report request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received for generating report is %+v", req)
	fileLoc, err := v.uc.GenerateVaccinationReport(req)
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]string{
		"file": fileLoc,
	})
}

func (v SController) GetVaccinationRecordDashBoard(c echo.Context) error {
	var err error
	req := new(requests.StudentManagementRequest)
	model := new(models.StudentManagement)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, vaccinated, err := v.uc.GetVaccinationDashBoardData()
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]int{
		"total_students":      total,
		"vaccinated_students": vaccinated,
	})
}

func NewStudentManagementServiceController(e *echo.Echo, req requests.StudentManagementRequestHandler, uc usecase.StudentManagementUsecaseHandler, resp response.StudentManagementResponseHandler) StudentManagementController {
	studentServiceController := SController{
		req:  req,
		uc:   uc,
		resp: resp,
	}
	e.POST("school-vaccine-portal/student-management/students/bulk-upload", studentServiceController.EditStudentRecord)
	e.POST("school-vaccine-portal/student-management/students", studentServiceController.CreateStudentRecord)
	e.PATCH("school-vaccine-portal/student-management/students", studentServiceController.EditStudentRecord)
	e.POST("school-vaccine-portal/student-management/vaccine-records", studentServiceController.CreateVaccineRecord)
	e.GET("school-vaccine-portal/student-management/vaccine-records/students/:id", studentServiceController.GetStudentVaccinationRecord)
	e.GET("school-vaccine-portal/student-management/vaccine-records/students", studentServiceController.GetStudentVaccinationRecord)
	e.GET("school-vaccine-portal/student-management/vaccine-records/dashboard", studentServiceController.GetVaccinationRecordDashBoard)
	e.GET("school-vaccine-portal/student-management/vaccine-records/genrate-report", studentServiceController.GenerateReport)
	return e
}
