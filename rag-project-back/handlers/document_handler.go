package handlers

import (
	"fmt"
	"rag-project/services"

	"github.com/gofiber/fiber/v2"
)

type DocumentHandler struct {
	service services.RagService
}

func NewDocumentHandler(service services.RagService) *DocumentHandler {
	return &DocumentHandler{service: service}
}

// UploadPDF: อัปโหลดไฟล์ (ฟังก์ชันนี้ยังเหมือนเดิม แต่ถ้าอยู่ใน Group /api ก็จะต้องมี Token ถึงจะเข้าได้)
func (h *DocumentHandler) UploadPDF(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Upload file failed"})
	}

	// สร้างโฟลเดอร์ uploads ไว้กันเหนียว (ถ้ายังไม่มี)
	filePath := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Save file failed"})
	}

	docID, err := h.service.ProcessPDF(c.Context(), file.Filename, filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Upload Success!",
		"filename":    file.Filename,
		"document_id": docID,
	})
}

// Chat: ฟังก์ชันคุยกับ AI (เวอร์ชันอัปเกรด ใช้ Username จาก Token)
func (h *DocumentHandler) Chat(c *fiber.Ctx) error {
	var req struct {
		Question   string `json:"question"`
		DocumentID uint   `json:"document_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// --- ดึง Username ที่ Middleware ฝากไว้ใน Locals ---
	// c.Locals("username") ค่านี้มาจากไฟล์ middleware/auth.go
	usernameLocal := c.Locals("username")

	// แปลง Interface{} ให้เป็น string
	username, ok := usernameLocal.(string)
	if !ok || username == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: User identification failed"})
	}

	// เรียก Service โดยส่ง username ไปเป็น sessionID เพื่อให้ AI จำประวัติของคนนี้ได้
	answer, err := h.service.Chat(c.Context(), req.Question, req.DocumentID, username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"answer": answer,
	})
}
