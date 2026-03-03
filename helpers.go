package gtfsparser

import (
	"archive/zip"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// zipFileReader wraps both the zip archive and the file inside it,
// closing both with a single Close() call.
type zipFileReader struct {
	zip  *zip.ReadCloser
	file io.ReadCloser
}

func (r *zipFileReader) Read(p []byte) (n int, err error) {
	return r.file.Read(p)
}

func (r *zipFileReader) Close() error {
	r.file.Close()
	return r.zip.Close()
}

// openFileFromZip opens a named file from within a zip archive.
// Returns nil, nil if the file does not exist (caller treats it as optional).
// Returns an io.ReadCloser that closes both the file and the zip on Close().
func openFileFromZip(zipPath, fileName string) (io.ReadCloser, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip %s: %w", zipPath, err)
	}

	for _, f := range r.File {
		if f.Name == fileName {
			rc, err := f.Open()
			if err != nil {
				r.Close()
				return nil, fmt.Errorf("opening %s in zip: %w", fileName, err)
			}
			return &zipFileReader{zip: r, file: rc}, nil
		}
	}

	r.Close()
	return nil, nil // file not found — caller decides if optional or required
}

func sanitizeHeaders(headers []string) []string {
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\uFEFF")
	}
	return headers
}

func getCol(row []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

func parseOptionalFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
