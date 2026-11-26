package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStoreSaveAndSanitize(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStore(tmpDir)

	if store.IsRemote() {
		t.Fatalf("expected LocalStore IsRemote() == false")
	}

	res, err := store.Save(context.Background(), "../../secret /my sample (1).png", "image/png", bytes.NewReader([]byte("pngdata")), 7)
	if err != nil {
		t.Fatalf("LocalStore Save failed: %v", err)
	}
	if !strings.HasPrefix(res.URL, "/data/uploads/") {
		t.Fatalf("unexpected URL prefix: %q", res.URL)
	}
	if strings.Contains(res.Filename, "..") || strings.Contains(res.Filename, "/") || strings.Contains(res.Filename, " ") {
		t.Fatalf("filename not sanitized safely: %q", res.Filename)
	}

	written, err := os.ReadFile(filepath.Join(tmpDir, res.Filename))
	if err != nil {
		t.Fatalf("failed reading saved local file: %v", err)
	}
	if string(written) != "pngdata" {
		t.Fatalf("expected 'pngdata', got %q", string(written))
	}
}

func TestVercelBlobStoreSaveSuccessAndHeaders(t *testing.T) {
	var gotAuth, gotVer, gotCT, gotSuffix, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		gotVer = r.Header.Get("x-api-version")
		gotCT = r.Header.Get("x-content-type")
		gotSuffix = r.Header.Get("x-add-random-suffix")
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"url":         "https://example.public.blob.vercel-storage.com" + r.URL.Path,
			"pathname":    strings.TrimPrefix(r.URL.Path, "/"),
			"contentType": gotCT,
		})
	}))
	defer srv.Close()

	store := NewVercelBlobStore("vercel_blob_rw_test123", srv.URL)
	if !store.IsRemote() || store.Mode() != "vercel-blob" {
		t.Fatalf("unexpected VercelBlobStore mode/remote flags")
	}

	res, err := store.Save(context.Background(), "tech pack final.pdf", "", bytes.NewReader([]byte("%PDF-1.4")), 8)
	if err != nil {
		t.Fatalf("VercelBlobStore Save failed: %v", err)
	}

	if gotAuth != "Bearer vercel_blob_rw_test123" {
		t.Errorf("Authorization header mismatch: %q", gotAuth)
	}
	if gotVer != "7" {
		t.Errorf("x-api-version mismatch: %q", gotVer)
	}
	if gotCT != "application/pdf" {
		t.Errorf("x-content-type mismatch: %q", gotCT)
	}
	if gotSuffix != "1" {
		t.Errorf("x-add-random-suffix mismatch: %q", gotSuffix)
	}
	if !strings.HasPrefix(gotPath, "/selvedge/") || !strings.HasSuffix(gotPath, "_tech_pack_final.pdf") {
		t.Errorf("unexpected target path: %q", gotPath)
	}
	if gotBody != "%PDF-1.4" {
		t.Errorf("unexpected body: %q", gotBody)
	}
	if !strings.HasPrefix(res.URL, "https://example.public.blob.vercel-storage.com/selvedge/") {
		t.Errorf("unexpected returned URL: %q", res.URL)
	}
}

func TestVercelBlobStoreRejectsOversizedUpload(t *testing.T) {
	store := NewVercelBlobStore("token", "https://blob.vercel-storage.com")
	oversized := bytes.NewReader(make([]byte, int(DefaultVercelMaxUploadBytes)+10))
	_, err := store.Save(context.Background(), "huge.png", "image/png", oversized, DefaultVercelMaxUploadBytes+10)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestNewFromEnvSwitchesStore(t *testing.T) {
	t.Setenv("BLOB_READ_WRITE_TOKEN", "")
	if s := NewFromEnv("./tmp-uploads"); s.IsRemote() {
		t.Fatalf("expected LocalStore when token empty")
	}

	t.Setenv("BLOB_READ_WRITE_TOKEN", "vercel_blob_rw_dummy")
	if s := NewFromEnv("./tmp-uploads"); !s.IsRemote() || s.Mode() != "vercel-blob" {
		t.Fatalf("expected VercelBlobStore when token set")
	}
}
