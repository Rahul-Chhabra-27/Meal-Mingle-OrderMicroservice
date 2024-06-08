package model

import "gorm.io/gorm"

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
	Discount int64
	RestaurantId uint `gorm:"foreignKey:RestaurantId;references:ID"` 
}

type User struct {
	gorm.Model
	Name string
	Password string
	Email   string  `gorm:"unique"`
	Phone	string  `gorm:"unique"`
	Address string
	City string
}

type Restaurant struct {
	gorm.Model 
	Name string `gorm:"unique"`
	Address string 	
	Phone string `gorm:"unique"`
	Availability string `gorm:"default:open"`
	Rating float32 `gorm:"default:0"`
	RestaurantOwnerMail string
	City string
}