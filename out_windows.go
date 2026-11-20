//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

// Windows MIDI out through winmm.dll, called directly (no cgo).

var (
	winmm       = syscall.NewLazyDLL("winmm.dll")
	pGetNumDevs = winmm.NewProc("midiOutGetNumDevs")
	pGetDevCaps = winmm.NewProc("midiOutGetDevCapsW")
	pOpen       = winmm.NewProc("midiOutOpen")
	pShortMsg   = winmm.NewProc("midiOutShortMsg")
	pClose      = winmm.NewProc("midiOutClose")
)

type midiOutCaps struct {
	Mid, Pid      uint16
	DriverVersion uint32
	Pname         [32]uint16
	Technology    uint16
	Voices        uint16
	Notes         uint16
	ChannelMask   uint16
	Support       uint32
}

type winOut struct {
	h    uintptr
	name string
}

func openWinmm(port string) (Out, error) {
	n, _, _ := pGetNumDevs.Call()
	if n == 0 {
		return nil, fmt.Errorf("this PC has no MIDI output devices (is the virtual port, e.g. loopMIDI, installed?)")
	}
	want := strings.ToLower(port)
	for i := uintptr(0); i < n; i++ {
		var caps midiOutCaps
		if r, _, _ := pGetDevCaps.Call(i, uintptr(unsafe.Pointer(&caps)), unsafe.Sizeof(caps)); r != 0 {
			continue
		}
		name := syscall.UTF16ToString(caps.Pname[:])
		if want != "" && !strings.Contains(strings.ToLower(name), want) {
			continue
		}
		var h uintptr
		if r, _, _ := pOpen.Call(uintptr(unsafe.Pointer(&h)), i, 0, 0, 0); r != 0 {
			return nil, fmt.Errorf("could not open MIDI output %q (winmm error %d)", name, r)
		}
		return &winOut{h: h, name: name}, nil
	}
	return nil, fmt.Errorf("no MIDI output matching %q on this PC", port)
}

func (o *winOut) Name() string { return o.name }

func (o *winOut) Send(msg []byte) error {
	var dw uint32
	for i := 0; i < len(msg) && i < 3; i++ {
		dw |= uint32(msg[i]) << (8 * i)
	}
	if r, _, _ := pShortMsg.Call(o.h, uintptr(dw)); r != 0 {
		return fmt.Errorf("midiOutShortMsg failed (winmm error %d)", r)
	}
	return nil
}

func (o *winOut) Close() error {
	pClose.Call(o.h)
	return nil
}
