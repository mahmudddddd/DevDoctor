package report

import (
	"fmt"
	"strings"
	"unicode"
)

func safeText(value string) string {
	var output strings.Builder
	for _, character := range value {
		if unicode.Is(unicode.Cc, character) || unicode.Is(unicode.Cf, character) {
			fmt.Fprintf(&output, "\\u%04X", character)
			continue
		}
		output.WriteRune(character)
	}
	return output.String()
}
