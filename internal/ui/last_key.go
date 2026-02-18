package ui

import "github.com/gdamore/tcell/v2"

var lastKey tcell.Key

// RecordLastKey stores the last input key for focus heuristics.
func RecordLastKey(event *tcell.EventKey) {
	if event == nil {
		return
	}
	lastKey = event.Key()
}

// ConsumeBacktab reports whether the last key was Backtab and clears it.
func ConsumeBacktab() bool {
	if lastKey == tcell.KeyBacktab {
		lastKey = tcell.KeyNUL
		return true
	}
	return false
}
