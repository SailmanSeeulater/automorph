package media

import (
	"slices"
	"testing"
)

func TestWidths(t *testing.T) {
	cases := map[int][]int{
		1920: {640, 1280, 1920},
		1000: {640, 1000},
		640:  {640},
		300:  {300},
	}
	for max, want := range cases {
		if got := Widths(max); !slices.Equal(got, want) {
			t.Errorf("Widths(%d) = %v, want %v", max, got, want)
		}
	}
}

func TestParseURL(t *testing.T) {
	key, w, ok := ParseURL(URL("2026/11/abc123", 1280))
	if !ok || key != "2026/11/abc123" || w != 1280 {
		t.Errorf("round trip failed: %q %d %v", key, w, ok)
	}
	for _, bad := range []string{"/media/../etc/passwd", "https://x.dev/a.webp", "/media/2026/11/ABC-1.webp"} {
		if _, _, ok := ParseURL(bad); ok {
			t.Errorf("ParseURL(%q) should fail", bad)
		}
	}
}
