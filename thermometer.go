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
	return &Thermometer{
		path: path,
		done:   make(chan bool),
	}
}

func (t *Thermometer) Stop() {
	t.done <- true
}

func (t *Thermometer) Temperature() float64 {
	return t.temperature
}

func (t *Thermometer) getTemperature() float64 {
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

	return (celsius)
}

func (t *Thermometer) RunLoop() {
	t.temperature = t.getTemperature()
	for {
		select {
		case <-time.After(interval):
			t.temperature = t.getTemperature()
		case <-t.done:
			break
		}
	}
}
