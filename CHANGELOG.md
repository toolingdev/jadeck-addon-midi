# Changelog

## 1.0.2

- **Fix:** on Windows and Linux every action answered "not supported on
  this OS" because the MIDI output wasn't wired up. Windows (winmm) and
  Linux (ALSA raw MIDI) output work again, and the port picker lists the
  computer's MIDI outputs.
- Built with add-on SDK v0.3.0.

## 1.0.1

- Renamed to jadeck (SDK v0.2.0).

## 1.0.0

- Generic slides: one MIDI note per action.
