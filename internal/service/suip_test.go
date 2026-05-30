package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Rolan335/parser/internal/domain"
	"github.com/Rolan335/parser/internal/extractor"
)

type mockExtractor struct {
	fn func(path string) (extractor.Result, error)
}

func (m *mockExtractor) Extract(path string) (extractor.Result, error) { return m.fn(path) }

type mockRepo struct {
	saveFn func(ctx context.Context, m *domain.FileMetadata) (int64, error)
	listFn func(ctx context.Context, f domain.ListFilter) ([]domain.FileMetadata, error)
}

func (m *mockRepo) Save(ctx context.Context, fm *domain.FileMetadata) (int64, error) {
	return m.saveFn(ctx, fm)
}
func (m *mockRepo) List(ctx context.Context, f domain.ListFilter) ([]domain.FileMetadata, error) {
	return m.listFn(ctx, f)
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

func result(fields map[string]string, html string) extractor.Result {
	return extractor.Result{Fields: fields, HTMLRaw: []byte(html)}
}

func TestParse_Success(t *testing.T) {
	fields := map[string]string{
		"Producer":        "cairo 1.16.0",
		"Title":           "Hello",
		"CreationDate":    "Thu Feb 26 19:47:07 2026 CET",
		"Pages":           "4",
		"PDF version":     "1.5",
		"Page size":       "595.276 x 841.89 pts (A4)",
		"Page rot":        "0",
		"Form":            "none",
		"Encrypted":       "no",
		"Optimized":       "no",
		"Tagged":          "yes",
		"JavaScript":      "no",
		"Custom Metadata": "no",
		"Metadata Stream": "yes",
		"UserProperties":  "no",
		"Suspects":        "no",
		"File size":       "173444 bytes",
	}
	htmlSnippet := `<textarea class="form-control textarea">Producer: cairo</textarea>`

	var (
		gotPath  string
		gotSaved *domain.FileMetadata
	)
	ex := &mockExtractor{
		fn: func(path string) (extractor.Result, error) {
			gotPath = path
			return result(fields, htmlSnippet), nil
		},
	}
	repo := &mockRepo{
		saveFn: func(_ context.Context, m *domain.FileMetadata) (int64, error) {
			gotSaved = m
			return 42, nil
		},
	}

	body := bytes.NewBufferString("fake pdf content")
	got, err := New(ex, repo).Parse(context.Background(), "foo.pdf", body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
	if got.FileName != "foo.pdf" {
		t.Errorf("FileName = %q", got.FileName)
	}
	if got.SizeBytes != 173444 {
		t.Errorf("SizeBytes = %d", got.SizeBytes)
	}
	if got.Producer != "cairo 1.16.0" {
		t.Errorf("Producer = %q", got.Producer)
	}
	if got.Title != "Hello" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.CreationDate != "Thu Feb 26 19:47:07 2026 CET" {
		t.Errorf("CreationDate = %q", got.CreationDate)
	}
	if got.Pages != 4 {
		t.Errorf("Pages = %d", got.Pages)
	}
	if got.PDFVersion != "1.5" {
		t.Errorf("PDFVersion = %q", got.PDFVersion)
	}
	if got.PageSize != "595.276 x 841.89 pts (A4)" {
		t.Errorf("PageSize = %q", got.PageSize)
	}
	if got.Form != "none" {
		t.Errorf("Form = %q", got.Form)
	}
	if got.Encrypted || got.Optimized || got.JavaScript || got.CustomMetadata || got.UserProperties || got.Suspects {
		t.Errorf("ожидал yes-флаги только у Tagged и MetadataStream, got = %+v", got)
	}
	if !got.Tagged || !got.MetadataStream {
		t.Errorf("Tagged/MetadataStream должны быть true")
	}
	if string(got.RawHTML) != htmlSnippet {
		t.Errorf("RawHTML = %q, want %q", got.RawHTML, htmlSnippet)
	}
	if time.Since(got.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt suspicious: %v", got.CreatedAt)
	}
	if filepath.Base(gotPath) != "foo.pdf" {
		t.Errorf("extractor path = %q", gotPath)
	}
	if gotSaved != got {
		t.Errorf("repo.Save received different *FileMetadata than was returned")
	}
}

func TestParse_FallbackFilename(t *testing.T) {
	for _, in := range []string{"", ".", "/"} {
		t.Run("input="+in, func(t *testing.T) {
			var gotPath string
			ex := &mockExtractor{fn: func(path string) (extractor.Result, error) {
				gotPath = path
				return result(map[string]string{}, ""), nil
			}}
			repo := &mockRepo{saveFn: func(context.Context, *domain.FileMetadata) (int64, error) {
				return 1, nil
			}}
			_, err := New(ex, repo).Parse(context.Background(), in, bytes.NewBufferString("x"))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if filepath.Base(gotPath) != "upload" {
				t.Errorf("fallback name = %q, want upload", filepath.Base(gotPath))
			}
		})
	}
}

func TestParse_SizeFallbackToUpload(t *testing.T) {
	ex := &mockExtractor{fn: func(string) (extractor.Result, error) {
		return result(map[string]string{"PDF version": "1.7"}, ""), nil
	}}
	repo := &mockRepo{saveFn: func(context.Context, *domain.FileMetadata) (int64, error) {
		return 1, nil
	}}

	upload := "fake pdf content"
	got, err := New(ex, repo).Parse(context.Background(), "x.pdf", bytes.NewBufferString(upload))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.SizeBytes != int64(len(upload)) {
		t.Errorf("SizeBytes = %d, want upload size %d", got.SizeBytes, len(upload))
	}
}

func TestParse_ExtractorError(t *testing.T) {
	boom := errors.New("pdfyeah blew up")
	ex := &mockExtractor{fn: func(string) (extractor.Result, error) { return extractor.Result{}, boom }}
	repo := &mockRepo{saveFn: func(context.Context, *domain.FileMetadata) (int64, error) {
		t.Fatal("repo.Save should not be called when extractor fails")
		return 0, nil
	}}

	_, err := New(ex, repo).Parse(context.Background(), "x.pdf", bytes.NewBufferString("data"))
	if !errors.Is(err, ErrExtract) {
		t.Errorf("err = %v, want wrapping ErrExtract", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("inner error not preserved through wrap: %v", err)
	}
}

func TestParse_RepoError(t *testing.T) {
	boom := errors.New("db down")
	ex := &mockExtractor{fn: func(string) (extractor.Result, error) {
		return result(map[string]string{"PDF version": "1.5"}, ""), nil
	}}
	repo := &mockRepo{saveFn: func(context.Context, *domain.FileMetadata) (int64, error) {
		return 0, boom
	}}

	_, err := New(ex, repo).Parse(context.Background(), "x.pdf", bytes.NewBufferString("data"))
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestParse_ReaderError(t *testing.T) {
	boom := errors.New("read failed")
	ex := &mockExtractor{fn: func(string) (extractor.Result, error) {
		t.Fatal("extractor should not be called when reader fails")
		return extractor.Result{}, nil
	}}
	repo := &mockRepo{}

	_, err := New(ex, repo).Parse(context.Background(), "x.pdf", errReader{err: boom})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrapping %v", err, boom)
	}
}

func TestList(t *testing.T) {
	want := []domain.FileMetadata{{ID: 1, FileName: "a"}, {ID: 2, FileName: "b"}}
	wantFilter := domain.ListFilter{Producer: "cairo 1.16.0", Limit: 10}

	var gotFilter domain.ListFilter
	repo := &mockRepo{listFn: func(_ context.Context, f domain.ListFilter) ([]domain.FileMetadata, error) {
		gotFilter = f
		return want, nil
	}}

	got, err := New(nil, repo).List(context.Background(), wantFilter)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(gotFilter, wantFilter) {
		t.Errorf("filter not passed through: got %+v want %+v", gotFilter, wantFilter)
	}
}

func TestList_Error(t *testing.T) {
	boom := errors.New("query failed")
	repo := &mockRepo{listFn: func(context.Context, domain.ListFilter) ([]domain.FileMetadata, error) {
		return nil, boom
	}}

	_, err := New(nil, repo).List(context.Background(), domain.ListFilter{})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

var _ io.Reader = errReader{}
