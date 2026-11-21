//go:build windows || linux

package main

import "testing"

// The OS's MIDI output must be wired up, or every action says "not
// supported on this OS".
func TestOutputWired(t *testing.T) {
	if openOut == nil || listPorts == nil {
		t.Fatal("no MIDI output for this OS")
	}
}
