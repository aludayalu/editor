package main

import (
	"os"

	"golang.org/x/term"
)

// written by ChatGPT
func terminalListener(events chan<- Event) {
	fd := int(os.Stdin.Fd())
	oldState, _ := term.MakeRaw(fd)
	defer term.Restore(fd, oldState)

	os.Stdout.Write([]byte("\x1b[?1003h\x1b[?1006h"))
	defer os.Stdout.Write([]byte("\x1b[?1003l\x1b[?1006l"))

	buf := make([]byte, 128)

	for {
		n, err := os.Stdin.Read(buf)

		if err != nil {
			close(events)
			return
		}

		i := 0
		for i < n {
			b := buf[i]

			if b == 3 {
				events <- Event{
					Type: ENUM_EVENT_KEY,
					KeyData: &KeyEventData{
						Key: "CTRL+C",
					},
				}
				return
			}

			if b == 0x1b {
				if i+1 < n && buf[i+1] == '[' {
					if i+2 < n && buf[i+2] == '<' {
						j := i + 3
						nums := [3]int{}
						k := 0
						val := 0

						for j < n {
							c := buf[j]

							if c >= '0' && c <= '9' {
								val = val*10 + int(c-'0')
							} else if c == ';' {
								if k < 3 {
									nums[k] = val
									k++
								}
								val = 0
							} else if c == 'M' || c == 'm' {
								if k < 3 {
									nums[k] = val
								}

								events <- Event{
									Type: ENUM_EVENT_MOUSE,
									MouseData: &MouseEventData{
										Button:  nums[0],
										X:       nums[1],
										Y:       nums[2],
										Pressed: map[bool]int{true: 1, false: 0}[c == 'M'],
									},
								}

								j++
								break
							} else {
								break
							}

							j++
						}

						i = j
						continue
					}

					if i+2 < n {
						var key string

						switch buf[i+2] {
						case 'A':
							key = "ArrowUp"
						case 'B':
							key = "ArrowDown"
						case 'C':
							key = "ArrowRight"
						case 'D':
							key = "ArrowLeft"
						default:
							key = "ESC"
						}

						events <- Event{
							Type: ENUM_EVENT_KEY,
							KeyData: &KeyEventData{
								Key: key,
							},
						}

						i += 3
						continue
					}
				}

				events <- Event{
					Type: ENUM_EVENT_KEY,
					KeyData: &KeyEventData{
						Key: "ESC",
					},
				}

				i++
				continue
			}

			events <- Event{
				Type: ENUM_EVENT_KEY,
				KeyData: &KeyEventData{
					Key: string(b),
				},
			}

			i++
		}
	}
}