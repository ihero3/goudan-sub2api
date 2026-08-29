package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 1x1 red PNG
var testPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
	0x0C, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0xF8, 0xCF, 0xC0, 0xF0,
	0x9F, 0x00, 0x05, 0xFE, 0x02, 0xFE, 0xA7, 0x35, 0x81, 0x69, 0x8C, 0x21,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func newUploadContext(t *testing.T, filename string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/uploads/image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func TestBlogUploadHandler_UploadAndServe(t *testing.T) {
	dir := t.TempDir()
	h := NewBlogUploadHandler(dir)

	c, w := newUploadContext(t, "cover.png", testPNG)
	h.UploadImage(c)

	if w.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	idx := strings.Index(body, "/api/v1/public/uploads/blogs/")
	if idx < 0 {
		t.Fatalf("response missing image url: %s", body)
	}
	rest := body[idx+len("/api/v1/public/uploads/blogs/"):]
	name := rest[:strings.IndexAny(rest, `"}`)]
	if !strings.HasSuffix(name, ".png") {
		t.Fatalf("unexpected stored filename: %q", name)
	}
	if _, err := os.Stat(filepath.Join(dir, "uploads", "blogs", name)); err != nil {
		t.Fatalf("uploaded file not stored: %v", err)
	}

	// Serve it back
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/uploads/blogs/"+name, nil)
	c2.Params = gin.Params{{Key: "filename", Value: "/" + name}}
	h.ServeImage(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("serve status = %d, want 200", w2.Code)
	}
	if !bytes.Equal(w2.Body.Bytes(), testPNG) {
		t.Fatalf("served bytes mismatch (got %d bytes)", w2.Body.Len())
	}
}

func TestBlogUploadHandler_RejectsNonImage(t *testing.T) {
	h := NewBlogUploadHandler(t.TempDir())

	c, w := newUploadContext(t, "evil.txt", []byte("not an image at all"))
	h.UploadImage(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}

func TestBlogUploadHandler_RejectsOversize(t *testing.T) {
	h := NewBlogUploadHandler(t.TempDir())

	big := append([]byte{}, testPNG...)
	big = append(big, bytes.Repeat([]byte{0x00}, int(maxBlogImageSize))...)

	c, w := newUploadContext(t, "big.png", big)
	h.UploadImage(c)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body=%s", w.Code, w.Body.String())
	}
}

func TestBlogUploadHandler_ServeBlocksTraversal(t *testing.T) {
	dir := t.TempDir()
	h := NewBlogUploadHandler(dir)

	secret := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secret, []byte("top secret"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	for _, target := range []string{
		"/../secret.txt",
		"/..%2fsecret.txt",
		"/....//secret.txt",
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public/uploads/blogs"+target, nil)
		c.Params = gin.Params{{Key: "filename", Value: target}}
		h.ServeImage(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("traversal %q: status = %d, want 404", target, w.Code)
		}
		if strings.Contains(w.Body.String(), "top secret") {
			t.Fatalf("traversal %q leaked file content", target)
		}
	}
}

func TestBlogUploadHandler_MissingFile(t *testing.T) {
	h := NewBlogUploadHandler(t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/uploads/image", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=abc")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadImage(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
}
