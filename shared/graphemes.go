package shared

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode"
	"golang.org/x/term"
	"golang.org/x/text/width"

)

// written originally by AI

// TerminalConfig holds measured terminal behavior
type terminalGraphemeConfig struct {
	// Basic measurements
	ASCIIWidth       uint64 `json:"ascii_width"`
	CJKWidth         uint64 `json:"cjk_width"`
	FullwidthWidth   uint64 `json:"fullwidth_width"`

	// Emoji categories
	BasicEmojiWidth  uint64 `json:"basic_emoji_width"`
	EmojiVS16Width   uint64 `json:"emoji_vs16_width"`
	EmojiVS15Width   uint64 `json:"emoji_vs15_width"`

	// Complex emoji
	FlagWidth        uint64 `json:"flag_width"`
	SkinToneWidth    uint64 `json:"skin_tone_width"`
	KeycapWidth      uint64 `json:"keycap_width"`
	TagSequenceWidth uint64 `json:"tag_sequence_width"`

	// ZWJ sequences
	ZWJ2Width        uint64 `json:"zwj_2_width"`
	ZWJ3Width        uint64 `json:"zwj_3_width"`
	ZWJ4Width        uint64 `json:"zwj_4_width"`
	ZWJComplexWidth  uint64 `json:"zwj_complex_width"`

	// Combining behavior
	CombiningWidth    uint64 `json:"combining_width"`
	MultipleCombining uint64 `json:"multiple_combining"`

	// What does ZWJ look like when broken?
	// Some terminals show nothing, some show a replacement char
	ZWJAloneWidth uint64 `json:"zwj_alone_width"`

	// Derived flags
	SupportsZWJ        bool `json:"supports_zwj"`
	SupportsSkinTones  bool `json:"supports_skin_tones"`
	SupportsFlags      bool `json:"supports_flags"`
	SupportsTags       bool `json:"supports_tags"`
	SupportsVariation  bool `json:"supports_variation"`
	CombiningAddsWidth bool `json:"combining_adds_width"`

	Measured bool `json:"measured"`
}

var graphemeConfig = terminalGraphemeConfig{
	// Defaults (assume good terminal)
	ASCIIWidth:         1,
	CJKWidth:           2,
	FullwidthWidth:     2,
	BasicEmojiWidth:    2,
	EmojiVS16Width:     2,
	EmojiVS15Width:     1,
	FlagWidth:          2,
	SkinToneWidth:      2,
	KeycapWidth:        2,
	TagSequenceWidth:   2,
	ZWJ2Width:          2,
	ZWJ3Width:          2,
	ZWJ4Width:          2,
	ZWJComplexWidth:    2,
	CombiningWidth:     1,
	MultipleCombining:  1,
	ZWJAloneWidth:      0,
	SupportsZWJ:        true,
	SupportsSkinTones:  true,
	SupportsFlags:      true,
	SupportsTags:       true,
	SupportsVariation:  true,
	CombiningAddsWidth: false,
	Measured:           false,
}

type probeTest struct {
	name     string
	char     string
	category string
}

