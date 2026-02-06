package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	thermalBasePath = "/sys/class/thermal"
)

func ReadTemperatures() []TemperatureSensor {
	var sensors []TemperatureSensor

	// Check if thermal directory exists
	if _, err := os.Stat(thermalBasePath); os.IsNotExist(err) {
		return sensors
	}

	// List thermal zones
	thermalZones, err := filepath.Glob(filepath.Join(thermalBasePath, "thermal_zone*"))
	if err != nil {
		return sensors
	}

	for _, zonePath := range thermalZones {
		sensor, err := readThermalZone(zonePath)
		if err == nil {
			sensors = append(sensors, sensor)
		}
	}

	// Also try hwmon sensors (commonly used for CPU, motherboard temperatures)
	hwmonPaths, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, hwmonPath := range hwmonPaths {
		sensors = append(sensors, readHwmonSensors(hwmonPath)...)
	}

	return sensors
}

func readThermalZone(zonePath string) (TemperatureSensor, error) {
	sensor := TemperatureSensor{}

	// Read temperature (in millidegree Celsius)
	tempPath := filepath.Join(zonePath, "temp")
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return sensor, err
	}
	tempStr := strings.TrimSpace(string(data))
	tempMilli, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		return sensor, err
	}
	sensor.Value = float64(tempMilli) / 1000.0
	sensor.Path = zonePath

	// Read sensor name
	typePath := filepath.Join(zonePath, "type")
	typeData, err := os.ReadFile(typePath)
	if err == nil {
		sensor.Name = strings.TrimSpace(string(typeData))
	} else {
		sensor.Name = filepath.Base(zonePath)
	}

	// Read thresholds if available
	highPath := filepath.Join(zonePath, "trip_point_0_temp")
	if data, err := os.ReadFile(highPath); err == nil {
		if highMilli, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			if highMilli >= 0 {
				sensor.High = float64(highMilli) / 1000.0
			}
		}
	}

	criticalPath := filepath.Join(zonePath, "trip_point_1_temp")
	if data, err := os.ReadFile(criticalPath); err == nil {
		if critMilli, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			if critMilli >= 0 {
				sensor.Critical = float64(critMilli) / 1000.0
			}
		}
	}

	// If thresholds not set, use sensible defaults
	if sensor.High == 0 {
		sensor.High = 80.0
	}
	if sensor.Critical == 0 {
		sensor.Critical = 100.0
	}

	return sensor, nil
}

func readHwmonSensors(hwmonPath string) []TemperatureSensor {
	var sensors []TemperatureSensor

	// Read hwmon name
	namePath := filepath.Join(hwmonPath, "name")
	nameData, err := os.ReadFile(namePath)
	if err != nil {
		return sensors
	}
	hwmonName := strings.TrimSpace(string(nameData))

	// Find temperature input files
	tempInputs, _ := filepath.Glob(filepath.Join(hwmonPath, "temp*_input"))
	for _, inputPath := range tempInputs {
		base := strings.TrimSuffix(filepath.Base(inputPath), "_input")
		labelPath := filepath.Join(hwmonPath, base+"_label")
		critPath := filepath.Join(hwmonPath, base+"_crit")
		maxPath := filepath.Join(hwmonPath, base+"_max")

		// Read temperature value
		data, err := os.ReadFile(inputPath)
		if err != nil {
			continue
		}
		tempStr := strings.TrimSpace(string(data))
		tempMilli, err := strconv.ParseInt(tempStr, 10, 64)
		if err != nil {
			continue
		}
		value := float64(tempMilli) / 1000.0

		// Determine sensor name
		var name string
		if labelData, err := os.ReadFile(labelPath); err == nil {
			name = strings.TrimSpace(string(labelData))
		} else {
			name = fmt.Sprintf("%s_%s", hwmonName, base)
		}

		sensor := TemperatureSensor{
			Name:     name,
			Value:    value,
			High:     80.0,
			Critical: 100.0,
			Path:     inputPath,
		}

		// Read critical threshold
		if critData, err := os.ReadFile(critPath); err == nil {
			if critMilli, err := strconv.ParseInt(strings.TrimSpace(string(critData)), 10, 64); err == nil {
				if critMilli >= 0 {
					sensor.Critical = float64(critMilli) / 1000.0
				}
			}
		}

		// Read max threshold as high
		if maxData, err := os.ReadFile(maxPath); err == nil {
			if maxMilli, err := strconv.ParseInt(strings.TrimSpace(string(maxData)), 10, 64); err == nil {
				if maxMilli >= 0 {
					sensor.High = float64(maxMilli) / 1000.0
				}
			}
		}

		sensors = append(sensors, sensor)
	}
	return sensors
}