// Fluid prototype: Go runtime + embedded Lua + monochrome ASCII-art terminal display.
//
// Screen is a grid of W x H character cells, one character per cell, no colors.
//
// Run:   go run . game.lua
// Keys:  arrows = btn(0..3), z = btn(4), x = btn(5), q / Esc / Ctrl-C = quit
//
// Platform code:
//
//	term_windows.go  console setup + real key-down/key-up input (ReadConsoleInput)
//	term_other.go    macOS/Linux raw mode + byte-based input (key-repeat fallback)
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	W   = 64 // columns
	H   = 24 // rows
	FPS = 30
)

type screen [H][W]rune

// ---------- input ----------

type input struct {
	mu    sync.Mutex
	down  [6]bool      // real key state (Windows: from key-down / key-up events)
	until [6]time.Time // fallback for terminals without key-up: held until this time
	latch [6]bool      // pressed at least once since last frame, so quick taps are never missed
	quit  bool
}

// setDown is used when the platform reports real key-down / key-up.
func (in *input) setDown(b int, d bool) {
	in.mu.Lock()
	in.down[b] = d
	if d {
		in.latch[b] = true
	}
	in.mu.Unlock()
}

// tap is used when the platform only reports "a key character arrived".
func (in *input) tap(b int) {
	in.mu.Lock()
	in.until[b] = time.Now().Add(150 * time.Millisecond)
	in.latch[b] = true
	in.mu.Unlock()
}

func (in *input) held(b int) bool {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.down[b] || in.latch[b] || time.Now().Before(in.until[b])
}

func (in *input) endFrame() {
	in.mu.Lock()
	in.latch = [6]bool{}
	in.mu.Unlock()
}

func (in *input) setQuit() { in.mu.Lock(); in.quit = true; in.mu.Unlock() }

func (in *input) quitting() bool { in.mu.Lock(); defer in.mu.Unlock(); return in.quit }

// ---------- Lua API ----------
//
// cls()                              clear screen
// put(x, y, ch)                      one character
// print(text, x, y)                  text; "\n" goes to the next line, back at column x
// art(x, y, text)                    multi-line ASCII art; spaces are transparent
// fill(x0, y0, x1, y1, ch)           fill a rectangle with one character
// btn(i) -> bool                     0=left 1=right 2=up 3=down 4=z 5=x
//
// Coordinates are 0-based cells; numbers are floored, so x = 10.6 draws at 10.

func register(L *lua.LState, s *screen, in *input) {
	num := func(n int) int { return int(math.Floor(float64(L.CheckNumber(n)))) }
	set := func(x, y int, ch rune) {
		if x >= 0 && x < W && y >= 0 && y < H {
			s[y][x] = ch
		}
	}
	firstRune := func(str string) rune {
		for _, r := range str {
			return r
		}
		return ' '
	}

	L.SetGlobal("cls", L.NewFunction(func(L *lua.LState) int {
		for y := range s {
			for x := range s[y] {
				s[y][x] = ' '
			}
		}
		return 0
	}))

	L.SetGlobal("put", L.NewFunction(func(L *lua.LState) int {
		set(num(1), num(2), firstRune(L.CheckString(3)))
		return 0
	}))

	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		x0, y := num(2), num(3)
		x := x0
		for _, r := range text {
			if r == '\n' {
				x, y = x0, y+1
				continue
			}
			set(x, y, r)
			x++
		}
		return 0
	}))

	L.SetGlobal("art", L.NewFunction(func(L *lua.LState) int {
		x0, y0 := num(1), num(2)
		text := L.CheckString(3)
		for dy, line := range strings.Split(text, "\n") {
			dx := 0
			for _, r := range line {
				if r != ' ' {
					set(x0+dx, y0+dy, r)
				}
				dx++
			}
		}
		return 0
	}))

	L.SetGlobal("fill", L.NewFunction(func(L *lua.LState) int {
		x0, y0, x1, y1 := num(1), num(2), num(3), num(4)
		ch := firstRune(L.CheckString(5))
		if x0 > x1 {
			x0, x1 = x1, x0
		}
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				set(x, y, ch)
			}
		}
		return 0
	}))

	L.SetGlobal("btn", L.NewFunction(func(L *lua.LState) int {
		i := L.CheckInt(1)
		L.Push(lua.LBool(i >= 0 && i < 6 && in.held(i)))
		return 1
	}))
}

// call a global Lua function if it exists (_init / _update / _draw).
func call(L *lua.LState, name string) error {
	fn := L.GetGlobal(name)
	if fn.Type() != lua.LTFunction {
		return nil
	}
	return L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true})
}

// ---------- renderer ----------
// Only cells that changed since the last frame are redrawn; output is flushed once per frame.

func render(w *bufio.Writer, s, prev *screen, full bool) {
	for y := 0; y < H; y++ {
		cursorOK := false
		for x := 0; x < W; x++ {
			ch := s[y][x]
			if ch < ' ' { // empty cell or control character -> space
				ch = ' '
			}
			if !full && prev[y][x] == ch {
				cursorOK = false
				continue
			}
			prev[y][x] = ch
			if !cursorOK {
				fmt.Fprintf(w, "\x1b[%d;%dH", y+1, x+1)
				cursorOK = true
			}
			w.WriteRune(ch)
		}
	}
	w.Flush()
}

// ---------- main loop ----------

func run(L *lua.LState, s, prev *screen, in *input) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	startInput(in)
	if err := call(L, "_init"); err != nil {
		return err
	}

	out := bufio.NewWriterSize(os.Stdout, 64*1024)
	tick := time.NewTicker(time.Second / FPS)
	defer tick.Stop()

	full := true
	for range tick.C {
		if in.quitting() {
			return nil
		}
		if err := call(L, "_update"); err != nil {
			return err
		}
		if err := call(L, "_draw"); err != nil {
			return err
		}
		in.endFrame()
		render(out, s, prev, full)
		full = false
	}
	return nil
}

func main() {
	path := "game.lua"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	L := lua.NewState()
	defer L.Close()

	var s, prev screen
	in := &input{}
	register(L, &s, in)

	// Load the game before touching the terminal so syntax errors print normally.
	if err := L.DoFile(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	restore, err := setupTerminal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = run(L, &s, &prev, in)
	restore()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
