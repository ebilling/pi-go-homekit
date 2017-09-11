package main

import (
	"github.com/brutella/hc/accessory"
	"log"
	"os"
	"strconv"
	"time"
)

type PoolPumpController struct {
	config      Config
	pump        *accessory.Switch
	sweep       *accessory.Switch
	temp        *Thermometer
	thermometer *accessory.Thermometer
	pin         string
	done        chan bool
}

func NewPoolPumpController(path string) *PoolPumpController {
	mftr := "Bonnie Labs"
	config := *NewConfig(path)
	ppc := PoolPumpController {
		config:    config,
		done:      make(chan bool),
	}

	pumpinfo := accessory.Info{
		Name:         "Pool Pump",
		Manufacturer: mftr,
	}

	sweepinfo := accessory.Info{
		Name:         "Pool Sweep",
		Manufacturer: mftr,
	}

	ppc.pump = accessory.NewSwitch(pumpinfo)
	ppc.pump.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			ppc.turnPumpOn()
		} else {
			ppc.turnAllOff()
		}
	})

	ppc.sweep = accessory.NewSwitch(sweepinfo)
	ppc.sweep.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			ppc.turnSweepOn()
		} else {
			ppc.turnAllOff()
		}
	})

	tpath, _ := config.Get("path.temperature")
	ppc.temp = NewThermometer(tpath)

	ppc.thermometer = accessory.NewTemperatureSensor(accessory.Info{
		Name:         "Pool Temp",
		Manufacturer: mftr,
	}, ppc.temp.Temperature(), 0.0, 100.0, 1.0)

	ppc.pin, _ = config.Get("homekit.pin")
	log.Println("Homekit Pin:" + ppc.pin)

	return &ppc
}

func (ppc *PoolPumpController) cmd(command string) {
	path, exists := ppc.config.Get("path.cmdfifo")
	if !exists {
		log.Println("path.cmdfifo not specified in the configuration file")
		return
	}
	fifo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		log.Println("Command Open Error: " + err.Error())
		return
	}
	defer fifo.Close()
	_, err = fifo.WriteString(command + "\n")
	if err != nil {
		log.Println("Command Write Error: " + err.Error())
	}
}

func (ppc *PoolPumpController) turnPumpOn() {
	ppc.cmd("PUMP_ON")
	log.Println("Turning Pump On")
}

func (ppc *PoolPumpController) turnSweepOn() {
	ppc.cmd("SWEEP_ON")
	log.Println("Turning Sweep On")
}

func (ppc *PoolPumpController) turnAllOff() {
	ppc.cmd("OFF")
	log.Println("Turning Pumps Off")
}

func (ppc *PoolPumpController) Update() {
	ppc.temp.readTemperature()
	path, _ := ppc.config.Get("path.status")
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	data := make([]byte, 100)
	count, err := file.Read(data)
	if err != nil {
		log.Println(err)
	}
	if count < 1 {
		log.Println("Status doesn't seem to be valid")
	}

	status, err := strconv.ParseInt(string(data[:count]), 10, 64)
	if err != nil {
		log.Println("Could not convert status: " + err.Error())
	}
	if status <= 0 {
		ppc.pump.Switch.On.SetValue(false)
		ppc.sweep.Switch.On.SetValue(false)
	} else if status%2 == 1 {
		ppc.pump.Switch.On.SetValue(true)
		ppc.sweep.Switch.On.SetValue(false)
	} else {
		ppc.pump.Switch.On.SetValue(true)
		ppc.sweep.Switch.On.SetValue(true)
	}
}

func (ppc *PoolPumpController) RunLoop() {
	interval := 5 * time.Second
	for {
		select {
		case <- time.After(interval):
			ppc.Update()
		case <- ppc.done:
			break
		}
	}
}

func (ppc *PoolPumpController) Start() {
	go ppc.RunLoop()
}

func (ppc *PoolPumpController) Stop() {
	ppc.done <- true
}
