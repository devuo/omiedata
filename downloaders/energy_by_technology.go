package downloaders

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devuo/omiedata/types"
)

// EnergyByTechnologyDownloader downloads energy by technology data files
type EnergyByTechnologyDownloader struct {
	*GeneralDownloader
	systemType types.SystemType
}

// NewEnergyByTechnologyDownloader creates a new energy by technology downloader
func NewEnergyByTechnologyDownloader(systemType types.SystemType) *EnergyByTechnologyDownloader {
	urlMask := "AGNO_YYYY/MES_MM/TXT/INT_PBC_TECNOLOGIAS_H_SYS_DD_MM_YYYY_DD_MM_YYYY.TXT"
	outputMask := "EnergyByTechnology_SYS_YYYYMMDD.TXT"

	return &EnergyByTechnologyDownloader{
		GeneralDownloader: NewGeneralDownloader(urlMask, outputMask),
		systemType:        systemType,
	}
}

// URLResponses returns a channel of HTTP responses for the date range
func (d *EnergyByTechnologyDownloader) URLResponses(ctx context.Context, dateIni, dateEnd time.Time, verbose bool) <-chan ResponseResult {
	// Override to use custom URL generation
	resultChan := make(chan ResponseResult)

	go func() {
		defer close(resultChan)

		for date := dateIni; !date.After(dateEnd); date = date.AddDate(0, 0, 1) {
			select {
			case <-ctx.Done():
				return
			default:
				result := d.downloadSingleDate(ctx, date, verbose)
				resultChan <- result
			}
		}
	}()

	return resultChan
}

// downloadSingleDate downloads data for a single date with custom URL generation
func (d *EnergyByTechnologyDownloader) downloadSingleDate(ctx context.Context, date time.Time, verbose bool) ResponseResult {
	url := d.generateURL(date)

	var lastErr error
	for attempt := 0; attempt <= d.GeneralDownloader.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ResponseResult{Date: date, URL: url, Error: ctx.Err()}
			case <-time.After(d.GeneralDownloader.config.RetryDelay * time.Duration(attempt)):
			}
		}

		if verbose {
			if attempt > 0 {
				fmt.Printf("Retrying (%d/%d) %s...\n", attempt, d.GeneralDownloader.config.MaxRetries, url)
			} else {
				fmt.Printf("Requesting %s...\n", url)
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := d.GeneralDownloader.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return ResponseResult{Response: resp, Date: date, URL: url}
		}

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
		Error: types.NewOMIEError(types.ErrCodeDownload, fmt.Sprintf("failed after %d attempts", d.GeneralDownloader.config.MaxRetries), lastErr),
	}
}

// generateURL generates the URL for a specific date, replacing SYS with system type
func (d *EnergyByTechnologyDownloader) generateURL(date time.Time) string {
	url := d.GetCompleteURL()
	url = strings.ReplaceAll(url, "YYYY", fmt.Sprintf("%04d", date.Year()))
	url = strings.ReplaceAll(url, "MM", fmt.Sprintf("%02d", date.Month()))
	url = strings.ReplaceAll(url, "DD", fmt.Sprintf("%02d", date.Day()))
	url = strings.ReplaceAll(url, "SYS", fmt.Sprintf("%d", int(d.systemType)))
	return url
}

// generateFilename generates the output filename, replacing SYS with system type
func (d *EnergyByTechnologyDownloader) generateFilename(date time.Time) string {
	filename := d.outputMask
	filename = strings.ReplaceAll(filename, "YYYY", fmt.Sprintf("%04d", date.Year()))
	filename = strings.ReplaceAll(filename, "MM", fmt.Sprintf("%02d", date.Month()))
	filename = strings.ReplaceAll(filename, "DD", fmt.Sprintf("%02d", date.Day()))
	filename = strings.ReplaceAll(filename, "SYS", fmt.Sprintf("%d", int(d.systemType)))
	return filename
}
