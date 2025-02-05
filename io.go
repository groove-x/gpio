package gpio

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Pin represents a single pin, which can be used either for reading or writing
type Pin struct {
	Number    uint
	direction direction
	f         *os.File
}

func retry(retryN int, retryDuration time.Duration, fn func() error) error {
	for i := 0; ; i++ {
		err := fn()
		if err != nil {
			if i == retryN-1 {
				return err
			} else {
				fmt.Println(err.Error())
				fmt.Printf("retrying...")
				time.Sleep(retryDuration)
			}
		} else {
			break
		}
	}
	return nil
}

func NewInput(p uint) (Pin, error) {
	return NewInputWithRetry(p, 1, 0)
}

// NewInputWithRetry opens the given pin number for reading. The number provided should be the pin number known by the kernel
func NewInputWithRetry(p uint, retryN int, retryDuration time.Duration) (Pin, error) {
	pin := Pin{
		Number: p,
	}

	err := retry(retryN, retryDuration, func() error {
		err := exportGPIO(pin)
		return err
	})
	if err != nil {
		return Pin{}, err
	}

	time.Sleep(10 * time.Millisecond)
	pin.direction = inDirection

	err = retry(retryN, retryDuration, func() error {
		err = setDirection(pin, inDirection, 0)
		if err != nil {
			return err
		}
		pin, err = openPin(pin, false)
		return err
	})
	if err != nil {
		return Pin{}, err
	}

	return pin, nil
}

// NewOutputWithRetry opens the given pin number for writing. The number provided should be the pin number known by the kernel
// NewOutputWithRetry also needs to know whether the pin should be initialized high (true) or low (false)
func NewOutputWithRetry(p uint, initHigh bool, retryN int, retryDuration time.Duration) (Pin, error) {
	var err error

	pin := Pin{
		Number: p,
	}

	err = retry(retryN, retryDuration, func() error {
		return exportGPIO(pin)
	})
	if err != nil {
		return Pin{}, err
	}

	time.Sleep(10 * time.Millisecond)
	initVal := uint(0)
	if initHigh {
		initVal = uint(1)
	}
	pin.direction = outDirection


	err = retry(retryN, retryDuration, func() error {
		return setDirection(pin, outDirection, initVal)
	})
	if err != nil {
		return Pin{}, err
	}

	err = retry(retryN, retryDuration, func() error {
		pin, err = openPin(pin, true)
		return err
	})
	if err != nil {
		return Pin{}, err
	}
	return pin, nil
}

// Close releases the resources related to Pin. This doen't unexport Pin, use Cleanup() instead
func (p Pin) Close() {
	if p.f != nil {
		p.f.Close()
		p.f = nil
	}
}

// Cleanup close Pin and unexport it
func (p Pin) Cleanup() {
	p.Close()
	unexportGPIO(p)
}

// Read returns the value read at the pin as reported by the kernel. This should only be used for input pins
func (p Pin) Read() (value uint, err error) {
	if p.direction != inDirection {
		return 0, errors.New("pin is not configured for input")
	}
	return readPin(p)
}

// SetLogicLevel sets the logic level for the Pin. This can be
// either "active high" or "active low"
func (p Pin) SetLogicLevel(logicLevel LogicLevel) error {
	return setLogicLevel(p, logicLevel)
}

// High sets the value of an output pin to logic high
func (p Pin) High() error {
	if p.direction != outDirection {
		return errors.New("pin is not configured for output")
	}
	return writePin(p, 1)
}

// Low sets the value of an output pin to logic low
func (p Pin) Low() error {
	if p.direction != outDirection {
		return errors.New("pin is not configured for output")
	}
	return writePin(p, 0)
}
