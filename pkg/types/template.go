package types

import (
	"io"
)

type PdfTemplate interface {
	Get() ([]byte, error)
	GetPageSize() PageSize
	GetUnit() DimensionUnit
	SaveW(w io.Writer) error
	SaveF(filename string) error
	GetOrientation() PageOrientation
	RenderW(pdfDoc PdfDocument, w io.Writer) error
	RenderF(pdfDoc PdfDocument, filename string) error
}

type PdfTemplateAttributes map[string]string

type PdfTemplateAttributeContainer interface {
	Builder() PdfTemplateBuilder
	Attribute(name, value string) PdfTemplateAttributeContainer
	Attributes(attrs PdfTemplateAttributes) PdfTemplateAttributeContainer
}

type PdfTemplateHeaderCell interface {
	Parent() PdfTemplateHeader
	Text(text string) PdfTemplateHeaderCell
	Attribute(name, value string) PdfTemplateHeaderCell
	Attributes(attrs PdfTemplateAttributes) PdfTemplateHeaderCell
	StyleList(name string, more ...string) PdfTemplateHeaderCell
}

type PdfTemplateFooterCell interface {
	Parent() PdfTemplateFooter
	Text(text string) PdfTemplateFooterCell
	Attribute(name, value string) PdfTemplateFooterCell
	Attributes(attrs PdfTemplateAttributes) PdfTemplateFooterCell
	StyleList(name string, more ...string) PdfTemplateFooterCell
}

type PdfTemplatePageCell interface {
	Parent() PdfTemplatePage
	Text(text string) PdfTemplatePageCell
	Attribute(name, value string) PdfTemplatePageCell
	Attributes(attrs PdfTemplateAttributes) PdfTemplatePageCell
	StyleList(name string, more ...string) PdfTemplatePageCell
}

type PdfTemplateHeader interface {
	Builder() PdfTemplateBuilder
	AddCell() PdfTemplateHeaderCell
	Attribute(name, value string) PdfTemplateHeader
	Attributes(attrs PdfTemplateAttributes) PdfTemplateHeader
	StyleList(name string, more ...string) PdfTemplateHeader
}

type PdfTemplateFooter interface {
	Builder() PdfTemplateBuilder
	AddCell() PdfTemplateFooterCell
	Attribute(name, value string) PdfTemplateFooter
	Attributes(attrs PdfTemplateAttributes) PdfTemplateFooter

	StyleList(name string, more ...string) PdfTemplateFooter
}

type PdfTemplatePage interface {
	AddCell() PdfTemplatePageCell
	Builder() PdfTemplateBuilder
	Attribute(name, value string) PdfTemplatePage
	BookmarkTitle(bookmark string) PdfTemplatePage
	StyleList(name string, more ...string) PdfTemplatePage
	Attributes(attrs PdfTemplateAttributes) PdfTemplatePage
}

type PdfTemplateWatermark interface {
	Builder() PdfTemplateBuilder
	Attribute(name, value string) PdfTemplateWatermark
	Attributes(attrs PdfTemplateAttributes) PdfTemplateWatermark

	StyleList(name string, more ...string) PdfTemplateWatermark
}

type PdfTemplateBuilder interface {
	AddPage() PdfTemplatePage
	Header() PdfTemplateHeader
	Footer() PdfTemplateFooter
	Build() (PdfTemplate, error)
	Title(title string) PdfTemplateBuilder
	Style(name string, attrs PdfTemplateAttributes) PdfTemplateBuilder
	StyleList(name string, more ...string) PdfTemplateBuilder
	ShowBookmarks(show bool) PdfTemplateBuilder
	PageBookmarkTemplate(template string) PdfTemplateBuilder
	AddFontFromFile(fontFamily string, style FontStyle, filepath string) PdfTemplateBuilder
	Watermark(text string) PdfTemplateWatermark
	Attribute(name, value string) PdfTemplateBuilder
	Attributes(attrs PdfTemplateAttributes) PdfTemplateBuilder
}

type PdfTemplateLoader interface {
	LoadR(r io.Reader) (PdfTemplate, error)
	LoadF(filename string) (PdfTemplate, error)
}
