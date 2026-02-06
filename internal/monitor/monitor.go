package monitor

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const compactHeightThreshold = 10

type Monitor struct {
	temperatureSensors []TemperatureSensor
	batteryStatus      BatteryStatus
	extraGroups        []SensorGroup
	lastUpdate         time.Time
	width, height      int
}

type TemperatureSensor struct {
	Name      string
	Value     float64 // in Celsius
	High      float64 // high threshold
	Critical  float64 // critical threshold
	Path      string  // sysfs path
}

type BatteryStatus struct {
	Capacity      int // percentage
	Status        string // Charging, Discharging, Full, Unknown
	Voltage       float64 // volts
	Current       float64 // amperes
	Power         float64 // watts
	Health        string // Health status
	Temperature   float64 // Celsius
	Energy        float64 // watt-hours
	CapacityLevel string // capacity level (Full, Normal, etc.)
}

func NewMonitor() Monitor {
	return Monitor{
		temperatureSensors: []TemperatureSensor{},
		batteryStatus:      BatteryStatus{},
		extraGroups:        []SensorGroup{},
		lastUpdate:         time.Now(),
	}
}

// RegisterSensorGroup adds a new group of sensors to the monitor.
// This enables easy extension with new types of system monitoring.
func (m *Monitor) RegisterSensorGroup(group SensorGroup) {
	m.extraGroups = append(m.extraGroups, group)
}

func (m Monitor) Init() tea.Cmd {
	return m.tick()
}

func (m Monitor) Update(msg tea.Msg) (Monitor, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		m = m.updateSensors()
		m.lastUpdate = time.Now()
		return m, m.tick()
	}
	return m, nil
}

func (m Monitor) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Use compact view for small panes
	if m.height < compactHeightThreshold {
		return m.compactView()
	}

	var sb strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		PaddingBottom(1)
	sb.WriteString(titleStyle.Render("System Status Monitor"))
	sb.WriteString("\n\n")

	// Temperatures section
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Temperatures"))
	sb.WriteString("\n")
	if len(m.temperatureSensors) == 0 {
		sb.WriteString("  No temperature sensors found\n")
	} else {
		for _, sensor := range m.temperatureSensors {
			color := "42" // green
			if sensor.Value >= sensor.Critical {
				color = "9" // red
			} else if sensor.Value >= sensor.High {
				color = "214" // orange
			}
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			tempStr := style.Render(fmt.Sprintf("%6.1f°C", sensor.Value))
			fmt.Fprintf(&sb, "  %-8s  %s\n", tempStr, sensor.Path)
		}
	}
	sb.WriteString("\n")

	// Battery section
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Battery"))
	sb.WriteString("\n")
	bat := m.batteryStatus
	if bat.Capacity == 0 && bat.Status == "" {
		sb.WriteString("  No battery information\n")
	} else {
		capacityColor := "42"
		if bat.Capacity < 20 {
			capacityColor = "9"
		} else if bat.Capacity < 50 {
			capacityColor = "214"
		}
		capacityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(capacityColor))
		fmt.Fprintf(&sb, "  Capacity: %s\n", capacityStyle.Render(fmt.Sprintf("%d%%", bat.Capacity)))
		fmt.Fprintf(&sb, "  Status: %s\n", bat.Status)
		if bat.Voltage > 0 {
			fmt.Fprintf(&sb, "  Voltage: %.2fV\n", bat.Voltage)
		}
		if bat.Current != 0 {
			fmt.Fprintf(&sb, "  Current: %.2fA\n", bat.Current)
		}
		if bat.Power > 0 {
			fmt.Fprintf(&sb, "  Power: %.2fW\n", bat.Power)
		}
		if bat.Health != "" {
			fmt.Fprintf(&sb, "  Health: %s\n", bat.Health)
		}
		if bat.Temperature > 0 {
			fmt.Fprintf(&sb, "  Temperature: %.1f°C\n", bat.Temperature)
		}
		if bat.Energy > 0 {
			fmt.Fprintf(&sb, "  Energy: %.2f Wh\n", bat.Energy)
		}
		if bat.CapacityLevel != "" {
			fmt.Fprintf(&sb, "  Capacity Level: %s\n", bat.CapacityLevel)
		}
	}

	// Extra sensor groups
	for _, group := range m.extraGroups {
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render(group.Name))
		sb.WriteString("\n")
		if len(group.Sensors) == 0 {
			sb.WriteString("  No sensors\n")
		} else {
			for _, sensor := range group.Sensors {
				color := "42" // green
				if sensor.Critical() {
					color = "9" // red
				} else if sensor.Warning() {
					color = "214" // orange
				}
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				fmt.Fprintf(&sb, "  %-20s: %s\n", sensor.Name(), style.Render(sensor.Value()))
			}
		}
	}

	// Footer
	sb.WriteString("\n")
	footerStyle := lipgloss.NewStyle().Faint(true)
	sb.WriteString(footerStyle.Render(fmt.Sprintf("Last updated: %s | Press 'q' to quit", m.lastUpdate.Format("15:04:05"))))

	return sb.String()
}

