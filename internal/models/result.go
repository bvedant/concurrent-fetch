package models

// Result represents processed data with error handling
type Result struct {
	Data  []byte
	Error error
}
