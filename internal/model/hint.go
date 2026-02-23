package model

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

type Hint struct {
	Mnemonic    string
	Description string
}

func (h Hint) IsBlank() bool  { return h.Mnemonic == "" && h.Description == "" }
func (h Hint) String() string { return h.Mnemonic }

type Hints []Hint

func (h Hints) Len() int      { return len(h) }
func (h Hints) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h Hints) Less(i, j int) bool {
	if h[i].Mnemonic == tcell.KeyNames[tcell.KeyEnter] {
		return true
	}

	m, err1 := strconv.Atoi(h[i].Mnemonic)
	n, err2 := strconv.Atoi(h[j].Mnemonic)
	if err1 == nil && err2 == nil {
		return m < n
	}

	if err1 == nil {
		return true
	}
	if err2 == nil {
		return false
	}

	return h[i].Description < h[j].Description
}
