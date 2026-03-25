package main

import (
	"fmt"
)

func main() {
	var events chan Event = make(chan Event)
	go terminalListener(events)
	go resizeListener(events)

	for event := range events {
		fmt.Println(event)
		if event.Type == ENUM_EVENT_KEY && event.KeyData.Key == "CTRL+C" {
			return
		}
	}
}