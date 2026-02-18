package view

import (
	"fmt"
	"strings"
	"tbunny/internal/skins"
)

func VhostDisplayName(vhost string) string {
	if vhost == "" {
		return "(all)"
	}

	return vhost
}

func ExchangeDisplayName(exchange string) string {
	if exchange == "" {
		return "(AMQP default)"
	}

	return exchange
}

// SkinTitle decorates a title.
func SkinTitle(format string, style *skins.Frame) string {
	bgColor := style.Title.BgColor
	if bgColor == skins.DefaultColor {
		bgColor = skins.TransparentColor
	}

	format = strings.ReplaceAll(format, "[fg:bg", "["+style.Title.FgColor.String()+":"+bgColor.String())
	format = strings.Replace(format, "[hilite", "["+style.Title.HighlightColor.String(), 1)
	format = strings.Replace(format, "[key", "["+style.Menu.NumKeyColor.String(), 1)
	format = strings.Replace(format, "[filter", "["+style.Title.FilterColor.String(), 1)
	format = strings.Replace(format, "[count", "["+style.Title.CounterColor.String(), 1)
	format = strings.ReplaceAll(format, ":bg:", ":"+bgColor.String()+":")

	return format
}

// SkinStatsContent decorates stats view content with theme colors.
func SkinStatsContent(content string, style *skins.Stats) string {
	bgColor := style.BgColor.String()

	content = strings.ReplaceAll(content, "[caption:bg]", "["+style.CaptionFgColor.String()+":"+bgColor+"]")
	content = strings.ReplaceAll(content, "[caption]", "["+style.CaptionFgColor.String()+":]")
	content = strings.ReplaceAll(content, "[label:bg]", "["+style.LabelFgColor.String()+":"+bgColor+"]")
	content = strings.ReplaceAll(content, "[label]", "["+style.LabelFgColor.String()+":]")
	content = strings.ReplaceAll(content, "[value:bg]", "["+style.ValueFgColor.String()+":"+bgColor+"]")
	content = strings.ReplaceAll(content, "[value]", "["+style.ValueFgColor.String()+":]")

	return content
}

func FormatBytes[T int | int32 | int64](bytes T) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := int64(bytes) / unit; n >= unit; n /= unit {
		div *= int64(unit)
		exp++
	}

	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
