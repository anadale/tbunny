package ui

import (
	"fmt"
	"strings"
	"tbunny/internal/skins"

	"github.com/rivo/tview"
)

var LogoBig = []string{
	` _____ ____                          `,
	`|_   _| __ ) _   _ _ __  _ __  _   _ `,
	`  | | |  _ \| | | | '_ \| '_ \| | | |`,
	`  | | | |_) | |_| | | | | | | | |_| |`,
	`  |_| |____/ \__,_|_| |_|_| |_|\__, |`,
	`                               |___/ `,
}

type Splash struct {
	*tview.Flex
}

func NewSplash(skin *skins.Skin, version string) *Splash {
	s := Splash{Flex: tview.NewFlex()}

	l := tview.NewTextView()
	l.SetDynamicColors(true)
	l.SetTextAlign(tview.AlignCenter)
	s.layoutLogo(l, skin)

	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetTextAlign(tview.AlignCenter)
	s.layoutVersion(v, version, skin)

	s.SetDirection(tview.FlexRow)
	s.AddItem(l, 10, 1, false)
	s.AddItem(v, 1, 1, false)

	return &s
}

func (*Splash) layoutLogo(t *tview.TextView, skin *skins.Skin) {
	logo := strings.Join(LogoBig, fmt.Sprintf("\n[%s::b]", skin.Body.LogoColor))
	_, _ = fmt.Fprintf(t, "%s[%s::b]%s\n",
		strings.Repeat("\n", 2),
		skin.Body.LogoColor,
		logo)
}

func (*Splash) layoutVersion(t *tview.TextView, version string, skin *skins.Skin) {
	_, _ = fmt.Fprintf(t, "[%s::b]Revision [red::b]%s", skin.FgColor(), version)
}
