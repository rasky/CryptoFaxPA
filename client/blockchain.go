package main

import (
	"bytes"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/antonholmquist/jason"
	chart "github.com/wcharczuk/go-chart"
)

type Elem struct {
	t time.Time
	v float64
}

// Return a graph in PNG format representing the 1-month BTC-USD exchange
func GetBitcoinGraph() []byte {
	res, err := http.Get("https://api.coindesk.com/v1/bpi/historical/close.json")
	if err != nil {
		log.Printf("[ERROR] coinbase: cannot fetch JSON: %v", err)
		return nil
	}
	defer res.Body.Close()

	v, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		log.Printf("[ERROR] coinbase: cannot unmarshal bitcoin JSON: %v", err)
		return nil
	}

	bpi, err := v.GetObject("bpi")
	if err != nil {
		log.Printf("[ERROR] coinbase: cannot decode JSON API")
		return nil
	}

	var elems []Elem
	for jt, jv := range bpi.Map() {
		t, err := time.Parse("2006-01-02", jt)
		if err != nil {
			continue
		}
		v, err := jv.Float64()
		if err != nil {
			continue
		}

		elems = append(elems, Elem{t, v})
	}

	sort.Slice(elems, func(i, j int) bool {
		return elems[i].t.Before(elems[j].t)
	})

	var times []time.Time
	var values []float64

	for _, e := range elems {
		times = append(times, e.t)
		values = append(values, e.v)
	}

	graph := chart.Chart{
		Width:  360,
		Height: 180,
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: times,
				YValues: values,
			},
		},
	}

	var buf bytes.Buffer
	graph.Render(chart.PNG, &buf)
	return buf.Bytes()
}
