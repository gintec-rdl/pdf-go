package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ElementType int

// Controls how cells are rendered in succession of each other.
type CellDisplay int

type PageSize string
type PageOrientation string

const (
	PAGE_SIZE_A3      = "A3"
	PAGE_SIZE_A4      = "A4"
	PAGE_SIZE_A5      = "A5"
	PAGE_SIZE_LEGAL   = "Legal"
	PAGE_SIZE_LETTER  = "Letter"
	PAGE_SIZE_TABLOID = "Tabloid"

	PO_LANDSCAPE = "L"
	PO_PORTRAIT  = "P"
)

var stringToCellDisplayMap map[string]CellDisplay = map[string]CellDisplay{
	"row":    DISPLAY_ROW,
	"stack":  DISPLAY_STACK,
	"column": DISPLAY_COLUMN,
}

var cellDisplayToStringMap map[CellDisplay]string = map[CellDisplay]string{
	DISPLAY_ROW:    "row",
	DISPLAY_STACK:  "stack",
	DISPLAY_COLUMN: "column",
}

func (me *CellDisplay) Parse(value string) error {
	value = strings.ToLower(value)
	d, ok := stringToCellDisplayMap[value]
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

func (me *CellDisplay) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.String())
}

func (me *CellDisplay) String() string {
	return cellDisplayToStringMap[*me]
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
	FS_REGULAR FontStyle = 1 << iota
	FS_BOLD
	FS_ITALIC
	FS_UNDERLINE
	FS_STRIKETHROUGH
)

func (fs FontStyle) String() string {
	if fs&FS_REGULAR > 0 {
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
	names := strings.Split(string(data), "|")
	if len(names) == 0 {
		return errors.New("empty font style")

	}
	for _, name := range names {
		style, ok := fontStrToFontStyeMap[name]
		if !ok {
			return errors.Errorf("unsupported font style %s", name)
		}
		*fs |= style.Style
	}
	return nil
}

func (fs *FontStyle) MarshalText() ([]byte, error) {
	names := []string{}
	for style, name := range fontStyleToFontNameMap {
		if *fs&style > 0 {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, errors.Errorf("unsupported font style value `%v`", fs)
	}
	return []byte(strings.Join(names, "|")), nil
}

func (fs *FontStyle) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return errors.Wrapf(err, "invalid font style %s", data)
	}
	return fs.UnmarshalText([]byte(name))
}

func (fs *FontStyle) MarshalJSON() ([]byte, error) {
	name, err := fs.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(name))
}

type styleinfo struct {
	Style FontStyle
	Value string
}

var (
	fontStrToFontStyeMap map[string]styleinfo = map[string]styleinfo{
		"regular":       {Style: FS_REGULAR, Value: ""},
		"bold":          {Style: FS_BOLD, Value: "B"},
		"italic":        {Style: FS_ITALIC, Value: "I"},
		"underline":     {Style: FS_UNDERLINE, Value: "U"},
		"strikethrough": {Style: FS_STRIKETHROUGH, Value: "S"},
	}

	fontStyleToFontNameMap map[FontStyle]string = map[FontStyle]string{
		FS_REGULAR:       "regular",
		FS_BOLD:          "bold",
		FS_ITALIC:        "italic",
		FS_UNDERLINE:     "underline",
		FS_STRIKETHROUGH: "strikethrough",
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
	Value         float64
	OriginalValue float64
	Unit          DimensionUnit
}

type Border struct {
	Brush Brush
}

type UnitType int

const (
	UT_LENGTH UnitType = 1 << iota
	UT_FONT_SIZE

	UT_LENGH_WIDTH
	UT_LENGTH_HEIGHT
)

type UnitConverter func(in float64, flags UnitType, w, h, fontSize float64) float64

var (
	// returns a value that is relative to parent's value of the same unit.
	percentageConverter = func(in float64, flags UnitType, w, h, fontSize float64) float64 {
		if flags&UT_FONT_SIZE > 0 {
			in = fontSize * in // relative to parent's font size
		} else if flags&UT_LENGTH > 0 {
			if flags&UT_LENGH_WIDTH > 0 {
				in = in * w // get percentage
			} else if flags&UT_LENGTH_HEIGHT > 0 {
				in = in * h // get percentage
			}
		}

		return in
	}
	conversionTable map[DimensionUnit]map[DimensionUnit]UnitConverter = map[DimensionUnit]map[DimensionUnit]UnitConverter{
		DU_MILIMETER: {
			DU_CENTIMETER: func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in * .1 },
			DU_INCH:       func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in * 0.0393701 },
			DU_MILIMETER:  func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in },
			//DU_PERCENT:    func(in float64, flags UnitType) float64 { return in },
		},
		DU_CENTIMETER: {
			DU_CENTIMETER: func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in },
			DU_INCH:       func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in * 0.393701 },
			DU_MILIMETER:  func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in * 10 },
			//DU_PERCENT:    func(in float64, flags UnitType) float64 { return in },
		},
		DU_INCH: {
			DU_CENTIMETER: func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in * 2.54 },
			DU_INCH:       func(in float64, flags UnitType, w, h, fontSize float64) float64 { return in },
			DU_MILIMETER:  func(in float64, flags UnitType, w, h, fontSize float64) float64 { return 25.4 },
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
func (me Dimension) GetValue(w, h, fontSize float64, flags UnitType, dstUnit DimensionUnit) float64 {
	fn, ok := conversionTable[me.Unit][dstUnit]
	if ok {
		return fn(me.Value, flags, w, h, fontSize)
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

func (d *Dimension) String() string {
	valueStr := strings.TrimSuffix(fmt.Sprintf("%.02f", d.OriginalValue), ".00")
	return valueStr + d.Unit.String()
}

func (d *Dimension) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Dimension) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
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
	me.OriginalValue = value
	me.Value = value / 100.00 // why??
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
	Attrs     []*Attribute `json:"attributes"`
	StyleList []string     `json:"style_list"`

	TextStyle     TextBrush `json:"-"`
	BookmarkTitle string    `json:"bookmark_title"`
	Brush         Brush     `json:"-"`
	Border        struct {
		Left   *Border
		Top    *Border
		Right  *Border
		Bottom *Border
	} `json:"-"`
	Background *Brush `json:"-"`
}

