package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)
//search :
type searchResultsMsg []string
func searchYoutube(query string) tea.Cmd {
	return func() tea.Msg {
		searchTerm := fmt.Sprintf("ytsearch5:%s", query)
		out, err := exec.Command(
		"yt-dlp",
		searchTerm,
		"--print", "%(title)s - %(webpage_url)s",
		"--flat-playlist",
		"--skip-download",
		).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		return searchResultsMsg(lines)
	}
}
// Model: holds your app's state
type model struct {
	count     int
	textInput textinput.Model
	editing   bool   // true when the search box is focused and taking input
	output    string // stores result/error from pause()
	selecting bool     // true when showing the results list
	results   []string // parsed search results
	cursor    int 
}

// initialModel just builds and returns the starting state.
// It does NOT start the program — that's main()'s job.
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	return model{textInput: ti}
}

// Init: runs once when the program starts (no initial command needed here)
func (m model) Init() tea.Cmd {
	return nil
}

// Message types to carry pause()'s result back into Update
type cmdOutputMsg string
type cmdErrMsg struct{ err error }

func pause() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("bash", "-c", `echo '{ "command": ["cycle", "pause"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
func playSelected(url string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("mpv", "--no-video", "--input-ipc-server=/tmp/mpvsocket", url).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
// Update: handles events (like keypresses) and returns an updated model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// If we're in "editing" mode, most keys should go to the text input,
		// except a couple of special ones (esc to leave editing, enter to submit).
		if m.editing {
			switch msg.String() {
			case "esc":
				m.editing = false
				m.textInput.Blur()
				return m, nil
			case "enter":
				m.editing = false
				m.textInput.Blur()
				query := m.textInput.Value()
				return m, searchYoutube(query)
			}
			// forward everything else (letters, backspace, etc.) to the text input
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
if m.selecting {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case "enter":
		m.selecting = false
		selected := m.results[m.cursor]
		// extract the URL — it's after " - " in your format string
		parts := strings.Split(selected, " - ")
		url := parts[len(parts)-1]
		return m, playSelected(url)
	case "esc":
		m.selecting = false
	}
	return m, nil
}
		// Not editing: treat keys as normal navigation/commands
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			return m, pause()
		case "down", "j":
			m.count--
		case "s":
			m.editing = true
			m.textInput.Focus()
			return m, textinput.Blink
		}
	case searchResultsMsg:
	m.results = []string(msg)
	m.cursor = 0
	m.selecting = true

	case cmdOutputMsg:
		m.output = string(msg)

	case cmdErrMsg:
		m.output = fmt.Sprintf("error: %v", msg.err)
	}
	return m, nil
}

// View: renders the model as a string
func (m model) View() string {
	if m.editing {
		return fmt.Sprintf("\n  Search: %s\n\n  (enter to submit, esc to cancel)\n", m.textInput.View())
	}
	if m.selecting {
	s := "\n  Select a result:\n\n"
	for i, r := range m.results {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s += fmt.Sprintf("  %s %s\n", cursor, r)
	}
	s += "\n  ↑/k up • ↓/j down • enter to play • esc to cancel\n"
	return s
	}

	return fmt.Sprintf(
		"\n  Count: %d\n\n  Output: %s\n\n  ↑/k Play/Pause • ↓/j decrement • s to search • q to quit\n",
		m.count, m.output,
	)
}
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}