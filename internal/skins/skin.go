package skins

import "github.com/gdamore/tcell/v2"

type (
	Skin struct {
		Body   Body   `yaml:"body" json:"body"`
		Frame  Frame  `yaml:"frame" json:"frame"`
		Info   Info   `yaml:"info" json:"info"`
		Views  Views  `yaml:"views" json:"views"`
		Dialog Dialog `yaml:"dialog" json:"dialog"`
		Help   Help   `yaml:"help" json:"help"`
	}

	Body struct {
		FgColor   Color `yaml:"fgColor" json:"fgColor"`
		BgColor   Color `yaml:"bgColor" json:"bgColor"`
		LogoColor Color `yaml:"logoColor" json:"logoColor"`
	}

	Dialog struct {
		FgColor              Color `json:"fgColor" yaml:"fgColor"`
		BgColor              Color `json:"bgColor" yaml:"bgColor"`
		ButtonFgColor        Color `json:"buttonFgColor" yaml:"buttonFgColor"`
		ButtonBgColor        Color `json:"buttonBgColor" yaml:"buttonBgColor"`
		ButtonFocusFgColor   Color `json:"buttonFocusFgColor" yaml:"buttonFocusFgColor"`
		ButtonFocusBgColor   Color `json:"buttonFocusBgColor" yaml:"buttonFocusBgColor"`
		LabelFgColor         Color `json:"labelFgColor" yaml:"labelFgColor"`
		FieldFgColor         Color `json:"fieldFgColor" yaml:"fieldFgColor"`
		DropdownFgColor      Color `json:"dropdownFgColor" yaml:"dropdownFgColor"`
		DropdownBgColor      Color `json:"dropdownBgColor" yaml:"dropdownBgColor"`
		DropdownFocusFgColor Color `json:"dropdownFocusFgColor" yaml:"dropdownFocusFgColor"`
		DropdownFocusBgColor Color `json:"dropdownFocusBgColor" yaml:"dropdownFocusBgColor"`
	}

	Frame struct {
		Title  Title  `json:"title" yaml:"title"`
		Border Border `json:"border" yaml:"border"`
		Menu   Menu   `json:"menu" yaml:"menu"`
		Crumb  Crumb  `json:"crumb" yaml:"crumb"`
	}

	Views struct {
		Table Table `json:"table" yaml:"table"`
		Stats Stats `json:"stats" yaml:"stats"`
		Json  Json  `json:"json" yaml:"json"`
	}

	Title struct {
		FgColor        Color `json:"fgColor" yaml:"fgColor"`
		BgColor        Color `json:"bgColor" yaml:"bgColor"`
		HighlightColor Color `json:"highlightColor" yaml:"highlightColor"`
		CounterColor   Color `json:"counterColor" yaml:"counterColor"`
		FilterColor    Color `json:"filterColor" yaml:"filterColor"`
	}

	Info struct {
		SectionColor Color `json:"sectionColor" yaml:"sectionColor"`
		FgColor      Color `json:"fgColor" yaml:"fgColor"`
		RevColor     Color `json:"revColor" yaml:"revColor"`
	}

	Border struct {
		FgColor    Color `json:"fgColor" yaml:"fgColor"`
		FocusColor Color `json:"focusColor" yaml:"focusColor"`
	}

	Menu struct {
		FgColor     Color     `json:"fgColor" yaml:"fgColor"`
		FgStyle     TextStyle `json:"fgStyle" yaml:"fgStyle"`
		KeyColor    Color     `json:"keyColor" yaml:"keyColor"`
		NumKeyColor Color     `json:"numKeyColor" yaml:"numKeyColor"`
	}

	Crumb struct {
		FgColor     Color `json:"fgColor" yaml:"fgColor"`
		BgColor     Color `json:"bgColor" yaml:"bgColor"`
		ActiveColor Color `json:"activeColor" yaml:"activeColor"`
	}

	Table struct {
		FgColor       Color       `json:"fgColor" yaml:"fgColor"`
		BgColor       Color       `json:"bgColor" yaml:"bgColor"`
		CursorFgColor Color       `json:"cursorFgColor" yaml:"cursorFgColor"`
		CursorBgColor Color       `json:"cursorBgColor" yaml:"cursorBgColor"`
		MarkColor     Color       `json:"markColor" yaml:"markColor"`
		Header        TableHeader `json:"header" yaml:"header"`
	}

	Stats struct {
		BgColor        Color `json:"bgColor" yaml:"bgColor"`
		LabelFgColor   Color `json:"labelFgColor" yaml:"fgColor"`
		ValueFgColor   Color `json:"valueFgColor" yaml:"valueFgColor"`
		CaptionFgColor Color `json:"captionFgColor" yaml:"captionFgColor"`
		CaptionBgColor Color `json:"captionBgColor" yaml:"captionBgColor"`
	}

	Json struct {
		BgColor           Color `json:"bgColor" yaml:"bgColor"`
		PropertyNameColor Color `json:"propertyNameColor" yaml:"propertyNameColor"`
		StringColor       Color `json:"stringColor" yaml:"stringColor"`
		NumberColor       Color `json:"numberColor" yaml:"numberColor"`
		BooleanColor      Color `json:"booleanColor" yaml:"booleanColor"`
		NullColor         Color `json:"nullColor" yaml:"nullColor"`
		BraceColor        Color `json:"braceColor" yaml:"braceColor"`
		BracketColor      Color `json:"bracketColor" yaml:"bracketColor"`
		PunctuationColor  Color `json:"punctuationColor" yaml:"punctuationColor"`
		IndentWidth       int   `json:"indentWidth" yaml:"indentWidth"`
	}

	TableHeader struct {
		FgColor     Color `json:"fgColor" yaml:"fgColor"`
		BgColor     Color `json:"bgColor" yaml:"bgColor"`
		SorterColor Color `json:"sorterColor" yaml:"sorterColor"`
	}

	Help struct {
		FgColor      Color `json:"fgColor" yaml:"fgColor"`
		BgColor      Color `json:"bgColor" yaml:"bgColor"`
		SectionColor Color `json:"sectionColor" yaml:"sectionColor"`
		KeyColor     Color `json:"keyColor" yaml:"keyColor"`
		NumKeyColor  Color `json:"numKeyColor" yaml:"numKeyColor"`
	}
)