func (e *Element) GetAttributeValue(name string) (string, bool) {
	for _, a := range e.Attrs {
		if a.Name == name {
			return a.Value, true
		}
	}
	return "", false
}

func (e *Element) InitBorders() {
	newBrush := func() *Brush {
		return &Brush{
			Stroke:      true,
			StrokeWidth: .2,
			StrokeColor: Color{
				Alpha: 1,
			},
			CapStyle: CS_CAP,
		}
	}
	if e.Border.Left == nil {
		e.Border.Left = &Border{
			Brush: *newBrush(),
		}
	}
	if e.Border.Right == nil {
		e.Border.Right = &Border{
			Brush: *newBrush(),
		}
	}
	if e.Border.Top == nil {
		e.Border.Top = &Border{
			Brush: *newBrush(),
		}
	}
	if e.Border.Bottom == nil {
		e.Border.Bottom = &Border{
			Brush: *newBrush(),
		}
	}
}

func (e *Element) SetBorderWidth(width float64) {
	e.Border.Top.Brush.StrokeWidth = width
	e.Border.Left.Brush.StrokeWidth = width
	e.Border.Right.Brush.StrokeWidth = width
	e.Border.Bottom.Brush.StrokeWidth = width
}

func (e *Element) SetBorderColor(color int, alpha float64) {
	e.Border.Top.Brush.StrokeColor.Apply(color, alpha)
	e.Border.Left.Brush.StrokeColor.Apply(color, alpha)
	e.Border.Right.Brush.StrokeColor.Apply(color, alpha)
	e.Border.Bottom.Brush.StrokeColor.Apply(color, alpha)
}

func (e Element) DrawBorder(c Canvas, left, top, right, bottom float64) {
	if e.Border.Left != nil {
		c.DrawLine(left, top, left, bottom, &e.Border.Left.Brush)
	}

	if e.Border.Right != nil {
		c.DrawLine(right, top, right, bottom, &e.Border.Right.Brush)
	}

	if e.Border.Top != nil {
		c.DrawLine(left, top, right, top, &e.Border.Top.Brush)
	}

	if e.Border.Bottom != nil {
		c.DrawLine(left, bottom, right, bottom, &e.Border.Bottom.Brush)
	}
}

func (e Element) Type() ElementType {
	panic("stub. unsupported")
}

func (e *Element) GetElement() *Element {
	return e
}

// Inherit parent brush styles
func (me *Element) Inherit(parent *Element) {
	if parent.Background != nil {
		if me.Background == nil {
			me.Background = new(Brush)
			me.Background.Stroke = true
			me.Background.StrokeColor.Apply(0, 1)
		}
		me.Background.Copy(parent.Background)
	}
	me.TextStyle.Copy(&parent.TextStyle)
}

type Header struct {
	Element
	Cells []*Cell `json:"cells"`
}

type Page struct {
	Element
	Cells []*Cell `json:"cells"`

	PageIndex int `json:"-"` // used internally
}

type Footer struct {
	Element
	Cells []*Cell `json:"cells"`
}

type Cell struct {
	Element
	Text string `json:"text"` // Text to render. Empty string will render a blank box. Use height and width to control size.

	Width    *Dimension `json:"-"` // Width of cell. Omit to use font width
	Height   *Dimension `json:"-"` // Height of cell. Omit to use font size
	Absolute bool       `json:"-"` // Render cell at an absolute position`
	Left     float64    `json:"-"` // Left position if absolute
	Top      float64    `json:"-"` // Top position if absolute
}

type Style struct {
	Name       string       `json:"name"`
	Attributes []*Attribute `json:"attributes"`
}

