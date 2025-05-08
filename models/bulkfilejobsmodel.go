package models

import "time"

type BulkFileJobsModel struct {
	Id               int       `json:"Id"`
	FileName         string    `json:"file_name"`
	FilePath         string    `json:"file_path"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ProcessedRecords int       `json:"processed_records"`
	TotalRecords     int       `json:"total_records"`
	RequestId        string    `json:"request_id"`
	RequestType      string    `json:"request_Type"`
}
