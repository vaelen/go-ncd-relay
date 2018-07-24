/***

MIT License

Copyright (c) 2018 Andrew C. Young

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

***/

package relay_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/vaelen/go-ncd-relay/relay"
)

func TestPacketChecksum(t *testing.T) {
	expectedChecksum := byte(0xDC)
	input := []byte{0xAA, 0x04, 0xFE, 0x30, 0x00, 0x00}
	checksum := relay.Checksum(input)
	if checksum != expectedChecksum {
		t.Errorf("Checksum not correct. Expected: %x, Got: %x", expectedChecksum, checksum)
	}
}

func TestPacket(t *testing.T) {
	expectedPacket := []byte{0xAA, 0x04, 0xFE, 0x30, 0x00, 0x00, 0xDC}
	payload := []byte{254, 48, 0, 0}
	packet := relay.CreatePacket(payload)
	if !bytes.Equal(packet, expectedPacket) {
		t.Errorf("Packet not created properly. Expected: %x, Got: %x", expectedPacket, packet)
	}
}

func ExampleController() {

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Turn on a relay
	err = controller.TurnOnRelay(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't turn on relay: %v", err)
	}

	// Get the current status of the relay
	relayStatus, err := controller.GetRelayStatus(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't get relay status: %v", err)
	}
	log.Printf("Relay status: %t\n", relayStatus)

	// Turn off the relay
	err = controller.TurnOffRelay(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't turn off relay: %v", err)
	}

	// Get the current status of the relay
	relayStatus, err = controller.GetRelayStatus(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't get relay status: %v", err)
	}
	log.Printf("Relay status: %t\n", relayStatus)

	// Turn off the relay
	err = controller.TurnOffRelay(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't turn off relay: %v", err)
	}

	// Turn on an entire bank
	err = controller.SetBankStatus(ctx, 1, 0xFF)
	if err != nil {
		log.Fatalf("Couldn't set bank status: %v", err)
	}

	// Get the current status of the bank
	bankStatus, err := controller.GetBankStatus(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't get bank status: %v", err)
	}
	log.Printf("Bank status: %08b\n", bankStatus)

	// Turn off an entire bank
	err = controller.SetBankStatus(ctx, 1, 0x00)
	if err != nil {
		log.Fatalf("Couldn't set bank status: %v", err)
	}

	// Get the current status of the bank
	bankStatus, err = controller.GetBankStatus(ctx, 1)
	if err != nil {
		log.Fatalf("Couldn't get bank status: %v", err)
	}
	log.Printf("Bank status: %08b\n", bankStatus)

}
