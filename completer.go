// Copyright (c) 2024 Neomantra BV

package ticker_autocomplete

// Completion is the ticker and metadata returned by a Completer
type Completion struct {
	Ticker string `json:"ticker"`           // Ticker symbol of the instrument
	Name   string `json:"name"`             // Name of the instrument's security
	Type   string `json:"type,omitempty"`   // Type of the instrument (e.g. "stock", "etf", "index")
	Region string `json:"region,omitempty"` // Region of the instrument
	Market string `json:"exch,omitempty"`   // Exchange where the instrument is listed
}

// TickerCompleter is the interface for getting ticker completions.
type Completer interface {
	// Returns all completions for the given prompt, up to the given limit (or all if limit < 0).
	GetCompletions(prompt string, limit int) []Completion
	GetAll() []Completion
}
