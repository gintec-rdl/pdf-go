package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type pageImpl struct {
	elementImpl[types.PdfTemplatePage]
	page *types.Page
}

type pageCell struct {
	elementCell[types.PdfTemplatePageCell, types.PdfTemplatePage]
}

func (p *pageImpl) AddCell() types.PdfTemplatePageCell {
	var newCell types.Cell
	p.page.Cells = append(p.page.Cells, &newCell)
	cell := &pageCell{
		elementCell[types.PdfTemplatePageCell, types.PdfTemplatePage]{
			parent:  p,
			cell:    &newCell,
			builder: p.container.builder,
		},
	}
	cell.self = cell
	cell.container.builder = p.container.builder
	cell.container.attributes = &newCell.Attrs
	return cell
}

func (p *pageImpl) BookmarkTitle(bookmark string) types.PdfTemplatePage {
	p.page.BookmarkTitle = bookmark
	return p
}

func (p *pageImpl) StyleList(name string, more ...string) types.PdfTemplatePage {
	if p.page.StyleList == nil {
		p.page.StyleList = make([]string, 0)
	}
	p.page.StyleList = append(p.page.StyleList, name)
	p.page.StyleList = append(p.page.StyleList, more...)
	return p
}
