package api

import (
	"net/http"
	"smart-stock/controllers"
	"smart-stock/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	app = gin.New()

	// 2. --- เพิ่มตั้งค่า CORS ตรงนี้ (ต้องอยู่ก่อน api := app.Group) ---
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173","https://smart-stock-ui.vercel.app"}, // อนุญาตให้ React พอร์ต 5173 เข้าถึงได้
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// ----------------------------------------------------

	api := app.Group("/api")

	api.POST("/auth/register", controllers.Register)
	api.POST("/auth/login", controllers.Login)

	protected := api.Group("/")
	protected.Use(middlewares.JWTAuth())
	{
		protected.POST("/chemicals", controllers.CreateChemical)
		protected.GET("/chemicals", controllers.GetChemicals)
		protected.POST("/inventory/receive", controllers.ReceiveChemical)
		protected.GET("/inventory/balance", controllers.GetStockBalance)
		protected.POST("/inventory/dispense", controllers.DispenseChemical)
		protected.GET("/inventory/lots", controllers.GetInventoryLots)
		protected.GET("/inventory/history", controllers.GetTransactionHistory)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
