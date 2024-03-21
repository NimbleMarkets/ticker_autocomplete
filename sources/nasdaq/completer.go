// sources/nasdaq/completer.go
// Copyright (c) 2024 Neomantra BV

package nasdaq

import (
	"math"
	"strings"

	tac "github.com/NimbleMarkets/ticker_autocomplete"
	trie "github.com/Vivino/go-autocomplete-trie"
)

// NasdaqCompleter implements ticker_autocomplete.TickerCompleter for NASDAQ-source symbol list
type NasdaqCompleter struct {
	completions []tac.Completion
	indexMap    map[string]int // map from Ticker to index in completions array
	index       *trie.Trie
}

// NewCompleter returns a new Nasdaq Completer, loading symbols from cache or the Internet.
func NewCompleter() (*NasdaqCompleter, error) {
	// Load nasdaqtraded.txt file (possibly from cache, downloading and caching otherwise)
	nqts, err := FetchNasdaqTraded()
	if err != nil {
		return nil, err
	}

	// Allocate the NasdaqCompleter
	completer := &NasdaqCompleter{
		completions: make([]tac.Completion, len(nqts)),
		indexMap:    make(map[string]int, len(nqts)),
		index:       trie.New(),
	}

	// populate everything from the []NasdaqTraded
	for i, nqt := range nqts {
		// Push the Completion, based on the NasdaqTraded
		completer.completions[i] = tac.Completion{
			Ticker: nqt.Symbol,
			Name:   nqt.Name,
			Type:   nqt.Type,
			Region: "US",
			Market: nqt.ListingExchange,
		}

		// Populate secondary indices and tries, using uppercase keys
		idxContent := strings.ToUpper(nqt.Symbol + " " + nqt.Name)
		completer.indexMap[idxContent] = i
		completer.index.Insert(idxContent)
	}

	return completer, nil
}

// GetAll returns all completions.
func (c *NasdaqCompleter) GetAll() []tac.Completion {
	return c.completions
}

// GetCompletions returns completions for the given prompt.
func (c *NasdaqCompleter) GetCompletions(prompt string, limit int) []tac.Completion {
	var results []tac.Completion
	if limit < 0 {
		limit = math.MaxInt32
	}

	// Try to complete the prompt using ticker symbols
	tickers := c.index.Search(prompt, limit)
	for _, ticker := range tickers {
		idx := c.indexMap[strings.ToUpper(ticker)]
		results = append(results, c.completions[idx])
	}

	return results
}
