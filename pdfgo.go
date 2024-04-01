package pdfgo

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"
)

// Element type
type ElementType int

// Controls how cells are rendered in succession of each other.
type CellDisplay int

var displayMap map[string]CellDisplay = map[string]CellDisplay{
	"column": DISPLAY_COLUMN,
	"row":    DISPLAY_ROW,
	"stack":  DISPLAY_STACK,
}

func (me *CellDisplay) Parse(value string) error {
	value = strings.ToLower(value)
	d, ok := displayMap[value]
	if ok {
		*me = d
		return nil
	}
	return errors.Errorf("unsupported display value: %s", value)
}

func (me *CellDisplay) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return me.Parse(str)
}

const (
	// Header section of the document
	HEADER ElementType = iota

	// Second level element of the document.
	PAGE

	// Footer section of the document
	FOOTER

	// Cell element where items are rendered in
	CELL

	// Document element. Root element
	DOCUMENT
)

const (
	// Renders the current cell and moves the pointer to the right to render the next cell
	DISPLAY_COLUMN CellDisplay = iota

	// Renders the current cell then resets the pointer to the left, onto the next line
	DISPLAY_ROW

	// Renders the current cell then adjustst the pointer onto the next line, but below the current cell.
	DISPLAY_STACK
)

type DimensionUnit int

const (
	// Milimeter unit (mm)
	DU_MILIMETER DimensionUnit = iota

	// Centimeter unit (cm)
	DU_CENTIMETER

	// Inch unit (in)
	DU_INCH

	// Percentile unit (%)
	DU_PERCENT
)

type FontStyle int

const (
	FS_NORMAL FontStyle = 1 << iota
	FS_BOLD
	FS_ITALIC
	FS_UNDERLINE
)

func (fs FontStyle) IsBold() bool {
	return fs&FS_BOLD > 0
}

func (fs FontStyle) IsItalicized() bool {
	return fs&FS_ITALIC > 0
}

func (fs FontStyle) IsUnderlined() bool {
	return fs&FS_UNDERLINE > 0
}

func (fs FontStyle) String() string {
	if fs&FS_NORMAL > 0 {
		return ""
	}
	style := []string{}
	for _, i := range fontStrToFontStyeMap {
		if fs&i.Style > 0 {
			style = append(style, i.Value)
		}
	}
	return strings.Join(style, "")
}

func (fs *FontStyle) UnmarshalText(data []byte) error {
	info, ok := fontStrToFontStyeMap[string(data)]
	if !ok {
		return errors.Errorf("invalid font style %s", data)
	}
	*fs = info.Style
	return nil
}

type styleinfo struct {
	Style FontStyle
	Value string
}

var (
	fontStrToFontStyeMap map[string]styleinfo = map[string]styleinfo{
		"normal":    {Style: FS_NORMAL, Value: ""},
		"bold":      {Style: FS_BOLD, Value: "B"},
		"italic":    {Style: FS_ITALIC, Value: "I"},
		"underline": {Style: FS_UNDERLINE, Value: "U"},
	}

	annotationToUnit map[string]DimensionUnit = map[string]DimensionUnit{
		"mm": DU_MILIMETER,
		"in": DU_INCH,
		"%":  DU_PERCENT,
		"cm": DU_CENTIMETER,
	}

	unitToAnnotationMap map[DimensionUnit]string = map[DimensionUnit]string{
		DU_CENTIMETER: "cm",
		DU_MILIMETER:  "mm",
		DU_INCH:       "in",
		DU_PERCENT:    "%",
	}
)

func (me *DimensionUnit) UnmarshalText(data []byte) error {
	notation := string(data)
	unit, ok := annotationToUnit[notation]
	if ok {
		*me = unit
		return nil
	}
	return errors.Errorf("invalid unit %s", notation)
}

func (me *DimensionUnit) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return me.UnmarshalText([]byte(str))
}

func (me *DimensionUnit) String() string {
	return unitToAnnotationMap[*me]
}

type Dimension struct {
	Value float64
	Unit  DimensionUnit
}

type Border struct {
	Color int
	Width float64
}

type UnitType int

const (
	UT_LENGTH UnitType = 1 << iota
	UT_FONT_SIZE

	UT_LENGH_WIDTH
	UT_LENGTH_HEIGHT
)

type UnitConverter func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64

