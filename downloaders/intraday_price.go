package downloaders

import (
	"fmt"
	"strings"
	"time"

	"github.com/devuo/omiedata/types"
)

// IntradayPriceDownloader downloads intraday price data files
type IntradayPriceDownloader struct {
	*GeneralDownloader
	session types.SessionType
}

// NewIntradayPriceDownloader creates a new intraday price downloader
func NewIntradayPriceDownloader(session types.SessionType) *IntradayPriceDownloader {
	urlMask := "AGNO_YYYY/MES_MM/TXT/INT_PIB_EV_H_1_SS_DD_MM_YYYY_DD_MM_YYYY.TXT"
	outputMask := "PrecioIntra_SS_YYYYMMDD.txt"

	return &IntradayPriceDownloader{
		GeneralDownloader: NewGeneralDownloader(urlMask, outputMask),
		session:           session,
	}
}

// generateURL generates the URL for a specific date, replacing SS with session
func (d *IntradayPriceDownloader) generateURL(date time.Time) string {
	url := d.GetCompleteURL()
	url = strings.ReplaceAll(url, "YYYY", fmt.Sprintf("%04d", date.Year()))
	url = strings.ReplaceAll(url, "MM", fmt.Sprintf("%02d", date.Month()))
	url = strings.ReplaceAll(url, "DD", fmt.Sprintf("%02d", date.Day()))
	url = strings.ReplaceAll(url, "SS", fmt.Sprintf("%d", int(d.session)))
	return url
}

// generateFilename generates the output filename, replacing SS with session
func (d *IntradayPriceDownloader) generateFilename(date time.Time) string {
	filename := d.outputMask
	filename = strings.ReplaceAll(filename, "YYYY", fmt.Sprintf("%04d", date.Year()))
	filename = strings.ReplaceAll(filename, "MM", fmt.Sprintf("%02d", date.Month()))
	filename = strings.ReplaceAll(filename, "DD", fmt.Sprintf("%02d", date.Day()))
	filename = strings.ReplaceAll(filename, "SS", fmt.Sprintf("%d", int(d.session)))
	return filename
}
