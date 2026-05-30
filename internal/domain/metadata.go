package domain

import "time"

type FileMetadata struct {
	ID             int64     `json:"id"`
	FileName       string    `json:"file_name"`
	SizeBytes      int64     `json:"size_bytes"`
	Producer       string    `json:"producer,omitempty"`
	Title          string    `json:"title,omitempty"`
	CreationDate   string    `json:"creation_date,omitempty"`
	Pages          int       `json:"pages"`
	PDFVersion     string    `json:"pdf_version,omitempty"`
	PageSize       string    `json:"page_size,omitempty"`
	PageRot        int       `json:"page_rot"`
	Form           string    `json:"form,omitempty"`
	Encrypted      bool      `json:"encrypted"`
	Optimized      bool      `json:"optimized"`
	Tagged         bool      `json:"tagged"`
	JavaScript     bool      `json:"javascript"`
	CustomMetadata bool      `json:"custom_metadata"`
	MetadataStream bool      `json:"metadata_stream"`
	UserProperties bool      `json:"user_properties"`
	Suspects       bool      `json:"suspects"`
	RawHTML        []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListFilter struct {
	FileName   string
	Producer   string
	PDFVersion string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}
