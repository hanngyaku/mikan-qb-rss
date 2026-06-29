package rss

import (
	"strings"
	"testing"
)

func TestParseAndExtractTitle(t *testing.T) {
	title, err := ParseTitle(strings.NewReader(`<rss><channel><title>Mikan Project - 尖帽子的魔法工房</title></channel></rss>`))
	if err != nil || AnimeName(title) != "尖帽子的魔法工房" {
		t.Fatalf("got title=%q err=%v", title, err)
	}
}
