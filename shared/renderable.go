package shared

type RenderResult struct {
	Height int
	Width int
	Data string
}

type Renderable interface {
	Render(render_info Render_Info)RenderResult
}


type Render_Info struct {
	SuggestedDimensions SuggestedDimensions
}

type SuggestedDimensions struct {
	Height int
	Width int
}