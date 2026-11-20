//go:build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Linux MIDI out through ALSA raw MIDI device files (no cgo). "port" is
// either a device path (/dev/snd/midiC1D0) or a substring of a sound
// card name from /proc/asound/cards; empty = first raw MIDI device.

type fileOut struct {
	f    *os.File
	name string
}

func openALSA(port string) (Out, error) {
	path, name := "", ""
	switch {
	case strings.HasPrefix(port, "/dev/"):
		path, name = port, port
	default:
		devs, _ := filepath.Glob("/dev/snd/midiC*D*")
		if len(devs) == 0 {
			return nil, fmt.Errorf("no ALSA raw MIDI devices (/dev/snd/midiC*D*); load snd-virmidi for a virtual port")
		}
		cards := cardNames()
		want := strings.ToLower(port)
		for _, d := range devs {
			var c, dev int
			fmt.Sscanf(filepath.Base(d), "midiC%dD%d", &c, &dev)
			n := cards[c]
			if want == "" || strings.Contains(strings.ToLower(n), want) {
				path, name = d, fmt.Sprintf("%s (%s)", n, d)
				break
			}
		}
		if path == "" {
			return nil, fmt.Errorf("no MIDI output matching %q", port)
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return &fileOut{f: f, name: name}, nil
}

func cardNames() map[int]string {
	out := map[int]string{}
	f, err := os.Open("/proc/asound/cards")
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var n int
		line := sc.Text()
		if _, err := fmt.Sscanf(strings.TrimSpace(line), "%d", &n); err == nil {
			if i := strings.Index(line, "]: "); i >= 0 {
				out[n] = strings.TrimSpace(line[i+3:])
			}
		}
	}
	return out
}

func (o *fileOut) Name() string          { return o.name }
func (o *fileOut) Send(msg []byte) error { _, err := o.f.Write(msg); return err }
func (o *fileOut) Close() error          { return o.f.Close() }
