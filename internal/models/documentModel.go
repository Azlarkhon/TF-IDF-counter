package models

import "time"

type Document struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	Name        string        `gorm:"size:100;not null" json:"name"` // имя файла или произвольное название
	FilePath    string        `gorm:"not null" json:"-"`             // путь до файла на диске
	UserID      int           `gorm:"not null" json:"user_id"`
	User        User          `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Collections []*Collection `gorm:"many2many:collection_documents;" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
