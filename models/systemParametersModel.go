package models

import "time"

type Metric struct {
	ID                           uint      `gorm:"primaryKey" json:"id"`
	FilesProcessed               uint      `gorm:"default:0" json:"files_processed"`
	LatestFileProcessedTimestamp time.Time `json:"latest_file_processed_timestamp"`
}
