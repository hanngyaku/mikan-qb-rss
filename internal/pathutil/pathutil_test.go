package pathutil

import "testing"

func TestCleanDirName(t *testing.T) {
	if got := CleanDirName(`show:/\*?<>|". `); got != `show_________` {
		t.Fatalf("got %q", got)
	}
	if got := CleanDirName(".."); got != "untitled" {
		t.Fatalf("got %q", got)
	}
	if got := Join("/downloads/anime", "show"); got != "/downloads/anime/show" {
		t.Fatalf("got %q", got)
	}
}
