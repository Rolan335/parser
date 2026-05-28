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
)

type mockExtractor struct {
	fn func(path string) (map[string]any, error)
}

func (m *mockExtractor) Extract(path string) (map[string]any, error) { return m.fn(path) }

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

// errReader всегда возвращает ошибку на Read.
type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

func TestParse_Success(t *testing.T) {
	rawFromExif := map[string]any{
		"MIMEType":            "application/pdf",
		"FileType":            "PDF",
		"Title":               "Hello",
		"Producer":            "TestProd",
		"PageCount":           3,
		"Directory":           "/tmp/parser-xxx",
		"SourceFile":          "/tmp/parser-xxx/foo.pdf",
		"FilePermissions":     "-rw-r--r--",
		"FileAccessDate":      "2026:01:01 00:00:00+00:00",
		"FileModifyDate":      "2026:01:01 00:00:00+00:00",
		"FileInodeChangeDate": "2026:01:01 00:00:00+00:00",
		"ExifToolVersion":     12.57,
	}

	var (
		gotPath  string
		gotSaved *domain.FileMetadata
	)
	ex := &mockExtractor{
		fn: func(path string) (map[string]any, error) {
			gotPath = path
			return rawFromExif, nil
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
		t.Errorf("FileName = %q, want foo.pdf", got.FileName)
	}
	if got.SizeBytes != int64(len("fake pdf content")) {
		t.Errorf("SizeBytes = %d, want %d", got.SizeBytes, len("fake pdf content"))
	}
	if got.MimeType != "application/pdf" {
		t.Errorf("MimeType = %q", got.MimeType)
	}
	if got.Format != "PDF" {
		t.Errorf("Format = %q", got.Format)
	}
	if got.Title != "Hello" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.Producer != "TestProd" {
		t.Errorf("Producer = %q", got.Producer)
	}
	if time.Since(got.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt suspicious: %v", got.CreatedAt)
	}

	if filepath.Base(gotPath) != "foo.pdf" {
		t.Errorf("extractor got path = %q, want basename foo.pdf", gotPath)
	}

	for _, k := range systemFields {
		if _, ok := got.Raw[k]; ok {
			t.Errorf("system field %q should have been stripped from raw", k)
		}
	}
	if _, ok := got.Raw["PageCount"]; !ok {
		t.Errorf("non-system field PageCount missing from raw")
	}

	if gotSaved != got {
		t.Errorf("repo.Save received different *FileMetadata than was returned")
	}
}

func TestParse_FallbackFilename(t *testing.T) {
	cases := []string{"", ".", "/"}
	for _, in := range cases {
		t.Run("input="+in, func(t *testing.T) {
			var gotPath string
			ex := &mockExtractor{fn: func(path string) (map[string]any, error) {
				gotPath = path
				return map[string]any{}, nil
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

func TestParse_ExtractorError(t *testing.T) {
	boom := errors.New("exiftool blew up")
	ex := &mockExtractor{fn: func(string) (map[string]any, error) { return nil, boom }}
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
	ex := &mockExtractor{fn: func(string) (map[string]any, error) {
		return map[string]any{"FileType": "PDF"}, nil
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
	ex := &mockExtractor{fn: func(string) (map[string]any, error) {
		t.Fatal("extractor should not be called when reader fails")
		return nil, nil
	}}
	repo := &mockRepo{}

	_, err := New(ex, repo).Parse(context.Background(), "x.pdf", errReader{err: boom})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrapping %v", err, boom)
	}
}

func TestList(t *testing.T) {
	want := []domain.FileMetadata{{ID: 1, FileName: "a"}, {ID: 2, FileName: "b"}}
	wantFilter := domain.ListFilter{MimeType: "application/pdf", Limit: 10}

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

var _ io.Reader = errReader{} // ensure errReader satisfies io.Reader at compile time
