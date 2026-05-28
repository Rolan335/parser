package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Rolan335/parser/internal/domain"
	"github.com/Rolan335/parser/internal/extractor"
	"github.com/Rolan335/parser/internal/repository"
)

type Service struct {
	ex   extractor.Extractor
	repo repository.Repository
}

func New(ex extractor.Extractor, repo repository.Repository) *Service {
	return &Service{ex: ex, repo: repo}
}

func (s *Service) Parse(ctx context.Context, fileName string, src io.Reader) (*domain.FileMetadata, error) {
	dir, err := os.MkdirTemp("", "parser-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	name := filepath.Base(fileName)
	if name == "" || name == "." || name == "/" {
		name = "upload"
	}
	path := filepath.Join(dir, name)

	tmp, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	size, err := io.Copy(tmp, src)
	tmp.Close()
	if err != nil {
		return nil, fmt.Errorf("write temp: %w", err)
	}

	raw, err := s.ex.Extract(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExtract, err)
	}
	for _, k := range systemFields {
		delete(raw, k)
	}

	mimeType, _ := raw["MIMEType"].(string)
	format, _ := raw["FileType"].(string)
	title, _ := raw["Title"].(string)
	producer, _ := raw["Producer"].(string)

	m := &domain.FileMetadata{
		FileName:  fileName,
		SizeBytes: size,
		MimeType:  mimeType,
		Format:    format,
		Title:     title,
		Producer:  producer,
		Raw:       raw,
		CreatedAt: time.Now().UTC(),
	}

	id, err := s.repo.Save(ctx, m)
	if err != nil {
		return nil, err
	}
	m.ID = id
	return m, nil
}

func (s *Service) List(ctx context.Context, f domain.ListFilter) ([]domain.FileMetadata, error) {
	return s.repo.List(ctx, f)
}
