package posts

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello, World!":             "hello-world",
		"Hello, Wörld! Første post": "hello-world-forste-post",
		"  Ærø & Straße — 2026  ":   "aero-strasse-2026",
		"Crème brûlée":              "creme-brulee",
		"日本語":                       "post",
		"---":                       "post",
		"a--b__c":                   "a-b-c",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
