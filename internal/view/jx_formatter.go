package view

import (
	"fmt"
	"strings"
	"tbunny/internal/skins"

	"github.com/go-faster/jx"
)

type JxFormatter struct {
	b                 strings.Builder
	indentWidth       int
	indentLevel       int
	bgColor           string
	propertyNameColor string
	stringColor       string
	numberColor       string
	booleanColor      string
	nullColor         string
	braceColor        string
	bracketColor      string
	punctuationColor  string
}

func NewJxFormatter(skin *skins.Skin) *JxFormatter {
	j := skin.Views.Json

	f := JxFormatter{
		b:                 strings.Builder{},
		indentWidth:       j.IndentWidth,
		bgColor:           j.BgColor.String(),
		propertyNameColor: j.PropertyNameColor.String(),
		stringColor:       j.StringColor.String(),
		numberColor:       j.NumberColor.String(),
		booleanColor:      j.BooleanColor.String(),
		nullColor:         j.NullColor.String(),
		braceColor:        j.BraceColor.String(),
		bracketColor:      j.BracketColor.String(),
		punctuationColor:  j.PunctuationColor.String(),
	}

	return &f
}

func (f *JxFormatter) Format(d *jx.Decoder) (string, error) {
	err := f.formatValue(d)
	if err != nil {
		return "", err
	}

	return f.b.String(), nil
}

func (f *JxFormatter) formatValue(d *jx.Decoder) (err error) {
	switch d.Next() {
	case jx.Array:
		err = f.formatArray(d)
	case jx.Object:
		err = f.formatObject(d)
	case jx.String:
		err = f.formatString(d)
	case jx.Number:
		err = f.formatNumber(d)
	case jx.Bool:
		err = f.formatBoolean(d)
	case jx.Null:
		err = f.formatNull(d)
	default:
		err = fmt.Errorf("unsupported type: %v", d.Next())
	}

	return err
}

func (f *JxFormatter) formatObject(d *jx.Decoder) error {
	f.b.WriteString(fmt.Sprintf("[%s:%s:-]{\n", f.braceColor, f.bgColor))
	f.indentLevel++

	iter, err := d.ObjIter()
	if err != nil {
		return err
	}

	first := true

	for iter.Next() {
		if !first {
			f.b.WriteString(fmt.Sprintf("[%s:%s:-],\n", f.punctuationColor, f.bgColor))
		}

		first = false

		f.writeIndent()
		f.b.WriteString(fmt.Sprintf("[%s:%s:-]%q[%s:%s:-]: ", f.propertyNameColor, f.bgColor, string(iter.Key()), f.punctuationColor, f.bgColor))

		err = f.formatValue(d)
		if err != nil {
			return err
		}
	}

	f.b.WriteString("\n")

	f.indentLevel--
	f.writeIndent()

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]}", f.braceColor, f.bgColor))

	return nil
}

func (f *JxFormatter) formatArray(d *jx.Decoder) error {
	f.b.WriteString(fmt.Sprintf("[%s:%s:-][\n", f.bracketColor, f.bgColor))
	f.indentLevel++

	iter, err := d.ArrIter()
	if err != nil {
		return err
	}

	first := true

	for iter.Next() {
		if !first {
			f.b.WriteString(fmt.Sprintf("[%s:%s:-],\n", f.punctuationColor, f.bgColor))
		}

		first = false

		f.writeIndent()

		err = f.formatValue(d)
		if err != nil {
			return err
		}
	}

	if !first {
		f.b.WriteString("\n")
	}

	f.indentLevel--
	f.writeIndent()

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]]", f.bracketColor, f.bgColor))

	return nil
}

func (f *JxFormatter) formatString(d *jx.Decoder) error {
	v, err := d.Str()
	if err != nil {
		return err
	}

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]%q", f.stringColor, f.bgColor, v))

	return nil
}

func (f *JxFormatter) formatNumber(d *jx.Decoder) error {
	v, err := d.Num()
	if err != nil {
		return err
	}

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]%v", f.numberColor, f.bgColor, v))

	return nil
}

func (f *JxFormatter) formatBoolean(d *jx.Decoder) error {
	v, err := d.Bool()
	if err != nil {
		return err
	}

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]%t", f.booleanColor, f.bgColor, v))

	return nil
}

func (f *JxFormatter) formatNull(d *jx.Decoder) error {
	err := d.Null()
	if err != nil {
		return err
	}

	f.b.WriteString(fmt.Sprintf("[%s:%s:-]%s", f.nullColor, f.bgColor, "null"))

	return nil
}

func (f *JxFormatter) writeIndent() {
	n := f.indentLevel * f.indentWidth

	f.b.Grow(n)
	for i := 0; i < n; i++ {
		f.b.WriteByte(' ')
	}
}
