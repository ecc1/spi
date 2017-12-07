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
		return nil, fmt.Errorf("%s: %v", spiDevice, err)
	}
	var dev *Device
	if customCS == 0 {
		// Ensure exclusive access if using default chip-select.
		err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
		switch err {
		case nil:
			dev = &Device{fd: fd, speed: speed}
		case unix.EWOULDBLOCK:
			_ = unix.Close(fd)
			err = fmt.Errorf("%s: device is in use", spiDevice)
		default:
			_ = unix.Close(fd)
			err = fmt.Errorf("%s: %v", spiDevice, err)
		}
		return dev, err
	}
	cs, err := gpio.Output(customCS, true, false)
	if err != nil {
		_ = unix.Close(fd)
		err = fmt.Errorf("GPIO %d for chip select: %v", customCS, err)
	} else {
		dev = &Device{fd: fd, speed: speed, cs: cs}
	}
	return dev, err
}

// Close closes the SPI device.
func (dev *Device) Close() error {
	return unix.Close(dev.fd)
}

// Write writes len(buf) bytes from buf to dev.
func (dev *Device) Write(buf []byte) error {
	if dev.cs != nil {
		_ = dev.cs.Write(true)
		defer func() { _ = dev.cs.Write(false) }()
	}
	n, err := unix.Write(dev.fd, buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("wrote %d bytes instead of %d", n, len(buf))
	}
	return nil
}

// Read reads from dev into buf, blocking if necessary
// until exactly len(buf) bytes have been read.
func (dev *Device) Read(buf []byte) error {
	if dev.cs != nil {
		_ = dev.cs.Write(true)
		defer func() { _ = dev.cs.Write(false) }()
	}
	for off := 0; off < len(buf); {
		n, err := unix.Read(dev.fd, buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}

// Transfer uses buf for an SPI transfer operation (send and receive).
// The received data overwrites buf.
func (dev *Device) Transfer(buf []byte) error {
	if dev.cs != nil {
		_ = dev.cs.Write(true)
		defer func() { _ = dev.cs.Write(false) }()
	}
	bufAddr := uint64(uintptr(unsafe.Pointer(&buf[0])))
	tr := spi_ioc_transfer{
		tx_buf:        bufAddr,
		rx_buf:        bufAddr,
		len:           uint32(len(buf)),
		speed_hz:      uint32(dev.speed),
		delay_usecs:   1,
		bits_per_word: 8,
	}
	return dev.syscall(spi_IOC_MESSAGE(1), (*int)(unsafe.Pointer(&tr)))
}

// Mode returns the mode of the SPI device.
func (dev *Device) Mode() (int, error) {
	var mode int
	err := dev.syscall(spi_IOC_RD_MODE, &mode)
	return mode, err
}

// SetMode sets the mode of the SPI device.
func (dev *Device) SetMode(mode int) error {
	return dev.syscall(spi_IOC_WR_MODE, &mode)
}

// LSBFirst returns bit order of the SPI device.
func (dev *Device) LSBFirst() (bool, error) {
	var b int
	err := dev.syscall(spi_IOC_RD_LSB_FIRST, &b)
	if b != 0 {
		return true, err
	}
	return false, err
}

// SetLSBFirst sets the bit order of the SPI device.
func (dev *Device) SetLSBFirst(lsb bool) error {
	var b int
	if lsb {
		b = 1
	}
	return dev.syscall(spi_IOC_WR_LSB_FIRST, &b)
}

// BitsPerWord returns the word size of the SPI device.
func (dev *Device) BitsPerWord() (int, error) {
	var bits int
	err := dev.syscall(spi_IOC_RD_BITS_PER_WORD, &bits)
	return bits, err
}

// SetBitsPerWord sets the word size of the SPI device.
func (dev *Device) SetBitsPerWord(bits int) error {
	return dev.syscall(spi_IOC_WR_BITS_PER_WORD, &bits)
}

// MaxSpeed returns the maximum speed of the SPI device, in Hertz.
func (dev *Device) MaxSpeed() (int, error) {
	var speed int
	err := dev.syscall(spi_IOC_RD_MAX_SPEED_HZ, &speed)
	return speed, err
}

// SetMaxSpeed sets the maximum speed of the SPI device, in Hertz.
func (dev *Device) SetMaxSpeed(speed int) error {
	return dev.syscall(spi_IOC_WR_MAX_SPEED_HZ, &speed)
}

func (dev *Device) syscall(op uint, arg *int) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL,
		uintptr(dev.fd), uintptr(op), uintptr(unsafe.Pointer(arg)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}
