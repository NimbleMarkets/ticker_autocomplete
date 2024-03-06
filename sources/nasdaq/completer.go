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
	tickerMap   map[string]int // map from Ticker to index in completions array
	nameMap     map[string]int // map from Nicker to index in completions array
	tickerTrie  *trie.Trie
	nameTrie    *trie.Trie
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
		tickerMap:   make(map[string]int, len(nqts)),
		nameMap:     make(map[string]int, len(nqts)),
		tickerTrie:  trie.New(),
		nameTrie:    trie.New(),
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
		upperTicker := strings.ToUpper(nqt.Symbol)
		completer.tickerMap[upperTicker] = i
		completer.tickerTrie.Insert(upperTicker)

		upperName := strings.ToUpper(nqt.Name)
		completer.nameMap[upperName] = i
		completer.nameTrie.Insert(upperName)
	}

	return completer, nil
}

// GetCompletions returns completions for the given prompt.
func (c *NasdaqCompleter) GetCompletions(prompt string, limit int) []tac.Completion {
	var results []tac.Completion
	if limit < 0 {
		limit = math.MaxInt32
	}

	// Try to complete the prompt using ticker symbols
	tickers := c.tickerTrie.Search(prompt, limit)
	for _, ticker := range tickers {
		idx := c.tickerMap[strings.ToUpper(ticker)]
		results = append(results, c.completions[idx])
	}

	// Try to complete the prompt using company names
	names := c.nameTrie.Search(prompt, limit)
	for _, name := range names {
		idx := c.nameMap[strings.ToUpper(name)]
		results = append(results, c.completions[idx])
	}

	// TODO: pick only the top results based on some kind of scoring
	return results
}
