package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
)

var (
	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.SetHeight(20)
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CursorLineStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("177")).
		Foreground(lipgloss.Color("160"))
	blurTextarea(&t)
	return t
}

func focusTextarea(m *textarea.Model) {
	m.CursorLineStyle = cursorLineStyle
	m.PlaceholderStyle = focusedPlaceholderStyle
	m.Focus()
}

func blurTextarea(m *textarea.Model) {
	m.CursorLineStyle = lipgloss.NewStyle()
	m.PlaceholderStyle = placeholderStyle
	m.Blur()
}

type model struct {
	width  int
	keymap keymap
	help   help.Model
	inputs []textarea.Model
	focus  int
}

func newModel() model {
	m := model{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = newTextarea()
	}
	focusTextarea(&m.inputs[m.focus])
	m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
			m.inputs[m.focus].Blur()
			blurTextarea(&m.inputs[m.focus])
			m.focus++
			if m.focus > len(m.inputs)-1 {
				m.focus = 0
			}
		case key.Matches(msg, m.keymap.prev):
			blurTextarea(&m.inputs[m.focus])
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) - 1
			}
			m.inputs[m.focus].Focus()
		case key.Matches(msg, m.keymap.add):
			m.inputs = append(m.inputs, newTextarea())
		case key.Matches(msg, m.keymap.remove):
			m.inputs = m.inputs[:len(m.inputs)-1]
			if m.focus > len(m.inputs)-1 {
				m.focus = len(m.inputs) - 1
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	focusTextarea(&m.inputs[m.focus])
	m.updateKeybindings()
	m.sizeInputs()

	// Workaround to help unmap ctrl+n/ctrl+p from textarea
	var keystroke string
	if msg, ok := msg.(tea.KeyMsg); ok {
		keystroke = msg.String()
	}

	var cmds []tea.Cmd

	if keystroke != "ctrl+n" && keystroke != "ctrl+p" {
		for i := range m.inputs {
			newModel, cmd := m.inputs[i].Update(msg)
			m.inputs[i] = newModel
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetWidth(m.width / len(m.inputs))
	}
}

func (m *model) updateKeybindings() {
	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
}

func (m model) View() string {
	if m.width == 0 {
		return "Hang on..."
	}

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}

func main() {
	if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
