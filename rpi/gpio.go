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
func OpenPin(number int, mode gpio.Mode, direction gpio.PullDirection) (gpio.Pin, error) {
	initOnce.Do(initRPi)
	pull(uint8(number), direction)
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

func pull(pin uint8, direction gpio.PullDirection) {
	clk_offset := uint32(pin / 32)
	gppud_clk := (*uint32)(unsafe.Pointer(&gpiomem[BCM2835_GPPUDCLK0+clk_offset*4]))
	shift := (pin % 32)

	switch direction {
	case gpio.PullUp:
		*gppud = 2
	case gpio.PullDown:
		*gppud = 1
	default:
		*gppud = 0
	}

	time.Sleep(5 * time.Microsecond)
	*gppud_clk = 1 << shift
	time.Sleep(5 * time.Microsecond)
	*gppud = 0
	*gppud_clk = 0
}
