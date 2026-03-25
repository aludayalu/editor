package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// written by CHatGPT
func resizeListener(events chan<- Event) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	for range sig {
		w, h, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			continue
		}

		events <- Event{
			Type: ENUM_EVENT_RESIZE,
			ResideData: &ResizeEventData{
				Height: h,
				Width:  w,
			},
		}
	}
}