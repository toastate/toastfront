package builder

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

var windowCRregexp = regexp.MustCompile(`\r?\n`)

func replaceWindowsCarriageReturn(b []byte) []byte {
	return windowCRregexp.ReplaceAll(b, []byte("\n"))
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	return io.Copy(destination, source)
}
