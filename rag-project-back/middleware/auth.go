package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("my-super-secret-key") // ต้องตรงกับหน้า Auth Handler

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. ดึง Token จาก Header "Authorization: Bearer <token>"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Missing Token"})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// 2. ตรวจสอบความถูกต้องของ Token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Invalid Token"})
		}

		// 3. ดึงข้อมูล Username ออกมาจาก Token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Invalid Claims"})
		}

		// 4. ฝากชื่อ User ไว้ใน Locals (เพื่อให้ Handler ถัดไปเอาไปใช้)
		c.Locals("username", claims["username"])

		return c.Next()
	}
}
