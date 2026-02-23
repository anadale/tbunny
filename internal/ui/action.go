package ui

import (
	"log/slog"
	"maps"
	"slices"
	"tbunny/internal/model"
	"tbunny/internal/sl"

	"github.com/gdamore/tcell/v2"
)

type (
	ActionHandler func(key *tcell.EventKey) *tcell.EventKey

	ActionOptions struct {
		Visible    bool
		ShowInMenu bool
		Group      int
	}

	KeyAction struct {
		Description string
		Action      ActionHandler
		Options     ActionOptions
	}

	KeyMap map[tcell.Key]KeyAction

	BindingKeysFn func(KeyMap)
)

func NewKeyAction(description string, action ActionHandler) KeyAction {
	return NewKeyActionWithOptions(description, action, ActionOptions{
		Visible:    true,
		ShowInMenu: true,
	})
}

func NewHiddenKeyAction(description string, action ActionHandler) KeyAction {
	return NewKeyActionWithOptions(description, action, ActionOptions{
		Visible:    false,
		ShowInMenu: false,
	})
}

func NewKeyActionWithGroup(description string, action ActionHandler, showInMenu bool, group int) KeyAction {
	return NewKeyActionWithOptions(description, action, ActionOptions{
		Visible:    true,
		ShowInMenu: showInMenu,
		Group:      group,
	})
}

func NewKeyActionWithOptions(description string, action ActionHandler, opts ActionOptions) KeyAction {
	return KeyAction{
		Description: description,
		Action:      action,
		Options:     opts,
	}
}

func NewKeyMap() KeyMap {
	m := make(KeyMap)

	return m
}

func (m KeyMap) Add(key tcell.Key, action KeyAction) {
	m[key] = action
}

func (m KeyMap) Merge(other KeyMap) {
	for key, action := range other {
		m[key] = action
	}
}

func (m KeyMap) MenuHints() model.Hints {
	return m.makeHints(false)
}

func (m KeyMap) HelpHints() model.Hints {
	return m.makeHints(true)
}

func (m KeyMap) makeHints(showNonMenu bool) model.Hints {
	hints := make(model.Hints, 0, len(m))
	groupedKeys := m.groupedKeys(showNonMenu)

	for _, keys := range groupedKeys {
		slices.Sort(keys)

		for _, k := range keys {
			if name, ok := tcell.KeyNames[k]; ok {
				hints = append(hints, model.Hint{
					Mnemonic:    name,
					Description: m[k].Description,
				})
			} else {
				slog.Error("Failed to get mnemonic for key", sl.Key, k)
			}
		}
	}

	return hints
}

func (m KeyMap) groupedKeys(showNonMenu bool) [][]tcell.Key {
	grouped := make(map[int][]tcell.Key)

	for k, action := range m {
		if action.Options.Visible && (showNonMenu || action.Options.ShowInMenu) {
			grouped[action.Options.Group] = append(grouped[action.Options.Group], k)
		}
	}

	groups := slices.Sorted(maps.Keys(grouped))
	keys := make([][]tcell.Key, 0, len(groups))

	for _, g := range groups {
		keys = append(keys, grouped[g])
	}

	return keys
}
