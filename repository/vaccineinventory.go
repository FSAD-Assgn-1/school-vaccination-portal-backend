package repository

import (
	"fmt"
	"log"
	"school_vaccination_portal/databases/mysql"
	"school_vaccination_portal/models"
	"school_vaccination_portal/requests"
)

type VaccineInventoryHandler interface {
	GetVaccineInventory(filter string) ([]models.VaccineInventory, error)
	CreateInventory(drive *models.VaccineInventory) error
	UpdateVaccineInventory(drive *requests.VaccineInventoryUpdateRequest) error
}

type Vacci struct {
	DB *mysql.MysqlConnect
}

func (v *Vacci) GetVaccineInventory(filter string) ([]models.VaccineInventory, error) {
	drives := []models.VaccineInventory{}
	var err error
	if filter == "" {
		err = v.DB.Table("vaccination_inventory").Order("drive_date ASC").Find(&drives).Error
		if err != nil {
			log.Println("error in fetching drives", err.Error())
			return drives, err
		}
		return drives, nil
	}
	fmt.Println("Filter being applied", filter)
	err = v.DB.Table("vaccination_inventory").Where(filter).Order("drive_date ASC").Find(&drives).Error
	if err != nil {
		log.Println("error in fetching drives", err.Error())
		return drives, err
	}
	return drives, nil
}
func (v *Vacci) CreateInventory(drive *models.VaccineInventory) error {
	return v.DB.Table("vaccination_inventory").Create(drive).Error
}

func (v *Vacci) UpdateVaccineInventory(drive *requests.VaccineInventoryUpdateRequest) error {
	updateMap := map[string]interface{}{}

	if drive.DriveDate != nil {
		updateMap["drive_date"] = drive.DriveDate
	}
	if drive.VaccineName != nil {
		updateMap["vaccine_name"] = drive.VaccineName
	}
	if drive.Doses != nil {
		updateMap["doses"] = drive.Doses
	}
	if drive.Classes != nil {
		updateMap["classes"] = drive.Classes
	}
	return v.DB.Table("vaccination_inventory").Where(fmt.Sprintf("id = %d", drive.Id)).Updates(updateMap).Error
}

func NewVaccineInventoryHandler(db *mysql.MysqlConnect) VaccineInventoryHandler {
	return &Vacci{
		DB: db,
	}
}
