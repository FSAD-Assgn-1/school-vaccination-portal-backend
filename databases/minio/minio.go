package minio

import (
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetMinIOClient() (*minio.Client, error) {
	var err error
	minioClient, err := minio.New(fmt.Sprintf("%s:%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT")), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_USERNAME"), os.Getenv("MINIO_PASSWORD"), ""),
		Secure: false, // Set to true for HTTPS
		Region: os.Getenv("MINIO_REGION"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
		return minioClient, err
	} else {
		log.Println("MinIO connected successfully")
	}
	return minioClient, err
}
