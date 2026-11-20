package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/toolingdev/jadeck/sdk/addon"
)

type fakeOut struct{ sent [][]byte }

func (f *fakeOut) Name() string { return "Fake MIDI" }
func (f *fakeOut) Send(m []byte) error {
	f.sent = append(f.sent, append([]byte(nil), m...))
	return nil
}
func (f *fakeOut) Close() error { return nil }

func TestNoteTap(t *testing.T) {
	fo := &fakeOut{}
	d, err := New(Config{Channel: 2, HoldMs: 1}, func(string) (Out, error) { return fo, nil })
	if err != nil {
		t.Fatal(err)
	}
	det, err := d.Do(context.Background(), "next")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fo.sent[0], []byte{0x91, 62, 100}) || !bytes.Equal(fo.sent[1], []byte{0x81, 62, 0}) {
		t.Fatalf("% x", fo.sent)
	}
	if det["note"] != 62 || !d.Info().Available {
		t.Fatal(det)
	}
	if _, err := d.Do(context.Background(), "explode"); !errors.Is(err, addon.ErrUnknownAction) {
		t.Fatal(err)
	}
}

func TestPortMissingFailsLoudlyThenRecovers(t *testing.T) {
	var fo *fakeOut
	open := func(string) (Out, error) {
		if fo == nil {
			return nil, errors.New("no MIDI output matching \"loopMIDI\"")
		}
		return fo, nil
	}
	d, _ := New(Config{Port: "loopMIDI", HoldMs: 1}, open)
	if d.Info().Available {
		t.Fatal("claims available")
	}
	if _, err := d.Do(context.Background(), "next"); err == nil {
		t.Fatal("sending with no port must fail, not pretend")
	}
	fo = &fakeOut{} // loopMIDI started
	if _, err := d.Do(context.Background(), "next"); err != nil {
		t.Fatalf("did not pick up the late port: %v", err)
	}
}

func TestBadConfig(t *testing.T) {
	if _, err := New(Config{Channel: 17}, nil); err == nil {
		t.Fatal("channel 17")
	}
	if _, err := New(Config{Actions: map[string]int{"x": 200}}, nil); err == nil {
		t.Fatal("note 200")
	}
}
