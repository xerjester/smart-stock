package api

import (
	"net/http"

	"smart-stock/configs" // นำเข้า configs

	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	// 1. สั่งให้เชื่อมต่อ Database ทันทีที่ API ตื่น
	configs.ConnectDatabase()

	app = gin.Default()

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"project": "SMART-STOCK API",
			"message": "ยินดีต้อนรับสู่ระบบหลังบ้าน! เข้าใช้งาน API ได้ที่ /api/ping",
		})
	})

	api := app.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong! ระบบสต๊อกพร้อมใช้งาน"})
		})
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
