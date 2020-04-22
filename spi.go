package spi

import (
	"fmt"
	"unsafe"

	"github.com/ecc1/gpio"
	"golang.org/x/sys/unix"
)

// Device represents an SPI device.
type Device struct {
	fd    int
	speed int
	cs    gpio.OutputPin
}

// Open opens the given SPI device at the specified speed (in Hertz)
// If customCS in not zero, that pin number is used as a custom chip-select.
func Open(spiDevice string, speed int, customCS int) (*Device, error) {
	fd, err := unix.Open(spiDevice, unix.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", spiDevice, err)
	}
	// Ensure exclusive access.
	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	switch err {
	case nil:
		if customCS == 0 {
			return &Device{fd: fd, speed: speed}, nil
		}
	case unix.EWOULDBLOCK:
		_ = unix.Close(fd)
		return nil, fmt.Errorf("%s: device is in use", spiDevice)
	default:
		_ = unix.Close(fd)
		return nil, fmt.Errorf("%s: %w", spiDevice, err)
	}
	// Use specified GPIO pin as custom chip-select.
	cs, err := gpio.Output(customCS, true, false)
	if err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("GPIO %d for chip select: %w", customCS, err)
	}
	return &Device{fd: fd, speed: speed, cs: cs}, nil
}

// Close closes the SPI device.
func (dev *Device) Close() error {
	return unix.Close(dev.fd)
}

// Transfer performs an SPI transfer operation (send and receive).
func (dev *Device) Transfer(snd, rcv []byte) error {
	if len(snd) != len(rcv) {
		return fmt.Errorf("transfer buffers must be the same length (snd = %d, rcv = %d)", len(snd), len(rcv))
	}
	if dev.cs != nil {
		dev.cs.Write(true)
		defer dev.cs.Write(false)
	}
	sndAddr := uint64(uintptr(unsafe.Pointer(&snd[0])))
	rcvAddr := uint64(uintptr(unsafe.Pointer(&rcv[0])))
	tr := spi_ioc_transfer{
		tx_buf:        sndAddr,
		rx_buf:        rcvAddr,
		len:           uint32(len(snd)),
		speed_hz:      uint32(dev.speed),
		delay_usecs:   0,
		bits_per_word: 8,
	}
	return dev.syscall(spi_IOC_MESSAGE(1), unsafe.Pointer(&tr))
}

// Mode returns the mode of the SPI device.
func (dev *Device) Mode() (uint8, error) {
	var mode uint8
	err := dev.syscallU8(spi_IOC_RD_MODE, &mode)
	return mode, err
}

// SetMode sets the mode of the SPI device.
func (dev *Device) SetMode(mode uint8) error {
	return dev.syscallU8(spi_IOC_WR_MODE, &mode)
}

// LSBFirst returns bit order of the SPI device.
func (dev *Device) LSBFirst() (bool, error) {
	var b uint8
	err := dev.syscallU8(spi_IOC_RD_LSB_FIRST, &b)
	if b != 0 {
		return true, err
	}
	return false, err
}

// SetLSBFirst sets the bit order of the SPI device.
func (dev *Device) SetLSBFirst(lsb bool) error {
	var b uint8
	if lsb {
		b = 1
	}
	return dev.syscallU8(spi_IOC_WR_LSB_FIRST, &b)
}

// BitsPerWord returns the word size of the SPI device.
func (dev *Device) BitsPerWord() (int, error) {
	var bits uint8
	err := dev.syscallU8(spi_IOC_RD_BITS_PER_WORD, &bits)
	return int(bits), err
}

// SetBitsPerWord sets the word size of the SPI device.
func (dev *Device) SetBitsPerWord(n int) error {
	bits := uint8(n)
	return dev.syscallU8(spi_IOC_WR_BITS_PER_WORD, &bits)
}

// MaxSpeed returns the maximum speed of the SPI device, in Hertz.
func (dev *Device) MaxSpeed() (int, error) {
	var speed uint32
	err := dev.syscallU32(spi_IOC_RD_MAX_SPEED_HZ, &speed)
	return int(speed), err
}

// SetMaxSpeed sets the maximum speed of the SPI device, in Hertz.
func (dev *Device) SetMaxSpeed(n int) error {
	speed := uint32(n)
	return dev.syscallU32(spi_IOC_WR_MAX_SPEED_HZ, &speed)
}

func (dev *Device) syscallU8(op uint, arg *uint8) error {
	return dev.syscall(op, unsafe.Pointer(arg))
}

func (dev *Device) syscallU32(op uint, arg *uint32) error {
	return dev.syscall(op, unsafe.Pointer(arg))
}

func (dev *Device) syscall(op uint, arg unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(dev.fd), uintptr(op), uintptr(arg))
	if errno != 0 {
		return error(errno)
	}
	return nil
}
