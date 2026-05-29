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
	go Render() // need to add sync between Render and process events

	ProcessEvents(events)
}

func Render() {
	arena_group := arena.NewArenaGroup(allocationBufferInitialSize)
	_ = arena_group // to remove

	SetCursor(1, 1)

	fmt.Println(elements.Text{Text: "😭🙏🏼"}.Render(shared.Render_Info{Arena_Group: arena_group}).Buffer)
}