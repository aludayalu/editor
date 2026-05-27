package shared

const (
	_ int = iota
	ENUM_EVENT_MOUSE
	ENUM_EVENT_KEY
	ENUM_EVENT_RESIZE
)

type Event struct {
	Type int
	MouseData *MouseEventData
	KeyData *KeyEventData
	ResizeData *ResizeEventData
}

type MouseEventData struct {
	Modifier string
	X int
	Y int
	ClickType int // left is 1, middle is 2, right is 3
	Dragging bool
	Scrolling int // 0 none, 1 up, 2 down, 3 left, 4 right
	Released bool
	Hovering  bool
}

type KeyEventData struct {
	Key string
	Data []byte
}

type ResizeEventData struct {
	Height int
	Width  int
}