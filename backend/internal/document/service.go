package document

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	pdf "github.com/ledongthuc/pdf"
	rscpdf "rsc.io/pdf"
)

var (
	ErrFileRequired = errors.New("file is required")
	ErrFileTooLarge = errors.New("file is too large")
	ErrInvalidType  = errors.New("only PDF and TXT files are allowed")
	ErrNoText       = errors.New("file does not contain readable text")
	ErrUnsafeName   = errors.New("file name is not safe")
)

type Store interface {
	CreateDocument(ctx context.Context, params CreateDocumentParams, chunks []Chunk) (Document, error)
	ListDocuments(ctx context.Context, userID string) ([]Document, error)
}

type EmbeddingIndexer interface {
	IndexMissingEmbeddings(ctx context.Context, limit int) (int, error)
}

type CreateDocumentParams struct {
	UserID       string
	OriginalName string
	StoredName   string
	MimeType     string
	SizeBytes    int64
}

type Document struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"originalName"`
	StoredName   string    `json:"storedName"`
	MimeType     string    `json:"mimeType"`
	SizeBytes    int64     `json:"sizeBytes"`
	Status       string    `json:"status"`
	ChunkCount   int       `json:"chunkCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Chunk struct {
	ChunkIndex int
	Content    string
	TokenCount int
}

type Service struct {
	store          Store
	indexer        EmbeddingIndexer
	uploadDir      string
	maxUploadBytes int64
}

func NewService(store Store, indexer EmbeddingIndexer, uploadDir string, maxUploadBytes int64) *Service {
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 * 1024 * 1024
	}

	return &Service{
		store:          store,
		indexer:        indexer,
		uploadDir:      uploadDir,
		maxUploadBytes: maxUploadBytes,
	}
}

func (s *Service) Upload(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (Document, error) {
	if fileHeader == nil {
		return Document{}, ErrFileRequired
	}

	mimeType, err := validateFile(fileHeader, s.maxUploadBytes)
	if err != nil {
		return Document{}, err
	}

	storedName, err := storedFileName(fileHeader.Filename)
	if err != nil {
		return Document{}, err
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return Document{}, err
	}

	source, err := fileHeader.Open()
	if err != nil {
		return Document{}, err
	}
	defer source.Close()

	destinationPath := filepath.Join(s.uploadDir, storedName)
	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return Document{}, err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return Document{}, err
	}

	text, err := extractText(destinationPath, mimeType)
	if err != nil {
		_ = os.Remove(destinationPath)
		return Document{}, err
	}

	chunks, err := chunkText(text, 220, 40)
	if err != nil {
		_ = os.Remove(destinationPath)
		return Document{}, err
	}

	doc, err := s.store.CreateDocument(ctx, CreateDocumentParams{
		UserID:       userID,
		OriginalName: fileHeader.Filename,
		StoredName:   storedName,
		MimeType:     mimeType,
		SizeBytes:    fileHeader.Size,
	}, chunks)
	if err != nil {
		_ = os.Remove(destinationPath)
		return Document{}, err
	}

	if s.indexer != nil {
		if _, err := s.indexer.IndexMissingEmbeddings(ctx, len(chunks)); err != nil {
			return Document{}, err
		}
	}

	return doc, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Document, error) {
	return s.store.ListDocuments(ctx, userID)
}

func extractText(path string, mimeType string) (string, error) {
	switch {
	case mimeType == "application/pdf":
		return extractPDFText(path)
	case strings.HasPrefix(mimeType, "text/plain"):
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return normalizeText(string(content)), nil
	default:
		return "", ErrInvalidType
	}
}

func extractPDFText(path string) (string, error) {
	text, err := extractPDFTextWithLedongthuc(path)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	text, err = extractPDFTextWithRSC(path)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	return extractPDFTextFromStructure(path)
}

func extractPDFTextWithLedongthuc(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var builder strings.Builder
	for pageNumber := 1; pageNumber <= reader.NumPage(); pageNumber++ {
		page := reader.Page(pageNumber)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}

		builder.WriteString(text)
		builder.WriteString("\n")
	}

	return normalizeText(builder.String()), nil
}

