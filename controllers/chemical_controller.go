package controllers

import (
	"net/http"

	"smart-stock/configs"
	"smart-stock/models"

	"github.com/gin-gonic/gin"
)

// 1. ฟังก์ชันเพิ่มสารเคมีใหม่ (Create)
func CreateChemical(c *gin.Context) {
	// เพิ่มตัวเช็กว่า DB พร้อมใช้งานไหม
	if configs.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not initialized"})
		return
	}

	var input models.Chemical
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตอนนี้บรรทัดนี้จะไม่พังแล้วเพราะเช็ก DB แล้ว
	if err := configs.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Save failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Success"})
}

// 2. ฟังก์ชันดึงรายการสารเคมีทั้งหมด (Read)
func GetChemicals(c *gin.Context) {
	var chemicals []models.Chemical

	// สั่ง GORM ให้ไปค้นหาข้อมูลทั้งหมดในตาราง Chemical
	if err := configs.DB.Find(&chemicals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลล้มเหลว"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ดึงข้อมูลสำเร็จ",
		"data":    chemicals,
	})
}
