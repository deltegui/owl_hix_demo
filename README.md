# OWL demo

The CRUD demo (products and contacts, a chart, generic forms, a theme selector) and a complex maintenance, **ledger entries**, equivalent to `ldg_entry` of LittleCash, made with **OWL** and **Hix**, and laid out like LittleCash: a folder per feature, each with its bootstrap, handler and service. Where the CRUD demo used only Hix and `net/http`, this one uses OWL for what it is for: routes, middleware, CSRF, validation, localization and pagination.

```sh
make serve     # builds the four WASM files and serves http://localhost:8080
make release   # the same four, optimized with TinyGo
make check     # 84 checks in a real browser (node 22+, chromium, network for Bootstrap and Chart.js)
```

## Layout

```
cmd/owldemo/main.go        wire(): the cipher, the Mux, the middleware, the list of modules
sys/
  ledger/                  the accounting: accounts, groups, basic operations, balance (plain Go, builds for the browser too)
  currency/                the currencies and a daily exchange rate
  bootstrap/               Context (what a feature gets) and BootstrapFunc
  config/                  the environment: OWLDEMO_ADDR, OWLDEMO_SECRET
  persistence/             Repository[T] and an in-memory implementation
feature/
  product/                 product_bootstrap.go, product_handler.go, product_service.go
    web/                   what the feature shows, on both sides: the model, the screen, the chart
  contact/                 the same, for contacts
  entry/                   the ledger entries maintenance (below)
    web/                   the list, the editor, the viewer and the models of its API
  forms/                   the generic form: it describes a struct and the browser draws it
  menu/                    the side menu (like LittleCash, it plugs a function into the layout)
  language/                POST /language/:lang
  about/                   a page that is only HTML
web/
  layout/                  the page around the features: menu, theme, language, islands
  locale/                  the translations of the whole app (shared, errors, pagination); the ones of a page are next to its code
  static/                  wasm_exec.js and the WASM files (embedded in the server)
  ui/                      Bootstrap wrappers (bs), the generic CRUD screen (crud), chart, theme, island
client/                    the three WASM entry points: lite, full and forms
```

A feature has the three files LittleCash has (`*_bootstrap.go` registers the routes in a sub-mux, `*_handler.go` has the handlers, `*_service.go` the logic and the data), plus a `web/` package for what must also build for the browser: the model, the screen and the island. The server renders each island from its state and the browser hydrates it (see `CRUD_DEMO` for how). A `web/` package imports Hix and `web/ui`, never OWL, so it can go in a WASM file.

## Ledger entries

`feature/entry` is the maintenance of `ldg_entry` in LittleCash: double-entry accounting, where an entry is balanced lines of debit and credit. It is the complex one: a list with a filter, sort and total, and an editor with two modes.

- **The list** filters by concept, category, operation and dates, sorts by clicking a header (an arrow shows the order, a second click turns it around), paginates keeping the order, and adds up the amounts of the whole filter, not only of the page. View, edit and delete (with a confirmation) are in each row.
- **Two layers**, like LittleCash's main and alternate layers: the editor or the viewer open over the list, which stays alive underneath, so the filter, the order and the page are where the user left them.
- **Basic mode** writes an entry as a sentence: *Expense of 99.50 USD from Bank to Operating expenses*. The accounts each side offers depend on the operation (an income takes money from an income account into an asset), and changing the operation replaces an account that does not fit.
- **Advanced mode** is a table of lines (account, debit, credit, currency, notes): typing in one side hides the other, an arrow shows if the amount raises or lowers the account, each line shows its amount in euros, and the sums of the two sides are red until they are equal. A line in another currency asks the server for its rate (each line has its own `hx.Action`, since the last call of one would cancel the call of another).
- **Changing the mode is the server's decision** (`POST /api/entries/convert`): a basic entry becomes two lines, and lines become a sentence only if they are a basic operation (a debit and a credit, one currency, accounts in the groups of an operation). An entry that is not one opens in advanced mode with the button disabled and the reason.
- **Validation** is the one of LittleCash: the valtruc tags of the model (`ctx.Validate`), then the rules of each mode as custom errors (`customMandatory` for an incomplete sentence, `customCreditDebitOnlyOneSetted` for a line), by field, even on a line (`lines[1].debit`); and a rule of the domain that is not about a field (the entry does not balance, in euros) is a `core.DomainError` whose code is translated in `web/locale/errors.go` and shown over the form.
- **The rules are in `sys/ledger`**, plain Go that the server uses to decide and the browser to draw the same arrows and lists. Amounts are `int64` cents, like LittleCash.

