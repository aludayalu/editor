package elements

import (
	"ltz/arena"
	"ltz/shared"
)

// Text handles wrapping wrt parent element although might overflow in height by default.

type Text struct {
	Styles
	Listeners
	Text string // Use SanatizedText for dangerous text values
}

func (element Text)Render(render_info shared.Render_Info)shared.RenderResult {
	graphemes := shared.Graphemes(element.Text)
	output_buffer_len := len(graphemes)

	output_buffer := arena.AllocSlice[shared.Cell](render_info.Arena_Group, uint64(output_buffer_len))

	buffer_index := 0

	currently_in_ZW := false
	starting_index := -1
	for grapheme_index := 0; grapheme_index < len(graphemes); grapheme_index++ {
		if !currently_in_ZW && graphemes[grapheme_index].Width == 0 {
			starting_index = grapheme_index
			currently_in_ZW = true
			continue
		}

		if currently_in_ZW && graphemes[grapheme_index].Width != 0 {
			total_length_bytes := 0
			for i := starting_index; i < grapheme_index; i++ {
				total_length_bytes += len(graphemes[i].Data)
			}
			total_length_bytes += len(graphemes[grapheme_index].Data)
			output_buffer[buffer_index].Data = arena.AllocSlice[byte](render_info.Arena_Group, uint64(total_length_bytes))

			for i := starting_index; i < grapheme_index; i++ {
				output_buffer[buffer_index].Data = append(output_buffer[buffer_index].Data, graphemes[i].Data...)
			}

			output_buffer[buffer_index].Data = append(output_buffer[buffer_index].Data, graphemes[grapheme_index].Data...)

			output_buffer[buffer_index].DataVisualWidth = graphemes[grapheme_index].Width

			buffer_index += 1
			continue
		}
		
		output_buffer[buffer_index].Data = graphemes[grapheme_index].Data
		output_buffer[buffer_index].DataVisualWidth = graphemes[grapheme_index].Width
		buffer_index += 1
		continue
	}

	return shared.RenderResult{
		Buffer: &output_buffer,
	}
}