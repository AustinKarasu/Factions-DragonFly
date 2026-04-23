package command

import "strings"

func stringsJoin(lines []string) string {
	return strings.Join(lines, "\n")
}
