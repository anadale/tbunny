package utils

import (
	"log/slog"
	"tbunny/internal/sl"

	"golang.design/x/clipboard"
)

var clipboardSupported bool

func init() {
	err := clipboard.Init()

	clipboardSupported = err == nil

	if err != nil {
		slog.Debug("Failed to initialize clipboard", sl.Error, err)
	} else {
		slog.Debug("Clipboard initialized")
	}
}

func IsClipboardSupported() bool {
	return clipboardSupported
}
