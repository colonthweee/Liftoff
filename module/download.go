package module

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cat2/liftoff/types"
	"cat2/liftoff/util"
)

const (
	maxRetries       = 3
	baseTimeout      = 30 * time.Second
	maxRedirects     = 10
	maxContentLength = 1024 * 1024 * 1024 
)

type DownloadManager struct {
	log    *util.Logger
	client *http.Client
}

func NewDownloadManager(log *util.Logger) *DownloadManager {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		DialContext: (&net.Dialer{
			Timeout:   baseTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: baseTimeout,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   baseTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		},
	}

	return &DownloadManager{
		log:    log,
		client: client,
	}
}

func (d *DownloadManager) Download(config types.DownloadConfig) error {
	for _, file := range config.Files {
		if err := d.downloadFile(file); err != nil {
			return fmt.Errorf("failed to download %s: %w", file.URL, err)
		}
	}
	return nil
}

func (d *DownloadManager) validateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	if !parsedURL.IsAbs() {
		return fmt.Errorf("URL must be absolute")
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed")
	}

	if strings.TrimSpace(parsedURL.Host) == "" {
		return fmt.Errorf("URL host is required")
	}

	return nil
}

type downloadResult struct {
	data []byte
	err  error
}

func (d *DownloadManager) downloadWithRetry(urlStr string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		result := make(chan downloadResult, 1)
		go func() {
			data, err := d.downloadOnce(urlStr)
			result <- downloadResult{data, err}
		}()

		select {
		case res := <-result:
			if res.err == nil {
				return res.data, nil
			}
			lastErr = res.err
			d.log.Warn(fmt.Sprintf("Download attempt %d failed: %v", attempt+1, res.err))
		case <-time.After(baseTimeout):
			lastErr = fmt.Errorf("download timeout")
			d.log.Warn(fmt.Sprintf("Download attempt %d timed out", attempt+1))
		}
	}

	return nil, fmt.Errorf("all download attempts failed: %v", lastErr)
}

func (d *DownloadManager) downloadOnce(urlStr string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	if contentLength > maxContentLength {
		return nil, fmt.Errorf("content length %d exceeds maximum allowed size", contentLength)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxContentLength))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func (d *DownloadManager) verifyChecksum(data []byte, expectedSHA256 string) error {
	hasher := sha256.New()
	hasher.Write(data)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(actualChecksum, expectedSHA256) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actualChecksum)
	}

	return nil
}

func (d *DownloadManager) downloadFile(file types.DownloadFile) error {
	if err := d.validateURL(file.URL); err != nil {
		return fmt.Errorf("invalid URL %s: %w", file.URL, err)
	}

	expandedDest := os.ExpandEnv(file.Dest)
	destDir := filepath.Dir(expandedDest)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	d.log.Info(fmt.Sprintf("Downloading %s to %s", file.URL, expandedDest))

	data, err := d.downloadWithRetry(file.URL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if file.SHA256 != "" {
		d.log.Info("Verifying file checksum")
		if err := d.verifyChecksum(data, file.SHA256); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		d.log.Success("Checksum verified successfully")
	} else {
		d.log.Warn("No checksum provided for verification")
	}

	tmpFile, err := os.CreateTemp(destDir, "download-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write file: %w", err)
	}
	tmpFile.Close()

	finalPath := expandedDest
	if file.Rename != "" {
		finalPath = filepath.Join(filepath.Dir(expandedDest), file.Rename)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("failed to move file to destination: %w", err)
	}

	d.log.Success(fmt.Sprintf("Successfully downloaded file to %s", finalPath))
	return nil
}
