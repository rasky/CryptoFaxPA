#!/bin/bash
echo "Bringing down existing interfaces..."
sudo ifdown wlan0
sudo ifdown ap0

echo "Starting Wi-Fi AP..."
sudo ifup ap0
sleep 2

echo "Starting Wi-Fi client..."
sudo ifup wlan0
