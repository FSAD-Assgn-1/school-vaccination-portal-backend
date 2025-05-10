package usecase

import (
	"errors"
	"fmt"
	"log"
	"os"
	"school_vaccination_portal/models"
	"school_vaccination_portal/repository"
	"school_vaccination_portal/requests"
	"strings"

	"github.com/xuri/excelize/v2"
)

type StudentManagementUsecaseHandler interface {
	CreateStudentRecords(records *[]models.StudentManagement) []models.DBInsertionRecord
	UpdateStudentRecord(records models.StudentManagement) (models.StudentManagement, error)
	GetVaccinationDashBoardData() (int, int, error)
	CreateVaccinationRecords(records *[]models.StudentVaccineRecord) []models.VaccineInsertionDBRecord
	GetStudentVaccinationRecords(request *requests.GetStudentVaccinationRecordRequest) (int, []models.GetStudentCompleteDetails, error)
	GenerateVaccinationReport(request *requests.GenerateReportRequest) (string, error)
}

type StudentManagementUsecase struct {
	studentManagementRepo        repository.StudentManagementRepositoryHandler
	studentVaccinationRecordRepo repository.StudentVaccinationRecordRepositoryHandler
	bulkFileJobsRepo             repository.BulkFileJobsRepositoryHandler
	vaccineInventoryRepo         repository.VaccineInventoryHandler
}

func (u *StudentManagementUsecase) UpdateStudentRecord(records models.StudentManagement) (models.StudentManagement, error) {
	var err error
	//verify if student records with given entry exists????
	studentData, err := u.studentManagementRepo.GetStudents(fmt.Sprintf("id = %d", records.Id))
	if err != nil {
		log.Println(fmt.Sprintf("observing error in getting student record with id %d", records.Id), err.Error())
		return records, err
	}
	if len(studentData) == 0 || studentData[0].Id != records.Id {
		return records, fmt.Errorf("student record with id %d is not available ====>>>>", records.Id)
	}
	if err = u.studentManagementRepo.UpdateStudents(records); err != nil {
		log.Println("record update failed")
		return records, err
	}
	studentData, err = u.studentManagementRepo.GetStudents(fmt.Sprintf("id = %d", records.Id))
	return studentData[0], err
}
func (u *StudentManagementUsecase) CreateStudentRecords(records *[]models.StudentManagement) []models.DBInsertionRecord {
	return u.studentManagementRepo.CreateStudentRecord(records)
}

func (u *StudentManagementUsecase) GetVaccinationDashBoardData() (int, int, error) {
	totalStudents, err := u.studentVaccinationRecordRepo.GetStudentVaccinationRecordCount("", "LEFT JOIN student_vaccination_records v ON s.id = v.student_id")
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return totalStudents, 0, err
	}
	vaccnatedStudents, err := u.studentVaccinationRecordRepo.GetStudentVaccinationRecordCount("", "INNER JOIN student_vaccination_records v ON s.id = v.student_id")
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
	}
	return totalStudents, vaccnatedStudents, err

}

func (v *StudentManagementUsecase) CreateVaccinationRecords(records *[]models.StudentVaccineRecord) []models.VaccineInsertionDBRecord {
	//verify if drive actually exists
	validRecords := new([]models.StudentVaccineRecord)
	inValidRecords := []models.VaccineInsertionDBRecord{}
	for _, j := range *records {
		//checking if drive exists
		driveData, err := v.verifyDriveExists(j.DriveId, "")
		log.Println("drive Data ", driveData, err == nil)
		if err != nil || len(driveData) == 0 {
			invalid := models.VaccineInsertionDBRecord{
				Record:      j,
				Status:      false,
				ErrorReason: fmt.Sprintf("no drive exists with drive_id : %d", j.DriveId),
			}
			log.Println("invalid due to wrong drive id", invalid)
			inValidRecords = append(inValidRecords, invalid)
			continue
		}
		//check if student is valid
		resp, _ := v.studentManagementRepo.GetStudents(fmt.Sprintf("id = %d", j.StudentId))
		if len(resp) != 1 {
			invalid := models.VaccineInsertionDBRecord{
				Record:      j,
				Status:      false,
				ErrorReason: fmt.Sprintf("no student exists with student_id : %d", j.StudentId),
			}
			log.Println("invalid due to wrong student id", invalid)
			inValidRecords = append(inValidRecords, invalid)
			continue
		}
		*validRecords = append(*validRecords, j)
	}
	//proceed for insertion
	log.Println("Valid Records", validRecords, "invalidRecords", inValidRecords)
	resp := v.studentVaccinationRecordRepo.CreateVaccinationRecord(validRecords)
	return append(resp, inValidRecords...)
}

