package database

import (
	"fmt"
	"log"
	"rag-project/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// แก้ Port เป็น 5433 ตามที่เราคุยกัน
	dsn := "host=localhost user=postgres password=password dbname=rag_db port=5433 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("เชื่อมต่อ Database พัง! : ", err)
	}

	fmt.Println("✅ Connected to Database!")

	// เปิดใช้งาน Vector Extension
	DB.Exec("CREATE EXTENSION IF NOT EXISTS vector")

	// สร้าง Table
	err = DB.AutoMigrate(&models.Document{}, &models.DocumentChunk{}, &models.ChatHistory{}, &models.User{})
	if err != nil {
		log.Fatal("สร้าง Table ไม่ได้ : ", err)
	}
}
