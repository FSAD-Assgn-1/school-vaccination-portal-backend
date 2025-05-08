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

type VaccineInventoryController interface{}
type VController struct {
	req  requests.VaccineInventoryRequestHandler
	uc   usecase.VaccineInventoryUsecaseHandler
	resp response.VacinneInventoryResponseHandler
}

func (v VController) GetVaccinationDriveDetails(c echo.Context) error {
	var err error
	req := new(requests.GetVaccineInventoryRequest)
	model := new(models.VaccineInventory)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	var data []models.VaccineInventory
	if data, err = v.uc.GetVaccineDriveDetails(model); err != nil {
		log.Println("error in creating drive", err.Error())
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessVaccineInventoryResponse(req, data))
}

func (v VController) CreateVaccinationDrive(c echo.Context) error {
	var err error
	req := new(requests.VaccineInventoryCreateRequest)
	model := new(models.VaccineInventory)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	if err = v.uc.CreatevaccineDrive(model); err != nil {
		log.Println("error in creating drive", err.Error())
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusCreated, v.resp.ProcessVaccineInventoryResponse(req, model))
}
func (v VController) EditVaccinationDrive(c echo.Context) error {
	var err error
	req := new(requests.VaccineInventoryUpdateRequest)
	model := new(models.VaccineInventory)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	if err = v.uc.EditVaccineDrive(req); err != nil {
		log.Println("error in creating drive", err.Error())
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	var data []models.VaccineInventory
	model.Id = req.Id
	if data, err = v.uc.GetVaccineDriveDetails(model); err != nil {
		log.Println("error in getting drive", err.Error())
		return c.JSON(http.StatusBadRequest, v.resp.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessVaccineInventoryResponse(req, data))
}

func NewVaccineInventoryServiceController(e *echo.Echo, req requests.VaccineInventoryRequestHandler, uc usecase.VaccineInventoryUsecaseHandler, resp response.VacinneInventoryResponseHandler) VaccineInventoryController {
	vaccineServiceController := VController{
		req:  req,
		uc:   uc,
		resp: resp,
	}
	e.POST("school-vaccine-portal/vaccine-inventory/drives", vaccineServiceController.CreateVaccinationDrive)
	e.GET("school-vaccine-portal/vaccine-inventory/drives", vaccineServiceController.GetVaccinationDriveDetails)
	e.GET("school-vaccine-portal/vaccine-inventory/drives/:id", vaccineServiceController.GetVaccinationDriveDetails)
	e.PATCH("school-vaccine-portal/vaccine-inventory/drives", vaccineServiceController.EditVaccinationDrive)
	return e
}
