package configs

import (
	"errors"
	"fmt"
	"os"
	"smart-stock/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// เปลี่ยนให้มีการ return error ออกมา
func ConnectDatabase() error {
	// ถ้าต่อแล้วก็ผ่านไปเลย
	if DB != nil {
		return nil
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is empty in this environment")
	}

	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // ปิด Prepared Statement เพื่อให้คุยกับ Supabase Pooler ได้
	}), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("GORM Connection error: %v", err)
	}

	// (ตอนนี้ตารางถูกสร้างไปแล้ว เราเลยคอมเมนต์บรรทัด Migrate ทิ้งไปก่อนเพื่อความเร็ว)
	err = database.AutoMigrate(
		&models.User{},
		&models.Chemical{},
		&models.InventoryLot{},
		&models.Transaction{},
	)


	DB = database
	return nil
}
