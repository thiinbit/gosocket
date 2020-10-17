/*
 * // Copyright (c) 2020 @thiinbit. All rights reserved.
 * // Use of this source code is governed by an MIT-style
 * // license that can be found in the LICENSE file
 *
 */

package gosocket

import (
	"fmt"
)

const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	colorMagenta //洋红
)

func Red(s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorRed, s)
}

func Green(s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorGreen, s)
}

func Yellow(s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorYellow, s)
}

func Blue(s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorBlue, s)
}

func Magenta(s interface{}) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorMagenta, s)
}
