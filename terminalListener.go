package main

import (
	"ltz/shared"
	"os"
	"golang.org/x/term"
)

func terminalListener(events chan<- shared.Event, listener_cleanup *func(), hasStarted chan <- int) {
	fd := int(os.Stdin.Fd())

	oldState, _ := term.MakeRaw(fd)

	os.Stdout.Write([]byte("\x1b[?1049h")) // alternate screen buffer

	os.Stdout.Write([]byte("\x1b[?25l")) // hides cursor

	os.Stdout.Write([]byte("\x1b[?2004h")) // enable bracketed paste

	os.Stdout.Write([]byte("\x1b[?1003h\x1b[?1006h")) // mouse enable

	cleanup := func() {
		os.Stdout.Write([]byte("\x1b[?1003l\x1b[?1006l")) // mouse disable
		os.Stdout.Write([]byte("\x1b[0m")) // resets terminal styles
		os.Stdout.Write([]byte("\x1b[?2004l")) // disables bracketed paste
		os.Stdout.Write([]byte("\x1b[?25h")) // unhides cursor
		os.Stdout.Write([]byte("\x1b[?1049l")) // restores screen buffer
		os.Stdout.Sync()
		term.Restore(fd, oldState)
	}

	*listener_cleanup = cleanup

	hasStarted <- 1

	readBuf := make([]byte, 1024)

	for {
		n, err := os.Stdin.Read(readBuf)

		if err != nil {
			panic(err)
		}

		buffer := readBuf[0:n]

		if buffer[0] >= 32 && buffer[0] <= 127 {
			events <- shared.Event{
				Type: shared.ENUM_EVENT_KEY,
				KeyData: &shared.KeyEventData{
					Key: string(buffer[0]),
				},
			}
		}
	}
}