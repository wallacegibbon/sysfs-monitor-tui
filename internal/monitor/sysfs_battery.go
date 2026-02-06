package monitor

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	powerSupplyBasePath = "/sys/class/power_supply"
)

func ReadBatteryStatus() BatteryStatus {
	status := BatteryStatus{}

	// Find battery directories by scanning all power supplies and checking type
	var batteryPath string
	entries, err := os.ReadDir(powerSupplyBasePath)
	if err != nil {
		return status
	}
	for _, entry := range entries {
		typePath := filepath.Join(powerSupplyBasePath, entry.Name(), "type")
		data, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "Battery" {
			batteryPath = filepath.Join(powerSupplyBasePath, entry.Name())
			break
		}
	}
	if batteryPath == "" {
		return status
	}

	// Read capacity
	capacityPath := filepath.Join(batteryPath, "capacity")
	if data, err := os.ReadFile(capacityPath); err == nil {
		if cap, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			status.Capacity = cap
		}
	}

	// Read status
	statusPath := filepath.Join(batteryPath, "status")
	if data, err := os.ReadFile(statusPath); err == nil {
		status.Status = strings.TrimSpace(string(data))
	}

	// Read voltage (in microvolts)
	voltagePath := filepath.Join(batteryPath, "voltage_now")
	if data, err := os.ReadFile(voltagePath); err == nil {
		if microvolts, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			status.Voltage = float64(microvolts) / 1_000_000.0
		}
	}

	// Read current (in microamperes)
	currentPath := filepath.Join(batteryPath, "current_now")
	if data, err := os.ReadFile(currentPath); err == nil {
		if microamps, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			status.Current = float64(microamps) / 1_000_000.0
		}
	}

	// Read power (in microwatts)
	powerPath := filepath.Join(batteryPath, "power_now")
	if data, err := os.ReadFile(powerPath); err == nil {
		if microwatts, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			status.Power = float64(microwatts) / 1_000_000.0
		}
	}

	// If power not available but voltage and current are, calculate power
	if status.Power == 0 && status.Voltage > 0 && status.Current != 0 {
		status.Power = status.Voltage * status.Current
	}

	// Read health
	healthPath := filepath.Join(batteryPath, "health")
	if data, err := os.ReadFile(healthPath); err == nil {
		status.Health = strings.TrimSpace(string(data))
	}

	// Read temperature (in tenths of degree Celsius)
	tempPath := filepath.Join(batteryPath, "temp")
	if data, err := os.ReadFile(tempPath); err == nil {
		if temp, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			status.Temperature = float64(temp) / 10.0
		}
	}

	// Read energy (in micro-watt-hours)
	energyPath := filepath.Join(batteryPath, "energy_now")
	if data, err := os.ReadFile(energyPath); err == nil {
		if microWh, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			status.Energy = float64(microWh) / 1_000_000.0
		}
	}

	// Read capacity level
	capacityLevelPath := filepath.Join(batteryPath, "capacity_level")
	if data, err := os.ReadFile(capacityLevelPath); err == nil {
		status.CapacityLevel = strings.TrimSpace(string(data))
	}

	return status
}

// Helper function to check if battery exists
func batteryExists() bool {
	_, err := os.Stat(powerSupplyBasePath)
	if os.IsNotExist(err) {
		return false
	}
	entries, err := os.ReadDir(powerSupplyBasePath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		typePath := filepath.Join(powerSupplyBasePath, entry.Name(), "type")
		data, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "Battery" {
			return true
		}
	}
	return false
}