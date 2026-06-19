package api

import (
	"net/http"

	"smart-stock/configs"
	"smart-stock/controllers" // อย่าลืม import controllers

	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	configs.ConnectDatabase()

	app = gin.Default()

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "online",
			"project": "SMART-STOCK API",
		})
	})

	api := app.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong! ระบบสต๊อกพร้อมใช้งาน"})
		})

		// --- เพิ่ม 2 เส้นทางใหม่ตรงนี้ ---
		api.POST("/chemicals", controllers.CreateChemical) // รับข้อมูลเพื่อสร้างใหม่ (POST)
		api.GET("/chemicals", controllers.GetChemicals)    // ดึงข้อมูลทั้งหมด (GET)
		api.POST("/inventory/receive", controllers.ReceiveChemical)
		api.GET("/inventory/balance", controllers.GetStockBalance)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
