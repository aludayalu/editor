package main

import (
	"flag"
	"fmt"
	"ltz/engine"
	"ltz/shared"
	"os"
	"time"

	"golang.org/x/term"
)

func main() {
	var events chan shared.Event = make(chan shared.Event)
	engine.LoadGraphemeConfig()

	debugMode := flag.Bool("debug", false, "To enable keyboard debugging information")
	toProbe := flag.Bool("grapheme", false, "Test your terminal's grapheme rendering quirks and save it to ensure unicode graphemes are more correctly rendered.\nRun this test whenever your terminal is glitchy.")
	
	flag.Parse()

	listener_cleanup := func(){}

	if *toProbe {
		probeTerminal()

		err :=  engine.SaveGraphemeConfig()

		if err == nil {
			fmt.Println("Grapheme configuration has been saved!")
		} else {
			fmt.Println("Unable to save grapheme config", err)
		}

		return
	}

	go terminalListener(events, &listener_cleanup)

	go resizeListener(events)

	if *debugMode {
		KeyboardDebugging(events)
	} else {
		time.Sleep(time.Millisecond * 100)
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
			fmt.Println(event.ResideData)
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

	// Enter alternate screen
	os.Stdout.Write([]byte("\x1b[?1049h"))
	defer os.Stdout.Write([]byte("\x1b[?1049l"))

	// Hide cursor (optional but cleaner)
	os.Stdout.Write([]byte("\x1b[?25l"))
	defer os.Stdout.Write([]byte("\x1b[?25h"))

	engine.Probe()
}