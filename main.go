package main

import (
	"fmt"
	"flag"
)

func main() {
	var events chan Event = make(chan Event)

	debugMode := flag.Bool("debug", false, "to enable keyboard debugging information")
	flag.Parse()

	listener_cleanup := func(){}

	go terminalListener(events, &listener_cleanup)

	go resizeListener(events)

	if *debugMode {
		KeyboardDebugging(events)
	}

	listener_cleanup()
}

func KeyboardDebugging(events <-chan Event) {
	for event := range events {
		if event.Type == ENUM_EVENT_MOUSE {
			fmt.Println(event.MouseData)
		}
		if event.Type == ENUM_EVENT_KEY {
			fmt.Println(event.KeyData)
		}
		if event.Type == ENUM_EVENT_RESIZE {
			fmt.Println(event.ResideData)
		}
	}
}