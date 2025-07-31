package downloaders

import (
	"fmt"
	"strings"
	"time"
)

// SupplyDemandCurveDownloader downloads supply/demand curve data files
type SupplyDemandCurveDownloader struct {
	*GeneralDownloader
	hour int // Hour of the day (1-24)
}

// NewSupplyDemandCurveDownloader creates a new supply/demand curve downloader
func NewSupplyDemandCurveDownloader(hour int) *SupplyDemandCurveDownloader {
	urlMask := "AGNO_YYYY/MES_MM/TXT/INT_CURVA_ACUM_UO_MIB_1_HH_DD_MM_YYYY_DD_MM_YYYY.TXT"
	outputMask := "OfferAndDemandCurve_HH_YYYYMMDD.TXT"
	
	return &SupplyDemandCurveDownloader{
		GeneralDownloader: NewGeneralDownloader(urlMask, outputMask),
		hour:              hour,
	}
}

// generateURL generates the URL for a specific date, replacing HH with hour
func (d *SupplyDemandCurveDownloader) generateURL(date time.Time) string {
	url := d.GetCompleteURL()
	url = strings.ReplaceAll(url, "YYYY", fmt.Sprintf("%04d", date.Year()))
	url = strings.ReplaceAll(url, "MM", fmt.Sprintf("%02d", date.Month()))
	url = strings.ReplaceAll(url, "DD", fmt.Sprintf("%02d", date.Day()))
	url = strings.ReplaceAll(url, "HH", fmt.Sprintf("%d", d.hour))
	return url
}

// generateFilename generates the output filename, replacing HH with hour
func (d *SupplyDemandCurveDownloader) generateFilename(date time.Time) string {
	filename := d.outputMask
	filename = strings.ReplaceAll(filename, "YYYY", fmt.Sprintf("%04d", date.Year()))
	filename = strings.ReplaceAll(filename, "MM", fmt.Sprintf("%02d", date.Month()))
	filename = strings.ReplaceAll(filename, "DD", fmt.Sprintf("%02d", date.Day()))
	filename = strings.ReplaceAll(filename, "HH", fmt.Sprintf("%d", d.hour))
	return filename
}