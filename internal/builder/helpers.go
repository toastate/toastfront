package builder

import (
	"regexp"
)

var windowCRregexp = regexp.MustCompile(`\r?\n`)

func replaceWindowsCarriageReturn(b []byte) []byte {
	return windowCRregexp.ReplaceAll(b, []byte("\n"))
}
