package model

import "time"

// User 是用户模型
// MVP 阶段桌面应用通常单用户,但保留多用户能力以便未来扩展
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID         string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;size:64"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;size:128"`
	PasswordHash string    `json:"-" gorm:"size:128"` // 不暴露到 JSON
	DisplayName  string    `json:"display_name" gorm:"size:128"`
	Role         string    `json:"role" gorm:"size:32;default:'user'"` // admin / user
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}