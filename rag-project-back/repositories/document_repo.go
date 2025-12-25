package repositories

import (
	"rag-project/models"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	CreateDocument(doc *models.Document) error
	CreateChunks(chunks []models.DocumentChunk) error
	SearchChunks(embedding pgvector.Vector, docID uint) ([]models.DocumentChunk, error)

	// เพิ่มฟังก์ชันสำหรับจัดการ Chat History
	SaveChatHistory(sessionID, role, message string) error
	GetChatHistory(sessionID string, limit int) ([]models.ChatHistory, error)
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) CreateDocument(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *documentRepository) CreateChunks(chunks []models.DocumentChunk) error {
	return r.db.Create(&chunks).Error
}

func (r *documentRepository) SearchChunks(embedding pgvector.Vector, docID uint) ([]models.DocumentChunk, error) {
	var chunks []models.DocumentChunk

	// ค้นหาโดยกรองเฉพาะ document_id ที่ระบุ และเรียงตามระยะห่างของ Vector (Cosine Distance)
	err := r.db.Where("document_id = ?", docID).
		Order(gorm.Expr("embedding <=> ?", embedding)).
		Limit(3). // ดึงมา 3 ก้อนที่เหมือนที่สุด
		Find(&chunks).Error

	return chunks, err
}

// ---  Chat History ---

func (r *documentRepository) SaveChatHistory(sessionID, role, message string) error {
	return r.db.Create(&models.ChatHistory{
		SessionID: sessionID,
		Role:      role,
		Message:   message,
	}).Error
}

func (r *documentRepository) GetChatHistory(sessionID string, limit int) ([]models.ChatHistory, error) {
	var histories []models.ChatHistory

	// ดึงประวัติล่าสุดตาม limit ที่กำหนด (เรียงจากใหม่ไปเก่า แล้วค่อยไปกลับด้านใน Service เอา)
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at desc").
		Limit(limit).
		Find(&histories).Error

	return histories, err
}
