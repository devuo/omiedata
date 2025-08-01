package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/devuo/omiedata/importers"
	"github.com/devuo/omiedata/types"
)

func main() {
	// Create importer with verbose output
	options := importers.ImportOptions{
		Verbose:       true,
		MaxRetries:    3,
		RetryDelay:    time.Second,
		MaxConcurrent: 3,
	}
	importer := importers.NewMarginalPriceImporter(options)

	// Define date range - last week
	end := time.Now().AddDate(0, 0, -1) // Yesterday
	start := end.AddDate(0, 0, -6)      // 7 days ago

	fmt.Printf("Importing marginal price data from %s to %s\n", 
		start.Format("2006-01-02"), end.Format("2006-01-02"))

	ctx := context.Background()
	results, err := importer.Import(ctx, start, end)
	if err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	dataList, ok := results.([]*types.MarginalPriceData)
	if !ok {
		log.Fatal("Unexpected result type")
	}

	fmt.Printf("\nSuccessfully imported data for %d days:\n", len(dataList))
	
	for _, data := range dataList {
		fmt.Printf("\nDate: %s\n", data.Date.Format("2006-01-02"))
		
		// Show some sample prices
		fmt.Println("Spain prices (first 6 hours):")
		for hour := 1; hour <= 6; hour++ {
			if price, exists := data.SpainPrices[hour]; exists {
				fmt.Printf("  Hour %2d: %8.2f EUR/MWh\n", hour, price)
			}
		}
		
		// Show energy data if available
		if len(data.IberianEnergy) > 0 {
			fmt.Println("Iberian energy (first 3 hours):")
			for hour := 1; hour <= 3; hour++ {
				if energy, exists := data.IberianEnergy[hour]; exists {
					fmt.Printf("  Hour %2d: %10.1f MWh\n", hour, energy)
				}
			}
		}
		
		// Calculate daily average price
		var totalPrice, totalHours float64
		for _, price := range data.SpainPrices {
			totalPrice += price
			totalHours++
		}
		if totalHours > 0 {
			avgPrice := totalPrice / totalHours
			fmt.Printf("Average daily price: %.2f EUR/MWh\n", avgPrice)
		}
	}
}