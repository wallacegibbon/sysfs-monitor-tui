package monitor

// Sensor represents a generic system sensor that can be monitored
type Sensor interface {
	// Name returns a human-readable identifier
	Name() string
	// Value returns the current reading as a string suitable for display
	Value() string
	// Warning returns true if the sensor is in warning state
	Warning() bool
	// Critical returns true if the sensor is in critical state
	Critical() bool
	// Refresh updates the sensor reading from the system
	Refresh() error
}

// SensorGroup represents a collection of sensors under a category
type SensorGroup struct {
	Name    string
	Sensors []Sensor
}

// GenericSensor is a simple implementation of Sensor for basic key-value pairs
type GenericSensor struct {
	name      string
	value     string
	warning   bool
	critical  bool
	refreshFn func() (string, bool, bool, error)
}

func NewGenericSensor(name string, refreshFn func() (string, bool, bool, error)) *GenericSensor {
	return &GenericSensor{
		name:      name,
		refreshFn: refreshFn,
	}
}

func (g *GenericSensor) Name() string {
	return g.name
}

func (g *GenericSensor) Value() string {
	return g.value
}

func (g *GenericSensor) Warning() bool {
	return g.warning
}

func (g *GenericSensor) Critical() bool {
	return g.critical
}

func (g *GenericSensor) Refresh() error {
	if g.refreshFn != nil {
		value, warning, critical, err := g.refreshFn()
		if err != nil {
			return err
		}
		g.value = value
		g.warning = warning
		g.critical = critical
	}
	return nil
}