func getProbeTests() []probeTest {
	return []probeTest{
		// ASCII
		{"ASCII a", "a", "ascii"},
		{"ASCII z", "z", "ascii"},

		// CJK
		{"CJK Han", "中", "cjk"},
		{"CJK Hiragana", "あ", "cjk"},
		{"CJK Korean", "한", "cjk"},

		// Fullwidth
		{"Fullwidth A", "Ａ", "fullwidth"},

		// Basic emoji
		{"Emoji face", "😀", "emoji_basic"},
		{"Emoji heart", "❤", "emoji_basic"},
		{"Emoji star", "⭐", "emoji_basic"},
		{"Emoji hand", "👍", "emoji_basic"},

		// Variation selectors
		{"VS16 heart", "❤️", "vs16"},
		{"VS15 heart", "❤︎", "vs15"},

		// Flags
		{"Flag US", "🇺🇸", "flag"},
		{"Flag JP", "🇯🇵", "flag"},
		{"Flag GB", "🇬🇧", "flag"},

		// Tag sequences
		{"Tag England", "🏴󠁧󠁢󠁥󠁮󠁧󠁿", "tag"},

		// Skin tones
		{"Skin light", "👍🏻", "skin"},
		{"Skin dark", "👍🏿", "skin"},

		// Keycaps
		{"Keycap 1", "1️⃣", "keycap"},
		{"Keycap hash", "#️⃣", "keycap"},

		// ZWJ - critical tests
		{"ZWJ alone", "a\u200Db", "zwj_alone"}, // measure what ZWJ shows as
		{"ZWJ2 hair", "👨‍🦰", "zwj2"},
		{"ZWJ3 family", "👨‍👩‍👧", "zwj3"},
		{"ZWJ4 family", "👨‍👩‍👧‍👦", "zwj4"},
		{"ZWJ kiss", "👩‍❤️‍💋‍👨", "zwj_complex"},
		{"ZWJ pilot", "👨‍✈️", "zwj_profession"},

		// Combining
		{"Combining acute", "e\u0301", "combining"},
		{"Combining double", "e\u0301\u0303", "combining2"},

		// Box drawing (TUI)
		{"Box horiz", "─", "box"},
		{"Box vert", "│", "box"},
	}
}

// Probe runs tests to detect terminal behavior
func ProbeGraphemes() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	tests := getProbeTests()
	results := make(map[string]uint64)
	categoryWidths := make(map[string][]uint64)

	for _, t := range tests {
		w, err := measureWidth(t.char)
		if err != nil {
			continue
		}
		results[t.name] = w
		categoryWidths[t.category] = append(categoryWidths[t.category], w)
	}

	processResults(results, categoryWidths)
	graphemeConfig.Measured = true
	return nil
}

func processResults(results map[string]uint64, categories map[string][]uint64) {
	mode := func(vals []uint64) uint64 {
		if len(vals) == 0 {
			return 2
		}
		counts := make(map[uint64]uint64)
		for _, v := range vals {
			counts[v]++
		}
		maxCount, maxVal := uint64(0), vals[0]
		for v, c := range counts {
			if c > maxCount {
				maxCount, maxVal = c, v
			}
		}
		return maxVal
	}

	if vals, ok := categories["ascii"]; ok {
		graphemeConfig.ASCIIWidth = mode(vals)
	}
	if vals, ok := categories["cjk"]; ok {
		graphemeConfig.CJKWidth = mode(vals)
	}
	if vals, ok := categories["fullwidth"]; ok {
		graphemeConfig.FullwidthWidth = mode(vals)
	}
	if vals, ok := categories["emoji_basic"]; ok {
		graphemeConfig.BasicEmojiWidth = mode(vals)
	}
	if vals, ok := categories["vs16"]; ok {
		graphemeConfig.EmojiVS16Width = mode(vals)
	}
	if vals, ok := categories["vs15"]; ok {
		graphemeConfig.EmojiVS15Width = mode(vals)
	}
	graphemeConfig.SupportsVariation = (graphemeConfig.EmojiVS16Width != graphemeConfig.EmojiVS15Width)

	if vals, ok := categories["flag"]; ok {
		graphemeConfig.FlagWidth = mode(vals)
		graphemeConfig.SupportsFlags = (graphemeConfig.FlagWidth <= 2)
	}
	if vals, ok := categories["tag"]; ok {
		graphemeConfig.TagSequenceWidth = mode(vals)
		graphemeConfig.SupportsTags = (graphemeConfig.TagSequenceWidth <= 2)
	}
	if vals, ok := categories["skin"]; ok {
		graphemeConfig.SkinToneWidth = mode(vals)
		graphemeConfig.SupportsSkinTones = (graphemeConfig.SkinToneWidth <= graphemeConfig.BasicEmojiWidth)
	}
	if graphemeConfig.SkinToneWidth > graphemeConfig.BasicEmojiWidth {
		graphemeConfig.SkinToneWidth = graphemeConfig.BasicEmojiWidth
	}
	if vals, ok := categories["keycap"]; ok {
		graphemeConfig.KeycapWidth = mode(vals)
	}

	// ZWJ alone: "a\u200Db" - subtract 2 for 'a' and 'b'
	if vals, ok := categories["zwj_alone"]; ok {
		measured := mode(vals)
		if measured > 2 {
			graphemeConfig.ZWJAloneWidth = measured - 2
		} else {
			graphemeConfig.ZWJAloneWidth = 0
		}
	}

	if vals, ok := categories["zwj2"]; ok {
		graphemeConfig.ZWJ2Width = mode(vals)
	}
	if vals, ok := categories["zwj3"]; ok {
		graphemeConfig.ZWJ3Width = mode(vals)
	}
	if vals, ok := categories["zwj4"]; ok {
		graphemeConfig.ZWJ4Width = mode(vals)
	}
	if vals, ok := categories["zwj_complex"]; ok {
		graphemeConfig.ZWJComplexWidth = mode(vals)
	}

	// ZWJ supported if 👨‍👩‍👧 renders as <= 2 cells
	graphemeConfig.SupportsZWJ = (graphemeConfig.ZWJ3Width <= 2)

	if vals, ok := categories["combining"]; ok {
		graphemeConfig.CombiningWidth = mode(vals)
		graphemeConfig.CombiningAddsWidth = (graphemeConfig.CombiningWidth > 1)
	}
	if vals, ok := categories["combining2"]; ok {
		graphemeConfig.MultipleCombining = mode(vals)
	}
}