var (
	// returns a value that is relative to parent's value of the same type.
	percentageConverter = func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 {
		var r rect
		ptSize, _ := pdf.GetFontSize()
		calculatePageRect(pdf, &r)

		if flags&UT_FONT_SIZE > 0 {
			in = ptSize * in // relative to parent's font size
		} else if flags&UT_LENGTH > 0 {
			if flags&UT_LENGH_WIDTH > 0 {
				in = in * r.w // get percentage
			} else if flags&UT_LENGTH_HEIGHT > 0 {
				in = in * r.h // get percentage
			}
		}

		return in
	}
	conversionTable map[DimensionUnit]map[DimensionUnit]UnitConverter = map[DimensionUnit]map[DimensionUnit]UnitConverter{
		DU_MILIMETER: {
			DU_CENTIMETER: func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in * .1 },
			DU_INCH:       func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in * 0.0393701 },
			DU_MILIMETER:  func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in },
			//DU_PERCENT:    func(in float64, flags UnitType) float64 { return in },
		},
		DU_CENTIMETER: {
			DU_CENTIMETER: func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in },
			DU_INCH:       func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in * 0.393701 },
			DU_MILIMETER:  func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in * 10 },
			//DU_PERCENT:    func(in float64, flags UnitType) float64 { return in },
		},
		DU_INCH: {
			DU_CENTIMETER: func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in * 2.54 },
			DU_INCH:       func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return in },
			DU_MILIMETER:  func(in float64, flags UnitType, pdf *gofpdf.Fpdf) float64 { return 25.4 },
			//DU_PERCENT:    func(in float64, flags UnitType) float64 {},
		},
		DU_PERCENT: {
			DU_CENTIMETER: percentageConverter,
			DU_INCH:       percentageConverter,
			DU_MILIMETER:  percentageConverter,
		},
	}
)

// Calculates values based on unit
func (me Dimension) GetValue(pdf *gofpdf.Fpdf, flags UnitType, dstUnit DimensionUnit) float64 {
	fn, ok := conversionTable[me.Unit][dstUnit]
	if ok {
		return fn(me.Value, flags, pdf)
	}
	return 0
}

func (me *Dimension) UnmarshalJSON(data []byte) error {
	var dimStr string
	if err := json.Unmarshal(data, &dimStr); err != nil {
		return err
	}
	return me.UnmarshalText([]byte(dimStr))
}

func (me *Dimension) UnmarshalText(data []byte) error {
	dimStr := strings.ToLower(string(data))
	intPart := "0.0"

	unitParser := func(units map[string]DimensionUnit) error {
		for notation, unit := range units {
			if strings.HasSuffix(dimStr, notation) {
				me.Unit = unit
				intPart = strings.TrimSuffix(dimStr, notation)
				return nil
			}
		}
		return errors.Errorf("invalid unit %s", dimStr)
	}

	// parseunit
	if err := unitParser(annotationToUnit); err != nil {
		return err
	}

	// parse value
	value, err := strconv.ParseFloat(intPart, 32)
	if err != nil {
		return errors.Wrap(err, "invalid unit value")
	}
	me.Value = value / 100.00
	return nil
}

// Element interface for determining type of element
type IElement interface {
	Type() ElementType
	GetElement() *Element
}

type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AttributeIterator func(prefix string, attrs []Attribute, e IElement, specialHandlerOnly ...bool) error

type DocumentElement ElementType
type PageElement ElementType
type HeaderElement ElementType
type FooterElement ElementType
type CellElement ElementType

const E_DOCUMENT DocumentElement = 0
const E_HEADER HeaderElement = 0
const E_FOOTER FooterElement = 0
const E_PAGE PageElement = 0
const E_CELL CellElement = 0

type Element struct {
	Attrs []Attribute `json:"attributes"`

	// Used internally
	fontSize      Dimension
	fontStyle     FontStyle
	fontColor     int
	bgColor       int
	borderColor   int     // not inheritable
	borderWidth   float64 // not inheritable
	bookmarkTitle string  //
	border        struct {
		left   *Border
		top    *Border
		right  *Border
		bottom *Border
	}
}

func (e *Element) initBorders() {
	if e.border.left == nil {
		e.border.left = &Border{}
	}
	if e.border.right == nil {
		e.border.right = &Border{}
	}
	if e.border.top == nil {
		e.border.top = &Border{}
	}
	if e.border.bottom == nil {
		e.border.bottom = &Border{}
	}
}

