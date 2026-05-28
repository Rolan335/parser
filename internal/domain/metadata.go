package domain

import "time"

// FileMetadata пишется в БД и отдаётся клиенту.
type FileMetadata struct {
	ID        int64          `json:"id"`
	FileName  string         `json:"file_name"`
	SizeBytes int64          `json:"size_bytes"`
	MimeType  string         `json:"mime_type"`
	Format    string         `json:"format"`
	Title     string         `json:"title,omitempty"`
	Producer  string         `json:"producer,omitempty"`
	Raw       map[string]any `json:"raw"`
	CreatedAt time.Time      `json:"created_at"`
}

// ListFilter параметры выборки для GET /suip-data.
type ListFilter struct {
	FileName string
	MimeType string
	Format   string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}
