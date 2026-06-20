package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ต้องเป็นคีย์ตัวเดียวกับที่ใช้สร้าง Token ใน auth_controller นะครับ
var jwtKey = []byte("smart_stock_secret_key_2026")

// JWTAuth คือฟังก์ชันสกัดกั้นเพื่อตรวจ Token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. ขอดูบัตรผ่านจาก Header (ในชื่อ Authorization)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access Denied: กรุณาล็อกอินก่อนใช้งาน"})
			c.Abort() // เตะออกทันที ไม่ให้ไปต่อ
			return
		}

		// 2. รูปแบบของบัตรผ่านจะเป็น "Bearer xxxxx.yyyyy.zzzzz" เราต้องตัดคำว่า "Bearer " ออก
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. ตรวจสอบความถูกต้องและวันหมดอายุของ Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access Denied: บัตรผ่านไม่ถูกต้องหรือหมดอายุ"})
			c.Abort()
			return
		}

		// 4. ถ้าบัตรผ่านถูกต้อง ให้ดึงข้อมูลที่ซ่อนไว้ออกมาฝากไว้ใน Context เผื่อ API อื่นอยากใช้
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("username", claims["username"])
			c.Set("user_id", claims["user_id"])
			c.Set("role", claims["role"])
		}

		// 5. ปล่อยให้เข้าไปใช้งาน API ต่อได้!
		c.Next()
	}
}