package gonewire

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type SensorType int32

const (
	// same for DS18S20
	DS1820  SensorType = 10
	DS1822  SensorType = 22
	DS18B20 SensorType = 28
)

var (
	typeStringMap = map[SensorType]string{
		DS1820:  "DS1820",
		DS1822:  "DS1822",
		DS18B20: "DS18B20",
	}
)

type Sensor struct {
	id           string
	sensorType   SensorType
	typeString   string
	fh           *os.File
	currentValue string
}

func newSensor(rootDir, subDir string) (*Sensor, error) {
	var err error
	subDir = strings.TrimRight(subDir, "/")
	rootDir = strings.TrimRight(rootDir, "/")

	typ, err := strconv.ParseInt(subDir[0:2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not convert type: %w", err)
	}

	sensor := &Sensor{
		id:         subDir[3:],
		sensorType: SensorType(typ),
	}

	var ok bool
	if sensor.typeString, ok = typeStringMap[sensor.sensorType]; !ok {
		return nil, fmt.Errorf("unsupproted type %d", sensor.sensorType)
	}

	if sensor.fh, err = os.Open(rootDir + "/" + subDir + "/w1_slave"); err != nil {
		return nil, fmt.Errorf("could not open slave file: %w", err)
	}

	return sensor, nil
}

func (s *Sensor) ID() string {
	return s.id
}

func (s *Sensor) TypeString() string {
	return s.typeString
}

func (s *Sensor) parseValue() error {
	s.fh.Seek(0, 0)
	b, err := ioutil.ReadAll(s.fh)
	if err != nil {
		return fmt.Errorf("could not read slave file: %w", err)
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return errors.New("unsupported file format")
	}

	if !strings.Contains(lines[0], "YES") {
		return errors.New("checksum error")
	}

	lastIdx := strings.LastIndex(lines[1], "=")
	if lastIdx == -1 {
		return errors.New("could not find value")
	}

	s.currentValue = lines[1][lastIdx+1:]

	return nil
}

func (s *Sensor) close() {
	s.fh.Close()
}
