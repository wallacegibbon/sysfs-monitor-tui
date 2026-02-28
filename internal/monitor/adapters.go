package monitor

import "fmt"

// TemperatureSensorAdapter adapts TemperatureSensor to the Sensor interface
type TemperatureSensorAdapter struct {
	*TemperatureSensor
}

func (t TemperatureSensorAdapter) Name() string {
	return t.TemperatureSensor.Name
}

func (t TemperatureSensorAdapter) Value() string {
	return fmt.Sprintf("%.1f°C", t.TemperatureSensor.Value)
}

func (t TemperatureSensorAdapter) Warning() bool {
	return t.TemperatureSensor.Value >= t.TemperatureSensor.High
}

func (t TemperatureSensorAdapter) Critical() bool {
	return t.TemperatureSensor.Value >= t.TemperatureSensor.Critical
}

func (t TemperatureSensorAdapter) Refresh() error {
	// Temperature sensors are refreshed via batch readTemperatures()
	// Individual refresh not supported; rely on global update
	return nil
}

// BatterySensorAdapter adapts BatteryStatus to the Sensor interface
type BatterySensorAdapter struct {
	*BatteryStatus
}

func (b BatterySensorAdapter) Name() string {
	return "Battery"
}

func (b BatterySensorAdapter) Value() string {
	return fmt.Sprintf("%d%%", b.BatteryStatus.Capacity)
}

func (b BatterySensorAdapter) Warning() bool {
	return b.BatteryStatus.Capacity < 20
}

func (b BatterySensorAdapter) Critical() bool {
	return b.BatteryStatus.Capacity < 10
}

func (b BatterySensorAdapter) Refresh() error {
	// Battery status is refreshed via batch readBatteryStatus()
	return nil
}

// CreateSensorGroups creates default sensor groups from existing data
func CreateSensorGroups(temps []TemperatureSensor, battery BatteryStatus) []SensorGroup {
	groups := []SensorGroup{}

	// Temperature group
	if len(temps) > 0 {
		tempSensors := make([]Sensor, len(temps))
		for i := range temps {
			tempSensors[i] = TemperatureSensorAdapter{&temps[i]}
		}
		groups = append(groups, SensorGroup{
			Name:    "Temperatures",
			Sensors: tempSensors,
		})
	}

	// Battery group
	if battery.Capacity > 0 || battery.Status != "" {
		groups = append(groups, SensorGroup{
			Name:    "Battery",
			Sensors: []Sensor{BatterySensorAdapter{&battery}},
		})
	}

	return groups
}
