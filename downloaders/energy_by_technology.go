package downloaders

import (
	"fmt"
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