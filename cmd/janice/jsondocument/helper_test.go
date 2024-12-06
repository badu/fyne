package jsondocument_test

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/cmd/janice/jsondocument"
	"github.com/stretchr/testify/assert"
)

func TestHelper(t *testing.T) {
	r := strings.NewReader("test")
	r2 := jsondocument.MakeURIReadCloser(r, "alpha")
	assert.Equal(t, "alpha", r2.URI().Name())
}
