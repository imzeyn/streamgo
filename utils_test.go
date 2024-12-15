package streamgo

import (
	"strings"
	"testing"
)

// This function have is old versions
func oldClearURL(s string) string {
	var builder strings.Builder
	builder.WriteString("/")

	for _, v := range strings.Split(s, "/") {
		if strings.TrimSpace(v) != "" {
			builder.WriteString(v + "/")
		}
	}

	return builder.String()
}


func Benchmark(b *testing.B) {
    url := "/bpl6z3jkii/12b/1j6ifkgi4///abj//ea3//c2ss81q/jdt1nm5ihf4of8jlq9xy/z/8ue7dbb73jhfpgk7duxq5qj0nxqa0hshv//pszgfxk0///////"
	b.Run("oldClearURL", func(b *testing.B) {
        for i := 0; i < 1; i++ {
            oldClearURL(url)
        }
 	})

	b.Run("ClearURL", func(b *testing.B) {
        for i := 0; i < 1; i++ {
            ClearURL(url)            
        }

	})
}

 