func (v *StudentManagementUsecase) verifyDriveExists(id int, name string) ([]models.VaccineInventory, error) {
	var selectionString string
	if id != 0 {
		selectionString = fmt.Sprintf("id = %d", id)
	} else {
		selectionString = fmt.Sprintf("vaccine_name = '%s'", name)
	}

	return v.vaccineInventoryRepo.GetVaccineInventory(selectionString)

}

func (v *StudentManagementUsecase) GetStudentVaccinationRecords(request *requests.GetStudentVaccinationRecordRequest) (int, []models.GetStudentCompleteDetails, error) {
	var studentDetails []models.GetStudentCompleteDetails
	var vaccinationDetails []models.StudentVaccinationDetail
	var err error
	var total int
	driveRegister := make(map[int]models.VaccineInventory)

	joinCondtion := "LEFT JOIN student_vaccination_records v ON s.id = v.student_id"
	queryString := ""

	//if id is given
	if request.Id != 0 {
		queryString = fmt.Sprintf("s.id = '%d'", request.Id)
	}
	if request.RollNo != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.roll_number = '%s'", request.RollNo)
		} else {
			queryString += " AND " + fmt.Sprintf("s.roll_number = '%s'", request.RollNo)
		}
	}
	if request.Class != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.class = '%s'", request.Class)
		} else {
			queryString += " AND " + fmt.Sprintf("s.class = '%s'", request.Class)
		}
	}
	if request.Name != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.name LIKE '%%%s%%'", request.Name)
		} else {
			queryString += " AND " + fmt.Sprintf("s.name LIKE '%%%s%%'", request.Name)
		}
	}
	if request.VaccineName != "" {
		//get drive info or id by vaccine name
		// var driveInfo VaccineDriveResponse
		drive, err := v.verifyDriveExists(0, request.VaccineName)
		if err != nil {
			log.Println("error fetching vaccination record", err.Error())
			return total, studentDetails, err
		}
		for _, j := range drive {
			driveRegister[j.Id] = j
		}
		log.Printf("Drive info is as below %+v", driveRegister)
		if len(driveRegister) == 0 {
			return total, studentDetails, fmt.Errorf("no vaccination drive with vaccine : %s", request.VaccineName)
		}
		driveIds := []int{}
		for i, _ := range driveRegister {
			driveIds = append(driveIds, i)
		}
		placeholders := make([]string, len(driveIds))
		for i, id := range driveIds {
			placeholders[i] = fmt.Sprintf("%d", id)
		}
		inClause := strings.Join(placeholders, ", ")
		log.Printf("student_vaccination_records.drive_id IN (%s)", inClause)
		//get count
		if queryString == "" {
			queryString = fmt.Sprintf("v.drive_id IN (%s)", inClause)
		} else {
			queryString += " AND " + fmt.Sprintf("v.drive_id IN (%s)", inClause)

		}
		//all record scenario
	}
	log.Println("query string being used", queryString)
	total, err = v.studentVaccinationRecordRepo.GetStudentVaccinationRecordCount(queryString, joinCondtion)
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return total, studentDetails, err
	}
	vaccinationDetails, err = v.studentVaccinationRecordRepo.GetStudentVaccinationRecord(queryString, request.Pagination)
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return total, studentDetails, err
	}

	//genrate consolidated Response
	for _, j := range vaccinationDetails {
		studentDetail := models.GetStudentCompleteDetails{}
		studentDetail.Id = j.Id
		studentDetail.Name = j.Name
		studentDetail.Gender = j.Gender
		studentDetail.Class = j.Class
		studentDetail.RollNo = j.RollNumber
		studentDetail.PhoneNo = j.PhoneNo
		if j.DriveId == 0 {
			studentDetail.Vaccination = false
		} else {
			studentDetail.Vaccination = true
			drive, ok := driveRegister[j.Id]
			if !ok {
				var driveInfo []models.VaccineInventory
				driveInfo, err = v.verifyDriveExists(j.DriveId, "")
				if err != nil {
					log.Println("error fetching vaccination record", err.Error())
					continue
				}
				log.Printf("response from vaccine service %+v", driveInfo)
				driveRegister[j.Id] = driveInfo[0]
				drive = driveInfo[0]
				log.Printf(" drive is %+v drive info is %+v", drive, driveInfo)
			}
			log.Printf(" drive is %+v", drive)
			studentDetail.VaccineName = drive.VaccineName
			studentDetail.VaccineDate = drive.DriveDate.Format("2006-01-02")
		}
		studentDetails = append(studentDetails, studentDetail)
	}
	return total, studentDetails, err
}

