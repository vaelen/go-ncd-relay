/*
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
*/

// This example uses the ADC inputs on the controller as a voltmeter.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/vaelen/go-ncd-relay/relay"
)

func read(controller *relay.Controller) []uint16 {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	allAD10, err := controller.ReadAllAD10(ctx)
	if err != nil {
		log.Fatalf("Couldn't read all ADs in 10bit mode: %v", err)
	}
	return allAD10
}

const V float32 = 5.0 / 1024.0

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

	for {
		allAD10 := read(controller)
		allV := make([]string, len(allAD10))
		for n, i := range allAD10 {
			allV[n] = fmt.Sprintf("[%02d, %04d, %01.3fV]", n, i, float32(i) * V)
		}
		log.Println(allV)
		time.Sleep(time.Millisecond * 10)
	}

}
