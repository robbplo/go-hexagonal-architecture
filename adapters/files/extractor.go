package files

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/linkai/go-chatbot-api/domain/model"
	"github.com/linkai/go-chatbot-api/domain/ports"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, fileName, _ string, data []byte) (ports.ExtractedDocument, error) {
	kind, err := model.DetectKnowledgeFileKind(fileName)
	if err != nil {
		return ports.ExtractedDocument{}, err
	}

	var text string
	switch kind {
	case model.KnowledgeFileKindTXT, model.KnowledgeFileKindMD:
		text = normalizeText(string(data))
	case model.KnowledgeFileKindDOCX:
		text, err = extractDOCX(data)
	case model.KnowledgeFileKindPDF:
		text, err = extractPDF(data)
	default:
		err = domainerrors.ErrUnsupportedFileType
	}
	if err != nil {
		return ports.ExtractedDocument{}, fmt.Errorf("extract %s: %w", fileName, err)
	}
	text = normalizeText(text)
	if text == "" {
		return ports.ExtractedDocument{}, &domainerrors.ValidationError{Field: "file", Message: "extracted text is empty"}
	}
	return ports.ExtractedDocument{
		Text: text,
		Kind: kind,
	}, nil
}

func extractDOCX(data []byte) (string, error) {
	readerAt := bytes.NewReader(data)
	reader, err := zip.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return "", err
	}

	var document *zip.File
	for _, file := range reader.File {
		if filepath.Clean(file.Name) == "word/document.xml" {
			document = file
			break
		}
	}
	if document == nil {
		return "", fmt.Errorf("word/document.xml not found")
	}

	handle, err := document.Open()
	if err != nil {
		return "", err
	}
	defer handle.Close()

	decoder := xml.NewDecoder(handle)
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		charData, ok := token.(xml.CharData)
		if !ok {
			continue
		}
		builder.WriteString(string(charData))
		builder.WriteString(" ")
	}
	return builder.String(), nil
}

func extractPDF(data []byte) (string, error) {
	file, err := os.CreateTemp("", "chatbot-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return "", err
	}
	readerFile, reader, err := pdf.Open(file.Name())
	if err != nil {
		return "", err
	}
	defer readerFile.Close()

	var builder strings.Builder
	totalPages := reader.NumPage()
	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		rows, err := page.GetTextByRow()
		if err != nil {
			continue
		}
		for _, row := range rows {
			for _, word := range row.Content {
				builder.WriteString(word.S)
				builder.WriteByte(' ')
			}
			builder.WriteByte('\n')
		}
	}
	return builder.String(), nil
}

func normalizeText(value string) string {
	lines := strings.Split(value, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
