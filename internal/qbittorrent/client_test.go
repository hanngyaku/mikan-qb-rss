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
	var preferences RSSPreferences
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			_, _ = io.WriteString(w, "Ok.")
		case "/api/v2/torrents/categories":
			_, _ = io.WriteString(w, `{}`)
		case "/api/v2/torrents/createCategory", "/api/v2/rss/addFeed":
			_ = r.ParseForm()
		case "/api/v2/app/preferences":
			_, _ = io.WriteString(w, `{"rss_processing_enabled":true,"rss_auto_downloading_enabled":false,"rss_refresh_interval":30}`)
		case "/api/v2/app/setPreferences":
			_ = r.ParseForm()
			if err := json.Unmarshal([]byte(r.FormValue("json")), &preferences); err != nil {
				t.Fatal(err)
			}
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
	current, err := client.RSSPreferences(ctx)
	if err != nil || !current.ProcessingEnabled || current.RefreshInterval != 30 {
		t.Fatalf("unexpected preferences %#v err=%v", current, err)
	}
	current.AutoDownloadingEnabled = true
	if err := client.SetRSSPreferences(ctx, current); err != nil || !preferences.AutoDownloadingEnabled {
		t.Fatalf("preferences not updated: %#v err=%v", preferences, err)
	}
}

func TestFindFeedPath(t *testing.T) {
	items := map[string]json.RawMessage{
		"folder": json.RawMessage(`{"show":{"url":"https://example.com/rss","uid":"1"}}`),
	}
	path, found := findFeedPath(items, "", "https://example.com/rss")
	if !found || path != `folder\show` {
		t.Fatalf("got path=%q found=%v", path, found)
	}
}
