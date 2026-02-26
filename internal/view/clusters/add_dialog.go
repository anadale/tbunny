package clusters

import (
	"net/url"
	"strings"
	"tbunny/internal/cluster"
	"tbunny/internal/model"
	"tbunny/internal/ui"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/tview"
)

type AddClusterFn func(name string, parameters cluster.ConnectionParameters)

func ShowAddClusterDialog(app model.App, okFn AddClusterFn) {
	f := ui.NewModalForm()

	f.AddInputField("Name:", "", 30, nil, nil)
	f.AddInputField("URI:", "", 30, nil, nil)
	f.AddInputField("Username:", "", 30, nil, nil)
	f.AddInputField("Password:", "", 30, nil, nil)

	f.AddButtons([]string{"Cancel", "Create"})

	nameField := f.GetFormItem(0).(*tview.InputField)
	uriField := f.GetFormItem(1).(*tview.InputField)
	usernameField := f.GetFormItem(2).(*tview.InputField)
	passwordField := f.GetFormItem(3).(*tview.InputField)

	nameField.SetPlaceholder("localhost")
	uriField.SetPlaceholder("http://localhost:15672")
	usernameField.SetPlaceholder("guest")
	passwordField.SetPlaceholder("guest")

	f.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex != 1 {
			app.DismissModal()
			return
		}

		name := strings.TrimSpace(nameField.GetText())
		if !checkClusterName(name) {
			f.SetFocus(0)
			return
		}

		uri := strings.TrimSpace(uriField.GetText())
		if !checkUri(uri) {
			f.SetFocus(1)
			return
		}

		username := strings.TrimSpace(usernameField.GetText())
		if username == "" {
			f.SetFocus(2)
		}

		password := passwordField.GetText()
		if password == "" {
			password = "guest"
		}

		params := cluster.ConnectionParameters{
			Uri:      uri,
			Username: username,
			Password: password,
		}

		okFn(name, params)
	})

	f.SetTitle("Add cluster")

	modal := ui.NewModalDialog(f, 60, 10)
	app.ShowModal(modal)
}

func checkUri(text string) bool {
	uri, err := url.ParseRequestURI(text)
	if err != nil {
		return false
	}
	if uri.Scheme != "http" && uri.Scheme != "https" {
		return false
	}
	if uri.Host == "" {
		return false
	}

	return true
}

// checkClusterNames checks that provided string is a valid cross-platform file name:
// - it's not a path (no / and \)
// - not "." and not ".."
// - doesn't start with '.' (hidden files like .env are forbidden)
// - Windows-safe: no <>:"/\|?*, NUL, control symbols
// - doesn't end with a space or dot (important for Windows)
// - not a reserved Windows word (CON, NUL, COM1...) — basename is checked until the first dot.
func checkClusterName(s string) bool {
	if s == "" {
		return false
	}
	if !utf8.ValidString(s) {
		return false
	}

	// Special names
	if s == "." || s == ".." {
		return false
	}

	// Doesn't start ith a dot
	if strings.HasPrefix(s, ".") {
		return false
	}

	// Not a path
	if strings.ContainsRune(s, '/') || strings.ContainsRune(s, '\\') {
		return false
	}

	// NUL is forbidden
	if strings.ContainsRune(s, '\x00') {
		return false
	}

	// Not a spaces
	if strings.TrimSpace(s) == "" {
		return false
	}

	// Windows: doesn't end with a space or dot
	if strings.HasSuffix(s, " ") || strings.HasSuffix(s, ".") {
		return false
	}

	// Windows-forbidden characters
	const bad = `<>:"/\|?*`
	if strings.ContainsAny(s, bad) {
		return false
	}

	// Control characters
	for _, r := range s {
		if r < 32 || r == 127 || unicode.IsControl(r) {
			return false
		}
	}

	// Reserved Windows names:
	base := s
	if i := strings.IndexByte(s, '.'); i >= 0 {
		base = s[:i]
	}
	up := strings.ToUpper(strings.TrimSpace(base))

	switch up {
	case "CON", "PRN", "AUX", "NUL":
		return false
	}
	if len(up) == 4 {
		prefix := up[:3]
		last := up[3]
		if (prefix == "COM" || prefix == "LPT") && last >= '1' && last <= '9' {
			return false
		}
	}

	return true
}
