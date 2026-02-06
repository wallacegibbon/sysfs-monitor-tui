# System Monitoring Agents

This document describes the monitoring agents available in the sysfs-tui-monitor system.

## Overview

The sysfs-tui-monitor is built around a flexible agent architecture where each monitoring component (agent) implements the `Sensor` interface. Agents can be grouped into logical categories and displayed in the TUI interface.

## Built-in Agents

### 1. Temperature Monitoring Agent

**Purpose**: Monitors CPU and system temperatures via Linux sysfs thermal interfaces.

**Sysfs Paths**:
- `/sys/class/thermal/thermal_zone*` - Standard thermal zones
- `/sys/class/hwmon/hwmon*` - Hardware monitoring sensors

**Data Collected**:
- Temperature in °C (converted from millidegree Celsius)
- Sensor name (from `type` file or hwmon label)
- High threshold (default: 80°C)
- Critical threshold (default: 100°C)

**Implementation**: `ReadTemperatures()` in `sysfs_temperature.go`

### 2. Battery Monitoring Agent

**Purpose**: Monitors battery status and health via Linux sysfs power supply interface.

**Sysfs Path**: `/sys/class/power_supply/`

**Detection Method**: Scans all entries in power supply directory and checks the `type` file for "Battery" value. This supports non-standard battery naming (e.g., `sbs-5-000b`).

**Data Collected**:
- **Basic**: Capacity (%), Status (Charging/Discharging/Full/Unknown), Voltage (V), Current (A), Power (W)
- **Extended** (when available):
  - Health status (Good/Overheat/Dead/Over voltage/Unspecified failure/Unknown)
  - Temperature (°C) - converted from tenths of degree Celsius
  - Energy (Wh) - converted from micro-watt-hours
  - Capacity Level (Full/Normal/Low/Critical)

**Implementation**: `ReadBatteryStatus()` in `sysfs_battery.go`

**Key Enhancement**: The agent now correctly detects batteries regardless of naming convention by checking device type instead of relying on directory name patterns.

## Extensible Architecture

### Sensor Interface

All agents implement the `Sensor` interface defined in `sensor.go`:

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

Sensors are organized into `SensorGroup` collections:

```go
type SensorGroup struct {
    Name    string
    Sensors []Sensor
}
```

Groups appear as separate sections in the TUI display.

### Adapters

The system includes adapters to convert existing data structures to the `Sensor` interface:

- `TemperatureSensorAdapter`: Adapts `TemperatureSensor` to `Sensor`
- `BatterySensorAdapter`: Adapts `BatteryStatus` to `Sensor`
- `GenericSensor`: Simple implementation for custom sensors

## Creating Custom Agents

### Method 1: Using GenericSensor

For simple key-value sensors:

```go
sensor := NewGenericSensor("CustomSensor", func() (string, bool, bool, error) {
    value := readSomeValue()
    warning := value > warningThreshold
    critical := value > criticalThreshold
    return fmt.Sprintf("%.1f units", value), warning, critical, nil
})
```

### Method 2: Implementing Sensor Interface

For complex sensors requiring more control:

```go
type CustomSensor struct {
    name    string
    value   float64
    // ... additional fields
}

func (c *CustomSensor) Name() string { return c.name }
func (c *CustomSensor) Value() string { return fmt.Sprintf("%.2f", c.value) }
func (c *CustomSensor) Warning() bool { return c.value > 50 }
func (c *CustomSensor) Critical() bool { return c.value > 80 }
func (c *CustomSensor) Refresh() error {
    c.value = readFromSystem()
    return nil
}
```

### Method 3: Extending Existing Agents

You can create wrapper agents that augment existing functionality:

```go
type EnhancedBatterySensor struct {
    *BatteryStatus
    cycleCount int
}

func (e *EnhancedBatterySensor) Value() string {
    return fmt.Sprintf("%d%% (%d cycles)", e.Capacity, e.cycleCount)
}
```

## Registering Agents

Add custom sensor groups to the monitor:

```go
monitor := monitor.NewMonitor()
customGroup := monitor.SensorGroup{
    Name: "Custom Sensors",
    Sensors: []monitor.Sensor{customSensor1, customSensor2},
}
monitor.RegisterSensorGroup(customGroup)
```

## Agent Lifecycle

1. **Initialization**: Agents are created when `NewMonitor()` is called
2. **Refresh**: All sensors are refreshed every 2 seconds via `tickMsg`
3. **Display**: Sensor values are rendered in the TUI via `View()` method
4. **Cleanup**: No explicit cleanup required (Go garbage collection)

## Sysfs Data Sources

