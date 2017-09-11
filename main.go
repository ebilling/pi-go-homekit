package main

import (
	"github.com/brutella/hc"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: pi-go-homekit CONFIG")
	}
	ppc := NewPoolPumpController(os.Args[1])
	ppc.Start()

	transport, err := hc.NewIPTransport(hc.Config{Pin: ppc.pin},
		ppc.thermometer.Accessory,
		ppc.pump.Accessory,
		ppc.sweep.Accessory)

	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		ppc.Stop()
		transport.Stop()
	})

	transport.Start()
	log.Println("Exiting")
}
