package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultLocalMaxUploadBytes caps local disk uploads at 16 MiB.
	DefaultLocalMaxUploadBytes int64 = 16 << 20
	// DefaultVercelMaxUploadBytes stays safely below Vercel Functions' 4.5 MB payload cap.
	DefaultVercelMaxUploadBytes int64 = 4 << 20

	defaultBlobAPIBaseURL = "https://blob.vercel-storage.com"
	vercelBlobAPIVersion  = "7"
)

var (
	ErrFileTooLarge = errors.New("uploaded file exceeds maximum allowed size")
	unsafeCharRe    = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)

// UploadResult describes a persisted media or document asset.
type UploadResult struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	StorageMode string `json:"storage"`
}

// Store abstracts uploaded file persistence across local disk and Vercel Blob.
type Store interface {
	Save(ctx context.Context, originalFilename string, contentType string, r io.Reader, declaredSize int64) (*UploadResult, error)
	Mode() string
	IsRemote() bool
	MaxBytes() int64
	LocalDir() string
}

// NewFromEnv selects VercelBlobStore when BLOB_READ_WRITE_TOKEN is set, otherwise LocalStore.
func NewFromEnv(localDir string) Store {
	if token := strings.TrimSpace(os.Getenv("BLOB_READ_WRITE_TOKEN")); token != "" {
		return NewVercelBlobStore(token, "")
	}
	return NewLocalStore(localDir)
}

// LocalStore persists uploaded files under a local workspace/data directory.
type LocalStore struct {
	dir      string
	maxBytes int64
}

func NewLocalStore(dir string) *LocalStore {
	if strings.TrimSpace(dir) == "" {
		dir = "./data/uploads"
	}
	maxBytes := DefaultLocalMaxUploadBytes
	if isServerlessEnv() {
		maxBytes = DefaultVercelMaxUploadBytes
	}
	return &LocalStore{dir: dir, maxBytes: maxBytes}
}

func (s *LocalStore) Mode() string     { return "local-disk" }
func (s *LocalStore) IsRemote() bool   { return false }
func (s *LocalStore) MaxBytes() int64  { return s.maxBytes }
func (s *LocalStore) LocalDir() string { return s.dir }

func (s *LocalStore) Save(_ context.Context, originalFilename string, contentType string, r io.Reader, declaredSize int64) (*UploadResult, error) {
	if declaredSize > s.maxBytes {
		return nil, fmt.Errorf("%w (%d MB max)", ErrFileTooLarge, s.maxBytes/(1<<20))
	}
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	payload, err := readLimited(r, s.maxBytes)
	if err != nil {
		return nil, err
	}

	cleanFilename := sanitizeFilename(originalFilename)
	dstPath := filepath.Join(s.dir, cleanFilename)
	if err := os.WriteFile(dstPath, payload, 0644); err != nil {
		return nil, fmt.Errorf("failed to write uploaded file: %w", err)
	}

	return &UploadResult{
		URL:         "/data/uploads/" + cleanFilename,
		Filename:    cleanFilename,
		Size:        int64(len(payload)),
		ContentType: detectContentType(cleanFilename, contentType, payload),
		StorageMode: s.Mode(),
	}, nil
}

// VercelBlobStore uploads files directly to the Vercel Blob REST API.
type VercelBlobStore struct {
	token      string
	apiBaseURL string
	prefix     string
	maxBytes   int64
	client     *http.Client
}

func NewVercelBlobStore(token, apiBaseURL string) *VercelBlobStore {
	base := strings.TrimRight(strings.TrimSpace(apiBaseURL), "/")
	if base == "" {
		base = defaultBlobAPIBaseURL
	}
	return &VercelBlobStore{
		token:      token,
		apiBaseURL: base,
		prefix:     "selvedge",
		maxBytes:   DefaultVercelMaxUploadBytes,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *VercelBlobStore) Mode() string     { return "vercel-blob" }
func (s *VercelBlobStore) IsRemote() bool   { return true }
func (s *VercelBlobStore) MaxBytes() int64  { return s.maxBytes }
func (s *VercelBlobStore) LocalDir() string { return "" }

func (s *VercelBlobStore) Save(ctx context.Context, originalFilename string, contentType string, r io.Reader, declaredSize int64) (*UploadResult, error) {
	if declaredSize > s.maxBytes {
		return nil, fmt.Errorf("%w (%d MB max on serverless)", ErrFileTooLarge, s.maxBytes/(1<<20))
	}

	payload, err := readLimited(r, s.maxBytes)
	if err != nil {
		return nil, err
	}

	cleanFilename := sanitizeFilename(originalFilename)
	detectedType := detectContentType(cleanFilename, contentType, payload)
	targetURL := fmt.Sprintf("%s/%s/%s", s.apiBaseURL, s.prefix, cleanFilename)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create blob request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("x-api-version", vercelBlobAPIVersion)
	req.Header.Set("x-content-type", detectedType)
	req.Header.Set("x-add-random-suffix", "1")
	req.ContentLength = int64(len(payload))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blob upload request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("blob upload rejected (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		URL         string `json:"url"`
		Pathname    string `json:"pathname"`
		ContentType string `json:"contentType"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("invalid blob response JSON: %w", err)
	}
	if strings.TrimSpace(parsed.URL) == "" {
		return nil, errors.New("blob response missing url")
	}

	finalType := parsed.ContentType
	if finalType == "" {
		finalType = detectedType
	}

	return &UploadResult{
		URL:         parsed.URL,
		Filename:    cleanFilename,
		Size:        int64(len(payload)),
		ContentType: finalType,
		StorageMode: s.Mode(),
	}, nil
}

func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}
	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("%w (%d MB max)", ErrFileTooLarge, maxBytes/(1<<20))
	}
	return payload, nil
}

func sanitizeFilename(raw string) string {
	base := filepath.Base(strings.TrimSpace(raw))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "upload.bin"
	}
	cleanBase := strings.ReplaceAll(base, " ", "_")
	cleanBase = unsafeCharRe.ReplaceAllString(cleanBase, "_")
	cleanBase = strings.TrimLeft(cleanBase, ".")
	if cleanBase == "" {
		cleanBase = "upload.bin"
	}
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), cleanBase)
}

func detectContentType(filename, headerType string, prefix []byte) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".webm":
		return "audio/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ai":
		return "application/postscript"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	}

	if headerType != "" && headerType != "application/octet-stream" {
		return headerType
	}
	if mimeByExt := mime.TypeByExtension(ext); mimeByExt != "" {
		return mimeByExt
	}
	if len(prefix) > 0 {
		return http.DetectContentType(prefix)
	}
	return "application/octet-stream"
}

func isServerlessEnv() bool {
	return os.Getenv("VERCEL") != "" || os.Getenv("VERCEL_ENV") != ""
}
