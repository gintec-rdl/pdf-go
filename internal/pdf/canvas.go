package pdf

import (
	"github.com/gintec-rdl/pdf-go/pkg/types"
	"github.com/jung-kurt/gofpdf"
)

type PdfCanvas struct {
	_pdf        *gofpdf.Fpdf
	ctx         ContextStack
	_parentUnit types.DimensionUnit
}

func NewPdfCanvas(pdf *gofpdf.Fpdf, parentUnit types.DimensionUnit) types.Canvas {
	return &PdfCanvas{_pdf: pdf, _parentUnit: parentUnit}
}

func (c *PdfCanvas) Save() {
	c.ctx.PushD(c._pdf.GetAlpha())
	c.ctx.Push(c._pdf.GetLineWidth())
	c.ctx.Push(c.GetFontSize())
	c.ctx.PushT(c._pdf.GetTextColor())
	c.ctx.PushT(c._pdf.GetDrawColor())
	c.ctx.PushT(c._pdf.GetFillColor())
	c.ctx.Push(c._pdf.GetCellMargin())
}

func (c *PdfCanvas) Restore() {
	// must match reverse order of .Save()
	c._pdf.SetCellMargin(PopSolo[float64](&c.ctx))
	c._pdf.SetFillColor(PopTrio[int](&c.ctx))
	c._pdf.SetDrawColor(PopTrio[int](&c.ctx))
	c._pdf.SetTextColor(PopTrio[int](&c.ctx))
	c._pdf.SetFontSize(PopSolo[float64](&c.ctx))
	c._pdf.SetLineWidth(PopSolo[float64](&c.ctx))
	c._pdf.SetAlpha(PopDuald2[float64, string](&c.ctx))
}

func (c *PdfCanvas) GetFontSize() float64 {
	ptSize, _ := c._pdf.GetFontSize()
	return ptSize
}

func (c *PdfCanvas) ApplyDrawingBrush(brush *types.Brush) {
	bm := "Normal"
	if brush.Fill {
		c._pdf.SetAlpha(brush.FillColor.Alpha, bm)
		c._pdf.SetFillColor(brush.FillColor.RGBfn())
	}
	if brush.Stroke {
		c._pdf.SetAlpha(brush.StrokeColor.Alpha, bm)
		c._pdf.SetDrawColor(brush.StrokeColor.RGBfn())
		c._pdf.SetLineWidth(brush.StrokeWidth)
	}
	c._pdf.SetLineCapStyle(string(brush.CapStyle))
	c._pdf.SetLineJoinStyle(string(brush.JoinStyle))
}

func (c *PdfCanvas) ApplyTypingBrush(brush *types.TextBrush) {
	bm := "Normal"
	var mode int = -1
	if brush.Fill {
		mode++
		c._pdf.SetAlpha(brush.FillColor.Alpha, bm)
		c._pdf.SetFillColor(brush.FillColor.RGBfn())
	}
	if brush.Stroke {
		mode++
		c._pdf.SetAlpha(brush.StrokeColor.Alpha, bm)
		c._pdf.SetDrawColor(brush.StrokeColor.RGBfn())
	}
	c._pdf.SetTextColor(brush.StrokeColor.RGBfn())

	ptSize, _ := c._pdf.GetFontSize()

	c._pdf.SetTextRenderingMode(mode)
	c._pdf.SetLineWidth(brush.StrokeWidth)
	c._pdf.SetFontStyle(brush.FontStyle.String())
	c._pdf.SetLineCapStyle(string(brush.CapStyle))
	c._pdf.SetLineJoinStyle(string(brush.JoinStyle))
	c._pdf.SetFont(brush.FontName, brush.FontStyle.String(), brush.FontSize.GetValue(0, 0, ptSize, types.UT_FONT_SIZE, c._parentUnit))
}

func (c *PdfCanvas) DrawText(w, h float64, text string, brush *types.TextBrush) {
	c.Save()
	c.ApplyTypingBrush(brush)
	c._pdf.CellFormat(w, h, text, "", int(brush.DisplayStyle), brush.Alignment, false, 0, "")
	c.Restore()
}

func (c *PdfCanvas) DrawRect(rect types.Rect, brush *types.Brush) {
	c.Save()
	c.ApplyDrawingBrush(brush)
	c._pdf.Rect(rect.Left, rect.Top, rect.Right, rect.Bottom, brush.DrawStyle())
	c.Restore()
}

func (c *PdfCanvas) DrawCircle(x, y, r float64, brush *types.Brush) {
	c.Save()
	c.ApplyDrawingBrush(brush)
	c._pdf.Circle(x, y, r, brush.DrawStyle())
	c.Restore()
}

func (c *PdfCanvas) DrawImage(x, y, w, h, float64, brush *types.Brush) {
	//c._pdf.ImageOptions()
}

func (c *PdfCanvas) DrawLine(x1, y1, x2, y2 float64, brush *types.Brush) {
	c.Save()
	c.ApplyDrawingBrush(brush)
	c._pdf.Line(x1, y1, x2, y2)
	c.Restore()
}

// Returns the rect of the entire page
func (c *PdfCanvas) GetPageRect() *types.Rect {
	w, h := c._pdf.GetPageSize()
	return &types.Rect{
		Left:   0,
		Top:    0,
		Right:  w,
		Bottom: h,
	}
}

// Returns the rect of the drawing area, with margins taken into account
func (c *PdfCanvas) GetDrawingRect() *types.Rect {
	w, h := c._pdf.GetPageSize()
	ml, mt, mr, mb := c._pdf.GetMargins()

	return &types.Rect{
		Left:   ml,
		Top:    mt,
		Right:  w - (ml + mr),
		Bottom: h - (mt + mb),
	}
}

func (c *PdfCanvas) GetWidth() float64 {
	return c.GetDrawingRect().Right
}

func (c *PdfCanvas) GetHeight() float64 {
	return c.GetDrawingRect().Bottom
}

func (c *PdfCanvas) GetTextWidth(text string) float64 {
	return c._pdf.GetStringWidth(text)
}
func (c *PdfCanvas) GetTextHeight() float64 {
	_, lh := c._pdf.GetFontSize()
	return lh
}

func (c *PdfCanvas) GetX() float64 {
	return c._pdf.GetX()
}

func (c *PdfCanvas) SetX(x float64) {
	c._pdf.SetX(x)
}

func (c *PdfCanvas) GetY() float64 {
	return c._pdf.GetY()
}

func (c *PdfCanvas) SetY(y float64) {
	c._pdf.SetY(y)
}

func (c *PdfCanvas) GetXY() (float64, float64) {
	return c._pdf.GetXY()
}

func (c *PdfCanvas) SetXY(x, y float64) {
	c._pdf.SetXY(x, y)
}
