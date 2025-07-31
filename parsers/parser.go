package parsers

import (
	"io"
	"net/http"
)

// Parser defines the interface for parsing OMIE data files
type Parser interface {
	// ParseResponse parses data from an HTTP response
	ParseResponse(resp *http.Response) (interface{}, error)
	
	// ParseFile parses data from a file
	ParseFile(filename string) (interface{}, error)
	
	// ParseReader parses data from any io.Reader
	ParseReader(reader io.Reader) (interface{}, error)
}