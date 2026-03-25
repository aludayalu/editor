package main

import (
	"os"
	"golang.org/x/term"
)

// written originally by ChatGPT
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

			if b >= 1 && b <= 26 {
				var key string

				switch b {
				case 13, 10:
					key = "Enter"
				case 8:
					key = "Backspace"
				default:
					key = "CTRL+" + string('A'+b-1)
				}

				events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: key}}

				if b == 3 {
					return
				}

				i++
				continue
			}

			if b == 0x1b {
				if i+1 < n && buf[i+1] != '[' {
					if i+2 < n {
						nb := buf[i+1]

						if nb >= 1 && nb <= 26 {
							key := "ALT+CTRL+" + string('A'+nb-1)

							events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: key}}

							i += 2
							continue
						}

						events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: "ALT+" + string(nb)}}

						i += 2
						continue
					}
				}

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

								events <- Event{Type: ENUM_EVENT_MOUSE, MouseData: &MouseEventData{Button: nums[0], X: nums[1], Y: nums[2], Pressed: map[bool]int{true: 1, false: 0}[c == 'M']}}

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

					j := i + 2
					val := 0
					mod := 1

					for j < n {
						c := buf[j]

						if c >= '0' && c <= '9' {
							val = val*10 + int(c-'0')
						} else if c == ';' {
							val = 0
						} else {
							if val != 0 {
								mod = val
							}

							var base string

							switch c {
							case 'A':
								base = "ArrowUp"
							case 'B':
								base = "ArrowDown"
							case 'C':
								base = "ArrowRight"
							case 'D':
								base = "ArrowLeft"
							default:
								base = "ESC"
							}

							prefix := ""

							if mod >= 2 {
								switch mod {
									case 2:
										prefix = "SHIFT+"
									case 3:
										prefix = "ALT+"
									case 4:
										prefix = "ALT+SHIFT+"
									case 5:
										prefix = "CTRL+"
									case 6:
										prefix = "CTRL+SHIFT+"
									case 7:
										prefix = "CTRL+ALT+"
									case 8:
										prefix = "CTRL+ALT+SHIFT+"
								}
							}

							events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: prefix + base}}

							j++
							break
						}

						j++
					}

					i = j
					continue
				}

				events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: "ESC"}}

				i++
				continue
			}

			events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: string(b)}}

			i++
		}
	}
}