package elements

import (
	"ltz/shared"
)

type Text struct {
	Styles map[string]string
	Listeners map[int]func()
	Text string
}

func (element Text)Render(render_info shared.Render_Info)shared.RenderResult {

	return shared.RenderResult{

	}
}