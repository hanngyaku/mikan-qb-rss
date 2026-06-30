package service

import "testing"

func TestMikanMetadataParsing(t *testing.T) {
	if id := mikanBangumiID("https://mikanani.me/RSS/Bangumi?bangumiId=3906&subgroupid=1244"); id != 3906 {
		t.Fatalf("got %d", id)
	}
	html := `<div class="bangumi-poster" style="background-image: url('/images/Bangumi/poster.jpg?format=webp');"></div>`
	match := mikanPosterPattern.FindStringSubmatch(html)
	if len(match) != 2 || match[1] != "/images/Bangumi/poster.jpg?format=webp" {
		t.Fatalf("got %#v", match)
	}
	page := []byte(`<p class="bangumi-info">放送日期：&#x661F;&#x671F;&#x4E8C;</p><p class="header2-desc"> 简介<br />第二行 </p>`)
	if got := pageText(page, `放送日期：(.*?)</p>`); got != "星期二" {
		t.Fatalf("got %q", got)
	}
	if got := pageText(page, `<p class="header2-desc">(.*?)</p>`); got != "简介\n第二行" {
		t.Fatalf("got %q", got)
	}
}
