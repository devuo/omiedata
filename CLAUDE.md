# CLAUDE.md - OMIE Go Library Migration Guide

This file provides guidance to Claude Code (claude.ai/code) when migrating the OMIEData Python library to Go.

## Overview

This is a migration project to create a Go library (`omie-go`) that provides the same functionality as the existing Python library (`omie-python`) for accessing OMIE (Iberian Electricity Market Operator) data. The Python library is well-tested and serves as the reference implementation.

## Migration Strategy

### 1. Architecture Mapping

The Go library should follow a similar layered architecture to the Python version:

**Python → Go Package Mapping:**
- `OMIEData/Downloaders/` → `downloaders/` - HTTP client functionality
- `OMIEData/FileReaders/` → `parsers/` - File parsing logic
- `OMIEData/DataImport/` → `importers/` - High-level API
- `OMIEData/Enums/` → `types/` - Enums and data structures

### 2. Core Components to Implement

#### Types Package (`types/`)
```go
// Enums from all_enums.py
type SystemType int
const (
    Spain SystemType = 1
    Portugal SystemType = 2
    Iberian SystemType = 9
)

type TechnologyType string
const (
    Coal TechnologyType = "COAL"
    FuelGas TechnologyType = "FUEL_GAS"
    // ... etc
)

// Data structures for parsed results
type MarginalPriceData struct {
    Date            time.Time
    SpainPrices     map[int]float64  // hour -> EUR/MWh
    PortugalPrices  map[int]float64  // hour -> EUR/MWh
    IberianEnergy   map[int]float64  // hour -> MWh
    BilateralEnergy map[int]float64  // hour -> MWh
}
```

#### Downloaders Package (`downloaders/`)
Base interface similar to `OMIEDownloader`:
```go
type Downloader interface {
    GetCompleteURL() string
    DownloadData(dateIni, dateEnd time.Time, outputFolder string, verbose bool) error
    URLResponses(dateIni, dateEnd time.Time, verbose bool) <-chan Response
}
```

Implement specific downloaders:
- `MarginalPriceDownloader` (from `marginal_price_downloader.py`)
- `EnergyByTechnologyDownloader` (from `energy_by_technology_downloader.py`)
- `SupplyDemandCurveDownloader` (from `supply_demand_curve_downloader.py`)
- `IntradayPriceDownloader` (from `intra_day_price_downloader.py`)

#### Parsers Package (`parsers/`)
File parsing logic from FileReaders:
```go
type Parser interface {
    ParseResponse(resp *http.Response) (interface{}, error)
    ParseFile(filename string) (interface{}, error)
}
```

Key parsers to implement:
- `MarginalPriceParser` (from `marginal_price_file_reader.py`)
- `EnergyByTechnologyParser` (from `energy_by_technology_files_reader.py`)
- `SupplyDemandCurveParser` (from `supply_demand_curve_file_reader.py`)

#### Importers Package (`importers/`)
High-level API combining downloaders and parsers:
```go
type MarginalPriceImporter struct {
    downloader Downloader
    parser     Parser
}

func (i *MarginalPriceImporter) Import(start, end time.Time) ([]MarginalPriceData, error)
```

### 3. Key Implementation Details

#### Character Encoding
Python uses ISO-8859-1 (Latin-1) encoding. In Go:
```go
import "golang.org/x/text/encoding/charmap"

decoder := charmap.ISO8859_1.NewDecoder()
reader := transform.NewReader(response.Body, decoder)
```

#### Number Parsing
Python uses babel for locale-aware parsing. In Go, handle European format:
```go
// Convert "123,45" to 123.45
value := strings.Replace(field, ",", ".", -1)
price, err := strconv.ParseFloat(value, 64)
```

#### Date/Time Handling
- Files use DD/MM/YYYY format
- Hours are 1-24 (not 0-23)
- Handle DST changes (23 or 25 hours)

#### URL Construction
Base URL: `https://www.omie.es/sites/default/files/dados/`

URL patterns from Python downloaders:
- Marginal Price: `AGNO_YYYY/MES_MM/TXT/INT_PBC_EV_H_1_DD_MM_YYYY_DD_MM_YYYY.TXT`
- Energy by Tech: `AGNO_YYYY/MES_MM/TXT/INT_PBC_TECNOLOGIAS_H_SYS_DD_MM_YYYY_DD_MM_YYYY.TXT`
- Supply/Demand: `AGNO_YYYY/MES_MM/TXT/INT_CURVA_ACUM_UO_MIB_1_HH_DD_MM_YYYY_DD_MM_YYYY.TXT`
- Intraday: `AGNO_YYYY/MES_MM/TXT/INT_PIB_EV_H_1_SS_DD_MM_YYYY_DD_MM_YYYY.TXT`

### 4. Testing Strategy

#### Test Data
Copy test files from `omie-python/tests/downloaders_tests/InputTesting/`:
- `PMD_20060101.txt` - Old format (Cent/kWh)
- `PMD_20090601.txt` - Transition format
- `PMD_20221030.txt` - Current format (EUR/MWh)
- `EnergyByTechnology_9_20201113.TXT` - Technology breakdown
- `OfferAndDemandCurve_1_20090102.TXT` - Market curves
- `PrecioIntra_2_20090102.txt` - Intraday prices

#### Test Coverage
1. **Parser Tests**: Test each parser with sample files from different years
2. **Downloader Tests**: Mock HTTP responses or use integration tests
3. **Format Evolution**: Test handling of format changes over time
4. **Edge Cases**: DST changes, missing data, malformed files

### 5. Development Commands

```bash
# Initialize Go module
go mod init github.com/devuo/omiedata

# Run tests
go test ./...

# Build library
go build ./...

# Run specific test
go test ./parsers -run TestMarginalPriceParser

# Generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 6. Implementation Order

1. **Phase 1 - Core Types** (types/)
   - Define all enums (SystemType, TechnologyType, etc.)
   - Define data structures (MarginalPriceData, etc.)

2. **Phase 2 - Parsers** (parsers/)
   - Implement file parsing logic
   - Test with provided sample files
   - Handle format variations by year

3. **Phase 3 - Downloaders** (downloaders/)
   - Implement base downloader with HTTP client
   - Add specific downloaders for each data type
   - Handle retries and error cases

4. **Phase 4 - Importers** (importers/)
   - Combine downloaders and parsers
   - Provide high-level API
   - Add concurrent download support

5. **Phase 5 - Examples**
   - Port Python examples to Go
   - Add documentation and README

### 7. Key Differences from Python

1. **Concurrency**: Use goroutines for parallel downloads
2. **Error Handling**: Explicit error returns instead of exceptions
3. **Data Structures**: Use structs instead of DataFrames
4. **Memory**: Stream processing to handle large date ranges

### 8. Common Pitfalls

1. **Hour Indexing**: OMIE uses 1-24, not 0-23
2. **DST Handling**: Some days have 23 or 25 hours
3. **Empty Values**: Missing data appears as empty strings
4. **Format Changes**: Pre-2009 files use different units
5. **Encoding**: Always decode from ISO-8859-1

### 9. Reference Files

Key Python files to study:
- `marginal_price_file_reader.py` - Complex parsing logic
- `general_omie_downloader.py` - URL construction patterns
- `all_enums.py` - Complete enum definitions
- Test files - Expected parsing results

### 10. Success Criteria

The Go library is complete when:
1. All data types from Python are supported
2. Tests pass with the same test data
3. API is idiomatic Go (not a direct port)
4. Performance is better than Python version
5. Concurrent downloads are supported