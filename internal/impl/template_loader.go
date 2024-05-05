package impl

import (
	"encoding/json"
	"io"
	"os"

	"github.com/gintec-rdl/pdf-go/pkg/types"
	"github.com/pkg/errors"
)

type PdfTemplateLoaderImpl struct {
}

// Returns
func (l *PdfTemplateLoaderImpl) LoadF(filename string) (types.PdfTemplate, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return l.LoadR(fd)
}

func (l *PdfTemplateLoaderImpl) LoadR(r io.Reader) (types.PdfTemplate, error) {
	dec := json.NewDecoder(r)

	var doc types.Document

	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}

	if err := validateDocument(&doc); err != nil {
		return nil, err
	}

	return &PdfTemplateImpl{document: doc}, nil
}

func NewTemplateLoader() types.PdfTemplateLoader {
	return &PdfTemplateLoaderImpl{}
}

func validateDocument(doc *types.Document) error {
	if len(doc.Pages) == 0 {
		return errors.New("no page data provided")
	}

	// validate and initialize document
	if doc.DisplayUnit == types.DU_PERCENT {
		return errors.New("relative units cannot be used at the document level")
	}

	// apply defaults
	doc.TextStyle.FontSize.Value = 12
	doc.TextStyle.FontSize.Unit = types.DU_MILIMETER
	doc.TextStyle.Alignment = "LM"
	doc.TextStyle.Brush.Fill = false
	doc.TextStyle.Brush.Stroke = true
	doc.TextStyle.FontName = "courier"
	doc.TextStyle.Brush.StrokeWidth = 0.2
	doc.TextStyle.StrokeColor.Apply(0, 1)
	doc.TextStyle.Brush.FillColor.Apply(0, 0)
	doc.TextStyle.CapStyle = types.CS_CAP
	doc.TextStyle.JoinStyle = types.JS_MITER
	doc.Brush.CapStyle = types.CS_CAP
	doc.Brush.JoinStyle = types.JS_MITER
	doc.Brush.Stroke = true
	doc.Brush.StrokeColor.Apply(0, 1)
	doc.Brush.StrokeWidth = .2

	attrWalker := func(prefix string, attrs []*types.Attribute, e types.IElement, parent types.IElement, specialHandlerOnly ...bool) error {
		if prefix != "" {
			prefix = prefix + "."
		}
		// combine in order of (defined list, inline list) to allow overriding
		combinedAttributes := []*types.Attribute{}
		styles := e.GetElement().StyleList
		for _, styleName := range styles {
			style, ok := doc.GetStyleByName(styleName)
			if !ok {
				return errors.Errorf("`%s`: style does not exist", styleName)
			}
			combinedAttributes = append(combinedAttributes, style.Attributes...)
		}
		combinedAttributes = append(combinedAttributes, attrs...)
		for _, attr := range combinedAttributes {
			var genOk bool
			var genericHandler AttributeHandler

			if len(specialHandlerOnly) == 0 || !specialHandlerOnly[0] {
				genericHandler, genOk = attributeHandlers[attr.Name]
			}

			specialHandler, specOk := attributeHandlers[prefix+attr.Name]
			if genOk {
				if err := genericHandler(e, parent, attr.Value); err != nil {
					return errors.Wrapf(err, "attribute `%s`", attr.Name)
				}
			}
			if specOk {
				if err := specialHandler(e, parent, attr.Value); err != nil {
					return errors.Wrapf(err, "attribute `%s`", attr.Name)
				}
			}
			if !genOk && !specOk {
				return errors.Errorf("unsupported attribute `%s`", attr.Name)
			}
		}
		return nil
	}

	// document attributes
	if err := attrWalker("document", doc.Attrs, doc, nil); err != nil {
		return errors.Wrapf(err, "error in document")
	}

	// inherit document font style for watermark
	doc.Watermark.TextStyle.Copy(&doc.TextStyle)

	if err := attrWalker("document.watermark", doc.Watermark.Attributes, doc, nil, true); err != nil {
		return errors.Wrapf(err, "error in document watermark")
	}

	// initialize header and footer
	doc.Head.Inherit(&doc.Element)
	doc.Foot.Inherit(&doc.Element)

	if err := attrWalker("header", doc.Head.Attrs, &doc.Head, doc); err != nil {
		return errors.Wrapf(err, "error in header")
	}

	// header cells
	for i, hc := range doc.Head.Cells {
		hc.Inherit(&doc.Head.Element)
		if err := attrWalker("cell", hc.Attrs, hc, &doc.Head); err != nil {
			return errors.Wrapf(err, "error in header cell %d", i)
		}
	}

	if err := attrWalker("footer", doc.Foot.Attrs, &doc.Foot, doc); err != nil {
		return errors.Wrapf(err, "error in footer")
	}
	for i, fc := range doc.Foot.Cells {
		fc.Inherit(&doc.Foot.Element)
		if err := attrWalker("cell", fc.Attrs, fc, &doc.Foot); err != nil {
			return errors.Wrapf(err, "error in footer cell %d", i)
		}
	}

	for i, page := range doc.Pages {
		page.PageIndex = i

		// page defaults (inherit from document)
		page.Inherit(&doc.Element)

		// page attributes
		if err := attrWalker("page", page.Attrs, page, doc); err != nil {
			return errors.Wrapf(err, "error in page %d", i)
		}

		// render cells
		for ic, cell := range page.Cells {
			// defaults
			cell.Inherit(&page.Element)

			// parse attributes
			if err := attrWalker("cell", cell.Attrs, cell, page); err != nil {
				return errors.Wrapf(err, "error in cell %d of page %d", ic, i)
			}
		}
	}
	return nil
}
