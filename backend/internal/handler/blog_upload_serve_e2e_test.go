package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 通过真实 gin 引擎验证静态图片路由：
//   - 正常文件返回 200 且字节一致
//   - 目录穿越 / 不存在的文件返回 404
func TestBlogUploadHandler_ServeImageOverHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	h := NewBlogUploadHandler(dir)

	// 预置一张图片 + 一个不应被访问到的敏感文件
	if err := os.WriteFile(filepath.Join(dir, "uploads", "blogs", "ok.png"), testPNG, 0o644); err != nil {
		t.Fatalf("seed image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("top secret"), 0o644); err != nil {
		t.Fatalf("seed secret: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/public/uploads/blogs/*filename", h.ServeImage)
	srv := httptest.NewServer(router)
	defer srv.Close()

	// 正常访问
	resp, err := http.Get(srv.URL + "/api/v1/public/uploads/blogs/ok.png")
	if err != nil {
		t.Fatalf("get image: %v", err)
	}
	body := make([]byte, len(testPNG))
	n, _ := resp.Body.Read(body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid image status = %d, want 200", resp.StatusCode)
	}
	if n != len(testPNG) || string(body[:n]) != string(testPNG) {
		t.Fatalf("served bytes mismatch (got %d bytes, want %d)", n, len(testPNG))
	}

	// 穿越与不存在的文件都必须 404，且不能泄漏目录外的文件内容
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}
	for _, path := range []string{
		"/api/v1/public/uploads/blogs/../secret.txt",
		"/api/v1/public/uploads/blogs/..%2fsecret.txt",
		"/api/v1/public/uploads/blogs/....//secret.txt",
		"/api/v1/public/uploads/blogs/missing.png",
		"/api/v1/public/uploads/blogs/",
	} {
		resp, err := client.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("get %q: %v", path, err)
		}
		buf := make([]byte, 256)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatalf("%q: status = 200, want 404", path)
		}
		if strings.Contains(string(buf[:n]), "top secret") {
			t.Fatalf("%q leaked file content outside uploads dir", path)
		}
	}
}
