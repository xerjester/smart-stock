package api

import (
	"net/http"
	
	// เปลี่ยนเป็น path จริงของคุณ
	"smart-stock/configs"
	"smart-stock/controllers" 

	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	configs.ConnectDatabase()

	app = gin.Default()
	api := app.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "API พร้อมใช้งาน!"})
		})

		// เพิ่ม Endpoint สำหรับเบิกของตรงนี้
		api.POST("/dispense", controllers.DispenseChemical)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}