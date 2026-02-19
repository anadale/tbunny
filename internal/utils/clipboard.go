package utils

import (
	"log/slog"
	"sync"
	"tbunny/internal/sl"

	"github.com/atotto/clipboard"
)

var (
	initOnce  sync.Once
	supported bool
)

func IsClipboardSupported() bool {
	initOnce.Do(func() {
		_, err := clipboard.ReadAll()

		supported = err == nil

		if err != nil {
			slog.Debug("Clipboard is not available", sl.Error, err)
		} else {
			slog.Debug("Clipboard available")
		}
	})

	return supported
}
