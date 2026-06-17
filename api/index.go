package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	// สร้าง Gin instance
	app = gin.Default()

	// สร้าง Route Group สำหรับ API
	api := app.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong! ระบบสต๊อกพร้อมใช้งาน"})
		})
		
		// TODO: เพิ่ม Route สำหรับรับเข้า-เบิกจ่ายตรงนี้
		// api.POST("/items", controllers.CreateItem)
	}
}

// Handler คือฟังก์ชันบังคับที่ Vercel จะเรียกใช้
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}