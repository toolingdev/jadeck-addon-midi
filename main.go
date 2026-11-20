// MIDI notes: a generic jadeck add-on. Sends one MIDI note per
// slide action, for slide software that listens for MIDI.
package main

import (
	"encoding/json"

	"github.com/toolingdev/jadeck/sdk/addon"
)

// Version is set by the release build.
var Version = "1.0.0"

func main() { addon.Serve(spec()) }

func spec() addon.Spec {
	return addon.Spec{
		Manifest: addon.Manifest{
			ID: "midi", Name: "MIDI notes", Version: Version, Publisher: "jadeck", APIVersion: addon.APIVersion, Repo: "toolingdev/jadeck-addon-midi", Generic: true,
			Description: "Sends a MIDI note per slide action, for software that listens for MIDI.",
			Reaches:     "MIDI outputs on this PC (e.g. a loopMIDI port)",
			ConnectionTypes: []addon.ConnectionType{{
				ID: "midi", Label: "MIDI notes", Capabilities: []string{addon.CapSlides},
				Description: "Install a virtual MIDI port (loopMIDI on Windows), then teach your slide software each note. Windows and Linux.",
				Fields: []addon.Field{
					{Key: "port", Label: "MIDI output", Type: "select", OptionsFrom: "midi.ports", Default: "loopMIDI",
						Help: "Matches the output's name; partial names work."},
					{Key: "channel", Label: "MIDI channel", Type: "number", Default: 1, Min: addon.Num(1), Max: addon.Num(16)},
					{Key: "actions", Label: "Notes", Type: "notes", Keys: []string{"prev", "next", "black", "logo", "clear"},
						Default: map[string]int{"prev": 60, "next": 62, "black": 64, "logo": 65, "clear": 66},
						Help:    "Each button sends its note; teach the same notes in your slide software."},
					{Key: "holdMs", Label: "Note length (ms)", Type: "number", Default: 120, Min: addon.Num(10), Max: addon.Num(1000), Advanced: true},
				},
				Sample:     map[string]any{"port": "loopMIDI Port", "channel": 1},
				SampleTest: "Ready — sends to loopMIDI Port",
			}},
			Widgets: []addon.WidgetDef{{Kind: "core", Type: "slides.prevnext", Label: "Prev / Next", Sample: map[string]any{
				"/api/presentation/status": map[string]any{"available": true, "port": "loopMIDI Port", "actions": []string{"prev", "next", "black", "logo", "clear"}}}}},
		},
		Open: map[string]addon.OpenFunc{"midi": func(raw json.RawMessage) (any, error) {
			var c Config
			if err := json.Unmarshal(raw, &c); err != nil {
				return nil, err
			}
			return New(c, openOut)
		}},
		Options: map[string]func() []addon.Option{"midi.ports": func() []addon.Option {
			var out []addon.Option
			if listPorts != nil {
				for _, n := range listPorts() {
					out = append(out, addon.Option{Value: n, Label: n})
				}
			}
			return out
		}},
	}
}
