package util

import "strings"

func TrimmedString(s string, sep string) string {

	slc := strings.Split(s, sep)

	return slc[len(slc)-1]

}
