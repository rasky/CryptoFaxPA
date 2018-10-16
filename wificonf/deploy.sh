#!/bin/bash
set -euxo pipefail

env GOOS=linux GOARCH=arm GOARM=7 packr build -ldflags="-s -w"
ssh pi@cryptofaxpa.local killall wificonf || true
ssh pi@cryptofaxpa.local rm ./wificonf || true
scp wificonf pi@cryptofaxpa.local:
ssh pi@cryptofaxpa.local sudo ./wificonf -listen 0.0.0.0:8080
