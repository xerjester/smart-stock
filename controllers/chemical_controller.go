package controllers

import (
	"net/http"

	"smart-stock/configs"
	"smart-stock/models"

	"github.com/gin-gonic/gin"
)

// 1. ฟังก์ชันเพิ่มสารเคมีใหม่ (Create)
func CreateChemical(c *gin.Context) {
	var input models.Chemical
	
	// รับข้อมูล JSON จากหน้าเว็บหรือ Postman มาแปลงเป็น Struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง: " + err.Error()})
		return
	}

	// สั่ง GORM ให้บันทึกข้อมูลลงฐานข้อมูล
	if err := configs.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "เพิ่มสารเคมีใหม่สำเร็จ",
		"data":    input,
	})
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