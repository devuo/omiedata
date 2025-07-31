package parsers

import (
	"testing"
	"time"

	"github.com/devuo/omiedata/types"
)

func TestMarginalPriceParser_ParseFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "parse old format file",
			filename: "../testdata/PMD_20060101.txt",
			wantErr:  false,
		},
		{
			name:     "parse transition format file",
			filename: "../testdata/PMD_20090601.txt",
			wantErr:  false,
		},
		{
			name:     "parse current format file",
			filename: "../testdata/PMD_20221030.txt",
			wantErr:  false,
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
			if data.Date.IsZero() {
				t.Errorf("date should not be zero")
			}

			if len(data.SpainPrices) == 0 {
				t.Errorf("should have Spain prices")
			}

			t.Logf("Parsed data for %s with %d Spain prices", 
				data.Date.Format("2006-01-02"), len(data.SpainPrices))

			// Print first few hours for debugging
			for hour := 1; hour <= 3; hour++ {
				if price, exists := data.SpainPrices[hour]; exists {
					t.Logf("Hour %d: %.2f EUR/MWh", hour, price)
				}
			}
		})
	}
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