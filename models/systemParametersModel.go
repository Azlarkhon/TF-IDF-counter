package models

import "time"

type Metric struct {
	ID                           uint      `gorm:"primaryKey" json:"id"`
	FilesProcessed               uint      `gorm:"default:0" json:"files_processed"`
	LatestFileProcessedTimestamp time.Time `json:"latest_file_processed_timestamp"`
	MinTimeProcessed             float64   `gorm:"type:decimal(10,3);default:1000.0" json:"min_time_processed"`
	AvgTimeProcessed             float64   `gorm:"type:decimal(10,3);default:0.0" json:"avg_time_processed"`
	MaxTimeProcessed             float64   `gorm:"type:decimal(10,3);default:0.0" json:"max_time_processed"`
	TotalFileSizeMB              float64   `gorm:"type:decimal(10,3);default:0.0" json:"total_file_size_mb"`
	AvgFileSizeMB                float64   `gorm:"type:decimal(10,3);default:0.0" json:"avg_file_size_mb"`
	WordStats                    []Word    `gorm:"foreignKey:MetricID" json:"word_stats"`
}

type Word struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MetricID  uint      `json:"metric_id"`
	Word      string    `gorm:"size:255" json:"word"`
	Count     int      `gorm:"default:0" json:"count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