func (e *Element) setBorderWidth(width float64) {
	e.border.left.Width = width
	e.border.top.Width = width
	e.border.right.Width = width
	e.border.bottom.Width = width
}

func (e *Element) setBorderColor(color int) {
	e.border.left.Color = color
	e.border.top.Color = color
	e.border.right.Color = color
	e.border.bottom.Color = color
}

func (e Element) drawBorder(pdf *gofpdf.Fpdf, left, top, right, bottom float64) {
	linewidth := pdf.GetLineWidth()
	oldx, oldy := pdf.GetXY()
	r, g, b := pdf.GetDrawColor()

	if e.border.left != nil {
		pdf.SetLineWidth(e.border.left.Width)
		pdf.SetDrawColor(rgb(e.border.left.Color))
		pdf.Line(left, top, left, bottom)
	}

	if e.border.right != nil {
		pdf.SetLineWidth(e.border.right.Width)
		pdf.SetDrawColor(rgb(e.border.right.Color))
		pdf.Line(right, top, right, bottom)
	}

	if e.border.top != nil {
		pdf.SetLineWidth(e.border.top.Width)
		pdf.SetDrawColor(rgb(e.border.top.Color))
		pdf.Line(left, top, right, top)
	}

	if e.border.bottom != nil {
		pdf.SetLineWidth(e.border.bottom.Width)
		pdf.SetDrawColor(rgb(e.border.bottom.Color))
		pdf.Line(left, bottom, right, bottom)
	}

	pdf.SetLineWidth(linewidth)
	pdf.SetXY(oldx, oldy)
	pdf.SetDrawColor(r, g, b)
}

func (e Element) Type() ElementType {
	panic("stub. unsupported")
}

func (e *Element) GetElement() *Element {
	return e
}

func (me *Element) Inherit(parent *Element) {
	me.bgColor = parent.bgColor
	me.fontSize = parent.fontSize
	me.fontStyle = parent.fontStyle
	me.fontColor = parent.fontColor
}

func (me *Element) ResetBorder(width float64) {
	me.borderColor = 0
	me.borderWidth = width
}

type Header struct {
	Element
	Cells []*Cell `json:"cells"`
}

type Page struct {
	Element
	Cells []Cell `json:"cells"`

	PageIndex int // used internally
}

type Footer struct {
	Element
	Cells []Cell `json:"cells"`
}

type Cell struct {
	Element
	Text      string `json:"text"`       // Text to render. Empty string will render a blank box. Use height and width to control size.
	TextAlign string `json:"text-align"` // (L)EFT, (C)ENTER, (R)IGHT, (T)OP, (B)OTTOM, (M)IDDLE, (A)BASELINE

	// Determines where to move line pointer next after rendering this cell
	Display  CellDisplay `json:"display"`
	Width    *Dimension  `json:"width"`    // Width of cell. Omit to use font width
	Height   *Dimension  `json:"height"`   // Height of cell. Omit to use font size
	Absolute bool        `json:"absolute"` // Render cell at an absolute position

	// used internally. Refer to CellFormat
	// Absolute left (x) position of cell
	left float64
	// Absolute top (y) position of the cell
	top float64
}

type Document struct {
	Element
	DisplayUnit          DimensionUnit `json:"units"`                  // Document display units. All numbers will eventually be converted to this unit
	Foot                 Footer        `json:"footer"`                 // Footer section
	Head                 Header        `json:"header"`                 // Header section
	Pages                []Page        `json:"pages"`                  // Pages
	PageBookmarks        bool          `json:"bookmarks"`              // Whether to show page bookmarks
	PageBookmarkTemplate string        `json:"page_bookmark_template"` // Template for all page bookmarks. This will be overriden by each page
	Watermark            struct {
		Text       string      `json:"text"`
		Attributes []Attribute `json:"attributes"`

		// internal use
		fontColor int
		fontSize  Dimension
	} `json:"watermark"` // Document watermark. Will be placed on every page
}

type contextStack[T any] []T

func (s *contextStack[T]) Push(e T) T {
	*s = append(*s, e)
	return e
}

func (s *contextStack[T]) Peek() T {
	l := len(*s)
	a := *s
	val := a[l-1]
	return val
}

func (s *contextStack[T]) Pop() T {
	l := len(*s)
	a := *s
	val := a[l-1]
	*s = a[:l-1]
	return val
}

