package model

import "time"

// Project 是项目,用于组织场景
type Project struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID        string    `json:"uuid" gorm:"uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"not null;size:128"`
	Description string    `json:"description" gorm:"type:text"`
	Color       string    `json:"color" gorm:"size:16"` // 颜色标签,如 #3370FF
	OwnerID     int64     `json:"owner_id" gorm:"index;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Project) TableName() string {
	return "projects"
}