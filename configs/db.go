package configs

import (
	"log"
	"os"
	"smart-stock/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// ป้องกันไม่ให้ต่อซ้ำถ้า DB มีค่าอยู่แล้ว
	if DB != nil {
		return
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("❌ DATABASE_URL is not set")
		return
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// เปลี่ยนจาก log.Fatal เป็น log.Println เพื่อให้ระบบไม่ดับ
		log.Println("❌ Failed to connect to Supabase: ", err)
		return
	}

	err = database.AutoMigrate(
		&models.User{},
		&models.Chemical{},
		&models.InventoryLot{},
		&models.Transaction{},
	)
	if err != nil {
		log.Println("❌ Migration failed: ", err)
		return
	}

	DB = database
	log.Println("✅ Supabase connected & Migrated successfully!")
}