func measureWidth(s string) (uint64, error) {
	fmt.Print("\033[1;1H\033[2K")

	col1, err := queryCol()
	if err != nil {
		return 0, err
	}

	fmt.Print(s)

	col2, err := queryCol()
	if err != nil {
		return 0, err
	}

	fmt.Print("\033[1;1H\033[2K")
	return col2 - col1, nil
}

func queryCol() (uint64, error) {
	fmt.Print("\033[6n")

	buf := make([]byte, 32)
	n := 0

	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) && n < len(buf) {
		os.Stdin.SetReadDeadline(deadline)
		nr, err := os.Stdin.Read(buf[n:])
		if nr > 0 {
			n += nr
			if buf[n-1] == 'R' {
				break
			}
		}
		if err != nil {
			break
		}
	}

	re := regexp.MustCompile(`\x1b\[(\d+);(\d+)R`)
	matches := re.FindSubmatch(buf[:n])
	if matches == nil {
		return 0, fmt.Errorf("parse failed: %q", buf[:n])
	}

	var col uint64
	fmt.Sscanf(string(matches[2]), "%d", &col)
	return col, nil
}

// Grapheme represents a single visual unit
type Grapheme struct {
	Data  []byte
	Width uint64
}

// Graphemes splits string into grapheme clusters based on ACTUAL terminal behavior
func Graphemes(s string) []Grapheme {
	if len(s) == 0 {
		return nil
	}

	var result []Grapheme
	runes := []rune(s)
	i := 0

	for i < len(runes) {
		start := i
		r := runes[i]
		i++

		switch {
		// ZWJ alone (when broken terminal splits sequences)
		case r == 0x200D:
			// Just the ZWJ by itself
			// Width determined by graphemeConfig.ZWJAloneWidth

		// Regional indicator (flags)
		case isRegionalIndicator(r):
			if i < len(runes) && isRegionalIndicator(runes[i]) {
				if graphemeConfig.SupportsFlags {
					i++ // consume pair
				}
				// else: don't consume, each indicator is separate grapheme
			}

		// Emoji or potential keycap base
		case IsEmoji(r) || isKeycapBase(r):
			if graphemeConfig.SupportsZWJ {
				i = consumeFullEmojiSequence(runes, i)
			} else {
				i = consumeNonZWJModifiers(runes, i)
			}

		// Regular character
		default:
			i = consumeCombiningMarks(runes, i)
		}

		cluster := runes[start:i]
		result = append(result, Grapheme{
			Data:  []byte(string(cluster)),
			Width: graphemeWidth(cluster),
		})
	}

	return result
}

