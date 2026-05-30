package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	uploadSize, err := io.Copy(tmp, src)
	tmp.Close()
	if err != nil {
		return nil, fmt.Errorf("write temp: %w", err)
	}

	res, err := s.ex.Extract(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExtract, err)
	}
	f := res.Fields

	size := parseFileSize(f["File size"])
	if size == 0 {
		size = uploadSize
	}

	m := &domain.FileMetadata{
		FileName:       fileName,
		SizeBytes:      size,
		Producer:       f["Producer"],
		Title:          f["Title"],
		CreationDate:   f["CreationDate"],
		Pages:          parseInt(f["Pages"]),
		PDFVersion:     f["PDF version"],
		PageSize:       f["Page size"],
		PageRot:        parseInt(f["Page rot"]),
		Form:           f["Form"],
		Encrypted:      yes(f["Encrypted"]),
		Optimized:      yes(f["Optimized"]),
		Tagged:         yes(f["Tagged"]),
		JavaScript:     yes(f["JavaScript"]),
		CustomMetadata: yes(f["Custom Metadata"]),
		MetadataStream: yes(f["Metadata Stream"]),
		UserProperties: yes(f["UserProperties"]),
		Suspects:       yes(f["Suspects"]),
		RawHTML:        res.HTMLRaw,
		CreatedAt:      time.Now().UTC(),
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

func yes(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "yes")
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseFileSize(s string) int64 {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, ' '); i > 0 {
		s = s[:i]
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
