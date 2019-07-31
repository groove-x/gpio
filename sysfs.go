package gpio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type direction uint

const (
	inDirection direction = iota
	outDirection
)

type Edge uint

const (
	EdgeNone Edge = iota
	EdgeRising
	EdgeFalling
	EdgeBoth
)

type LogicLevel uint

const (
	ActiveHigh LogicLevel = iota
	ActiveLow
)

type Value uint

const (
	Inactive Value = 0
	Active   Value = 1
)

func exportGPIO(p Pin) error {
	if _, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%d", int(p.Number))); err == nil {
		return nil
	}

	export, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio export file for writing: %s", err)
	}
	defer export.Close()
	_, err = export.Write([]byte(strconv.Itoa(int(p.Number))))
	if err != nil {
		return fmt.Errorf("failed to write gpio export file: %s", err)
	}
	return nil
}

func unexportGPIO(p Pin) error {
	export, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio unexport file for writing: %s", err)
	}
	defer export.Close()
	_, err = export.Write([]byte(strconv.Itoa(int(p.Number))))
	if err != nil {
		return fmt.Errorf("failed to write gpio unexport file: %s", err)
	}
	return nil
}

func setDirection(p Pin, d direction, initialValue uint) error {
	dir, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio %d direction file for writing: %s", p.Number, err)
	}
	defer dir.Close()

	switch {
	case d == inDirection:
		_, err = dir.Write([]byte("in"))
	case d == outDirection && initialValue == 0:
		_, err = dir.Write([]byte("low"))
	case d == outDirection && initialValue == 1:
		_, err = dir.Write([]byte("high"))
	default:
		return fmt.Errorf("setDirection called with invalid direction or initialValue: %d, %d", d, initialValue)
	}
	if err != nil {
		return fmt.Errorf("failed to write gpio direction file: %s", err)
	}
	return nil
}

func setEdgeTrigger(p Pin, e Edge) error {
	edge, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/edge", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio %d edge file for writing: %s", p.Number, err)
	}
	defer edge.Close()

	switch e {
	case EdgeNone:
		_, err = edge.Write([]byte("none"))
	case EdgeRising:
		_, err = edge.Write([]byte("rising"))
	case EdgeFalling:
		_, err = edge.Write([]byte("falling"))
	case EdgeBoth:
		_, err = edge.Write([]byte("both"))
	default:
		return fmt.Errorf("setEdgeTrigger called with invalid edge %d", e)
	}
	if err != nil {
		return fmt.Errorf("failed to write gpio edge file: %s", err)
	}
	return nil
}

func setLogicLevel(p Pin, l LogicLevel) error {
	level, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/active_low", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio %d active_low file for writing: %s", p.Number, err)
	}
	defer level.Close()

	switch l {
	case ActiveHigh:
		_, err = level.Write([]byte("0"))
	case ActiveLow:
		_, err = level.Write([]byte("1"))
	default:
		return errors.New("invalid logic level setting")
	}
	if err != nil {
		return fmt.Errorf("failed to write gpio active_low file: %s", err)
	}
	return nil
}

func openPin(p Pin, write bool) (Pin, error) {
	flags := os.O_RDONLY
	if write {
		flags = os.O_RDWR
	}
	f, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", p.Number), flags, 0600)
	if err != nil {
		return Pin{}, fmt.Errorf("failed to open gpio %d value file for reading: %s", p.Number, err)
	}
	p.f = f
	return p, nil
}

func readPin(p Pin) (val uint, err error) {
	file := p.f
	file.Seek(0, 0)
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read: %s", err)
	}
	c := buf[0]
	switch c {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	default:
		return 0, fmt.Errorf("read inconsistent value in pinfile, %c", c)
	}
}

func writePin(p Pin, v uint) error {
	var buf []byte
	switch v {
	case 0:
		buf = []byte{'0'}
	case 1:
		buf = []byte{'1'}
	default:
		return fmt.Errorf("invalid output value %d", v)
	}
	_, err := p.f.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to write: %s", err)
	}
	return nil
}