// compactView renders a minimal display suitable for small panes (≤3 lines)
func (m Monitor) compactView() string {
	var lines []string

	// Combine temperature and battery on first line if both present
	var firstLine strings.Builder
	// Temperature
	if len(m.temperatureSensors) > 0 {
		var highest TemperatureSensor
		for _, sensor := range m.temperatureSensors {
			if sensor.Value > highest.Value {
				highest = sensor
			}
		}
		color := "42" // green
		if highest.Value >= highest.Critical {
			color = "9" // red
		} else if highest.Value >= highest.High {
			color = "214" // orange
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		fmt.Fprintf(&firstLine, "🌡 %s", style.Render(fmt.Sprintf("%.1f°C", highest.Value)))
		if len(m.temperatureSensors) > 1 {
			fmt.Fprintf(&firstLine, " (%d)", len(m.temperatureSensors))
		}
	}
	// Battery
	bat := m.batteryStatus
	if bat.Capacity > 0 || bat.Status != "" {
		if firstLine.Len() > 0 {
			firstLine.WriteString(" | ")
		}
		capacityColor := "42"
		if bat.Capacity < 20 {
			capacityColor = "9"
		} else if bat.Capacity < 50 {
			capacityColor = "214"
		}
		capacityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(capacityColor))
		fmt.Fprintf(&firstLine, "🔋 %s %s", capacityStyle.Render(fmt.Sprintf("%d%%", bat.Capacity)), bat.Status)
		if bat.Voltage > 0 {
			fmt.Fprintf(&firstLine, " %.2fV", bat.Voltage)
		}
	}
	if firstLine.Len() > 0 {
		lines = append(lines, firstLine.String())
	}

	// Extra groups summary (second line)
	if len(m.extraGroups) > 0 {
		totalSensors := 0
		for _, group := range m.extraGroups {
			totalSensors += len(group.Sensors)
		}
		lines = append(lines, fmt.Sprintf("Extra: %d groups, %d sensors", len(m.extraGroups), totalSensors))
	}

	// Footer with update time (always last line)
	footerStyle := lipgloss.NewStyle().Faint(true)
	lines = append(lines, footerStyle.Render(fmt.Sprintf("Updated: %s", m.lastUpdate.Format("15:04:05"))))

	// Ensure we don't exceed 3 lines
	maxLines := 3
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

// tickMsg is a message sent periodically to update sensor readings
type tickMsg time.Time

func (m Monitor) tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Monitor) updateSensors() Monitor {
	// Update built-in sensors
	m.temperatureSensors = ReadTemperatures()
	m.batteryStatus = ReadBatteryStatus()

	// Refresh extra sensor groups
	for _, group := range m.extraGroups {
		for _, sensor := range group.Sensors {
			_ = sensor.Refresh() // Ignore errors for now
		}
	}
	return m
}

