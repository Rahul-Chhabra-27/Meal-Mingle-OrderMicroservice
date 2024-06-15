package config

import (
	"fmt"
	"log"
	"os"

	"order-microservice/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DatabaseDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
}

func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func ConnectDB() (*gorm.DB, *gorm.DB, error) {
	orderDB, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	orderDB.AutoMigrate(&model.Order{})
	orderItemDB, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	orderItemDB.AutoMigrate(&model.OrderItem{})

	return orderDB, orderItemDB, nil
}

func GetUserConnector() (*gorm.DB, error) {
	userDB, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	userDB.AutoMigrate(&model.User{})
	return userDB, nil
}
func GetRestaurantConnector() (*gorm.DB, error) {
	restaurantDB, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	restaurantDB.AutoMigrate(&model.Restaurant{})
	return restaurantDB, nil
}
func ValidateOrderFields(shippingAddress string, restaurantName string, orderAmount int64) bool {
	if shippingAddress == "" || restaurantName == "" || orderAmount == 0 {
		return false
	}
	return true
}
