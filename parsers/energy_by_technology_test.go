package parsers

import (
	"math"
	"testing"
	"time"

	"github.com/devuo/omiedata/types"
)

func TestEnergyByTechnologyParser_ParseFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "parse energy by technology file",
			filename: "../testdata/EnergyByTechnology_9_20201113.TXT",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewEnergyByTechnologyParser()
			result, err := parser.ParseFile(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			data, ok := result.(*types.TechnologyEnergyDay)
			if !ok {
				t.Errorf("expected *types.TechnologyEnergyDay, got %T", result)
				return
			}

			// Validate basic structure
			validateBasicStructure(t, data)

			// Validate first hour data
			firstHour := data.Records[0]
			validateFirstHourData(t, &firstHour)

			// Print first few hours for debugging
			for i := 0; i < 3 && i < len(data.Records); i++ {
				record := data.Records[i]
				t.Logf("Hour %d: Nuclear=%.1f, Wind=%.1f, Hydro=%.1f, Coal=%.1f",
					record.Hour, record.Nuclear, record.Wind, record.Hydro, record.Coal)
			}

			// Test renewable energy calculation
			renewable := firstHour.Wind + firstHour.SolarPV + firstHour.SolarThermal + firstHour.Hydro
			expectedRenewable := 7371.1 + 3.7 + 25.7 + 2405.9 // Sum from testdata
			if math.Abs(renewable-expectedRenewable) > 0.1 {
				t.Errorf("renewable energy: expected %.1f, got %.1f", expectedRenewable, renewable)
			}
			t.Logf("Hour 1 renewable energy: %.1f MWh", renewable)
		})
	}
}

func validateBasicStructure(t *testing.T, data *types.TechnologyEnergyDay) {
	expectedDate := time.Date(2020, 11, 13, 0, 0, 0, 0, time.UTC)
	if !data.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, data.Date)
	}

	if data.System != types.Iberian {
		t.Errorf("expected system %v, got %v", types.Iberian, data.System)
	}

	if len(data.Records) != 24 {
		t.Errorf("expected 24 hourly records, got %d", len(data.Records))
	}

	t.Logf("Parsed data for %s (%s system) with %d records",
		data.Date.Format("2006-01-02"), data.System, len(data.Records))
}

func validateFirstHourData(t *testing.T, firstHour *types.TechnologyEnergy) {
	if firstHour.Hour != 1 {
		t.Errorf("expected hour 1, got %d", firstHour.Hour)
	}

	// Test European number format parsing - these values come from the testdata file
	// Line: 13/11/2020;1;1.432,0;;;6.088,9;2.405,9;3.191,6;7.371,1;25,7;3,7;6.292,4;;2.400,0;
	expectedValues := map[string]float64{
		"Coal":          1432.0, // 1.432,0
		"Nuclear":       6088.9, // 6.088,9
		"Hydro":         2405.9, // 2.405,9
		"CombinedCycle": 3191.6, // 3.191,6
		"Wind":          7371.1, // 7.371,1
		"SolarThermal":  25.7,   // 25,7
		"SolarPV":       3.7,    // 3,7
		"Cogeneration":  6292.4, // 6.292,4
		"ImportNoMIBEL": 2400.0, // 2.400,0
	}

	testValue := func(name string, actual, expected float64) {
		if math.IsNaN(actual) && !math.IsNaN(expected) {
			t.Errorf("%s: expected %.1f, got NaN", name, expected)
		} else if !math.IsNaN(expected) && math.Abs(actual-expected) > 0.1 {
			t.Errorf("%s: expected %.1f, got %.1f", name, expected, actual)
		}
	}

	// Test each value
	testValue("Coal", firstHour.Coal, expectedValues["Coal"])
	testValue("Nuclear", firstHour.Nuclear, expectedValues["Nuclear"])
	testValue("Hydro", firstHour.Hydro, expectedValues["Hydro"])
	testValue("CombinedCycle", firstHour.CombinedCycle, expectedValues["CombinedCycle"])
	testValue("Wind", firstHour.Wind, expectedValues["Wind"])
	testValue("SolarThermal", firstHour.SolarThermal, expectedValues["SolarThermal"])
	testValue("SolarPV", firstHour.SolarPV, expectedValues["SolarPV"])
	testValue("Cogeneration", firstHour.Cogeneration, expectedValues["Cogeneration"])
	testValue("ImportNoMIBEL", firstHour.ImportNoMIBEL, expectedValues["ImportNoMIBEL"])

	// Test that empty fields are properly handled as NaN
	validateEmptyFields(t, firstHour)
}

