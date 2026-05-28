package handler

import "time"

type parseMetadata struct {
	FileName  string         `json:"file_name"`
	SizeBytes int64          `json:"size_bytes"`
	MimeType  string         `json:"mime_type"`
	Format    string         `json:"format"`
	Title     string         `json:"title,omitempty"`
	Producer  string         `json:"producer,omitempty"`
	Raw       map[string]any `json:"raw"`
	CreatedAt time.Time      `json:"created_at"`
}

type parseResponse struct {
	parseMetadata
	File attachedFile `json:"file"`
}

type attachedFile struct {
	Name          string `json:"name"`
	ContentBase64 string `json:"content_base64"`
}

type listQuery struct {
	FileName string     `form:"file_name"`
	MimeType string     `form:"mime_type"`
	Format   string     `form:"format"`
	From     *time.Time `form:"from"`
	To       *time.Time `form:"to"`
	Limit    int        `form:"limit"  binding:"gte=0"`
	Offset   int        `form:"offset" binding:"gte=0"`
}
