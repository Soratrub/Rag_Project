package services

import (
	"context"
	"fmt"
	"os"
	"rag-project/models"
	"rag-project/repositories"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/ledongthuc/pdf"
	"github.com/pgvector/pgvector-go"
)

type PageContent struct {
	PageNumber int
	Content    string
}

type RagService interface {
	ProcessPDF(ctx context.Context, filename string, filePath string) (uint, error)
	Chat(ctx context.Context, question string, docID uint, sessionID string) (string, error)
}

type ragService struct {
	repo     repositories.DocumentRepository
	aiClient *genai.Client
}

func NewRagService(repo repositories.DocumentRepository, aiClient *genai.Client) RagService {
	return &ragService{
		repo:     repo,
		aiClient: aiClient,
	}
}

func (s *ragService) ProcessPDF(ctx context.Context, filename string, filePath string) (uint, error) {
	defer os.Remove(filePath)

	pages, err := readPDFWithPages(filePath)
	if err != nil {
		return 0, fmt.Errorf("อ่านไฟล์ PDF ไม่ได้: %v", err)
	}

	doc := models.Document{Filename: filename}
	if err := s.repo.CreateDocument(&doc); err != nil {
		return 0, fmt.Errorf("บันทึกไฟล์ลง DB ไม่ได้: %v", err)
	}

	emModel := s.aiClient.EmbeddingModel("text-embedding-004")
	var docChunks []models.DocumentChunk

	for _, page := range pages {
		chunks := splitText(page.Content, 1000) // ใช้ Smart Chunking

		for _, chunkText := range chunks {
			res, err := emModel.EmbedContent(ctx, genai.Text(chunkText))
			if err != nil {
				continue
			}

			vectorData := pgvector.NewVector(res.Embedding.Values)
			docChunks = append(docChunks, models.DocumentChunk{
				DocumentID: doc.ID,
				Content:    chunkText,
				PageNumber: page.PageNumber, // บันทึกเลขหน้า
				Embedding:  vectorData,
			})
		}
	}

	if len(docChunks) > 0 {
		if err := s.repo.CreateChunks(docChunks); err != nil {
			return 0, fmt.Errorf("บันทึก Chunks ไม่ได้: %v", err)
		}
	}

	return doc.ID, nil
}

func (s *ragService) Chat(ctx context.Context, question string, docID uint, sessionID string) (string, error) {
	// 1. ดึงประวัติการคุยเก่าๆ (History)
	histories, _ := s.repo.GetChatHistory(sessionID, 5) // ดึง 5 ข้อความล่าสุด
	historyContext := ""
	for i := len(histories) - 1; i >= 0; i-- {
		historyContext += fmt.Sprintf("%s: %s\n", histories[i].Role, histories[i].Message)
	}

	// 2. แปลงคำถามเป็น Vector
	emModel := s.aiClient.EmbeddingModel("text-embedding-004")
	res, err := emModel.EmbedContent(ctx, genai.Text(question))
	if err != nil {
		return "", fmt.Errorf("Embedding Error: %v", err)
	}
	queryVector := pgvector.NewVector(res.Embedding.Values)

	// 3. ค้นหาเนื้อหาที่เกี่ยวข้อง (Search)
	chunks, err := s.repo.SearchChunks(queryVector, docID)
	if err != nil {
		return "", fmt.Errorf("Search Error: %v", err)
	}

	if len(chunks) == 0 {
		return "ไม่พบข้อมูลในเอกสารฉบับนี้ครับ", nil
	}

	// 4. สร้าง Context พร้อมเลขหน้า
	contextText := ""
	for _, chunk := range chunks {
		contextText += fmt.Sprintf("(หน้า %d): %s\n---\n", chunk.PageNumber, chunk.Content)
	}

	// 5. สร้าง Prompt (รวม History + Context + Question)
	model := s.aiClient.GenerativeModel("gemini-2.5-flash")
	prompt := fmt.Sprintf("ประวัติการสนทนา:\n%s\n\nข้อมูลอ้างอิงจากเอกสาร:\n%s\nคำถาม: %s\nตอบ (กรุณาระบุเลขหน้าอ้างอิงถ้ามี):", historyContext, contextText, question)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	answer := "No response"
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		answer = fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
	}

	// 6. บันทึกบทสนทนาลง DB
	s.repo.SaveChatHistory(sessionID, "user", question)
	s.repo.SaveChatHistory(sessionID, "model", answer)

	return answer, nil
}

// --- Helper Functions ---

func readPDFWithPages(path string) ([]PageContent, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pages []PageContent
	totalPage := r.NumPage()

	for i := 1; i <= totalPage; i++ {
		p := r.Page(i)
		text, _ := p.GetPlainText(nil)
		if text != "" {
			pages = append(pages, PageContent{PageNumber: i, Content: text})
		}
	}
	return pages, nil
}

func splitText(text string, chunkSize int) []string {
	var chunks []string
	paragraphs := strings.Split(text, "\n\n") // แยกย่อหน้า

	currentChunk := ""
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		if len(currentChunk)+len(para) < chunkSize {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += para
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			currentChunk = para
			// ถ้าย่อหน้าเดียวยาวจัด ให้หั่นย่อย
			if len(para) > chunkSize {
				runes := []rune(para)
				for i := 0; i < len(runes); i += chunkSize {
					end := i + chunkSize
					if end > len(runes) {
						end = len(runes)
					}
					chunks = append(chunks, string(runes[i:end]))
				}
				currentChunk = ""
			}
		}
	}
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}
	return chunks
}
