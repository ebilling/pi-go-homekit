package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

var (
	interval = 60 * time.Second
)

type Thermometer struct {
	path      string
	temperature float64
	done        chan bool
}

func NewThermometer(path string) *Thermometer {
	th := Thermometer{
		path: path,
		done: make(chan bool),
	}
	th.readTemperature()
	return &th
}

func (t *Thermometer) Stop() {
	t.done <- true
}

func (t *Thermometer) Temperature() float64 {
	return t.temperature
}

func (t *Thermometer) readTemperature() float64 {
	file, err := os.Open(t.path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	data := make([]byte, 100)
	count, err := file.Read(data)
	if err != nil {
		log.Println(err)
	}
	if count < 3 {
		log.Println("Temperature doesn't seem to be valid")
	}

	celsius, err := strconv.ParseFloat(string(data[:count]), 64)
	if err != nil {
		log.Println("Could not convert temperature from device: " + err.Error())
	}
	t.temperature = celsius
	return (celsius)
}

func (t *Thermometer) RunLoop() {
	for {
		select {
		case <-time.After(interval):
			t.readTemperature()
		case <-t.done:
			break
		}
	}
}
