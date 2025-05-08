package usecase

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"school_vaccination_portal/models"
	"school_vaccination_portal/repository"
	"school_vaccination_portal/requests"
	"school_vaccination_portal/utils/validator"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type BulkFileJobUsecaseHandler interface {
	UploadBulkRequestFile(req *models.BulkFileJobsModel) error
	ProcessBulkStudentRecord(model *models.BulkFileJobsModel) error
	ProcessBulkVaccineRecord(model *models.BulkFileJobsModel) error
	GetBulkFileJobDetails(requestId string, pagination requests.Pagination) (int, []models.BulkFileJobsModel, error)
}

type BulkFileJobUsecase struct {
	studentManagementusecaseRepo StudentManagementUsecaseHandler
	bulkFileJobsRepo             repository.BulkFileJobsRepositoryHandler
}

func (b *BulkFileJobUsecase) GetBulkFileJobDetails(requestId string, pagination requests.Pagination) (int, []models.BulkFileJobsModel, error) {
	var count int
	var result []models.BulkFileJobsModel
	var err error
	//get count
	count, err = b.bulkFileJobsRepo.GetBulkFileJobCounts(requestId, pagination)
	//handle error
	if err != nil {
		log.Println("error in fetching count", err.Error())
		return count, result, err
	}
	//get data
	result, err = b.bulkFileJobsRepo.GetBulkFileJobs(requestId, pagination)
	//handle error
	if err != nil {
		log.Println("error in fetching count", err.Error())
		return count, result, err
	}
	return count, result, err
}

func (b *BulkFileJobUsecase) UploadBulkRequestFile(req *models.BulkFileJobsModel) error {
	var err error
	uploadLoc, err := b.bulkFileJobsRepo.UploadFileToMinio(req.FilePath, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "uploads/", req.RequestId)
	if err != nil {
		log.Printf("error in uploading to minIo %s", err.Error())
		return err
	}
	log.Println("Uploaded at: ", uploadLoc)
	req.FilePath = uploadLoc
	//Create Entry in DB
	if err = b.bulkFileJobsRepo.CreateFileUpload(req); err != nil {
		return fmt.Errorf("error in creating bulk upload file entry %s", err.Error())
	}
	//dump in rmq to be picked by async worker
	return b.bulkFileJobsRepo.SubmitToRabbitMQ(req, "bulk_upload")
}
func (b *BulkFileJobUsecase) ProcessBulkVaccineRecord(model *models.BulkFileJobsModel) error {
	b.bulkFileJobsRepo.UpdateFileUpload(&models.BulkFileJobsModel{Id: model.Id, Status: "PROCESSING"})
	//Get the file
	fileLoc, err := b.bulkFileJobsRepo.GetFileFromActiveServer(os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), model.FilePath)
	if err != nil {
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	//validate the file
	file, err := os.Open(fileLoc)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	defer file.Close()
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	if !(bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) || bytes.Equal(header, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})) {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	f, err := excelize.OpenFile(fileLoc)
	if err != nil {
		log.Printf("failed to open Excel file: %v", err)
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	defer f.Close()
	sheets := f.GetSheetList()
	log.Printf("Sheets: %v\n", sheets)
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	//Header Adjusted
	model.TotalRecords = len(rows) - 1
	if err != nil {
		log.Fatalf("failed to read rows: %v", err)
	}
	vaccineRecord := new([]models.StudentVaccineRecord)
	insertionRecords := []models.VaccineInsertionDBRecord{}
	validat := validator.NewValidator()
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 2 {
			log.Printf("failed to open Excel file: %v", err)
			model.ErrorMessage = "Missing Columns"
			model.Status = "FAILED"
			b.bulkFileJobsRepo.UpdateFileUpload(model)
			return nil
		}
		insertionRecord := models.VaccineInsertionDBRecord{}
		studentId, err := strconv.Atoi(row[0])
		if err != nil {
			log.Printf("invalid insertion Record: %v", err.Error())
			model.ErrorMessage = fmt.Sprintf("invalid entry at row %d", i+1)
			model.Status = "FAILED"
			b.bulkFileJobsRepo.UpdateFileUpload(model)
			return nil
		}
		driveId, err := strconv.Atoi(row[1])
		if err != nil {
			log.Printf("invalid insertion Record: %v", err.Error())
			model.ErrorMessage = fmt.Sprintf("invalid entry at row %d", i+1)
			model.Status = "FAILED"
			b.bulkFileJobsRepo.UpdateFileUpload(model)
			return nil
		}

		sReq := requests.StudentVaccinationRecordCreateRequest{
			StudentId: studentId,
			DriveId:   driveId,
		}
		sModel := models.StudentVaccineRecord{
			StudentId: studentId,
			DriveId:   driveId,
		}
		err = validat.Validate(sReq)
		if err != nil {
			insertionRecord.Record = sModel
			insertionRecord.Status = false
			insertionRecord.ErrorReason = "invalid input"
			insertionRecords = append(insertionRecords, insertionRecord)
			continue
		}
		*vaccineRecord = append(*vaccineRecord, sModel)
	}
	result := b.studentManagementusecaseRepo.CreateVaccinationRecords(vaccineRecord)

	log.Println("Request Processing Complete", result)
	result = append(result, insertionRecords...)

	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Student Id", "Drive Id", "Status", "Remarks"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	//Keeping track Of Processed File
	totalProcessed := 0
	for i, insertion := range result {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), insertion.Record.StudentId)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), insertion.Record.DriveId)
		if !insertion.Status {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), "Rejected")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), insertion.ErrorReason)
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), "Accepted")
			totalProcessed++
		}
	}
	model.ProcessedRecords = totalProcessed
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("error creating report file", err.Error(), reportFileName)
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	log.Println("Report File Created", reportFileName)
	//Upload report
	uploadedReportFile, err := b.bulkFileJobsRepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", model.RequestId)
	os.Remove(reportFileName)

	if err != nil {
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	log.Println("Report File Uploaded", reportFileName)
	model.FilePath = fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)
	//change status
	model.Status = "PROCESSED"
	//update db
	b.bulkFileJobsRepo.UpdateFileUpload(model)
	log.Println("Processing Complete", model)
	return nil
}

