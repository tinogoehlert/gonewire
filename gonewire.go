package gonewire

import (
	"context"
	"fmt"
	"os"
	"time"
)

const (
	defaultDirectory = "/sys/bus/w1/devices/"
	minFrequency     = 500 * time.Millisecond
	chanSize         = 8
)

type Value struct {
	ID    string
	Value string
	Type  string
}

type Gonewire struct {
	valueChannel chan Value
	directory    string
	sensormap    map[string]*Sensor
	errCballback func(err error, sensor *Sensor)
}

func New(dir string) (*Gonewire, error) {
	if dir == "" {
		dir = defaultDirectory
	}

	gw := &Gonewire{
		directory:    dir,
		valueChannel: make(chan Value, chanSize),
		sensormap:    make(map[string]*Sensor),
		errCballback: defaultErrorCallback,
	}
	if err := gw.readFolder(); err != nil {
		return nil, err
	}
	return gw, nil
}

func (gw *Gonewire) Values() chan Value {
	return gw.valueChannel
}

func (gw *Gonewire) Start(ctx context.Context, frequency time.Duration) {
	if frequency < minFrequency {
		frequency = minFrequency
	}
	for {
		select {
		case <-time.After(frequency):
			for _, sensor := range gw.sensormap {
				if err := sensor.parseValue(); err != nil {
					gw.errCballback(err, sensor)
					continue
				}
				gw.valueChannel <- Value{
					ID:    sensor.id,
					Value: sensor.currentValue,
					Type:  sensor.typeString,
				}
			}
		case <-ctx.Done():
			for _, sensor := range gw.sensormap {
				sensor.close()
			}
			return
		}
	}
}

func (gw *Gonewire) OnReadError(fn func(error, *Sensor)) {
	gw.errCballback = fn
}

func (gw *Gonewire) readFolder() error {
	dir, err := os.Open(gw.directory)
	if err != nil {
		return fmt.Errorf("could not read dir: %w", err)
	}
	defer dir.Close()

	subdirs, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("could not read sub dirs: %w", err)
	}

	for _, subdir := range subdirs {
		if subdir == "w1_bus_master1" {
			continue
		}
		s, err := newSensor(gw.directory, subdir)
		if err != nil {
			fmt.Println("could not init sensor: %w", err)
			continue
		}
		gw.sensormap[s.id] = s
	}

	return nil
}

func defaultErrorCallback(err error, sensor *Sensor) {
	return
}
