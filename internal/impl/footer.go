package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type footerImpl struct {
	elementImpl[types.PdfTemplateFooter]
	footer *types.Footer
}

type footerCell struct {
	elementCell[types.PdfTemplateFooterCell, types.PdfTemplateFooter]
}

func (f *footerImpl) AddCell() types.PdfTemplateFooterCell {
	var newCell types.Cell
	f.footer.Cells = append(f.footer.Cells, &newCell)
	cell := &footerCell{
		elementCell[types.PdfTemplateFooterCell, types.PdfTemplateFooter]{
			parent:  f,
			cell:    &newCell,
			builder: f.container.builder,
		},
	}
	cell.self = cell
	cell.container.builder = f.container.builder
	cell.container.attributes = &newCell.Attrs
	return cell
}

func (f *footerImpl) StyleList(name string, more ...string) types.PdfTemplateFooter {
	if f.footer.StyleList == nil {
		f.footer.StyleList = make([]string, 0)
	}
	f.footer.StyleList = append(f.footer.StyleList, name)
	f.footer.StyleList = append(f.footer.StyleList, more...)
	return f
}
