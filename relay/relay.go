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
	"context"
	"fmt"
	"io"
)

// ErrInvalidResponse is returned when an invalid response was received from the relay controller
var ErrInvalidResponse = fmt.Errorf("invalid response")

// ErrTimedOut is returned when a command to the relay controller times out
var ErrTimedOut = fmt.Errorf("timed out waiting for response")

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
func (c *Controller) TurnOnRelay(ctx context.Context, index uint16) error {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 48, lsb, msb})
	return c.ExecuteCommand(ctx, packet)
}

// TurnOffRelay turns off the given relay using 1 based indexing
func (c *Controller) TurnOffRelay(ctx context.Context, index uint16) error {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 47, lsb, msb})
	return c.ExecuteCommand(ctx, packet)
}

// GetRelayStatus returns the current state of the given relay using 1 based indexing
func (c *Controller) GetRelayStatus(ctx context.Context, index uint16) (bool, error) {
	lsb := byte(index - 1)
	msb := byte(index >> 8)
	packet := CreatePacket([]byte{254, 44, lsb, msb})
	payload, err := c.ExecuteRead(ctx, packet, 1)
	if err != nil {
		return false, err
	}
	return payload[0] == 1, nil
}

// SetBankStatus sets the current state of all 8 relays in a given bank at the same time
func (c *Controller) SetBankStatus(ctx context.Context, bank uint8, status uint8) error {
	packet := CreatePacket([]byte{254, 140, status, bank})
	return c.ExecuteCommand(ctx, packet)
}

// GetBankStatus returns the current state of all 8 relays in a given bank
func (c *Controller) GetBankStatus(ctx context.Context, bank uint8) (uint8, error) {
	packet := CreatePacket([]byte{254, 124, bank})
	payload, err := c.ExecuteRead(ctx, packet, 1)
	if err != nil {
		return 0, err
	}
	return payload[0], nil
}

// TurnOnRelayByBank turns on the given relay in the given bank using 1 based indexing
func (c *Controller) TurnOnRelayByBank(ctx context.Context, index uint8, bank uint8) error {
	packet := CreatePacket([]byte{254, 48, 107 + index, bank})
	return c.ExecuteCommand(ctx, packet)
}

// TurnOffRelayByBank turns off the given relay in the given bank using 1 based indexing
func (c *Controller) TurnOffRelayByBank(ctx context.Context, index uint8, bank uint8) error {
	packet := CreatePacket([]byte{254, 99 + index, bank})
	return c.ExecuteCommand(ctx, packet)
}

// ReadAD8 reads one of the AD channels with 8 bit granularity (0-255)
func (c *Controller) ReadAD8(ctx context.Context, channel uint8) (uint8, error) {
	packet := CreatePacket([]byte{254, 149 + channel})
	v, err := c.ExecuteRead(ctx, packet, 1)
	if err != nil || len(v) < 1 {
		return 0, err
	}
	return v[0], err
}

// ReadAllAD8 reads all of the AD channels with 8 bit granularity (0-255)
func (c *Controller) ReadAllAD8(ctx context.Context) ([]uint8, error) {
	packet := CreatePacket([]byte{254, 166})
	return c.ExecuteRead(ctx, packet, 8)
}

// ReadAD10 reads one of the AD channels with 10 bit granularity (0-1024)
func (c *Controller) ReadAD10(ctx context.Context, channel uint8) (uint16, error) {
	packet := CreatePacket([]byte{254, 149 + channel})
	v, err := c.ExecuteRead(ctx, packet, 1)
	if err != nil || len(v) < 2 {
		return 0, err
	}
	return parse10Bit(v), nil
}

// ReadAllAD10 reads all of the AD channels with 10 bit granularity (0-1024)
func (c *Controller) ReadAllAD10(ctx context.Context) ([]uint16, error) {
	packet := CreatePacket([]byte{254, 166})
	v, err := c.ExecuteRead(ctx, packet, 16)
	if err != nil || len(v) < 16 {
		return nil, err
	}
	r := make([]uint16, 8)
	for i := 0; i < 16; i += 2 {
		r[i/2] = parse10Bit(v[i : i+2])
	}
	return r, nil
}

func parse10Bit(b []byte) uint16 {
	return ((uint16(b[0]) & 3) << 8) + uint16(b[1])
}

// ExecuteCommand executes a command that does not return data
func (c *Controller) ExecuteCommand(ctx context.Context, packet Packet) error {
	response, err := c.sendCommand(ctx, packet, 4)
	if err != nil {
		return err
	}
	if !response.IsValid() {
		return ErrInvalidResponse
	}
	return nil
}

// ExecuteRead executes a command that returns data
func (c *Controller) ExecuteRead(ctx context.Context, packet Packet, responseLength int) ([]byte, error) {
	response, err := c.sendCommand(ctx, packet, responseLength+3)
	if err != nil {
		return nil, err
	}
	if !response.IsValid() {
		return nil, ErrInvalidResponse
	}
	return response.Payload(), nil
}

func (c *Controller) sendCommand(ctx context.Context, packet Packet, responseLength int) (response Packet, err error) {
	response = make([]byte, responseLength)
	done := make(chan struct{})

	go func() {
		defer func() {
			close(done)
		}()
		var bytesWritten int
		bytesToWrite := packet

		for len(bytesToWrite) > 0 {
			bytesWritten, err = c.stream.Write(bytesToWrite)
			if err != nil {
				return
			}
			bytesToWrite = bytesToWrite[bytesWritten:]
		}

		var bytesRead int
		var totalBytesRead int

		for totalBytesRead < responseLength {
			bytesRead, err = c.stream.Read(response[totalBytesRead:])
			if err != nil {
				return
			}
			totalBytesRead += bytesRead
		}
	}()

	select {
	case <-done:
		// Finished
	case <-ctx.Done():
		// Timed out
		err = ErrTimedOut
	}

	return response, err
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
