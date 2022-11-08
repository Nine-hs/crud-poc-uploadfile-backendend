package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Catalog struct {
	Id    uint   `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Title string `json:"title"`
	Model string `json:"model"`
	Year  string `json:"year"`
	Gear  string `json:"gear"`
	Image string `json:"image"`
}

func Database() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("database/data.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Catalog{})

	return db, nil
}
