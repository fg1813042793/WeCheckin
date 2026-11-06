package application

import (
	"math"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validateSettings(s *Settings) error {
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	s.LogoURL = strings.TrimSpace(s.LogoURL)
	// Accepted only for old clients; domains no longer control page or QR links.
	s.PublicBaseURL = ""
	if s.Version == 0 || s.Version == math.MaxUint64 || !validText(s.Title, 100) || s.Title == "" || !validText(s.Description, 1000) {
		return ErrInvalidArgument
	}
	if !validLogo(s.LogoURL) {
		return ErrInvalidLogo
	}
	return nil
}
func validateItem(s *ItemInput) error {
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	if !s.Kind.Valid() || s.SourceID == 0 || s.Sort < 0 || s.Sort > 99999 || !validText(s.Title, 255) || !validText(s.Description, 1000) {
		return ErrInvalidArgument
	}
	return nil
}
func ValidateOptions(q *OptionsQuery) error {
	q.Keyword = strings.TrimSpace(q.Keyword)
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if !q.Kind.Valid() || !validText(q.Keyword, 255) || q.Page < 1 || q.PageSize < 1 || q.PageSize > MaxPageSize || q.Page > math.MaxInt/q.PageSize {
		return ErrInvalidArgument
	}
	return nil
}
func validText(s string, max int) bool {
	if !utf8.ValidString(s) || utf8.RuneCountInString(s) > max {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return false
		}
	}
	return true
}
func safeURL(s string) (*url.URL, bool) {
	if strings.ContainsAny(s, "\\\r\n\t") || strings.Contains(s, " ") {
		return nil, false
	}
	u, err := url.Parse(s)
	if err != nil || u.Opaque != "" || u.User != nil {
		return nil, false
	}
	return u, true
}
func validLogo(s string) bool {
	if s == "" {
		return true
	}
	if len(s) > 2048 {
		return false
	}
	u, ok := safeURL(s)
	if !ok {
		return false
	}
	if strings.HasPrefix(s, "/") {
		return !strings.HasPrefix(s, "//") && u.Host == "" && u.Scheme == "" && !strings.HasPrefix(u.Path, "//") && !strings.Contains(u.Path, "\\")
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}
