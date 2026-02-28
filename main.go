package main

import (
	"fmt"
	"github.com/wallacegibbon/sysfs-monitor-tui/internal/monitor"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	mon monitor.Monitor
}

func initialModel() model {
	return model{
		mon: monitor.NewMonitor(),
	}
}

func (m model) Init() tea.Cmd {
	return m.mon.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	updatedMonitor, cmd := m.mon.Update(msg)
	m.mon = updatedMonitor
	return m, cmd
}

func (m model) View() string {
	return m.mon.View()
}
