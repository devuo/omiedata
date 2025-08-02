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

// EnergyByTechnologyParser parses energy by technology files
type EnergyByTechnologyParser struct{}

// NewEnergyByTechnologyParser creates a new energy by technology parser
func NewEnergyByTechnologyParser() *EnergyByTechnologyParser {
	return &EnergyByTechnologyParser{}
}

// ParseResponse parses energy by technology data from an HTTP response
func (p *EnergyByTechnologyParser) ParseResponse(resp *http.Response) (interface{}, error) {
	reader := NewISO88591Reader(resp.Body)
	return p.ParseReader(reader)
}

// ParseFile parses energy by technology data from a file
func (p *EnergyByTechnologyParser) ParseFile(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, types.NewOMIEError(types.ErrCodeParse, "failed to open file", err)
	}
	defer file.Close()

	reader := NewISO88591Reader(file)
	return p.ParseReader(reader)
}

// ParseReader parses energy by technology data from a reader
func (p *EnergyByTechnologyParser) ParseReader(reader io.Reader) (interface{}, error) {
	lines, err := ReadLines(reader)
	if err != nil {
		return nil, err
	}

	if len(lines) < 3 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "insufficient lines in file", nil)
	}

	// Parse date and system from header
	date, system, err := p.parseHeader(lines[0])
	if err != nil {
		return nil, err
	}

	// Find column headers line and parse column mapping
	columnMapping, headerLineIndex := p.parseColumnHeaders(lines)
	if len(columnMapping) == 0 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "no technology columns found", nil)
	}

	// Parse data lines
	var records []types.TechnologyEnergy
	for i := headerLineIndex + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		record, err := p.parseDataLine(line, date, system, columnMapping)
		if err != nil {
			continue // Skip invalid lines
		}

		records = append(records, *record)
	}

	if len(records) == 0 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "no valid data records found", nil)
	}

	return &types.TechnologyEnergyDay{
		Date:    date,
		System:  system,
		Records: records,
	}, nil
}

// parseHeader extracts date and system type from the header
func (p *EnergyByTechnologyParser) parseHeader(headerLine string) (time.Time, types.SystemType, error) {
	// Extract date
	dateRegex := regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)
	dateMatches := dateRegex.FindAllString(headerLine, -1)

	if len(dateMatches) == 0 {
		return time.Time{}, 0, types.NewOMIEError(types.ErrCodeParse, "no date found in header", nil)
	}

	date, err := ParseDate(dateMatches[len(dateMatches)-1]) // Use the last date found
	if err != nil {
		return time.Time{}, 0, err
	}

	// Determine system type from header content
	system := types.Iberian // Default
	if strings.Contains(strings.ToLower(headerLine), "español") {
		system = types.Spain
	} else if strings.Contains(strings.ToLower(headerLine), "portugués") {
		system = types.Portugal
	}

	return date, system, nil
}

// parseColumnHeaders finds and parses the column headers to create technology mapping
func (p *EnergyByTechnologyParser) parseColumnHeaders(lines []string) (map[int]types.TechnologyType, int) {
	for i, line := range lines {
		fields := SplitCSV(line)
		if len(fields) < 3 {
			continue
		}

		// Check if this looks like a header line (contains technology names)
		if p.containsTechnologyNames(fields) {
			mapping := make(map[int]types.TechnologyType)

			for j, field := range fields {
				field = strings.TrimSpace(field)
				// Only add to mapping if it's a recognized technology
				if _, ok := isKnownTechnology(field); ok {
					tech := types.TechnologyTypeFromSpanish(field)
					mapping[j] = tech
				}
			}

			return mapping, i
		}
	}

	return nil, -1
}

// containsTechnologyNames checks if fields contain technology names
func (p *EnergyByTechnologyParser) containsTechnologyNames(fields []string) bool {
	knownTechs := []string{"CARBÓN", "NUCLEAR", "EÓLICA", "SOLAR", "HIDRÁULICA"}

	for _, field := range fields {
		field = strings.TrimSpace(strings.ToUpper(field))
		for _, tech := range knownTechs {
			if strings.Contains(field, tech) {
				return true
			}
		}
	}

	return false
}

// isKnownTechnology checks if a field name is a known technology
func isKnownTechnology(field string) (types.TechnologyType, bool) {
	knownTechs := map[string]types.TechnologyType{
		"CARBÓN":                           types.Coal,
		"FUEL-GAS":                         types.FuelGas,
		"AUTOPRODUCTOR":                    types.SelfProducer,
		"NUCLEAR":                          types.Nuclear,
		"HIDRÁULICA":                       types.Hydro,
		"CICLO COMBINADO":                  types.CombinedCycle,
		"EÓLICA":                           types.Wind,
		"SOLAR TÉRMICA":                    types.ThermalSolar,
		"SOLAR FOTOVOLTAICA":               types.PhotovoltaicSolar,
		"COGENERACIÓN/RESIDUOS/MINI HIDRA": types.Residuals,
		"IMPORTACIÓN INTER.":               types.Import,
		"IMPORTACIÓN INTER. SIN MIBEL":     types.ImportWithoutMIBEL,
	}

	tech, ok := knownTechs[field]
	return tech, ok
}

// parseDataLine parses a single data line
func (p *EnergyByTechnologyParser) parseDataLine(line string, date time.Time, system types.SystemType, columnMapping map[int]types.TechnologyType) (*types.TechnologyEnergy, error) {
	fields := SplitCSV(line)
	if len(fields) < 3 {
		return nil, types.NewOMIEError(types.ErrCodeParse, "insufficient fields", nil)
	}

	// Parse hour (usually in second column)
	hour, err := ParseHour(fields[1])
	if err != nil {
		return nil, err
	}

	// Create record
	record := &types.TechnologyEnergy{
		Date:   date,
		Hour:   hour,
		System: system,
	}

	// Parse technology values
	for colIndex, techType := range columnMapping {
		if colIndex >= len(fields) {
			continue
		}

		value, err := ParseFloat(fields[colIndex])
		if err != nil {
			continue // Skip invalid values
		}

		// Assign to appropriate field
		p.assignTechnologyValue(record, techType, value)
	}

	return record, nil
}

// assignTechnologyValue assigns a value to the appropriate field in TechnologyEnergy
func (p *EnergyByTechnologyParser) assignTechnologyValue(record *types.TechnologyEnergy, techType types.TechnologyType, value float64) {
	switch techType {
	case types.Coal:
		record.Coal = value
	case types.FuelGas:
		record.FuelGas = value
	case types.SelfProducer:
		record.SelfProducer = value
	case types.Nuclear:
		record.Nuclear = value
	case types.Hydro:
		record.Hydro = value
	case types.CombinedCycle:
		record.CombinedCycle = value
	case types.Wind:
		record.Wind = value
	case types.ThermalSolar:
		record.SolarThermal = value
	case types.PhotovoltaicSolar:
		record.SolarPV = value
	case types.Residuals:
		record.Cogeneration = value
	case types.Import:
		record.ImportInt = value
	case types.ImportWithoutMIBEL:
		record.ImportNoMIBEL = value
	}
}
