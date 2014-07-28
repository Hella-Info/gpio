package rpi

import (
	"log"
	"syscall"
	"time"
	"unsafe"

	"github.com/Hella-Info/gpio"
)

var (
	gpfsel, gpset, gpclr, gplev []*uint32
	gppud                       *uint32
	gpiomem                     []byte
)

func initGPIO(memfd int) {
	var err error
	gpiomem, err = syscall.Mmap(memfd, BCM2835_GPIO_BASE, BCM2835_BLOCK_SIZE, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("rpi: unable to mmap GPIO page: %v", err)
	}
	gpfsel = []*uint32{
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL0])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL1])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL2])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL3])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL4])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPFSEL5])),
	}
	gpset = []*uint32{
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPSET0])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPSET1])),
	}
	gpclr = []*uint32{
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPCLR0])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPCLR1])),
	}
	gplev = []*uint32{
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPLEV0])),
		(*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPLEV1])),
	}
	gppud = (*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPPUD]))
}

// pin represents a specalised RPi GPIO pin with fast paths for
// several operations.
type pin struct {
	gpio.Pin       // the underlying Pin implementation
	pin      uint8 // the actual pin number
}

// OpenPin returns a gpio.Pin implementation specalised for the RPi.
func OpenPin(number int, mode gpio.Mode) (gpio.Pin, error) {
	initOnce.Do(initRPi)
	p, err := gpio.OpenPin(number, mode)
	return &pin{Pin: p, pin: uint8(number)}, err
}

func (p *pin) Set() {
	offset := p.pin / 32
	shift := p.pin % 32
	*gpset[offset] = (1 << shift)
}

func (p *pin) Clear() {
	offset := p.pin / 32
	shift := p.pin % 32
	*gpclr[offset] = (1 << shift)
}

func (p *pin) Get() bool {
	offset := p.pin / 32
	shift := p.pin % 32
	return *gplev[offset]&(1<<shift) == (1 << shift)
}

func GPIOFSel(pin, mode uint8) {
	offset := pin / 10
	shift := (pin % 10) * 3
	value := *gpfsel[offset]
	mask := BCM2835_GPIO_FSEL_MASK << shift
	value &= ^uint32(mask)
	value |= uint32(mode) << shift
	*gpfsel[offset] = value & mask
}

func (p *pin) Pull(direction gpio.PullDirection) {
	gppud_clk := (*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPPUDCLK0+p.pin/8]))
	shift := (p.pin % 8)

	switch direction {
	case gpio.PullUp:
		*gppud = (*gppud & ^uint32(3)) | 2
	case gpio.PullDown:
		*gppud = (*gppud & ^uint32(3)) | 1
	default:
		*gppud &= ^uint32(3)
	}

	time.Sleep(150 * time.Nanosecond)
	*gppud_clk = 1 << shift
	time.Sleep(150 * time.Nanosecond)
	*gppud &= ^uint32(3)
	*gppud_clk = 0
}
