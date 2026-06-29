package pathutil

import (
	"path/filepath"
	"regexp"
	"strings"
)

var invalidName = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func CleanDirName(name string) string {
	name = invalidName.ReplaceAllString(strings.TrimSpace(name), "_")
	name = strings.TrimRight(name, ". ")
	if name == "" || name == "." || name == ".." {
		return "untitled"
	}
	return name
}

func Join(root, name string) string {
	return filepath.Join(root, CleanDirName(name))
}
