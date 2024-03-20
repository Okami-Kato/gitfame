package output

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/Okami-Kato/gitfame/internal/domain"
)

type Format string

const (
	Tabular   Format = "tabular"
	CSV       Format = "csv"
	JSON      Format = "json"
	JSONLines Format = "json-lines"
)

type Writer interface {
	Write([]domain.FameEntry) error
}

var ErrUnsupportedFormat = errors.New("unsupported format")

func NewWriter(format Format, w io.Writer) (Writer, error) {
	var writer Writer
	switch format {
	case Tabular:
		writer = NewTabularWriter(w)
	case CSV:
		writer = NewCSVWriter(w)
	case JSON:
		writer = NewJSONWriter(w)
	case JSONLines:
		writer = NewJSONLinesWriter(w)
	default:
		return nil, ErrUnsupportedFormat
	}
	return writer, nil
}

type TabularWriter struct {
	writer *tabwriter.Writer
}

func NewTabularWriter(w io.Writer) Writer {
	return &TabularWriter{
		writer: tabwriter.NewWriter(w, 0, 0, 1, ' ', 0),
	}
}

func (w *TabularWriter) Write(arr []domain.FameEntry) error {
	_, err := fmt.Fprintf(w.writer, "%s\t%s\t%s\t%s\n", "Name", "Lines", "Commits", "Files")
	if err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	for _, entry := range arr {
		_, err = fmt.Fprintf(w.writer, "%s\t%d\t%d\t%d\n", entry.Name, entry.Lines, entry.Commits, entry.Files)
		if err != nil {
			return fmt.Errorf("error writing entry: %w", err)
		}
	}
	err = w.writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing: %w", err)
	}
	return nil
}

type CSVWriter struct {
	writer *csv.Writer
}

func NewCSVWriter(w io.Writer) Writer {
	return &CSVWriter{csv.NewWriter(w)}
}

func (w *CSVWriter) Write(arr []domain.FameEntry) error {
	if err := w.writer.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
		return fmt.Errorf("error writing header to csv: %w", err)
	}
	for _, entry := range arr {
		if err := w.writer.Write(toRecord(&entry)); err != nil {
			return fmt.Errorf("error writing record to csv: %w", err)
		}
	}
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return fmt.Errorf("error flushing to csv: %w", err)
	}
	return nil
}

func toRecord(e *domain.FameEntry) []string {
	return []string{
		e.Name,
		strconv.Itoa(e.Lines),
		strconv.Itoa(e.Commits),
		strconv.Itoa(e.Files),
	}
}

type JSONWriter struct {
	encoder *json.Encoder
}

func NewJSONWriter(w io.Writer) Writer {
	return &JSONWriter{json.NewEncoder(w)}
}

func (w *JSONWriter) Write(arr []domain.FameEntry) error {
	if err := w.encoder.Encode(arr); err != nil {
		return fmt.Errorf("error encoding slice of entries: %w", err)
	}
	return nil
}

type JSONLinesWriter struct {
	encoder *json.Encoder
}

func NewJSONLinesWriter(w io.Writer) Writer {
	return &JSONLinesWriter{json.NewEncoder(w)}
}

func (w *JSONLinesWriter) Write(arr []domain.FameEntry) error {
	for _, entry := range arr {
		if err := w.encoder.Encode(entry); err != nil {
			return fmt.Errorf("error encoding entry: %w", err)
		}
	}
	return nil
}
