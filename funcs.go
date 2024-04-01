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
func parseColor(str string) (color int, err error) {
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
		}
		return int(n), nil
	}
	return 0, err
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

func drawRect(pdf *gofpdf.Fpdf, drawcolor int, x, y, w, h float64) {
	dr, dg, db := pdf.GetDrawColor()

	r, g, b := rgb(drawcolor)
	pdf.SetDrawColor(r, g, b)

	pdf.Rect(x, y, w, h, "D")
	pdf.SetDrawColor(dr, dg, db)
}

// paint background
func paintBackground(pdf *gofpdf.Fpdf) {
	w, h, _ := pdf.PageSize(0)
	fillRect(pdf, 0xf0f8ff, 0, 0, w, h)
}

// draw border around the content area
func drawBorder(pdf *gofpdf.Fpdf) {
	var pr rect
	calculatePageRect(pdf, &pr)
	drawRect(pdf, 0xa2adb1, pr.x, pr.y, pr.w, pr.h)
}

func drawLine(pdf *gofpdf.Fpdf, x1, y1, x2, y2 float64, color int) {
	r, g, b := pdf.GetDrawColor()
	pdf.SetDrawColor(rgb(color))
	pdf.Line(x1, y1, x2, y2)
	pdf.SetDrawColor(r, g, b)
}

type label struct {
	title string
	value string
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

func waterMark(pdf *gofpdf.Fpdf, text string) {
	var pr rect

	calculatePageRect(pdf, &pr)
	s, _ := pdf.GetFontSize()

	pdf.TransformBegin()
	pdf.SetFontSize(80)
	pdf.SetFontStyle("B")
	pdf.SetTextColor(rgb(0xfe8080))
	pdf.SetDrawColor(rgb(0x200000))
	pdf.TransformTranslate(pr.cx()-45, pr.h)
	pdf.TransformRotate(45, 0, 0)
	pdf.Text(0, 0, text)
	pdf.TransformEnd()

	pdf.SetFontSize(s)
}
