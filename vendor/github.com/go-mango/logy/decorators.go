package logy

import (
	"fmt"
	"strconv"
)

func colorful(c int, s string) string {
	return "\033[" + strconv.Itoa(c) + "m" + s + "\033[0m"
}

func transform(s interface{}) string {
	f := "%v"

	switch s.(type) {
	case error:
		f = s.(error).Error()
	case fmt.Stringer:
		f = s.(fmt.Stringer).String()
	case string:
		f = s.(string)
	}

	return f
}
