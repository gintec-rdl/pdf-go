package examples_test

import (
	"os"
	"path"
	"testing"

	pdfgo "github.com/gintec-rdl/pdf-go"
)

func testdataFile(filename string) string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, "../testdata", filename)
}

func testDataOutputFile(filename string) string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, "../testdata/output", filename)
}

func makeTestFileNames(in string) (string, string) {
	return testdataFile(in + ".json"), testDataOutputFile(in + ".pdf")
}

func TestDocument(t *testing.T) {
	err := pdfgo.Generate(makeTestFileNames("document_test"))
	if err != nil {
		t.Fatal(err)
		return
	}
}
