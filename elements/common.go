package elements

import (
	"ltz/shared"
)

type Styles map[string]string

type Listeners struct {
	OnClick func()
}

type Element interface {
	shared.Renderable
}