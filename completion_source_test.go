package ticker_autocomplete

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCompletionSource(t *testing.T) {
	expected := &MockCompleter{}
	factory := func() (Completer, error) {
		time.Sleep(1 * time.Millisecond)
		return expected, nil
	}

	cs := NewCompletionSource(factory)
	assert.NotNil(t, cs)

	c := cs.GetCompleter()
	assert.Nil(t, c, "GetCompleter should return nil before the first refresh is complete")

	time.Sleep(10 * time.Millisecond)
	c = cs.GetCompleter()
	assert.Same(t, expected, c, "GetCompleter should return the completer after the first refresh is complete")
}

func TestRefresh(t *testing.T) {
	expected := &MockCompleter{}
	count := 0
	factory := func() (Completer, error) {
		count++
		time.Sleep(1 * time.Millisecond)
		return expected, nil
	}

	cs := NewCompletionSource(factory)
	assert.Equal(t, 0, count, "Factory should not be called until the first refresh completes")

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, count, "Factory should have been called once during initialization")

	err := cs.Refresh()
	assert.Nil(t, err, "Refresh should not return an error")
	assert.Equal(t, 2, count, "Refresh invokes the factory")
}

func TestAutoRefresh(t *testing.T) {
	expected := &MockCompleter{}
	count := 0
	factory := func() (Completer, error) {
		count++
		return expected, nil
	}

	cs := newCompletionSource(factory, 1*time.Millisecond, 1*time.Millisecond)

	time.Sleep(10 * time.Millisecond)
	assert.Greater(t, count, 1, "Factory be refreshed many times in 10ms")

	c := cs.GetCompleter()
	assert.Same(t, expected, c)
	assert.Nil(t, cs.LastError)
}

func TestRefresh_Retry(t *testing.T) {
	expected := &MockCompleter{}
	count := 0
	err := errors.New("expected error")
	factory := func() (Completer, error) {
		count++
		return expected, err
	}

	cs := newCompletionSource(factory, 1*time.Minute, 1*time.Nanosecond)

	time.Sleep(100 * time.Millisecond)
	assert.Greater(t, count, 1, "Factory be retried many times")

	c := cs.GetCompleter()
	assert.Same(t, err, cs.LastError)
	assert.Nil(t, c)
}

type MockCompleter struct {
	Results []Completion
}

func (c *MockCompleter) GetCompletions(prompt string, limit int) []Completion {
	var results []Completion
	if limit < 0 {
		limit = math.MaxInt32
	}

	return results[:limit]
}

func (c *MockCompleter) GetAll() []Completion {
	return c.Results
}
