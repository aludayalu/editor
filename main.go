package main

import (
	"fmt"
	"flag"
)

func main() {
	var events chan Event = make(chan Event)

	debugMode := flag.Bool("debug", false, "to enable keyboard debugging information")
	flag.Parse()

	go terminalListener(events)
	go resizeListener(events)

	if *debugMode {
		KeyboardDebugging(events)
	}
}

func KeyboardDebugging(events <-chan Event) {
	for event := range events {
		if event.Type == ENUM_EVENT_KEY && event.KeyData.Key == "CTRL+C" {
			return
		}
		fmt.Print("\r")
		fmt.Println(event.KeyData.Key, event.KeyData.Data)
	}
}