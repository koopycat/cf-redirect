// Package csvio implements the two-column redirect interchange format.
package csvio

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/koopycat/cf-redirect/internal/domain"
)

var header = []string{"source", "target"}

// Read accepts comma- or semicolon-separated rows with exactly two fields.
// A source,target header is optional and must use the same separator as the data.
func Read(r io.Reader) ([]domain.Redirect, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("CSV is empty; expected source and target rows")
	}

	commaRedirects, commaErr := readDelimited(data, ',')
	if commaErr == nil {
		return commaRedirects, nil
	}
	semicolonRedirects, semicolonErr := readDelimited(data, ';')
	if semicolonErr == nil {
		return semicolonRedirects, nil
	}
	return nil, fmt.Errorf("invalid comma- or semicolon-separated CSV (comma: %v; semicolon: %v)", commaErr, semicolonErr)
}

func readDelimited(data []byte, comma rune) ([]domain.Redirect, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = comma
	reader.FieldsPerRecord = 2

	first, err := reader.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("CSV is empty; expected source and target rows")
	}
	if err != nil {
		return nil, fmt.Errorf("read CSV row 1: %w", err)
	}

	var redirects []domain.Redirect
	seen := make(map[string]struct{})
	if first[0] == header[0] && first[1] == header[1] {
		// The header is optional and is not an import row.
	} else if err := appendRedirect(&redirects, seen, first, 1); err != nil {
		return nil, err
	}

	row := 2
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return redirects, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row %d: %w", row, err)
		}
		if err := appendRedirect(&redirects, seen, record, row); err != nil {
			return nil, err
		}
		row++
	}
}

func appendRedirect(redirects *[]domain.Redirect, seen map[string]struct{}, record []string, row int) error {
	redirect := domain.New(record[0], record[1])
	if err := redirect.Validate(); err != nil {
		return fmt.Errorf("CSV row %d: %w", row, err)
	}
	if _, exists := seen[redirect.Source]; exists {
		return fmt.Errorf("CSV row %d: duplicate source %q", row, redirect.Source)
	}
	seen[redirect.Source] = struct{}{}
	*redirects = append(*redirects, redirect)
	return nil
}

func Write(w io.Writer, redirects []domain.Redirect) error {
	writer := csv.NewWriter(w)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	for _, redirect := range redirects {
		if err := writer.Write([]string{redirect.Source, redirect.Target}); err != nil {
			return fmt.Errorf("write CSV row for %q: %w", redirect.Source, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}
	return nil
}
