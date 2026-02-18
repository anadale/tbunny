package utils

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if path == "" || !strings.HasPrefix(path, "~") {
		return filepath.Abs(path)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Заменяем ~ на путь к домашней директории
	if path == "~" {
		return home, nil
	}

	// Важно использовать filepath.Join для корректных разделителей (/ или \)
	return filepath.Join(home, path[2:]), nil
}
