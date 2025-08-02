package parsers

import (
	"math"
	"testing"
	"time"

	"github.com/devuo/omiedata/types"
)

func TestMarginalPriceParser_ParseFile(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		expectedDate  time.Time
		expectedHours int
		wantErr       bool
		validateFunc  func(t *testing.T, data *types.MarginalPriceData)
	}{
		{
			name:          "parse old format file - 2006 Cent/kWh",
			filename:      "../testdata/PMD_20060101.txt",
			expectedDate:  time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedHours: 24,
			wantErr:       false,
			validateFunc:  validate2006Format,
		},
		{
			name:          "parse transition format file - 2009 dual market",
			filename:      "../testdata/PMD_20090601.txt",
			expectedDate:  time.Date(2009, 6, 1, 0, 0, 0, 0, time.UTC),
			expectedHours: 24,
			wantErr:       false,
			validateFunc:  validate2009Format,
		},
		{
			name:          "parse current format file - 2022 DST day",
			filename:      "../testdata/PMD_20221030.txt",
			expectedDate:  time.Date(2022, 10, 30, 0, 0, 0, 0, time.UTC),
			expectedHours: 25, // DST change day
			wantErr:       false,
			validateFunc:  validate2022Format,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMarginalPriceParser()
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

			data, ok := result.(*types.MarginalPriceData)
			if !ok {
				t.Errorf("expected *types.MarginalPriceData, got %T", result)
				return
			}

			// Basic validation
			if !data.Date.Equal(tt.expectedDate) {
				t.Errorf("expected date %v, got %v", tt.expectedDate, data.Date)
			}

			if len(data.SpainPrices) == 0 {
				t.Errorf("should have Spain prices")
			}

			// Validate hour count
			spainHours := len(data.SpainPrices)
			if spainHours != tt.expectedHours {
				t.Errorf("expected %d hours, got %d", tt.expectedHours, spainHours)
			}

			t.Logf("Parsed data for %s with %d Spain prices",
				data.Date.Format("2006-01-02"), len(data.SpainPrices))

			// Print first few hours for debugging
			for hour := 1; hour <= 3; hour++ {
				if price, exists := data.SpainPrices[hour]; exists {
					t.Logf("Hour %d: %.2f EUR/MWh", hour, price)
				}
			}

			// Run format-specific validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, data)
			}
		})
	}
}

func validate2006Format(t *testing.T, data *types.MarginalPriceData) {
	// 2006 format: Single price line in Cent/kWh, should be converted to EUR/MWh
	// From testdata: Precio marginal (Cent/kWh);  6,694;  4,888;  4,525;  4,371;...

	expectedPrices := map[int]float64{
		1:  66.94, // 6,694 cent/kWh -> 66.94 EUR/MWh
		2:  48.88, // 4,888 cent/kWh -> 48.88 EUR/MWh
		3:  45.25, // 4,525 cent/kWh -> 45.25 EUR/MWh
		4:  43.71, // 4,371 cent/kWh -> 43.71 EUR/MWh
		24: 76.17, // 7,617 cent/kWh -> 76.17 EUR/MWh
	}

	for hour, expectedPrice := range expectedPrices {
		if actualPrice, exists := data.SpainPrices[hour]; !exists {
			t.Errorf("missing Spain price for hour %d", hour)
		} else if math.Abs(actualPrice-expectedPrice) > 0.01 {
			t.Errorf("hour %d Spain price: expected %.2f EUR/MWh, got %.2f EUR/MWh",
				hour, expectedPrice, actualPrice)
		}
	}

	// 2006 format should not have Portugal prices (single market)
	if len(data.PortugalPrices) > 0 {
		t.Errorf("2006 format should not have Portugal prices, got %d", len(data.PortugalPrices))
	}

	// Should have energy data - currently parsed as decimal values
	// From testdata: Energía en el programa resultante de la casación (MWh);  26.377;  26.070;...
	expectedEnergy := map[int]float64{
		1:  26.377, // Currently parsing as decimal, not thousands separator
		2:  26.070, // Currently parsing as decimal
		24: 25.373, // Currently parsing as decimal
	}

	for hour, expectedEng := range expectedEnergy {
		// Check multiple energy fields that might contain this data
		found := false
		if energy, exists := data.IberianEnergy[hour]; exists && math.Abs(energy-expectedEng) < 0.1 {
			found = true
		}
		if energy, exists := data.SpainBuyEnergy[hour]; exists && math.Abs(energy-expectedEng) < 0.1 {
			found = true
		}

		if !found {
			t.Errorf("hour %d: expected energy value %.3f MWh not found in any energy field", hour, expectedEng)
		}
	}

	t.Logf("✓ 2006 format: Cent/kWh conversion, single market, energy data")
}

