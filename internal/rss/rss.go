package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func ParseTitle(r io.Reader) (string, error) {
	var feed struct {
		Channel struct {
			Title string `xml:"title"`
		} `xml:"channel"`
	}
	if err := xml.NewDecoder(io.LimitReader(r, 2<<20)).Decode(&feed); err != nil {
		return "", fmt.Errorf("parse RSS: %w", err)
	}
	title := strings.TrimSpace(feed.Channel.Title)
	if title == "" {
		return "", fmt.Errorf("RSS channel title is empty")
	}
	return title, nil
}

func AnimeName(title string) string {
	const prefix = "Mikan Project - "
	title = strings.TrimSpace(title)
	if strings.HasPrefix(title, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(title, prefix))
	}
	return title
}
