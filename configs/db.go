package configs

import (
	"log"
	"os"

	"smart-stock/models" // อ้างอิงโฟลเดอร์ models ของโปรเจกต์
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// ดึง Connection String จาก Environment Variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("⚠️ DATABASE_URL is not set")
		return
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to Supabase: ", err)
	}

	// สร้างตารางใน Supabase อัตโนมัติ
	err = database.AutoMigrate(
		&models.User{},
		&models.Chemical{},
		&models.InventoryLot{},
		&models.Transaction{},
	)
	if err != nil {
		log.Fatal("❌ Migration failed: ", err)
	}

	DB = database
	log.Println("✅ Supabase connected & Migrated successfully!")
}