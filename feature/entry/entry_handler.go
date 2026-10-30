package entry

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/deltegui/owl/v3"
	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/pagination"

	"owldemo/feature/entry/web"
	"owldemo/sys/ledger"
	"owldemo/sys/persistence"
	"owldemo/web"
	"owldemo/web/layout"
	"owldemo/web/ui/crud"
)

const viewEntryIndex string = "entry/entry_index"

func texts(ctx owl.Ctx) crud.Texts { return crud.Texts(ctx.GetLocalizer(viewEntryIndex)) }

func listing(list pagination.List[ledger.Entry], sum int64) entryweb.Listing {
	p := list.Pagination
	out := entryweb.Listing{
		Items: make([]entryweb.Row, 0, len(list.Items)),
		Page:  p.CurrentPage,
		Pages: max(1, (p.TotalElements+p.ElementsPerPage-1)/p.ElementsPerPage),
		Total: p.TotalElements,
		Sum:   euros(sum),
	}
	for _, e := range list.Items {
		out.Items = append(out.Items, row(e))
	}
	return out
}

// queryOf reads the filter and the order of the list from the query string.
func queryOf(ctx owl.Ctx) (entryFilter, entryOrder, int) {
	filter := entryFilter{Concept: ctx.GetQueryParam("concept")}
	filter.CategoryID, _ = strconv.Atoi(ctx.GetQueryParam("category"))
	filter.Operation, _ = strconv.Atoi(ctx.GetQueryParam("operation"))
	if day, err := time.Parse(dateLayout, ctx.GetQueryParam("from")); err == nil {
		filter.From = day
	}
	if day, err := time.Parse(dateLayout, ctx.GetQueryParam("to")); err == nil {
		filter.To = day
	}
	order := entryOrder{By: ctx.GetQueryParam("orderBy"), Desc: ctx.GetQueryParam("order") != "asc"}
	page, _ := strconv.Atoi(ctx.GetQueryParam("page"))
	return filter, order, page
}

// handleShowIndex renders the page: the maintenance is one island, which the
// server draws from its first page and the browser hydrates.
func handleShowIndex(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		list, sum, err := service.Find(entryFilter{}, entryOrder{By: "carriedOutAt", Desc: true}, pagination.Pagination{CurrentPage: 1})
		if err != nil {
			return ctx.StringInternalServerError("Error finding entries")
		}
		first := listing(list, sum)
		t := texts(ctx)
		state, _ := json.Marshal(entryweb.State{Listing: &first, Lookups: lookups(), Texts: t})
		return layout.Render(ctx, layout.Page{
			Title:   t.Get("Title"),
			Bundle:  "entries",
			Islands: []layout.Island{{Name: "entries", State: string(state), Build: entryweb.Island}},
		})
	}
}

func handleList(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		filter, order, page := queryOf(ctx)
		list, sum, err := service.Find(filter, order, pagination.Pagination{CurrentPage: page})
		if err != nil {
			return ctx.StringInternalServerError("Error finding entries")
		}
		return ctx.Json(listing(list, sum))
	}
}

func idOf(ctx owl.Ctx) (int, bool) {
	id, err := strconv.Atoi(ctx.GetURLParam("id"))
	return id, err == nil && id > 0
}

func handleShowOne(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		id, ok := idOf(ctx)
		if !ok {
			return ctx.StringBadRequest("Invalid Id")
		}
		e, err := service.FindOne(id)
		if err != nil {
			return failure(ctx, err)
		}
		return ctx.Json(toDTO(e))
	}
}

// handleShowCreate is a new entry with what a new entry starts with, as the
// editor needs it.
func handleShowCreate() owl.Handler {
	return func(ctx owl.Ctx) error { return ctx.Json(blank(time.Now())) }
}

