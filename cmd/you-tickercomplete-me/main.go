// Copyright (c) 2024 Neomantra BV
//
// Simple CLI and TUI test program for ticker_autocomplete
//
// Adapted from:
//   https://github.com/charmbracelet/bubbletea/blob/master/examples/autocomplete/main.go

package main

import (
	"fmt"
	"os"

	tac "github.com/NimbleMarkets/ticker_autocomplete"
	nqtac "github.com/NimbleMarkets/ticker_autocomplete/sources/nasdaq"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

///////////////////////////////////////////////////////////////////////////

const MAX_COMPLETIONS = 5

var nasdaqCompleter tac.Completer

type gotTickerCompletionMsg []tac.Completion

func getTickerCompletionCmd(prompt string) tea.Cmd {
	return func() tea.Msg {
		// get the completion, but also force the type to be gotTickerCompletionMsg
		var tc gotTickerCompletionMsg = nasdaqCompleter.GetCompletions(prompt, MAX_COMPLETIONS)
		return tc
	}
}

///////////////////////////////////////////////////////////////////////////

func main() {
	// Set up the completer
	var err error
	nasdaqCompleter, err = nqtac.NewCompleter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create Nasdaq Completer: %s\n", err.Error())
		os.Exit(1)
	}

	// If there is a command-line argument, then we just output prompts for that
	if len(os.Args) > 1 {
		for _, prompt := range os.Args[1:] {
			completions := nasdaqCompleter.GetCompletions(prompt, MAX_COMPLETIONS)
			for _, c := range completions {
				fmt.Printf("%s:%s   %s\n", c.Market, c.Ticker, c.Name)
			}
		}
		os.Exit(0)
	}

	// Run a simple BubbleTea TUI
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////

type model struct {
	textInput textinput.Model
	textArea  textarea.Model
	help      help.Model
	keymap    keymap
}

type keymap struct{}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next")),
		key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "prev")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "Enter a ticker: "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()
	ti.CharLimit = 8
	ti.Width = 80
	ti.ShowSuggestions = true

	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.SetWidth(80)

	h := help.New()
	km := keymap{}
	return model{textInput: ti, textArea: ta, help: h, keymap: km}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
		var tiUpdateCmd tea.Cmd
		m.textInput, tiUpdateCmd = m.textInput.Update(msg)
		return m, tea.Batch(tiUpdateCmd, getTickerCompletionCmd(m.textInput.Value()))

	case gotTickerCompletionMsg:
		var suggestions []string
		var completionsText string
		for i, r := range msg {
			if i > MAX_COMPLETIONS {
				break
			}
			suggestions = append(suggestions, r.Ticker)
			completionsText += fmt.Sprintf("%s:%s   %s\n", r.Market, r.Ticker, r.Name)
		}
		m.textInput.SetSuggestions(suggestions)
		m.textArea.SetValue(completionsText)

		var tiUpdateCmd, taUpdateCmd tea.Cmd
		m.textInput, tiUpdateCmd = m.textInput.Update(msg)
		m.textArea, taUpdateCmd = m.textArea.Update(completionsText)
		return m, tea.Batch(tiUpdateCmd, taUpdateCmd)
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf(
		"  %s\n\n%s\n\n%s\n\n",
		m.textInput.View(),
		m.textArea.View(),
		m.help.View(m.keymap),
	)
}
