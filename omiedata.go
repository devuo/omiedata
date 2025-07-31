// Package omiedata provides access to OMIE (Iberian Electricity Market Operator) data.
//
// This library allows you to download and parse electricity market data from the OMIE website,
// including marginal prices, energy by technology, supply/demand curves, and intraday prices
// for Spain and Portugal.
//
// Basic usage example:
//
//	importer := omiedata.NewMarginalPriceImporter()
//	data, err := importer.ImportSingleDate(ctx, time.Now().AddDate(0, 0, -1))
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Use data...
package omiedata

import (
	"github.com/devuo/omiedata/importers"
	"github.com/devuo/omiedata/types"
)

// Re-export key types for easier access
type (
	// System types
	SystemType = types.SystemType
	
	// Technology types
	TechnologyType = types.TechnologyType
	
	// Data types
	MarginalPriceData    = types.MarginalPriceData
	TechnologyEnergy     = types.TechnologyEnergy
	TechnologyEnergyDay  = types.TechnologyEnergyDay
	MarketPoint          = types.MarketPoint
	MarketCurve          = types.MarketCurve
	IntradayPrice        = types.IntradayPrice
	
	// Import options
	ImportOptions = importers.ImportOptions
	
	// Importers
	MarginalPriceImporter        = importers.MarginalPriceImporter
	EnergyByTechnologyImporter   = importers.EnergyByTechnologyImporter
)

// System type constants
const (
	Spain    = types.Spain
	Portugal = types.Portugal
	Iberian  = types.Iberian
)

// Technology type constants
const (
	Coal                = types.Coal
	FuelGas             = types.FuelGas
	SelfProducer        = types.SelfProducer
	Nuclear             = types.Nuclear
	Hydro               = types.Hydro
	CombinedCycle       = types.CombinedCycle
	Wind                = types.Wind
	ThermalSolar        = types.ThermalSolar
	PhotovoltaicSolar   = types.PhotovoltaicSolar
	Residuals           = types.Residuals
	Import              = types.Import
	ImportWithoutMIBEL  = types.ImportWithoutMIBEL
)

// Convenience constructor functions

// NewMarginalPriceImporter creates a new marginal price importer with default settings
func NewMarginalPriceImporter() *MarginalPriceImporter {
	return importers.NewDefaultMarginalPriceImporter()
}

// NewMarginalPriceImporterWithOptions creates a new marginal price importer with custom options
func NewMarginalPriceImporterWithOptions(options ImportOptions) *MarginalPriceImporter {
	return importers.NewMarginalPriceImporter(options)
}

// NewEnergyByTechnologyImporter creates a new energy by technology importer with default settings
func NewEnergyByTechnologyImporter(systemType SystemType) *EnergyByTechnologyImporter {
	return importers.NewDefaultEnergyByTechnologyImporter(systemType)
}

// NewEnergyByTechnologyImporterWithOptions creates a new energy by technology importer with custom options
func NewEnergyByTechnologyImporterWithOptions(systemType SystemType, options ImportOptions) *EnergyByTechnologyImporter {
	return importers.NewEnergyByTechnologyImporter(systemType, options)
}