func (b *BulkFileJobUsecase) ProcessBulkStudentRecord(model *models.BulkFileJobsModel) error {
	//change status to processing
	b.bulkFileJobsRepo.UpdateFileUpload(&models.BulkFileJobsModel{Id: model.Id, Status: "PROCESSING"})
	//Get the file
	fileLoc, err := b.bulkFileJobsRepo.GetFileFromActiveServer(os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), model.FilePath)
	if err != nil {
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	//validate the file
	file, err := os.Open(fileLoc)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	defer file.Close()
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	if !(bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) || bytes.Equal(header, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})) {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	f, err := excelize.OpenFile(fileLoc)
	if err != nil {
		log.Printf("failed to open Excel file: %v", err)
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	defer f.Close()
	sheets := f.GetSheetList()
	log.Printf("Sheets: %v\n", sheets)
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	//Header Adjusted
	model.TotalRecords = len(rows) - 1
	if err != nil {
		log.Fatalf("failed to read rows: %v", err)
	}
	studentSet := new([]models.StudentManagement)
	insertionRecords := []models.DBInsertionRecord{}
	validat := validator.NewValidator()
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 5 {
			log.Printf("failed to open Excel file: %v", err)
			model.ErrorMessage = "Missing Columns"
			model.Status = "FAILED"
			b.bulkFileJobsRepo.UpdateFileUpload(model)
			return nil
		}
		sReq := requests.StudentManagementCreateRequest{
			Name:    row[0],
			Class:   row[1],
			Gender:  row[2],
			RollNo:  row[3],
			PhoneNo: row[4],
		}
		sModel := models.StudentManagement{
			Name:       row[0],
			Class:      row[1],
			Gender:     row[2],
			RollNumber: row[3],
			PhoneNo:    row[4],
		}
		err := validat.Validate(sReq)
		if err != nil {
			insertionRecord := models.DBInsertionRecord{
				Record:      sModel,
				Status:      false,
				ErrorReason: err.Error(),
			}
			insertionRecords = append(insertionRecords, insertionRecord)
			continue
		}
		*studentSet = append(*studentSet, sModel)
	}
	result := b.studentManagementusecaseRepo.CreateStudentRecords(studentSet)

	log.Println("Request Processing Complete", result)
	result = append(result, insertionRecords...)

	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Name", "Class", "Gender", "Roll Number", "Phone Number", "Status", "Remarks"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	//Keeping track Of Processed File
	totalProcessed := 0
	for i, insertion := range result {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), insertion.Record.Name)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), insertion.Record.Class)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), insertion.Record.Gender)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), insertion.Record.RollNumber)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("E%d", rowNum), insertion.Record.PhoneNo)
		if !insertion.Status {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Rejected")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("G%d", rowNum), insertion.ErrorReason)
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Accepted")
			totalProcessed++
		}
	}
	model.ProcessedRecords = totalProcessed
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("error creating report file", err.Error(), reportFileName)
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	log.Println("Report File Created", reportFileName)
	//Upload report
	uploadedReportFile, err := b.bulkFileJobsRepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", model.RequestId)

	if err != nil {
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.bulkFileJobsRepo.UpdateFileUpload(model)
		return nil
	}
	log.Println("Report File Uploaded", reportFileName)
	model.FilePath = fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)
	//change status
	model.Status = "PROCESSED"
	//update db
	b.bulkFileJobsRepo.UpdateFileUpload(model)
	log.Println("Processing Complete", model)
	return nil
}

func NewBulkFileJobUsecaseHandler(studentUcRepo StudentManagementUsecaseHandler, bulkfileJobsRepo repository.BulkFileJobsRepositoryHandler) BulkFileJobUsecaseHandler {
	return &BulkFileJobUsecase{
		studentManagementusecaseRepo: studentUcRepo,
		bulkFileJobsRepo:             bulkfileJobsRepo,
	}
}
