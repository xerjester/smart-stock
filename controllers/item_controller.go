package controllers

import (
	"net/http"

	// เปลี่ยน "your_project_name" เป็นชื่อ module ใน go.mod ของคุณ
	"smart-stock/configs" 
	"smart-stock/models"

	"github.com/gin-gonic/gin"
)

// โครงสร้างของข้อมูลที่รับมาจากหน้าเว็บ (JSON Payload)
type DispenseRequest struct {
	ChemicalID string  `json:"chemical_id" binding:"required"`
	UserID     string  `json:"user_id" binding:"required"` // สมมติว่าส่งรหัสคนเบิกมาด้วย
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
}

func DispenseChemical(c *gin.Context) {
	var req DispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบ"})
		return
	}

	// เริ่ม Database Transaction เพื่อป้องกันสต๊อกเพี้ยน
	tx := configs.DB.Begin()
	// หากฟังก์ชันพังกลางคัน ให้ Rollback ข้อมูลกลับ
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. ค้นหาล็อตของสารเคมีนี้ที่ยังไม่หมด เรียงตามวันหมดอายุ (FEFO)
	var lots []models.InventoryLot
	err := tx.Where("chemical_id = ? AND quantity_remain > 0 AND status = 'ACTIVE'", req.ChemicalID).
		Order("expiration_date ASC").
		Find(&lots).Error

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลสต๊อกล้มเหลว"})
		return
	}

	remainToDispense := req.Quantity // ยอดที่ต้องการเบิกทั้งหมด

	// 2. วนลูปหักยอดทีละล็อต (จากล็อตที่เก่าที่สุดก่อน)
	for i := range lots {
		if remainToDispense <= 0 {
			break // เบิกครบตามจำนวนแล้ว ออกจากลูปได้เลย
		}

		// คำนวณว่าจะหักจากล็อตนี้เท่าไหร่
		deductAmount := lots[i].QuantityRemain
		if remainToDispense < deductAmount {
			deductAmount = remainToDispense
		}

		// 3. อัปเดตยอดคงเหลือในล็อต
		lots[i].QuantityRemain -= deductAmount
		if lots[i].QuantityRemain == 0 {
			lots[i].Status = "DEPLETED" // ของหมดล็อตแล้ว เปลี่ยนสถานะ
		}

		if err := tx.Save(&lots[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "อัปเดตยอดสต๊อกล้มเหลว"})
			return
		}

		// 4. บันทึกประวัติ (Audit Trail)
		history := models.Transaction{
			LotID:           lots[i].ID,
			UserID:          req.UserID,
			TransactionType: "OUT", // ระบุว่าเป็นการเบิกออก
			Quantity:        deductAmount,
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกประวัติการเบิกล้มเหลว"})
			return
		}

		// หักยอดที่ต้องการเบิกลง
		remainToDispense -= deductAmount
	}

	// 5. ตรวจสอบว่าของในสต๊อกพอให้เบิกครบไหม
	if remainToDispense > 0 {
		tx.Rollback() // ยกเลิกการหักยอดทั้งหมดที่ทำมาในลูป
		c.JSON(http.StatusBadRequest, gin.H{"error": "สารเคมีในสต๊อกไม่เพียงพอ"})
		return
	}

	// บันทึกการเปลี่ยนแปลงทั้งหมดลงฐานข้อมูล (Commit)
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message": "เบิกจ่ายสารเคมีสำเร็จ",
		"dispensed_quantity": req.Quantity,
	})
}