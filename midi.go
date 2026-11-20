// Package midi is a generic MIDI-note presentation driver: each action
// sends one tapped note (on, short hold, off) on a MIDI output port.
// Software that listens for MIDI (Proclaim's Settings → MIDI, for
// example) is taught which note means which action.
//
// addon.json:
//
//	{"driver": "midi", "label": "Proclaim", "port": "loopMIDI",
//	 "channel": 1, "actions": {"prev": 60, "next": 62, "black": 64,
//	 "logo": 65, "clear": 66}}
//
// "port" matches an output device name by case-insensitive substring
// (empty = first device). Outputs: Windows (winmm), Linux (ALSA raw
// MIDI, /dev/snd/midiC*D*). macOS is not supported yet.
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/toolingdev/jadeck/sdk/addon"
)

// Out sends raw short MIDI messages.
type Out interface {
	Name() string
	Send(msg []byte) error
	Close() error
}

// openOut and listPorts are provided per OS (out_windows.go, out_linux.go).
var (
	openOut   func(port string) (Out, error)
	listPorts func() []string
)

var DefaultActions = map[string]int{"prev": 60, "next": 62, "black": 64, "logo": 65, "clear": 66}

type Config struct {
	Port     string         `json:"port"`
	Channel  int            `json:"channel"`
	Velocity int            `json:"velocity"`
	HoldMs   int            `json:"holdMs"`
	Actions  map[string]int `json:"actions"`
}

type Driver struct {
	cfg  Config
	open func(string) (Out, error)

	mu      sync.Mutex
	out     Out
	openErr error
	last    map[string]any
}

// New builds the driver. open is the port opener (tests pass a fake).
func New(c Config, open func(string) (Out, error)) (*Driver, error) {
	if c.Channel == 0 {
		c.Channel = 1
	}
	if c.Channel < 1 || c.Channel > 16 {
		return nil, fmt.Errorf("midi channel %d outside 1..16", c.Channel)
	}
	if c.Velocity == 0 {
		c.Velocity = 100
	}
	if c.HoldMs == 0 {
		c.HoldMs = 120
	}
	if len(c.Actions) == 0 {
		c.Actions = DefaultActions
	}
	for a, n := range c.Actions {
		if n < 0 || n > 127 {
			return nil, fmt.Errorf("midi note %d for %q outside 0..127", n, a)
		}
	}
	d := &Driver{cfg: c, open: open}
	d.ensureOpen()
	return d, nil
}

// ensureOpen (re)opens the port; a port that appears later (loopMIDI
// started after us) is picked up on the next action.
func (d *Driver) ensureOpen() (Out, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.out != nil {
		return d.out, nil
	}
	if d.open == nil {
		d.openErr = fmt.Errorf("MIDI output is not supported on this OS yet")
		return nil, d.openErr
	}
	d.out, d.openErr = d.open(d.cfg.Port)
	return d.out, d.openErr
}

func (d *Driver) Info() addon.PresentationInfo {
	out, err := d.ensureOpen()
	d.mu.Lock()
	defer d.mu.Unlock()
	info := addon.PresentationInfo{
		Details: map[string]any{"channel": d.cfg.Channel, "notes": d.cfg.Actions, "lastSent": d.last},
	}
	if err != nil {
		info.Target = err.Error()
		return info
	}
	info.Available = true
	info.Target = out.Name()
	return info
}

func (d *Driver) Actions() []string {
	out := make([]string, 0, len(d.cfg.Actions))
	for a := range d.cfg.Actions {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

func (d *Driver) Do(ctx context.Context, action string) (map[string]any, error) {
	note, ok := d.cfg.Actions[action]
	if !ok {
		return nil, fmt.Errorf("%w %q", addon.ErrUnknownAction, action)
	}
	out, err := d.ensureOpen()
	if err != nil {
		return nil, fmt.Errorf("MIDI port not open: %w", err)
	}
	ch := byte(d.cfg.Channel-1) & 0x0F
	if err := out.Send([]byte{0x90 | ch, byte(note), byte(d.cfg.Velocity)}); err != nil {
		d.drop()
		return nil, err
	}
	select {
	case <-time.After(time.Duration(d.cfg.HoldMs) * time.Millisecond):
	case <-ctx.Done():
	}
	if err := out.Send([]byte{0x80 | ch, byte(note), 0}); err != nil {
		d.drop()
		return nil, err
	}
	detail := map[string]any{"action": action, "note": note, "channel": d.cfg.Channel, "port": out.Name()}
	d.mu.Lock()
	d.last = detail
	d.mu.Unlock()
	return detail, nil
}

// drop closes a failed port so the next action reopens it.
func (d *Driver) drop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.out != nil {
		d.out.Close()
		d.out = nil
	}
}

func (d *Driver) Close() error {
	d.drop()
	return nil
}
