package types

import "time"

// MarginalPriceData contains the marginal prices and energy data for a specific date
type MarginalPriceData struct {
	Date            time.Time
	SpainPrices     map[int]float64 // hour (1-24) -> EUR/MWh
	PortugalPrices  map[int]float64 // hour (1-24) -> EUR/MWh
	SpainBuyEnergy  map[int]float64 // hour (1-24) -> MWh
	SpainSellEnergy map[int]float64 // hour (1-24) -> MWh
	IberianEnergy   map[int]float64 // hour (1-24) -> MWh
	BilateralEnergy map[int]float64 // hour (1-24) -> MWh
}

// NewMarginalPriceData creates a new MarginalPriceData with initialized maps
func NewMarginalPriceData(date time.Time) *MarginalPriceData {
	return &MarginalPriceData{
		Date:            date,
		SpainPrices:     make(map[int]float64),
		PortugalPrices:  make(map[int]float64),
		SpainBuyEnergy:  make(map[int]float64),
		SpainSellEnergy: make(map[int]float64),
		IberianEnergy:   make(map[int]float64),
		BilateralEnergy: make(map[int]float64),
	}
}

// TechnologyEnergy contains energy generation by technology for a specific hour
type TechnologyEnergy struct {
	Date          time.Time
	Hour          int
	System        SystemType
	Coal          float64 // MWh
	FuelGas       float64 // MWh
	SelfProducer  float64 // MWh
	Nuclear       float64 // MWh
	Hydro         float64 // MWh
	CombinedCycle float64 // MWh
	Wind          float64 // MWh
	SolarThermal  float64 // MWh
	SolarPV       float64 // MWh
	Cogeneration  float64 // MWh (includes residuals and mini-hydro)
	ImportInt     float64 // MWh
	ImportNoMIBEL float64 // MWh
}

// MarketPoint represents a single point in the supply/demand curve
type MarketPoint struct {
	Energy  float64       // MWh
	Price   float64       // EUR/MWh
	Matched MatchedStatus // Offered (O) or Matched (C)
}

// MarketCurve contains the supply and demand curves for a specific hour
type MarketCurve struct {
	Date   time.Time
	Hour   int
	Supply []MarketPoint // Sell offers (Tipo "V")
	Demand []MarketPoint // Buy offers (Tipo "C")
}

// IntradayPrice contains intraday session prices
type IntradayPrice struct {
	Date           time.Time
	Session        SessionType
	Hour           int
	SpainPrice     float64 // EUR/MWh
	PortugalPrice  float64 // EUR/MWh
	SpainEnergy    float64 // MWh
	PortugalEnergy float64 // MWh
}

// MarginalPriceRecord represents a single record from marginal price file parsing
type MarginalPriceRecord struct {
	Date    time.Time
	Concept DataTypeInMarginalPriceFile
	Values  map[int]float64 // hour -> value
}

// TechnologyEnergyDay contains all technology energy data for a single day
type TechnologyEnergyDay struct {
	Date    time.Time
	System  SystemType
	Records []TechnologyEnergy // One record per hour
}

// MarketCurveDay contains all market curves for a single day
type MarketCurveDay struct {
	Date   time.Time
	Curves []MarketCurve // One curve per hour
}

// IntradaySession contains all prices for a single intraday session
type IntradaySession struct {
	Date    time.Time
	Session SessionType
	Prices  []IntradayPrice // One price per hour
}
