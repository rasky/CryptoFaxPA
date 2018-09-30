package common

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/antonholmquist/jason"
)

type geoIpService struct {
	Url    string
	TzKey  []string
	ErrKey []string
}

var geoIpServices = []geoIpService{
	{
		Url:    "http://ip-api.com/json/%s",
		TzKey:  []string{"timezone"},
		ErrKey: []string{"message"},
	},
	{
		Url:   "https://freegeoip.app/json/%s",
		TzKey: []string{"time_zone"},
	},
	{
		Url:    "http://geoip.nekudo.com/api/%s",
		TzKey:  []string{"location", "time_zone"},
		ErrKey: []string{"msg"},
	},
	{
		Url:    "https://timezoneapi.io/api/ip?%s",
		TzKey:  []string{"data", "timezone", "id"},
		ErrKey: []string{"meta", "message"},
	},
}

var (
	currentLoc atomic.Value
)

func updateTimezone(ip string) bool {
	// Contact all APIs in parallel, and use the first non-error returned
	tzchan := make(chan string, len(geoIpServices))
	for _, gip := range geoIpServices {
		go func(gip geoIpService) {
			if res, err := http.Get(fmt.Sprintf(gip.Url, ip)); err == nil {
				defer res.Body.Close()
				if v, err := jason.NewObjectFromReader(res.Body); err == nil {
					// Check if there's an error key; if so, ignore
					if _, err := v.GetString(gip.ErrKey...); err == nil {
						return
					}
					// Check if we managed to read the timezone (and it is not empty)
					if tz, err := v.GetString(gip.TzKey...); err == nil && tz != "" {
						tzchan <- tz
					}
				}
			}
		}(gip)
	}

	select {
	case tz := <-tzchan:
		if loc, err := time.LoadLocation(tz); err == nil {
			log.Printf("[INFO] current timezone: %v", tz)
			currentLoc.Store(loc)
			return true
		}
	case <-time.After(10 * time.Second):
		// timeout, ignore
	}

	return false
}

// PollTimezone starts polling for timezone changes using the current IP
// address. It should be launched at startup.
func PollTimezone() {
	for {
		done := updateTimezone("") // use auto-detected IP
		if done {
			time.Sleep(6 * time.Hour)
		} else {
			time.Sleep(10 * time.Minute)
		}
	}
}

// NowHere is like time.Now(), but returns a time adjusted with the current
// location (if known).
func NowHere() time.Time {
	if loc := currentLoc.Load(); loc != nil {
		return time.Now().In(loc.(*time.Location))
	}
	return time.Now()
}
