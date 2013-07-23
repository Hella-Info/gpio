package gpio_test

import (
	"testing"

	"github.com/davecheney/gpio/rpi"
	"github.com/davecheney/gpio"
)


func TestOpenPin(t *testing.T) {
	checkRoot(t)
	pin, err := gpio.OpenPin(rpi.GPIO_P1_22, gpio.ModeInput)
	if err != nil {
		t.Fatal(err)
	}
	err = pin.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// test opening pins from a non priviledged user fails
func TestOpenPinUnpriv(t *testing.T) {
	checkNotRoot(t)
	pin, err := gpio.OpenPin(rpi.GPIO_P1_22, gpio.ModeInput)
	if err == nil {
		pin.Close()
		t.Fatalf("OpenPin is expected to fail for non priv user")
	}
}

func TestSetDirection(t *testing.T) {
	checkRoot(t)
	pin, err := gpio.OpenPin(rpi.GPIO_P1_22, gpio.ModeInput)
	if err != nil {
		t.Fatal(err)
	}
	defer pin.Close()
	pin.SetMode(gpio.ModeOutput)
	if pin.Err() != nil {
		t.Fatal(err)
	}
	dir := pin.Mode()
	if err := pin.Err(); dir != gpio.ModeOutput || err != nil {
		t.Fatalf("pin.Mode(): expected %v %v , got %v %v", gpio.ModeOutput, nil, dir, err)
	}
}