func validateEmptyFields(t *testing.T, firstHour *types.TechnologyEnergy) {
	if !math.IsNaN(firstHour.FuelGas) {
		t.Errorf("FuelGas: expected NaN for empty field, got %.1f", firstHour.FuelGas)
	}
	if !math.IsNaN(firstHour.SelfProducer) {
		t.Errorf("SelfProducer: expected NaN for empty field, got %.1f", firstHour.SelfProducer)
	}
	if !math.IsNaN(firstHour.ImportInt) {
		t.Errorf("ImportInt: expected NaN for empty field, got %.1f", firstHour.ImportInt)
	}
}

func TestEnergyByTechnologyParser_EuropeanNumberFormat(t *testing.T) {
	// Test the European number format parsing that was the root cause of the bug
	testCases := []struct {
		input    string
		expected float64
	}{
		{"1.432,0", 1432.0},   // thousands.decimal
		{"6.088,9", 6088.9},   // thousands.decimal
		{"2.405,9", 2405.9},   // thousands.decimal
		{"25,7", 25.7},        // decimal only
		{"3,7", 3.7},          // decimal only
		{"2.400,0", 2400.0},   // thousands.decimal
		{"15.934,8", 15934.8}, // thousands.decimal (from recent data)
		{"292,0", 292.0},      // decimal only
		{"50,8", 50.8},        // decimal only
		{"", math.NaN()},      // empty should be NaN
	}

	for _, tc := range testCases {
		t.Run("parse_"+tc.input, func(t *testing.T) {
			result, err := ParseFloat(tc.input)

			if tc.input == "" {
				// Empty string should return NaN
				if err != nil {
					t.Errorf("unexpected error for empty string: %v", err)
				}
				if !math.IsNaN(result) {
					t.Errorf("expected NaN for empty string, got %f", result)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error parsing '%s': %v", tc.input, err)
				return
			}

			if math.IsNaN(tc.expected) {
				if !math.IsNaN(result) {
					t.Errorf("expected NaN, got %f", result)
				}
			} else {
				if math.Abs(result-tc.expected) > 0.01 {
					t.Errorf("expected %f, got %f", tc.expected, result)
				}
			}
		})
	}
}

func TestEnergyByTechnologyParser_ColumnMapping(t *testing.T) {
	parser := NewEnergyByTechnologyParser()

	// Test header line parsing
	headerLine := "Fecha;Hora;CARBÓN;FUEL-GAS;AUTOPRODUCTOR;NUCLEAR;HIDRÁULICA;CICLO COMBINADO;EÓLICA;SOLAR TÉRMICA;SOLAR FOTOVOLTAICA;COGENERACIÓN/RESIDUOS/MINI HIDRA;IMPORTACIÓN INTER.;IMPORTACIÓN INTER. SIN MIBEL;"
	fields := SplitCSV(headerLine)

	mapping, _ := parser.parseColumnHeaders([]string{"", "", headerLine})

	expectedMappings := map[int]types.TechnologyType{
		2:  types.Coal,
		3:  types.FuelGas,
		4:  types.SelfProducer,
		5:  types.Nuclear,
		6:  types.Hydro,
		7:  types.CombinedCycle,
		8:  types.Wind,
		9:  types.ThermalSolar,
		10: types.PhotovoltaicSolar,
		11: types.Residuals,
		12: types.Import,
		13: types.ImportWithoutMIBEL,
	}

	if len(mapping) != len(expectedMappings) {
		t.Errorf("expected %d mappings, got %d", len(expectedMappings), len(mapping))
	}

	for col, expectedTech := range expectedMappings {
		if actualTech, exists := mapping[col]; !exists {
			t.Errorf("missing mapping for column %d (%s)", col, fields[col])
		} else if actualTech != expectedTech {
			t.Errorf("column %d: expected %s, got %s", col, expectedTech, actualTech)
		}
	}
}
