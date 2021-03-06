package main

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

type RPButtonEvent struct {
	Pin  int
	When time.Time
}

type RPButtonMonitor struct {
	Events chan RPButtonEvent

	exit int32
	pins []rpio.Pin
}

func NewRPButtonMonitor(pinids ...int) *RPButtonMonitor {
	mon := &RPButtonMonitor{
		Events: make(chan RPButtonEvent, 16),
	}

	if runtime.GOOS == "darwin" { // debugging mode
		return mon
	}

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, id := range pinids {
		pin := rpio.Pin(id)
		pin.Input()
		pin.PullOff()
		mon.pins = append(mon.pins, pin)
	}

	go mon.detectEdges()
	return mon
}

func (mon *RPButtonMonitor) Shutdown() {
	atomic.StoreInt32(&mon.exit, 1)
	for atomic.LoadInt32(&mon.exit) != 2 {
		time.Sleep(50 * time.Millisecond)
	}
	rpio.Close()
}

func (mon *RPButtonMonitor) detectEdges() {
	for atomic.LoadInt32(&mon.exit) == 0 {
		for _, pin := range mon.pins {
			if pin.Read() == rpio.Low {
				mon.Events <- RPButtonEvent{
					Pin:  int(pin),
					When: time.Now(),
				}

				// wait until button is released
				for {
					if pin.Read() == rpio.High {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	atomic.StoreInt32(&mon.exit, 2)
}