// handleSave creates the entry, or updates it when the URL has an id. It checks
// what the tags of the model say (valtruc), and then what only the mode knows.
func handleSave(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		var dto entryweb.Entry
		if err := ctx.ParseJson(&dto); err != nil {
			return ctx.StringBadRequest("Invalid JSON")
		}
		if raw := ctx.GetURLParam("id"); raw != "" {
			id, ok := idOf(ctx)
			if !ok {
				return ctx.StringBadRequest("Invalid Id")
			}
			dto.ID = id
		} else {
			dto.ID = 0
		}

		validate(&ctx, dto)
		if !ctx.ModelState.Valid {
			return web.Invalid(ctx, viewEntryIndex)
		}
		e, err := toDomain(dto)
		if err != nil {
			return domainFailure(ctx, err)
		}
		saved, err := service.Save(e)
		if err != nil {
			return failure(ctx, err)
		}
		return ctx.Json(toDTO(saved))
	}
}

func handleDelete(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		id, ok := idOf(ctx)
		if !ok {
			return ctx.StringBadRequest("Invalid Id")
		}
		if err := service.Delete(id); err != nil {
			return failure(ctx, err)
		}
		return ctx.NoContent()
	}
}

// handleLine answers the rate of a line: how many euros a unit of its currency
// is worth on the day of the entry.
func handleLine(service entryService) owl.Handler {
	return func(ctx owl.Ctx) error {
		var req entryweb.LineRequest
		if err := ctx.ParseJson(&req); err != nil {
			return ctx.StringBadRequest("Invalid JSON")
		}
		line, err := service.PriceLine(req.Line, req.ExchangeDate)
		if err != nil {
			return ctx.StringBadRequest("%s", err.Error())
		}
		return ctx.Json(line)
	}
}

// handleConvert turns the entry of the editor into the other mode.
func handleConvert() owl.Handler {
	return func(ctx owl.Ctx) error {
		var req entryweb.ConvertRequest
		if err := ctx.ParseJson(&req); err != nil {
			return ctx.StringBadRequest("Invalid JSON")
		}
		out, err := convert(req.To, req.Entry)
		if err != nil {
			return domainFailure(ctx, err)
		}
		return ctx.Json(out)
	}
}

// validate checks the entry: the tags of the model, and the rules of each mode
// (the same two LittleCash has: the basic mode needs its sentence complete, and
// every line of the advanced mode has either a debit or a credit).
func validate(ctx *owl.Ctx, dto entryweb.Entry) {
	ctx.Validate(dto)
	loc := ctx.GetLocalizer(viewEntryIndex)
	custom := func(identifier, field string) []core.ValidationError {
		return []core.ValidationError{core.CustomValidationError{
			ErrorMessage: loc.Get(identifier), StructName: "Entry", FieldName: field, Identifier: valtrucID(identifier),
		}}
	}
	fail := func(field, identifier string) {
		ctx.ModelState.Errors[field] = custom(identifier, field)
		ctx.ModelState.Valid = false
	}

	if dto.Mode == entryweb.ModeBasic {
		if dto.Amount == nil {
			fail("Amount", "customMandatory")
		}
		if dto.OperationId == 0 {
			fail("OperationId", "customMandatory")
		}
		if dto.CreditAccount == 0 {
			fail("CreditAccount", "customMandatory")
		}
		if dto.DebitAccount == 0 {
			fail("DebitAccount", "customMandatory")
		}
		return
	}
	if len(dto.Lines) < 2 {
		fail("Lines", "customMinTwoLines")
	}
	for i, line := range dto.Lines {
		if (line.Debit == nil) == (line.Credit == nil) {
			path := "Lines[" + strconv.Itoa(i) + "]"
			fail(path+".Debit", "customCreditDebitOnlyOneSetted")
			fail(path+".Credit", "customCreditDebitOnlyOneSetted")
		}
	}
}

func domainFailure(ctx owl.Ctx, err error) error {
	var domain core.DomainError
	if errors.As(err, &domain) {
		return web.Failure(ctx, domain)
	}
	return web.Failure(ctx, errInvalid.Wrap(err))
}

func failure(ctx owl.Ctx, err error) error {
	if errors.Is(err, persistence.ErrNotFound) {
		return web.Failure(ctx, errNotFound)
	}
	ctx.Status(http.StatusInternalServerError)
	return ctx.String("Internal error")
}
