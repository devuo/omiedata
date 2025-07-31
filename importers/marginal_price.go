package importers

import (
	"context"
	"fmt"
	"time"

	"github.com/devuo/omiedata/downloaders"
	"github.com/devuo/omiedata/parsers"
	"github.com/devuo/omiedata/types"
)

// MarginalPriceImporter imports marginal price data
type MarginalPriceImporter struct {
	downloader *downloaders.MarginalPriceDownloader
	parser     *parsers.MarginalPriceParser
	options    ImportOptions
}

// NewMarginalPriceImporter creates a new marginal price importer
func NewMarginalPriceImporter(options ImportOptions) *MarginalPriceImporter {
	downloader := downloaders.NewMarginalPriceDownloader()
	
	// Configure downloader
	config := downloaders.DownloadConfig{
		MaxRetries:     options.MaxRetries,
		RetryDelay:     options.RetryDelay,
		RequestTimeout: 30 * time.Second,
		MaxConcurrent:  options.MaxConcurrent,
	}
	downloader.SetConfig(config)
	
	return &MarginalPriceImporter{
		downloader: downloader,
		parser:     parsers.NewMarginalPriceParser(),
		options:    options,
	}
}

// NewDefaultMarginalPriceImporter creates a marginal price importer with default options
func NewDefaultMarginalPriceImporter() *MarginalPriceImporter {
	return NewMarginalPriceImporter(ImportOptions{
		Verbose:       false,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		MaxConcurrent: 5,
	})
}

// Import downloads and parses marginal price data for a date range
func (i *MarginalPriceImporter) Import(ctx context.Context, start, end time.Time) (interface{}, error) {
	responseChan := i.downloader.URLResponses(ctx, start, end, i.options.Verbose)
	
	var results []*types.MarginalPriceData
	var errors []error
	
	for result := range responseChan {
		if result.Error != nil {
			errors = append(errors, result.Error)
			continue
		}
		
		// Parse the response
		parsed, err := i.parser.ParseResponse(result.Response)
		result.Response.Body.Close()
		
		if err != nil {
			errors = append(errors, fmt.Errorf("parse error for %s: %w", result.Date.Format("2006-01-02"), err))
			continue
		}
		
		if data, ok := parsed.(*types.MarginalPriceData); ok {
			results = append(results, data)
		}
	}
	
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("no data imported, %d errors occurred: %v", len(errors), errors[0])
	}
	
	return results, nil
}

// ImportSingleDate downloads and parses marginal price data for a single date
func (i *MarginalPriceImporter) ImportSingleDate(ctx context.Context, date time.Time) (interface{}, error) {
	results, err := i.Import(ctx, date, date)
	if err != nil {
		return nil, err
	}
	
	if dataList, ok := results.([]*types.MarginalPriceData); ok && len(dataList) > 0 {
		return dataList[0], nil
	}
	
	return nil, types.NewOMIEError(types.ErrCodeNotFound, "no data found for date", nil)
}

// ImportToDataFrame imports data and returns it in a flattened format
// This method provides a pandas-like interface for easier data analysis
func (i *MarginalPriceImporter) ImportToDataFrame(ctx context.Context, start, end time.Time) ([]types.MarginalPriceRecord, error) {
	results, err := i.Import(ctx, start, end)
	if err != nil {
		return nil, err
	}
	
	dataList, ok := results.([]*types.MarginalPriceData)
	if !ok {
		return nil, types.NewOMIEError(types.ErrCodeParse, "unexpected result type", nil)
	}
	
	var records []types.MarginalPriceRecord
	
	for _, data := range dataList {
		// Convert to flattened records
		if len(data.SpainPrices) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.PriceSpain,
				Values:  data.SpainPrices,
			})
		}
		
		if len(data.PortugalPrices) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.PricePortugal,
				Values:  data.PortugalPrices,
			})
		}
		
		if len(data.IberianEnergy) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.EnergyIberian,
				Values:  data.IberianEnergy,
			})
		}
		
		if len(data.BilateralEnergy) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.EnergyIberianWithBilateral,
				Values:  data.BilateralEnergy,
			})
		}
		
		if len(data.SpainBuyEnergy) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.EnergyBuySpain,
				Values:  data.SpainBuyEnergy,
			})
		}
		
		if len(data.SpainSellEnergy) > 0 {
			records = append(records, types.MarginalPriceRecord{
				Date:    data.Date,
				Concept: types.EnergySellSpain,
				Values:  data.SpainSellEnergy,
			})
		}
	}
	
	return records, nil
}