/*
Package co2monitor implements a client for USB CO2 monitors
like the CO2Mini / AIRCO2NTROL MINI / AIRCO2NTROL COACH.

It uses the Linux HIDRAW API to communicate to the USB device.

This is an implementation of the reverse engineering documented at
https://hackaday.io/project/5301-reverse-engineering-a-low-cost-usb-co-monitor
*/
package co2monitor

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// OpCo2 indicates a CO2 ppm value
	OpCo2 = 'P'
	// OpTemp indicates a temperature value
	OpTemp = 'B'
	// OpHum indicates a humidity value
	OpHum = 'A'
)

var (
	key = []byte{0xc4, 0xc6, 0xc0, 0x92, 0x40, 0x23, 0xdc, 0x96}
)

// Conn is a USB connection to the CO2 monitor
type Conn interface {
	// Read reads one operation/value packet from the monitor
	Read() (operation rune, value int, err error)
	// Close closes the USB connection
	Close() (err error)
}

type conn struct {
	device *os.File
}

// Open opens a USB connection to the CO2 monitor using the provided hidraw device (e.g. /dev/hidraw1)
func Open(name string) (c Conn, err error) {
	device, err := os.OpenFile(name, os.O_APPEND|os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}

	// Send "Feature" SET_REPORT w/ encryption key
	setReport := new([9]byte)
	for i, v := range key {
		setReport[i+1] = v
	}
	const hidiocsfeature9 = 0xC0094806

	rawConn, err := device.SyscallConn()
	if err != nil {
		panic(err)
	}
	var ep syscall.Errno
	rawConn.Control(func(fd uintptr) {
		_, _, ep = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(hidiocsfeature9), uintptr(unsafe.Pointer(setReport)))
	})
	if ep != 0 {
		device.Close()
		return nil, syscall.Errno(ep)
	}

	return &conn{
		device: device,
	}, nil
}

func (c *conn) Close() (err error) {
	return c.device.Close()
}

func (c *conn) Read() (operation rune, value int, err error) {
	buffer := make([]byte, 8)
	_, err = c.device.Read(buffer)
	if err != nil {
		return
	}

	data := buffer
	if !check(data) {
		data = decrypt(data)
		if !check(data) {
			err = fmt.Errorf("checksum error: %x", data)
		}
	}

	operation = rune(data[0])
	value = ((int)(data[1]) << 8) | (int)(data[2])
	return
}

// TempToCelsius converts a temperature value returned by Read() to degrees celsius
func TempToCelsius(value int) (temperature float64) {
	return float64(value)/16.0 - 273.15

}

// HumidityToRH converts a humidity value returned by Read() to %RH
func HumidityToRH(value int) (humidity float64) {
	return float64(value) / 100.0
}

func check(data []byte) bool {
	return (data[0]+data[1]+data[2])&0xff == data[3] &&
		data[4] == 0x0d && data[5] == 0x00 && data[6] == 0x00 && data[7] == 0x00
}

func decrypt(data []byte) []byte {
	cstate := []byte{0x48, 0x74, 0x65, 0x6d, 0x70, 0x39, 0x39, 0x65}
	shuffle := []int{2, 4, 0, 7, 1, 6, 5, 3}

	phase1 := make([]byte, 8)
	for i, o := range shuffle {
		phase1[o] = data[i]
	}

	phase2 := make([]byte, 8)
	for i := 0; i < 8; i++ {
		phase2[i] = phase1[i] ^ key[i]
	}

	phase3 := make([]byte, 8)
	for i := 0; i < 8; i++ {
		phase3[i] = ((phase2[i] >> 3) | (phase2[(i-1+8)%8] << 5)) & 0xff
	}

	ctmp := make([]byte, 8)
	for i := 0; i < 8; i++ {
		ctmp[i] = ((cstate[i] >> 4) | (cstate[i] << 4)) & 0xff
	}

	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		out[i] = (byte)(((0x100 + (int)(phase3[i]) - (int)(ctmp[i])) & (int)(0xff)))
	}
	return out
}
