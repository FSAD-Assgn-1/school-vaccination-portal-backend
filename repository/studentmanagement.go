package repository

import (
	"fmt"
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/models"
)

type StudentManagementRepositoryHandler interface {
	CreateStudentRecord(record *[]models.StudentManagement) []models.DBInsertionRecord
	GetStudents(filter string) ([]models.StudentManagement, error)
	UpdateStudents(student models.StudentManagement) error
}
type StudentManagementRepository struct {
	DB *mysql.MysqlConnect
}

func (r *StudentManagementRepository) CreateStudentRecord(record *[]models.StudentManagement) []models.DBInsertionRecord {
	dataRecords := []models.DBInsertionRecord{}
	for _, j := range *record {
		dataRecord := models.DBInsertionRecord{}
		err := r.DB.Table("student_management").Create(&j).Error
		dataRecord.Record = j
		dataRecord.Status = true
		if err != nil {
			dataRecord.Status = false
			dataRecord.ErrorReason = err.Error()
		}
		dataRecords = append(dataRecords, dataRecord)
	}
	return dataRecords
}
func (r *StudentManagementRepository) GetStudents(selectionString string) ([]models.StudentManagement, error) {
	dbResponse := []models.StudentManagement{}
	var err error
	if selectionString == "" {
		err = r.DB.Table("student_management").Find(&dbResponse).Error
		return dbResponse, err
	} else {
		err = r.DB.Table("student_management").Where(selectionString).Find(&dbResponse).Error
		return dbResponse, err
	}
}
func (r *StudentManagementRepository) UpdateStudents(student models.StudentManagement) error {
	toupdate := map[string]interface{}{}

	if student.Name != "" {
		toupdate["name"] = student.Name
	}
	if student.RollNumber != "" {
		toupdate["roll_number"] = student.RollNumber
	}
	if student.Class != "" {
		toupdate["class"] = student.Class
	}
	if student.Gender != "" {
		toupdate["gender"] = student.Gender
	}
	if student.PhoneNo != "" {
		toupdate["phone_no"] = student.PhoneNo
	}
	selectionstring := fmt.Sprintf("id = %d", student.Id)
	return r.DB.Table("student_management").Where(selectionstring).Updates(toupdate).Error
}

func NewStudentRepositoryHandler(DB *mysql.MysqlConnect) StudentManagementRepositoryHandler {
	return &StudentManagementRepository{
		DB: DB,
	}
}
