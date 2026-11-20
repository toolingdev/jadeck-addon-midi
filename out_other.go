//go:build !windows && !linux

package main

// No MIDI output on this OS yet (macOS needs CoreMIDI via cgo). The
// driver reports this plainly in its status instead of pretending.
