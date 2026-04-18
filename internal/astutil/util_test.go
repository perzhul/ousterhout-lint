package astutil

import "testing"

func TestIsConstructor(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"New", true},
		{"NewUser", true},
		{"NewFoo", true},
		{"NewHTTPClient", true},
		{"Newline", false},
		{"Newer", false},
		{"NewsFeed", false},
		{"new", false},
		{"", false},
		{"Get", false},
		{"NewX", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsConstructor(tc.name)
			if got != tc.want {
				t.Fatalf("IsConstructor(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
