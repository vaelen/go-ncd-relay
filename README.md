# go-ncd-relay
A [Go](https://golang.org/) driver for the professional relay boards from [National Control Devices](https://ncd.io/).

[![Go Report Card](https://goreportcard.com/badge/github.com/vaelen/go-ncd-relay/relay)](https://goreportcard.com/report/github.com/vaelen/go-ncd-relay/relay)
[![Go Docs](https://godoc.org/github.com/vaelen/go-ncd-relay/relay?status.svg)](https://godoc.org/github.com/vaelen/go-ncd-relay/relay)

Copyright (c) 2018, Andrew C. Young

Released under the MIT license

Example usage:
```go
package main

import (
	"log"

	"github.com/jacobsa/go-serial/serial"
	"github.com/vaelen/go-ncd-relay/relay"
)

func main() {

	serialOptions := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the serial port (using go-serial)
	port, err := serial.Open(serialOptions)
	if err != nil {
		log.Fatalf("Couldn't open serial port: %v", err)
	}
	defer port.Close()

	// Create the controller
	controller := relay.New(port)

	// Turn on a relay
	err = controller.TurnOnRelay(1)
	if err != nil {
		log.Fatalf("Couldn't turn on relay: %v", err)
	}

	// Get the current status of the relay
	relayStatus, err := controller.GetRelayStatus(1)
	if err != nil {
		log.Fatalf("Couldn't get relay status: %v", err)
	}
	log.Printf("Relay status: %t\n", relayStatus)

	// Turn off the relay
	err = controller.TurnOffRelay(1)
	if err != nil {
		log.Fatalf("Couldn't turn off relay: %v", err)
	}

	// Get the current status of the relay
	relayStatus, err = controller.GetRelayStatus(1)
	if err != nil {
		log.Fatalf("Couldn't get relay status: %v", err)
	}
	log.Printf("Relay status: %t\n", relayStatus)

	// Turn off the relay
	err = controller.TurnOffRelay(1)
	if err != nil {
		log.Fatalf("Couldn't turn off relay: %v", err)
	}

	// Turn on an entire bank
	err = controller.SetBankStatus(1, 0xFF)
	if err != nil {
		log.Fatalf("Couldn't set bank status: %v", err)
	}

	// Get the current status of the bank
	bankStatus, err := controller.GetBankStatus(1)
	if err != nil {
		log.Fatalf("Couldn't get bank status: %v", err)
	}
	log.Printf("Bank status: %08b\n", bankStatus)

	// Turn off an entire bank
	err = controller.SetBankStatus(1, 0x00)
	if err != nil {
		log.Fatalf("Couldn't set bank status: %v", err)
	}

	// Get the current status of the bank
	bankStatus, err = controller.GetBankStatus(1)
	if err != nil {
		log.Fatalf("Couldn't get bank status: %v", err)
	}
	log.Printf("Bank status: %08b\n", bankStatus)

}
```
