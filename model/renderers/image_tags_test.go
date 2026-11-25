package renderers

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestRenderContentWithImageTags(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		imageCount  int
		imageOffset int
		want        string
		wantOffset  int
	}{
		{
			name:        "prefixes when there are no placeholders",
			content:     "describe this image",
			imageCount:  2,
			imageOffset: 0,
			want:        "[img-0][img-1] describe this image",
			wantOffset:  2,
		},
		{
			name:        "replaces explicit placeholders in order",
			content:     "compare [img] and [img]",
			imageCount:  2,
			imageOffset: 3,
			want:        "compare [img-3] and [img-4]",
			wantOffset:  5,
		},
		{
			name:        "prefixes extra images after placeholders are exhausted",
			content:     "compare [img]",
			imageCount:  2,
			imageOffset: 0,
			want:        "[img-1] compare [img-0]",
			wantOffset:  2,
		},
		{
			name:        "leaves leftover placeholders when there are fewer images",
			content:     "compare [img] and [img]",
			imageCount:  1,
			imageOffset: 0,
			want:        "compare [img-0] and [img]",
			wantOffset:  1,
		},
		{
			name:        "preserves already-numbered placeholders",
			content:     "compare [img-0] and [img-1]",
			imageCount:  2,
			imageOffset: 0,
			want:        "compare [img-0] and [img-1]",
			wantOffset:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOffset := renderContentWithImageTags(tt.content, tt.imageCount, tt.imageOffset)
			if got != tt.want {
				t.Fatalf("content = %q, want %q", got, tt.want)
			}
			if gotOffset != tt.wantOffset {
				t.Fatalf("offset = %d, want %d", gotOffset, tt.wantOffset)
			}
		})
	}
}

func TestRenderContentWithImageTagsMatchesLegacy(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	parts := []string{"a", " ", "[img]", "[img-2]", "\n", "世界", "\u3000", "\xff", ""}
	for range 1000 {
		var input strings.Builder
		for range rng.Intn(30) {
			input.WriteString(parts[rng.Intn(len(parts))])
		}
		count, offset := rng.Intn(15), rng.Intn(20)
		content := input.String()
		want, wantOffset := legacyImageTags(content, count, offset)
		got, gotOffset := renderContentWithImageTags(content, count, offset)
		if got != want || gotOffset != wantOffset {
			t.Fatalf("content %q, count %d: got %q (%d), want %q (%d)", content, count, got, gotOffset, want, wantOffset)
		}
	}
}

// legacyImageTags mirrors the previous implementation to check that changing
// the construction strategy preserves placeholder and whitespace behavior.
func legacyImageTags(content string, imageCount, imageOffset int) (string, int) {
	if imageCount == 0 || strings.Contains(content, "[img-") {
		return content, imageOffset + imageCount
	}
	var prefix strings.Builder
	for i := range imageCount {
		tag := fmt.Sprintf("[img-%d]", imageOffset+i)
		if strings.Contains(content, "[img]") {
			content = strings.Replace(content, "[img]", tag, 1)
		} else {
			prefix.WriteString(tag)
		}
	}
	if prefix.Len() > 0 && content != "" {
		if r, _ := utf8.DecodeRuneInString(content); r != utf8.RuneError && !unicode.IsSpace(r) {
			prefix.WriteByte(' ')
		}
	}
	return prefix.String() + content, imageOffset + imageCount
}

func BenchmarkRenderContentWithImageTags(b *testing.B) {
	content := strings.Repeat("[img]", 10000)
	for range b.N {
		renderContentWithImageTags(content, 10000, 0)
	}
}

func BenchmarkRenderContentWithImageTagsLegacy(b *testing.B) {
	content := strings.Repeat("[img]", 10000)
	for range b.N {
		legacyImageTags(content, 10000, 0)
	}
}
