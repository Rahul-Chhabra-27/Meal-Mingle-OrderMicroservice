package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"order-microservice/model"
)

func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func ConnectDB(dsn string) (*gorm.DB, *gorm.DB, error) {
	orderDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	orderDB.AutoMigrate(&model.Order{})
	orderItemDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	orderItemDB.AutoMigrate(&model.OrderItem{})

	return orderDB ,orderItemDB, nil
}

func GetUserConnector(dsn string) (*gorm.DB, error) {
	userDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	userDB.AutoMigrate(&model.User{})
	return userDB, nil
}
func GetRestaurantConnector(dsn string) (*gorm.DB, error) {
	restaurantDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	restaurantDB.AutoMigrate(&model.Restaurant{})
	return restaurantDB, nil
}
func ValidateOrderFields(shippingAddress string ,restaurantName string, orderAmount int64,) bool {
	if shippingAddress == "" || restaurantName == "" || orderAmount == 0  {
		return false
	}
	return true
}