package models

import "gorm.io/gorm"

type ChatHistory struct {
	gorm.Model
	SessionID string `json:"session_id" gorm:"index"` // ใช้แยกห้องแชทของแต่ละคน
	Role      string `json:"role"`                    // "user" หรือ "model"
	Message   string `json:"message"`
}
