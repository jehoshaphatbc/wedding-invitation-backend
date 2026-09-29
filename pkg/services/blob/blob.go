package blob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
)

type BlobService struct {
	token string
}

func NewBlobService(cfg *config.Config) *BlobService {
	// You can retrieve this from config.GetEnv(...) or directly
	return &BlobService{
		token: cfg.BlobReadWriteToken,
	}
}

type PutBlobResult struct {
	URL                string `json:"url"`
	DownloadURL        string `json:"downloadUrl"`
	Pathname           string `json:"pathname"`
	ContentType        string `json:"contentType"`
	ContentDisposition string `json:"contentDisposition"`
}

func (s *BlobService) Upload(ctx context.Context, pathname string, file multipart.File, size int64, contentType string) (*PutBlobResult, error) {
	if s.token == "" {
		return nil, fmt.Errorf("BLOB_READ_WRITE_TOKEN is not configured")
	}

	apiURL := fmt.Sprintf("https://vercel.com/api/blob?pathname=%s", url.QueryEscape(pathname))

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, file)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	if size > 0 {
		req.ContentLength = size
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("x-api-version", "7")
	req.Header.Set("x-vercel-blob-access", "public")
	req.Header.Set("x-add-random-suffix", "0")
	if contentType != "" {
		req.Header.Set("x-content-type", contentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vercel blob upload error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result PutBlobResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &result, nil
}

func (s *BlobService) Delete(ctx context.Context, urls []string) error {
	if s.token == "" {
		return fmt.Errorf("BLOB_READ_WRITE_TOKEN is not configured")
	}

	if len(urls) == 0 {
		return nil
	}

	apiURL := "https://vercel.com/api/blob/delete"

	payload := map[string][]string{
		"urls": urls,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("x-api-version", "7")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vercel blob delete error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
