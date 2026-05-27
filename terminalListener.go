package main

import (
	"errors"

	"ltz/shared"
	"os"
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
	
	readBuf := make([]byte, 128) // cache line size on macbook m4 air
	validBytesCount := 0 // firstN
	consumedCount := 0

	getNextNBytes := func(requestBytesCount int) ([]byte, error) {
		if (requestBytesCount > 128) {
			return nil, errors.New("For performance reasons we are limiting nBytes reads to the size of the readBuf buffer")
		}

		if (requestBytesCount <= (validBytesCount - consumedCount)) {
			out_slice := readBuf[consumedCount : consumedCount + requestBytesCount]
			consumedCount += requestBytesCount
			return out_slice, nil
		}

		newValidBuf := readBuf[consumedCount:validBytesCount]

		copy(readBuf, newValidBuf)

		validBytesCount = len(newValidBuf)
		consumedCount = 0

		for {
			addedBytes, err := os.Stdin.Read(readBuf[validBytesCount:]); if err != nil {panic(err)}
			validBytesCount += addedBytes

			if (validBytesCount >= requestBytesCount) {
				out_slice := readBuf[consumedCount : consumedCount + requestBytesCount]
				consumedCount += requestBytesCount
				return out_slice, nil
			}
		}
	}

	nextNBytesInBuffer := func(requestBytesCount int) bool {
		return requestBytesCount <= (validBytesCount - consumedCount)
	}

	for {
		sequence, err := getNextNBytes(1); if err != nil {panic(err)}
		first_byte := sequence[0]

		if first_byte >= 32 && first_byte <= 126 {
			events <- shared.Event{
				Type: shared.ENUM_EVENT_KEY,
				KeyData: &shared.KeyEventData{
					Key: string(first_byte),
				},
			}
			continue
		}

		if first_byte == 127 {
			events <- shared.Event{
				Type: shared.ENUM_EVENT_KEY,
				KeyData: &shared.KeyEventData{
					Key: "BACKSPACE",
				},
			}
			continue
		}

		{
			event, matches := MatchesCTRL_X(first_byte)

			if matches {
				events <- event
				continue
			}
		}

		if first_byte == 27 {
			timeout_time := time.Time{}
			timeout_mutex := sync.Mutex{}
			starting_timeout := !nextNBytesInBuffer(1)

			if starting_timeout {
				timeout_time = time.Now()
				go func() {
					time.Sleep(time.Millisecond * 40)
					timeout_mutex.Lock()
					if !timeout_time.Equal(time.Time{}) {
						timeout_time = time.Time{}
						events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ESC"}}
					}
					timeout_mutex.Unlock()
				}()
			}

			sequence, err = getNextNBytes(1); if err != nil {panic(err)}

			if starting_timeout {
				timeout_mutex.Lock()
				if timeout_time.Equal(time.Time{}) {
					consumedCount -= 1
					continue
				}
				timeout_mutex.Unlock()
			}

			if sequence[0] == 98 {
				events <- shared.Event{
					Type: shared.ENUM_EVENT_KEY,
					KeyData: &shared.KeyEventData{
						Key: "ALT+ARROW_LEFT",
					},
				}
				continue
			}

			if sequence[0] == 102 {
				events <- shared.Event{
					Type: shared.ENUM_EVENT_KEY,
					KeyData: &shared.KeyEventData{
						Key: "ALT+ARROW_RIGHT",
					},
				}
				continue
			}

			if sequence[0] == 91 {
				sequence, err = getNextNBytes(1); if err != nil {panic(err)}
				
				if sequence[0] == 65 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_UP"}}
					continue
				}

				if sequence[0] == 66 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_DOWN"}}
					continue
				}

				if sequence[0] == 67 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_RIGHT"}}
					continue
				}

				if sequence[0] == 68 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ARROW_LEFT"}}
					continue
				}

				if sequence[0] == 49 {
					sequence, err = getNextNBytes(3); if err != nil {panic(err)}

					if sequence[0] == 59 {
						if sequence[1] == 51 {
							if sequence[2] == 65 {
								events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+ARROW_UP"}}
							}

							if sequence[2] == 66 {
								events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "ALT+ARROW_DOWN"}}
							}

							continue
						}
						
						if sequence[1] == 50 {
							var KeyString string

							if sequence[2] == 68 {
								KeyString = "SHIFT+ARROW_LEFT"
							}

							if sequence[2] == 67 {
								KeyString = "SHIFT+ARROW_RIGHT"
							}

							if sequence[2] == 65 {
								KeyString = "SHIFT+ARROW_UP"
							}

							if sequence[2] == 66 {
								KeyString = "SHIFT+ARROW_DOWN"
							}

							events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: KeyString}}
							continue
						}
					}

					if sequence[0] == 48 {
						getNextNBytes(2) // doing this as sequence contains [53 117] we don't check

						if sequence[2] == 59 {
							if sequence[1] == 57 {
								events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "CTRL+M"}}
								continue
							}

							if sequence[1] == 53 {
								events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "CTRL+I"}}
								continue
							}

							continue
						}

						continue
					}

					continue
				}

				if sequence[0] == 60 {
					Cb := 0
					Cx := 0
					Cy := 0

					for {
						sequence, err = getNextNBytes(1); if err != nil {panic(err)}

						if sequence[0] == ';' {
							break
						}

						Cb = (Cb * 10) + int(sequence[0]) - 48
					}

					for {
						sequence, err = getNextNBytes(1); if err != nil {panic(err)}

						if sequence[0] == ';' {
							break
						}

						Cx = (Cx * 10) + int(sequence[0]) - 48
					}

					for {
						sequence, err = getNextNBytes(1); if err != nil {panic(err)}

						if sequence[0] == 'M' || sequence[0] == 'm' {
							break
						}

						Cy = (Cy * 10) + int(sequence[0]) - 48
					}

					base := Cb & 3

					shift := Cb&4 != 0
					alt := Cb&8 != 0
					ctrl := Cb&16 != 0

					motion := Cb&32 != 0
					wheel := Cb&64 != 0

					modifier := ""
					if shift {
						modifier += "shift"
					}
					if alt {
						if modifier != "" {
							modifier += " "
						}
						modifier += "alt"
					}
					if ctrl {
						if modifier != "" {
							modifier += " "
						}
						modifier += "ctrl"
					}

					mouseData := shared.MouseEventData{
						Modifier: modifier,
						X:        Cx,
						Y:        Cy,
					}

					if wheel {
						switch base {
							case 0:
								mouseData.Scrolling = 1 // up
							case 1:
								mouseData.Scrolling = 2 // down
							case 2:
								mouseData.Scrolling = 3 // left
							case 3:
								mouseData.Scrolling = 4 // right
						}
					} else {
						if motion {
							if base == 3 {
								mouseData.Hovering = true
							} else {
								mouseData.Dragging = true
								mouseData.ClickType = base + 1
							}
						} else {
							if base <= 2 {
								mouseData.ClickType = base + 1
							}
						}

						if sequence[0] == 'm' {
							mouseData.Released = true
							if base <= 2 {
								mouseData.ClickType = base + 1
							}
						}
					}

					event := shared.Event{
						Type:      shared.ENUM_EVENT_MOUSE,
						MouseData: &mouseData,
					}

					events <- event

					continue
				}

				if sequence[0] == 50 {
					temp_sequence, err := getNextNBytes(1); if err != nil {panic(err)}

					if temp_sequence[0] == 126 {
						events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "INSERT"}}
						continue
					}

					getNextNBytes(2)

					pasteContent := make([]byte, 0, 4096)
					pattern := []byte{27, 91, 50, 48, 49, 126}
					patternMatchCount := 0

					for {
						sequence, err = getNextNBytes(1); if err != nil {panic(err)}

						if patternMatchCount == 0 && sequence[0] == 27 {
							patternMatchCount += 1
							continue
						}

						if patternMatchCount == 0 {
							pasteContent = append(pasteContent, sequence[0])
							continue
						}
						
						if patternMatchCount != 0 {
							if pattern[patternMatchCount] != sequence[0] {
								pasteContent = append(pasteContent, pattern[:patternMatchCount]...)
							} else {
								patternMatchCount += 1
							}

							if patternMatchCount == 6 {
								events <- shared.Event{
									Type: shared.ENUM_EVENT_KEY,
									KeyData: &shared.KeyEventData{
										Key: "PASTE",
										Data: pasteContent,
									},
								}

								break
							}

							continue
						}
					}

					continue
				}

				if sequence[0] == 53 {
					getNextNBytes(1)
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "PAGE_UP"}}
					continue
				}

				if sequence[0] == 54 {
					getNextNBytes(1)
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "PAGE_DOWN"}}
					continue
				}

				if sequence[0] == 72 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "HOME"}}
					continue
				}

				if sequence[0] == 70 {
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "END"}}
					continue
				}

				if sequence[0] == 51 {
					getNextNBytes(1)
					events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: "DELETE"}}
					continue
				}
				
				continue
			}

			continue
		}

		if sequence[0] == 194 {
			sequence, err = getNextNBytes(1); if err != nil {panic(err)}

			key := ""

			if sequence[0] == 161 {
				key = "ALT+1"
			}

			if sequence[0] == 163 {
				key = "ALT+3"
			}

			if sequence[0] == 167  {
				key = "ALT+5"
			}

			if sequence[0] == 182 {
				key = "ALT+7"
			}

			if sequence[0] == 176 {
				key = "ALT+0"
			}

			events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: key}}
			continue
		}

		if sequence[0] == 226 {
			sequence, err = getNextNBytes(2); if err != nil {panic(err)}

			key := ""

			if sequence[0] == 132 && sequence[1] == 162 {
				key = "ALT+2"
			}

			if sequence[0] == 130 && sequence[1] == 185 {
				key = "ALT+4"
			}

			if sequence[0] == 128 && sequence[1] == 162 {
				key = "ALT+8"
			}

			events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: key}}
			continue
		}

		if sequence[0] == 204 {
			sequence, err = getNextNBytes(1); if err != nil {panic(err)}

			key := ""

			if sequence[0] == 144 {
				key = "ALT+9"
			}

			events <- shared.Event{Type: shared.ENUM_EVENT_KEY, KeyData: &shared.KeyEventData{Key: key}}
			continue
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
			key = "TAB" // could be CTRL+I
		case 10:
			key = "CTRL+J"
		case 11:
			key = "CTRL+K"
		case 12:
			key = "CTRL+L"
		case 13:
			key = "CTRL+M" // could also be full size keyboard numpad side enter
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