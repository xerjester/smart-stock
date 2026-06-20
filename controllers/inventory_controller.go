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
		Joins("LEFT JOIN inventory_lots ON chemicals.id::text = inventory_lots.chemical_id AND inventory_lots.status = 'ACTIVE'").
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

// โครงสร้างสำหรับรับข้อมูลการเบิก
type DispenseRequest struct {
	ChemicalID string  `json:"chemical_id" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UserID     string  `json:"user_id" binding:"required"`
}

// ฟังก์ชันเบิกจ่ายแบบ FEFO
func DispenseChemical(c *gin.Context) {
	if err := configs.ConnectDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Connection Failed"})
		return
	}

	var req DispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	tx := configs.DB.Begin() // เริ่ม Transaction ถ้าพังกลางทางจะได้ยกเลิกทั้งหมด

	// 1. ดึงล็อตทั้งหมดของสารเคมีนี้ ที่ยังมีของเหลือ (ACTIVE) โดยเรียงจากวันหมดอายุใกล้สุดไปไกลสุด
	var lots []models.InventoryLot
	if err := tx.Where("chemical_id = ? AND status = 'ACTIVE' AND quantity_remain > 0", req.ChemicalID).
		Order("expiration_date ASC").
		Find(&lots).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลล็อตได้"})
		return
	}

	// 2. เช็กว่าสต๊อกรวมมีพอให้เบิกไหม
	var totalAvailable float64
	for _, lot := range lots {
		totalAvailable += lot.QuantityRemain
	}

	if req.Quantity > totalAvailable {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "สต๊อกไม่เพียงพอ",
			"available": totalAvailable,
			"requested": req.Quantity,
		})
		return
	}

	// 3. เริ่มกระบวนการตัดสต๊อก (FEFO)
	remainToDispense := req.Quantity
	var dispensedDetails []map[string]interface{}

	for _, lot := range lots {
		if remainToDispense <= 0 {
			break // เบิกครบแล้ว ออกจากลูป
		}

		var deductAmount float64

		if lot.QuantityRemain >= remainToDispense {
			// ล็อตนี้มีของพอให้ตัดจนครบ
			deductAmount = remainToDispense
			lot.QuantityRemain -= deductAmount
			remainToDispense = 0
		} else {
			// ล็อตนี้มีของไม่พอ ต้องตัดจนเกลี้ยง (0) แล้วไปเอาล็อตถัดไป
			deductAmount = lot.QuantityRemain
			remainToDispense -= lot.QuantityRemain
			lot.QuantityRemain = 0
			lot.Status = "EMPTY" // เปลี่ยนสถานะล็อตว่าของหมดแล้ว
		}

		// อัปเดตข้อมูลล็อตในฐานข้อมูล
		if err := tx.Save(&lot).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "อัปเดตสต๊อกล้มเหลว"})
			return
		}

		// บันทึกประวัติการเบิก (OUT) ลงตาราง Transaction
		history := models.Transaction{
			LotID:           lot.ID,
			UserID:          req.UserID,
			TransactionType: "OUT",
			Quantity:        deductAmount,
			Remarks:         "เบิกใช้งาน",
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกประวัติล้มเหลว"})
			return
		}

		// เก็บประวัติไว้สรุปส่งกลับไปให้หน้าเว็บดู
		dispensedDetails = append(dispensedDetails, map[string]interface{}{
			"batch_number": lot.BatchNumber,
			"deducted":     deductAmount,
		})
	}

	tx.Commit() // บันทึกทุกอย่างลง Database

	c.JSON(http.StatusOK, gin.H{
		"message":           "เบิกจ่ายสำเร็จ",
		"dispensed_details": dispensedDetails,
	})
}

// โครงสร้างข้อมูลสำหรับส่งกลับไปให้หน้าเว็บ (แบบละเอียด)
type LotDetailResponse struct {
	ChemicalCode   string  `json:"chemical_code"`
	Name           string  `json:"name"`
	BatchNumber    string  `json:"batch_number"`
	QuantityRemain float64 `json:"quantity_remain"`
	BaseUnit       string  `json:"base_unit"`
	ExpirationDate string  `json:"expiration_date"`
}

// ฟังก์ชันดึงข้อมูลรายล็อต
func GetInventoryLots(c *gin.Context) {
	if err := configs.ConnectDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Connection Failed"})
		return
	}

	var results []LotDetailResponse
	// ใช้ SQL JOIN ดึงชื่อสารเคมีมาประกบกับเลขล็อต และเอาเฉพาะที่ยังมีของอยู่ (>0)
	query := `
		SELECT 
			c.chemical_code, 
			c.name, 
			i.batch_number, 
			i.quantity_remain, 
			c.base_unit, 
			TO_CHAR(i.expiration_date, 'YYYY-MM-DD') as expiration_date
		FROM inventory_lots i
		JOIN chemicals c ON i.chemical_id = c.id::text
		WHERE i.quantity_remain > 0
		ORDER BY c.chemical_code ASC, i.expiration_date ASC
	`

	if err := configs.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลล้มเหลว"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// โครงสร้างสำหรับส่งข้อมูลประวัติกลับไปหน้าเว็บ
type TransactionHistoryResponse struct {
	TransactionDate string  `json:"transaction_date"`
	TransactionType string  `json:"transaction_type"` // IN หรือ OUT
	ChemicalCode    string  `json:"chemical_code"`
	Name            string  `json:"name"`
	BatchNumber     string  `json:"batch_number"`
	Quantity        float64 `json:"quantity"`
	Remarks         string  `json:"remarks"`
	UserID          string  `json:"user_id"`
}

// ฟังก์ชันดึงประวัติการเคลื่อนไหว
func GetTransactionHistory(c *gin.Context) {
	if err := configs.ConnectDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Connection Failed"})
		return
	}

	var results []TransactionHistoryResponse
	// ใช้ SQL JOIN ดึงประวัติข้าม 3 ตาราง: transactions -> inventory_lots -> chemicals
	query := `
		SELECT 
			TO_CHAR(t.created_at, 'YYYY-MM-DD HH24:MI:SS') as transaction_date,
			t.transaction_type,
			c.chemical_code,
			c.name,
			i.batch_number,
			t.quantity,
			t.remarks,
			t.user_id
		FROM transactions t
		JOIN inventory_lots i ON t.lot_id::text = i.id::text
		JOIN chemicals c ON i.chemical_id::text = c.id::text
		ORDER BY t.created_at DESC
	`

	if err := configs.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลประวัติล้มเหลว"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}
