package parsers

import (
	"bufio"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"github.com/devuo/omiedata/types"
)

// ParseFloat parses a European-formatted float (comma as decimal separator)
func ParseFloat(s string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return math.NaN(), nil
	}
	
	// Replace comma with dot for European format
	s = strings.Replace(strings.TrimSpace(s), ",", ".", -1)
	return strconv.ParseFloat(s, 64)
}

// ParseDate parses a date in DD/MM/YYYY format
func ParseDate(s string) (time.Time, error) {
	return time.Parse("02/01/2006", strings.TrimSpace(s))
}

// NewISO88591Reader creates a reader that decodes from ISO-8859-1 to UTF-8
func NewISO88591Reader(r io.Reader) io.Reader {
	decoder := charmap.ISO8859_1.NewDecoder()
	return transform.NewReader(r, decoder)
}

// ReadLines reads all lines from a reader and returns them as a slice
func ReadLines(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return nil, types.NewOMIEError(types.ErrCodeParse, "failed to read lines", err)
	}
	
	return lines, nil
}

// SplitCSV splits a CSV line by semicolon separator
func SplitCSV(line string) []string {
	return strings.Split(line, ";")
}

// FindDatesInString finds dates in DD/MM/YYYY format in a string
func FindDatesInString(s string) []string {
	// Simple regex-like approach for DD/MM/YYYY pattern
	var dates []string
	words := strings.Fields(s)
	
	for _, word := range words {
		// Check if word matches DD/MM/YYYY pattern
		if len(word) == 10 && word[2] == '/' && word[5] == '/' {
			if _, err := ParseDate(word); err == nil {
				dates = append(dates, word)
			}
		}
	}
	
	return dates
}

// ParseHour parses hour value, handling 1-24 format
func ParseHour(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, types.NewOMIEError(types.ErrCodeParse, "empty hour value", nil)
	}
	
	hour, err := strconv.Atoi(s)
	if err != nil {
		return 0, types.NewOMIEError(types.ErrCodeParse, "invalid hour format", err)
	}
	
	if hour < 1 || hour > 25 { // Allow 25 for DST changes
		return 0, types.NewOMIEError(types.ErrCodeParse, "hour out of range (1-25)", nil)
	}
	
	return hour, nil
}

// IsValidPriceValue checks if a price value is valid (not NaN or negative for prices)
func IsValidPriceValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// IsValidEnergyValue checks if an energy value is valid
func IsValidEnergyValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}