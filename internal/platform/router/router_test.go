package router

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ppl-amehta/selvedge/internal/platform/storage"
)

func TestHealthEndpoint(t *testing.T) {
	r := NewRouter(AppServices{
		UploadStore: storage.NewLocalStore(t.TempDir()),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}

	if res["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %v", res["status"])
	}

	if res["version"] != "1.0.0-production" {
		t.Errorf("expected version '1.0.0-production', got %v", res["version"])
	}

	if res["storage"] != "local-disk" {
		t.Errorf("expected storage 'local-disk', got %v", res["storage"])
	}
}

func TestUploadEndpointWithStore(t *testing.T) {
	r := NewRouter(AppServices{
		UploadStore: storage.NewLocalStore(t.TempDir()),
	})

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "tech pack.png")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	_, _ = fw.Write([]byte("fake-png-bytes"))
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on upload, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode upload response failed: %v", err)
	}
	if res["success"] != true {
		t.Fatalf("expected success=true, got %v", res)
	}
	urlStr, _ := res["url"].(string)
	if !strings.HasPrefix(urlStr, "/data/uploads/") || !strings.HasSuffix(urlStr, "_tech_pack.png") {
		t.Fatalf("unexpected uploaded url: %q", urlStr)
	}
}
