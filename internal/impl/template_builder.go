package impl

import (
	"slices"
	"strings"

	"github.com/gintec-rdl/pdf-go/pkg/types"
	"github.com/pkg/errors"
)

type container struct {
	builder    *BuilderImpl
	attributes *[]*types.Attribute
}

func (c *container) Attribute(name, value string) types.PdfTemplateAttributeContainer {
	index := slices.IndexFunc(*c.attributes, func(attr *types.Attribute) bool {
		return strings.EqualFold(attr.Name, name)
	})
	if index != -1 {
		attr := *c.attributes
		attr[index].Value = value
	} else {
		*c.attributes = append(*c.attributes, &types.Attribute{Name: name, Value: value})
	}
	return c
}
func (c *container) Attributes(attrs types.PdfTemplateAttributes) types.PdfTemplateAttributeContainer {
	for name, value := range attrs {
		c.Attribute(name, value)
	}
	return c
}
func (c *container) Builder() types.PdfTemplateBuilder {
	return c.builder
}

type BuilderImpl struct {
	container container
	document  types.Document
}

func NewTemplateBuilder(orientation types.PageOrientation, pageSize types.PageSize, units types.DimensionUnit) types.PdfTemplateBuilder {
	builder := &BuilderImpl{}
	builder.document.DisplayUnit = units
	builder.document.PageSize = pageSize
	builder.document.Orientation = orientation
	builder.document.Attrs = make([]*types.Attribute, 0)
	builder.container.attributes = &builder.document.Attrs
	builder.container.builder = builder
	return builder
}

func (c *BuilderImpl) Attribute(name, value string) types.PdfTemplateBuilder {
	return c
}
func (c *BuilderImpl) Attributes(attrs types.PdfTemplateAttributes) types.PdfTemplateBuilder {
	c.container.Attributes(attrs)
	return c
}

func (b *BuilderImpl) Builder() types.PdfTemplateBuilder {
	return b
}

func (b *BuilderImpl) Build() (types.PdfTemplate, error) {
	if b.document.Pages == nil || len(b.document.Pages) == 0 {
		return nil, errors.New("document needs at least one page")
	}
	if b.document.Head.Attrs == nil {
		b.document.Head.Attrs = make([]*types.Attribute, 0)
	}
	if b.document.Head.Cells == nil {
		b.document.Head.Cells = make([]*types.Cell, 0)
	}
	if b.document.Foot.Attrs == nil {
		b.document.Foot.Attrs = make([]*types.Attribute, 0)
	}
	if b.document.Foot.Cells == nil {
		b.document.Foot.Cells = make([]*types.Cell, 0)
	}
	if b.document.Styles == nil {
		b.document.Styles = make([]*types.Style, 0)
	}

	if err := validateDocument(&b.document); err != nil {
		return nil, err
	}

	return &PdfTemplateImpl{document: b.document}, nil
}

func (b *BuilderImpl) Title(title string) types.PdfTemplateBuilder {
	b.container.Attribute("title", title)
	return b
}

func (b *BuilderImpl) StyleList(name string, more ...string) types.PdfTemplateBuilder {
	if b.document.StyleList == nil {
		b.document.StyleList = make([]string, 0)
	}
	b.document.StyleList = append(b.document.StyleList, name)
	b.document.StyleList = append(b.document.StyleList, more...)
	return b
}

func (b *BuilderImpl) Style(name string, attrs types.PdfTemplateAttributes) types.PdfTemplateBuilder {
	if b.document.Styles == nil {
		b.document.Styles = make([]*types.Style, 0)
	}
	style, ok := b.document.GetStyleByName(name)
	if !ok {
		style = new(types.Style)
		style.Name = name
		style.Attributes = make([]*types.Attribute, 0)
		b.document.Styles = append(b.document.Styles, style)
	}
	for k, v := range attrs {
		attr := new(types.Attribute)
		attr.Name = k
		attr.Value = v
		style.Attributes = append(style.Attributes, attr)
	}
	return b
}

func (b *BuilderImpl) ShowBookmarks(show bool) types.PdfTemplateBuilder {
	b.document.PageBookmarks = show
	return b
}

func (b *BuilderImpl) PageBookmarkTemplate(template string) types.PdfTemplateBuilder {
	b.document.PageBookmarkTemplate = template
	return b
}

func (b *BuilderImpl) AddFontFromFile(fontFamily string, style types.FontStyle, filepath string) types.PdfTemplateBuilder {
	if b.document.Fonts == nil {
		b.document.Fonts = make([]*types.Font, 0)
	}
	font := &types.Font{}
	font.Style = style
	font.Name = fontFamily
	font.Data.FilePath = filepath
	b.document.Fonts = append(b.document.Fonts, font)
	return b
}

func (b *BuilderImpl) AddPage() types.PdfTemplatePage {
	var newPage = new(types.Page)

	newPage.Cells = make([]*types.Cell, 0)
	newPage.Attrs = make([]*types.Attribute, 0)

	b.document.Pages = append(b.document.Pages, newPage)
	page := &pageImpl{page: newPage}
	page.self = page
	page.container.builder = b
	page.container.attributes = &newPage.Attrs
	return page
}

func (b *BuilderImpl) Header() types.PdfTemplateHeader {
	header := &headerImpl{header: &b.document.Head}
	if b.document.Head.Attrs == nil {
		b.document.Head.Attrs = make([]*types.Attribute, 0)
	}
	if b.document.Head.Cells == nil {
		b.document.Head.Cells = make([]*types.Cell, 0)
	}
	header.self = header
	header.container.builder = b
	header.container.attributes = &b.document.Head.Attrs
	return header
}
func (b *BuilderImpl) Footer() types.PdfTemplateFooter {
	footer := &footerImpl{footer: &b.document.Foot}
	if b.document.Foot.Attrs == nil {
		b.document.Foot.Attrs = make([]*types.Attribute, 0)
	}
	if b.document.Foot.Cells == nil {
		b.document.Foot.Cells = make([]*types.Cell, 0)
	}
	footer.self = footer
	footer.container.builder = b
	footer.container.attributes = &b.document.Foot.Attrs
	return footer
}
func (b *BuilderImpl) Watermark(text string) types.PdfTemplateWatermark {
	b.document.Watermark.Text = text
	if b.document.Watermark.Attributes == nil {
		b.document.Watermark.Attributes = make([]*types.Attribute, 0)
	}
	wm := &watermarkImpl{builder: b, doc: &b.document}
	wm.container.attributes = &b.document.Watermark.Attributes
	wm.container.builder = b
	return wm
}