The files are the ones of any feature: `entry_bootstrap.go` (routes), `entry_handler.go` (the page and the JSON API), `entry_service.go` (filter, order, pagination, save), `entry_mapping.go` (the models of the editor and the entries of the domain, and the conversion) and `entry_lookup.go` (what the screens pick from), plus `web/`.

## What OWL does here

- **Routes and middleware.** `Mux.CreateSubMux("/api/products")`, `Get`/`Post`/`Put`/`Delete` with `:id`, `Use(middleware.Csrf(...))`, as in LittleCash.
- **Views are Hix.** `ctx.View(owl.HixView(...))` renders with `hx/ssr`. `layout.Render(ctx, page)` is the one place that builds the HTML of a page (what `views/shared/layout.html` is in LittleCash).
- **Localization.** Spanish by default (the first of `locale.Languages`), English with the buttons of the menu, which come from that same list (a form, so it works without JavaScript; `POST /language/:lang` keeps only paths of this site in `Next`). The texts of a page are a `localizer.Dict` next to its code (`feature/<feature>/web/<feature>_locale.go`, which also compiles to WASM), plus `web/locale/shared.go`; each one has a test with `localizer.Check`. The server translates the page with `hxui.NewView(ctx, dict)`, and the islands translate themselves: the state of an island carries only the language, and both the server and the browser build the words with `TextsFor(lang)` from the same dictionaries, so hydration matches.
- **CSRF.** The middleware checks every request that is not a GET. The page has the token in `<meta name="csrf-token">`, the language buttons send it as a hidden field and the islands send it in the `X-Csrf-Token` header (`crud.Send`).
- **Validation.** `ctx.Validate(item)` with the valtruc tags of the model. `ctx.JsonInvalid(texts)` (OWL) answers 422 with the errors by field, translated with the file of the view (the names of the fields come from there), and the island shows them under each field.
- **Pagination.** `pagination.Pagination` and `pagination.List[T]` in the services, as in LittleCash. The JSON answer is `crud.Page`.

## The WASM files

| | TinyGo | gzip | brotli |
|---|---|---|---|
| `lite.wasm` (theme selector) | 352 KB | 130 KB | 104 KB |
| `full.wasm` (screens, chart, JSON) | 1.50 MB | 504 KB | 361 KB |
| `forms.wasm` (generic form) | 1.29 MB | 454 KB | 330 KB |
| `entries.wasm` (ledger entries) | 1.57 MB | 533 KB | 381 KB |

Each page loads the one it needs (`Bundle` in `layout.Page`); a Go WASM cannot load code later. About loads `lite`, the lists load `full`, the forms load `forms` and the entries load `entries`.

## Checks

`make check` builds the WASM, starts its own server and a headless Chromium, and runs 84 checks:

- **The HTML of the server**, with no JavaScript, in Spanish: the tables, totals and the state of each island, the menu from the menu feature, and the binary each page loads.
- **CSRF and language:** a POST with no token, or a made up one, is refused (403); with the token it is accepted; a GET needs none; the 422 is translated and by field; the language form goes back to the page it came from and never to another site; with the cookie the page is in English.
- **Hydration** of every page: no mismatch, no error in the console, and none of the elements the server sent is created again.
- **The screens** (in English): filter, pagination, validation errors, create, edit, enable/disable and delete, the chart following the products, a rule that only the server checks, the generic forms and the theme.
- **The entries:** the list against the API (rows, order, filter, total), the view over the list, editing in basic mode (the accounts follow the operation), the server's rules on their fields, a new entry, the advanced mode (lines, sums, a line in dollars, the balance), an entry that cannot be a sentence, an entry that does not balance, and delete.

## Not here

A database (the repository is a map), sessions and login (LittleCash has them: `session.Manager` (`session.NewManager(cipher, opts...)`) and `middleware.Authorize` would go in `sys/bootstrap` and each feature's sub-mux), and the URL reflecting the filter and page.
