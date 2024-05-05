package pdfgo

import (
	"github.com/gintec-rdl/pdf-go/internal/impl"
	"github.com/gintec-rdl/pdf-go/internal/pdf"
	"github.com/gintec-rdl/pdf-go/pkg/types"
)

func CreatePdfDocument(orientation types.PageOrientation, pageSize types.PageSize, units types.DimensionUnit) (types.PdfDocument, error) {
	return pdf.NewPdfDocument(orientation, pageSize, units)
}

func CreatePdfDocumentT(t types.PdfTemplate) (types.PdfDocument, error) {
	return pdf.NewPdfDocument(t.GetOrientation(), t.GetPageSize(), t.GetUnit())
}

func CreatePdfTemplateBuilder(orientation types.PageOrientation, pageSize types.PageSize, units types.DimensionUnit) types.PdfTemplateBuilder {
	return impl.NewTemplateBuilder(orientation, pageSize, units)
}

func CreatePdfTemplateLoader() types.PdfTemplateLoader {
	return impl.NewTemplateLoader()
}