func (cell *Cell) Render(pdf *gofpdf.Fpdf, section string, icell int, doc *Document, page *Page, it AttributeIterator) error {
	var ctx contextStack[any]

	cellx, celly := pdf.GetXY()
	var cellw, cellh float64

	// defaults
	if page != nil {
		cell.Inherit(&page.Element)
	}

	// parse attributes if iterator is avalable
	if it != nil {
		if err := it("cell", cell.Attrs, cell); err != nil {
			return errors.Wrapf(err, "error in %s, cell %d", section, icell)
		}
	}

	// allow absolute positioning on pages only
	if page != nil && cell.Absolute {
		cellx = cell.left
		celly = cell.top
	}

	// apply cell font size
	if page != nil {
		// page size (needed in case cell font size is percentange)
		pdf.SetFontSize(page.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))
	} else {
		pdf.SetFontSize(doc.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))
	}
	pdf.SetFontSize(cell.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))

	if cell.Width == nil {
		// fallback to string width for the width
		cellw = pdf.GetStringWidth(cell.Text)
	} else {
		cellw = cell.Width.GetValue(pdf, UT_LENGTH|UT_LENGH_WIDTH, doc.DisplayUnit)
	}
	if cell.Height == nil {
		// fallback to
		_, fontHeight := pdf.GetFontSize()
		cellh = fontHeight
	} else {
		cellh = cell.Width.GetValue(pdf, UT_LENGTH|UT_LENGTH_HEIGHT, doc.DisplayUnit)
	}

	if cell.bgColor > 0 {
		oa, bm := pdf.GetAlpha()
		pdf.SetAlpha(alpha(cell.bgColor), "Normal")
		fillRect(pdf, cell.bgColor, cellx, celly, cellw, cellh)
		pdf.SetAlpha(oa, bm)
	}

	// apply cell styles
	ctx.Push(rgb2i(pdf.GetTextColor()))
	pdf.SetFontStyle(cell.fontStyle.String())
	pdf.SetTextColor(rgb(cell.fontColor))

	// save cordinates for future border drawing
	cellx, celly = pdf.GetXY()

	pdf.CellFormat(cellw, cellh, cell.Text, "", int(cell.Display), cell.TextAlign, false, 0, "")

	// draw border
	left := cellx
	top := celly
	right := left + cellw
	bottom := celly + cellh
	cell.drawBorder(pdf, left, top, right, bottom)

	pdf.SetTextColor(rgb(ctx.Pop().(int)))

	return nil
}

func (d Document) Type() ElementType { return DOCUMENT }
func (d Header) Type() ElementType   { return HEADER }
func (d Footer) Type() ElementType   { return FOOTER }
func (d Page) Type() ElementType     { return PAGE }
func (d Cell) Type() ElementType     { return CELL }

func (d *Document) GetElement() *Element { return &d.Element }
func (d *Header) GetElement() *Element   { return &d.Element }
func (d *Footer) GetElement() *Element   { return &d.Element }
func (d *Page) GetElement() *Element     { return &d.Element }
func (d *Cell) GetElement() *Element     { return &d.Element }

// Generate a PDF file from the given template file
func Generate(tplfile string, pdfout string) error {
	fd, err := os.Open(tplfile)
	if err != nil {
		return errors.Wrapf(err, "failed to open template file: %s", tplfile)
	}

	defer fd.Close()

	dec := json.NewDecoder(fd)
	if dec == nil {
		return errors.Wrapf(err, "failed to open")
	}

	var docTemplate Document

	if err = dec.Decode(&docTemplate); err != nil {
		return err
	}

	if len(docTemplate.Pages) == 0 {
		return errors.New("no page data provided")
	}

	return createPdf(&docTemplate, pdfout)
}

func _parseHex(v string) (int, error) {
	l := len(v)
	if l > 1 {
		if v[0] == '#' {
			v = v[1:]
		}
	}
	n, err := strconv.ParseUint(v, 16, 32)
	if err == nil {
		return int(n), nil
	}
	return 0, err
}

