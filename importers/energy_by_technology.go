package importers

import (
	"context"
	"fmt"
	"time"

	"github.com/devuo/omiedata/downloaders"
	"github.com/devuo/omiedata/parsers"
	"github.com/devuo/omiedata/types"
)

// EnergyByTechnologyImporter imports energy by technology data
type EnergyByTechnologyImporter struct {
	downloader *downloaders.EnergyByTechnologyDownloader
	parser     *parsers.EnergyByTechnologyParser
	options    ImportOptions
	systemType types.SystemType
}

// NewEnergyByTechnologyImporter creates a new energy by technology importer
func NewEnergyByTechnologyImporter(systemType types.SystemType, options ImportOptions) *EnergyByTechnologyImporter {
	downloader := downloaders.NewEnergyByTechnologyDownloader(systemType)
	
	// Configure downloader
	config := downloaders.DownloadConfig{
		MaxRetries:     options.MaxRetries,
		RetryDelay:     options.RetryDelay,
		RequestTimeout: 30 * time.Second,
		MaxConcurrent:  options.MaxConcurrent,
	}
	downloader.SetConfig(config)
	
	return &EnergyByTechnologyImporter{
		downloader: downloader,
		parser:     parsers.NewEnergyByTechnologyParser(),
		options:    options,
		systemType: systemType,
	}
}

// NewDefaultEnergyByTechnologyImporter creates an energy by technology importer with default options
func NewDefaultEnergyByTechnologyImporter(systemType types.SystemType) *EnergyByTechnologyImporter {
	return NewEnergyByTechnologyImporter(systemType, ImportOptions{
		Verbose:       false,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		MaxConcurrent: 5,
	})
}

// Import downloads and parses energy by technology data for a date range
func (i *EnergyByTechnologyImporter) Import(ctx context.Context, start, end time.Time) (interface{}, error) {
	responseChan := i.downloader.URLResponses(ctx, start, end, i.options.Verbose)
	
	var results []*types.TechnologyEnergyDay
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
		
		if data, ok := parsed.(*types.TechnologyEnergyDay); ok {
			results = append(results, data)
		}
	}
	
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("no data imported, %d errors occurred: %v", len(errors), errors[0])
	}
	
	return results, nil
}

// ImportSingleDate downloads and parses energy by technology data for a single date
func (i *EnergyByTechnologyImporter) ImportSingleDate(ctx context.Context, date time.Time) (interface{}, error) {
	results, err := i.Import(ctx, date, date)
	if err != nil {
		return nil, err
	}
	
	if dataList, ok := results.([]*types.TechnologyEnergyDay); ok && len(dataList) > 0 {
		return dataList[0], nil
	}
	
	return nil, types.NewOMIEError(types.ErrCodeNotFound, "no data found for date", nil)
}

// ImportToRecords imports data and returns it as a flat list of records
func (i *EnergyByTechnologyImporter) ImportToRecords(ctx context.Context, start, end time.Time) ([]types.TechnologyEnergy, error) {
	results, err := i.Import(ctx, start, end)
	if err != nil {
		return nil, err
	}
	
	dataList, ok := results.([]*types.TechnologyEnergyDay)
	if !ok {
		return nil, types.NewOMIEError(types.ErrCodeParse, "unexpected result type", nil)
	}
	
	var records []types.TechnologyEnergy
	
	for _, dayData := range dataList {
		records = append(records, dayData.Records...)
	}
	
	return records, nil
}