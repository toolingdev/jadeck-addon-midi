# MIDI notes — PTZ controller add-on

Sends a MIDI note per slide action, for software that listens for MIDI.

Install: it ships with the controller (Settings → Add-ons); newer releases of this repository arrive as updates there.

Build every platform and the `.ptzaddon` package:

```bash
go run github.com/toolingdev/ptz-controller/sdk/addon/cmd/ptz-addon build
```

Try it on a controller: Settings → Add-ons → Developer → Load a folder (this folder),
or install `dist/midi.ptzaddon` with Install from file.

Release: push a tag `v1.2.3`; the workflow builds, signs (secret `PTZ_SIGNING_KEY`)
and attaches `midi.ptzaddon` to the GitHub release.
