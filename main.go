package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"encoding/json"
	"time"


	"github.com/charmbracelet/bubbles/progress"
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
	textInput   textinput.Model
	editing     bool
	selecting   bool
	volume 		int
	results     []string
	progress    progress.Model
	position    float64 
	duration    float64
	mpvPID 	    int
	cursor      int
	nowPlaying  string 
}
type positionMsg struct {
	position float64
	duration float64
}
func getPosition() tea.Cmd {
	return func() tea.Msg {
		posOut, err := exec.Command("bash", "-c",
			`echo '{ "command": ["get_property", "time-pos"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		durOut, err := exec.Command("bash", "-c",
			`echo '{ "command": ["get_property", "duration"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		var posResp, durResp struct {
			Data float64 `json:"data"`
		}
		json.Unmarshal(posOut, &posResp)
		json.Unmarshal(durOut, &durResp)

		return positionMsg{position: posResp.Data, duration: durResp.Data}
	}
}
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)	
	})
}

type tickMsg time.Time
// initialModel just builds and returns the starting state.
// It does NOT start the program — that's main()'s job.
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	return model{
		textInput: ti,
		progress:  progress.New(progress.WithDefaultGradient()),
	}
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
func h10() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("bash", "-c", `echo '{ "command": ["seek", "-5"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
func ha10() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("bash", "-c", `echo '{ "command": ["seek", "5"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
func va10() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("bash", "-c", `echo '{ "command": ["add", "volume", "-5:"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
func v10() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("bash", "-c", `echo '{ "command": ["add", "volume", "+5"] }' | socat - /tmp/mpvsocket`).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
type mpvStartedMsg int

func playSelected(url string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("mpv", "--no-video", "--input-ipc-server=/tmp/mpvsocket", url)
		if err := cmd.Start(); err != nil {
			return cmdErrMsg{err}
		}
		
		return mpvStartedMsg(cmd.Process.Pid)
	}
}
func quit(pid int) tea.Cmd {
	return func() tea.Msg {
		if pid == 0 {
			return cmdOutputMsg("") // nothing was ever started
		}
		out, err := exec.Command("kill", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return cmdErrMsg{err}
		}
		return cmdOutputMsg(string(out))
	}
}
// ###########################################   Update: handles events (like keypresses) and returns an updated model
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
		parts := strings.Split(selected, " - ")
		url := parts[len(parts)-1]
		title := strings.Join(parts[:len(parts)-1], " - ")
		m.nowPlaying = title
		return m, playSelected(url)
	case "esc":
		m.selecting = false
	}
	return m, nil
}
		// Not editing: treat keys as normal navigation/commands
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Sequence(quit(m.mpvPID), tea.Quit)
		case "space"," ", "p":
			return m, pause()
		case "s":
			m.editing = true
			m.textInput.Focus()
			return m, textinput.Blink
		case "down", "j":
    // Only lower the volume if it's strictly greater than 0
    if m.volume > 0 {
        m.volume -= 5
        // Optional safety clamp if m.volume was e.g. 3
        if m.volume < 0 {
            m.volume = 0
        }
        return m, va10()
    }
    return m, nil // Volume is already 0, do nothing

case "up", "k":
    // Prevent the volume from exceeding 100%
    if m.volume < 130 {
        m.volume += 5
	        if m.volume > 130 {
            m.volume = 130
        }
        return m, v10()
    }
    return m, nil // Volume is already 100, do nothing
			return m, v10()
		case "right", "l":
			return m, ha10()
		case "left", "h":
			return m, h10()
		}
	case searchResultsMsg:
	m.results = []string(msg)
	m.cursor = 0
	m.selecting = true

	case mpvStartedMsg:
	m.mpvPID = int(msg)
	m.volume = 100
	return m, tickCmd()

	case cmdErrMsg:
		m.nowPlaying = fmt.Sprintf("error: %v", msg.err)


	case tickMsg: 
		return m, tea.Batch(getPosition(), tickCmd())

	case positionMsg: 
		m.position = msg.position
		m.duration = msg.duration
		if m.duration > 0 {
			cmd := m.progress.SetPercent(m.position / m.duration)
			return m, cmd
		}
	case progress.FrameMsg: 
		newModel, cmd := m.progress.Update(msg)
		if newM, ok := newModel.(progress.Model); ok {
			m.progress = newM
		}
		return m, cmd
	}
	return m, nil
}

// #################################################### View: renders the model as a string
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
	"\n  Now Playing: %s\n\n%s\nVolume: %d\n\nspace to pause - Up/Down to Increase/Decrease volume - Right/Left to skip/go back 5s \ns to search - q to quit\n\n",
	m.nowPlaying,
	m.progress.View(),
	m.volume,
)
}
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}