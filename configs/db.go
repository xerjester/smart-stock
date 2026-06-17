package configs

import (
	"log"
	"os"

	"smart-stock/models" // เรียกใช้ models จากโปรเจกต์ของเรา
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// ดึง URL จาก Environment Variable ของ Vercel (หรือเครื่องเราตอนรันทดสอบ)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to Supabase: ", err)
	}

	// สร้างตารางอัตโนมัติตาม struct ในโฟลเดอร์ models
	err = database.AutoMigrate(
		&models.User{},
		&models.Chemical{},
		&models.InventoryLot{},
		&models.Transaction{},
	)
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}

	DB = database
	log.Println("Supabase connected successfully!")
}