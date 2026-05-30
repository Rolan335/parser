package extractor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Result struct {
	Fields  map[string]string
	HTMLRaw []byte
}

type Extractor interface {
	Extract(path string) (Result, error)
}

const pdfyeahURL = "https://www.pdfyeah.com/view-pdf-metadata"

var textareaRe = regexp.MustCompile(`(?s)<textarea[^>]*class="form-control textarea"[^>]*>(.*?)</textarea>`)

type PdfyeahExtractor struct {
	client *http.Client
}

func New() *PdfyeahExtractor {
	return &PdfyeahExtractor{client: &http.Client{Timeout: 60 * time.Second}}
}

func (e *PdfyeahExtractor) Extract(path string) (Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return Result{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile("upfile", filepath.Base(path))
	if err != nil {
		return Result{}, fmt.Errorf("multipart: %w", err)
	}
	if _, err := io.Copy(fw, f); err != nil {
		return Result{}, fmt.Errorf("copy upload: %w", err)
	}
	if err := mw.WriteField("submitfile", "Read PDF Metadata"); err != nil {
		return Result{}, fmt.Errorf("multipart submitfile: %w", err)
	}
	if err := mw.Close(); err != nil {
		return Result{}, fmt.Errorf("multipart close: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pdfyeahURL, body)
	if err != nil {
		return Result{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("User-Agent", "Mozilla/5.0 parser")

	resp, err := e.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("pdfyeah request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return Result{}, fmt.Errorf("pdfyeah status %d", resp.StatusCode)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}
	return parseHTML(html)
}

func parseHTML(html []byte) (Result, error) {
	m := textareaRe.FindSubmatch(html)
	if len(m) < 2 {
		return Result{}, fmt.Errorf("metadata textarea not found in response")
	}

	fields := make(map[string]string)
	sc := bufio.NewScanner(strings.NewReader(string(m[1])))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		fields[key] = val
	}
	return Result{Fields: fields, HTMLRaw: append([]byte(nil), m[0]...)}, nil
}
