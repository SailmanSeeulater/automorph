package media

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Every image is stored at these widths (only the ones smaller than the image itself),
// plus its own width capped at MaxWidth.
var standardWidths = []int{640, 1280}

const MaxWidth = 1920

// Widths lists the stored vriant widths for an image whose largest variant is maxWidth.
func Widths(maxWidth int) []int {
	var widths []int
	for _, w := range standardWidths {
		if w < maxWidth {
			widths = append(widths, w)
		}
	}
	return append(widths, maxWidth)
}

// URL is the public path of one variant, e.g. /
func URL(key string, width int) string {
	return fmt.Sprintf("/media/%s-%d.webp", key, width)
}

// Srcset lists every variant for <img srcset>, so browsers download the smallest one that fits.
func Srcset(key string, maxWidth int) string {
	parts := make([]string, 0, 3)
	for _, w := range Widths(maxWidth) {
		parts = append(parts, fmt.Sprintf("%s %dw", URL(key, w), w))
	}
	return strings.Join(parts, ", ")
}

var urlPattern = regexp.MustCompile(`^/media/(\d{4}/\d{2}/[a-z0-9]+)-(\d+)\.webp$`)

// ParseUrl is the reverse of URL. It's how Markdown images get turned into srcsets.
func ParseURL(url string) (key string, width int, ok bool) {
	m := urlPattern.FindStringSubmatch(url)
	if m == nil {
		return "", 0, false
	}
	width, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}
	return m[1], width, true
}
