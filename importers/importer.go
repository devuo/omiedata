package importers

import (
	"context"
	"time"
)

// Importer defines the interface for high-level data importers
type Importer interface {
	// Import downloads and parses data for a date range
	Import(ctx context.Context, start, end time.Time) (interface{}, error)
	
	// ImportSingleDate downloads and parses data for a single date
	ImportSingleDate(ctx context.Context, date time.Time) (interface{}, error)
}

// ImportOptions holds configuration options for importing data
type ImportOptions struct {
	Verbose       bool
	MaxRetries    int
	RetryDelay    time.Duration
	MaxConcurrent int
}