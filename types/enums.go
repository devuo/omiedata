package types

// SystemType represents the different market systems
type SystemType int

const (
	Spain    SystemType = 1
	Portugal SystemType = 2
	Iberian  SystemType = 9
)

// String returns the string representation of SystemType
func (s SystemType) String() string {
	switch s {
	case Spain:
		return "SPAIN"
	case Portugal:
		return "PORTUGAL"
	case Iberian:
		return "IBERIAN"
	default:
		return "UNKNOWN"
	}
}

// TechnologyType represents different energy generation technologies
type TechnologyType string

const (
	Coal               TechnologyType = "COAL"
	FuelGas            TechnologyType = "FUEL_GAS"
	SelfProducer       TechnologyType = "SELF_PRODUCER"
	Nuclear            TechnologyType = "NUCLEAR"
	Hydro              TechnologyType = "HYDRO"
	CombinedCycle      TechnologyType = "COMBINED_CYCLE"
	Wind               TechnologyType = "WIND"
	ThermalSolar       TechnologyType = "THERMAL_SOLAR"
	PhotovoltaicSolar  TechnologyType = "PHOTOVOLTAIC_SOLAR"
	Residuals          TechnologyType = "RESIDUALS"
	Import             TechnologyType = "IMPORT"
	ImportWithoutMIBEL TechnologyType = "IMPORT_WITHOUT_MIBEL"
)

// NameInFile returns the Spanish name as it appears in OMIE files
func (t TechnologyType) NameInFile() string {
	switch t {
	case Coal:
		return "CARBÓN"
	case FuelGas:
		return "FUEL-GAS"
	case SelfProducer:
		return "AUTOPRODUCTOR"
	case Nuclear:
		return "NUCLEAR"
	case Hydro:
		return "HIDRÁULICA"
	case CombinedCycle:
		return "CICLO COMBINADO"
	case Wind:
		return "EÓLICA"
	case ThermalSolar:
		return "SOLAR TÉRMICA"
	case PhotovoltaicSolar:
		return "SOLAR FOTOVOLTAICA"
	case Residuals:
		return "COGENERACIÓN/RESIDUOS/MINI HIDRA"
	case Import:
		return "IMPORTACIÓN INTER."
	case ImportWithoutMIBEL:
		return "IMPORTACIÓN INTER. SIN MIBEL"
	default:
		return string(t)
	}
}

// TechnologyTypeFromSpanish converts Spanish name to TechnologyType
func TechnologyTypeFromSpanish(spanish string) TechnologyType {
	switch spanish {
	case "CARBÓN":
		return Coal
	case "FUEL-GAS":
		return FuelGas
	case "AUTOPRODUCTOR":
		return SelfProducer
	case "NUCLEAR":
		return Nuclear
	case "HIDRÁULICA":
		return Hydro
	case "CICLO COMBINADO":
		return CombinedCycle
	case "EÓLICA":
		return Wind
	case "SOLAR TÉRMICA":
		return ThermalSolar
	case "SOLAR FOTOVOLTAICA":
		return PhotovoltaicSolar
	case "COGENERACIÓN/RESIDUOS/MINI HIDRA":
		return Residuals
	case "IMPORTACIÓN INTER.":
		return Import
	case "IMPORTACIÓN INTER. SIN MIBEL":
		return ImportWithoutMIBEL
	default:
		return TechnologyType(spanish)
	}
}

// DataTypeInMarginalPriceFile represents the different data types in marginal price files
type DataTypeInMarginalPriceFile string

const (
	PriceSpain                 DataTypeInMarginalPriceFile = "PRICE_SP"
	PricePortugal              DataTypeInMarginalPriceFile = "PRICE_PT"
	EnergyIberian              DataTypeInMarginalPriceFile = "ENER_IB"
	EnergyIberianWithBilateral DataTypeInMarginalPriceFile = "ENER_IB_BILLAT"
	EnergyBuySpain             DataTypeInMarginalPriceFile = "ENER_BUY_SP"
	EnergySellSpain            DataTypeInMarginalPriceFile = "ENER_SELL_SP"
)

// SessionType represents intraday market sessions
type SessionType int

const (
	Session1 SessionType = 1
	Session2 SessionType = 2
	Session3 SessionType = 3
	Session4 SessionType = 4
	Session5 SessionType = 5
	Session6 SessionType = 6
)

// OfferType represents market offer types
type OfferType string

const (
	Buy  OfferType = "C" // Compra/Demand
	Sell OfferType = "V" // Venta/Supply
)

// MatchedStatus represents whether an offer was matched
type MatchedStatus string

const (
	Offered MatchedStatus = "O" // Ofertada
	Matched MatchedStatus = "C" // Casada
)