func (v *StudentManagementUsecase) GenerateVaccinationReport(request *requests.GenerateReportRequest) (string, error) {
	var studentDetails []models.GetStudentCompleteDetails
	var vaccinationDetails []models.StudentVaccinationDetail
	var err error
	driveRegister := make(map[int]models.VaccineInventory)
	queryString := ""

	if request.VaccineName != "" {
		drive, err := v.verifyDriveExists(0, request.VaccineName)
		if err != nil {
			log.Println("error in getting vaccine name", err.Error())
			return "", err
		}
		if len(drive) == 0 {
			return "", fmt.Errorf("no data for vaccine name %s", request.VaccineName)
		}
		for _, j := range drive {
			driveRegister[j.Id] = j
		}
		log.Printf("Drive info is as below %+v", driveRegister)
		driveIds := []int{}
		for i, _ := range driveRegister {
			driveIds = append(driveIds, i)
		}
		placeholders := make([]string, len(driveIds))
		fmt.Println(driveIds)
		for i, id := range driveIds {
			placeholders[i] = fmt.Sprintf("%d", id)
		}
		inClause := strings.Join(placeholders, ", ")
		queryString = fmt.Sprintf("v.drive_id IN (%s)", inClause)
	}
	if request.Class != "" {
		query := fmt.Sprintf("s.class = '%s'", request.Class)
		if queryString != "" {
			queryString += " AND " + query
		} else {
			queryString = query
		}
	}
	log.Println("query being used", queryString)
	vaccinationDetails, err = v.studentVaccinationRecordRepo.GetStudentVaccinationRecord(queryString, requests.Pagination{})
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return "", err
	}
	log.Printf("details fetched %+v", vaccinationDetails)
	for _, j := range vaccinationDetails {
		studentDetail := models.GetStudentCompleteDetails{}
		studentDetail.Id = j.Id
		studentDetail.Name = j.Name
		studentDetail.Gender = j.Gender
		studentDetail.Class = j.Class
		studentDetail.RollNo = j.RollNumber
		studentDetail.PhoneNo = j.PhoneNo
		if j.DriveId == 0 {
			studentDetail.Vaccination = false
		} else {
			studentDetail.Vaccination = true
			drive, ok := driveRegister[j.Id]
			if !ok {
				var driveInfo []models.VaccineInventory
				driveInfo, err = v.verifyDriveExists(j.DriveId, "")
				if err != nil {
					log.Println("error fetching vaccination record", err.Error())
					continue
				}
				log.Printf("response from vaccine service %+v", driveInfo)
				driveRegister[j.Id] = driveInfo[0]
				drive = driveInfo[0]
				log.Printf(" drive is %+v drive info is %+v", drive, driveInfo)
			}
			log.Printf(" drive is %+v", drive)
			studentDetail.VaccineName = drive.VaccineName
			studentDetail.VaccineDate = drive.DriveDate.Format("2006-01-02")
		}
		studentDetails = append(studentDetails, studentDetail)
	}
	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Name", "Class", "Gender", "Roll Number", "Phone Number", "Vaccination Status", "Vaccine Name", "Vaccination Date"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	for i, student := range studentDetails {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), student.Name)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), student.Class)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), student.Gender)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), student.RollNo)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("E%d", rowNum), student.PhoneNo)
		if !student.Vaccination {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Non Vaccinated")
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Vaccinated")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("G%d", rowNum), student.VaccineName)
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("H%d", rowNum), student.VaccineDate)
		}
	}
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("Unable to save report File Locally", err.Error())
		return "", errors.New("Internal server Error")
	}
	uploadedReportFile, err := v.bulkFileJobsRepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", request.RequestId)
	if err != nil {
		log.Println("error in uploading file to minio", err.Error())
		return "", err
	}
	filePath := fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)

	return filePath, nil
}

func NewStudentManagementUsecaseHandler(studentRepo repository.StudentManagementRepositoryHandler, studentvaccinationrepo repository.StudentVaccinationRecordRepositoryHandler, bulkfileJobsRepo repository.BulkFileJobsRepositoryHandler, vaccineinventoryRepo repository.VaccineInventoryHandler) StudentManagementUsecaseHandler {
	return &StudentManagementUsecase{studentManagementRepo: studentRepo, studentVaccinationRecordRepo: studentvaccinationrepo, bulkFileJobsRepo: bulkfileJobsRepo, vaccineInventoryRepo: vaccineinventoryRepo}
}
