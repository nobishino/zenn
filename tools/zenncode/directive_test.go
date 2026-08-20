package main

import (
	"strings"
	"testing"
)

func TestValidateDirective(t *testing.T) {
	code := "```go\nvar x int\n```\n"
	for _, tt := range []struct {
		name string
		src  string
		want string // substring of the expected error, "" if it should pass
	}{
		{"no directive", code, ""},
		{"skip", "<!-- zenncode: skip -->\n" + code, ""},
		{"playground none", "<!-- zenncode: playground=none -->\n" + code, ""},
		{"playground with another value", "<!-- zenncode: playground=off -->\n" + code, "the only supported value is none"},
		{"bare playground", "<!-- zenncode: playground -->\n" + code, "the only supported value is none"},
		{"planned key", "<!-- zenncode: file=gosample/wasm -->\n" + code, "not implemented yet"},
		{"run is gone", "<!-- zenncode: run=false -->\n" + code, "there is no run= key"},
		{"unknown key", "<!-- zenncode: nonsense=1 -->\n" + code, "unknown directive key"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDirective(block(t, tt.src).Directive)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want == "":
			case err == nil:
				t.Fatalf("expected an error mentioning %q", tt.want)
			case !strings.Contains(err.Error(), tt.want):
				t.Errorf("error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}
