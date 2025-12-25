package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"rag-project/database"
	"rag-project/handlers"
	"rag-project/middleware"
	"rag-project/repositories"
	"rag-project/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY is not set")
	}

	// 1. Setup Database
	database.ConnectDB()

	// 2. Setup Google AI Client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal("Failed to create Gemini client: ", err)
	}
	defer client.Close()

	// 3. Dependency Injection (ต่อท่อการทำงาน)
	// DB -> Repo -> Service -> Handler
	docRepo := repositories.NewDocumentRepository(database.DB)
	ragService := services.NewRagService(docRepo, client)
	docHandler := handlers.NewDocumentHandler(ragService)

	// Setup Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // ตั้ง Limit ไฟล์ Upload ที่ 10MB
	})

	// Enable CORS for frontend (development)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Routes Management ---

	// Public Routes (โซนเปิด: ใครก็เข้าได้)
	// ใช้สำหรับสมัครสมาชิกและเข้าสู่ระบบเพื่อเอา Token
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)

	// Protected Routes (โซนปิด: ต้องมี Token เท่านั้น)
	// สร้าง Group "/api" และแปะ Middleware ตรวจบัตร (Protected) ไว้หน้าประตู
	api := app.Group("/api", middleware.Protected())

	// เวลาเรียกต้องยิงไปที่ /api/upload และ /api/chat
	api.Post("/upload", docHandler.UploadPDF)
	api.Post("/chat", docHandler.Chat)

	// 6. Start Server
	// ตรวจสอบและสร้างโฟลเดอร์ uploads ถ้ายังไม่มี
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}

	fmt.Println("🚀 Server running on port 3000")
	// เริ่มรัน Server
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
