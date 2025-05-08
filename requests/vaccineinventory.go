package requests

import (
	"log"
	"school_vaccination_portal/models"
	"time"

	"github.com/labstack/echo/v4"
)

type VaccineInventoryRequestHandler interface {
	Bind(c echo.Context, request interface{}, model *models.VaccineInventory) error
}

type VaccineInventoryRequest struct {
}

type GetVaccineInventoryRequest struct {
	Id     int    `param:"id"`
	Name   string `query:"vaccine_name"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

type VaccineInventoryCreateRequest struct {
	DriveDate   time.Time `json:"drive_date" validate:"checkValidDriveDate"`
	VaccineName string    `json:"vaccine_name" validate:"required"`
	Doses       int       `json:"doses" validate:"required"`
	Classes     string    `json:"classes" validate:"required"`
}

type VaccineInventoryUpdateRequest struct {
	Id          int        `json:"id" validate:"required"`
	DriveDate   *time.Time `json:"drive_date,omitempty"`
	VaccineName *string    `json:"vaccine_name,omitempty"`
	Doses       *int       `json:"doses,omitempty"`
	Classes     *string    `json:"classes,omitempty"`
}

func (r VaccineInventoryRequest) Bind(c echo.Context, req interface{}, model *models.VaccineInventory) error {
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
	case *VaccineInventoryCreateRequest:
		model.DriveDate = req.(*VaccineInventoryCreateRequest).DriveDate.UTC()
		model.VaccineName = req.(*VaccineInventoryCreateRequest).VaccineName
		model.Doses = req.(*VaccineInventoryCreateRequest).Doses
		model.Classes = req.(*VaccineInventoryCreateRequest).Classes
	case *GetVaccineInventoryRequest:
		model.Id = req.(*GetVaccineInventoryRequest).Id
		model.VaccineName = req.(*GetVaccineInventoryRequest).Name
		log.Println("req is ", model)
	case **VaccineInventoryUpdateRequest:
		model.Id = req.(*VaccineInventoryUpdateRequest).Id
	default:
		log.Println("request type Unknown for transformation", v)
	}

	return nil
}
func NewVaccineDriveRequestHandler() VaccineInventoryRequestHandler {
	return VaccineInventoryRequest{}
}
