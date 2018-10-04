#!/bin/bash
set -euo pipefail
set -x

GOOS=linux GOARCH=arm GOARM=7 go build
ssh pi@cryptofaxpa.local rm realcrypto/client
scp client pi@cryptofaxpa.local:/home/pi/realcrypto
ssh pi@cryptofaxpa.local sudo killall client || true
ssh pi@cryptofaxpa.local tail -f /var/log/syslog

