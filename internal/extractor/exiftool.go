package extractor

import (
	"fmt"
	"sync"

	"github.com/barasher/go-exiftool"
)

type Extractor interface {
	Extract(path string) (map[string]any, error)
}

// ExifExtractor использует exiftool в режиме stay_open, мьютекс сериализует доступ к stdin/stdout.
type ExifExtractor struct {
	mu sync.Mutex
	et *exiftool.Exiftool
}

func New() (*ExifExtractor, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, fmt.Errorf("init exiftool: %w", err)
	}
	return &ExifExtractor{et: et}, nil
}

func (e *ExifExtractor) Close() error {
	return e.et.Close()
}

func (e *ExifExtractor) Extract(path string) (map[string]any, error) {
	e.mu.Lock()
	res := e.et.ExtractMetadata(path)
	e.mu.Unlock()

	if len(res) == 0 {
		return nil, fmt.Errorf("no metadata returned for %s", path)
	}
	if res[0].Err != nil {
		return nil, res[0].Err
	}
	return res[0].Fields, nil
}
