package impl

import "github.com/gintec-rdl/pdf-go/pkg/types"

type elementCell[T any, P any] struct {
	parent    P
	self      T
	container container
	cell      *types.Cell
	builder   *BuilderImpl
}

func (c *elementCell[T, P]) Parent() P {
	return c.parent
}

func (c *elementCell[T, P]) Text(text string) T {
	c.cell.Text = text
	return c.self
}

func (c *elementCell[T, P]) Attribute(name, value string) T {
	c.container.Attribute(name, value)
	return c.self
}
func (c *elementCell[T, P]) Attributes(attrs types.PdfTemplateAttributes) T {
	c.container.Attributes(attrs)
	return c.self
}

func (c *elementCell[T, P]) StyleList(name string, more ...string) T {
	if c.cell.StyleList == nil {
		c.cell.StyleList = make([]string, 0)
	}
	c.cell.StyleList = append(c.cell.StyleList, name)
	c.cell.StyleList = append(c.cell.StyleList, more...)
	return c.self
}