func (s *Style) Apply(e *Element) {
	e.Attrs = append(e.Attrs, s.Attributes...)
}

type FontData struct {
	FilePath string `json:"-"` // will contain the file path if it points to a font file
	Data     string `json:"data"`
}

const MAX_FONT_FILE_SIZE = 500 * 1024

func (fd *FontData) UnmarshalJSON(in []byte) error {
	var dt string
	if err := json.Unmarshal(in, &dt); err != nil {
		return err
	}
	// check for 'file://' protocol directive
	if strings.HasPrefix(dt, "file://") {
		filename := filepath.Clean(dt[7:])

		// load font data
		stat, err := os.Stat(filename)
		if err != nil {
			return errors.Wrapf(err, "failed to stat font file `%s`", filename)
		}
		if stat.Size() > MAX_FONT_FILE_SIZE {
			return errors.Errorf("font file '%s' exceeds '%d' bytes", filepath.Base(filename), MAX_FONT_FILE_SIZE)
		}
		fd.FilePath = filename
		return nil
	} else {
		if len(dt) == 0 {
			return errors.New("missing font data")
		}
	}
	return nil
}

func (fd *FontData) MarshalJSON() ([]byte, error) {
	if fd.FilePath != "" {
		return json.Marshal(fmt.Sprintf("file://%s", fd.FilePath))
	}
	return json.Marshal(fd.Data)
}

type Font struct {
	Data  FontData  `json:"data"`
	Name  string    `json:"name"`
	Style FontStyle `json:"style"`
}

type Document struct {
	Element
	Styles               []*Style        `json:"styles"`
	Fonts                []*Font         `json:"fonts"`
	PageSize             PageSize        `json:"size,omitempty"`         // Document size: (A4,Letter, etc)
	DisplayUnit          DimensionUnit   `json:"units,omitempty"`        // Document display units. All numbers will eventually be converted to this unit
	Orientation          PageOrientation `json:"orientation,omitempty"`  // Orientation: (P)ortrait or (L)andscape
	Foot                 Footer          `json:"footer"`                 // Footer section
	Head                 Header          `json:"header"`                 // Header section
	Pages                []*Page         `json:"pages"`                  // Pages
	PageBookmarks        bool            `json:"bookmarks"`              // Whether to show page bookmarks
	PageBookmarkTemplate string          `json:"page_bookmark_template"` // Template for all page bookmarks. This can be overriden a page
	Watermark            struct {
		Text       string       `json:"text"`
		StyleList  []string     `json:"style_list"`
		Attributes []*Attribute `json:"attributes"`

		// internal use
		TextStyle TextBrush `json:"-"`
	} `json:"watermark"` // Document watermark. Will be placed on every page

	Title string `json:"-"`
}

func (d *Document) HasStyle(name string) bool {
	return slices.ContainsFunc[[]*Style, *Style](d.Styles, func(s *Style) bool {
		return s.Name == name
	})
}

func (d *Document) GetStyleByName(name string) (*Style, bool) {
	index := slices.IndexFunc[[]*Style, *Style](d.Styles, func(s *Style) bool {
		return s.Name == name
	})
	if index >= 0 {
		return d.Styles[index], true
	}
	return nil, false
}

func (cell *Cell) Render(c Canvas, icell int, doc *Document, page *Page, isPageCell bool) {
	cellx, celly := c.GetXY()
	var cellw, cellh float64
	dc := c.GetDrawingRect()

	// allow absolute positioning on pages only
	if isPageCell && cell.Absolute {
		cellx = cell.Left
		celly = cell.Top
	}

	// TODO take into account cell margin

	if cell.Width == nil {
		// fallback to string width for the width
		cellw = c.GetTextWidth(cell.Text)
	} else {
		cellw = cell.Width.GetValue(dc.Right, 0, 0, UT_LENGTH|UT_LENGH_WIDTH, doc.DisplayUnit)
	}
	if cell.Height == nil {
		// fallback to
		cellh = c.GetTextHeight()
	} else {
		cellh = cell.Height.GetValue(0, dc.Bottom, 0, UT_LENGTH|UT_LENGTH_HEIGHT, doc.DisplayUnit)
	}

	rect := Rect{
		Left:   cellx,
		Top:    celly,
		Right:  cellw,
		Bottom: cellh,
	}

	if cell.Background != nil {
		c.DrawRect(rect, cell.Background)
	}

	cellx, celly = c.GetXY()
	c.DrawText(cellw, cellh, cell.Text, &cell.TextStyle)

	// draw border
	cell.DrawBorder(c, cellx, celly, cellx+cellw, celly+cellh)
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

func NewDimension(value float64, unit DimensionUnit) *Dimension {
	return &Dimension{
		Unit:          unit,
		Value:         value,
		OriginalValue: value,
	}
}

func MustParseDimension(str string) *Dimension {
	var d Dimension
	if err := d.UnmarshalText([]byte(str)); err != nil {
		panic(err)
	}
	return &d
}