func (s *Skin) FgColor() tcell.Color {
	return s.Body.FgColor.Color()
}

func (s *Skin) BgColor() tcell.Color {
	return s.Body.BgColor.Color()
}

func newSkin() *Skin {
	return &Skin{
		Body:   newBody(),
		Frame:  newFrame(),
		Info:   newInfo(),
		Views:  newViews(),
		Dialog: newDialog(),
		Help:   newHelp(),
	}
}

func newBody() Body {
	return Body{
		FgColor:   "cadetblue",
		BgColor:   "black",
		LogoColor: "orange",
	}
}

func newDialog() Dialog {
	return Dialog{
		FgColor:              "skyblue",
		BgColor:              "black",
		ButtonFgColor:        "skyblue",
		ButtonBgColor:        "black",
		ButtonFocusFgColor:   "black",
		ButtonFocusBgColor:   "dodgerblue",
		LabelFgColor:         "white",
		FieldFgColor:         "white",
		DropdownFgColor:      "black",
		DropdownBgColor:      "skyblue",
		DropdownFocusFgColor: "black",
		DropdownFocusBgColor: "dodgerblue",
	}
}

func newViews() Views {
	return Views{
		Table: newTable(),
		Stats: newStats(),
		Json:  newJson(),
	}
}

func newFrame() Frame {
	return Frame{
		Title:  newTitle(),
		Border: newBorder(),
		Menu:   newMenu(),
		Crumb:  newCrumb(),
	}
}

func newTitle() Title {
	return Title{
		FgColor:        "aqua",
		BgColor:        "black",
		HighlightColor: "fuchsia",
		CounterColor:   "papayawhip",
		FilterColor:    "seagreen",
	}
}

func newInfo() Info {
	return Info{
		SectionColor: "orange",
		FgColor:      "white",
		RevColor:     "aqua",
	}
}

func newBorder() Border {
	return Border{
		FgColor:    "skyblue",
		FocusColor: "lightskyblue",
	}
}

func newTable() Table {
	return Table{
		FgColor:       "skyblue",
		BgColor:       "black",
		CursorFgColor: "black",
		CursorBgColor: "skyblue",
		MarkColor:     "palegreen",
		Header:        newTableHeader(),
	}
}

func newStats() Stats {
	return Stats{
		BgColor:        "black",
		LabelFgColor:   "white",
		ValueFgColor:   "skyblue",
		CaptionFgColor: "aqua",
		CaptionBgColor: "black",
	}
}

func newJson() Json {
	return Json{
		BgColor:           "black",
		PropertyNameColor: "white",
		StringColor:       "orange",
		NumberColor:       "fuchsia",
		BooleanColor:      "aqua",
		NullColor:         "red",
		BraceColor:        "skyblue",
		BracketColor:      "skyblue",
		PunctuationColor:  "skyblue",
		IndentWidth:       2,
	}
}

func newTableHeader() TableHeader {
	return TableHeader{
		FgColor:     "white",
		BgColor:     "black",
		SorterColor: "skyblue",
	}
}

func newCrumb() Crumb {
	return Crumb{
		FgColor:     "black",
		BgColor:     "aqua",
		ActiveColor: "orange",
	}
}

func newMenu() Menu {
	return Menu{
		FgColor:     "white",
		KeyColor:    "dodgerblue",
		NumKeyColor: "fuchsia",
	}
}

func newHelp() Help {
	return Help{
		FgColor:      "cadetblue",
		BgColor:      "black",
		SectionColor: "green",
		KeyColor:     "dodgerblue",
		NumKeyColor:  "fuchsia",
	}
}
