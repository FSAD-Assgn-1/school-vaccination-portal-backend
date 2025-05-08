package response

import (
	"fmt"
	"log"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"
	"school_vaccination_portal/utils/validator"
	"time"

	"github.com/labstack/echo/v4"
)

type VacinneInventoryResponseHandler interface {
	ProcessVaccineInventoryResponse(req interface{}, data interface{}) interface{}
	ProcessErrorResponse(err error) interface{}
}

type VaccineInventoryResponse struct {
	Message string      `json:"message_string"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error_string,omitempty"`
	Links   interface{} `json:"links,omitempty"`
}
type VaccineInventoryGetResponse struct {
	Id        int         `json:"id"`
	Vaccine   string      `json:"vaccine_name"`
	DriveDate string      `json:"drive_date"`
	Doses     int         `json:"doses"`
	Classes   string      `json:"classes"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	Links     interface{} `json:"_links,omitempty"`
}

func (r VaccineInventoryResponse) ProcessErrorResponse(err error) interface{} {
	resp := VaccineInventoryResponse{}
	switch v := err.(type) {
	case *validator.ValidationError:
		log.Println("error type", v)
		resp.Message = "Invalid Input"
		resp.Data = []string{}
		resp.Error = err.(*validator.ValidationError).Fields
	case *echo.HTTPError:
		log.Println("error type", v)
		if err.(*echo.HTTPError).Code == 415 {
			resp.Message = "Invalid Request"
			resp.Data = []string{}
			resp.Error = map[string]string{
				"error": "Unsupported Media Type. Please use application/json in request header Content-Type",
			}
		}

	default:
		log.Println("error type", v)
		resp.Message = "unable to schedule vaccination drive"
		resp.Data = []string{}
		resp.Error = err.Error()
	}
	return resp
}

func (resp VaccineInventoryResponse) ProcessVaccineInventoryResponse(req, data interface{}) interface{} {
	r := VaccineInventoryResponse{}
	switch req.(type) {
	case *requests.VaccineInventoryCreateRequest:
		r.Message = "Drive Scheduled Successfully"
		r.Data = data
		r.Links = map[string]interface{}{
			"self": map[string]string{
				"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/vaccine-inventory/drives/%d", data.(*models.VaccineInventory).Id),
				"method": "GET",
			},
			"edit": map[string]string{
				"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/vaccine-inventory/drives/%d", data.(*models.VaccineInventory).Id),
				"method": "PATCH",
			},
		}
	case *requests.GetVaccineInventoryRequest, *requests.VaccineInventoryUpdateRequest:
		if len(data.([]models.VaccineInventory)) < 1 {
			r.Message = "No Upcoming Drives in 30 days"
			r.Data = []string{}
		} else {
			r.Message = "Drive Information fetched Succesully"
			collectionData := []VaccineInventoryGetResponse{}
			for _, j := range data.([]models.VaccineInventory) {
				vaccineDriveResponse := VaccineInventoryGetResponse{}
				vaccineDriveResponse.Id = j.Id
				vaccineDriveResponse.Vaccine = j.VaccineName
				vaccineDriveResponse.DriveDate = j.DriveDate.Format("2006-01-02")
				vaccineDriveResponse.Doses = j.Doses
				vaccineDriveResponse.Classes = j.Classes
				vaccineDriveResponse.CreatedAt = j.CreatedAt.Format("2006-01-02 15:04:05")
				vaccineDriveResponse.UpdatedAt = j.UpdatedAt.Format("2006-01-02 15:04:05")
				vaccineDriveResponse.Links = geneRateHateOasForVaccination(j)
				collectionData = append(collectionData, vaccineDriveResponse)
			}
			r.Data = collectionData
			if len(collectionData) == 1 {
				r.Data = collectionData[0]
			}
		}

	}
	return r
}
func geneRateHateOasForVaccination(data models.VaccineInventory) interface{} {
	hateOas := map[string]interface{}{}
	hateOas["self"] = map[string]string{
		"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/vaccine-inventory/drives/%d", data.Id),
		"method": "GET",
	}
	if time.Until(data.DriveDate) > 0 {
		hateOas["edit"] = map[string]string{
			"href":   fmt.Sprintf("http://localhost:8080/school-vaccine-portal/vaccine-inventory/drives/%d", data.Id),
			"method": "PATCH",
		}
	}
	return hateOas
}

func NewVacinneInventoryResponseHandler() VacinneInventoryResponseHandler {
	return VaccineInventoryResponse{}
}
