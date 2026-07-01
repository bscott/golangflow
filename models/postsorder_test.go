package models_test

import (
	"testing"

	"github.com/bscott/golangflow/models"
)

// TestPostsOrder covers the homepage date-sort helper. It is a pure function,
// so it runs without a Buffalo runtime or a database.
func TestPostsOrder(t *testing.T) {
	cases := map[string]string{
		"oldest":  "created_at asc",
		"newest":  "created_at desc",
		"":        "created_at desc",
		"garbage": "created_at desc",
	}
	for in, want := range cases {
		if got := models.PostsOrder(in); got != want {
			t.Errorf("PostsOrder(%q) = %q, want %q", in, got, want)
		}
	}
}
