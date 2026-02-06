# System Monitoring Agents

## Overview

Flexible agent architecture where each monitoring component implements the `Sensor` interface. Agents are grouped into logical categories for TUI display.

## Built-in Agents

### 1. Temperature Monitoring Agent
- **Purpose**: Monitors CPU/system temperatures via Linux sysfs thermal interfaces
- **Sysfs Paths**: `/sys/class/thermal/thermal_zone*`, `/sys/class/hwmon/hwmon*`
- **Data**: Temperature (°C), sensor name, thresholds (high: 80°C, critical: 100°C)
- **Threshold Validation**: Negative threshold values (e.g., `trip_point_*_temp`, `crit`, `max`) are ignored; default thresholds apply
- **Implementation**: `ReadTemperatures()` in `sysfs_temperature.go`

### 2. Battery Monitoring Agent
- **Purpose**: Monitors battery status and health via sysfs power supply interface
- **Sysfs Path**: `/sys/class/power_supply/`
- **Detection**: Checks `type` file for "Battery" value (supports non-standard naming)
- **Data**: Capacity (%), status, voltage, current, power, health, temperature, energy, capacity level
- **Implementation**: `ReadBatteryStatus()` in `sysfs_battery.go`

## Architecture

### Sensor Interface
```go
type Sensor interface {
    Name() string      // Human-readable identifier
    Value() string     // Current reading as display string
    Warning() bool     // True if in warning state
    Critical() bool    // True if in critical state
    Refresh() error    // Update reading from system
}
```

### Sensor Groups
```go
type SensorGroup struct {
    Name    string
    Sensors []Sensor
}
```

### Adapters
- `TemperatureSensorAdapter`: Adapts `TemperatureSensor` to `Sensor`
- `BatterySensorAdapter`: Adapts `BatteryStatus` to `Sensor`
- `GenericSensor`: Simple implementation for custom sensors

## Creating Custom Agents

### Method 1: Using GenericSensor
```go
sensor := NewGenericSensor("CustomSensor", func() (string, bool, bool, error) {
    value := readSomeValue()
    warning := value > warningThreshold
    critical := value > criticalThreshold
    return fmt.Sprintf("%.1f units", value), warning, critical, nil
})
```

### Method 2: Implementing Sensor Interface
Implement `Name()`, `Value()`, `Warning()`, `Critical()`, `Refresh()` methods.

## Compact Display Mode

For small terminal panes (height < 10 lines), automatic compact view (≤3 lines):

**Compact View Format**:
1. **First line**: Multiple temperatures (all shown in row) and battery status
   - 🌡 65.0°C 72.5°C (all temperatures with color coding)
   - 🔋 85% Charging 3.70V (capacity with color coding)
   - Separated by " | " if both present
2. **Second line** (optional): Extra sensor groups summary with warning/critical counts
3. **Third line**: Update timestamp

**Non-compact View**:
- Two-column layout: temperatures on left, battery info on right
- Side-by-side display with 4-space separation
- Extra sensor groups follow below

**Color Coding**:
- **Temperature**: Green (< high), Orange (≥ high), Red (≥ critical)
- **Battery**: Green (≥ 50%), Orange (20-49%), Red (< 20%)
- **Extra Groups**: Green (no warnings/critical), Orange (any warnings), Red (any critical)

## Future Agent Extensions

Potential agents to implement:
1. **Memory Usage Agent**: Monitor RAM via `/proc/meminfo`
2. **CPU Usage Agent**: Monitor CPU via `/proc/stat`
3. **Disk Usage Agent**: Monitor disk space via `sysfs`/`statfs`
4. **Network Agent**: Monitor interfaces via `/sys/class/net`
5. **Process Agent**: Monitor process metrics via `/proc`

## Contributing New Agents

1. Create agent in `internal/monitor/`
2. Implement `Sensor` interface
3. Add sysfs reading logic
4. Create adapter if needed
5. Update `CreateSensorGroups()` (if exists)
6. Add tests
7. Update this document

---
*Last Updated: 2026-02-07*
*System: Linux sysfs monitoring agents*