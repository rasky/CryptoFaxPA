#!/bin/bash
set -euo pipefail

env GOOS=linux GOARCH=arm GOARM=7 packr build
ssh pi@cryptofax.local killall wificonf || true
scp wificonf pi@cryptofax.local:
ssh pi@cryptofax.local ./wificonf -listen 0.0.0.0:8080
