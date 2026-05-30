package handler

import (
	"time"

	"github.com/Rolan335/parser/internal/domain"
)

type parseResponse struct {
	domain.FileMetadata
	File attachedFile `json:"file"`
}

type attachedFile struct {
	Name          string `json:"name"`
	ContentBase64 string `json:"content_base64"`
}

type listQuery struct {
	FileName   string     `form:"file_name"`
	Producer   string     `form:"producer"`
	PDFVersion string     `form:"pdf_version"`
	From       *time.Time `form:"from"`
	To         *time.Time `form:"to"`
	Limit      int        `form:"limit"  binding:"gte=0"`
	Offset     int        `form:"offset" binding:"gte=0"`
}
