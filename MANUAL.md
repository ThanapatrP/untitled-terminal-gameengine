# Fluid Manual

Fluid is a tiny fantasy-console-style game engine. Games are written in Lua and drawn as
monochrome ASCII art directly in your terminal.

This manual has two parts:

- **Part 1: Guide**: for people learning to make games with Fluid.
- **Part 2: Agent Reference**: a compact, exact spec for AI coding agents.

Both parts describe the same engine version (prototype, Lua API v0).

---

# Part 1: Guide

## Running a game

```
go run . game.lua
```

If you leave out the file name, Fluid loads `game.lua` from the current folder.

Run it in a real terminal. On Windows use **Windows Terminal**, not an IDE output pane.

| Key | Action |
|---|---|
| Arrow keys | `btn(0)`–`btn(3)` |
| `z` | `btn(4)` |
| `x` | `btn(5)` |
| `q`, `Ctrl+C` (and `Esc` on Windows) | Quit |

## The screen

The screen is a grid of **64 columns × 24 rows** of characters. There are no pixels and no
colors: every cell holds exactly one character.

```
(0,0) ──────────── x ────────────► (63,0)
  │
  y
  │
  ▼
(0,23)                            (63,23)
```

- `x` is the column (0–63) and `y` is the row (0–23). Both start at 0 in the top-left corner.
- Positions can be decimals such as `10.6`. Fluid rounds them **down** (`10.6` → `10`), so you
  can keep smooth positions in your game logic and pass them straight to the draw functions.
- Anything drawn outside the screen is silently cut off. This is safe and not an error.

> **Cells are tall.** A terminal character is about twice as tall as it is wide, so moving
> 1 row looks like moving about 2 columns. To make movement look equally fast in both
> directions, move vertically at about half the speed you move horizontally.

## The game loop

Define up to three functions and Fluid calls them for you:

```lua
function _init()
  -- runs once, when the game starts
end

function _update()
  -- runs 30 times per second: read buttons, move things
end

function _draw()
  -- runs 30 times per second, right after _update: draw the screen
end
```

All three are optional. The screen is **not** cleared for you between frames, so most games
start `_draw()` with `cls()`.

## Drawing

### `cls()`
Clears the whole screen to spaces.

### `put(x, y, ch)`
Draws one character. If `ch` is longer than one character, only the first is used.

```lua
put(10, 5, "@")
```

### `print(text, x, y)`
Draws text starting at column `x`, row `y`. A `"\n"` in the text moves to the next row, back
at column `x`. Spaces in the text are drawn, so they cover whatever was underneath.

```lua
print("SCORE " .. score, 1, 0)
print("line one\nline two", 10, 10)
```

> Fluid replaces Lua's built-in `print`. It always needs `x` and `y`, so you cannot use it to
> write debug messages to the console.

### `art(x, y, text)`
Draws multi-line ASCII art with its top-left corner at `(x, y)`. **Spaces are transparent**:
they leave what is underneath visible, so a character can stand in front of the background.

Lua's long strings `[[ ... ]]` are the easiest way to write art. Lua drops the first line
break right after `[[`, so the art starts on the first row:

```lua
local hero = [[
 o
/|\
/ \]]

art(30, 10, hero)
```

### `fill(x0, y0, x1, y1, ch)`
Fills a rectangle with one character. Both corners are included, and they can be given in
any order.

```lua
fill(0, 21, 63, 23, "#")   -- ground
fill(14, 8, 49, 13, " ")   -- clear a box for a message
```

## Input

### `btn(i)`
Returns `true` while button `i` is held down.

| `i` | Key |
|---|---|
| 0 | Left |
| 1 | Right |
| 2 | Up |
| 3 | Down |
| 4 | `z` |
| 5 | `x` |

A very short tap still counts as pressed for at least one frame, so a quick tap is never
missed. Any other number returns `false`.

There is no "just pressed" function yet. If you need one, remember last frame's value:

