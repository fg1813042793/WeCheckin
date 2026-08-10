package review

import "strings"

func functionBody(src, signature string) string {
	start := strings.Index(src, signature)
	if start < 0 {
		return ""
	}
	body := src[start:]
	depth := 0
	for index, r := range body {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return body[:index+1]
			}
		}
	}
	return body
}
