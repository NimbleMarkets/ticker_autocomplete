package ticker_autocomplete

import (
	"sync/atomic"
	"time"
)

type CompletionSource struct {
	RefreshFrequency time.Duration
	RetryFrequency   time.Duration
	LastError        error // this seems a little hacky but its the only way to communicate the error to the web request
	completer        *atomic.Value
	factory          func() (Completer, error)
}

func NewCompletionSource(factory func() (Completer, error)) *CompletionSource {
	cs := newCompletionSource(factory, 8*time.Hour, 1*time.Minute)
	go cs.goRefreshCompleter()
	return cs
}

func newCompletionSource(factory func() (Completer, error), refresh time.Duration, retry time.Duration) *CompletionSource {
	cs := &CompletionSource{
		RefreshFrequency: refresh,
		RetryFrequency:   retry,
		completer:        &atomic.Value{},
		factory:          factory,
	}

	go cs.goRefreshCompleter()
	return cs
}

// GetCompleter returns the current completer. It is safe to call this concurrently.
// It is possible that nil will be returned if the completer has never successfully loaded.
func (ch *CompletionSource) GetCompleter() Completer {
	c, ok := ch.completer.Load().(Completer)

	if !ok {
		return nil
	}
	return c
}

// goRefreshCompleter refreshes the completer at the RefreshFrequency.
// If the completer fails to refresh, it will retry at the RetryFrequency.
// If the completer fails to refresh, the LastError will be set to the error.
func (ch *CompletionSource) goRefreshCompleter() {
	for {
		// Refresh the completer and store it retrying until success
		success := false
		for !success {
			err := ch.Refresh()
			if err != nil {
				time.Sleep(ch.RetryFrequency)
				continue
			}
			success = true
		}

		time.Sleep(ch.RefreshFrequency)
	}
}

// Refresh refreshes the completer and returns it. Exposed for testing.
func (ch *CompletionSource) Refresh() error {
	completer, err := ch.factory()
	if err != nil {
		ch.LastError = err
		return err
	}
	ch.LastError = nil
	ch.completer.Store(completer)
	return nil
}
