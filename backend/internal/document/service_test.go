package document

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeStore struct {
	params CreateDocumentParams
}

func (f *fakeStore) CreateDocument(_ context.Context, params CreateDocumentParams) (Document, error) {
	f.params = params
	return Document{
		ID:           "document-1",
		OriginalName: params.OriginalName,
		StoredName:   params.StoredName,
		MimeType:     params.MimeType,
		SizeBytes:    params.SizeBytes,
		Status:       "processing",
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func TestUploadRejectsInvalidType(t *testing.T) {
	service := NewService(&fakeStore{}, t.TempDir(), 1024)
	file := newFileHeader(t, "script.exe", "application/octet-stream", []byte("MZ executable"))

	_, err := service.Upload(context.Background(), "user-1", file)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected invalid type, got %v", err)
	}
}

func TestUploadRejectsSpoofedPDF(t *testing.T) {
	service := NewService(&fakeStore{}, t.TempDir(), 1024)
	file := newFileHeader(t, "document.pdf", "application/pdf", []byte("not really a pdf"))

	_, err := service.Upload(context.Background(), "user-1", file)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected invalid type, got %v", err)
	}
}

func TestUploadAcceptsPDFWhenMultipartHeaderIsOctetStream(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, t.TempDir(), 1024)
	file := newFileHeader(t, "document.pdf", "application/octet-stream", []byte("%PDF-1.7\n1 0 obj\n<<>>\nendobj\n"))

	_, err := service.Upload(context.Background(), "user-1", file)
	if err != nil {
		t.Fatalf("upload pdf with octet-stream header: %v", err)
	}

	if store.params.MimeType != "application/pdf" {
		t.Fatalf("expected detected pdf mime type, got %q", store.params.MimeType)
	}
}

func TestUploadRejectsOversizeFile(t *testing.T) {
	service := NewService(&fakeStore{}, t.TempDir(), 2)
	file := newFileHeader(t, "notes.txt", "text/plain", []byte("too large"))

	_, err := service.Upload(context.Background(), "user-1", file)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected file too large, got %v", err)
	}
}

func TestUploadRejectsUnsafeFileName(t *testing.T) {
	service := NewService(&fakeStore{}, t.TempDir(), 1024)
	file := &multipart.FileHeader{
		Filename: "../notes.txt",
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"text/plain"},
		},
		Size: 5,
	}

	_, err := service.Upload(context.Background(), "user-1", file)
	if !errors.Is(err, ErrUnsafeName) {
		t.Fatalf("expected unsafe name, got %v", err)
	}
}

func TestUploadSavesFileAndMetadata(t *testing.T) {
	store := &fakeStore{}
	uploadDir := t.TempDir()
	service := NewService(store, uploadDir, 1024)
	file := newFileHeader(t, "notes.txt", "text/plain", []byte("hello"))

	doc, err := service.Upload(context.Background(), "user-1", file)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if doc.ID != "document-1" {
		t.Fatalf("expected document id, got %q", doc.ID)
	}

	if store.params.OriginalName != "notes.txt" {
		t.Fatalf("expected original name to be stored, got %q", store.params.OriginalName)
	}

	if store.params.StoredName == "" || filepath.Ext(store.params.StoredName) != ".txt" {
		t.Fatalf("expected safe stored txt name, got %q", store.params.StoredName)
	}

	if _, err := os.Stat(filepath.Join(uploadDir, store.params.StoredName)); err != nil {
		t.Fatalf("expected uploaded file to exist: %v", err)
	}
}

func newFileHeader(t *testing.T, filename string, contentType string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}

	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file, got %d", len(files))
	}

	return files[0]
}
