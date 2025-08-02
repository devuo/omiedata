package downloaders

import (
	"context"
	"net/http"
	"time"
)

// Downloader defines the interface for downloading OMIE data
type Downloader interface {
	// GetCompleteURL returns the complete URL pattern for this downloader
	GetCompleteURL() string

	// DownloadData downloads data for a date range and saves to folder
	DownloadData(ctx context.Context, dateIni, dateEnd time.Time, outputFolder string, verbose bool) error

	// URLResponses returns a channel of HTTP responses for the date range
	URLResponses(ctx context.Context, dateIni, dateEnd time.Time, verbose bool) <-chan ResponseResult
}

// ResponseResult wraps an HTTP response with potential error
type ResponseResult struct {
	Response *http.Response
	Date     time.Time
	URL      string
	Error    error
}

// DownloadConfig holds configuration for downloading
type DownloadConfig struct {
	MaxRetries     int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
	MaxConcurrent  int
}
