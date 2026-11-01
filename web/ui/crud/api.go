package crud

import (
	"context"
	"strconv"

	"github.com/deltegui/owl/v3/island/client"
)

// Send, FieldErrors, ErrorMessage and Escape are OWL's (owl/island/client), here
// under the names the screens already use.
func Send[T any](ctx context.Context, method, url string, body any) (T, error) {
	return client.Send[T](ctx, method, url, body)
}

var (
	FieldErrors  = client.FieldErrors
	ErrorMessage = client.ErrorMessage
	Escape       = client.Escape
)

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
