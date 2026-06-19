package controllers

import (
	"net/http"
	"time"

	"smart-stock/configs"
	"smart-stock/models"

	"github.com/gin-gonic/gin"
)

// โครงสร้างข้อมูลสำหรับรับค่าจากหน้าเว็บ/Postman
type ReceiveRequest struct {
	ChemicalID     string  `json:"chemical_id" binding:"required"`
	BatchNumber    string  `json:"batch_number" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	ExpirationDate string  `json:"expiration_date" binding:"required"` // ส่งมาในรูปแบบ YYYY-MM-DD
	UserID         string  `json:"user_id" binding:"required"`
}

// ฟังก์ชันรับของเข้าสต๊อก
func ReceiveChemical(c *gin.Context) {
	var req ReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน หรือรูปแบบผิด: " + err.Error()})
		return
	}

	// แปลงวันที่จาก String เป็น time.Time
	expDate, err := time.Parse("2006-01-02", req.ExpirationDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบวันหมดอายุต้องเป็น YYYY-MM-DD"})
		return
	}

	// เริ่ม Database Transaction
	tx := configs.DB.Begin()

	// 1. สร้างล็อตใหม่ในตาราง InventoryLot
	newLot := models.InventoryLot{
		ChemicalID:     req.ChemicalID,
		BatchNumber:    req.BatchNumber,
		QuantityRemain: req.Quantity,
		ReceivedDate:   time.Now(),
		ExpirationDate: expDate,
		Status:         "ACTIVE",
	}

	if err := tx.Create(&newLot).Error; err != nil {
		tx.Rollback() // พังปุ๊บ ยกเลิกทันที
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างล็อตใหม่ได้"})
		return
	}

	// 2. บันทึกประวัติลงตาราง Transaction (Audit Trail)
	history := models.Transaction{
		LotID:           newLot.ID,
		UserID:          req.UserID,
		TransactionType: "IN",
		Quantity:        req.Quantity,
		Remarks:         "รับสินค้าเข้าใหม่ ล็อต: " + req.BatchNumber,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกประวัติได้"})
		return
	}

	// ทำงานสำเร็จทั้ง 2 ตาราง กด Commit บันทึกจริงลง Database
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "รับสินค้าเข้าสต๊อกสำเร็จ",
		"lot_id":  newLot.ID,
	})
}

// โครงสร้างสำหรับส่งข้อมูลกลับไปให้หน้า Dashboard
type ChemicalBalanceResponse struct {
	ChemicalID   string  `json:"chemical_id"`
	ChemicalCode string  `json:"chemical_code"`
	Name         string  `json:"name"`
	TotalRemain  float64 `json:"total_remain"`
	BaseUnit     string  `json:"base_unit"`
}

// ฟังก์ชันดึงสรุปยอดคงเหลือของสารเคมีแต่ละชนิด
func GetStockBalance(c *gin.Context) {
	// ป้องกันเคส Vercel Cold Start แล้วลืมต่อฐานข้อมูล
	if err := configs.ConnectDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Connection Failed"})
		return
	}

	var balances []ChemicalBalanceResponse

	// เขียน Query เพื่อ Join ตาราง Chemicals เข้ากับ InventoryLots 
	// และ SUM ยอด QuantityRemain เฉพาะล็อตที่สถานะยังเป็น ACTIVE
	err := configs.DB.Table("chemicals").
		Select(`
			chemicals.id as chemical_id, 
			chemicals.chemical_code, 
			chemicals.name, 
			chemicals.base_unit, 
			COALESCE(SUM(inventory_lots.quantity_remain), 0) as total_remain
		`).
		Joins("LEFT JOIN inventory_lots ON chemicals.id = inventory_lots.chemical_id AND inventory_lots.status = 'ACTIVE'").
		Group("chemicals.id").
		Scan(&balances).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลสรุปยอดล้มเหลว", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ดึงข้อมูลสต๊อกสำเร็จ",
		"data":    balances,
	})
}