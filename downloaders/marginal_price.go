package downloaders

// MarginalPriceDownloader downloads marginal price data files
type MarginalPriceDownloader struct {
	*GeneralDownloader
}

// NewMarginalPriceDownloader creates a new marginal price downloader
func NewMarginalPriceDownloader() *MarginalPriceDownloader {
	urlMask := "AGNO_YYYY/MES_MM/TXT/INT_PBC_EV_H_1_DD_MM_YYYY_DD_MM_YYYY.TXT"
	outputMask := "PMD_YYYYMMDD.txt"
	
	return &MarginalPriceDownloader{
		GeneralDownloader: NewGeneralDownloader(urlMask, outputMask),
	}
}