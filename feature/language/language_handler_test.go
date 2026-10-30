package language

import "testing"

func TestLocalPath(t *testing.T) {
	cases := map[string]string{
		"/products?page=2":      "/products?page=2",
		"/":                     "/",
		"":                      "/",
		"//evil.example":        "/",
		"/\\evil.example":       "/",
		"https://evil.example/": "/",
		"products":              "/",
	}
	for in, want := range cases {
		if got := localPath(in); got != want {
			t.Errorf("localPath(%q) = %q, want %q", in, got, want)
		}
	}
}
