package util

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxRetries       = 3
	baseTimeout      = 30 * time.Second
	maxRedirects     = 10
	maxContentLength = 1024 * 1024 * 1024 
)

type secureHttpClient struct {
	client *http.Client
	log    *Logger
}

func newSecureHttpClient(log *Logger) *secureHttpClient {
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

	return &secureHttpClient{
		client: client,
		log:    log,
	}
}

type downloadResult struct {
	data []byte
	err  error
}

func (c *secureHttpClient) downloadWithRetry(urlStr string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		result := make(chan downloadResult, 1)
		go func() {
			data, err := c.download(urlStr)
			result <- downloadResult{data, err}
		}()

		select {
		case res := <-result:
			if res.err == nil {
				return res.data, nil
			}
			lastErr = res.err
			c.log.Warn(fmt.Sprintf("Download attempt %d failed: %v", attempt+1, res.err))
		case <-time.After(baseTimeout):
			lastErr = fmt.Errorf("download timeout")
			c.log.Warn(fmt.Sprintf("Download attempt %d timed out", attempt+1))
		}
	}

	return nil, fmt.Errorf("all download attempts failed: %v", lastErr)
}

func (c *secureHttpClient) download(urlStr string) ([]byte, error) {
	if err := validateURL(urlStr); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
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

func validateURL(urlStr string) error {
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

func validateGitURL(urlStr string) error {
	if err := validateURL(urlStr); err != nil {
		return err
	}

	parsedURL, _ := url.Parse(urlStr)
	host := strings.ToLower(parsedURL.Host)

	
	trustedHosts := map[string]bool{
		"github.com":    true,
		"gitlab.com":    true,
		"bitbucket.org": true,
		"dev.azure.com": true,
	}

	if !trustedHosts[host] {
		return fmt.Errorf("untrusted Git host: %s", host)
	}

	return nil
}

func verifyChecksum(data []byte, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return fmt.Errorf("no checksum provided for verification")
	}

	hasher := sha256.New()
	hasher.Write(data)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(actualChecksum, expectedSHA256) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actualChecksum)
	}

	return nil
}

type DownloadConfig struct {
	URL            string
	ExpectedSHA256 string
	MaxSize        int64
	Timeout        time.Duration
}
