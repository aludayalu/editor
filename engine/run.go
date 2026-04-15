package engine

import (
	"fmt"
	"ltz/elements"
	"ltz/shared"
	"sync"
)

var buffer []byte = nil
var bufferMutex sync.Mutex

var dom elements.Element = nil
var domMutex sync.Mutex

func SetCursor(row int, col int) {
    fmt.Printf("\033[%d;%dH", row, col)
}

func ClearScreen() {
	fmt.Print("\033[2J")
}

func Run(events chan shared.Event) {
	SetCursor(1, 1)

	domMutex.Lock()
	dom = Test()
	domMutex.Unlock()

	buffer := make([][]string, shared.CurrentTerminalDimensions.Height)
	
	for i := 0; i < shared.CurrentTerminalDimensions.Height; i++ {
		buffer[i] = make([]string, shared.CurrentTerminalDimensions.Width)
	}

	FirstPaint := dom.Render(shared.Render_Info{
		Dimensions: shared.RenderingDimensions{
			SuggestedHeight: shared.CurrentTerminalDimensions.Height,
			SuggestedWidth: shared.CurrentTerminalDimensions.Width,
		},
		// Buffer: &buffer,
	})

	IncrementalPrint(FirstPaint)

	ProcessEvents(events)
}

func Test() elements.Element {
	return elements.Text{Text: "Hello!"}
}

func IncrementalPrint(result shared.RenderResult) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
}