package theme

import (
	"testing"

	"fyne.io/fyne/v2"
)

func BenchmarkTheme_current(b *testing.B) {
	fyne.CurrentSettings().SetTheme(LightTheme())

	for n := 0; n < b.N; n++ {
		Current()
	}
}
