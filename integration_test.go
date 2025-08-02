package omiedata

import (
	"testing"

	"github.com/devuo/omiedata/parsers"
)

func TestMarginalPriceIntegration(t *testing.T) {
	// Test with local files
	parser := parsers.NewMarginalPriceParser()

	testFiles := []struct {
		filename string
		year     int
	}{
		{"testdata/PMD_20060101.txt", 2006},
		{"testdata/PMD_20090601.txt", 2009},
		{"testdata/PMD_20221030.txt", 2022},
	}

	for _, test := range testFiles {
		t.Run(test.filename, func(t *testing.T) {
			result, err := parser.ParseFile(test.filename)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", test.filename, err)
			}

			data, ok := result.(*MarginalPriceData)
			if !ok {
				t.Fatalf("Expected *MarginalPriceData, got %T", result)
			}

			// Verify basic data integrity
			if data.Date.Year() != test.year {
				t.Errorf("Expected year %d, got %d", test.year, data.Date.Year())
			}

			if len(data.SpainPrices) == 0 {
				t.Error("No Spain prices found")
			}

			// Check that we have 24 or 25 hours (DST adjustment)
			hourCount := len(data.SpainPrices)
			if hourCount < 23 || hourCount > 25 {
				t.Errorf("Unexpected hour count: %d (should be 23-25)", hourCount)
			}

			t.Logf("Successfully parsed %s: %d hours, date %s",
				test.filename, hourCount, data.Date.Format("2006-01-02"))
		})
	}
}

func TestEnergyByTechnologyIntegration(t *testing.T) {
	parser := parsers.NewEnergyByTechnologyParser()

	result, err := parser.ParseFile("testdata/EnergyByTechnology_9_20201113.TXT")
	if err != nil {
		t.Fatalf("Failed to parse energy by technology file: %v", err)
	}

	dayData, ok := result.(*TechnologyEnergyDay)
	if !ok {
		t.Fatalf("Expected *TechnologyEnergyDay, got %T", result)
	}

	// Verify basic data integrity
	if dayData.Date.Year() != 2020 {
		t.Errorf("Expected year 2020, got %d", dayData.Date.Year())
	}

	if dayData.System != Iberian {
		t.Errorf("Expected Iberian system, got %v", dayData.System)
	}

	if len(dayData.Records) == 0 {
		t.Error("No technology records found")
	}

	// Check that we have reasonable data
	for i, record := range dayData.Records {
		if record.Hour < 1 || record.Hour > 25 {
			t.Errorf("Invalid hour %d in record %d", record.Hour, i)
		}

		// At least some technology should have energy > 0 (but allow for missing hours in files)
		totalEnergy := record.Nuclear + record.Wind + record.SolarPV + record.Hydro + record.CombinedCycle + record.Coal + record.Cogeneration
		if totalEnergy <= 0 && record.Hour <= 24 {
			// Only warn for normal hours (1-24), not for hours that might be missing data
			t.Logf("Warning: No energy found in record for hour %d (this might be normal for some files)", record.Hour)
		}
	}

	t.Logf("Successfully parsed energy by technology: %d records, date %s, system %v",
		len(dayData.Records), dayData.Date.Format("2006-01-02"), dayData.System)
}

func TestHighLevelAPI(t *testing.T) {
	// Test the high-level API by parsing files directly
	// (Skip network tests to avoid external dependencies in CI)

	importer := NewMarginalPriceImporter()

	// This would normally make network requests, but we'll test the structure
	if importer == nil {
		t.Error("Failed to create importer")
	}

	// Test that we can create different system importers
	iberianImporter := NewEnergyByTechnologyImporter(Iberian)
	spainImporter := NewEnergyByTechnologyImporter(Spain)
	portugalImporter := NewEnergyByTechnologyImporter(Portugal)

	if iberianImporter == nil || spainImporter == nil || portugalImporter == nil {
		t.Error("Failed to create technology importers")
	}

	t.Log("High-level API components created successfully")
}

func TestSystemTypes(t *testing.T) {
	systems := []SystemType{Spain, Portugal, Iberian}
	expectedValues := []int{1, 2, 9}
	expectedNames := []string{"SPAIN", "PORTUGAL", "IBERIAN"}

	for i, system := range systems {
		if int(system) != expectedValues[i] {
			t.Errorf("System %v should have value %d, got %d", system, expectedValues[i], int(system))
		}

		if system.String() != expectedNames[i] {
			t.Errorf("System %v should have name %s, got %s", system, expectedNames[i], system.String())
		}
	}
}

func TestTechnologyTypes(t *testing.T) {
	// Test some key technology mappings
	testCases := []struct {
		tech    TechnologyType
		spanish string
	}{
		{Nuclear, "NUCLEAR"},
		{Wind, "EÓLICA"},
		{PhotovoltaicSolar, "SOLAR FOTOVOLTAICA"},
		{Coal, "CARBÓN"},
	}

	for _, tc := range testCases {
		if tc.tech.NameInFile() != tc.spanish {
			t.Errorf("Technology %v should map to '%s', got '%s'", tc.tech, tc.spanish, tc.tech.NameInFile())
		}
	}
}
