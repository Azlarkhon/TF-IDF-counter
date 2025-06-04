package models

import "time"

type Collection struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `gorm:"size:100;not null" json:"name"`
	UserID    int         `gorm:"not null" json:"user_id"`
	User      User        `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Documents []*Document `gorm:"many2many:collection_documents;" json:"documents"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
