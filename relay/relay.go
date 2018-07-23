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

// Package relay includes code for interacting with the relay controllers produced by National Control Devices (ncd.io).
package relay

import (
	"fmt"
	"io"
)

// ErrInvalidResponse is returned when an invalid response was received from the relay controller
var ErrInvalidResponse = fmt.Errorf("invalid response")

//////////////////////
///// Controller /////
//////////////////////

// A Controller represents a relay controller
type Controller struct {
	stream io.ReadWriter
}

// New creates a new instance of a relay controller.
// NCD relay controllers can communicate via a wide range of technologies.
// Because of this, the controller expects the caller to create and manage
// the stream that is used for communication with the relay controller.
func New(stream io.ReadWriter) *Controller {
	c := &Controller{
		stream: stream,
	}
	return c
}

// TurnOnRelay turns on the given relay using 1 based indexing
func (c *Controller) TurnOnRelay(index uint16) error {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 48, lsb, msb})
	return c.ExecuteCommand(packet)
}

// TurnOffRelay turns off the given relay using 1 based indexing
func (c *Controller) TurnOffRelay(index uint16) error {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 47, lsb, msb})
	return c.ExecuteCommand(packet)
}

// GetRelayStatus returns the current state of the given relay using 1 based indexing
func (c *Controller) GetRelayStatus(index uint16) (bool, error) {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 44, lsb, msb})
	payload, err := c.ExecuteRead(packet, 1)
	if err != nil {
		return false, err
	}
	return payload[0] == 1, nil
}

// SetBankStatus sets the current state of all 8 relays in a given bank at the same time
func (c *Controller) SetBankStatus(bank uint8, status uint8) error {
	packet := CreatePacket([]byte{254, 140, status, bank})
	return c.ExecuteCommand(packet)
}

// GetBankStatus returns the current state of all 8 relays in a given bank
func (c *Controller) GetBankStatus(bank uint8) (uint8, error) {
	packet := CreatePacket([]byte{254, 124, bank})
	payload, err := c.ExecuteRead(packet, 1)
	if err != nil {
		return 0, err
	}
	return payload[0], nil
}

// TurnOnRelayByBank turns on the given relay in the given bank using 1 based indexing
func (c *Controller) TurnOnRelayByBank(index uint8, bank uint8) error {
	packet := CreatePacket([]byte{254, 48, 107 + index, bank})
	return c.ExecuteCommand(packet)
}

// TurnOffRelayByBank turns off the given relay in the given bank using 1 based indexing
func (c *Controller) TurnOffRelayByBank(index uint8, bank uint8) error {
	packet := CreatePacket([]byte{254, 99 + index, bank})
	return c.ExecuteCommand(packet)
}

// ExecuteCommand executes a command that does not return data
func (c *Controller) ExecuteCommand(packet Packet) error {
	response, err := c.sendCommand(packet, 4)
	if err != nil {
		return err
	}
	if !response.IsValid() {
		return ErrInvalidResponse
	}
	return nil
}

// ExecuteRead executes a command that returns data
func (c *Controller) ExecuteRead(packet Packet, responseLength int) ([]byte, error) {
	response, err := c.sendCommand(packet, responseLength+3)
	if err != nil {
		return nil, err
	}
	if !response.IsValid() {
		return nil, ErrInvalidResponse
	}
	return response.Payload(), nil
}

func (c *Controller) sendCommand(packet Packet, responseLength int) (Packet, error) {
	var bytesWritten int
	var err error
	bytesToWrite := packet

	for len(bytesToWrite) > 0 {
		bytesWritten, err = c.stream.Write(bytesToWrite)
		if err != nil {
			return nil, err
		}
		bytesToWrite = bytesToWrite[bytesWritten:]
	}

	var bytesRead int
	var totalBytesRead int
	response := make([]byte, responseLength)

	for totalBytesRead < responseLength {
		bytesRead, err = c.stream.Read(response[totalBytesRead:])
		if err != nil {
			return nil, err
		}
		totalBytesRead += bytesRead
	}

	return response, nil
}

//////////////////
///// Packet /////
//////////////////

// A Packet represents a series of bytes that make up a control packet
type Packet []byte

// CreatePacket creates a packet from the given payload
func CreatePacket(payload []byte) Packet {
	packet := make([]byte, 0, len(payload)+3)
	packet = append(packet, 170, byte(len(payload)))
	packet = append(packet, payload...)
	packet = append(packet, Checksum(packet))
	return packet
}

// IsValid returns true if the packet is valid
func (packet Packet) IsValid() bool {
	return packet.validHandshake() && packet.validLength() && packet.validChecksum()
}

// Payload returns the packet's payload
func (packet Packet) Payload() []byte {
	return packet[2 : len(packet)-1]
}

func (packet Packet) validHandshake() bool {
	return packet[0] == 170
}

func (packet Packet) validLength() bool {
	return packet[1] == byte(len(packet)-3)
}

func (packet Packet) validChecksum() bool {
	return packet[len(packet)-1] == Checksum(packet[0:len(packet)-1])
}

////////////////
///// Misc /////
////////////////

// Checksum generates a packet checksum
func Checksum(packet []byte) byte {
	var chk byte
	for _, b := range packet {
		chk += b
	}
	return chk
}