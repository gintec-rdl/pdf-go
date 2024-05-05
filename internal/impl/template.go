package impl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gintec-rdl/pdf-go/pkg/types"
)

type PdfTemplateImpl struct {
	document types.Document
}

func (tpl *PdfTemplateImpl) GetOrientation() types.PageOrientation {
	return tpl.document.Orientation
}

func (tpl *PdfTemplateImpl) GetPageSize() types.PageSize {
	return tpl.document.PageSize
}

func (tpl *PdfTemplateImpl) GetUnit() types.DimensionUnit {
	return tpl.document.DisplayUnit
}

func (tpl PdfTemplateImpl) RenderF(pdfDoc types.PdfDocument, filename string) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()
	return tpl.RenderW(pdfDoc, fd)
}

func (tpl PdfTemplateImpl) Get() ([]byte, error) {
	var bb bytes.Buffer
	err := tpl.SaveW(&bb)
	if err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func (tpl PdfTemplateImpl) SaveF(filename string) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()
	return tpl.SaveW(fd)
}

func (tpl PdfTemplateImpl) SaveW(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "")
	return encoder.Encode(&tpl.document)
}

func (tpl PdfTemplateImpl) RenderW(pdfDoc types.PdfDocument, w io.Writer) error {
	// set document title, bookmarks, etc
	pdfDoc.SetTitle(tpl.document.Title)

	// setup fonts
	if err := pdfDoc.InitializeFonts(&tpl.document.Fonts); err != nil {
		return err
	}

	// bookmark templating function
	type BookmarkTitleResolver func(key string, page *types.Page) string
	titleTemplateMap := func(title string, total int, page *types.Page) string {
		vars := map[string]BookmarkTitleResolver{
			"${page}":  func(key string, page *types.Page) string { return fmt.Sprintf("%04d", page.PageIndex+1) },
			"${total}": func(key string, page *types.Page) string { return fmt.Sprintf("%04d", total) },
		}
		for k, v := range vars {
			title = strings.ReplaceAll(title, k, v("", page))
		}
		return title
	}

	totalPages := len(tpl.document.Pages)

	// footer and header render
	footerNHeaderFn := func(p types.PdfPage, pageIndex int, inFooter bool) {
		var cells *[]*types.Cell
		pageSource := tpl.document.Pages[pageIndex]

		c := p.GetCanvas()

		if inFooter {
			// footer
			cells = &tpl.document.Foot.Cells
		} else {
			// header
			cells = &tpl.document.Head.Cells
		}

		rc := c.GetDrawingRect()

		for i, cell := range *cells {
			x := c.GetX() // cache X because .SetX resets 'X' coordinate
			if inFooter {
			} else {
				// center vertically in the margin area
				c.SetY((rc.Top * .5) - (c.GetTextHeight() * .5))
			}
			c.SetX(x)
			ogTxt := cell.Text
			cell.Text = titleTemplateMap(ogTxt, totalPages, pageSource)
			cell.Render(c, i, &tpl.document, pageSource, false)
			cell.Text = ogTxt
		}
	}

	for _, page := range tpl.document.Pages {
		pdfPage := pdfDoc.AddNewPage(footerNHeaderFn)

		c := pdfPage.GetCanvas()
		dc := c.GetDrawingRect()

		if tpl.document.PageBookmarks {
			// TODO Create bookmarks
			var title string
			if page.BookmarkTitle == "" {
				if tpl.document.PageBookmarkTemplate == "" {
					title = titleTemplateMap("Page ${page}", totalPages, page)
				} else {
					title = titleTemplateMap(tpl.document.PageBookmarkTemplate, totalPages, page)
				}
			} else {
				title = titleTemplateMap(page.BookmarkTitle, totalPages, page)
			}
			pdfDoc.SetBookmark(title)
		}

		// background
		if tpl.document.Background != nil {
			c.DrawRect(*dc, tpl.document.Background)
		}

		// TODO footer and header

		// page cells
		for j, cell := range page.Cells {
			cell.Render(c, j, &tpl.document, page, true)
		}

		dc = c.GetDrawingRect()

		// draw border
		page.DrawBorder(c, dc.Left, dc.Top, dc.Right+dc.Left, dc.Bottom+dc.Top)
	}

	return pdfDoc.Save(w)
}
