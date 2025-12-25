package models

import (
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// Document: เก็บข้อมูลไฟล์ต้นฉบับ (ชื่อไฟล์, วันที่อัปโหลด)
type Document struct {
	gorm.Model
	Filename string `json:"filename"`
}

// DocumentChunk: เก็บเนื้อหาที่ถูกหั่นแล้ว + ค่า Vector
type DocumentChunk struct {
	gorm.Model
	DocumentID uint   `json:"document_id"`
	Content    string `json:"content"`     // เนื้อหา text ที่ตัดมา
	PageNumber int    `json:"page_number"` // เพิ่มบรรทัดนี้
	// สำคัญ: Google Gemini รุ่น text-embedding-004 ใช้ขนาด 768
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"embedding"`
}
