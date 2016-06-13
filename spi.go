package spi

import (
	"fmt"
	"syscall"
	"unsafe"
)

type Device struct {
	fd    int
	speed int
}

const (
	// FIX THIS to support platforms other than Intel Edison.
	spiDevice = "/dev/spidev5.1"
)

func Open(speed int) (*Device, error) {
	fd, err := syscall.Open(spiDevice, syscall.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", spiDevice, err)
	}
	err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		return nil, fmt.Errorf("%s: device is in use", spiDevice)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %v", spiDevice, err)
	}
	return &Device{fd: fd, speed: speed}, nil
}

func (dev *Device) Close() error {
	return syscall.Close(dev.fd)
}

// Write writes len(buf) bytes from buf to dev.
func (dev *Device) Write(buf []byte) error {
	n, err := syscall.Write(dev.fd, buf)
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
	for off := 0; off < len(buf); {
		n, err := syscall.Read(dev.fd, buf[off:])
		if err != nil {
			return err
		}
		off += n
	}
	return nil
}

func (dev *Device) Transfer(buf []byte) error {
	bufAddr := uint64(uintptr(unsafe.Pointer(&buf[0])))
	tr := spi_ioc_transfer{
		tx_buf:        bufAddr,
		rx_buf:        bufAddr,
		len:           uint32(len(buf)),
		speed_hz:      uint32(dev.speed),
		delay_usecs:   1,
		bits_per_word: 8,
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_MESSAGE(1)), uintptr(unsafe.Pointer(&tr)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}

func (dev *Device) Mode() (mode int, err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_RD_MODE), uintptr(unsafe.Pointer(&mode)))
	if errno != 0 {
		err = error(errno)
	}
	return
}

func (dev *Device) SetMode(mode int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_WR_MODE), uintptr(unsafe.Pointer(&mode)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}

func (dev *Device) LsbFirst() (lsb bool, err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_RD_LSB_FIRST), uintptr(unsafe.Pointer(&lsb)))
	if errno != 0 {
		err = error(errno)
	}
	return
}

func (dev *Device) SetLsbFirst(lsb bool) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_WR_LSB_FIRST), uintptr(unsafe.Pointer(&lsb)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}

func (dev *Device) BitsPerWord() (bits int, err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_RD_BITS_PER_WORD), uintptr(unsafe.Pointer(&bits)))
	if errno != 0 {
		err = error(errno)
	}
	return
}

func (dev *Device) SetBitsPerWord(bits int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_WR_BITS_PER_WORD), uintptr(unsafe.Pointer(&bits)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}

func (dev *Device) MaxSpeed() (speed int, err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_RD_MAX_SPEED_HZ), uintptr(unsafe.Pointer(&speed)))
	if errno != 0 {
		err = error(errno)
	}
	return
}

func (dev *Device) SetMaxSpeed(speed int) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(dev.fd),
		uintptr(spi_IOC_WR_MAX_SPEED_HZ), uintptr(unsafe.Pointer(&speed)))
	if errno != 0 {
		return error(errno)
	}
	return nil
}
