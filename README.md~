# Fluid

A tiny fantasy-console-style game engine. Write games in **Lua**, play them as
**monochrome ASCII art** right in your terminal.

```
  COINS 3/16   DEATHS 1

        ~~~~                                ~~~~
                                 * * *
                      ***      =========
                                                           #
             o                                            ##
             A    * *                        ^^^    W    ###
####################     ###################################
```

> **Status: early prototype.** The API is small and may change.

## Features

- 64×24 character screen, one character per cell
- Lua 5.1 scripting with a PICO-8-style loop: `_init`, `_update`, `_draw`
- Six buttons (arrows, `z`, `x`)
- Smooth key holding on Windows (real key-down / key-up events)
- Single Go binary, no external runtime needed

## Quick start

Requires Go (see `go.mod` for the version).

```
git clone <this repo>
cd <repo>
go run . game.lua
```

On Windows, run it in **Windows Terminal**.

Controls: arrows move, `z` jumps, `q` quits.

## Make a game

```lua
local x, y = 32, 12

function _update()
  if btn(0) then x = x - 1 end
  if btn(1) then x = x + 1 end
end

function _draw()
  cls()
  print("hello, fluid", 1, 0)
  put(x, y, "@")
end
```

See [MANUAL.md](MANUAL.md) for the full API. It has a guide for people and a compact
reference for AI coding agents.

## Project layout

| File | Purpose |
|---|---|
| `main.go` | Lua API, game loop, renderer |
| `term_windows.go` | Windows console setup and key input |
| `term_other.go` | macOS / Linux terminal setup and key input |
| `game.lua` | Example: side-scrolling platformer |
| `MANUAL.md` | API manual |

## Known limitations

- On macOS and Linux, holding a key stutters briefly after the first press, because the
  terminal reports no key-up events.
- The terminal window must be at least 64×24.
- No sound, colors or `btnp` yet.
- Lua scripts currently have access to the full standard library, including `os` and `io`.
  Only run games you trust.

## Roadmap

- Terminal size check
- Sandboxed Lua (no `os` / `io`)
- `btnp()`
- Sound: built-in chip synth (`beep`, `play`)
- Play over SSH
- ESP32 runtime
