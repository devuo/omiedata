package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/devuo/omiedata/importers"
	"github.com/devuo/omiedata/types"
)

func main() {
	// Create importer for Iberian system with verbose output
	options := importers.ImportOptions{
		Verbose: true,
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
		MaxConcurrent: 1,
	}
	importer := importers.NewEnergyByTechnologyImporter(types.Iberian, options)

	// Import data for a specific date
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	fmt.Printf("Importing energy by technology data for %s (Iberian system)\n",
		date.Format("2006-01-02"))

	ctx := context.Background()
	result, err := importer.ImportSingleDate(ctx, date)
	if err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	dayData, ok := result.(*types.TechnologyEnergyDay)
	if !ok {
		log.Fatal("Unexpected result type")
	}

	fmt.Printf("\nEnergy data for %s (%s system):\n",
		dayData.Date.Format("2006-01-02"), dayData.System)
	fmt.Printf("Number of hours: %d\n", len(dayData.Records))

	// Show first few hours
	fmt.Println("\nTechnology breakdown (first 3 hours):")
	for i, record := range dayData.Records {
		if i >= 3 {
			break
		}

		fmt.Printf("\nHour %d:\n", record.Hour)
		printIfNotNaN("  Nuclear", record.Nuclear)
		printIfNotNaN("  Wind", record.Wind)
		printIfNotNaN("  Solar PV", record.SolarPV)
		printIfNotNaN("  Solar Therm", record.SolarThermal)
		printIfNotNaN("  Hydro", record.Hydro)
		printIfNotNaN("  Combined", record.CombinedCycle)
		printIfNotNaN("  Coal", record.Coal)
		printIfNotNaN("  Imports", record.ImportInt)

		// Calculate total renewable energy
		renewable := sumNonNaN(record.Wind, record.SolarPV, record.SolarThermal, record.Hydro)
		total := sumNonNaN(renewable, record.Nuclear, record.CombinedCycle, record.Coal, 
			record.FuelGas, record.Cogeneration, record.SelfProducer)

		if total > 0 && !math.IsNaN(renewable) {
			renewablePct := (renewable / total) * 100
			fmt.Printf("  Renewable %%:  %8.1f%%\n", renewablePct)
		}
	}

	// Calculate daily totals
	var dailyTotals map[string]float64 = make(map[string]float64)

	for _, record := range dayData.Records {
		addToTotal(dailyTotals, "Nuclear", record.Nuclear)
		addToTotal(dailyTotals, "Wind", record.Wind)
		addToTotal(dailyTotals, "Solar PV", record.SolarPV)
		addToTotal(dailyTotals, "Solar Thermal", record.SolarThermal)
		addToTotal(dailyTotals, "Hydro", record.Hydro)
		addToTotal(dailyTotals, "Combined Cycle", record.CombinedCycle)
		addToTotal(dailyTotals, "Coal", record.Coal)
		addToTotal(dailyTotals, "Fuel Gas", record.FuelGas)
		addToTotal(dailyTotals, "Cogeneration", record.Cogeneration)
		addToTotal(dailyTotals, "Imports", record.ImportInt)
	}

	fmt.Println("\nDaily totals:")
	for tech, total := range dailyTotals {
		if total > 0 {
			fmt.Printf("  %-15s: %10.1f MWh\n", tech, total)
		}
	}
}

// printIfNotNaN prints a value only if it's not NaN
func printIfNotNaN(label string, value float64) {
	if !math.IsNaN(value) {
		fmt.Printf("%-14s: %8.1f MWh\n", label, value)
	}
}

// sumNonNaN sums values, ignoring NaN values
func sumNonNaN(values ...float64) float64 {
	sum := 0.0
	hasValue := false
	for _, v := range values {
		if !math.IsNaN(v) {
			sum += v
			hasValue = true
		}
	}
	if !hasValue {
		return math.NaN()
	}
	return sum
}

// addToTotal adds a value to the total if it's not NaN
func addToTotal(totals map[string]float64, key string, value float64) {
	if !math.IsNaN(value) {
		totals[key] += value
	}
}
