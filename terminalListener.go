package main

import (
	"fmt"
	"ltz/shared"
	"os"
	"strconv"
	"sync"
	"time"

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
	readBuf := make([]byte, 1024) // TODO: set this to 1024

	loopMode := 0 // 0:Normal 1:N Bytes 2:Bytes until pattern
	bytes_left_to_read := 0
	n_bytes_buffer := make([]byte, 0, 1024)

	var escapeTimeoutStartTime time.Time
	var escapeTimeoutMutex sync.Mutex

	state := 0 // 0:None 1:ESC+X 2:ESC+91+X 3:ALT+SPACE.(2)

	escapeTimeoutFunction := func(initialTimeout time.Time) {
		time.Sleep(time.Millisecond * 40)
		escapeTimeoutMutex.Lock()
		defer escapeTimeoutMutex.Unlock()

		if ((escapeTimeoutStartTime).Equal(initialTimeout)) {
			loopMode = 0
			bytes_left_to_read = 0
			n_bytes_buffer = n_bytes_buffer[:0]
			state = 0
			events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ESC"}}
			escapeTimeoutStartTime = time.Time{}
		}
	}

	requestReadN := func(nBytes int, newState int) {
		loopMode = 1
		n_bytes_buffer = n_bytes_buffer[:0]
		bytes_left_to_read = nBytes
		state = newState
	}

	resetLoop := func() {
		state = 0
		loopMode = 0
	}
	
	for {
		n, err := os.Stdin.Read(readBuf)

		if err != nil {
			panic(err)
		}

		if n == 0 {
			continue
		}

		{
			escapeTimeoutMutex.Lock()

			escapeTimeoutStartTime = time.Time{}

			escapeTimeoutMutex.Unlock()
		}

		current_buffer := readBuf[0:n]

		fmt.Println(current_buffer, loopMode, state)

		for i := 0; i < n; i++ {
			b := current_buffer[i]

			if loopMode == 0 {
				if b >= 32 && b <= 126 {
					fmt.Println("Basic", b)
					events <- shared.Event{
						Type: shared.ENUM_EVENT_KEY,
						KeyData: &shared.KeyEventData{
							Key: string(b),
						},
					}
					continue
				}

				CTRL_X_Event, is_CTRL_X := MatchesCTRL_X(b)
				if is_CTRL_X {
					fmt.Println("CTRL Basic", CTRL_X_Event.KeyData.Key)
					events <- CTRL_X_Event
					continue
				}

				if b == 27 {
					if len(current_buffer) == 1 {
						events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ESC"}}
						continue
					}
					
					if i == len(current_buffer) - 1 {
						escapeTimeoutStartTime = time.Now()
						go escapeTimeoutFunction(escapeTimeoutStartTime)
					}
					
					requestReadN(1, 1)
					continue
				}

				if b == 194 {
					requestReadN(1, 3)
					continue
				}
				continue
			}

			if loopMode == 1 {
				if bytes_left_to_read != 0 {
					n_bytes_buffer = append(n_bytes_buffer, b)
					bytes_left_to_read -= 1
				}

				// ^ always passes through

				if bytes_left_to_read == 0 {
					if state == 1 {
						first_byte := n_bytes_buffer[0]

						if first_byte == 91 {
							requestReadN(1, 2)
							continue
						}

						if first_byte == 102 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+ARROW_RIGHT"}}
							resetLoop()
							continue
						}
						if first_byte == 98 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+ARROW_LEFT"}}
							resetLoop()
							continue
						}

						if first_byte == 127 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+DELETE"}}
							resetLoop()
							continue
						}

						// if unprocessed then package escape and parse the rest
						events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ESC"}}
						resetLoop()
						i -= 1
						escapeTimeoutStartTime = time.Time{}
						continue
					}

					if state == 2 {
						first_byte := n_bytes_buffer[0]

						if first_byte == 65 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_UP"}}
							resetLoop()
							continue
						}
						if first_byte == 66 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_DOWN"}}
							resetLoop()
							continue
						}
						if first_byte == 67 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_RIGHT"}}
							resetLoop()
							continue
						}
						if first_byte == 68 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_LEFT"}}
							resetLoop()
							continue
						}

						cleanup()
						panic("Unhandled ESC+91+X=" + strconv.FormatInt(int64(n_bytes_buffer[0]), 10))
					}

					if state == 3 {
						if n_bytes_buffer[0] == 160 {
							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+SPACE"}}
							resetLoop()
							continue
						}

						cleanup()
						panic("Unhandled 194+X=" + strconv.FormatInt(int64(n_bytes_buffer[0]), 10))
					}
				}

				continue
			}
		}
	}
}

func MatchesCTRL_X(b byte) (shared.Event, bool) {
	out := shared.Event{Type: shared.ENUM_EVENT_KEY}
	var key string

	switch b {
		case 0:
			key = "CTRL+SPACE"
		case 1:
			key = "CTRL+A"
		case 2:
			key = "CTRL+B"
		case 3:
			key = "CTRL+C"
		case 4:
			key = "CTRL+D"
		case 5:
			key = "CTRL+E"
		case 6:
			key = "CTRL+F"
		case 7:
			key = "CTRL+G"
		case 8:
			key = "CTRL+DELETE" // could be CTRL+H
		case 9:
			key = "CTRL+I"
		case 10:
			key = "CTRL+J"
		case 11:
			key = "CTRL+K"
		case 12:
			key = "CTRL+L"
		case 13:
			key = "CTRL+M"
		case 14:
			key = "CTRL+N"
		case 15:
			key = "CTRL+O"
		case 16:
			key = "CTRL+P"
		case 17:
			key = "CTRL+Q"
		case 18:
			key = "CTRL+R"
		case 19:
			key = "CTRL+S"
		case 20:
			key = "CTRL+T"
		case 21:
			key = "SUPER+DELETE" // coule be CTRL+U
		case 22:
			key = "CTRL+V"
		case 23:
			key = "CTRL+W"
		case 24:
			key = "CTRL+X"
		case 25:
			key = "CTRL+Y"
		case 26:
			key = "CTRL+Z"
		case 31:
			key = "CTRL+?"
	}

	out.KeyData = &shared.KeyEventData{Key: key}

	return out, len(key) != 0
}

func MatchesEscapedCTRL_X() (shared.Event, bool) {
	// to handle M, I for ghostty
	out := shared.Event{Type: shared.ENUM_EVENT_KEY}
	var key string

	out.KeyData = &shared.KeyEventData{Key: key}

	return out, len(key) != 0
}