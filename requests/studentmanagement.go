package requests

import (
	"log"
	"school_vaccination_portal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type StudentManagementRequestHandler interface {
	Bind(c echo.Context, request interface{}, model interface{}) error
}
type StudentManagementRequest struct{}
type StudentManagementCreateRequest struct {
	Name    string `json:"name" validate:"required"`
	Class   string `json:"class" validate:"checkValidGrade"`
	Gender  string `json:"gender" validate:"required"`
	RollNo  string `json:"roll_no" validate:"required"`
	PhoneNo string `json:"phone_no" validate:"required"`
}
type StudentManagementUpdateRequest struct {
	Id      int    `json:"id" validate:"required"`
	Name    string `json:"name,omitempty"`
	Class   string `json:"class,omitempty" validate:"omitempty,checkValidGradeUpdate"`
	Gender  string `json:"gender,omitempty"`
	RollNo  string `json:"roll_no,omitempty"`
	PhoneNo string `json:"phone_no,omitempty"`
}
type StudentVaccinationRecordCreateRequest struct {
	StudentId int `json:"student_id" validate:"required"`
	DriveId   int `json:"drive_id" validate:"required"`
}
type GetStudentVaccinationRecordRequest struct {
	Id          int    `param:"id"`
	RollNo      string `query:"roll_no"`
	VaccineName string `query:"vaccine_name"`
	Class       string `query:"class" validate:"omitempty,checkValidGradeUpdate"`
	Name        string `query:"name"`
	Pagination  Pagination
}
type GenerateReportRequest struct {
	Class       string `query:"class" validate:"omitempty,checkValidGradeUpdate"`
	VaccineName string `query:"vaccine_name"`
	RequestId   string
}

func (r StudentManagementRequest) Bind(c echo.Context, req interface{}, model interface{}) error {
	var err error

	if err = c.Bind(req); err != nil {
		log.Println("Error in reading request", err.Error())
		return err
	}
	if err = c.Validate(req); err != nil {
		log.Println("error in validating request", err.Error())
		return err
	}
	switch v := req.(type) {
	case *StudentManagementCreateRequest:
		data := models.StudentManagement{}
		data.Name = req.(*StudentManagementCreateRequest).Name
		data.Class = req.(*StudentManagementCreateRequest).Class
		data.Gender = req.(*StudentManagementCreateRequest).Gender
		data.RollNumber = req.(*StudentManagementCreateRequest).RollNo
		data.PhoneNo = req.(*StudentManagementCreateRequest).PhoneNo
		modelptr := model.(*[]models.StudentManagement)
		*modelptr = append(*modelptr, data)
	case *StudentManagementUpdateRequest:
		model.(*models.StudentManagement).Id = req.(*StudentManagementUpdateRequest).Id
		model.(*models.StudentManagement).Name = req.(*StudentManagementUpdateRequest).Name
		model.(*models.StudentManagement).Class = req.(*StudentManagementUpdateRequest).Class
		model.(*models.StudentManagement).Gender = req.(*StudentManagementUpdateRequest).Gender
		model.(*models.StudentManagement).RollNumber = req.(*StudentManagementUpdateRequest).RollNo
		model.(*models.StudentManagement).PhoneNo = req.(*StudentManagementUpdateRequest).PhoneNo
	case *StudentVaccinationRecordCreateRequest:
		data := models.StudentVaccineRecord{}
		data.StudentId = req.(*StudentVaccinationRecordCreateRequest).StudentId
		data.DriveId = req.(*StudentVaccinationRecordCreateRequest).DriveId
		modelptr := model.(*[]models.StudentVaccineRecord)
		*modelptr = append(*modelptr, data)

	case *GetStudentVaccinationRecordRequest:
		req.(*GetStudentVaccinationRecordRequest).Pagination = GetPagination(req.(*GetStudentVaccinationRecordRequest).Pagination)
	case *GenerateReportRequest:
		req.(*GenerateReportRequest).RequestId = uuid.NewString()

	default:
		log.Println("request type Unknown for transformation", v)
	}

	return nil
}

func NewStudentManagementRequestHandler() StudentManagementRequestHandler {
	return StudentManagementRequest{}
}