// consumeFullEmojiSequence consumes everything including ZWJ (for good terminals)
func consumeFullEmojiSequence(runes []rune, i int) int {
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == 0xFE0E || r == 0xFE0F:
			i++
		case r >= 0x1F3FB && r <= 0x1F3FF:
			i++
		case r >= 0x1F9B0 && r <= 0x1F9B3:
			i++
		case r == 0x20E3:
			i++
		case r == 0x200D:
			i++
			if i < len(runes) && isEmojiOrGenderSign(runes[i]) {
				i++
				i = consumeFullEmojiSequence(runes, i)
			}
		case r >= 0xE0020 && r <= 0xE007F:
			i++
		default:
			return i
		}
	}
	return i
}

// consumeNonZWJModifiers consumes modifiers but STOPS at ZWJ (for broken terminals)
func consumeNonZWJModifiers(runes []rune, i int) int {
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == 0xFE0E || r == 0xFE0F:
			i++
		case r >= 0x1F3FB && r <= 0x1F3FF:
			i++
		case r >= 0x1F9B0 && r <= 0x1F9B3:
			i++ // hair always consumed with base
		case r == 0x20E3:
			i++
		case r == 0x200D:
			return i // STOP - ZWJ becomes separate grapheme
		case r >= 0xE0020 && r <= 0xE007F:
			if graphemeConfig.SupportsTags {
				i++
			} else {
				return i
			}
		default:
			return i
		}
	}
	return i
}

func consumeCombiningMarks(runes []rune, i int) int {
	for i < len(runes) {
		r := runes[i]
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r) {
			i++
			continue
		}
		if r == 0xFE0E || r == 0xFE0F {
			i++
			continue
		}
		break
	}
	return i
}

func graphemeWidth(cluster []rune) uint64 {
	if len(cluster) == 0 {
		return 0
	}

	first := cluster[0]

	// ZWJ alone
	if first == 0x200D {
		return graphemeConfig.ZWJAloneWidth
	}

	// Zero-width specials
	if first == 0x200B || first == 0x200C || first == 0xFEFF || first == 0x2060 {
		return 0
	}

	// Regional indicator
	if isRegionalIndicator(first) {
		if graphemeConfig.SupportsFlags && len(cluster) >= 2 && isRegionalIndicator(cluster[1]) {
			return graphemeConfig.FlagWidth
		}
		return graphemeConfig.BasicEmojiWidth // single indicator
	}

	// Tag sequence
	for _, r := range cluster {
		if r >= 0xE0020 && r <= 0xE007F {
			if graphemeConfig.SupportsTags {
				return graphemeConfig.TagSequenceWidth
			}
			return graphemeConfig.BasicEmojiWidth
		}
	}

	// Count ZWJ and emoji in cluster
	hasZWJ := false
	emojiCount := 0
	for _, r := range cluster {
		if r == 0x200D {
			hasZWJ = true
		} else if IsEmoji(r) && !isModifier(r) {
			emojiCount++
		}
	}

	// ZWJ sequence (only if terminal supports, otherwise we split earlier)
	if hasZWJ && graphemeConfig.SupportsZWJ {
		switch emojiCount {
		case 2:
			return graphemeConfig.ZWJ2Width
		case 3:
			return graphemeConfig.ZWJ3Width
		case 4:
			return graphemeConfig.ZWJ4Width
		default:
			return graphemeConfig.ZWJComplexWidth
		}
	}

	// Skin tone
	for _, r := range cluster {
		if r >= 0x1F3FB && r <= 0x1F3FF {
			return graphemeConfig.SkinToneWidth
		}
	}

	// Keycap
	if isKeycapBase(first) {
		for _, r := range cluster {
			if r == 0x20E3 {
				return graphemeConfig.KeycapWidth
			}
		}
	}

	// Variation selectors
	hasVS16, hasVS15 := false, false
	for _, r := range cluster {
		if r == 0xFE0F {
			hasVS16 = true
		}
		if r == 0xFE0E {
			hasVS15 = true
		}
	}

	if IsEmoji(first) {
		if hasVS15 && graphemeConfig.SupportsVariation {
			return graphemeConfig.EmojiVS15Width
		}
		if hasVS16 {
			return graphemeConfig.EmojiVS16Width
		}
		return graphemeConfig.BasicEmojiWidth
	}

	// East Asian Width
	switch width.LookupRune(first).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return graphemeConfig.CJKWidth
	}

	// Combining marks
	if len(cluster) > 1 && graphemeConfig.CombiningAddsWidth {
		combiningCount := uint64(0)
		for _, r := range cluster[1:] {
			if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r) {
				combiningCount++
			}
		}
		return 1 + combiningCount
	}

	return graphemeConfig.ASCIIWidth
}