```lua
local was_down = false
function _update()
  local down = btn(4)
  if down and not was_down then
    -- pressed this frame
  end
  was_down = down
end
```

> **Platform note.** On Windows, Fluid reads real key-down and key-up events, so holding a
> key is smooth. On macOS and Linux the terminal only reports typed characters, so holding a
> key relies on the keyboard's auto-repeat. There will be a short pause after the first
> press, the same pause you see when holding a key in a text editor.

## Lua version

Fluid uses **Lua 5.1**. Some features of newer Lua versions are not available:

| Don't use (Lua 5.3+) | Use instead |
|---|---|
| `a // b` | `math.floor(a / b)` |
| `a & b`, `a \| b`, `a << n` | arithmetic, or avoid |
| `math.tointeger(x)` | `math.floor(x)` |
| `table.unpack(t)` | `unpack(t)` |

The `math`, `string` and `table` libraries work as usual. Use `math.random(a, b)` for
random numbers.

## Characters you can use

Use characters that take up **one column**: ASCII is always safe, and so are most simple
symbols such as `·`. Wide characters (Chinese, Japanese, Korean, emoji) and Thai vowel or
tone marks that sit above or below a letter will push the rest of the row out of line.
Control characters such as tabs are shown as spaces.

## Errors

- A **syntax error** prints the file and line, and the game does not start.
- A **runtime error** (for example calling `nil`) stops the game, restores your terminal
  and prints the error with its file and line.

## A complete small game

```lua
local x, y = 30, 12

function _update()
  if btn(0) then x = x - 1 end
  if btn(1) then x = x + 1 end
  if btn(2) then y = y - 0.5 end   -- vertical is half speed (cells are tall)
  if btn(3) then y = y + 0.5 end
  x = math.max(0, math.min(63, x))
  y = math.max(1, math.min(23, y))
end

function _draw()
  cls()
  print("arrows to move, q to quit", 1, 0)
  put(x, y, "@")
end
```

For a bigger example, see the side-scrolling platformer in `game.lua`. It keeps its level
as ASCII text, so you can edit the level by editing the text.

## Not available yet

These do not exist yet. Calling them gives a runtime error:
colors, sound (`beep`, `play`), "just pressed" (`btnp`), mouse, loading files or sprites,
changing the screen size or frame rate.

---

# Part 2: Agent Reference

Normative spec for generating or editing Fluid games. If anything here disagrees with
Part 1, follow this part.

## Runtime facts

- Language: **Lua 5.1** (gopher-lua). No `//`, no bitwise operators, no integer subtype,
  no `table.unpack` (use `unpack`), no `math.tointeger`. Use `math.floor` for integer division.
- Entry: `go run . <file.lua>`; the default file is `game.lua`. The file runs top to bottom
  once at load, before the terminal is set up.
- Screen: `W = 64` columns, `H = 24` rows, monochrome, one character per cell, no colors.
- Coordinates: 0-based; `x` = column, `y` = row; origin top-left. Every numeric argument is
  passed through `math.floor`. Writes outside `0 <= x < 64, 0 <= y < 24` are ignored without
  error.
- Frame rate: fixed **30 fps**. Order per frame: `_update()`, then `_draw()`, then the input
  latch is cleared, then the screen is rendered. The screen buffer persists between frames
  and is never cleared automatically.
- Callbacks (all optional globals): `_init()` runs once before the first frame; `_update()`
  and `_draw()` run every frame.
- Cell aspect ratio is about 1:2 (width:height). For visually equal speed, use vertical
  velocity of about 0.5 × horizontal velocity.
- Reserved keys: `q` and `Ctrl+C` quit (plus `Esc` on Windows). Games cannot read them.
- Any Lua error ends the program and prints `file:line: message`.

## API

