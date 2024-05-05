package pdfgo

import (
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer *message.Printer

func init() {
	printer = message.NewPrinter(language.AmericanEnglish)
}

// Parse 24 or 32 bit color hex code.
func parseColor(str string) (alpha float64, color int, err error) {
	alphaHint := false
	l := len(str)
	if l > 1 {
		if str[0] == '#' {
			str = str[1:]
		}
		if l == 9 {
			alphaHint = true
		}
	}
	n, err := strconv.ParseUint(str, 16, 32)
	if err == nil {
		if !alphaHint {
			n |= 0xFF << 24 // implicit full opacity
		} else {
			alpha = float64((int(n)&0xFF000000)>>24) / 255.0
		}
		color = int(n)
		return
	}
	return 0, 0, err
}

func rgb(c int) (r, g, b int) {
	r = (c & 0x00FF0000) >> 16
	g = (c & 0x0000FF00) >> 8
	b = c & 0x000000FF
	return
}

// returns the alpha channel
func alpha(c int) (a float64) {
	v := (c & 0xFF000000) >> 24
	a = float64(v) / 255.0
	return
}

func rgb2i(r, g, b int) int {
	return (r << 16) | (g << 8) | b
}

func fillRect(pdf *gofpdf.Fpdf, fillcolor int, x, y, w, h float64) {
	fr, fg, fb := pdf.GetFillColor()

	r, g, b := rgb(fillcolor)
	pdf.SetFillColor(r, g, b)

	pdf.Rect(x, y, w, h, "F")
	pdf.SetFillColor(fr, fg, fb)
}

type rect struct {
	w float64
	h float64
	x float64
	y float64

	raww float64
	rawh float64
}

// Returns the x coordinate of the centerpoint
func (r *rect) cx() float64 {
	return r.w * .5
}

// Returns the y coordinate of the centerpoint
func (r *rect) cy() float64 {
	return r.h * .5
}

// Returns the content area, taking margins into account
func calculatePageRect(pdf *gofpdf.Fpdf, r *rect) {
	getPageRect(pdf, r, true)
}

func getPageRect(pdf *gofpdf.Fpdf, r *rect, withMargins ...bool) {
	w, h := pdf.GetPageSize()
	ml, mt, mr, mb := 0.0, 0.0, 0.0, 0.0

	if len(withMargins) > 0 && withMargins[0] {
		ml, mt, mr, mb = pdf.GetMargins()
	}

	r.x = ml
	r.y = mt
	r.w = w - (mr + ml)
	r.h = h - (mb + mt)
	r.rawh = h
	r.raww = w
}