func extractPDFTextWithRSC(path string) (string, error) {
	reader, err := rscpdf.Open(path)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for pageNumber := 1; pageNumber <= reader.NumPage(); pageNumber++ {
		page := reader.Page(pageNumber)
		content := page.Content()
		for _, text := range content.Text {
			builder.WriteString(text.S)
			builder.WriteString(" ")
		}
		builder.WriteString("\n")
	}

	return normalizeText(builder.String()), nil
}

func extractPDFTextFromStructure(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	markers := []string{"/ActualText (", "/E (", "/T ("}
	seen := map[string]bool{}
	var builder strings.Builder

	for _, marker := range markers {
		offset := 0
		for {
			index := strings.Index(string(content[offset:]), marker)
			if index == -1 {
				break
			}

			openParen := offset + index + len(marker) - 1
			value, next := parsePDFLiteralString(content, openParen)
			offset = next

			value = normalizeText(value)
			if value == "" || seen[value] {
				continue
			}

			seen[value] = true
			builder.WriteString(value)
			builder.WriteString("\n")
		}
	}

	return normalizeText(builder.String()), nil
}

func parsePDFLiteralString(content []byte, openParen int) (string, int) {
	if openParen < 0 || openParen >= len(content) || content[openParen] != '(' {
		return "", openParen + 1
	}

	var builder strings.Builder
	depth := 1
	for index := openParen + 1; index < len(content); index++ {
		character := content[index]

		if character == '\\' {
			if index+1 >= len(content) {
				break
			}
			index++
			escaped := content[index]
			switch escaped {
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			case 'b', 'f':
				// Ignore rarely useful control escapes in extracted text.
			default:
				builder.WriteByte(escaped)
			}
			continue
		}

		switch character {
		case '(':
			depth++
			builder.WriteByte(character)
		case ')':
			depth--
			if depth == 0 {
				return builder.String(), index + 1
			}
			builder.WriteByte(character)
		default:
			builder.WriteByte(character)
		}
	}

	return builder.String(), len(content)
}

func chunkText(text string, chunkSize int, overlap int) ([]Chunk, error) {
	words := strings.Fields(normalizeText(text))
	if len(words) == 0 {
		return nil, ErrNoText
	}

	if chunkSize <= 0 {
		chunkSize = 220
	}

	if overlap < 0 {
		overlap = 0
	}

	if overlap >= chunkSize {
		overlap = chunkSize / 5
	}

	step := chunkSize - overlap
	chunks := make([]Chunk, 0, (len(words)/step)+1)
	for start := 0; start < len(words); start += step {
		end := start + chunkSize
		if end > len(words) {
			end = len(words)
		}

		content := strings.Join(words[start:end], " ")
		chunks = append(chunks, Chunk{
			ChunkIndex: len(chunks),
			Content:    content,
			TokenCount: len(words[start:end]),
		})

		if end == len(words) {
			break
		}
	}

	return chunks, nil
}

func normalizeText(text string) string {
	return strings.Join(strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	}), " ")
}

func validateFile(fileHeader *multipart.FileHeader, maxUploadBytes int64) (string, error) {
	if fileHeader.Filename == "" {
		return "", ErrUnsafeName
	}

	if strings.ContainsAny(fileHeader.Filename, `/\`) {
		return "", ErrUnsafeName
	}

	if filepath.Base(fileHeader.Filename) != fileHeader.Filename {
		return "", ErrUnsafeName
	}

	if fileHeader.Size <= 0 {
		return "", ErrFileRequired
	}

	if fileHeader.Size > maxUploadBytes {
		return "", ErrFileTooLarge
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if extension != ".pdf" && extension != ".txt" {
		return "", ErrInvalidType
	}

	mimeType, err := detectMimeType(fileHeader)
	if err != nil {
		return "", err
	}

	switch extension {
	case ".pdf":
		if mimeType != "application/pdf" {
			return "", ErrInvalidType
		}
	case ".txt":
		if !strings.HasPrefix(mimeType, "text/plain") {
			return "", ErrInvalidType
		}
	}

	return mimeType, nil
}

func detectMimeType(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	if n == 0 {
		return "", ErrFileRequired
	}

	return strings.ToLower(strings.TrimSpace(http.DetectContentType(buffer[:n]))), nil
}

func storedFileName(originalName string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	extension := strings.ToLower(filepath.Ext(originalName))
	return hex.EncodeToString(random) + extension, nil
}
