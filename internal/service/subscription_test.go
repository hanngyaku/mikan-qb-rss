package service

import (
	"testing"

	"github.com/example/mikan-qb-rss/internal/qbittorrent"
)

func TestSameRule(t *testing.T) {
	rule := qbittorrent.Rule{
		Enabled: true, AffectedFeeds: []string{"https://example.com/rss"},
		AssignedCategory: "MikanRSS", SavePath: "/downloads/anime/show",
	}
	copy := rule
	copy.LastMatch = "newer"
	copy.PreviouslyMatchedEpisodes = []string{"1"}
	if !sameRule(rule, copy) {
		t.Fatal("match history must not force a rule rewrite")
	}
	copy.SavePath = "/other"
	if sameRule(rule, copy) {
		t.Fatal("changed settings must rewrite the rule")
	}
}
