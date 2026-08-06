package media

import "strings"

// FullURL returns an absolute URL for a stored static resource path.
func FullURL(path, domain string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return domain + path
}
