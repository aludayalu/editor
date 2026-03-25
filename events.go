package main

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
	ResideData *ResizeEventData
}

type MouseEventData struct {
	Button int
	X int
	Y int
	Pressed int
}

type KeyEventData struct {
	Key string
}

type ResizeEventData struct {
	Height int
	Width  int
}