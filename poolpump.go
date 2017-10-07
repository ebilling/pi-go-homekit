package main

import (
	"github.com/brutella/hc/accessory"
	"log"
	"os"
	"strconv"
	"time"
)

var mftr = "Bonnie Labs"

type Temp struct {
	therm *Thermometer
	acc   *accessory.Thermometer
}

type PoolPumpController struct {
	config      Config
	pump        *accessory.Switch
	sweep       *accessory.Switch
	waterTemp   *Temp
	roofTemp    *Temp
	pin         string
	done        chan bool
}

func NewTemp(data Config, key string, name string) *Temp {
	info := accessory.Info{
		Name: name,
		Manufacturer: mftr,
	}
	th := NewThermometer(key)
	t := Temp{
		therm: th,
		acc:   accessory.NewTemperatureSensor(info, th.Temperature(), 0.0, 100.0, 1.0),
	}
	return &t
}

func (t *Temp) Update(data *Config) {
	t.acc.TempSensor.CurrentTemperature.SetValue(t.therm.Update(data))
}

func NewPoolPumpController(path string) *PoolPumpController {
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

	tpath, exists := config.Get("path.temperature")
	if !exists {
		log.Printf("No entry for path.temperature")
	}
	tempdata := NewConfig(tpath)
	ppc.waterTemp = NewTemp(*tempdata, "waterTempC", "Water Temp")
	ppc.roofTemp = NewTemp(*tempdata, "roofTempC", "Roof Temp")

	ppc.pin, _ = config.Get("homekit.pin")
	log.Println("Homekit Pin: " + ppc.pin)

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
	log.Println("Writing command")
	_, err = fifo.WriteString(command + "\n")
	if err != nil {
		log.Println("Command Write Error: " + err.Error())
	}
}

func (ppc *PoolPumpController) turnPumpOn() {
	log.Println("Turning Pump On")
	ppc.cmd("PUMP_ON")
}

func (ppc *PoolPumpController) turnSweepOn() {
	log.Println("Turning Sweep On")
	ppc.cmd("SWEEP_ON")
}

func (ppc *PoolPumpController) turnAllOff() {
	log.Println("Turning Pumps Off")
	ppc.cmd("OFF")
}

//TODO update the temperature in the accessory
func (ppc *PoolPumpController) Update() {
	tdatapath, _ := ppc.config.Get("path.temperature")
	tdata := NewConfig(tdatapath)
	ppc.waterTemp.Update(tdata)
	ppc.roofTemp.Update(tdata)

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
	tries := 0
	for {
		if tries % 12 == 0 {
			log.Println("Still Running")
		}
		select {
		case <- time.After(interval):
			ppc.Update()
		case <- ppc.done:
			break
		}
		tries++
	}
}

func (ppc *PoolPumpController) Start() {
	go ppc.RunLoop()
}

func (ppc *PoolPumpController) Stop() {
	ppc.done <- true
}
