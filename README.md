# CryptoFaxPA 2.0

This repository contains source code for the new model of CryptoFaxPA.

## Bill of materials:

* Raspberry PI 3 B+
* Pipsta thermal printer
* 2 push buttons (and 2 10kOhm resistors)

## Repository contents:

* `cryptofax.scad`: 3D model
* `backend`: contains the bot that reacts to Slack messages and preprocesses them before adding them to a MQTT queue
* `client`: contains the client that runs on the CryptoFaxPA device
* `common`: contains common code shared by the other components

## Compilation

Make sure you have Go 1.11 and issue `GOOS=linux GOARCH=arm go build` on your machine. Then `scp` the compiled `client` binary (along with the .ogg files) to the device

## Authors

* Giovanni Bajo (software)
* Laura Bordin (packaging)
* Matteo De Santi (product design & packaging)
* Elisabetta Pique' (project manager)
* Alessandro Ranellucci (hardware & software)
