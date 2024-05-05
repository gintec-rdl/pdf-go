package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type elementImpl[Self any] struct {
	self      Self
	container container
}

func (e *elementImpl[Self]) Attribute(name, value string) Self {
	e.container.Attribute(name, value)
	return e.self
}

func (e *elementImpl[Self]) Attributes(attrs types.PdfTemplateAttributes) Self {
	e.container.Attributes(attrs)
	return e.self
}

func (e *elementImpl[Self]) Builder() types.PdfTemplateBuilder {
	return e.container.builder
}