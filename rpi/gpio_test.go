package rpi

import (
	"github.com/Hella-Info/gpio"
)

// assert that rpi.pin implements gpio.Pin
var _ gpio.Pin = new(pin)
