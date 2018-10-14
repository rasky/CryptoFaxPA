#!/bin/bash
set -euo pipefail
set -x

GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w"
ssh pi@cryptofaxpa.local rm client
scp client pi@cryptofaxpa.local:
ssh pi@cryptofaxpa.local sudo killall client || true
ssh pi@cryptofaxpa.local tail -f /var/log/syslog
