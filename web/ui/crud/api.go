package crud

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/deltegui/hx/hxhttp"
	"github.com/deltegui/hx/hxjson"
)

// statusUnprocessableEntity is http.StatusUnprocessableEntity. net/http is not imported
// on purpose: it brings crypto, TLS and MIME into the WASM, hundreds of KB for a number.
const statusUnprocessableEntity = 422

// csrfHeader is the name OWL's CSRF middleware reads the token from.
const csrfHeader = "X-Csrf-Token"

// Send does what hxjson.Send does and adds the CSRF token, which the server
// checks on every request that is not a GET.
func Send[T any](ctx context.Context, method, url string, body any) (T, error) {
	var zero T
	request := hxhttp.Request{Method: method, URL: url, Header: map[string]string{
		"Accept":   "application/json",
		csrfHeader: csrfToken(),
	}}
	if body != nil {
		encoded, err := hxjson.Body(body)
		if err != nil {
			return zero, err
		}
		request.Body = encoded
	}
	response, err := hxhttp.Do(ctx, request)
	if err != nil {
		return zero, err
	}
	if err := response.Err(); err != nil {
		return zero, err
	}
	var result T
	if len(response.Bytes()) == 0 {
		return result, nil
	}
	if err := hxjson.Decode(response, &result); err != nil {
		return zero, err
	}
	return result, nil
}

// Client talks to the JSON API of one entity (see the handlers of its feature).
type Client[T any] struct{ Base string }

func (c Client[T]) List(ctx context.Context, query string, page int) (Page[T], error) {
	return Send[Page[T]](ctx, "GET", c.Base+"?"+query+"page="+strconv.Itoa(page), nil)
}

// Save creates the item when id is 0 and updates it otherwise.
func (c Client[T]) Save(ctx context.Context, id int, body map[string]any) (T, error) {
	if id == 0 {
		return Send[T](ctx, "POST", c.Base, body)
	}
	return Send[T](ctx, "PUT", c.Base+"/"+strconv.Itoa(id), body)
}

func (c Client[T]) Delete(ctx context.Context, id int) error {
	_, err := Send[struct{}](ctx, "DELETE", c.Base+"/"+strconv.Itoa(id), nil)
	return err
}

func (c Client[T]) Toggle(ctx context.Context, id int) (T, error) {
	return Send[T](ctx, "POST", c.Base+"/"+strconv.Itoa(id)+"/toggle", nil)
}

// FieldErrors reads the errors by field out of a 422 answer.
func FieldErrors(err error) (map[string][]string, bool) {
	var httpErr *hxhttp.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != statusUnprocessableEntity {
		return nil, false
	}
	var body Errors
	if json.Unmarshal([]byte(httpErr.Body), &body) != nil {
		return nil, false
	}
	return body.Errors, true
}

// ErrorMessage reads the message of a rule of the domain out of an answer that
// has one (see Errors), or says there is none.
func ErrorMessage(err error) (string, bool) {
	var httpErr *hxhttp.HTTPError
	if !errors.As(err, &httpErr) {
		return "", false
	}
	var body Errors
	if json.Unmarshal([]byte(httpErr.Body), &body) != nil || body.Message == "" {
		return "", false
	}
	return body.Message, true
}

// Escape percent-encodes a query value. It is here so the WASM does not import
// net/url, which adds size.
func Escape(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&15])
	}
	return b.String()
}