func createPdf(doc *Document, pdfout string) error {
	var (
		page_rect rect
		pdf       *gofpdf.Fpdf
		ctx       contextStack[any]
	)

	// validate and initialize document
	if doc.DisplayUnit == DU_PERCENT {
		return errors.New("relative units cannot be used for documents")
	}

	// apply defaults
	doc.fontSize.Unit = doc.DisplayUnit
	doc.fontSize.Value = 12

	pdf = gofpdf.New("P", doc.DisplayUnit.String(), "A4", "")

	// watermark rendering function
	watermarkfn := func(watermark string, fontSize float64, color int) {
		//
		var rc rect

		watermark = strings.Join(strings.Split(watermark, ""), " ") // add space
		oldx, oldy := pdf.GetXY()
		r, g, b := pdf.GetTextColor()
		oldalpha, oldBlendMode := pdf.GetAlpha()
		newalpha := alpha(color)
		ptSize, unitSize := pdf.GetFontSize()
		textWidth := pdf.GetStringWidth(watermark)

		getPageRect(pdf, &rc, true)

		newX := rc.cx() - (textWidth / 2.0)
		newY := rc.cy() + unitSize

		pdf.SetAlpha(newalpha, "Normal")
		pdf.SetFontSize(fontSize)
		pdf.SetTextColor(rgb(color))
		pdf.SetFontStyle("B")
		//pdf.SetXY(rc.cx()-(textWidth/2), rc.cy())
		pdf.SetXY(newX, newY)

		pdf.TransformBegin()
		pdf.TransformRotate(45, rc.cx(), rc.cy())
		pdf.CellFormat(textWidth, unitSize, watermark, "1TBLR", 0, "C", false, 0, "")
		pdf.TransformEnd()

		pdf.SetAlpha(oldalpha, oldBlendMode)
		pdf.SetXY(oldx, oldy)
		pdf.SetTextColor(r, g, b)
		pdf.SetFontSize(ptSize)
	}

	doc.fontColor = rgb2i(pdf.GetTextColor())

	pdf.SetFont("Courier", "", doc.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))

	calculatePageRect(pdf, &page_rect)

	attrWalker := func(prefix string, attrs []Attribute, e IElement, specialHandlerOnly ...bool) error {
		if prefix != "" {
			prefix = prefix + "."
		}
		for _, attr := range attrs {
			var genOk bool
			var genericHandler AttributeHandler

			if len(specialHandlerOnly) == 0 || !specialHandlerOnly[0] {
				genericHandler, genOk = attributeHandlers[attr.Name]
			}

			specialHandler, specOk := attributeHandlers[prefix+attr.Name]
			if genOk {
				if err := genericHandler(pdf, e, attr.Value); err != nil {
					return errors.Wrapf(err, "error parsing attribute value %s", attr.Name)
				}
			}
			if specOk {
				if err := specialHandler(pdf, e, attr.Value); err != nil {
					return errors.Wrapf(err, "error parsing attribute value %s", attr.Name)
				}
			}
		}
		return nil
	}

	// document attributes
	if err := attrWalker("document", doc.Attrs, doc); err != nil {
		return errors.Wrapf(err, "error in document")
	}

	// inherit document font style for watermark
	doc.Watermark.fontSize = doc.fontSize
	doc.Watermark.fontColor = doc.fontColor

	if err := attrWalker("document.watermark", doc.Watermark.Attributes, doc, true); err != nil {
		return errors.Wrapf(err, "error parsing document watermark")
	}

	// apply and save document font size
	// no need to save since document is root
	pdf.SetFontSize(doc.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))

	totalPages := len(doc.Pages)

	// initialize header and footer
	doc.Head.Inherit(&doc.Element)
	doc.Foot.Inherit(&doc.Element)

	if err := attrWalker("header", doc.Head.Attrs, doc); err != nil {
		return errors.Wrapf(err, "error parsing header")
	}

	if err := attrWalker("footer", doc.Foot.Attrs, doc); err != nil {
		return errors.Wrapf(err, "error parsing footer")
	}

	// header function
	pdf.SetHeaderFunc(func() {
		oldx, oldy := pdf.GetXY()
		//drawColor := pdf.GetDrawColor()
		//fontColor := pdf.GetTextColor()
		//fontSize =
		//pdf.SetMargins(10.0, 10.0, 10)
		//pdf.SetXY(10, 10)

		// initialize
		//text := "HEADERRRRRRR!"
		//w := pdf.GetStringWidth(text)
		//pdf.SetFontSize(12)
		//_, lh := pdf.GetFontSize()
		////pdf.SetTextColor(0x33, 0, 0)
		//pdf.CellFormat(w, lh, text, "", 1, "L", false, 0, "")

		fmt.Printf("in header--> x:%.2f, y:%.2f\n", oldx, oldy)
	})

	// footer function
	pdf.SetFooterFunc(func() {
		oldx, oldy := pdf.GetXY()
		left, top, right, bottom := pdf.GetMargins()
		//drawColor := pdf.GetDrawColor()
		//fontColor := pdf.GetTextColor()
		//fontSize =
		var rc rect

		getPageRect(pdf, &rc, true)

		pdf.SetXY(left, bottom+rc.h)

		/* // initialize
		text := "FOOTER!"
		w := pdf.GetStringWidth(text)
		_, lh := pdf.GetFontSize()
		pdf.CellFormat(w, lh, text, "", 1, "ML", false, 0, "")
		pdf.CellFormat(w, lh, text, "", 1, "ML", false, 0, "") */

		for idx, cell := range doc.Foot.Cells {
			cell.Render(pdf, "footer", idx, doc, nil, nil)
		}

		pdf.SetXY(oldx, oldy)

		fmt.Printf("in footer--> x:%.2f, y:%.2f (t:%.2f,r:%.2f,b:%.2f)\n", oldx, oldy, top, right, bottom)
	})

	// bookmark templating function
	bookmarkTitle := func(title string, current, total int) string {
		vars := map[string]string{
			"${page}":  strconv.Itoa(current),
			"${total}": strconv.Itoa(total),
		}
		for k, v := range vars {
			title = strings.ReplaceAll(title, k, v)
		}
		return title
	}

	for i, page := range doc.Pages {
		pdf.AddPage()
		page.PageIndex = i

		// header
		//doc.Head.Render(pdf, &page, doc)
		// footer

		//doc.Foot.Render(pdf, &page, doc)

		// apply document background first before any page adjustments
		if doc.bgColor > 0 {
			var rc rect
			getPageRect(pdf, &rc, false)
			fillRect(pdf, doc.bgColor, rc.x, rc.y, rc.w, rc.h)
		}

		// page defaults (inherit from document)
		page.Inherit(&doc.Element)

		// page attributes
		if err := attrWalker("page", page.Attrs, &page); err != nil {
			return errors.Wrapf(err, "error in page %d", i)
		}

		// render bookmarks
		if doc.PageBookmarks {
			var title string
			if page.bookmarkTitle == "" {
				if doc.PageBookmarkTemplate == "" {
					title = bookmarkTitle("Page ${page}", i+1, totalPages)
				} else {
					title = bookmarkTitle(doc.PageBookmarkTemplate, i+1, totalPages)
				}
			} else {
				title = bookmarkTitle(page.bookmarkTitle, i+1, totalPages)
			}
			pdf.Bookmark(title, 1, -1)
		}

		// apply page font size
		pdf.SetFontSize(ctx.Push(page.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit)).(float64))

		// render cells
		for ic, cell := range page.Cells {
			cellx, celly := pdf.GetXY()
			var cellw, cellh float64

			// defaults
			cell.Inherit(&page.Element)

			// parse attributes
			if err := attrWalker("cell", cell.Attrs, &cell); err != nil {
				return errors.Wrapf(err, "error in page %d, cell %d", i, ic)
			}

			if cell.Absolute {
				cellx = cell.left
				celly = cell.top
			}

			// apply cell font size
			pdf.SetFontSize(ctx.Peek().(float64)) // page size (needed in case cell size is percentange)
			pdf.SetFontSize(cell.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))

			if cell.Width == nil {
				// fallback to string width for the width
				cellw = pdf.GetStringWidth(cell.Text)
			} else {
				cellw = cell.Width.GetValue(pdf, UT_LENGTH|UT_LENGH_WIDTH, doc.DisplayUnit)
			}
			if cell.Height == nil {
				// fallback to
				_, fontHeight := pdf.GetFontSize()
				cellh = fontHeight
			} else {
				cellh = cell.Width.GetValue(pdf, UT_LENGTH|UT_LENGTH_HEIGHT, doc.DisplayUnit)
			}

			if cell.bgColor > 0 {
				fillRect(pdf, cell.bgColor, cellx, celly, cellw, cellh)
			}

			// apply cell styles
			ctx.Push(rgb2i(pdf.GetTextColor()))
			pdf.SetFontStyle(cell.fontStyle.String())
			pdf.SetTextColor(rgb(cell.fontColor))

			// save cordinates for future border drawing
			cellx, celly = pdf.GetXY()

			pdf.CellFormat(cellw, cellh, cell.Text, "", int(cell.Display), cell.TextAlign, false, 0, "")

			// draw border
			left := cellx
			top := celly
			right := left + cellw
			bottom := celly + cellh
			cell.drawBorder(pdf, left, top, right, bottom)

			pdf.SetTextColor(rgb(ctx.Pop().(int)))
		}

		// restore page styles
		pdf.SetFontSize(ctx.Pop().(float64))
		pdf.SetFontStyle(page.fontStyle.String())

		// draw page border
		var br rect
		getPageRect(pdf, &br, true)
		page.drawBorder(pdf, br.x, br.y, br.w+br.x, br.h+br.y*2)
	}

	// restore document styles
	pdf.SetFontStyle(doc.fontStyle.String())
	pdf.SetFontSize(doc.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit))

	// watermark
	if doc.Watermark.Text != "" {
		watermarkfn(
			doc.Watermark.Text,
			doc.Watermark.fontSize.GetValue(pdf, UT_FONT_SIZE, doc.DisplayUnit),
			doc.Watermark.fontColor,
		)
	}

	//invoiceItems(pdf)

	//paintBackground(pdf)
	/* pdf.AddPage()
	invoiceHeader(pdf, &labels)
	invoiceItems(pdf, &items)
	drawBorder(pdf) */
	//drawBorder(pdf)

	/* if !utils.IsEmptyString(watermark) {
		waterMark(pdf, watermark)
	} */

	err := pdf.OutputFileAndClose(pdfout)
	if err != nil {
		return errors.Wrap(err, "error writing pdf file")
	}

	return err
}

