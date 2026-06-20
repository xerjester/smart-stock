package api

import (
	"net/http"
	"smart-stock/controllers"

	"github.com/gin-contrib/cors" // 1. อย่าลืม Import ตรงนี้
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	app = gin.New()

	// 2. --- เพิ่มตั้งค่า CORS ตรงนี้ (ต้องอยู่ก่อน api := app.Group) ---
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // อนุญาตให้ React พอร์ต 5173 เข้าถึงได้
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// ----------------------------------------------------

	api := app.Group("/api")

	api.POST("/chemicals", controllers.CreateChemical)
	api.GET("/chemicals", controllers.GetChemicals)
	api.POST("/inventory/receive", controllers.ReceiveChemical)
	api.GET("/inventory/balance", controllers.GetStockBalance)
	api.POST("/inventory/dispense", controllers.DispenseChemical)
	api.GET("/inventory/lots", controllers.GetInventoryLots)
	api.GET("/inventory/history", controllers.GetTransactionHistory)
	api.POST("/auth/register", controllers.Register)
	api.POST("/auth/login", controllers.Login)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
