package controllers

import (
	"net/http"
	"smart-stock/configs"
	"smart-stock/models" // ถ้าไฟล์ schema อยู่ใน package models ก็ import ปกติครับ
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("smart_stock_secret_key_2026")

// โครงสร้างตอนสมัครสมาชิก (ต้องมีชื่อเต็มด้วย)
type RegisterRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	FullName   string `json:"full_name" binding:"required"`
	Department string `json:"department"`
}

// โครงสร้างตอนล็อกอิน (ใช้แค่ 2 อย่าง)
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 1. API สมัครสมาชิก
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกข้อมูลให้ครบถ้วน (ต้องมี Username, Password, FullName)"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสรหัสผ่านล้มเหลว"})
		return
	}

	user := models.User{
		Username:   req.Username,
		Password:   string(hashedPassword),
		FullName:   req.FullName,
		Department: req.Department,
		Role:       "USER", // ค่า Default เริ่มต้น
	}

	if err := configs.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username นี้ถูกใช้งานแล้ว หรือข้อมูลไม่ถูกต้อง"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างบัญชีผู้ใช้งานสำเร็จ!"})
}

// 2. API ล็อกอิน
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอก Username และ Password"})
		return
	}

	var user models.User
	if err := configs.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบผู้ใช้งานในระบบ"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "รหัสผ่านไม่ถูกต้อง"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"username":  user.Username,
		"user_id":   user.ID,
		"role":      user.Role, // 💡 แอบแนบ Role ไปใน Token ด้วยเผื่อใช้ทำสิทธิ์แอดมินในอนาคต!
		"exp":       expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง Token ล้มเหลว"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "เข้าสู่ระบบสำเร็จ",
		"token":     tokenString,
		"username":  user.Username,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}