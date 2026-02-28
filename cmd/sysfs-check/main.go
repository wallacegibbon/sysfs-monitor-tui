package main

import (
	"fmt"
	"github.com/wallacegibbon/sysfs-monitor-tui/internal/monitor"
)

func main() {
	fmt.Println("Testing sysfs monitoring...")

	temps := monitor.ReadTemperatures()
	fmt.Printf("Found %d temperature sensors:\n", len(temps))
	for _, t := range temps {
		fmt.Printf("  %s: %.1f°C (high %.1f, critical %.1f)\n", t.Name, t.Value, t.High, t.Critical)
	}

	battery := monitor.ReadBatteryStatus()
	fmt.Printf("\nBattery status:\n")
	if battery.Capacity == 0 && battery.Status == "" {
		fmt.Println("  No battery information")
	} else {
		fmt.Printf("  Capacity: %d%%\n", battery.Capacity)
		fmt.Printf("  Status: %s\n", battery.Status)
		if battery.Voltage > 0 {
			fmt.Printf("  Voltage: %.2fV\n", battery.Voltage)
		}
		if battery.Current != 0 {
			fmt.Printf("  Current: %.2fA\n", battery.Current)
		}
		if battery.Power > 0 {
			fmt.Printf("  Power: %.2fW\n", battery.Power)
		}
		if battery.Health != "" {
			fmt.Printf("  Health: %s\n", battery.Health)
		}
		if battery.Temperature > 0 {
			fmt.Printf("  Temperature: %.1f°C\n", battery.Temperature)
		}
		if battery.Energy > 0 {
			fmt.Printf("  Energy: %.2f Wh\n", battery.Energy)
		}
		if battery.CapacityLevel != "" {
			fmt.Printf("  Capacity Level: %s\n", battery.CapacityLevel)
		}
	}
}
