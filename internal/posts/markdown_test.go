package posts

import (
	"strings"
	"testing"
)

func TestRenderMarkdownImages(t *testing.T) {
	out, err := RenderMarkdown(`![A *red* "boat"](/media/2026/11/abc123-1920.webp "At sea")`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`src="/media/2026/11/abc123-1920.webp"`,
		`srcset="/media/2026/11/abc123-640.webp 640w, /media/2026/11/abc123-1280.webp 1280w, /media/2026/11/abc123-1920.webp 1920w"`,
		`alt="A red &#34;boat&#34;"`,
		`title="At sea"`,
		`loading="lazy"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in\n%s", want, out)
		}
	}

	out, _ = RenderMarkdown(`![x](javascript:alert(1))`)
	if strings.Contains(out, "javascript") {
		t.Errorf("dangerous URL not removed: %s", out)
	}

	out, _ = RenderMarkdown(`![external](https://example.com/a.jpg)`)
	if strings.Contains(out, "srcset") || !strings.Contains(out, `src="https://example.com/a.jpg"`) {
		t.Errorf("external image rendered wrong: %s", out)
	}
}
