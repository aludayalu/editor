package shared

type Element interface {
	Renderable
}

type Props = map[string]any