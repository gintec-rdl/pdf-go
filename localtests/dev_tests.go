package localtests_test

import (
	"testing"

	pdfgo "github.com/gintec-rdl/pdf-go"
	"github.com/gintec-rdl/pdf-go/pkg/types"
)

func TestReport(t *testing.T) {
	loader := pdfgo.CreatePdfTemplateLoader()
	template, err := loader.LoadF("report.json")
	if err != nil {
		t.Fatal(err)
		return
	}
	pdfDoc, err := pdfgo.CreatePdfDocumentT(template)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = template.RenderF(pdfDoc, "reports2.pdf")
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestLoadTemplate(t *testing.T) {
	loader := pdfgo.CreatePdfTemplateLoader()
	template, err := loader.LoadF("template.json")
	if err != nil {
		t.Fatal(err)
		return
	}
	pdfDoc, err := pdfgo.CreatePdfDocumentT(template)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = template.RenderF(pdfDoc, "template.pdf")
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestBuilder(t *testing.T) {
	builder := pdfgo.CreatePdfTemplateBuilder(types.PO_PORTRAIT, types.PAGE_SIZE_A4, types.DU_MILIMETER)

	// documentTitle
	builder.Title("Report Builder test").
		ShowBookmarks(true).
		PageBookmarkTemplate("").
		Style("document", types.PdfTemplateAttributes{
			"background-color": "#f0f8ff",
		}).
		Style("page", types.PdfTemplateAttributes{
			"border-width":    ".7",
			"border-color":    "#80333333",
			"line-join-style": "round", // miter, round, bevel
			"line-cap-style":  "cap",   // cap, butt, square
		}).
		Style("centered-row", types.PdfTemplateAttributes{
			"background-color": "#101212EE",
			"font-color":       "#ffffff",
			"text-align":       "MC",
			"width":            "100%",
			"border-width":     ".5",
			"border-color":     "#80337733",
		}).Style("left-half", types.PdfTemplateAttributes{
		"background-color": "#80f50000",
		"font-color":       "#ffffff",
		"text-align":       "MR",
		"display":          "row",
		"width":            "50%",
		"border-width":     ".9",
		"border-color":     "#000F90",
	}).Style("right-half", types.PdfTemplateAttributes{
		"background-color": "#80f56666",
		"font-color":       "#ff000000",
		"width":            "50%",
	}).Style("row-match-parent", types.PdfTemplateAttributes{
		"width":      "100%",
		"font-color": "#000000",
	}).
		AddPage().StyleList("document", "page").BookmarkTitle("${page} - Override bookmark").
		AddCell().Text("Hello").StyleList("right-half").Parent().
		AddCell().Text("World").StyleList("left-half").Parent().
		AddCell().Text("Centered").StyleList("centered-row").Parent().
		Builder().AddPage().AddCell().Text("Second page").StyleList("centered-row", "row-match-parent")

	template, err := builder.Build()
	if err != nil {
		t.Fatal(err)
		return
	}

	// save to file
	err = template.SaveF("template.json")
	if err != nil {
		t.Fatal(err)
		return
	}
}
