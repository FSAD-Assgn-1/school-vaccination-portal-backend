package response

import (
	"fmt"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"
	"time"
)

type StudentManagementResponseHandler interface {
	ProcessResponse(req interface{}, data interface{}) StudentManagementConsolidatedResposne
}
type StudentManagementResponse struct {
	Id        int         `json:"id,omitempty"`
	Name      string      `json:"name"`
	Class     string      `json:"class"`
	Gender    string      `json:"gender"`
	RollNo    string      `json:"roll_no"`
	CreatedAt *time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time  `json:"update_at,omitempty"`
	PhoneNo   string      `json:"phone_no"`
	Links     interface{} `json:"_links,omitempty"`
}

type StudentManagementConsolidatedResposne struct {
	Message string      `json:"message_string"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error_string,omitempty"`
	Total   int         `json:"total,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	Offset  int         `json:"offset,omitempty"`
}

type VaccineRecordResponse struct {
	Id        int         `json:"id,omitempty"`
	StudentId int         `json:"student_id"`
	DriveId   int         `json:"drive_id"`
	Links     interface{} `json:"_links,omitempty"`
}

func (r StudentManagementConsolidatedResposne) ProcessResponse(req interface{}, result interface{}) StudentManagementConsolidatedResposne {
	resp := StudentManagementConsolidatedResposne{}
	switch req.(type) {
	case *requests.StudentManagementCreateRequest:
		data := StudentManagementResponse{}
		data.Id = result.([]models.DBInsertionRecord)[0].Record.Id
		data.Name = result.([]models.DBInsertionRecord)[0].Record.Name
		data.Class = result.([]models.DBInsertionRecord)[0].Record.Class
		data.Gender = result.([]models.DBInsertionRecord)[0].Record.Gender
		data.RollNo = result.([]models.DBInsertionRecord)[0].Record.RollNumber
		data.CreatedAt = result.([]models.DBInsertionRecord)[0].Record.CreatedAt
		data.UpdatedAt = result.([]models.DBInsertionRecord)[0].Record.UpdatedAt
		data.PhoneNo = result.([]models.DBInsertionRecord)[0].Record.PhoneNo
		if result.([]models.DBInsertionRecord)[0].Status {
			data.Links = geneRateHateOasForStudent(result.([]models.DBInsertionRecord)[0])
			resp.Message = "Student Onboarding succesful"
		} else {
			resp.Error = result.([]models.DBInsertionRecord)[0].ErrorReason
			resp.Message = "Student onboarding failed"
		}
		resp.Data = data
	case *requests.StudentManagementUpdateRequest:
		data := StudentManagementResponse{}
		data.Id = result.(models.StudentManagement).Id
		data.Name = result.(models.StudentManagement).Name
		data.Class = result.(models.StudentManagement).Class
		data.Gender = result.(models.StudentManagement).Gender
		data.RollNo = result.(models.StudentManagement).RollNumber
		data.CreatedAt = result.(models.StudentManagement).CreatedAt
		data.UpdatedAt = result.(models.StudentManagement).UpdatedAt
		data.PhoneNo = result.(models.StudentManagement).PhoneNo
		data.Links = geneRateHateOasForStudent(models.DBInsertionRecord{Record: result.(models.StudentManagement)})
		resp.Message = "Student Successfully Updated"
		resp.Data = data
	case *requests.StudentVaccinationRecordCreateRequest:
		v := VaccineRecordResponse{
			Id:        result.([]models.VaccineInsertionDBRecord)[0].Record.Id,
			StudentId: result.([]models.VaccineInsertionDBRecord)[0].Record.StudentId,
			DriveId:   result.([]models.VaccineInsertionDBRecord)[0].Record.DriveId,
		}
		if result.([]models.VaccineInsertionDBRecord)[0].Status {
			resp.Message = "Vaccination Record added Successfully"
		} else {
			resp.Message = "Vaccination Record addition failed"
			resp.Error = result.([]models.VaccineInsertionDBRecord)[0].ErrorReason
		}
		resp.Data = v
	case *requests.GetStudentVaccinationRecordRequest:
		resp.Message = "student record fetched successfully"
		resp.Data = result
		resp.Limit = req.(*requests.GetStudentVaccinationRecordRequest).Pagination.Limit
		resp.Offset = req.(*requests.GetStudentVaccinationRecordRequest).Pagination.Offset

	}

	return resp
}
func geneRateHateOasForStudent(data models.DBInsertionRecord) interface{} {
	hateOas := map[string]interface{}{}
	hateOas["self"] = map[string]string{
		"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/student-management/students/%d", data.Record.Id),
		"method": "GET",
	}
	hateOas["edit"] = map[string]string{
		"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/student-management/students/%d", data.Record.Id),
		"method": "PATCH",
	}

	return hateOas
}

func NewStudentManagementResponseHandler() StudentManagementResponseHandler {
	return StudentManagementConsolidatedResposne{}
}
