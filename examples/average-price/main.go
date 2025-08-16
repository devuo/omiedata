package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/devuo/omiedata/importers"
	"github.com/devuo/omiedata/types"
)

func main() {
	var startDate, endDate string
	flag.StringVar(&startDate, "start", "", "Start date in DD-MM-YYYY format")
	flag.StringVar(&endDate, "end", "", "End date in DD-MM-YYYY format")
	flag.Parse()

	if startDate == "" || endDate == "" {
		log.Fatal("Usage: average-price -start DD-MM-YYYY -end DD-MM-YYYY")
	}

	start, err := parseDate(startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}

	end, err := parseDate(endDate)
	if err != nil {
		log.Fatalf("Invalid end date: %v", err)
	}

	// Add one day to end date to include it fully (up to midnight of next day)
	end = end.AddDate(0, 0, 1)

	fmt.Printf("Fetching OMIE data from %s to %s (exclusive)\n",
		start.Format("02-01-2006"), end.Format("02-01-2006"))

	ctx := context.Background()
	importer := importers.NewDefaultMarginalPriceImporter()

	// Fetch data for the date range
	results, err := importer.Import(ctx, start, end)
	if err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	// Calculate average PT price
	var totalPrice float64
	var totalHours int

	dataList, ok := results.([]*types.MarginalPriceData)
	if !ok {
		log.Fatal("Unexpected result type from importer")
	}

	for _, data := range dataList {
		// Get Portugal prices for each hour
		for _, price := range data.PortugalPrices {
			if price > 0 {
				totalPrice += price
				totalHours++
			}
		}
	}

	if totalHours == 0 {
		log.Fatal("No valid PT price data found in the specified range")
	}

	avgPrice := totalPrice / float64(totalHours)
	fmt.Printf("\nAverage PT price: %.2f EUR/MWh\n", avgPrice)
	fmt.Printf("Based on %d hours of data\n", totalHours)
	fmt.Printf("Date range: %s to %s\n",
		start.Format("02-01-2006"),
		end.AddDate(0, 0, -1).Format("02-01-2006"))
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("02-01-2006", dateStr)
}
