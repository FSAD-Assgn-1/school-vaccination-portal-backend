package response

import (
	"log"
	"school_vaccination_portal/utils/validator"

	"github.com/labstack/echo/v4"
)

func ProcessErrorResponse(err error) interface{} {
	resp := StudentManagementConsolidatedResposne{}
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
				"error": "content type mismatch. Please use application/json in request header Content-Type",
			}
		}
	default:
		log.Println("error type", v)
		resp.Message = "request processing unavailable"
		resp.Data = []string{}
		resp.Error = err.Error()
	}
	return resp
}
