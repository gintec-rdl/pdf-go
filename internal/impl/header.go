package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type headerImpl struct {
	elementImpl[types.PdfTemplateHeader]
	header *types.Header
}

type headerCell struct {
	elementCell[types.PdfTemplateHeaderCell, types.PdfTemplateHeader]
}

func (h *headerImpl) AddCell() types.PdfTemplateHeaderCell {
	var newCell types.Cell
	h.header.Cells = append(h.header.Cells, &newCell)
	cell := &headerCell{
		elementCell[types.PdfTemplateHeaderCell, types.PdfTemplateHeader]{
			self:    nil,
			parent:  h,
			builder: h.container.builder,
		},
	}
	cell.self = cell
	cell.container.builder = h.container.builder
	cell.container.attributes = &newCell.Attrs
	return cell
}

func (h *headerImpl) StyleList(name string, more ...string) types.PdfTemplateHeader {
	if h.header.StyleList == nil {
		h.header.StyleList = make([]string, 0)
	}
	h.header.StyleList = append(h.header.StyleList, name)
	h.header.StyleList = append(h.header.StyleList, more...)
	return h
}
