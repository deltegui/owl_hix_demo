package entryweb

import (
	"context"
	"strconv"

	"owldemo/web/ui/crud"
)

const base = "/api/entries"

// The calls of the island to the API of the feature (see entry_handler.go).

func fetchList(ctx context.Context, q Query) (Listing, error) {
	query := ""
	add := func(key, value string) {
		if value != "" {
			query += key + "=" + crud.Escape(value) + "&"
		}
	}
	add("concept", q.Concept)
	add("category", q.Category)
	add("operation", q.Operation)
	add("from", q.From)
	add("to", q.To)
	add("orderBy", q.OrderBy)
	add("order", q.Order)
	return crud.Send[Listing](ctx, "GET", base+"?"+query+"page="+strconv.Itoa(max(q.Page, 1)), nil)
}

func fetchEntry(ctx context.Context, id int) (Entry, error) {
	if id == 0 {
		return crud.Send[Entry](ctx, "GET", "/api/entry-defaults", nil)
	}
	return crud.Send[Entry](ctx, "GET", base+"/"+strconv.Itoa(id), nil)
}

func saveEntry(ctx context.Context, e Entry) (Entry, error) {
	if e.ID == 0 {
		return crud.Send[Entry](ctx, "POST", base, e)
	}
	return crud.Send[Entry](ctx, "PUT", base+"/"+strconv.Itoa(e.ID), e)
}

func deleteEntry(ctx context.Context, id int) error {
	_, err := crud.Send[struct{}](ctx, "DELETE", base+"/"+strconv.Itoa(id), nil)
	return err
}

func priceLine(ctx context.Context, r LineRequest) (Line, error) {
	return crud.Send[Line](ctx, "POST", base+"/line", r)
}

func convertEntry(ctx context.Context, r ConvertRequest) (Entry, error) {
	return crud.Send[Entry](ctx, "POST", base+"/convert", r)
}
