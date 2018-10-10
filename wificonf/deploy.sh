#!/bin/bash
set -euxo pipefail

env GOOS=linux GOARCH=arm GOARM=7 packr build -ldflags="-s -w"
ssh pi@cryptofaxpa.local killall wificonf || true
ssh pi@cryptofaxpa.local rm ./realcrypto/wificonf || true
scp wificonf pi@cryptofaxpa.local:realcrypto/
ssh pi@cryptofaxpa.local ./realcrypto/wificonf -listen 0.0.0.0:8080