type AttributeHandler func(pdf *gofpdf.Fpdf, e IElement, val any) error

var (
	borderColorFn = func(e *Border, val string) (*Border, error) {
		color, err := parseColor(val)
		if err != nil {
			return nil, errors.Wrap(err, "invalid border color value")
		}
		if e == nil {
			e = new(Border)
		}
		e.Color = color
		return e, nil
	}
	borderWidthFn = func(e *Border, val []byte) (*Border, error) {
		w, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, errors.Wrap(err, "invalid border with value")
		}
		if e == nil {
			e = new(Border)
		}
		e.Width = w
		return e, nil
	}
	attributeHandlers map[string]AttributeHandler = map[string]AttributeHandler{
		"background-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			color, err := parseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.bgColor = color
			return nil
		},
		"border-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error { // for all border sides
			color, err := parseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.initBorders()
			doc.setBorderColor(color)
			return nil
		},
		"border-width": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			width, err := strconv.ParseFloat(val.(string), 32)
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.initBorders()
			doc.setBorderWidth(width)
			return nil
		},
		"font-style": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			var oldStyle = doc.fontStyle
			err := doc.fontStyle.UnmarshalText([]byte(val.(string)))
			if err != nil {
				return err
			}
			doc.fontStyle |= oldStyle
			return nil
		},
		"font-size": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			err := doc.fontSize.UnmarshalText([]byte(val.(string)))
			if err != nil {
				return err
			}
			return nil
		},
		"font-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			color, err := parseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.fontColor = color
			return nil
		},
		"document.title": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			pdf.SetTitle(val.(string), true)
			return nil
		},
		"document.watermark.font-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			color, err := parseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.(*Document)
			doc.Watermark.fontColor = color
			return nil
		},
		"document.watermark.font-size": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.(*Document)
			err := doc.Watermark.fontSize.UnmarshalText([]byte(val.(string)))
			if err != nil {
				return err
			}
			return nil
		},
		"border-left-width": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.border.left, []byte(val.(string)))
			doc.border.left = b
			return err
		},
		"border-right-width": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.border.right, []byte(val.(string)))
			doc.border.right = b
			return err
		},
		"border-top-width": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.border.top, []byte(val.(string)))
			doc.border.top = b
			return err
		},
		"border-bottom-width": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.border.bottom, []byte(val.(string)))
			doc.border.bottom = b
			return err
		},
		"border-left-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.border.left, val.(string))
			doc.border.left = b
			return err
		},
		"border-top-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.border.top, val.(string))
			doc.border.top = b
			return err
		},
		"border-right-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.border.right, val.(string))
			doc.border.right = b
			return err
		},
		"border-bottom-color": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.border.bottom, val.(string))
			doc.border.bottom = b
			return err
		},
		"cell.text-align": func(pdf *gofpdf.Fpdf, e IElement, val any) error {
			cell := e.(*Cell)
			cell.TextAlign = val.(string)
			return nil
		},
	}
)
