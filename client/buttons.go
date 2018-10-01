package main

import (
	"runtime"
	"time"

	rpio "github.com/stianeikeland/go-rpio"
)

type RPButton struct {
	rpio.Pin
	Edges chan time.Time
}

func NewRPButton(pinid int) *RPButton {
	if runtime.GOOS == "darwin" { // debuggin mode
		return &RPButton{
			Edges: make(chan time.Time),
		}
	}

	pin := rpio.Pin(pinid)
	pin.Input()
	pin.PullDown()
	pin.Detect(rpio.RiseEdge)

	rpb := &RPButton{
		Pin:   pin,
		Edges: make(chan time.Time, 16),
	}
	go rpb.detectEdges()
	return rpb
}

func (btn *RPButton) detectEdges() {
	for {
		if btn.EdgeDetected() {
			btn.Edges <- time.Now()
		}
		time.Sleep(50 * time.Millisecond)
	}
}
