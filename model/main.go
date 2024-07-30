package model

import "gorm.io/gorm"
const (
	AdminRole = "admin"
	UserRole = "user"
)

type OrderItem struct {
	gorm.Model
	Name    string
	Quantity int64
	Price   int64
	OrderId uint `gorm:"foreignKey:OrderId;references:ID"`
}

type Order struct {
	gorm.Model
	UserId uint `gorm:"foreignKey:UserId;references:ID"`
	TotalPrice int64
	Status string 
	Discount string
	RestaurantId uint `gorm:"foreignKey:RestaurantId;references:ID"` 
	ShippingAddress string
}

type User struct {
	gorm.Model
	Name string
	Password string
	Email   string  `gorm:"unique"`
	Phone	string  `gorm:"unique"`
	Address string
	City string
	Role string
}

type Restaurant struct {
	gorm.Model 
	Name string `gorm:"unique"`
	Phone string `gorm:"unique"`
	Availability string `gorm:"default:open"`
	Rating float32 `gorm:"default:0"`
	RestaurantOwnerMail string
	ImageUrl string
	OperationDays string
	OperationHours string
}

type Address struct {
	gorm.Model
	RestaurantId uint `gorm:"foreignKey:RestaurantId;references:ID;uniqueIndex:idx_restaurant_address"` // foreign key referencing the primary key of the Restaurant table
	StreetName string
	Pincode string
	City string
	Country string
}