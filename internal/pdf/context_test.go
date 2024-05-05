package pdf_test

import (
	"fmt"
	"testing"

	"github.com/gintec-rdl/pdf-go/internal/pdf"
	"github.com/stretchr/testify/assert"
)

func TestContextPush(t *testing.T) {
	ctx := pdf.ContextStack{}
	ctx.PushT("L", "I", "E")
	assert.Equal(t, "LIE", fmt.Sprintf("%s%s%s", ctx.Pop(), ctx.Pop(), ctx.Pop()))

	ctx.PushT("L", "E", "G")
	a, b, c := pdf.PopTrio[string](&ctx)
	assert.Equal(t, "LEG", fmt.Sprintf("%s%s%s", a, b, c))
}
