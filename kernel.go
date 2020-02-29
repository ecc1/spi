package spi

// Definitions from <linux/spi/spidev.h>
// C naming is used for ease in keeping this file in sync.

type spi_ioc_transfer struct {
	tx_buf uint64
	rx_buf uint64

	len      uint32
	speed_hz uint32

	delay_usecs   uint16
	bits_per_word uint8
	cs_change     uint8
	tx_nbits      uint8
	rx_nbits      uint8
	pad           uint16
}

// Not all of these are used, but are defined for completeness.
const (
	spi_CPHA = 0x01
	spi_CPOL = 0x02

	spi_MODE_0 = 0
	spi_MODE_1 = spi_CPHA
	spi_MODE_2 = spi_CPOL
	spi_MODE_3 = spi_CPOL | spi_CPHA

	spi_CS_HIGH   = 0x04
	spi_LSB_FIRST = 0x08
	spi_3WIRE     = 0x10
	spi_LOOP      = 0x20
	spi_NO_CS     = 0x40
	spi_READY     = 0x80
	spi_TX_DUAL   = 0x100
	spi_TX_QUAD   = 0x200
	spi_RX_DUAL   = 0x400
	spi_RX_QUAD   = 0x800

	spi_IOC_MESSAGE_base = 0x40006B00
	spi_IOC_MESSAGE_incr = 0x200000

	// Read / Write of SPI mode (spi_MODE_0..spi_MODE_3) (limited to 8 bits)
	spi_IOC_RD_MODE = 0x80016B01
	spi_IOC_WR_MODE = 0x40016B01

	// Read / Write SPI bit justification
	spi_IOC_RD_LSB_FIRST = 0x80016B02
	spi_IOC_WR_LSB_FIRST = 0x40016B02

	// Read / Write SPI device word length (1..N)
	spi_IOC_RD_BITS_PER_WORD = 0x80016B03
	spi_IOC_WR_BITS_PER_WORD = 0x40016B03

	// Read / Write SPI device default max speed Hz
	spi_IOC_RD_MAX_SPEED_HZ = 0x80046B04
	spi_IOC_WR_MAX_SPEED_HZ = 0x40046B04

	// Read / Write of the SPI mode field
	spi_IOC_RD_MODE32 = 0x80046B05
	spi_IOC_WR_MODE32 = 0x40046B05
)

func spi_IOC_MESSAGE(n uint) uint {
	return spi_IOC_MESSAGE_base + n*spi_IOC_MESSAGE_incr
}
