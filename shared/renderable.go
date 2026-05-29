package shared

import (
	"ltz/arena"
)

type Renderable interface {
	Render(render_info Render_Info)RenderResult
}

type Render_Info struct {
	Arena_Group *arena.ArenaGroup
	Dimensions RenderingDimensions
}

type RenderResult struct {
	Buffer *[]Cell // 2D in shape
	Rows int
	Columns int
}

type RenderingDimensions struct {
	SuggestedHeight int
	SuggestedWidth int
}

type Cell struct {
	Data []byte // contains ZW and Data
	DataVisualWidth uint64 // 1 or 2
}