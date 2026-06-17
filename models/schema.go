package models

import (
	"time"

	"gorm.io/gorm"
)

// User ข้อมูลบุคลากรและการจัดการสิทธิ์
type User struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FullName   string `gorm:"not null"`
	Department string
	Role       string `gorm:"not null;default:'USER'"` // SUPER_ADMIN, ADMIN, APPROVER, USER
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// Chemical ตารางข้อมูลสารเคมีหลัก
type Chemical struct {
	ID           string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ChemicalCode string  `gorm:"uniqueIndex;not null"` // เช่น CHM-1001
	Name         string  `gorm:"not null"`             // เช่น Ethanol 95%
	BaseUnit     string  `gorm:"not null"`             // หน่วยวัด เช่น ml, g, bottle
	MinimumLevel float64 // จุดสั่งซื้อเพิ่ม (แจ้งเตือนเมื่อของใกล้หมด)
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// InventoryLot ตารางจัดเก็บสต๊อกแยกตามล็อต (สำคัญมากสำหรับระบบ FEFO)
type InventoryLot struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ChemicalID     string    `gorm:"index;not null"`
	BatchNumber    string    // เลขล็อตจากผู้ผลิต
	QuantityRemain float64   // จำนวนที่เหลืออยู่ในล็อตนี้
	ReceivedDate   time.Time
	ExpirationDate time.Time `gorm:"index"` // ทำ Index เพื่อให้ค้นหาวันหมดอายุได้เร็วขึ้น
	Status         string    `gorm:"default:'ACTIVE'"` // ACTIVE, EXPIRED, DEPLETED
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// Transaction ตารางประวัติการทำรายการ (Audit Trail)
type Transaction struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	LotID           string    `gorm:"index;not null"`
	UserID          string    `gorm:"not null"` // ผู้ทำรายการ (เชื่อมกับ User.ID)
	TransactionType string    `gorm:"not null"` // "IN" (รับเข้า), "OUT" (เบิกจ่าย), "ADJUST" (ปรับปรุงยอด)
	Quantity        float64   `gorm:"not null"`
	Remarks         string    // หมายเหตุเพิ่มเติม
	TransactionDate time.Time `gorm:"autoCreateTime"` // บันทึกเวลาอัตโนมัติเมื่อมีการสร้าง Record
}