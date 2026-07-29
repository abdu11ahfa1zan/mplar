package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Model: holds your app's state
type model struct {
	count int
}

// Init: runs once when the program starts (no initial command needed here)
func (m model) Init() tea.Cmd {
	return nil
}

// Update: handles events (like keypresses) and returns an updated model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.count++
		case "down", "j":
			m.count--
		}
	}
	return m, nil
}

// View: renders the model as a string
func (m model) View() string {
	return fmt.Sprintf("\n  Count: %d\n\n  ↑/k increment • ↓/j decrement • q to quit\n", m.count)
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}