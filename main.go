package main

import (
	"flag"
	"fmt"
	"ltz/engine"
	"ltz/shared"
	"os"

	"golang.org/x/term"
)

func main() {
	var events chan shared.Event = make(chan shared.Event)

	debugMode := flag.Bool("key-debug", false, "To enable basic keyboard debugging by printing events")
	toProbe := flag.Bool("grapheme", false, "Test your terminal's grapheme rendering quirks and save it to ensure unicode graphemes are more correctly rendered.\nRun this test whenever your terminal is glitchy.")
	
	flag.Parse()

	{ // tests and saves grapheme configuration
		if *toProbe {
			probeTerminal()

			err :=  shared.SaveGraphemeConfig()

			if err == nil {
				fmt.Println("Grapheme configuration has been saved!")
			} else {
				fmt.Println("Unable to save grapheme config", err)
			}

			return
		}
	}

	{ // Initializing Configuration
		shared.LoadGraphemeConfig()
	}

	listener_cleanup := func(){}

	hasStarted := make(chan int)
	go terminalListener(events, &listener_cleanup, hasStarted)
	<-hasStarted

	go resizeListener(events)

	if *debugMode {
		KeyboardDebugging(events)
	} else {
		engine.Run(events)
	}

	listener_cleanup()
}

func KeyboardDebugging(events <-chan shared.Event) {
	for event := range events {
		if event.Type == shared.ENUM_EVENT_MOUSE {
			fmt.Println(event.MouseData)
		}
		if event.Type == shared.ENUM_EVENT_KEY {
			fmt.Println(event.KeyData)
		}
		if event.Type == shared.ENUM_EVENT_RESIZE {
			fmt.Println(event.ResizeData)
		}
	}
}

func probeTerminal() {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return
	}
	defer term.Restore(fd, oldState)

	os.Stdout.Write([]byte("\x1b[?1049h"))
	defer os.Stdout.Write([]byte("\x1b[?1049l"))

	os.Stdout.Write([]byte("\x1b[?25l"))
	defer os.Stdout.Write([]byte("\x1b[?25h"))

	shared.ProbeGraphemes()
}