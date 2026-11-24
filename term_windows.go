//go:build windows

package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows: read raw console input records instead of characters.
// Each record says which key went DOWN or UP, so holding a key is smooth
// and does not depend on the OS key-repeat delay.

var procReadConsoleInputW = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReadConsoleInputW")

const (
	enableExtendedFlags = 0x0080 // set alone: no line input, no echo, no Ctrl-C signal, no VT input, quick-edit off
	keyEvent            = 0x0001

	vkEscape = 0x1B
	vkLeft   = 0x25
	vkUp     = 0x26
	vkRight  = 0x27
	vkDown   = 0x28
	vkC      = 0x43
	vkQ      = 0x51
	vkX      = 0x58
	vkZ      = 0x5A

	ctrlPressed = 0x0004 | 0x0008 // RIGHT_CTRL_PRESSED | LEFT_CTRL_PRESSED
)

// INPUT_RECORD with the KEY_EVENT_RECORD variant of the union (20 bytes total).
type inputRecord struct {
	eventType   uint16
	_           uint16
	keyDown     int32
	repeatCount uint16
	vkCode      uint16
	scanCode    uint16
	char        uint16
	ctrlState   uint32
}

func setupTerminal() (restore func(), err error) {
	in := windows.Handle(os.Stdin.Fd())
	out := windows.Handle(os.Stdout.Fd())

	var inMode, outMode uint32
	if err := windows.GetConsoleMode(in, &inMode); err != nil {
		return nil, fmt.Errorf("stdin is not a console (run in Windows Terminal, cmd or PowerShell, not an IDE output pane): %w", err)
	}
	if err := windows.GetConsoleMode(out, &outMode); err != nil {
		return nil, fmt.Errorf("stdout is not a console: %w", err)
	}
	if err := windows.SetConsoleMode(in, enableExtendedFlags); err != nil {
		return nil, fmt.Errorf("cannot set console input mode: %w", err)
	}
	// ANSI escape codes for cursor movement
	_ = windows.SetConsoleMode(out, outMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|windows.ENABLE_PROCESSED_OUTPUT)

	// alternate screen, hide cursor, clear
	os.Stdout.WriteString("\x1b[?1049h\x1b[?25l\x1b[2J")
	return func() {
		os.Stdout.WriteString("\x1b[?25h\x1b[?1049l")
		_ = windows.SetConsoleMode(out, outMode)
		_ = windows.SetConsoleMode(in, inMode)
	}, nil
}

func startInput(st *input) { go readConsole(st) }

func readConsole(st *input) {
	h := windows.Handle(os.Stdin.Fd())
	var recs [32]inputRecord
	for {
		var n uint32
		r, _, _ := procReadConsoleInputW.Call(
			uintptr(h),
			uintptr(unsafe.Pointer(&recs[0])),
			uintptr(len(recs)),
			uintptr(unsafe.Pointer(&n)),
		)
		if r == 0 {
			return
		}
		for i := uint32(0); i < n; i++ {
			rec := &recs[i]
			if rec.eventType != keyEvent {
				continue
			}
			down := rec.keyDown != 0
			switch rec.vkCode {
			case vkLeft:
				st.setDown(0, down)
			case vkRight:
				st.setDown(1, down)
			case vkUp:
				st.setDown(2, down)
			case vkDown:
				st.setDown(3, down)
			case vkZ:
				st.setDown(4, down)
			case vkX:
				st.setDown(5, down)
			case vkQ, vkEscape:
				if down {
					st.setQuit()
				}
			case vkC:
				if down && rec.ctrlState&ctrlPressed != 0 {
					st.setQuit()
				}
			}
		}
	}
}