func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

func isKeycapBase(r rune) bool {
	return r == '#' || r == '*' || (r >= '0' && r <= '9')
}

func isModifier(r rune) bool {
	return (r >= 0x1F3FB && r <= 0x1F3FF) ||
		(r >= 0x1F9B0 && r <= 0x1F9B3) ||
		r == 0xFE0E || r == 0xFE0F ||
		r == 0x200D
}

func isEmojiOrGenderSign(r rune) bool {
	return IsEmoji(r) || r == 0x2640 || r == 0x2642 || r == 0x2695
}

func IsEmoji(r rune) bool {
	switch {
	case r >= 0x1F000 && r <= 0x1F0FF:
		return true
	case r >= 0x1F100 && r <= 0x1F1FF:
		return true
	case r >= 0x1F200 && r <= 0x1F2FF:
		return true
	case r >= 0x1F300 && r <= 0x1F5FF:
		return true
	case r >= 0x1F600 && r <= 0x1F64F:
		return true
	case r >= 0x1F680 && r <= 0x1F6FF:
		return true
	case r >= 0x1F780 && r <= 0x1F7FF:
		return true
	case r >= 0x1F900 && r <= 0x1F9FF:
		return true
	case r >= 0x1FA00 && r <= 0x1FAFF:
		return true
	case r >= 0x2600 && r <= 0x26FF:
		return true
	case r >= 0x2700 && r <= 0x27BF:
		return true
	case r >= 0x2300 && r <= 0x23FF:
		return true
	case r >= 0x2B00 && r <= 0x2BFF:
		return true
	case r == 0x00A9 || r == 0x00AE:
		return true
	case r == 0x203C || r == 0x2049:
		return true
	case r == 0x2122 || r == 0x2139:
		return true
	case r >= 0x2194 && r <= 0x21AA:
		return true
	case r >= 0x25AA && r <= 0x25FE:
		return true
	case r == 0x3030 || r == 0x303D:
		return true
	case r == 0x3297 || r == 0x3299:
		return true
	case r >= 0xE0020 && r <= 0xE007F:
		return true
	}
	return false
}

func StringWidth(s string) uint64 {
	w := uint64(0)
	for _, g := range Graphemes(s) {
		w += g.Width
	}
	return w
}

func SaveGraphemeConfig() error {
	dir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	path := filepath.Join(dir, "ltzsh", "graphemeConfig.json")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(graphemeConfig)

	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadGraphemeConfig() error {
	dir, err := os.UserConfigDir()

	if err != nil {
		return err
	}

	path := filepath.Join(dir, "ltzsh", "graphemeConfig.json")

	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	var loadedConfig terminalGraphemeConfig

	err = json.Unmarshal(data, &loadedConfig)

	if err == nil {
		graphemeConfig = loadedConfig
	}

	return err
}