package monitor

import (
	"fmt"
	"strings"
	"testing"
)

func TestCompactView(t *testing.T) {
	m := NewMonitor()
	// Set up some dummy temperature sensors
	m.temperatureSensors = []TemperatureSensor{
		{Name: "CPU", Value: 65.0, High: 80.0, Critical: 100.0, Path: "thermal_zone0"},
		{Name: "GPU", Value: 72.5, High: 85.0, Critical: 105.0, Path: "thermal_zone1"},
	}
	// Set up dummy battery status
	m.batteryStatus = BatteryStatus{
		Capacity: 85,
		Status:   "Charging",
		Voltage:  3.7,
	}
	// No extra groups

	output := m.compactView()
	lines := strings.Split(output, "\n")
	if len(lines) > 3 {
		t.Errorf("compactView should output at most 3 lines, got %d:\n%s", len(lines), output)
	}
	// Ensure temperature appears
	if !strings.Contains(output, "🌡") {
		t.Error("compactView should include temperature icon")
	}
	if !strings.Contains(output, "72.5°C") {
		t.Error("compactView should include highest temperature")
	}
	if !strings.Contains(output, "65.0°C") {
		t.Error("compactView should include all temperatures")
	}
	// Ensure battery appears
	if !strings.Contains(output, "🔋") {
		t.Error("compactView should include battery icon")
	}
	if !strings.Contains(output, "85%") {
		t.Error("compactView should include battery capacity")
	}
	fmt.Printf("Compact view output (%d lines):\n%s\n", len(lines), output)
}

func TestCompactViewNoSensors(t *testing.T) {
	m := NewMonitor()
	// No sensors, no battery
	output := m.compactView()
	lines := strings.Split(output, "\n")
	if len(lines) > 3 {
		t.Errorf("compactView should output at most 3 lines, got %d:\n%s", len(lines), output)
	}
	// Should only have footer line
	if len(lines) != 1 {
		t.Errorf("expected only footer line, got %d lines: %v", len(lines), lines)
	}
	if !strings.Contains(output, "Updated:") {
		t.Error("compactView should include update time")
	}
}

func TestCompactViewOnlyBattery(t *testing.T) {
	m := NewMonitor()
	m.batteryStatus = BatteryStatus{
		Capacity: 30,
		Status:   "Discharging",
	}
	output := m.compactView()
	lines := strings.Split(output, "\n")
	if len(lines) > 3 {
		t.Errorf("compactView should output at most 3 lines, got %d:\n%s", len(lines), output)
	}
	// Should have battery line and footer line (maybe 2 lines)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (battery + footer), got %d: %v", len(lines), lines)
	}
	if !strings.Contains(output, "🔋") {
		t.Error("compactView should include battery icon")
	}
	if !strings.Contains(output, "30%") {
		t.Error("compactView should include battery capacity")
	}
}

func TestCompactViewWithExtraGroups(t *testing.T) {
	m := NewMonitor()
	m.extraGroups = []SensorGroup{
		{
			Name: "Custom",
			Sensors: []Sensor{
				NewGenericSensor("Sensor1", func() (string, bool, bool, error) {
					return "OK", false, false, nil
				}),
			},
		},
	}
	output := m.compactView()
	lines := strings.Split(output, "\n")
	if len(lines) > 3 {
		t.Errorf("compactView should output at most 3 lines, got %d:\n%s", len(lines), output)
	}
	// Should have extra groups line and footer line (maybe 2 lines)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (extra + footer), got %d: %v", len(lines), lines)
	}
	if !strings.Contains(output, "Extra:") {
		t.Error("compactView should include extra groups summary")
	}
}

func TestViewUsesCompactWhenHeightSmall(t *testing.T) {
	m := NewMonitor()
	// Set up some data
	m.temperatureSensors = []TemperatureSensor{
		{Name: "CPU", Value: 65.0, High: 80.0, Critical: 100.0, Path: "thermal_zone0"},
	}
	m.batteryStatus = BatteryStatus{
		Capacity: 50,
		Status:   "Discharging",
	}
	// Set height below threshold
	m.height = compactHeightThreshold - 1
	m.width = 80
	output := m.View()
	lines := strings.Split(output, "\n")
	// Should be compact view (≤3 lines)
	if len(lines) > 3 {
		t.Errorf("View with small height should output ≤3 lines, got %d:\n%s", len(lines), output)
	}
	// Should contain compact indicators (icons)
	if !strings.Contains(output, "🌡") && !strings.Contains(output, "🔋") {
		t.Error("Compact view should include icons")
	}
}

func TestViewUsesFullWhenHeightLarge(t *testing.T) {
	m := NewMonitor()
	m.temperatureSensors = []TemperatureSensor{
		{Name: "CPU", Value: 65.0, High: 80.0, Critical: 100.0, Path: "thermal_zone0"},
	}
	m.batteryStatus = BatteryStatus{
		Capacity: 50,
		Status:   "Discharging",
	}
	// Set height above threshold
	m.height = compactHeightThreshold + 5
	m.width = 80
	output := m.View()
	// Should contain full view section headers
	if !strings.Contains(output, "Temperatures") {
		t.Error("Full view should include 'Temperatures' header")
	}
	if !strings.Contains(output, "Battery") {
		t.Error("Full view should include 'Battery' header")
	}
}