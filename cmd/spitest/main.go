package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/ecc1/spi"
)

var (
	device   = flag.String("d", "/dev/spidev5.1", "SPI `device`")
	speed    = flag.Int("s", 1000000, "SPI `speed` (Hz)")
	customCS = flag.Int("cs", 0, "use `GPIO#` as custom chip select")
)

func main() {
	flag.Parse()
	var values []byte
	for _, v := range flag.Args() {
		b, err := strconv.ParseUint(v, 16, 8)
		if err != nil {
			log.Fatal(err)
		}
		values = append(values, byte(b))
	}
	dev, err := spi.Open(*device, *speed, *customCS)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = dev.Close() }()
	if len(values)%2 == 1 {
		values = append(values, 0)
	}
	fmt.Printf("send: % X\n", values)
	err = dev.Transfer(values)
	if err != nil {
		log.Fatalf("%s: %v", *device, err)
	}
	fmt.Printf("recv: % X\n", values)
}