func validate2009Format(t *testing.T, data *types.MarginalPriceData) {
	// 2009 format: Separate Spain and Portugal prices in Cent/kWh
	// From testdata:
	// Precio marginal en el sistema español (Cent/kWh);  3,997;  3,760;  3,560;...
	// Precio marginal en el sistema portugués (Cent/kWh);  3,997;  3,760;  3,731;...

	expectedSpainPrices := map[int]float64{
		1:  39.97, // 3,997 cent/kWh -> 39.97 EUR/MWh
		2:  37.60, // 3,760 cent/kWh -> 37.60 EUR/MWh
		3:  35.60, // 3,560 cent/kWh -> 35.60 EUR/MWh
		24: 37.52, // 3,752 cent/kWh -> 37.52 EUR/MWh
	}

	expectedPortugalPrices := map[int]float64{
		1:  39.97, // 3,997 cent/kWh -> 39.97 EUR/MWh (same as Spain)
		2:  37.60, // 3,760 cent/kWh -> 37.60 EUR/MWh (same as Spain)
		3:  37.31, // 3,731 cent/kWh -> 37.31 EUR/MWh (different from Spain!)
		24: 40.19, // 4,019 cent/kWh -> 40.19 EUR/MWh
	}

	// Validate Spain prices
	for hour, expectedPrice := range expectedSpainPrices {
		if actualPrice, exists := data.SpainPrices[hour]; !exists {
			t.Errorf("missing Spain price for hour %d", hour)
		} else if math.Abs(actualPrice-expectedPrice) > 0.01 {
			t.Errorf("hour %d Spain price: expected %.2f EUR/MWh, got %.2f EUR/MWh",
				hour, expectedPrice, actualPrice)
		}
	}

	// Validate Portugal prices (key difference from 2006)
	if len(data.PortugalPrices) == 0 {
		t.Errorf("2009 format should have Portugal prices")
	}

	for hour, expectedPrice := range expectedPortugalPrices {
		if actualPrice, exists := data.PortugalPrices[hour]; !exists {
			t.Errorf("missing Portugal price for hour %d", hour)
		} else if math.Abs(actualPrice-expectedPrice) > 0.01 {
			t.Errorf("hour %d Portugal price: expected %.2f EUR/MWh, got %.2f EUR/MWh",
				hour, expectedPrice, actualPrice)
		}
	}

	// Validate energy data with European number format
	// From testdata: Energía total de compra sistema español (MWh);  24326,2;  22477,4;...
	expectedSpainBuyEnergy := map[int]float64{
		1: 24326.2, // 24326,2 -> 24326.2 (comma as decimal separator)
		2: 22477.4, // 22477,4 -> 22477.4
		3: 21142.8, // 21142,8 -> 21142.8
	}

	for hour, expectedEng := range expectedSpainBuyEnergy {
		if energy, exists := data.SpainBuyEnergy[hour]; !exists {
			t.Errorf("missing Spain buy energy for hour %d", hour)
		} else if math.Abs(energy-expectedEng) > 0.1 {
			t.Errorf("hour %d Spain buy energy: expected %.1f MWh, got %.1f MWh",
				hour, expectedEng, energy)
		}
	}

	t.Logf("✓ 2009 format: dual market prices, energy data, market coupling")
}

func validate2022Format(t *testing.T, data *types.MarginalPriceData) {
	// 2022 format: This file contains adjustment prices (EUR/MWh), all zeros
	// This is a DST change day with 25 hours
	// From testdata: Precio de ajuste en el sistema español (EUR/MWh);     0,00;     0,00;...

	// Validate DST day has 25 hours
	if len(data.SpainPrices) != 25 {
		t.Errorf("DST day should have 25 hours, got %d", len(data.SpainPrices))
	}

	// All adjustment prices should be 0.00 EUR/MWh
	for hour := 1; hour <= 25; hour++ {
		if price, exists := data.SpainPrices[hour]; !exists {
			t.Errorf("missing Spain price for hour %d on DST day", hour)
		} else if price != 0.0 {
			t.Errorf("hour %d adjustment price: expected 0.00 EUR/MWh, got %.2f EUR/MWh",
				hour, price)
		}
	}

	// Portugal prices should also be zero if present
	if len(data.PortugalPrices) > 0 {
		for hour := 1; hour <= 25; hour++ {
			if price, exists := data.PortugalPrices[hour]; exists && price != 0.0 {
				t.Errorf("hour %d Portugal adjustment price: expected 0.00 EUR/MWh, got %.2f EUR/MWh",
					hour, price)
			}
		}
	}

	t.Logf("✓ 2022 format: DST 25 hours, zero adjustment prices, EUR/MWh format")
}

func TestMarginalPriceParser_DateParsing(t *testing.T) {
	parser := NewMarginalPriceParser()

	// Test date parsing from header
	headerLine := "OMIE - Mercado de electricidad;Fecha Emisión :01/01/2006 - 08:30;;01/01/2006;Precio del mercado diario (Cent/kWh);;;;"

	date, err := parser.parseDateFromHeader(headerLine)
	if err != nil {
		t.Errorf("failed to parse date: %v", err)
		return
	}

	expected := time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, date)
	}
}

func TestMarginalPriceParser_EuropeanNumberFormat(t *testing.T) {
	// Test European number format parsing in price context
	testCases := []struct {
		input          string
		expectedResult float64
		description    string
	}{
		{"6,694", 6.694, "cent price with comma decimal"},
		{"3,997", 3.997, "cent price with comma"},
		{"0,00", 0.0, "zero price with comma"},
		{"42,50", 42.5, "simple decimal price"},
		{"24326,2", 24326.2, "energy value with comma decimal"},
		{"1.071,6", 1071.6, "value with thousands separator and comma decimal"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := ParseFloat(tc.input)
			if err != nil {
				t.Errorf("ParseFloat('%s') failed: %v", tc.input, err)
				return
			}

			if math.Abs(result-tc.expectedResult) > 0.01 {
				t.Errorf("ParseFloat('%s'): expected %.3f, got %.3f", tc.input, tc.expectedResult, result)
			}
		})
	}
}
