package downloaders

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/devuo/omiedata/types"
)

const (
	baseURL = "https://www.omie.es/sites/default/files/dados/"
)

// GeneralDownloader implements the base functionality for OMIE downloaders
type GeneralDownloader struct {
	urlMask    string
	outputMask string
	client     *http.Client
	config     DownloadConfig
}

// NewGeneralDownloader creates a new GeneralDownloader
func NewGeneralDownloader(urlMask, outputMask string) *GeneralDownloader {
	return &GeneralDownloader{
		urlMask:    urlMask,
		outputMask: outputMask,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: DownloadConfig{
			MaxRetries:     3,
			RetryDelay:     time.Second,
			RequestTimeout: 30 * time.Second,
			MaxConcurrent:  5,
		},
	}
}

// SetConfig updates the download configuration
func (d *GeneralDownloader) SetConfig(config DownloadConfig) {
	d.config = config
	d.client.Timeout = config.RequestTimeout
}

// GetCompleteURL returns the complete URL pattern
func (d *GeneralDownloader) GetCompleteURL() string {
	return baseURL + d.urlMask
}

// DownloadData downloads data for a date range and saves to folder
func (d *GeneralDownloader) DownloadData(ctx context.Context, dateIni, dateEnd time.Time, outputFolder string, verbose bool) error {
	// Ensure output folder exists
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return types.NewOMIEError(types.ErrCodeDownload, "failed to create output folder", err)
	}

	// Use the response channel to download and save files
	responseChan := d.URLResponses(ctx, dateIni, dateEnd, verbose)

	var errors []error
	for result := range responseChan {
		if result.Error != nil {
			errors = append(errors, result.Error)
			continue
		}

		// Generate output filename
		filename := d.generateFilename(result.Date)
		filepath := filepath.Join(outputFolder, filename)

		if verbose {
			fmt.Printf("Saving to %s...\n", filepath)
		}

		if err := d.saveResponse(result.Response, filepath); err != nil {
			errors = append(errors, types.NewOMIEError(types.ErrCodeDownload, "failed to save file", err))
		}

		result.Response.Body.Close()
	}

	if len(errors) > 0 {
		return fmt.Errorf("download completed with %d errors: %v", len(errors), errors[0])
	}

	return nil
}

// URLResponses returns a channel of HTTP responses for the date range
func (d *GeneralDownloader) URLResponses(ctx context.Context, dateIni, dateEnd time.Time, verbose bool) <-chan ResponseResult {
	resultChan := make(chan ResponseResult)

	go func() {
		defer close(resultChan)

		// Create a channel for dates
		dateChan := make(chan time.Time)

		// Create worker pool
		var wg sync.WaitGroup
		for i := 0; i < d.config.MaxConcurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for date := range dateChan {
					select {
					case <-ctx.Done():
						return
					default:
						result := d.downloadSingleDate(ctx, date, verbose)
						resultChan <- result
					}
				}
			}()
		}

		// Send dates to workers
		go func() {
			defer close(dateChan)
			for d := dateIni; !d.After(dateEnd); d = d.AddDate(0, 0, 1) {
				select {
				case <-ctx.Done():
					return
				case dateChan <- d:
				}
			}
		}()

		wg.Wait()
	}()

	return resultChan
}

// downloadSingleDate downloads data for a single date with retries
func (d *GeneralDownloader) downloadSingleDate(ctx context.Context, date time.Time, verbose bool) ResponseResult {
	url := d.generateURL(date)

	var lastErr error
	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ResponseResult{
					Date:  date,
					URL:   url,
					Error: ctx.Err(),
				}
			case <-time.After(d.config.RetryDelay * time.Duration(attempt)):
			}
		}

		if verbose {
			if attempt > 0 {
				fmt.Printf("Retrying (%d/%d) %s...\n", attempt, d.config.MaxRetries, url)
			} else {
				fmt.Printf("Requesting %s...\n", url)
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := d.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Check for success
		if resp.StatusCode == http.StatusOK {
			return ResponseResult{
				Response: resp,
				Date:     date,
				URL:      url,
			}
		}

		// Handle different error codes
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			lastErr = types.NewOMIEError(types.ErrCodeNotFound, fmt.Sprintf("data not available for date %s", date.Format("2006-01-02")), nil)
		} else {
			lastErr = types.NewOMIEError(types.ErrCodeNetwork, fmt.Sprintf("HTTP %d", resp.StatusCode), nil)
		}
	}

	return ResponseResult{
		Date:  date,
		URL:   url,
		Error: types.NewOMIEError(types.ErrCodeDownload, fmt.Sprintf("failed after %d attempts", d.config.MaxRetries), lastErr),
	}
}

// generateURL generates the URL for a specific date
func (d *GeneralDownloader) generateURL(date time.Time) string {
	url := d.GetCompleteURL()
	url = strings.ReplaceAll(url, "YYYY", fmt.Sprintf("%04d", date.Year()))
	url = strings.ReplaceAll(url, "MM", fmt.Sprintf("%02d", date.Month()))
	url = strings.ReplaceAll(url, "DD", fmt.Sprintf("%02d", date.Day()))
	return url
}

// generateFilename generates the output filename for a specific date
func (d *GeneralDownloader) generateFilename(date time.Time) string {
	filename := d.outputMask
	filename = strings.ReplaceAll(filename, "YYYY", fmt.Sprintf("%04d", date.Year()))
	filename = strings.ReplaceAll(filename, "MM", fmt.Sprintf("%02d", date.Month()))
	filename = strings.ReplaceAll(filename, "DD", fmt.Sprintf("%02d", date.Day()))
	return filename
}

// saveResponse saves an HTTP response to a file
func (d *GeneralDownloader) saveResponse(resp *http.Response, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
