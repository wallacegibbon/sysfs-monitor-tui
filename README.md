# sysfs-tui-monitor

A terminal-based system monitor for Linux that displays CPU temperatures and battery status using sysfs.

## Features

- **Temperature Monitoring**: Real-time CPU/core temperatures from `/sys/class/thermal/`
- **Battery Monitoring**: Capacity, status, voltage, current, power, and health from `/sys/class/power_supply/`
- **Color-coded Alerts**: Green (normal), orange (warning), red (critical)
- **Compact View**: Automatic 3-line view for small terminal panes
- **Extensible**: Add custom sensors via the `Sensor` interface

## Usage

Run the monitor:

```bash
go run main.go
```

Press `q` or `Ctrl+C` to quit.

### Normal View

![Normal View](normal-view.gif)

### Compact View

![Compact View](compact-view.gif)

## Installation

```bash
git clone <repo-url>
cd sysfs-tui-monitor
go run main.go
```

## Requirements

- Linux with sysfs
- Go 1.25+
- Terminal with color support

## License

MIT
