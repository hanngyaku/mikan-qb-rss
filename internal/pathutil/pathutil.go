package pathutil

import (
	"path"
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
	if strings.Contains(root, "/") {
		return path.Join(root, CleanDirName(name))
	}
	return filepath.Join(root, CleanDirName(name))
}
