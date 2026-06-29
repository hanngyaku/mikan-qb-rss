package service

import "testing"

func TestEpisodeNumber(t *testing.T) {
	name := `[64bitsub][Jishou Akuyaku Reijou][12][1920x1080][CHT].mp4`
	if episode, ok := episodeNumber(name); !ok || episode != 12 {
		t.Fatalf("got episode=%d ok=%v", episode, ok)
	}
	if _, ok := episodeNumber(`[show][1920x1080].mp4`); ok {
		t.Fatal("resolution must not be treated as episode")
	}
	if episode, ok := episodeNumber(`[ANi] Show - 12 [1080P][WEB-DL].mp4`); !ok || episode != 12 {
		t.Fatalf("got episode=%d ok=%v", episode, ok)
	}
}
