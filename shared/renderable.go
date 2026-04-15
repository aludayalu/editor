package shared

type RenderResult struct {
	Start Coordinate
	End Coordinate
}

type Coordinate struct {
	Column int
	Row int
}

type Renderable interface {
	Render(render_info Render_Info)RenderResult
}


type Render_Info struct {
	Buffer *[][]Cell // [row][column][index]
	Dimensions RenderingDimensions
}

type RenderingDimensions struct {
	SuggestedHeight int
	SuggestedWidth int
}

type Cell struct {
	Data string
}