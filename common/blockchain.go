package common

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/antonholmquist/jason"
	humanize "github.com/dustin/go-humanize"
	"github.com/guptarohit/asciigraph"
	chart "github.com/wcharczuk/go-chart"
)

type Elem struct {
	t time.Time
	v float64
}

func getCoindeskData() []Elem {
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
	return elems
}

// Return a graph in PNG format representing the 1-month BTC-USD exchange
func GetBitcoinGraph() []byte {
	elems := getCoindeskData()
	if elems == nil {
		return nil
	}

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

func GetBitcoinAsciiGraph(width, height int) string {
	elems := getCoindeskData()
	if elems == nil {
		return ""
	}

	var values []float64
	for _, e := range elems {
		values = append(values, e.v)
	}

	return asciigraph.Plot(values, asciigraph.Width(width), asciigraph.Height(height))
}

type BlockchainNerdInfo struct {
	Name  string
	Value string
}

func httpGetString(url string, outerr *error) string {
	res, err := http.Get(url)
	if err != nil {
	    *outerr = err
		return ""
	}
	var buf bytes.Buffer
	io.Copy(&buf, res.Body)
	res.Body.Close()
	return buf.String()
}

func satoshis(s string) string {
	v, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return fmt.Sprintf("%.8f", float64(v)/float64(100000000))
	}
	return s
}

func seconds(s string) string {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return fmt.Sprintf("%v", time.Duration(v*1000000000))
	}
	return s
}

func human(s string) string {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return fmt.Sprintf("%v", humanize.Comma(int64(v)))
	}
	return s
}

func GetBlockchainNerdInfos() (infos []BlockchainNerdInfo, err error) {
	return []BlockchainNerdInfo{
		{Name: "Current BTC price (USD)", Value: "$" + httpGetString("https://blockchain.info/q/24hrprice", &err)},
		{Name: "Market cap (USD)", Value: "$" + human(httpGetString("https://blockchain.info/q/marketcap", &err))},
		{Name: "Global hash rate (GigaHash)", Value: human(httpGetString("https://blockchain.info/q/hashrate", &err))},
		{Name: "Current difficulty target", Value: human(httpGetString("https://blockchain.info/q/getdifficulty", &err))},
		{Name: "Current block height", Value: httpGetString("https://blockchain.info/q/getblockcount", &err)},
		{Name: "Latest hash", Value: httpGetString("https://blockchain.info/q/latesthash", &err)},
		{Name: "Current block reward", Value: satoshis(httpGetString("https://blockchain.info/q/bcperblock", &err))},
		{Name: "Total bitcoins", Value: satoshis(httpGetString("https://blockchain.info/q/totalbc", &err))},
		{Name: "Probability of mining", Value: httpGetString("https://blockchain.info/q/probability", &err)},
		{Name: "ETA until next block", Value: seconds(httpGetString("https://blockchain.info/q/eta", &err))},
	}, err
}
