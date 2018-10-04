# CryptoFaxPA 2.0 📠

This repository contains source code for the new model of CryptoFaxPA.

<img src="pics/P1410474.jpg" width="640" />

<img src="pics/P1410484.jpg" width="640" />

<img src="pics/P1410472.jpg" width="640" />

<img src="pics/P1410462.jpg" width="640" />

<img src="pics/P1410470.jpg" width="640" />

## How it works

Plug it to a power socket and enjoy:

* GSM connection
* WiFi connection (in case no known networks are found, CryptoFaxPA will start in access point exposing a page where a WiFi network can be selected)
* a HELP button which instantly prints installation instructions
* a BLOCKCHAIN button which prints super-nerd blockchain information such as real-time Bitcoin value and other juicy things, along with a very pretty chart
* whenever a message is sent to the @CryptoFaxPA bot on Slack, it will be encrypted and sent to the device as a fax (actually, a cryptofax), and instantly printed
* while printing, a nice 56k modem sound is emitted
* images are printed as well, and the Slack bot will actually preprocess them and show them as preview to the sender asking for confirmation - we don't want to send bad-looking images
* in case a fax cannot be delivered to the device or printed successfully, it will be kept in spool

## Bill of materials:

* Raspberry PI 3 B+
* Pipsta thermal printer
* 2 push buttons (and 2 10kOhm resistors)
* PAM8302 audio amplifier
* loudspeakers
* GSM USB modem (such as the Huawei E3531)

## Repository contents:

* [CryptoFaxPA_Leaflet_A3.pdf](CryptoFaxPA_Leaflet_A3.pdf): leaflet
* [CryptoFaxPA.scad](CryptoFaxPA.scad): 3D model
* [backend](backend): contains the bot that reacts to Slack messages and preprocesses them before adding them to a MQTT queue
* [client](client): contains the client that runs on the CryptoFaxPA device
* [common](common): contains common code shared by the other components

## Compilation

Make sure you have Go 1.11 and issue `GOOS=linux GOARCH=arm go build` on your machine. Then `scp` the compiled `client` binary (along with the .ogg files) to the device

## Authors

* Giovanni Bajo (software)
* Laura Bordin (packaging & leaflet)
* Matteo De Santi (product design & packaging)
* Elisabetta Pique' (project manager)
* Alessandro Ranellucci (hardware & software)
