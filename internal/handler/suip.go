package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Rolan335/parser/internal/domain"
	"github.com/Rolan335/parser/internal/service"
)

const maxUploadBytes = 100 << 20

type Handler struct {
	svc *service.Service
	log *slog.Logger
}

func New(svc *service.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Parse(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := fh.Open()
	if err != nil {
		h.log.Error("open uploaded file", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read upload"})
		return
	}
	defer src.Close()

	name := fh.Filename
	m, err := h.svc.Parse(c.Request.Context(), name, src)
	if err != nil {
		if errors.Is(err, service.ErrExtract) {
			h.log.Warn("extract failed", "err", err, "file", name)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "cannot extract metadata"})
			return
		}
		h.log.Error("parse failed", "err", err, "file", name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	meta := parseMetadata{
		FileName:  m.FileName,
		SizeBytes: m.SizeBytes,
		MimeType:  m.MimeType,
		Format:    m.Format,
		Title:     m.Title,
		Producer:  m.Producer,
		Raw:       m.Raw,
		CreatedAt: m.CreatedAt,
	}
	body, err := json.Marshal(meta)
	if err != nil {
		h.log.Error("marshal response", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, parseResponse{
		parseMetadata: meta,
		File: attachedFile{
			Name:          fmt.Sprintf("%s.meta.json", name),
			ContentBase64: base64.StdEncoding.EncodeToString(body),
		},
	})
}

func (h *Handler) List(c *gin.Context) {
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.svc.List(c.Request.Context(), domain.ListFilter(q))
	if err != nil {
		h.log.Error("list failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}
