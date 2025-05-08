package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"school_vaccination_portal/controller"
	"school_vaccination_portal/databases/minio"
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/databases/rabbitmq"
	"school_vaccination_portal/models"
	"school_vaccination_portal/repository"
	"school_vaccination_portal/server"
	"school_vaccination_portal/usecase"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading env file", err.Error())
	}
}

func StartAsyncFileProcessing(queue string) {
	dbConnection, err := mysql.GetMySQLConnect()
	if err != nil {
		log.Println("error in connecting to db", err.Error())
		os.Exit(1)
	}
	minIo, err := minio.GetMinIOClient()
	if err != nil {
		log.Fatalln("error creating min IO Client", err.Error())
	}
	rabbitConnection := rabbitmq.GetRabbitConn()
	bulkfilejobrepo := repository.NewBulkFileJobsRepositoryHandler(dbConnection, minIo, rabbitConnection)
	studentManagementRepo := repository.NewStudentRepositoryHandler(dbConnection)
	studentvaccinationrecordrepo := repository.NewVaccineRecordRepositoryHandler(dbConnection)
	studentmanagementusecase := usecase.NewStudentManagementUsecaseHandler(studentManagementRepo, studentvaccinationrecordrepo, bulkfilejobrepo, nil)
	asyncfileprocessingUsecase := usecase.NewBulkFileJobUsecaseHandler(studentmanagementusecase, bulkfilejobrepo)
	rabbitConnection.Qos(10, 0, false)
	ch, err := rabbitConnection.Consume(queue, "", false, false, false, false, amqp.Table{})
	if err != nil {
		log.Fatalf("Unable to start Processing from queue %s", err.Error())
	}
	log.Println("Async Processor Started")
	forever := make(chan bool)
	for j := range ch {
		//UnMarshall and see what kind of data
		data := new(models.BulkFileJobsModel)
		if err = json.Unmarshal(j.Body, data); err != nil {
			log.Println("Unable to Unmarshall Data Packet for processing ", err.Error())
			j.Ack(false)
		}
		switch data.RequestType {
		case controller.BULK_STUDENT_RECORD:
			go asyncfileprocessingUsecase.ProcessBulkStudentRecord(data)
			j.Ack(false)
		case controller.BULK_VACCINE_RECORD:
			go asyncfileprocessingUsecase.ProcessBulkVaccineRecord(data)
			j.Ack(false)
		default:
			log.Println("Unknown Request Type")
		}
	}
	<-forever
}

func main() {
	service := flag.String("service", "", "Service Being Requested: server, bulkProcessor")
	flag.Parse()
	switch *service {
	case "schoool-vaccination-portal-server":
		server.Start()
	default:
		log.Println("Starting Bulk processor")
		StartAsyncFileProcessing("async-file-processing-queue")
	}
}
