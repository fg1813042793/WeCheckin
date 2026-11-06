package httpadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"wecheckin/backend/internal/modules/publicportal/application"
)

func decodeBody(c *app.RequestContext, target any) error {
	body := bytes.TrimSpace(c.Request.Body())
	if len(body) == 0 || len(body) > 32*1024 || body[0] != '{' {
		return application.ErrInvalidArgument
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return application.ErrInvalidArgument
	}
	return nil
}
func positive(raw string) (uint64, error) {
	if raw == "" || raw[0] < '1' || raw[0] > '9' {
		return 0, application.ErrInvalidArgument
	}
	for _, c := range raw {
		if c < '0' || c > '9' {
			return 0, application.ErrInvalidArgument
		}
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		return 0, application.ErrInvalidArgument
	}
	return n, nil
}
func queryValue(c *app.RequestContext, key string) (string, error) {
	count := 0
	value := ""
	c.Request.URI().QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			count++
			value = string(v)
		}
	})
	if count > 1 {
		return "", application.ErrInvalidArgument
	}
	return value, nil
}
func pageValue(c *app.RequestContext, key string, defaultValue int) (int, error) {
	raw, err := queryValue(c, key)
	if err != nil {
		return 0, err
	}
	if raw == "" {
		if c.Request.URI().QueryArgs().Has(key) {
			return 0, application.ErrInvalidArgument
		}
		return defaultValue, nil
	}
	n, err := positive(raw)
	if err != nil || n > math.MaxInt {
		return 0, application.ErrInvalidArgument
	}
	return int(n), nil
}
func optionsQuery(c *app.RequestContext) (application.OptionsQuery, error) {
	kind, err := queryValue(c, "kind")
	if err != nil {
		return application.OptionsQuery{}, err
	}
	keyword, err := queryValue(c, "keyword")
	if err != nil {
		return application.OptionsQuery{}, err
	}
	page, err := pageValue(c, "page", 1)
	if err != nil {
		return application.OptionsQuery{}, err
	}
	size, err := pageValue(c, "pageSize", 20)
	if err != nil {
		return application.OptionsQuery{}, err
	}
	q := application.OptionsQuery{Kind: application.Kind(kind), Keyword: keyword, Page: page, PageSize: size}
	err = application.ValidateOptions(&q)
	return q, err
}
