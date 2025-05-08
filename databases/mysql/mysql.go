package mysql

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type MysqlConnect struct {
	*gorm.DB
}

var mysqlConnect *MysqlConnect

func GetMySQLConnect() (*MysqlConnect, error) {
	var err error
	connection, err := gorm.Open("mysql", getConnectionString())
	if err != nil {
		log.Println("Error connecting to MySQL", err.Error())
		return mysqlConnect, err
	}
	mysqlConnect = &MysqlConnect{connection}
	log.Println("MySQL is Connected")
	return mysqlConnect, err
}
func getConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
}
func Close() {
	log.Println("Closing MySQL Connection")
	mysqlConnect.Close()
}
