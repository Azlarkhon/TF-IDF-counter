package dto

import "time"

type DocumentResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Content     string    `json:"content"`
	UplodadedAt time.Time `json:"uploaded_at"`
}
