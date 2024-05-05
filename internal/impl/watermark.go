package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type watermarkImpl struct {
	doc       *types.Document
	container container
	builder   *BuilderImpl
}

func (w *watermarkImpl) Attribute(name, value string) types.PdfTemplateWatermark {
	w.container.Attribute(name, value)
	return w
}
func (w *watermarkImpl) Attributes(attrs types.PdfTemplateAttributes) types.PdfTemplateWatermark {
	w.container.Attributes(attrs)
	return w
}

func (w *watermarkImpl) StyleList(name string, more ...string) types.PdfTemplateWatermark {
	if w.doc.Watermark.StyleList == nil {
		w.doc.Watermark.StyleList = make([]string, 0)
	}
	w.doc.Watermark.StyleList = append(w.doc.Watermark.StyleList, name)
	w.doc.Watermark.StyleList = append(w.doc.Watermark.StyleList, more...)
	return w
}

func (w *watermarkImpl) Builder() types.PdfTemplateBuilder {
	return w.builder
}
