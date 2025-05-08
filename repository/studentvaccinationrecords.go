package repository

import (
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"
)

type StudentVaccinationRecordRepositoryHandler interface {
	CreateVaccinationRecord(record *[]models.StudentVaccineRecord) []models.VaccineInsertionDBRecord
	GetStudentVaccinationRecord(selectionString string, pagination requests.Pagination) ([]models.StudentVaccinationDetail, error)
	GetStudentVaccinationRecordCount(selectionString, join string) (int, error)
}

type StudentVaccinationRecordReposiotry struct {
	DB *mysql.MysqlConnect
}

func (r *StudentVaccinationRecordReposiotry) CreateVaccinationRecord(record *[]models.StudentVaccineRecord) []models.VaccineInsertionDBRecord {
	dataRecords := []models.VaccineInsertionDBRecord{}
	for _, j := range *record {
		dataRecord := models.VaccineInsertionDBRecord{}
		err := r.DB.Table("student_vaccination_records").Create(&j).Error
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
func (r *StudentVaccinationRecordReposiotry) GetStudentVaccinationRecord(selectionString string, pagination requests.Pagination) ([]models.StudentVaccinationDetail, error) {
	insertionDetails := []models.StudentVaccinationDetail{}
	if selectionString == "" {
		if pagination.Limit == 0 {
			return insertionDetails, r.DB.Table("student_management s").
				Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
				Joins("LEFT JOIN student_vaccination_records v ON s.id = v.student_id").
				Find(&insertionDetails).Error
		}
		return insertionDetails, r.DB.Table("student_management s").
			Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Order("id ASC").
			Joins("LEFT JOIN student_vaccination_records v ON s.id = v.student_id").
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&insertionDetails).Error
	}
	if pagination.Limit == 0 {
		return insertionDetails, r.DB.Table("student_management s").
			Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Joins("LEFT JOIN student_vaccination_records v ON s.id = v.student_id").
			Where(selectionString).
			Find(&insertionDetails).Error
	}
	return insertionDetails, r.DB.Table("student_management s").
		Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
		Order("id ASC").
		Joins("LEFT JOIN student_vaccination_records v ON s.id = v.student_id").
		Where(selectionString).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&insertionDetails).Error
}
func (r *StudentVaccinationRecordReposiotry) GetStudentVaccinationRecordCount(selectionString, join string) (int, error) {
	insertionDetails := 0
	if selectionString == "" {
		return insertionDetails, r.DB.Table("student_management s").
			Select("s.id AS id, s.name, s.class, s.roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Joins(join).
			Count(&insertionDetails).Error
	}
	return insertionDetails, r.DB.Table("student_management s").
		Select("s.id AS id, s.name, s.class, s.roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
		Joins(join).
		Where(selectionString).
		Count(&insertionDetails).Error
}
func NewVaccineRecordRepositoryHandler(DB *mysql.MysqlConnect) StudentVaccinationRecordRepositoryHandler {
	return &StudentVaccinationRecordReposiotry{
		DB: DB,
	}
}
