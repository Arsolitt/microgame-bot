package queue

import "testing"

func TestMatchSubject(t *testing.T) {
	tests := []struct {
		subject string
		pattern string
		match   bool
	}{
		// exact match
		{"foo", "foo", true},
		{"foo.bar", "foo.bar", true},
		{"foo", "bar", false},

		// single wildcard '*'
		{"foo.bar", "foo.*", true},
		{"foo.bar", "*.bar", true},
		{"foo.bar.baz", "foo.*.baz", true},
		{"foo.bar", "*.*", true},
		{"foo.bar.baz", "foo.*", false},
		{"foo", "foo.*", false},

		// multiple wildcards
		{"foo.bar.baz", "*.*.*", true},
		{"foo.bar.baz.qux", "foo.*.*.qux", true},
		{"foo.bar", "*.*.baz", false},

		// '>' wildcard
		{"foo.bar", "foo.>", true},
		{"foo.bar.baz", "foo.>", true},
		{"foo.bar.baz.qux", "foo.>", true},
		{"foo", "foo.>", false},
		{"bar.baz", "foo.>", false},
		{"foo.bar.baz", ">", true},

		// mix '*' and '>'
		{"foo.bar.baz.qux", "foo.*.>", true},
		{"foo.bar.baz", "*.bar.>", true},

		// invalid '>' not at end (should not match)
		{"foo.bar.baz", "foo.>.bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.subject+"~"+tt.pattern, func(t *testing.T) {
			got := matchSubject(tt.subject, tt.pattern)
			if got != tt.match {
				t.Errorf("matchSubject(%q, %q) = %v, want %v", tt.subject, tt.pattern, got, tt.match)
			}
		})
	}
}
