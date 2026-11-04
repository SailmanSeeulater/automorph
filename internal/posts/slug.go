package posts

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const maxSlugLen = 80

// Letters that don't split into base letter + accent, so NFD can't simplify them.
var transliterate = strings.NewReplacer("ø", "o", "æ", "ae", "œ", "oe", "ß", "ss", "ð", "d", "þ", "th", "ł", "l", "đ", "d")

// Slugify turns a title into a URL slug: "Hello, Wörld!" → "hello-world".
func Slugify(title string) string {
	var b strings.Builder
	dash := false
	// NFD splits "ö" into "o" + a combining mark, so dropping the marks leaves plain ASCII.
	for _, r := range norm.NFD.String(transliterate.Replace(strings.ToLower(title))) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	slug := strings.TrimSuffix(b.String(), "-")
	if len(slug) > maxSlugLen {
		slug = strings.TrimSuffix(slug[:maxSlugLen], "-")
	}
	if slug == "" {
		slug = "post"
	}
	return slug
}

// UniqueSlug returns base, or base-2, base-3, … whichever is free first.
func UniqueSlug(ctx context.Context, taken func(context.Context, string) (bool, error), base string) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := taken(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}