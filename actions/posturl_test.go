package actions

import "testing"

// TestPostURL verifies that postURL builds the canonical, first-party
// golangflow.io URL for a post. This replaces the old goo.gl shortener,
// which Google shut down. It is a pure function, so it runs without a
// Buffalo runtime or a database.
func TestPostURL(t *testing.T) {
	got := postURL("abc-123")
	want := "https://golangflow.io/posts/abc-123"
	if got != want {
		t.Fatalf("postURL(%q) = %q, want %q", "abc-123", got, want)
	}

	// Empty ID still yields a well-formed, first-party base URL.
	if got := postURL(""); got != "https://golangflow.io/posts/" {
		t.Fatalf("postURL(\"\") = %q, want %q", got, "https://golangflow.io/posts/")
	}
}
