package controllers

import (
	"net/http"
	"smart-stock/configs"
	"smart-stock/models"

	"github.com/gin-gonic/gin"
)

func CreateChemical(c *gin.Context) {
	// 1. ลองเชื่อมต่อ Database แบบ On-Demand และดึง Error มาโชว์
	if err := configs.ConnectDatabase(); err != nil {
		// ถ้าพัง โชว์ Error ตรงๆ ไปเลย
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DB Connection Failed",
			"details": err.Error(),
		})
		return
	}

	// 2. รับข้อมูล JSON
	var input models.Chemical
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
		return
	}

	// 3. เซฟลง Database
	if err := configs.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "เพิ่มสารเคมีใหม่สำเร็จ",
		"data":    input,
	})
}

func GetChemicals(c *gin.Context) {
	if err := configs.ConnectDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DB Connection Failed",
			"details": err.Error(),
		})
		return
	}

	var chemicals []models.Chemical
	if err := configs.DB.Find(&chemicals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ดึงข้อมูลล้มเหลว"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": chemicals})
}