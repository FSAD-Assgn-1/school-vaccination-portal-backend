package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/databases/rabbitmq"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"

	"github.com/minio/minio-go/v7"
	"github.com/streadway/amqp"
)

type BulkFileJobsRepositoryHandler interface {
	UploadFileToMinio(filePath, bucketName, root, uniqueId string) (string, error)
	CreateFileUpload(model *models.BulkFileJobsModel) error
	SubmitToRabbitMQ(rmqData *models.BulkFileJobsModel, queueName string) error
	UpdateFileUpload(model *models.BulkFileJobsModel) error
	GetFileFromActiveServer(bucketName, fileLocation string) (string, error)
	GetBulkFileJobs(requestId string, pagination requests.Pagination) ([]models.BulkFileJobsModel, error)
	GetBulkFileJobCounts(requestId string, pagination requests.Pagination) (int, error)
}

type BulkFileJobsRepository struct {
	MinIoConn *minio.Client
	DB        *mysql.MysqlConnect
	Rabbit    *rabbitmq.RabbitChannel
}

func (b *BulkFileJobsRepository) UploadFileToMinio(filePath, bucketName, root, uniqueId string) (string, error) {
	uploadInfo, err := b.MinIoConn.FPutObject(context.Background(), bucketName, fmt.Sprintf("%s%s/%s", root, uniqueId, filepath.Base(filePath)), filePath, minio.PutObjectOptions{})
	os.Remove(filePath)
	return uploadInfo.Key, err
}
func (b *BulkFileJobsRepository) GetFileFromActiveServer(bucketName, fileLocation string) (string, error) {
	object, err := b.MinIoConn.GetObject(context.Background(), bucketName, fileLocation, minio.GetObjectOptions{})
	if err != nil {
		log.Println("Error in fetching object from server", err.Error())
		return "", err
	}
	defer object.Close()
	//create localFile
	tempFile, err := os.CreateTemp("", "school-vaccine-bulk-*")
	defer tempFile.Close()
	if err != nil {
		log.Println("error Creating temporary file for processing bulk request", err.Error())
		return "", err
	}
	_, err = io.Copy(tempFile, object)
	if err != nil {
		log.Println("error copying  temporary file for processing bulk request", err.Error())
		return "", err
	}
	return tempFile.Name(), nil
}
func (b *BulkFileJobsRepository) CreateFileUpload(model *models.BulkFileJobsModel) error {
	return b.DB.Table("bulk_file_jobs").Create(model).Error
}

func (b *BulkFileJobsRepository) SubmitToRabbitMQ(rmqData *models.BulkFileJobsModel, queueName string) error {
	body, _ := json.Marshal(rmqData)
	return b.Rabbit.Publish(
		"", "async-file-processing-queue", false, false, amqp.Publishing{
			DeliveryMode: 2,
			ContentType:  "text/plain",
			Body:         body,
		},
	)
}

func (b *BulkFileJobsRepository) UpdateFileUpload(model *models.BulkFileJobsModel) error {
	updates := map[string]interface{}{}
	if model.Status != "" {
		updates["status"] = model.Status
	}
	if model.FilePath != "" {
		updates["file_path"] = model.FilePath
	}
	if model.TotalRecords != 0 {
		updates["total_records"] = model.TotalRecords
	}
	if model.ProcessedRecords != 0 {
		updates["processed_records"] = model.ProcessedRecords
	}
	if model.ErrorMessage != "" {
		updates["error_message"] = model.ErrorMessage
	}
	return b.DB.Table("bulk_file_jobs").Updates(updates).Where("id = ?", model.Id).Error
}
func (b *BulkFileJobsRepository) GetBulkFileJobs(requestId string, pagination requests.Pagination) ([]models.BulkFileJobsModel, error) {
	result := []models.BulkFileJobsModel{}
	var err error
	if requestId == "" {
		log.Println("fetching bulkupload status requestes with no requestId")
		err = b.DB.Table("bulk_file_jobs").
			Order("id ASC").
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&result).Error
	} else {
		log.Println("fetching bulkupload status requestes with requestId", requestId)
		err = b.DB.Table("bulk_file_jobs").
			Order("id ASC").
			Where("request_id =?", requestId).
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&result).Error
	}
	return result, err
}

func (b *BulkFileJobsRepository) GetBulkFileJobCounts(requestId string, pagination requests.Pagination) (int, error) {
	var result int
	var err error
	if requestId == "" {
		log.Println("fetching bulkupload status request count with no requestId")
		err = b.DB.Table("bulk_file_jobs").
			Count(&result).Error
	} else {
		log.Println("fetching bulkupload status request count with no requestId")

		err = b.DB.Table("bulk_file_jobs").
			Where("request_id = ?", requestId).
			Count(&result).Error
	}
	return result, err
}

func NewBulkFileJobsRepositoryHandler(DB *mysql.MysqlConnect, MinIO *minio.Client, rabbit *rabbitmq.RabbitChannel) BulkFileJobsRepositoryHandler {
	return &BulkFileJobsRepository{
		DB:        DB,
		MinIoConn: MinIO,
		Rabbit:    rabbit,
	}
}
