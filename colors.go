// Package courier colors is a set of default term colours
package courier

import "fmt"

const (
	colorDefault = "\x1b[39m"
	colorRed     = "\x1b[0;31m"
	colorBlue    = "\x1b[94m"
	colorGreen   = "\x1b[32m"
)

func Red(s string) string {
	return fmt.Sprintf("%s%s%s", colorRed, s, colorDefault)
}

func Blue(s string) string {
	return fmt.Sprintf("%s%s%s", colorBlue, s, colorDefault)
}

func Green(s string) string {
	return fmt.Sprintf("%s%s%s", colorGreen, s, colorDefault)
}
