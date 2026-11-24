//go:build !windows

package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

// macOS / Linux: plain terminals only send characters, never key-up,
// so input falls back to in.tap() (held for a short window per key event).
// Smooth holding here needs the kitty keyboard protocol (future work).

func setupTerminal() (restore func(), err error) {
	fd := int(os.Stdin.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("cannot enter raw mode (run in a real terminal): %w", err)
	}
	// alternate screen, hide cursor, clear
	os.Stdout.WriteString("\x1b[?1049h\x1b[?25l\x1b[2J")
	return func() {
		os.Stdout.WriteString("\x1b[?25h\x1b[?1049l")
		_ = term.Restore(fd, old)
	}, nil
}

func startInput(in *input) { go readBytes(in) }

func readBytes(in *input) {
	r := bufio.NewReader(os.Stdin)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return
		}
		switch b {
		case 3, 'q': // Ctrl-C or q
			in.setQuit()
		case 'z', 'Z':
			in.tap(4)
		case 'x', 'X':
			in.tap(5)
		case 0x1b: // arrow keys: ESC [ A..D  (or ESC O A..D)
			b2, _ := r.ReadByte()
			if b2 != '[' && b2 != 'O' {
				continue
			}
			switch b3, _ := r.ReadByte(); b3 {
			case 'D':
				in.tap(0)
			case 'C':
				in.tap(1)
			case 'A':
				in.tap(2)
			case 'B':
				in.tap(3)
			}
		}
	}
}
