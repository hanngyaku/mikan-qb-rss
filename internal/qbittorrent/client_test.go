package qbittorrent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSSSetup(t *testing.T) {
	var rule Rule
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			_, _ = io.WriteString(w, "Ok.")
		case "/api/v2/torrents/categories":
			_, _ = io.WriteString(w, `{}`)
		case "/api/v2/torrents/createCategory", "/api/v2/rss/addFeed":
			_ = r.ParseForm()
		case "/api/v2/rss/setRule":
			_ = r.ParseForm()
			if err := json.Unmarshal([]byte(r.FormValue("ruleDef")), &rule); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := client.Login(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.EnsureCategory(ctx, "MikanRSS"); err != nil {
		t.Fatal(err)
	}
	if err := client.AddFeed(ctx, "https://example.com/rss", "Anime"); err != nil {
		t.Fatal(err)
	}
	want := Rule{Enabled: true, MustContain: "1080", UseRegex: true, AffectedFeeds: []string{"https://example.com/rss"}, AssignedCategory: "MikanRSS", SavePath: "/downloads/anime/Anime"}
	if err := client.SetRule(ctx, "Anime", want); err != nil {
		t.Fatal(err)
	}
	if rule.AssignedCategory != want.AssignedCategory || rule.SavePath != want.SavePath || len(rule.AffectedFeeds) != 1 {
		t.Fatalf("unexpected rule %#v", rule)
	}
}
