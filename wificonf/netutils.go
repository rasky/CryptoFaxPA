package main

import (
	"errors"
	"sort"
	"time"

	"pifke.org/wpasupplicant"
)

func WpaCurrentNetwork() (string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return "", err
	}
	defer w.Close()

	if status, err := w.Status(); err != nil {
		return "", err
	} else {
		return status.SSID(), nil
	}
}

// Scan for Wifi networks -- note that this function is blocking and takes
// a while.
func WpaScan() ([]string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return nil, err
	}
	defer w.Close()

	if err := w.Scan(); err != nil {
		return nil, err
	}

	// Wait for the end of scanning, or 5 seconds as timeout
	t0 := time.Now()
WaitLoop:
	for time.Since(t0) < 5*time.Second {
		select {
		case evt := <-w.EventQueue():
			if evt.Event == "SCAN_RESULTS" {
				break WaitLoop
			}
		case <-time.After(1 * time.Second):
		}
	}

	res, errs := w.ScanResults()
	if len(errs) != 0 {
		return nil, errs[0]
	}

	seen := make(map[string]bool)
	var ssids []string
	for _, bss := range res {
		s := bss.SSID()
		if !seen[s] {
			ssids = append(ssids, s)
			seen[s] = true
		}
	}

	sort.Strings(ssids)
	return ssids, nil
}

// Returns the list of known networks
func WpaKnownNetworks() ([]string, error) {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return nil, err
	}
	defer w.Close()

	res, err := w.ListNetworks()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var ssids []string
	for _, bss := range res {
		s := bss.SSID()
		if !seen[s] {
			ssids = append(ssids, s)
			seen[s] = true
		}
	}

	sort.Strings(ssids)
	return ssids, nil
}

func WpaAddNetwork(name, pass string) error {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return err
	}
	defer w.Close()

	idx, err := w.AddNetwork()
	if err != nil {
		return err
	}

	if err := w.SetNetwork(idx, "ssid", name); err != nil {
		return err
	}
	if pass != "" {
		if err := w.SetNetwork(idx, "psk", pass); err != nil {
			return err
		}
	} else {
		if err := w.SetNetwork(idx, "key_mgmt", "NONE"); err != nil {
			return err
		}
	}
	if err := w.EnableNetwork(idx); err != nil {
		return err
	}

	return w.SaveConfig()
}

func WpaRemoveNetwork(name string) error {
	w, err := wpasupplicant.Unixgram("wlan0")
	if err != nil {
		return err
	}
	defer w.Close()

	res, err := w.ListNetworks()
	if err != nil {
		return err
	}

	for idx, bss := range res {
		if bss.SSID() == name {
			if err := w.RemoveNetwork(idx); err != nil {
				return err
			}
			return w.SaveConfig()
		}
	}

	return errors.New("the specified network was not found in the configuration")
}
