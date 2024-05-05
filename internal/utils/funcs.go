package utils

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer *message.Printer

func init() {
	printer = message.NewPrinter(language.AmericanEnglish)
}

// Parse 24 or 32 bit color hex code.
// The first return value is the alpha value and the second value contains the 24bit RGB value
func ParseColor(str string) (float64, int, error) {
	str = strings.TrimPrefix(str, "#")

	n, err := strconv.ParseUint(str, 16, 32)
	if err != nil {
		return 0, 0, err
	}

	l := len(str)
	if l == 6 { // 24 bit color
		return 1, int(n), nil
	} else if l == 8 { // 32 bit color
		alpha := float64((0xFF000000&n)>>24) / 255
		return alpha, int(n & 0x00FFFFFF), nil
	} else {
		return 0, 0, errors.Errorf("invalid color value: %s. expected 24 or 32 bit color value", str)
	}
}

// Extract RGB channels from a 24 or 32 bit color value
func RGB(c int) (r, g, b int) {
	r = (c & 0x00FF0000) >> 16
	g = (c & 0x0000FF00) >> 8
	b = c & 0x000000FF
	return
}

// returns the alpha channel from a 32bit color value
func Alpha(c int) (a float64) {
	v := (c & 0xFF000000) >> 24
	a = float64(v) / 255.0
	return
}

func RGB2i(r, g, b int) int {
	return (r << 16) | (g << 8) | b
}

func ARGB2i(a, r, g, b int) int {
	return (a << 24) | (r << 16) | (g << 8) | b
}
