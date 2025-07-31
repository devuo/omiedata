# OMIEData Go Library

A Go library for accessing electricity market data from OMIE (Iberian Peninsula's Electricity Market Operator). This library provides data access for daily market (hourly prices, energy by technology, bid/ask curves) and intra-day market data for Spain and Portugal.

This is a Go port of the [OMIEData Python library](https://github.com/example/omiedata-python), designed to provide better performance and easier deployment while maintaining the same functionality.

## Features

- **Marginal Prices**: Hourly electricity prices for Spain and Portugal
- **Energy by Technology**: Generation breakdown by source (wind, solar, nuclear, etc.)
- **Supply/Demand Curves**: Market bid/ask curves (planned)
- **Intraday Prices**: Prices for the 6 daily adjustment sessions (planned)
- **Concurrent Downloads**: Fast parallel data fetching
- **Multiple Formats**: Support for historical format changes
- **Type Safety**: Full Go type safety with proper error handling

## Installation

```bash
go get github.com/devuo/omiedata
```

## Quick Start

### Marginal Prices

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/devuo/omiedata"
)

func main() {
    // Create importer
    importer := omiedata.NewMarginalPriceImporter()

    // Get data for yesterday
    ctx := context.Background()
    yesterday := time.Now().AddDate(0, 0, -1)
    
    data, err := importer.ImportSingleDate(ctx, yesterday)
    if err != nil {
        log.Fatal(err)
    }

    priceData := data.(*omiedata.MarginalPriceData)
    
    fmt.Printf("Date: %s\n", priceData.Date.Format("2006-01-02"))
    
    // Print hourly prices
    for hour := 1; hour <= 24; hour++ {
        if price, exists := priceData.SpainPrices[hour]; exists {
            fmt.Printf("Hour %2d: %.2f EUR/MWh\n", hour, price)
        }
    }
}
```

### Energy by Technology

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/devuo/omiedata"
)

func main() {
    // Create importer for Iberian system
    importer := omiedata.NewEnergyByTechnologyImporter(omiedata.Iberian)

    ctx := context.Background()
    date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
    
    result, err := importer.ImportSingleDate(ctx, date)
    if err != nil {
        log.Fatal(err)
    }

    dayData := result.(*omiedata.TechnologyEnergyDay)
    
    fmt.Printf("Energy data for %s:\n", dayData.Date.Format("2006-01-02"))
    
    // Show renewable energy for each hour
    for _, record := range dayData.Records {
        renewable := record.Wind + record.SolarPV + record.SolarThermal + record.Hydro
        fmt.Printf("Hour %2d: %.1f MWh renewable\n", record.Hour, renewable)
    }
}
```

### Date Range Import

```go
// Import data for a week
start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end := start.AddDate(0, 0, 6)

results, err := importer.Import(ctx, start, end)
if err != nil {
    log.Fatal(err)
}

dataList := results.([]*omiedata.MarginalPriceData)
fmt.Printf("Imported %d days of data\n", len(dataList))
```

## Configuration

You can customize the import behavior with options:

```go
options := omiedata.ImportOptions{
    Verbose:       true,           // Enable verbose logging
    MaxRetries:    5,              // Number of download retries
    RetryDelay:    2 * time.Second, // Delay between retries
    MaxConcurrent: 3,              // Maximum concurrent downloads
}

importer := omiedata.NewMarginalPriceImporterWithOptions(options)
```

## Data Types

### MarginalPriceData

Contains hourly electricity prices and energy volumes:

```go
type MarginalPriceData struct {
    Date            time.Time
    SpainPrices     map[int]float64 // hour (1-24) -> EUR/MWh
    PortugalPrices  map[int]float64 // hour (1-24) -> EUR/MWh
    SpainBuyEnergy  map[int]float64 // hour (1-24) -> MWh
    SpainSellEnergy map[int]float64 // hour (1-24) -> MWh
    IberianEnergy   map[int]float64 // hour (1-24) -> MWh
    BilateralEnergy map[int]float64 // hour (1-24) -> MWh
}
```

### TechnologyEnergy

Contains energy generation by technology for a specific hour:

```go
type TechnologyEnergy struct {
    Date              time.Time
    Hour              int
    System            SystemType
    Coal              float64 // MWh
    Nuclear           float64 // MWh
    Wind              float64 // MWh
    SolarPV           float64 // MWh
    // ... other technologies
}
```

## System Types

- `omiedata.Spain` (1) - Spanish market
- `omiedata.Portugal` (2) - Portuguese market  
- `omiedata.Iberian` (9) - Combined Iberian market

## Error Handling

The library uses structured error types:

```go
data, err := importer.ImportSingleDate(ctx, date)
if err != nil {
    if omieErr, ok := err.(*types.OMIEError); ok {
        switch omieErr.Code {
        case types.ErrCodeNotFound:
            fmt.Println("Data not available for this date")
        case types.ErrCodeNetwork:
            fmt.Println("Network error occurred")
        case types.ErrCodeParse:
            fmt.Println("Failed to parse data")
        }
    }
    return err
}
```

## Historical Data Format Changes

The library automatically handles OMIE's format changes over time:

- **Pre-2009**: Prices in Cent/kWh (automatically converted to EUR/MWh)
- **2009-2019**: Transition period with format variations
- **2019+**: Current EUR/MWh format

## Performance

This Go library is significantly faster than the Python equivalent:

- **Concurrent downloads**: 5 parallel connections by default
- **Streaming parsing**: Low memory usage for large date ranges
- **Native performance**: 10-50x faster than Python pandas processing

## Examples

See the [examples](./examples/) directory for complete working examples:

- `marginal_price_example.go` - Basic price data import
- `energy_by_technology_example.go` - Technology breakdown analysis

## Testing

Run tests with sample data:

```bash
go test ./...
```

The test suite includes sample files from different time periods to ensure compatibility with format changes.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Based on the original Python OMIEData library
- OMIE (Operador del Mercado Ibérico de Energía) for providing the data
- The Go community for excellent tooling and libraries