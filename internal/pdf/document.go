package pdf

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/gintec-rdl/pdf-go/pkg/types"
	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"
)

type PdfDocumentImpl struct {
	_pdf                  *gofpdf.Fpdf
	currPage              *PdfPageImpl
	sectionFuncsInstalled bool
}

type PdfPageImpl struct {
	_pdf *gofpdf.Fpdf
}

func (d *PdfDocumentImpl) AddNewPage(footerNHeaderFn func(p types.PdfPage, pageIndex int, inFooter bool)) types.PdfPage {
	d.currPage = &PdfPageImpl{_pdf: d._pdf}
	if !d.sectionFuncsInstalled {
		d.sectionFuncsInstalled = true
		d._pdf.SetHeaderFuncMode(func() {
			if footerNHeaderFn != nil {
				footerNHeaderFn(d.currPage, d._pdf.PageNo()-1, false)
			}
		}, true)

		d._pdf.SetFooterFuncLpi(func(lastPage bool) {
			if footerNHeaderFn != nil {
				footerNHeaderFn(d.currPage, d._pdf.PageNo()-1, true)
			}
		})
	}
	d._pdf.AddPage() // above callbacks are called by this function
	return d.currPage
}

func (d *PdfDocumentImpl) SetTitle(title string) {
	d._pdf.SetTitle(title, true)
}

func (d *PdfDocumentImpl) SetBookmark(title string) {
	d._pdf.Bookmark(title, -1, -1)
}

func (d *PdfDocumentImpl) GetPage(ipage int) (types.PdfPage, bool) {
	if ipage <= d._pdf.PageCount() {
		d._pdf.SetPage(ipage)
		return &PdfPageImpl{_pdf: d._pdf}, true
	}
	return nil, false
}

func (d *PdfDocumentImpl) GetPageCount() int {
	return d._pdf.PageCount()
}

func (d *PdfDocumentImpl) AddFont(fontname string, data []byte) {
	d._pdf.AddFontFromBytes(fontname, "", nil, data)
}

func (d *PdfDocumentImpl) SaveAndCloseF(dst string) error {
	return d._pdf.OutputFileAndClose(dst)
}

func (d *PdfDocumentImpl) SaveAndCloseW(w io.WriteCloser) error {
	return d._pdf.OutputAndClose(w)
}

func (d *PdfDocumentImpl) Save(w io.Writer) error {
	return d._pdf.Output(w)
}

func (d *PdfDocumentImpl) InitializeFonts(fonts *[]*types.Font) error {
	loadf := func(font *types.Font) error {
		if font.Data.FilePath != "" {
			fd, err := os.Open(font.Data.FilePath)
			if err != nil {
				return errors.Wrapf(err, "open font file `%s`", filepath.Base(font.Data.FilePath))
			}
			defer fd.Close()
			bytes, err := io.ReadAll(fd)
			if err != nil {
				return errors.Wrapf(err, "read font file `%s`", filepath.Base(font.Data.FilePath))
			}
			// Doesn't report errors immediately. Errors are reported when saving document 😕
			d._pdf.AddUTF8FontFromBytes(font.Name, font.Style.String(), bytes)
			return nil
		} else {
			bytes, err := hex.DecodeString(font.Data.Data)
			if err != nil {
				return errors.Wrapf(err, "decode font `%s`", font.Name)
			}
			d._pdf.AddUTF8FontFromBytes(font.Name, font.Style.String(), bytes)
			return nil
		}
	}
	for _, font := range *fonts {
		err := loadf(font)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PdfPageImpl) GetCanvas() types.Canvas {
	return &PdfCanvas{_pdf: p._pdf}
}

func NewPdfDocument(orientation types.PageOrientation, pageSize types.PageSize, units types.DimensionUnit) (types.PdfDocument, error) {
	if units == types.DU_PERCENT {
		return nil, errors.New("'%' unit cannot be used at document root level")
	}
	pdf := gofpdf.New(string(orientation), units.String(), string(pageSize), "")
	pdf.SetFont("courier", "", 12)
	return &PdfDocumentImpl{_pdf: pdf}, nil
}
