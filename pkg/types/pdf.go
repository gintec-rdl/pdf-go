package types

import (
	"fmt"
	"io"
	"strings"

	"github.com/gintec-rdl/pdf-go/internal/utils"
)

type Rect struct {
	Top    float64 `json:"-"`
	Left   float64 `json:"-"`
	Right  float64 `json:"-"`
	Bottom float64 `json:"-"`
}

func (r Rect) Width() float64 {
	return r.Right - r.Left
}

func (r Rect) Height() float64 {
	return r.Bottom - r.Top
}

func (r Rect) Cx() float64 {
	return r.Right * .5
}

func (r Rect) Cy() float64 {
	return r.Bottom * .5
}

type Color struct {
	ARGB    int     `json:"-"`
	Alpha   float64 `json:"-"`
	RGB     int     `json:"-"`
	R, G, B int     `json:"-"`
}

func (c *Color) Apply(color int, alpha float64) {
	c.RGB = color
	c.Alpha = alpha
	c.ARGB = (int(alpha*255.0) << 24 & 0xFF000000) | c.RGB
	c.R, c.G, c.B = utils.RGB(color)
}

func (c Color) RGBfn() (r, g, b int) {
	return c.R, c.G, c.B
}

func (c Color) String() string {
	return fmt.Sprintf("#%X", c.ARGB)
}

type CapStyle string
type JoinStyle string

const (
	CS_CAP    CapStyle = "cap"
	CS_BUTT   CapStyle = "butt"
	CS_SQUARE CapStyle = "square"
)

func (cc *CapStyle) Parse(in string) error {
	in = strings.ToLower(in)
	if in == "cap" {
		*cc = CS_CAP
		return nil
	}
	if in == "butt" {
		*cc = CS_BUTT
		return nil
	}
	if in == "square" {
		*cc = CS_SQUARE
		return nil
	}
	return fmt.Errorf("invalid cap style `%s`", in)
}

const (
	JS_MITER JoinStyle = "miter"
	JS_ROUND JoinStyle = "round"
	JS_BEVEL JoinStyle = "bevel"
)

func (cc *JoinStyle) Parse(in string) error {
	in = strings.ToLower(in)
	if in == "miter" {
		*cc = JS_MITER
		return nil
	}
	if in == "round" {
		*cc = JS_ROUND
		return nil
	}
	if in == "bevel" {
		*cc = JS_BEVEL
		return nil
	}
	return fmt.Errorf("invalid join style `%s`", in)
}

type Brush struct {
	Fill        bool      `json:"-"`
	Stroke      bool      `json:"-"`
	CapStyle    CapStyle  `json:"-"`
	JoinStyle   JoinStyle `json:"-"`
	FillColor   Color     `json:"-"`
	StrokeColor Color     `json:"-"`
	StrokeWidth float64   `json:"-"`

	drawStyleStr *string `json:"-"`
}

func (b *Brush) DrawStyle() string {
	if b.drawStyleStr == nil {
		var styles []string
		if b.Fill {
			styles = append(styles, "F")
		}
		if b.Stroke {
			styles = append(styles, "D")
		}
		s := strings.Join(styles, "")
		b.drawStyleStr = &s
	}
	return *b.drawStyleStr
}

func (b *Brush) Copy(other *Brush) {
	b.Fill = other.Fill
	b.Stroke = other.Stroke
	b.CapStyle = other.CapStyle
	b.JoinStyle = other.JoinStyle
	b.FillColor = other.FillColor
	b.StrokeColor = other.StrokeColor
	b.StrokeWidth = other.StrokeWidth
}

type TextBrush struct {
	Brush        `json:"-"`
	FontName     string      `json:"-"`
	Alignment    string      `json:"-"` // (L)EFT, (C)ENTER, (R)IGHT, (T)OP, (B)OTTOM, (M)IDDLE, (A)BASELINE
	FontSize     Dimension   `json:"-"`
	FontStyle    FontStyle   `json:"-"`
	DisplayStyle CellDisplay `json:"-"`
}

func (b *TextBrush) Copy(other *TextBrush) {
	b.Brush.Copy(&other.Brush)
	if other.FontName != "" {
		b.FontName = other.FontName
	}
	b.FontSize = other.FontSize
	b.Alignment = other.Alignment
	b.FontStyle = other.FontStyle
	b.DisplayStyle = other.DisplayStyle
}

type Canvas interface {
	DrawText(w, h float64, text string, brush *TextBrush)
	DrawRect(rect Rect, brush *Brush)
	DrawCircle(x, y, r float64, brush *Brush)
	DrawImage(x, y, w, h, float64, brush *Brush)
	DrawLine(x1, y1, x2, y2 float64, brush *Brush)
	GetDrawingRect() *Rect
	GetPageRect() *Rect
	GetWidth() float64
	GetHeight() float64
	GetTextWidth(text string) float64
	GetTextHeight() float64
	GetFontSize() float64
	GetX() float64
	SetX(float64)
	GetY() float64
	SetY(float64)
	GetXY() (float64, float64)
	SetXY(float64, float64)
	Save()
	Restore()
}

type PdfPage interface {
	GetCanvas() Canvas
}

type PdfDocument interface {
	AddNewPage(footerNHeaderFn func(p PdfPage, pageIndex int, inFooter bool)) PdfPage
	SetBookmark(title string)
	SetTitle(title string)
	GetPage(page int) (PdfPage, bool)
	GetPageCount() int
	InitializeFonts(fonts *[]Font) error
	AddFont(fontname string, data []byte)
	SaveAndCloseF(dst string) error
	SaveAndCloseW(w io.WriteCloser) error
	Save(w io.Writer) error
}