| Signature | Returns | Semantics |
|---|---|---|
| `cls()` | nothing | Set every cell to `" "`. Arguments are ignored. |
| `put(x, y, ch)` | nothing | Set one cell to the first UTF-8 character of `ch`. `""` gives `" "`. Spaces are drawn. |
| `print(text, x, y)` | nothing | Draw `text` from `(x, y)`. `"\n"` sets x back to the start x and adds 1 to y. Spaces are drawn (opaque). No wrapping: text past column 63 is clipped. **x and y are required. This replaces Lua's `print`.** |
| `art(x, y, text)` | nothing | Split `text` on `"\n"`; draw line `i` at row `y+i`. The space character `" "` is transparent (skipped); every other character is drawn. |
| `fill(x0, y0, x1, y1, ch)` | nothing | Set every cell in the inclusive rectangle to the first character of `ch`. Corner order does not matter. |
| `btn(i)` | boolean | `0` left, `1` right, `2` up, `3` down, `4` z, `5` x. True if held now **or** pressed at any time since the previous frame. Any other `i` gives `false`. |

`x`, `y`, `x0`, `y0`, `x1`, `y1` and `i` are numbers; `ch` and `text` are strings.
Passing `nil` or a table where a number or string is expected raises a Lua error.

### Input behaviour by platform
- Windows: real key-down and key-up events. Holding a key is continuous.
- macOS and Linux: no key-up events. A key counts as held for 150 ms after each character
  arrives, so expect a gap of about 300–500 ms (the OS repeat delay) after the first press.
  Do not design mechanics that depend on exact hold timing.

## Constraints and rules

1. Do **not** call functions that are not in the API table. In particular these do not
   exist: `spr`, `rect`, `line`, `circ`, `pset`, `color`, `btnp`, `sfx`, `music`, `beep`,
   `play`, `rnd`, `flr`, `mid`, `time`, `stat`. Use `math.random`, `math.floor`,
   `math.max`/`math.min` instead.
2. Use only single-width characters, preferably printable ASCII (0x20–0x7E). Do not use
   wide characters, emoji or combining marks in drawn text. Characters below 0x20 are
   rendered as spaces.
3. Start `_draw()` with `cls()` unless you intentionally redraw every cell.
4. Keep all game state in `local` variables at file scope; draw only in `_draw()`.
5. Do not use `print` for debugging. It draws on the screen and requires `x, y`.
6. Keep per-frame velocities below 1 cell when doing tile collision, or split the movement
   into sub-steps, to avoid passing through walls.
7. Use `[[ ... ]]` for multi-line art and level maps. The line break right after `[[` is
   dropped by Lua.
8. Rows are about 2× taller than columns. Scale vertical motion accordingly.
9. Reserve row 0 (or 23) for a HUD when needed; the game area is whatever rows remain.

## Template

```lua
-- state
local x, y = 32, 12

function _init()
end

function _update()
  if btn(0) then x = x - 1 end
  if btn(1) then x = x + 1 end
  if btn(2) then y = y - 0.5 end
  if btn(3) then y = y + 0.5 end
  x = math.max(0, math.min(63, x))
  y = math.max(1, math.min(23, y))
end

function _draw()
  cls()
  print("HUD", 0, 0)
  put(x, y, "@")
end
```

## Recommended patterns

- **Tile maps:** keep levels as an array of equal-width strings (pad shorter rows with
  spaces). Read a tile with `row:sub(x + 1, x + 1)`. Change a tile with
  `row:sub(1, x) .. ch .. row:sub(x + 2)`.
- **Scrolling:** keep a camera offset `cam`; draw the visible slice with
  `print(row:sub(cam + 1, cam + 64), 0, screen_y)`, and draw entities at `world_x - cam`.
- **Collision:** treat entities as boxes in cell units. Test every cell the box overlaps
  with `math.floor(pos)` to `math.floor(pos + size - 0.001)`. Resolve the X axis and the
  Y axis separately.
- **Edge-triggered input:** compare `btn(i)` against the previous frame's value stored in a
  local variable.
- **Reference implementation:** `game.lua` (side-scrolling platformer) uses all of the
  patterns above.
