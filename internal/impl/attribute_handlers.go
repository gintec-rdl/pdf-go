package impl

import (
	"slices"
	"strconv"
	"strings"

	"github.com/gintec-rdl/pdf-go/internal/utils"
	"github.com/gintec-rdl/pdf-go/pkg/types"
	"github.com/pkg/errors"
)

type AttributeHandler func(e types.IElement, parent types.IElement, val any) error

var (
	borderColorFn = func(e *types.Border, val string) (*types.Border, error) {
		alpha, color, err := utils.ParseColor(val)
		if err != nil {
			return nil, errors.Wrap(err, "invalid border color value")
		}
		if e == nil {
			e = new(types.Border)
			e.Brush.Stroke = true
		}
		e.Brush.StrokeColor.Apply(color, alpha)
		return e, nil
	}
	borderWidthFn = func(e *types.Border, val []byte) (*types.Border, error) {
		w, err := strconv.ParseFloat(string(val), 32)
		if err != nil {
			return nil, errors.Wrap(err, "invalid border width value")
		}
		if e == nil {
			e = new(types.Border)
			e.Brush.Stroke = true
		}
		e.Brush.StrokeWidth = w
		return e, nil
	}
	attributeHandlers map[string]AttributeHandler = map[string]AttributeHandler{
		"background-color": func(e types.IElement, parent types.IElement, val any) error {
			alpha, color, err := utils.ParseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			if doc.Background == nil {
				doc.Background = new(types.Brush)
				doc.Background.Fill = true
			}
			doc.Background.FillColor.Apply(color, alpha)
			return nil
		},
		"border-color": func(e types.IElement, parent types.IElement, val any) error { // for all border sides
			alpha, color, err := utils.ParseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.InitBorders()
			doc.SetBorderColor(color, alpha)
			return nil
		},
		"border-width": func(e types.IElement, parent types.IElement, val any) error {
			width, err := strconv.ParseFloat(val.(string), 32)
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.InitBorders()
			doc.SetBorderWidth(width)
			return nil
		},
		"font-style": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			return doc.TextStyle.FontStyle.UnmarshalText([]byte(val.(string)))
		},
		"font-size": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			err := doc.TextStyle.FontSize.UnmarshalText([]byte(val.(string)))
			if err != nil {
				return err
			}
			return nil
		},
		"font-color": func(e types.IElement, parent types.IElement, val any) error {
			alpha, color, err := utils.ParseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.GetElement()
			doc.TextStyle.StrokeColor.Apply(color, alpha)
			return nil
		},
		"document.title": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.(*types.Document)
			doc.Title = val.(string)
			return nil
		},
		"document.watermark.font-color": func(e types.IElement, parent types.IElement, val any) error {
			alpha, color, err := utils.ParseColor(val.(string))
			if err != nil {
				return err
			}
			doc := e.(*types.Document)
			doc.Watermark.TextStyle.StrokeColor.Apply(color, alpha)
			return nil
		},
		"document.watermark.font-size": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.(*types.Document)
			err := doc.Watermark.TextStyle.FontSize.UnmarshalText([]byte(val.(string)))
			if err != nil {
				return err
			}
			return nil
		},
		"border-left-width": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.Border.Left, []byte(val.(string)))
			doc.Border.Left = b
			return err
		},
		"border-right-width": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.Border.Right, []byte(val.(string)))
			doc.Border.Right = b
			return err
		},
		"border-top-width": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.Border.Top, []byte(val.(string)))
			doc.Border.Top = b
			return err
		},
		"border-bottom-width": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderWidthFn(doc.Border.Bottom, []byte(val.(string)))
			doc.Border.Bottom = b
			return err
		},
		"border-left-color": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.Border.Left, val.(string))
			doc.Border.Left = b
			return err
		},
		"border-top-color": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.Border.Top, val.(string))
			doc.Border.Top = b
			return err
		},
		"border-right-color": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.Border.Right, val.(string))
			doc.Border.Right = b
			return err
		},
		"border-bottom-color": func(e types.IElement, parent types.IElement, val any) error {
			doc := e.GetElement()
			b, err := borderColorFn(doc.Border.Bottom, val.(string))
			doc.Border.Bottom = b
			return err
		},
		"cell.text-align": func(e types.IElement, parent types.IElement, val any) error {
			cell := e.(*types.Cell)
			table := []rune("LCRBATM")
			values := strings.ToUpper(val.(string))
			for _, r := range values {
				if !slices.Contains(table, r) {
					return errors.Errorf("invalid text alignment flag `%c`. expected any of `LCRBATM`", r)
				}
			}
			cell.TextStyle.Alignment = val.(string)
			return nil
		},
		"cell.width": func(e types.IElement, parent types.IElement, val any) error {
			cell := e.(*types.Cell)
			if cell.Width == nil {
				cell.Width = new(types.Dimension)
			}
			return cell.Width.UnmarshalText([]byte(val.(string)))
		},
		"cell.height": func(e types.IElement, parent types.IElement, val any) error {
			cell := e.(*types.Cell)
			if cell.Height == nil {
				cell.Height = new(types.Dimension)
			}
			return cell.Height.UnmarshalText([]byte(val.(string)))
		},
		"cell.absolute": func(e types.IElement, parent types.IElement, val any) error {
			var err error
			cell := e.(*types.Cell)
			cell.Absolute, err = strconv.ParseBool(val.(string))
			return err
		},
		"cell.display": func(e types.IElement, parent types.IElement, val any) error {
			cell := e.(*types.Cell)
			return cell.TextStyle.DisplayStyle.Parse(val.(string))
		},
		"font-family": func(e types.IElement, parent types.IElement, val any) error {
			el := e.GetElement()
			el.TextStyle.FontName = val.(string)
			return nil
		},
		"line-join-style": func(e types.IElement, parent types.IElement, val any) error {
			el := e.GetElement()
			if err := el.TextStyle.JoinStyle.Parse(val.(string)); err != nil {
				return err
			}
			el.Brush.JoinStyle = el.TextStyle.JoinStyle
			return nil
		},
		"line-cap-style": func(e types.IElement, parent types.IElement, val any) error {
			el := e.GetElement()
			if err := el.TextStyle.CapStyle.Parse(val.(string)); err != nil {
				return err
			}
			el.Brush.CapStyle = el.TextStyle.CapStyle
			return nil
		},
	}
)
