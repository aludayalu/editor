package main

import (
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

// written originally by ChatGPT
func terminalListener(events chan<- Event) {
	fd := int(os.Stdin.Fd())
	oldState, _ := term.MakeRaw(fd)
	defer term.Restore(fd, oldState)

	os.Stdout.Write([]byte("\x1b[?1049h")) // alternate screen buffer
	defer os.Stdout.Write([]byte("\x1b[?1049l"))

	os.Stdout.Write([]byte("\x1b[?25l")) // hides cursor
	defer os.Stdout.Write([]byte("\x1b[?25h"))

	os.Stdout.Write([]byte("\x1b[?2004h")) // enable bracketed paste
	defer os.Stdout.Write([]byte("\x1b[?2004l"))

	defer os.Stdout.Write([]byte("\x1b[0m")) // reset terminal styles

	os.Stdout.Write([]byte("\x1b[?1003h\x1b[?1006h"))
	defer os.Stdout.Write([]byte("\x1b[?1003l\x1b[?1006l"))

	buf := make([]byte, 128)
	pending := make([]byte, 0, 256)

	emitKey := func(key string) {
		events <- Event{Type: ENUM_EVENT_KEY, KeyData: &KeyEventData{Key: key}}
	}

	controlKey := func(b byte) (string, bool) {
		switch b {
			case 0:
				return "CTRL+@", true
			case 8:
				return "Backspace", true
			case 9:
				return "Tab", true
			case 10, 13:
				return "Enter", true
			case 28:
				return "CTRL+\\", true
			case 29:
				return "CTRL+]", true
			case 30:
				return "CTRL+^", true
			case 31:
				return "CTRL+_", true
			case 127:
				return "Backspace", true
		}

		if b >= 1 && b <= 26 {
			return "CTRL+" + string('A'+b-1), true
		}

		return "", false
	}

	modifierPrefix := func(mod int) string {
		switch mod {
			case 2:
				return "SHIFT+"
			case 3:
				return "ALT+"
			case 4:
				return "ALT+SHIFT+"
			case 5:
				return "CTRL+"
			case 6:
				return "CTRL+SHIFT+"
			case 7:
				return "CTRL+ALT+"
			case 8:
				return "CTRL+ALT+SHIFT+"
		}

		return ""
	}

	parseSingleKey := func(data []byte) (string, int, bool) {
		if len(data) == 0 {
			return "", 0, true
		}

		b := data[0]

		if key, ok := controlKey(b); ok {
			return key, 1, false
		}

		if b < 0x80 {
			return string(b), 1, false
		}

		if !utf8.FullRune(data) {
			return "", 0, true
		}

		r, size := utf8.DecodeRune(data)
		return string(r), size, false
	}

	parseSS3 := func(data []byte) (int, bool) {
		if len(data) < 3 {
			return 0, true
		}

		if data[0] != 0x1b || data[1] != 'O' {
			return 0, false
		}

		var key string

		switch data[2] {
			case 'A':
				key = "ArrowUp"
			case 'B':
				key = "ArrowDown"
			case 'C':
				key = "ArrowRight"
			case 'D':
				key = "ArrowLeft"
			case 'H':
				key = "Home"
			case 'F':
				key = "End"
			case 'P':
				key = "F1"
			case 'Q':
				key = "F2"
			case 'R':
				key = "F3"
			case 'S':
				key = "F4"
			default:
				return 3, false
		}

		emitKey(key)
		return 3, false
	}

	parseCSI := func(data []byte) (int, bool) {
		if len(data) < 3 {
			return 0, true
		}

		if data[0] != 0x1b || data[1] != '[' {
			return 0, false
		}

		if data[2] == '<' {
			j := 3
			nums := [3]int{}
			k := 0
			val := 0
			haveVal := false

			for j < len(data) {
				c := data[j]

				if c >= '0' && c <= '9' {
					val = val*10 + int(c-'0')
					haveVal = true
				} else if c == ';' {
					if k < 3 {
						if haveVal {
							nums[k] = val
						} else {
							nums[k] = 0
						}
						k++
					}
					val = 0
					haveVal = false
				} else if c == 'M' || c == 'm' {
					if k < 3 {
						if haveVal {
							nums[k] = val
						} else {
							nums[k] = 0
						}
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

					return j + 1, false
				} else {
					return j + 1, false
				}

				j++
			}

			return 0, true
		}

		params := make([]int, 0, 4)
		val := 0
		haveVal := false
		j := 2

		appendParam := func() {
			if haveVal {
				params = append(params, val)
			} else {
				params = append(params, 0)
			}
			val = 0
			haveVal = false
		}

		for j < len(data) {
			c := data[j]

			if c >= '0' && c <= '9' {
				val = val*10 + int(c-'0')
				haveVal = true
			} else if c == ';' {
				appendParam()
			} else if c >= '@' && c <= '~' {
				if haveVal || len(params) > 0 {
					appendParam()
				}

				final := c
				mod := 1

				if len(params) >= 2 && params[1] != 0 {
					mod = params[1]
				}

				prefix := modifierPrefix(mod)
				var key string

				switch final {
					case 'A':
						key = prefix + "ArrowUp"
					case 'B':
						key = prefix + "ArrowDown"
					case 'C':
						key = prefix + "ArrowRight"
					case 'D':
						key = prefix + "ArrowLeft"
					case 'H':
						key = prefix + "Home"
					case 'F':
						key = prefix + "End"
					case 'P':
						key = prefix + "F1"
					case 'Q':
						key = prefix + "F2"
					case 'R':
						key = prefix + "F3"
					case 'S':
						key = prefix + "F4"
					case 'Z':
						if len(params) == 0 {
							key = "Shift+Tab"
						} else {
							key = prefix + "Tab"
						}
					case '~':
						if len(params) == 0 {
							return j + 1, false
						}

						switch params[0] {
							case 1:
								key = prefix + "Home"
							case 2:
								key = prefix + "Insert"
							case 3:
								key = prefix + "Delete"
							case 4:
								key = prefix + "End"
							case 5:
								key = prefix + "PageUp"
							case 6:
								key = prefix + "PageDown"
							case 11:
								key = prefix + "F1"
							case 12:
								key = prefix + "F2"
							case 13:
								key = prefix + "F3"
							case 14:
								key = prefix + "F4"
							case 15:
								key = prefix + "F5"
							case 17:
								key = prefix + "F6"
							case 18:
								key = prefix + "F7"
							case 19:
								key = prefix + "F8"
							case 20:
								key = prefix + "F9"
							case 21:
								key = prefix + "F10"
							case 23:
								key = prefix + "F11"
							case 24:
								key = prefix + "F12"
							case 25:
								key = prefix + "F13"
							case 26:
								key = prefix + "F14"
							case 28:
								key = prefix + "F15"
							case 29:
								key = prefix + "F16"
							case 31:
								key = prefix + "F17"
							case 32:
								key = prefix + "F18"
							case 33:
								key = prefix + "F19"
							case 34:
								key = prefix + "F20"
							case 200, 201:
								return j + 1, false
							default:
								return j + 1, false
						}
					default:
						return j + 1, false
				}

				emitKey(key)
				return j + 1, false
			} else {
				return j + 1, false
			}

			j++
		}

		return 0, true
	}

	parseEscape := func(data []byte) (int, bool) {
		if len(data) == 0 {
			return 0, true
		}

		if data[0] != 0x1b {
			return 0, false
		}

		if len(data) == 1 {
			return 0, true
		}

		if data[1] == '[' {
			if consumed, needMore := parseCSI(data); needMore {
				return 0, true
			} else if consumed > 0 {
				return consumed, false
			}
		}

		if data[1] == 'O' {
			if consumed, needMore := parseSS3(data); needMore {
				return 0, true
			} else if consumed > 0 {
				return consumed, false
			}
		}

		key, used, needMore := parseSingleKey(data[1:])
		if needMore {
			return 0, true
		}

		emitKey("ALT+" + key)
		return 1 + used, false
	}

	for {
		n, err := os.Stdin.Read(buf)

		if err != nil {
			close(events)
			return
		}

		pending = append(pending, buf[:n]...)

		i := 0
		for i < len(pending) {
			b := pending[i]

			if b == 0x1b {
				consumed, needMore := parseEscape(pending[i:])
				if needMore {
					break
				}

				if consumed > 0 {
					i += consumed
					continue
				}

				emitKey("ESC")
				i++
				continue
			}

			if key, ok := controlKey(b); ok {
				emitKey(key)

				if b == 3 {
					return
				}

				i++
				continue
			}

			if b < 0x80 {
				emitKey(string(b))
				i++
				continue
			}

			if !utf8.FullRune(pending[i:]) {
				break
			}

			r, size := utf8.DecodeRune(pending[i:])
			emitKey(string(r))
			i += size
		}

		if i > 0 {
			pending = append(pending[:0], pending[i:]...)
		}

		if len(pending) == 1 && pending[0] == 0x1b {
			emitKey("ESC")
			pending = pending[:0]
		}
	}
}