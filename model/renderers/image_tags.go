package renderers

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// renderContentWithImageTags preserves the legacy server-side placeholder
// semantics for explicit [img] tokens: replace placeholders in order, and
// only prepend tags for any remaining images without placeholders.
func renderContentWithImageTags(content string, imageCount int, imageOffset int) (string, int) {
	if imageCount == 0 {
		return content, imageOffset
	}

	if strings.Contains(content, "[img-") {
		return content, imageOffset + imageCount
	}

	var prefix, rendered strings.Builder
	remaining := content
	searchPlaceholders := true
	for i := range imageCount {
		imgTag := fmt.Sprintf("[img-%d]", imageOffset+i)
		if searchPlaceholders {
			if pos := strings.Index(remaining, "[img]"); pos >= 0 {
				rendered.WriteString(remaining[:pos])
				rendered.WriteString(imgTag)
				remaining = remaining[pos+len("[img]"):]
				continue
			}
			searchPlaceholders = false
		}
		prefix.WriteString(imgTag)
	}
	rendered.WriteString(remaining)
	content = rendered.String()

	if prefix.Len() > 0 && content != "" {
		if r, _ := utf8.DecodeRuneInString(content); r != utf8.RuneError && !unicode.IsSpace(r) {
			prefix.WriteByte(' ')
		}
	}

	return prefix.String() + content, imageOffset + imageCount
}
