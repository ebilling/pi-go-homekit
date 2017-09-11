package main

import (
	"github.com/brutella/hc"
	"log"
	"os"
)

func main() {
	ppc := NewPoolPumpController(os.Args[1])
	ppc.Start()

	transport, err := hc.NewIPTransport(hc.Config{Pin: ppc.pin},
		ppc.pump.Accessory,
		ppc.sweep.Accessory,
		ppc.thermometer.Accessory)

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
