package server

import (
	"log"
	"school_vaccination_portal/controller"
	"school_vaccination_portal/databases/minio"
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/databases/rabbitmq"
	"school_vaccination_portal/repository"
	"school_vaccination_portal/requests"
	"school_vaccination_portal/response"
	"school_vaccination_portal/usecase"
	"school_vaccination_portal/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func newRouter() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.PATCH,
			echo.POST,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))
	e.Use(middleware.Logger())
	e.Validator = validator.NewValidator()
	dbConn, err := mysql.GetMySQLConnect()
	if err != nil {
		log.Fatalln("error in connecting to db", err.Error())
	}
	minIo, err := minio.GetMinIOClient()
	if err != nil {
		log.Fatalln("error creating min IO Client", err.Error())
	}
	rabb := rabbitmq.GetRabbitConn()

	vaccineDriveRequest := requests.NewVaccineDriveRequestHandler()
	vaccineDriveReqpository := repository.NewVaccineInventoryHandler(dbConn)
	vaccineDriveUsecase := usecase.NewVaccineInventoryUsecaseHandler(vaccineDriveReqpository)
	vaccineResponse := response.NewVacinneInventoryResponseHandler()
	controller.NewVaccineInventoryServiceController(e, vaccineDriveRequest, vaccineDriveUsecase, vaccineResponse)

	studentmanagementRequest := requests.NewStudentManagementRequestHandler()
	bulkfilejobrepo := repository.NewBulkFileJobsRepositoryHandler(dbConn, minIo, rabb)
	studentManagementRepo := repository.NewStudentRepositoryHandler(dbConn)
	studentvaccinationrecordrepo := repository.NewVaccineRecordRepositoryHandler(dbConn)
	studentmanagementusecase := usecase.NewStudentManagementUsecaseHandler(studentManagementRepo, studentvaccinationrecordrepo, bulkfilejobrepo, vaccineDriveReqpository)
	studentmanagementresponse := response.NewStudentManagementResponseHandler()
	controller.NewStudentManagementServiceController(e, studentmanagementRequest, studentmanagementusecase, studentmanagementresponse)
	bulkjobsRequest := requests.NewBulkUploadRequestHandler()
	bulkjobUc := usecase.NewBulkFileJobUsecaseHandler(studentmanagementusecase, bulkfilejobrepo)
	controller.NewBulkUploadController(e, bulkjobsRequest, bulkjobUc)

	return e
}
