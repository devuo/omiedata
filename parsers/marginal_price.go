package parsers

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/devuo/omiedata/types"
)

// MarginalPriceParser parses marginal price files
type MarginalPriceParser struct {
	conceptsToLoad []types.DataTypeInMarginalPriceFile
}

// NewMarginalPriceParser creates a new marginal price parser
func NewMarginalPriceParser(conceptsToLoad ...types.DataTypeInMarginalPriceFile) *MarginalPriceParser {
	if len(conceptsToLoad) == 0 {
		// Load all concepts by default
		conceptsToLoad = []types.DataTypeInMarginalPriceFile{
			types.PriceSpain,
			types.PricePortugal,
			types.EnergyIberian,
			types.EnergyIberianWithBilateral,
			types.EnergyBuySpain,
			types.EnergySellSpain,
		}
	}
	
	return &MarginalPriceParser{
		conceptsToLoad: conceptsToLoad,
	}
}

// ParseResponse parses marginal price data from an HTTP response
func (p *MarginalPriceParser) ParseResponse(resp *http.Response) (interface{}, error) {
	reader := NewISO88591Reader(resp.Body)
	return p.ParseReader(reader)
}

// ParseFile parses marginal price data from a file
func (p *MarginalPriceParser) ParseFile(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, types.NewOMIEError(types.ErrCodeParse, "failed to open file", err)
	}
	defer file.Close()

	reader := NewISO88591Reader(file)
	return p.ParseReader(reader)
}

// ParseReader parses marginal price data from a reader
func (p *MarginalPriceParser) ParseReader(reader io.Reader) (interface{}, error) {
	lines, err := ReadLines(reader)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "empty file", nil)
	}

	// Parse date from first line
	date, err := p.parseDateFromHeader(lines[0])
	if err != nil {
		return nil, err
	}

	// Create result structure
	result := types.NewMarginalPriceData(date)
	records := []types.MarginalPriceRecord{}

	// Process all lines looking for data rows
	for _, line := range lines[1:] { // Skip header line
		if strings.TrimSpace(line) == "" {
			continue
		}

		record, err := p.parseDataLine(line, date)
		if err != nil {
			// Skip invalid lines but continue processing
			continue
		}

		if record != nil {
			records = append(records, *record)
			p.addRecordToResult(result, *record)
		}
	}

	if len(records) == 0 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "no valid data found", nil)
	}

	return result, nil
}

// parseDateFromHeader extracts the date from the header line
func (p *MarginalPriceParser) parseDateFromHeader(headerLine string) (time.Time, error) {
	// Use regex to find dates in DD/MM/YYYY format
	dateRegex := regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)
	matches := dateRegex.FindAllString(headerLine, -1)
	
	if len(matches) < 2 {
		return time.Time{}, types.NewOMIEError(types.ErrCodeParse, "expected at least 2 dates in header", nil)
	}

	// The second date is the one we want (data date)
	return ParseDate(matches[1])
}

// parseDataLine parses a single data line
func (p *MarginalPriceParser) parseDataLine(line string, date time.Time) (*types.MarginalPriceRecord, error) {
	fields := SplitCSV(line)
	if len(fields) < 2 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "insufficient fields in line", nil)
	}

	concept := strings.TrimSpace(fields[0])
	
	// Map Spanish concepts to our enum types
	conceptType, multiplier := p.mapConcept(concept)
	if conceptType == "" {
		return nil, nil // Not a concept we're interested in
	}

	// Check if this concept should be loaded
	shouldLoad := false
	for _, c := range p.conceptsToLoad {
		if c == conceptType {
			shouldLoad = true
			break
		}
	}
	
	if !shouldLoad {
		return nil, nil
	}

	// Parse hourly values
	values := make(map[int]float64)
	for i, field := range fields[1:] {
		if i >= 25 { // Maximum 25 hours (for DST)
			break
		}
		
		hour := i + 1 // Hours are 1-based
		if strings.TrimSpace(field) == "" {
			continue // Skip empty values
		}

		value, err := ParseFloat(field)
		if err != nil {
			continue // Skip invalid values
		}

		// Apply multiplier (for old format conversion)
		values[hour] = value * multiplier
	}

	return &types.MarginalPriceRecord{
		Date:    date,
		Concept: conceptType,
		Values:  values,
	}, nil
}

// mapConcept maps Spanish concept names to our enum types and returns multiplier
func (p *MarginalPriceParser) mapConcept(concept string) (types.DataTypeInMarginalPriceFile, float64) {
	conceptMap := map[string][2]interface{}{
		// Old format (Cent/kWh) - multiply by 10 to get EUR/MWh
		"Precio marginal (Cent/kWh)": {types.PriceSpain, 10.0},
		"Precio marginal en el sistema español (Cent/kWh)": {types.PriceSpain, 10.0},
		"Precio marginal en el sistema portugués (Cent/kWh)": {types.PricePortugal, 10.0},
		
		// New format (EUR/MWh)
		"Precio marginal (EUR/MWh)": {types.PriceSpain, 1.0},
		"Precio marginal en el sistema español (EUR/MWh)": {types.PriceSpain, 1.0},
		"Precio marginal en el sistema portugués (EUR/MWh)": {types.PricePortugal, 1.0},
		
		// Adjustment prices (also map to Spain/Portugal prices)
		"Precio de ajuste en el sistema español (EUR/MWh)": {types.PriceSpain, 1.0},
		"Precio de ajuste en el sistema portugués (EUR/MWh)": {types.PricePortugal, 1.0},
		
		// Energy concepts
		"Demanda+bombeos (MWh)": {types.EnergyIberian, 1.0},
		"Energía en el programa resultante de la casación (MWh)": {types.EnergyIberian, 1.0},
		"Energía total del mercado Ibérico (MWh)": {types.EnergyIberian, 1.0},
		"Energía total con bilaterales del mercado Ibérico (MWh)": {types.EnergyIberianWithBilateral, 1.0},
		"Energía total de compra sistema español (MWh)": {types.EnergyBuySpain, 1.0},
		"Energía total de venta sistema español (MWh)": {types.EnergySellSpain, 1.0},
		"Energía horaria sujeta al mecanismo de ajuste a los consumidores MIBEL (MWh)": {types.EnergyIberian, 1.0},
	}

	if mapping, exists := conceptMap[concept]; exists {
		return mapping[0].(types.DataTypeInMarginalPriceFile), mapping[1].(float64)
	}

	return "", 0.0
}

// addRecordToResult adds a parsed record to the result structure
func (p *MarginalPriceParser) addRecordToResult(result *types.MarginalPriceData, record types.MarginalPriceRecord) {
	switch record.Concept {
	case types.PriceSpain:
		for hour, value := range record.Values {
			result.SpainPrices[hour] = value
		}
	case types.PricePortugal:
		for hour, value := range record.Values {
			result.PortugalPrices[hour] = value
		}
	case types.EnergyBuySpain:
		for hour, value := range record.Values {
			result.SpainBuyEnergy[hour] = value
		}
	case types.EnergySellSpain:
		for hour, value := range record.Values {
			result.SpainSellEnergy[hour] = value
		}
	case types.EnergyIberian:
		for hour, value := range record.Values {
			result.IberianEnergy[hour] = value
		}
	case types.EnergyIberianWithBilateral:
		for hour, value := range record.Values {
			result.BilateralEnergy[hour] = value
		}
	}
}