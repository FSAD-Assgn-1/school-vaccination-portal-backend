package models

import "time"

type StudentManagement struct {
	Id         int        `json:"id"`
	Name       string     `json:"name"`
	Class      string     `json:"class"`
	Gender     string     `json:"gender"`
	RollNumber string     `json:"roll_number"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"update_at"`
	PhoneNo    string     `json:"phone_no"`
}
type DBInsertionRecord struct {
	Record      StudentManagement `json:"record"`
	Status      bool              `json:"status"`
	ErrorReason string            `json:"error_reason"`
}

type GetStudentCompleteDetails struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Gender      string `json:"gender"`
	RollNo      string `json:"roll_no"`
	PhoneNo     string `json:"phone_no"`
	Vaccination bool   `json:"vaccination"`
	VaccineName string `json:"vaccine_name,omitempty"`
	VaccineDate string `json:"vaccine_date,omitempty"`
}

type StudentVaccineRecord struct {
	Id        int        `json:"id"`
	StudentId int        `json:"student_id"`
	DriveId   int        `json:"drive_id"`
	CreatedAt *time.Time `json:"created_at"`
}
type VaccineInsertionDBRecord struct {
	Record      StudentVaccineRecord `json:"record"`
	Status      bool                 `json:"status"`
	ErrorReason string               `json:"error_reason"`
}

type StudentVaccinationDetail struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Class      string `json:"class"`
	Gender     string `json:"gender"`
	RollNumber string `json:"roll_number"`
	PhoneNo    string `json:"phone_no"`
	DriveId    int    `json:"drive_id"`
}
