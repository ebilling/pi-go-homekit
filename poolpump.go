package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"encoding/json"
	"log"
	"os"
)

var config map[string]string

func cmd(command string) {
	file, err := os.OpenFile(config["path.cmdfifo"], os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	if _, err := file.Write([]byte(command)); err != nil {
		log.Println(err)
	}
}

func turnPumpOn() {
	cmd("PUMP_ON")
}

func turnSweepOn() {
	cmd("SWEEP_ON")
}

func turnAllOff() {
	cmd("OFF")
}


func readConfig() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	pumpinfo := accessory.Info{
		Name:         "Pool Pump",
		Manufacturer: "Bonnie Labs",
	}

	sweepinfo := accessory.Info{
		Name:         "Pool Sweep",
		Manufacturer: "Bonnie Labs",
	}

	readConfig()

	pump := accessory.NewSwitch(pumpinfo)
	pump.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			turnPumpOn()
		} else {
			turnAllOff()
		}
	})

	sweep := accessory.NewSwitch(sweepinfo)
	sweep.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			turnSweepOn()
		} else {
			turnAllOff()
		}
	})

	t := NewThermometer(config["path.temperature"])
	go t.RunLoop()

	thermometer := accessory.NewTemperatureSensor(accessory.Info{
		Name:         "Pool Temp",
		Manufacturer: "Bonnie Labs",
	}, t.Temperature(), 0.0, 100.0, 1.0)
	
	log.Println("Pin:" + config["homekit.pin"])

	transport, err := hc.NewIPTransport(hc.Config{Pin: config["homekit.pin"]},
		pump.Accessory,
		sweep.Accessory,
		thermometer.Accessory)

	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
		transport.Stop()
	})

	transport.Start()
}
