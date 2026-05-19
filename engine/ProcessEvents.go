package engine

import (
	"ltz/shared"
)

func ProcessEvents(events chan shared.Event) {
	for event := range events {
		if event.Type == shared.ENUM_EVENT_KEY {
			if event.KeyData.Key == "CTRL+C" {
				return
			}
		}
	}
}