### Temperature Data Files
- `/sys/class/thermal/thermal_zone*/temp` - Temperature in millidegree Celsius
- `/sys/class/thermal/thermal_zone*/type` - Sensor type
- `/sys/class/hwmon/hwmon*/temp*_input` - Temperature inputs
- `/sys/class/hwmon/hwmon*/temp*_label` - Sensor labels
- `/sys/class/hwmon/hwmon*/temp*_crit` - Critical thresholds

### Battery Data Files
- `/sys/class/power_supply/*/type` - Device type (Battery/Mains/USB)
- `/sys/class/power_supply/*/capacity` - Capacity percentage
- `/sys/class/power_supply/*/status` - Charging status
- `/sys/class/power_supply/*/voltage_now` - Voltage in microvolts
- `/sys/class/power_supply/*/current_now` - Current in microamperes
- `/sys/class/power_supply/*/power_now` - Power in microwatts
- `/sys/class/power_supply/*/health` - Battery health
- `/sys/class/power_supply/*/temp` - Temperature in tenths of °C
- `/sys/class/power_supply/*/energy_now` - Energy in micro-watt-hours
- `/sys/class/power_supply/*/capacity_level` - Capacity level

## Error Handling

Agents handle missing sysfs files gracefully:
- Missing files return empty/zero values
- Parsing errors are silently ignored
- The TUI displays available information only
- No system crashes due to missing sysfs data

## Performance Considerations

- **Polling Interval**: Default 2 seconds (configurable in `tick()`)
- **Batch Reading**: Temperature and battery agents read multiple files in single pass
- **Lazy Evaluation**: Values are only read when displayed
- **Caching**: No long-term caching - always reads fresh from sysfs

## Testing Agents

### Unit Tests
```go
func TestBatteryAgent(t *testing.T) {
    status := ReadBatteryStatus()
    if status.Capacity < 0 || status.Capacity > 100 {
        t.Errorf("Invalid capacity: %d", status.Capacity)
    }
}
```

### Integration Tests
```bash
# Test battery detection
go run cmd/sysfs-check/main.go

# Test full TUI
go run main.go
```

## Future Agent Extensions

Potential agents to implement:

1. **Memory Usage Agent**: Monitor RAM usage via `/proc/meminfo`
2. **CPU Usage Agent**: Monitor CPU utilization via `/proc/stat`
3. **Disk Usage Agent**: Monitor disk space via `sysfs` or `statfs`
4. **Network Agent**: Monitor network interfaces via `/sys/class/net`
5. **Process Agent**: Monitor specific process metrics via `/proc`

## Compact Display Mode

For small terminal panes (height < 10 lines), the monitor automatically switches to a compact view that fits within 3 lines:

**Compact View Format**:
1. **First line**: Combined temperature and battery status
   - 🌡 Highest temperature with color coding (green/orange/red)
   - 🔋 Battery capacity with color coding (green/orange/red) + status + voltage
   - Separated by " | " if both present
2. **Second line** (optional): Extra sensor groups summary
   - Shows count of groups and total sensors
   - Color-coded based on warning/critical status (green/orange/red)
   - Includes warning/critical counts when present
3. **Third line**: Update timestamp (faint text)

**Color Coding**:
- **Temperature**: Green (< high threshold), Orange (≥ high), Red (≥ critical)
- **Battery**: Green (≥ 50%), Orange (20-49%), Red (< 20%)
- **Extra Groups**: Green (no warnings/critical), Orange (any warnings), Red (any critical)

**Implementation**: `compactView()` method in `monitor.go` with automatic switching based on terminal height (`compactHeightThreshold` constant). The summary now includes warning/critical counts and color coding for extra groups.

## Contributing New Agents

1. Create agent implementation in `internal/monitor/`
2. Implement the `Sensor` interface
3. Add appropriate sysfs reading logic
4. Create adapter if needed
5. Update `CreateSensorGroups()` to include new agent
6. Add tests
7. Update documentation (this file)

## Troubleshooting

### Agent Not Detecting Hardware
- Check sysfs paths exist: `ls -la /sys/class/thermal/` or `/sys/class/power_supply/`
- Verify file permissions: `cat /sys/class/power_supply/*/type`
- Check kernel modules are loaded for your hardware

### Incorrect Readings
- Verify unit conversions (millidegree, microvolts, etc.)
- Check sysfs file formats with `cat`
- Some hardware reports 0 for unused sensors

### Performance Issues
- Increase polling interval in `tick()` function
- Consider caching for frequently accessed values
- Implement selective refreshing for slow sensors

## License & Attribution

This agent architecture is part of the sysfs-tui-monitor project. Agents are designed to be simple, extensible, and resilient to missing system data.

---
*Last Updated: 2026-02-06*
*System: Linux sysfs monitoring agents*