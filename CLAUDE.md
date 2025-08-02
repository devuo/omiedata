# CLAUDE.md - OMIEData Go Library Development Guide

This file provides guidance to Claude Code (claude.ai/code) for maintaining and extending the OMIEData Go library.

## Overview

This is a Go library that provides access to [OMIE](https://www.omie.es/) (Iberian Electricity Market Operator) data. The library implements parsers and downloaders for various market data types including marginal prices, energy by technology, and intraday prices.

## Project Structure

```
omiedata/
├── types/           # Data structures and enums
├── parsers/         # File parsing logic
├── downloaders/     # HTTP download functionality
├── importers/       # High-level API combining parsers and downloaders
├── examples/        # Example applications
└── testdata/       # Sample files for testing
```

## Key Components

### Types Package (`types/`)
- `enums.go`: System types (Spain=1, Portugal=2, Iberian=9) and technology types
- `data.go`: Data structures for parsed results (MarginalPriceData, TechnologyEnergy, etc.)
- `errors.go`: Custom error types with error codes

### Parsers Package (`parsers/`)
Each parser implements the `Parser` interface:
- `MarginalPriceParser`: Parses daily market price files
- `EnergyByTechnologyParser`: Parses energy generation by technology
- `IntradayPriceParser`: Parses intraday session prices (planned)

### Downloaders Package (`downloaders/`)
Each downloader implements the `Downloader` interface:
- Constructs OMIE URLs based on date patterns
- Handles HTTP requests with retries
- Returns responses or saves to disk

### Importers Package (`importers/`)
High-level API that combines downloaders and parsers:
- `Import()`: Fetches data for a date range
- `ImportSingleDate()`: Fetches data for a single date
- Handles concurrent downloads with configurable limits

## Important Implementation Details

### Character Encoding
OMIE files use ISO-8859-1 (Latin-1) encoding:
```go
import "golang.org/x/text/encoding/charmap"
decoder := charmap.ISO8859_1.NewDecoder()
```

### Number Format
European decimal format (comma as decimal separator):
```go
// Convert "123,45" to 123.45
value := strings.Replace(field, ",", ".", -1)
```

### Date/Time Handling
- Files use DD/MM/YYYY format
- Hours are 1-24 (not 0-23)
- Handle DST changes (some days have 23 or 25 hours)

### URL Patterns
Base URL: `https://www.omie.es/sites/default/files/dados/`

File patterns:
- Marginal Price: `AGNO_YYYY/MES_MM/TXT/INT_PBC_EV_H_1_DD_MM_YYYY_DD_MM_YYYY.TXT`
- Energy by Tech: `AGNO_YYYY/MES_MM/TXT/INT_PBC_TECNOLOGIAS_H_SYS_DD_MM_YYYY_DD_MM_YYYY.TXT`
- Intraday: `AGNO_YYYY/MES_MM/TXT/INT_PIB_EV_H_1_SS_DD_MM_YYYY_DD_MM_YYYY.TXT`

## Adding New Features

### To Add a New Data Type

1. **Define data structures** in `types/data.go`
2. **Create parser** in `parsers/` implementing the `Parser` interface
3. **Create downloader** in `downloaders/` implementing the `Downloader` interface
4. **Create importer** in `importers/` combining parser and downloader
5. **Add tests** using sample files in `testdata/`
6. **Add example** in `examples/`

### Testing

The `testdata/` directory contains sample files from different time periods:
- `PMD_20060101.txt` - Old format (Cent/kWh)
- `PMD_20090601.txt` - Transition format
- `PMD_20221030.txt` - Current format (EUR/MWh)
- `EnergyByTechnology_9_20201113.TXT` - Technology breakdown
- `PrecioIntra_2_20090102.txt` - Intraday prices

Run tests:
```bash
go test ./...
go test -v ./parsers -run TestMarginalPriceParser
```

## Common Issues and Solutions

### Format Changes Over Time
- **Pre-2009**: Prices in Cent/kWh (parser converts to EUR/MWh)
- **2009-2019**: Various format transitions
- **2019+**: Current EUR/MWh format

### DST Handling
Some days have 23 or 25 hours due to daylight saving time changes. Parsers must handle variable hour counts.

### Missing Data
Empty fields in OMIE files indicate missing data. Parsers should handle gracefully.

### Error Handling
Use typed errors from `types/errors.go`:
- `ErrCodeNotFound`: Data not available for date
- `ErrCodeNetwork`: Network/download issues
- `ErrCodeParse`: File parsing errors

## Development Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run examples
go run ./examples/marginal-price
go run ./examples/energy-by-technology

# Build examples (optional)
go build ./examples/marginal-price
go build ./examples/energy-by-technology

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run
```

## Code Style Guidelines

1. Follow standard Go conventions
2. Use meaningful variable names
3. Add comments for exported functions
4. Handle errors explicitly
5. Use context for cancellation support
6. Log errors but don't panic in library code

## Future Enhancements

Currently planned features:
- Intraday price parser and importer
- Additional data types as needed
- Performance optimizations if required