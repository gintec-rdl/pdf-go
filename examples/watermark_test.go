package examples_test

import (
	"testing"

	pdfgo "github.com/gintec-rdl/pdf-go"
)

func TestWatermark(t *testing.T) {
	err := pdfgo.Generate(makeTestFileNames("watermark_test"))
	if err != nil {
		t.Fatal(err)
		return
	}
}
