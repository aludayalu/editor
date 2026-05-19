package engine

import (
	"fmt"
	"ltz/arena"
	"ltz/elements"
	"ltz/shared"
)

var buffer []byte = nil
var dom elements.Element = nil

var allocationBufferInitialSize uint64 = 100 * 1024 * 1024 // 100 MB

func SetCursor(row int, col int) {
    fmt.Printf("\033[%d;%dH", row, col)
}

func Run(events chan shared.Event) {
	arena_group := arena.NewArenaGroup(allocationBufferInitialSize)
	_ = arena_group // to remove

	SetCursor(1, 1)

	ProcessEvents(events)
}